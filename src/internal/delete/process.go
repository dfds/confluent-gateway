package delete

import (
	"context"
	"errors"
	"github.com/dfds/confluent-gateway/internal/models"
	. "github.com/dfds/confluent-gateway/internal/process"
	"github.com/dfds/confluent-gateway/internal/storage"
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
	TopicId string
}

func (p *process) Process(ctx context.Context, input ProcessInput) error {
	session := p.database.NewSession(ctx)

	state, err := p.prepareProcessState(session, input)
	if err != nil {
		if errors.Is(err, storage.ErrTopicNotFound) {
			// topic already exists => skip
			p.logger.Warning("Topic with id {TopicId} not found", input.TopicId)
			return nil
		}

		return err
	}

	if state.IsCompleted() {
		// already completed => skip
		return nil
	}

	return PrepareSteps[*StepContext]().
		Step(ensureTopicSchemasAreDeleted).
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

func ensureTopicExists(tx models.Transaction, input ProcessInput) error {
	topic, err := tx.GetTopic(input.TopicId)
	if err != nil {
		return err
	}

	if topic == nil {
		return storage.ErrTopicNotFound
	}

	return nil
}

type stateRepository interface {
	GetDeleteProcessState(topicId string) (*models.DeleteProcess, error)
	SaveDeleteProcessState(state *models.DeleteProcess) error
}

func getOrCreateProcessState(repo stateRepository, input ProcessInput) (*models.DeleteProcess, error) {
	state, err := repo.GetDeleteProcessState(input.TopicId)
	if !errors.Is(err, storage.ErrTopicNotFound) && err != nil {
		return nil, err
	}

	if state != nil && !state.IsCompleted() {
		// is process is unfinished => continue
		return state, nil
	}

	state = models.NewDeleteProcess(input.TopicId)

	if err := repo.SaveDeleteProcessState(state); err != nil {
		return nil, err
	}

	return state, nil
}

func (p *process) getStepContext(ctx context.Context, tx models.Transaction, state *models.DeleteProcess) *StepContext {
	logger := p.logger
	topic := NewTopicService(ctx, p.confluent, tx)
	schema := NewSchemaService(ctx, p.confluent, tx)
	outbox := p.factory(tx)

	return NewStepContext(logger, state, topic, schema, outbox)
}

// region Steps

func ensureTopicIsDeleted(stepContext *StepContext) error {
	stepContext.logger.Trace("Running {Step}", "EnsureTopicIsDeleted")
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

// region Schema deletion

func ensureTopicSchemasAreDeleted(stepContext *StepContext) error {
	stepContext.logger.Trace("Running {Step}", "EnsureTopicSchemasAreDeleted")
	return ensureTopicSchemasAreDeletedStep(stepContext)
}

func ensureTopicSchemasAreDeletedStep(step EnsureTopicSchemasAreDeletedStep) error {
	if step.AreSchemasDeleted() {
		return nil
	}

	err := step.DeleteSchemasByTopicId()
	if err != nil {
		return err
	}

	step.MarkSchemasAsDeleted()
	return nil
}

type EnsureTopicSchemasAreDeletedStep interface {
	DeleteSchemasByTopicId() error
	MarkSchemasAsDeleted()
	AreSchemasDeleted() bool
}

// endregion
