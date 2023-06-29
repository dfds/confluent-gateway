package create

import (
	"context"
	"errors"
	"fmt"

	"github.com/dfds/confluent-gateway/internal/models"
	. "github.com/dfds/confluent-gateway/internal/process"
	"github.com/dfds/confluent-gateway/logging"
)

type process struct {
	logger    logging.Logger
	database  models.Database
	confluent Confluent
	factory   OutboxFactory
}

func NewProcess(logger logging.Logger, database models.Database, confluent Confluent, factory OutboxFactory) Process {
	return &process{
		logger:    logger,
		database:  database,
		confluent: confluent,
		factory:   factory,
	}
}

type ProcessInput struct {
	TopicId      string
	CapabilityId models.CapabilityId
	ClusterId    models.ClusterId
	Topic        models.TopicDescription
}

func (p *process) Process(ctx context.Context, input ProcessInput) error {
	session := p.database.NewSession(ctx)

	state, err := p.prepareProcessState(session, input)
	if err != nil {
		if errors.Is(err, ErrTopicAlreadyExists) {
			// topic already exists => skip
			return nil
		}

		return err
	}

	return PrepareSteps[*StepContext]().
		Step(ensureHasValidServiceAccount).
		Step(ensureTopicIsCreated).
		Run(func(step func(*StepContext) error) error {
			return session.Transaction(func(tx models.Transaction) error {
				stepContext := p.getStepContext(ctx, tx, state)

				err := step(stepContext)
				if err != nil {
					return err
				}

				return tx.UpdateCreateProcessState(state)
			})
		})
}

func (p *process) prepareProcessState(session models.Session, input ProcessInput) (*models.CreateProcess, error) {
	var s *models.CreateProcess

	err := session.Transaction(func(tx models.Transaction) error {
		outbox := p.factory(tx)

		if err := ensureNewTopic(tx, input); err != nil {
			p.logger.Warning("{Topic} on {Cluster} for {Capability} already exists", input.Topic.Name, string(input.CapabilityId), string(input.ClusterId))
			return err
		}

		state, err := getOrCreateProcessState(tx, outbox, input)
		if err != nil {
			return err
		}

		s = state

		return nil
	})

	return s, err
}

var ErrTopicAlreadyExists = errors.New("topic already exists")

func ensureNewTopic(tx models.Transaction, input ProcessInput) error {
	topic, err := tx.GetTopic(input.TopicId)
	if err != nil {
		return err
	}

	if topic != nil {
		return ErrTopicAlreadyExists
	}

	return nil
}

type stateRepository interface {
	GetCreateProcessState(capabilityId models.CapabilityId, clusterId models.ClusterId, topicName string) (*models.CreateProcess, error)
	GetServiceAccount(capabilityId models.CapabilityId) (*models.ServiceAccount, error)
	SaveCreateProcessState(state *models.CreateProcess) error
}

func getOrCreateProcessState(repo stateRepository, outbox Outbox, input ProcessInput) (*models.CreateProcess, error) {
	capabilityId, clusterId, topic := input.CapabilityId, input.ClusterId, input.Topic

	state, err := repo.GetCreateProcessState(capabilityId, clusterId, topic.Name)
	if err != nil {
		return nil, err
	}

	if state != nil && !state.IsCompleted() {
		// is process is unfinished => continue
		return state, nil
	}

	state = models.NewCreateProcess(capabilityId, clusterId, input.TopicId, topic)

	if err := repo.SaveCreateProcessState(state); err != nil {
		return nil, err
	}

	err = outbox.Produce(&TopicProvisioningBegun{
		TopicId:      input.TopicId,
		CapabilityId: string(capabilityId),
		ClusterId:    string(clusterId),
		TopicName:    topic.Name,
	})

	if err != nil {
		return nil, err
	}

	return state, nil
}

func (p *process) getStepContext(ctx context.Context, tx models.Transaction, state *models.CreateProcess) *StepContext {
	logger := p.logger
	newAccountService := NewAccountService(ctx, tx)
	topic := NewTopicService(ctx, p.confluent, tx)
	outbox := p.factory(tx)

	return NewStepContext(logger, state, newAccountService, topic, outbox)
}

// region Steps

func ensureHasValidServiceAccount(stepContext *StepContext) error {
	stepContext.logger.Trace("Running {Step}", "EnsureServiceAccount")
	return ensureHasValidServiceAccountStep(stepContext)
}

type EnsureServiceAccountStep interface {
	HasClusterAccessWithValidAcls() bool
}

func ensureHasValidServiceAccountStep(step EnsureServiceAccountStep) error {
	if step.HasClusterAccessWithValidAcls() {
		return nil
	}

	return fmt.Errorf("no service account for capability to provision topic")
}

func ensureTopicIsCreated(stepContext *StepContext) error {
	stepContext.logger.Trace("Running {Step}", "EnsureTopicIsCreated")
	return ensureTopicIsCreatedStep(stepContext)
}

type EnsureTopicIsCreatedStep interface {
	IsCompleted() bool
	CreateTopic() error
	MarkAsCompleted()
	RaiseTopicProvisionedEvent() error
}

func ensureTopicIsCreatedStep(step EnsureTopicIsCreatedStep) error {
	if step.IsCompleted() {
		return nil
	}

	err := step.CreateTopic()
	if err != nil {
		return err
	}

	step.MarkAsCompleted()

	return step.RaiseTopicProvisionedEvent()
}

// endregion
