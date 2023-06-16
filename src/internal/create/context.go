package create

import (
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/logging"
	"github.com/dfds/confluent-gateway/messaging"
)

type StepContext struct {
	logger  logging.Logger
	state   *models.CreateProcess
	account AccountService
	vault   VaultService
	topic   TopicService
	outbox  Outbox
}

func NewStepContext(logger logging.Logger, state *models.CreateProcess, account AccountService, vault VaultService, topic TopicService, outbox Outbox) *StepContext {
	return &StepContext{logger: logger, state: state, account: account, vault: vault, topic: topic, outbox: outbox}
}

type AccountService interface {
	CreateServiceAccount(models.CapabilityId, models.ClusterId) error
	GetOrCreateClusterAccess(models.CapabilityId, models.ClusterId) (*models.ClusterAccess, error)
	GetClusterAccess(models.CapabilityId, models.ClusterId) (*models.ClusterAccess, error)
	CreateAclEntry(models.ClusterId, models.UserAccountId, *models.AclEntry) error
	CreateClusterApiKey(*models.ClusterAccess) error
}

type VaultService interface {
	StoreApiKey(models.CapabilityId, *models.ClusterAccess) error
}

type TopicService interface {
	CreateTopic(models.CapabilityId, models.ClusterId, string, models.TopicDescription) error
}

type Outbox interface {
	Produce(msg messaging.OutgoingMessage) error
}

type OutboxRepository interface {
	AddToOutbox(entry *messaging.OutboxEntry) error
}

type OutboxFactory func(repository OutboxRepository) Outbox

func (c *StepContext) MarkServiceAccountAsReady() {
	c.state.HasServiceAccount = true
}

func (c *StepContext) CreateServiceAccount() error {
	return c.account.CreateServiceAccount(c.state.CapabilityId, c.state.ClusterId)
}

func (c *StepContext) HasServiceAccount() bool {
	return c.state.HasServiceAccount
}

func (c *StepContext) HasClusterAccess() bool {
	return c.state.HasClusterAccess
}

func (c *StepContext) GetOrCreateClusterAccess() (*models.ClusterAccess, error) {
	return c.account.GetOrCreateClusterAccess(c.state.CapabilityId, c.state.ClusterId)
}

func (c *StepContext) CreateAclEntry(clusterAccess *models.ClusterAccess, nextEntry models.AclEntry) error {
	return c.account.CreateAclEntry(c.state.ClusterId, clusterAccess.UserAccountId, &nextEntry)
}

func (c *StepContext) MarkClusterAccessAsReady() {
	c.state.HasClusterAccess = true
}

func (c *StepContext) HasApiKey() bool {
	return c.state.HasApiKey
}

func (c *StepContext) GetClusterAccess() (*models.ClusterAccess, error) {
	return c.account.GetClusterAccess(c.state.CapabilityId, c.state.ClusterId)
}

func (c *StepContext) CreateApiKey(clusterAccess *models.ClusterAccess) error {
	return c.account.CreateClusterApiKey(clusterAccess)
}

func (c *StepContext) MarkApiKeyAsReady() {
	c.state.HasApiKey = true
}

func (c *StepContext) HasApiKeyInVault() bool {
	return c.state.HasApiKeyInVault
}

func (c *StepContext) StoreApiKey(clusterAccess *models.ClusterAccess) error {
	return c.vault.StoreApiKey(c.state.CapabilityId, clusterAccess)
}

func (c *StepContext) MarkApiKeyInVaultAsReady() {
	c.state.HasApiKeyInVault = true
}

func (c *StepContext) IsCompleted() bool {
	return c.state.IsCompleted()
}

func (c *StepContext) CreateTopic() error {
	return c.topic.CreateTopic(c.state.CapabilityId, c.state.ClusterId, c.state.TopicId, c.state.TopicDescription())
}

func (c *StepContext) MarkAsCompleted() {
	c.state.MarkAsCompleted()
}

func (c *StepContext) RaiseTopicProvisionedEvent() error {
	event := &TopicProvisioned{
		TopicId:      c.state.TopicId,
		CapabilityId: string(c.state.CapabilityId),
		ClusterId:    string(c.state.ClusterId),
		TopicName:    c.state.TopicName,
	}
	return c.outbox.Produce(event)
}
