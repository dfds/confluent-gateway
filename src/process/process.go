package process

import (
	"context"
	"fmt"
	"github.com/dfds/confluent-gateway/models"
)

type Process struct {
	Context   context.Context
	Session   DataSession
	Confluent Confluent
	Vault     Vault
	State     *models.ProcessState
}

func (p *Process) CapabilityRootId() models.CapabilityRootId {
	return p.State.CapabilityRootId
}

func (p *Process) ClusterId() models.ClusterId {
	return p.State.ClusterId
}

func (p *Process) createTopic() error {
	topic := p.State.Topic()
	return p.Confluent.CreateTopic(p.Context, p.State.ClusterId, topic.Name, topic.Partitions, topic.Retention.Milliseconds())
}

func (p *Process) CreateServiceAccount() (*models.ServiceAccountId, error) {
	return p.Confluent.CreateServiceAccount(p.Context, "sa-some-name", "sa description")
}

func (p *Process) SaveServiceAccount(newServiceAccount *models.ServiceAccount) error {
	return p.Session.CreateServiceAccount(newServiceAccount)
}

func (p *Process) GetServiceAccount() (*models.ServiceAccount, error) {
	return p.Session.GetServiceAccount(p.State.CapabilityRootId)
}

func (p *Process) CreateClusterAccess(clusterAccess *models.ClusterAccess) error {
	return p.Session.CreateClusterAccess(clusterAccess)
}

func (p *Process) UpdateClusterAccess(access *models.ClusterAccess) error {
	return p.Session.UpdateClusterAccess(access)
}

func (p *Process) CreateAclEntry(serviceAccountId models.ServiceAccountId, entry *models.AclEntry) error {
	return p.Confluent.CreateACLEntry(p.Context, p.ClusterId(), serviceAccountId, entry.AclDefinition)
}

func (p *Process) UpdateAclEntry(entry *models.AclEntry) error {
	return p.Session.UpdateAclEntry(entry)
}

func (p *Process) CreateApiKey(clusterAccess *models.ClusterAccess) (*models.ApiKey, error) {
	return p.Confluent.CreateApiKey(p.Context, clusterAccess.ClusterId, clusterAccess.ServiceAccountId)
}

func NewProcess(ctx context.Context, session DataSession, confluent Confluent, vault Vault, state *models.ProcessState) *Process {
	return &Process{
		Context:   ctx,
		Session:   session,
		Confluent: confluent,
		Vault:     vault,
		State:     state,
	}
}

func (p *Process) NewSession(session DataSession) *Process {
	return &Process{
		Session:   session,
		Context:   p.Context,
		State:     p.State,
		Confluent: p.Confluent,
		Vault:     p.Vault,
	}
}

func (p *Process) Execute(step Step) error {
	return p.Session.Transaction(func(session DataSession) error {
		np := p.NewSession(session)

		err := step(np)
		if err != nil {
			return err
		}

		return session.UpdateProcessState(p.State)
	})
}

// region Steps

func ensureServiceAccount(process *Process) error {
	fmt.Println("### EnsureServiceAccount")

	if process.State.HasServiceAccount {
		return nil
	}

	err := NewAccountHelper(process).CreateServiceAccount()
	if err != nil {
		return err
	}

	process.State.HasServiceAccount = true

	return err
}

func ensureServiceAccountAcl(process *Process) error {
	fmt.Println("### EnsureServiceAccountAcl")
	if process.State.HasClusterAccess {
		return nil
	}

	clusterAccess, err := NewAccountHelper(process).GetOrCreateClusterAccess()
	if err != nil {
		return err
	}

	entries := clusterAccess.GetAclPendingCreation()
	if len(entries) == 0 {
		// no acl entries left => mark as done
		process.State.HasClusterAccess = true
		return nil

	} else {
		nextEntry := entries[0]

		return NewAccountHelper(process).CreateAclEntry(clusterAccess.ServiceAccountId, &nextEntry)
	}
}

func ensureServiceAccountApiKey(process *Process) error {

	fmt.Println("### EnsureServiceAccountApiKey")
	if process.State.HasApiKey {
		return nil
	}

	clusterAccess, err := NewAccountHelper(process).GetOrCreateClusterAccess()
	if err != nil {
		return err
	}

	err2 := NewAccountHelper(process).CreateApiKey(clusterAccess)
	if err2 != nil {
		return err2
	}

	process.State.HasApiKey = true
	return nil
}

func ensureServiceAccountApiKeyAreStoredInVault(process *Process) error {
	aws := process.Vault
	capabilityRootId := process.State.CapabilityRootId

	fmt.Println("### EnsureServiceAccountApiKeyAreStoredInVault")
	if process.State.HasApiKeyInVault {
		return nil
	}

	clusterAccess, err := NewAccountHelper(process).GetOrCreateClusterAccess()
	if err != nil {
		return err
	}

	if err := aws.StoreApiKey(context.TODO(), capabilityRootId, clusterAccess.ClusterId, clusterAccess.ApiKey); err != nil {
		return err
	}

	process.State.HasApiKeyInVault = true

	return nil
}

func ensureTopicIsCreated(process *Process) error {

	fmt.Println("### EnsureTopicIsCreated")
	if process.State.IsCompleted() {
		return nil
	}

	err := process.createTopic()
	if err != nil {
		return err
	}

	process.State.MarkAsCompleted()

	return nil
}

// endregion
