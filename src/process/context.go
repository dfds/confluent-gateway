package process

import (
	"github.com/dfds/confluent-gateway/models"
)

type StepContext struct {
	State   *models.ProcessState
	Account AccountService
	Vault   VaultService
	Topic   TopicService
	Outbox  Outbox
}

func NewStepContext(state *models.ProcessState, account AccountService, vault VaultService, topic TopicService, outbox Outbox) *StepContext {
	return &StepContext{State: state, Account: account, Vault: vault, Topic: topic, Outbox: outbox}
}

type Outbox interface {
	Produce(msg interface{}) error
}

func (p *StepContext) MarkServiceAccountAsReady() {
	p.State.HasServiceAccount = true
}

func (p *StepContext) CreateServiceAccount() error {
	return p.Account.CreateServiceAccount(p.State.CapabilityRootId, p.State.ClusterId)
}

func (p *StepContext) HasServiceAccount() bool {
	return p.State.HasServiceAccount
}

func (p *StepContext) HasClusterAccess() bool {
	return p.State.HasClusterAccess
}

func (p *StepContext) GetOrCreateClusterAccess() (*models.ClusterAccess, error) {
	return p.Account.GetOrCreateClusterAccess(p.State.CapabilityRootId, p.State.ClusterId)
}

func (p *StepContext) CreateAclEntry(clusterAccess *models.ClusterAccess, nextEntry models.AclEntry) error {
	return p.Account.CreateAclEntry(p.State.ClusterId, clusterAccess.ServiceAccountId, &nextEntry)
}

func (p *StepContext) MarkClusterAccessAsReady() {
	p.State.HasClusterAccess = true
}

func (p *StepContext) HasApiKey() bool {
	return p.State.HasApiKey
}

func (p *StepContext) GetClusterAccess() (*models.ClusterAccess, error) {
	return p.Account.GetClusterAccess(p.State.CapabilityRootId, p.State.ClusterId)
}

func (p *StepContext) CreateApiKey(clusterAccess *models.ClusterAccess) error {
	return p.Account.CreateApiKey(clusterAccess)
}

func (p *StepContext) MarkApiKeyAsReady() {
	p.State.HasApiKey = true
}

func (p *StepContext) HasApiKeyInVault() bool {
	return p.State.HasApiKeyInVault
}

func (p *StepContext) StoreApiKey(clusterAccess *models.ClusterAccess) error {
	return p.Vault.StoreApiKey(p.State.CapabilityRootId, clusterAccess)
}

func (p *StepContext) MarkApiKeyInVaultAsReady() {
	p.State.HasApiKeyInVault = true
}

func (p *StepContext) IsCompleted() bool {
	return p.State.IsCompleted()
}

func (p *StepContext) CreateTopic() error {
	return p.Topic.CreateTopic(p.State.ClusterId, p.State.Topic())
}

func (p *StepContext) MarkAsCompleted() {
	p.State.MarkAsCompleted()
}

func (p *StepContext) RaiseTopicProvisionedEvent() error {
	event := TopicProvisioned{
		CapabilityRootId: string(p.State.CapabilityRootId),
		ClusterId:        string(p.State.ClusterId),
		TopicName:        p.State.TopicName,
	}
	return p.Outbox.Produce(event)
}
