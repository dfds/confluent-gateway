package create

import (
	"github.com/dfds/confluent-gateway/logging"
	"github.com/dfds/confluent-gateway/messaging"
	"github.com/dfds/confluent-gateway/models"
)

type StepContext struct {
	logger  logging.Logger
	state   *models.ProcessState
	account AccountService
	vault   VaultService
	topic   TopicService
	outbox  Outbox
}

func NewStepContext(logger logging.Logger, state *models.ProcessState, account AccountService, vault VaultService, topic TopicService, outbox Outbox) *StepContext {
	return &StepContext{logger: logger, state: state, account: account, vault: vault, topic: topic, outbox: outbox}
}

type AccountService interface {
	CreateServiceAccount(models.CapabilityRootId, models.ClusterId) error
	GetOrCreateClusterAccess(models.CapabilityRootId, models.ClusterId) (*models.ClusterAccess, error)
	GetClusterAccess(models.CapabilityRootId, models.ClusterId) (*models.ClusterAccess, error)
	CreateAclEntry(models.ClusterId, models.UserAccountId, *models.AclEntry) error
	CreateApiKey(*models.ClusterAccess) error
}

type VaultService interface {
	StoreApiKey(models.CapabilityRootId, *models.ClusterAccess) error
}

type TopicService interface {
	CreateTopic(models.CapabilityRootId, models.ClusterId, models.TopicDescription) error
}

type Outbox interface {
	Produce(msg messaging.OutgoingMessage) error
}

func (c *StepContext) MarkServiceAccountAsReady() {
	c.state.HasServiceAccount = true
}

func (c *StepContext) CreateServiceAccount() error {
	return c.account.CreateServiceAccount(c.state.CapabilityRootId, c.state.ClusterId)
}

func (c *StepContext) HasServiceAccount() bool {
	return c.state.HasServiceAccount
}

func (c *StepContext) HasClusterAccess() bool {
	return c.state.HasClusterAccess
}

func (c *StepContext) GetOrCreateClusterAccess() (*models.ClusterAccess, error) {
	return c.account.GetOrCreateClusterAccess(c.state.CapabilityRootId, c.state.ClusterId)
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
	return c.account.GetClusterAccess(c.state.CapabilityRootId, c.state.ClusterId)
}

func (c *StepContext) CreateApiKey(clusterAccess *models.ClusterAccess) error {
	return c.account.CreateApiKey(clusterAccess)
}

func (c *StepContext) MarkApiKeyAsReady() {
	c.state.HasApiKey = true
}

func (c *StepContext) HasApiKeyInVault() bool {
	return c.state.HasApiKeyInVault
}

func (c *StepContext) StoreApiKey(clusterAccess *models.ClusterAccess) error {
	return c.vault.StoreApiKey(c.state.CapabilityRootId, clusterAccess)
}

func (c *StepContext) MarkApiKeyInVaultAsReady() {
	c.state.HasApiKeyInVault = true
}

func (c *StepContext) IsCompleted() bool {
	return c.state.IsCompleted()
}

func (c *StepContext) CreateTopic() error {
	return c.topic.CreateTopic(c.state.CapabilityRootId, c.state.ClusterId, c.state.TopicDescription())
}

func (c *StepContext) MarkAsCompleted() {
	c.state.MarkAsCompleted()
}

func (c *StepContext) RaiseTopicProvisionedEvent() error {
	event := &TopicProvisioned{
		partitionKey:     c.state.Id.String(),
		CapabilityRootId: string(c.state.CapabilityRootId),
		ClusterId:        string(c.state.ClusterId),
		TopicName:        c.state.TopicName,
	}
	return c.outbox.Produce(event)
}
