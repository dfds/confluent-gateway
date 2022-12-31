package process

import (
	"github.com/dfds/confluent-gateway/models"
)

type Process struct {
	State   *models.ProcessState
	Account AccountService
	Vault   VaultService
	Topic   TopicService
	Outbox  Outbox
}

func NewProcess(state *models.ProcessState, account AccountService, vault VaultService, topic TopicService, outbox Outbox) *Process {
	return &Process{State: state, Account: account, Vault: vault, Topic: topic, Outbox: outbox}
}

type Outbox interface {
	Produce(msg interface{}) error
}

func (p *Process) markServiceAccountReady() {
	p.State.HasServiceAccount = true
}

func (p *Process) createServiceAccount() error {
	return p.Account.CreateServiceAccount(p.State.CapabilityRootId, p.State.ClusterId)
}

func (p *Process) hasServiceAccount() bool {
	return p.State.HasServiceAccount
}

func (p *Process) hasClusterAccess() bool {
	return p.State.HasClusterAccess
}

func (p *Process) getOrCreateClusterAccess() (*models.ClusterAccess, error) {
	return p.Account.GetOrCreateClusterAccess(p.State.CapabilityRootId, p.State.ClusterId)
}

func (p *Process) createAclEntry(clusterAccess *models.ClusterAccess, nextEntry models.AclEntry) error {
	return p.Account.CreateAclEntry(p.State.ClusterId, clusterAccess.ServiceAccountId, &nextEntry)
}

func (p *Process) markClusterAccessReady() {
	p.State.HasClusterAccess = true
}

func (p *Process) hasApiKey() bool {
	return p.State.HasApiKey
}

func (p *Process) getClusterAccess() (*models.ClusterAccess, error) {
	return p.Account.GetClusterAccess(p.State.CapabilityRootId, p.State.ClusterId)
}

func (p *Process) createApiKey(clusterAccess *models.ClusterAccess) error {
	return p.Account.CreateApiKey(clusterAccess)
}

func (p *Process) markApiKeyReady() {
	p.State.HasApiKey = true
}

func (p *Process) hasApiKeyInVault() bool {
	return p.State.HasApiKeyInVault
}

func (p *Process) storeApiKey(clusterAccess *models.ClusterAccess) error {
	return p.Vault.StoreApiKey(p.State.CapabilityRootId, clusterAccess)
}

func (p *Process) markApiKeyInVaultReady() {
	p.State.HasApiKeyInVault = true
}

func (p *Process) isCompleted() bool {
	return p.State.IsCompleted()
}

func (p *Process) createTopic() error {
	return p.Topic.CreateTopic(p.State.ClusterId, p.State.Topic())
}

func (p *Process) markAsCompleted() {
	p.State.MarkAsCompleted()
}

func (p *Process) topicProvisioned() error {
	event := TopicProvisioned{
		CapabilityRootId: string(p.State.CapabilityRootId),
		ClusterId:        string(p.State.ClusterId),
		TopicName:        p.State.TopicName,
	}
	return p.Outbox.Produce(event)
}
