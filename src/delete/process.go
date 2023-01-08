package create

import (
	"context"
	"errors"
	"github.com/dfds/confluent-gateway/logging"
	"github.com/dfds/confluent-gateway/messaging"
	"github.com/dfds/confluent-gateway/models"
	. "github.com/dfds/confluent-gateway/process"
	"github.com/satori/go.uuid"
	"strings"
)

type process struct {
	logger    logging.Logger
	database  models.Database
	confluent Confluent
	registry  messaging.OutgoingMessageRegistry
}

func NewProcess(logger logging.Logger, database models.Database, confluent Confluent, registry messaging.OutgoingMessageRegistry) Process {
	return &process{
		logger:    logger,
		database:  database,
		confluent: confluent,
		registry:  registry,
	}
}

type ProcessInput struct {
	CapabilityRootId models.CapabilityRootId
	ClusterId        models.ClusterId
	TopicName        string
}

func (p *process) Process(ctx context.Context, input ProcessInput) error {
	session := p.database.NewSession(ctx)

	state, err := p.prepareProcessState(session, input)
	if err != nil {
		if errors.Is(err, ErrTopicNotFound) {
			// topic already exists => skip
			p.logger.Warning("{Topic} on {Cluster} for {Capability} not found", input.TopicName, string(input.CapabilityRootId), string(input.ClusterId))
			return nil
		}

		return err
	}

	if state.IsCompleted() {
		// already completed => skip
		return nil
	}

	return PrepareSteps[*StepContext]().
		Step(ensureTopicIsDeleted).
		Run(func(step func(*StepContext) error) error {
			return session.Transaction(func(tx models.Transaction) error {
				stepContext := p.getStepContext(ctx, tx, state)

				err := step(stepContext)
				if err != nil {
					return err
				}

				return tx.UpdateDeleteProcessState(state)
			})
		})
}

func (p *process) prepareProcessState(session models.Session, input ProcessInput) (*models.DeleteProcess, error) {
	var s *models.DeleteProcess

	err := session.Transaction(func(tx models.Transaction) error {
		if err := ensureTopicExists(tx, input); err != nil {
			return err
		}

		state, err := getOrCreateProcessState(tx, input)
		if err != nil {
			return err
		}

		s = state

		return nil
	})

	return s, err
}

var ErrTopicNotFound = errors.New("topic not found")

func ensureTopicExists(tx models.Transaction, input ProcessInput) error {
	capabilityRootId, clusterId, topicName := input.CapabilityRootId, input.ClusterId, input.TopicName

	topic, err := tx.GetTopic(capabilityRootId, clusterId, topicName)
	if err != nil {
		return err
	}

	if topic == nil {
		return ErrTopicNotFound
	}

	return nil
}

type stateRepository interface {
	GetDeleteProcessState(capabilityRootId models.CapabilityRootId, clusterId models.ClusterId, topicName string) (*models.DeleteProcess, error)
	SaveDeleteProcessState(state *models.DeleteProcess) error
}

func getOrCreateProcessState(repo stateRepository, input ProcessInput) (*models.DeleteProcess, error) {
	capabilityRootId, clusterId, topicName := input.CapabilityRootId, input.ClusterId, input.TopicName

	state, err := repo.GetDeleteProcessState(capabilityRootId, clusterId, topicName)
	if err != nil {
		return nil, err
	}

	if state != nil && !state.IsCompleted() {
		// is process is unfinished => continue
		return state, nil
	}

	state = models.NewDeleteProcess(capabilityRootId, clusterId, topicName)

	if !strings.HasSuffix(topicName, "-cg") {
		// TODO -- stop faking
		state.MarkAsCompleted()
	}

	if err := repo.SaveDeleteProcessState(state); err != nil {
		return nil, err
	}

	return state, nil
}

func (p *process) getStepContext(ctx context.Context, tx models.Transaction, state *models.DeleteProcess) *StepContext {
	logger := p.logger
	topic := NewTopicService(ctx, p.confluent, tx)
	outbox := p.getOutbox(tx)

	return NewStepContext(logger, state, topic, outbox)
}

func (p *process) getOutbox(tx models.Transaction) *messaging.Outbox {
	return messaging.NewOutbox(p.logger, p.registry, tx, func() string { return uuid.NewV4().String() })
}

// region Steps

func ensureTopicIsDeleted(stepContext *StepContext) error {
	stepContext.logger.Trace("Running {Step}", "EnsureTopicIsCreated")
	return ensureTopicIsDeletedStep(stepContext)
}

type EnsureTopicIsDeletedStep interface {
	IsCompleted() bool
	DeleteTopic() error
	MarkAsCompleted()
	RaiseTopicDeletedEvent() error
}

func ensureTopicIsDeletedStep(step EnsureTopicIsDeletedStep) error {
	if step.IsCompleted() {
		return nil
	}

	err := step.DeleteTopic()
	if err != nil {
		return err
	}

	step.MarkAsCompleted()

	return step.RaiseTopicDeletedEvent()
}

// endregion
