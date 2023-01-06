package process

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
	CreateServiceAccount(capabilityRootId models.CapabilityRootId, clusterId models.ClusterId) error
	GetOrCreateClusterAccess(capabilityRootId models.CapabilityRootId, clusterId models.ClusterId) (*models.ClusterAccess, error)
	GetClusterAccess(capabilityRootId models.CapabilityRootId, clusterId models.ClusterId) (*models.ClusterAccess, error)
	CreateAclEntry(clusterId models.ClusterId, userAccountId models.UserAccountId, entry *models.AclEntry) error
	CreateApiKey(clusterAccess *models.ClusterAccess) error
}

type VaultService interface {
	StoreApiKey(capabilityRootId models.CapabilityRootId, clusterAccess *models.ClusterAccess) error
}

type TopicService interface {
	CreateTopic(clusterId models.ClusterId, topic models.TopicDescription) error
}

type Outbox interface {
	Produce(msg messaging.OutgoingMessage) error
}

func (p *StepContext) MarkServiceAccountAsReady() {
	p.state.HasServiceAccount = true
}

func (p *StepContext) CreateServiceAccount() error {
	return p.account.CreateServiceAccount(p.state.CapabilityRootId, p.state.ClusterId)
}

func (p *StepContext) HasServiceAccount() bool {
	return p.state.HasServiceAccount
}

func (p *StepContext) HasClusterAccess() bool {
	return p.state.HasClusterAccess
}

func (p *StepContext) GetOrCreateClusterAccess() (*models.ClusterAccess, error) {
	return p.account.GetOrCreateClusterAccess(p.state.CapabilityRootId, p.state.ClusterId)
}

func (p *StepContext) CreateAclEntry(clusterAccess *models.ClusterAccess, nextEntry models.AclEntry) error {
	return p.account.CreateAclEntry(p.state.ClusterId, clusterAccess.UserAccountId, &nextEntry)
}

func (p *StepContext) MarkClusterAccessAsReady() {
	p.state.HasClusterAccess = true
}

func (p *StepContext) HasApiKey() bool {
	return p.state.HasApiKey
}

func (p *StepContext) GetClusterAccess() (*models.ClusterAccess, error) {
	return p.account.GetClusterAccess(p.state.CapabilityRootId, p.state.ClusterId)
}

func (p *StepContext) CreateApiKey(clusterAccess *models.ClusterAccess) error {
	return p.account.CreateApiKey(clusterAccess)
}

func (p *StepContext) MarkApiKeyAsReady() {
	p.state.HasApiKey = true
}

func (p *StepContext) HasApiKeyInVault() bool {
	return p.state.HasApiKeyInVault
}

func (p *StepContext) StoreApiKey(clusterAccess *models.ClusterAccess) error {
	return p.vault.StoreApiKey(p.state.CapabilityRootId, clusterAccess)
}

func (p *StepContext) MarkApiKeyInVaultAsReady() {
	p.state.HasApiKeyInVault = true
}

func (p *StepContext) IsCompleted() bool {
	return p.state.IsCompleted()
}

func (p *StepContext) CreateTopic() error {
	return p.topic.CreateTopic(p.state.ClusterId, p.state.TopicDescription())
}

func (p *StepContext) MarkAsCompleted() {
	p.state.MarkAsCompleted()
}

func (p *StepContext) RaiseTopicProvisionedEvent() error {
	event := &TopicProvisioned{
		partitionKey:     p.state.Id.String(),
		CapabilityRootId: string(p.state.CapabilityRootId),
		ClusterId:        string(p.state.ClusterId),
		TopicName:        p.state.TopicName,
	}
	return p.outbox.Produce(event)
}
