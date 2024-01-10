package schema

import (
	"context"
	"errors"
	"fmt"
	"github.com/dfds/confluent-gateway/internal/confluent"
	"github.com/dfds/confluent-gateway/internal/models"
	. "github.com/dfds/confluent-gateway/internal/process"
	"github.com/dfds/confluent-gateway/internal/storage"
	"github.com/dfds/confluent-gateway/internal/vault"
	"github.com/dfds/confluent-gateway/logging"
)

type logger interface {
	LogDebug(string, ...string)
	LogWarning(string, ...string)
	LogError(error, string, ...string)
}

type process struct {
	logger    logging.Logger
	database  models.Database
	confluent Confluent
	factory   OutboxFactory
	vault     vault.Vault
}

func NewProcess(logger logging.Logger, database models.Database, confluent Confluent, vault vault.Vault, factory OutboxFactory) Process {
	return &process{
		logger:    logger,
		database:  database,
		factory:   factory,
		confluent: confluent,
		vault:     vault,
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
		Step(ensureServiceAccountSchemaRegistryAccessStep).
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
	newAccountService := NewSchemaAccountService(ctx, p.confluent, tx)
	vaultService := NewVaultService(ctx, p.vault)
	topicService := NewTopicService(tx)
	return NewStepContext(p.logger, ctx, schema, p.confluent, p.factory(tx), newAccountService, vaultService, topicService)
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

type EnsureServiceAccountSchemaRegistryAccessStep interface {
	logger
	GetClusterAccess() (*models.ClusterAccess, error)
	HasSchemaRegistryApiKey(clusterAccess *models.ClusterAccess) (bool, error)
	HasSchemaRegistryApiKeyInVault(clusterAccess *models.ClusterAccess) (bool, error)
	CreateServiceAccountRoleBinding(*models.ClusterAccess) error
	CreateSchemaRegistryApiKeyAndStoreInVault(clusterAccess *models.ClusterAccess, shouldOverwriteKey bool) error
	DeleteSchemaRegistryApiKey(clusterAccess *models.ClusterAccess) error
}

func ensureServiceAccountSchemaRegistryAccessStep(step *StepContext) error {
	inner := func(step EnsureServiceAccountSchemaRegistryAccessStep) error {
		step.LogDebug("Running {Step}", "EnsureServiceAccountSchemaRegistryAccessStep")

		clusterAccess, err := step.GetClusterAccess()
		if err != nil {
			if !errors.Is(err, storage.ErrTopicNotFound) {
				step.LogError(err, "unable to get cluster access")
			}
			return nil
		}

		err = step.CreateServiceAccountRoleBinding(clusterAccess)
		if err != nil {
			if errors.Is(err, confluent.ErrMissingSchemaRegistryIds) {
				step.LogError(err, "unable to setup schema registry access")
				return nil // fail silently to not take down whole service
			}
			return err
		}

		hasKeyInConfluent, err := step.HasSchemaRegistryApiKey(clusterAccess)
		if err != nil {
			return err
		}
		hasKeyInVault, err := step.HasSchemaRegistryApiKeyInVault(clusterAccess)
		if err != nil {
			return err
		}

		if hasKeyInVault && hasKeyInConfluent {
			return nil
		}

		recreateKey := false
		if hasKeyInConfluent && !hasKeyInVault {
			step.LogWarning("found existing api key in Confluent, but not in Parameter Store. Deleting key and creating again.")
			err = step.DeleteSchemaRegistryApiKey(clusterAccess)
			if err != nil {
				return err
			}
		} else if !hasKeyInConfluent && hasKeyInVault { // not sure if this can happen
			step.LogWarning("found existing key in Parameter Store, but not in Confluent. Creating new key and updating Parameter Store.")
			recreateKey = true
		}

		err = step.CreateSchemaRegistryApiKeyAndStoreInVault(clusterAccess, recreateKey)
		if err != nil {
			return err
		}
		step.LogWarning("granted schema registry access through schema registration")
		return nil
	}
	return inner(step)
}

// endregion
