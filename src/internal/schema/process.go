package schema

import (
	"context"
	"errors"
	"fmt"
	"github.com/dfds/confluent-gateway/internal/confluent"
	"github.com/dfds/confluent-gateway/internal/models"
	. "github.com/dfds/confluent-gateway/internal/process"
	"github.com/dfds/confluent-gateway/internal/storage"
	"github.com/dfds/confluent-gateway/logging"
)

type process struct {
	logger   logging.Logger
	database models.Database
	registry SchemaRegistry
	factory  OutboxFactory
}

func NewProcess(logger logging.Logger, database models.Database, registry SchemaRegistry, factory OutboxFactory) Process {
	return &process{
		logger:   logger,
		database: database,
		registry: registry,
		factory:  factory,
	}
}

type ProcessInput struct {
	MessageContractId string
	TopicId           string
	MessageType       string
	Description       string
	Schema            string
	SchemaVersion     int32
}

func (p *process) Process(ctx context.Context, input ProcessInput) error {
	session := p.database.NewSession(ctx)

	state, err := p.prepareProcessState(session, input)
	if err != nil {
		if errors.Is(err, storage.ErrTopicNotFound) { //TODO: What? Why do we just ignore this?
			// topic doesn't exists => skip
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
		Step(ensureSchemaIsRegistered).
		Run(func(step func(*StepContext) error) error {
			return session.Transaction(func(tx models.Transaction) error {
				stepContext := p.getStepContext(ctx, tx, state)

				err := step(stepContext)
				if err != nil {
					return err
				}

				return tx.UpdateSchemaProcessState(state)
			})
		})
}

func (p *process) prepareProcessState(session models.Session, input ProcessInput) (*models.SchemaProcess, error) {
	var s *models.SchemaProcess

	err := session.Transaction(func(tx models.Transaction) error {
		topic, err := tx.GetTopic(input.TopicId)
		if err != nil {
			return err
		}

		if topic == nil {
			return storage.ErrTopicNotFound
		}

		state, err := getOrCreateProcessState(tx, input, topic)
		if err != nil {
			return err
		}

		s = state

		return nil
	})

	return s, err
}

type schemaRepository interface {
	GetSchemaProcessState(messageContractId string) (*models.SchemaProcess, error)
	SaveSchemaProcessState(state *models.SchemaProcess) error
}

func getOrCreateProcessState(repo schemaRepository, input ProcessInput, topic *models.Topic) (*models.SchemaProcess, error) {
	schema, err := repo.GetSchemaProcessState(input.MessageContractId)
	if err != nil {
		return nil, err
	}

	if schema != nil && !schema.IsCompleted() {
		// is process is unfinished => continue
		return schema, nil
	}

	subject := fmt.Sprintf("%s-%s", topic.Name, input.MessageType)
	schema = models.NewSchemaProcess(topic.ClusterId, input.MessageContractId, input.TopicId, input.MessageType, input.Description, subject, input.Schema, input.SchemaVersion)

	if err := repo.SaveSchemaProcessState(schema); err != nil {
		return nil, err
	}

	return schema, nil
}

func (p *process) getStepContext(ctx context.Context, tx models.Transaction, schema *models.SchemaProcess) *StepContext {
	return NewStepContext(p.logger, ctx, schema, p.registry, p.factory(tx))
}

// region Steps

func ensureSchemaIsRegistered(stepContext *StepContext) error {
	stepContext.logger.Trace("Running {Step}", "EnsureSchemaIsRegistered")
	return ensureSchemaIsRegisteredStep(stepContext)
}

type EnsureSchemaIsRegisteredStep interface {
	IsCompleted() bool
	RegisterSchema() error
	MarkAsCompleted()
	RaiseSchemaRegisteredEvent() error
	RaiseSchemaRegistrationFailed(string) error
}

func ensureSchemaIsRegisteredStep(step EnsureSchemaIsRegisteredStep) error {
	if step.IsCompleted() {
		return nil
	}

	err := step.RegisterSchema()

	if errors.Is(err, confluent.ErrNoSchemaRegistry) {
		step.MarkAsCompleted()
		return step.RaiseSchemaRegistrationFailed(err.Error())
	}

	var mr *confluent.ClientError
	if errors.As(err, &mr) {
		step.MarkAsCompleted()
		return step.RaiseSchemaRegistrationFailed(mr.Error())
	}

	if err != nil {
		return err
	}

	step.MarkAsCompleted()

	return step.RaiseSchemaRegisteredEvent()
}

// endregion
