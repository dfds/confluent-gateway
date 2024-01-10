package schema

import (
	"context"
	"fmt"
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/logging"
	"github.com/dfds/confluent-gateway/messaging"
)

type StepContext struct {
	logger   logging.Logger
	ctx      context.Context
	account  AccountService
	vault    VaultService
	state    *models.SchemaProcess
	registry SchemaRegistry
	input    ProcessInput
	outbox   Outbox
	topic    TopicService
}
type AccountService interface {
	GetServiceAccount(models.CapabilityId) (*models.ServiceAccount, error)
	GetClusterAccess(models.CapabilityId, models.ClusterId) (*models.ClusterAccess, error)
	CreateSchemaRegistryApiKey(clusterAccess *models.ClusterAccess) (models.ApiKey, error)
	CreateServiceAccountRoleBinding(clusterAccess *models.ClusterAccess) error
	CountSchemaRegistryApiKeys(clusterAccess *models.ClusterAccess) (int, error)
	DeleteSchemaRegistryApiKey(clusterAccess *models.ClusterAccess) error
}

func NewStepContext(logger logging.Logger, ctx context.Context, schema *models.SchemaProcess, registry SchemaRegistry, outbox Outbox, account AccountService, vault VaultService, topic TopicService) *StepContext {
	return &StepContext{logger: logger, ctx: ctx, state: schema, registry: registry, outbox: outbox, account: account, vault: vault, topic: topic}
}

type TopicService interface {
	GetTopic(string) (*models.Topic, error)
}

type Outbox interface {
	Produce(msg messaging.OutgoingMessage) error
}

type OutboxRepository interface {
	AddToOutbox(entry *messaging.OutboxEntry) error
}

type OutboxFactory func(repository OutboxRepository) Outbox

func (c *StepContext) IsCompleted() bool {
	return c.state.IsCompleted()
}

func (c *StepContext) RegisterSchema() error {
	return c.registry.RegisterSchema(c.ctx, c.state.ClusterId, c.state.Subject, c.state.Schema)
}

func (c *StepContext) MarkAsCompleted() {
	c.state.MarkAsCompleted()
}

func (c *StepContext) RaiseSchemaRegisteredEvent() error {
	event := &SchemaRegistered{
		MessageContractId: c.state.MessageContractId,
	}
	return c.outbox.Produce(event)
}

func (c *StepContext) RaiseSchemaRegistrationFailed(reason string) error {
	event := &SchemaRegistrationFailed{
		MessageContractId: c.state.MessageContractId,
		Reason:            reason,
	}
	return c.outbox.Produce(event)
}

func (c *StepContext) LogDebug(format string, args ...string) {
	c.logger.Debug(format, args...)
}

func (c *StepContext) LogError(err error, format string, args ...string) {
	c.logger.Error(err, format, args...)
}

func (c *StepContext) LogWarning(format string, args ...string) {
	c.logger.Warning(format, args...)
}

func (c *StepContext) HasServiceAccount(capabilityId models.CapabilityId) bool {
	account, err := c.account.GetServiceAccount(capabilityId)
	if err != nil {
		c.LogError(err, fmt.Sprintf("encountered error when checking if ServiceAccount exists for CapabilityId %s", capabilityId))
		return false
	}
	return account != nil
}

func (c *StepContext) GetClusterAccess() (*models.ClusterAccess, error) {
	topic, err := c.topic.GetTopic(c.state.TopicId)
	if err != nil {
		return nil, err
	}
	return c.account.GetClusterAccess(topic.CapabilityId, topic.ClusterId)
}

func (c *StepContext) HasSchemaRegistryApiKey(clusterAccess *models.ClusterAccess) (bool, error) {
	count, err := c.account.CountSchemaRegistryApiKeys(clusterAccess)
	return count > 0, err
}

func (c *StepContext) HasSchemaRegistryApiKeyInVault(clusterAccess *models.ClusterAccess) (bool, error) {
	topic, err := c.topic.GetTopic(c.state.TopicId)
	if err != nil {
		return false, err
	}
	return c.vault.QuerySchemaRegistryApiKey(topic.CapabilityId, clusterAccess.ClusterId)
}

func (c *StepContext) CreateSchemaRegistryApiKeyAndStoreInVault(clusterAccess *models.ClusterAccess, shouldOverwriteKey bool) error {
	newKey, err := c.account.CreateSchemaRegistryApiKey(clusterAccess)
	if err != nil {
		return err
	}
	topic, err := c.topic.GetTopic(c.state.TopicId)
	if err != nil {
		return err
	}
	return c.vault.StoreSchemaRegistryApiKey(topic.CapabilityId, clusterAccess.ClusterId, newKey, shouldOverwriteKey)
}

func (c *StepContext) DeleteSchemaRegistryApiKey(clusterAccess *models.ClusterAccess) error {
	return c.account.DeleteSchemaRegistryApiKey(clusterAccess)
}

func (c *StepContext) CreateServiceAccountRoleBinding(clusterAccess *models.ClusterAccess) error {
	return c.account.CreateServiceAccountRoleBinding(clusterAccess)
}
