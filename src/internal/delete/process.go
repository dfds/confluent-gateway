package create

import (
	"context"
	"errors"
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
	CapabilityId models.CapabilityId
	ClusterId    models.ClusterId
	TopicId      string
	TopicName    string
}

func (p *process) Process(ctx context.Context, input ProcessInput) error {
	session := p.database.NewSession(ctx)

	state, err := p.prepareProcessState(session, input)
	if err != nil {
		if errors.Is(err, ErrTopicNotFound) {
			// topic already exists => skip
			p.logger.Warning("{Topic} on {Cluster} for {Capability} not found", input.TopicName, string(input.CapabilityId), string(input.ClusterId))
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
	topic, err := tx.GetTopic(input.TopicId)
	if err != nil {
		return err
	}

	if topic == nil {
		return ErrTopicNotFound
	}

	return nil
}

type stateRepository interface {
	GetDeleteProcessState(capabilityId models.CapabilityId, clusterId models.ClusterId, topicName string) (*models.DeleteProcess, error)
	SaveDeleteProcessState(state *models.DeleteProcess) error
}

func getOrCreateProcessState(repo stateRepository, input ProcessInput) (*models.DeleteProcess, error) {
	capabilityId, clusterId, topicName := input.CapabilityId, input.ClusterId, input.TopicName

	state, err := repo.GetDeleteProcessState(capabilityId, clusterId, topicName)
	if err != nil {
		return nil, err
	}

	if state != nil && !state.IsCompleted() {
		// is process is unfinished => continue
		return state, nil
	}

	state = models.NewDeleteProcess(capabilityId, clusterId, input.TopicId, topicName)

	if err := repo.SaveDeleteProcessState(state); err != nil {
		return nil, err
	}

	return state, nil
}

func (p *process) getStepContext(ctx context.Context, tx models.Transaction, state *models.DeleteProcess) *StepContext {
	logger := p.logger
	topic := NewTopicService(ctx, p.confluent, tx)
	outbox := p.factory(tx)

	return NewStepContext(logger, state, topic, outbox)
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