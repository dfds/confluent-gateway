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

func (p *Process) service() *AccountService {
	return NewAccountService(p.Context, p.Confluent, p.Session)
}

// region Steps

func ensureServiceAccount(process *Process) error {
	capabilityRootId := process.State.CapabilityRootId
	clusterId := process.State.ClusterId
	service := process.service()

	fmt.Println("### EnsureServiceAccount")

	if process.State.HasServiceAccount {
		return nil
	}

	err := service.CreateServiceAccount(capabilityRootId, clusterId)
	if err != nil {
		return err
	}

	process.State.HasServiceAccount = true

	return err
}

func ensureServiceAccountAcl(process *Process) error {
	service := process.service()
	capabilityRootId := process.State.CapabilityRootId
	clusterId := process.State.ClusterId

	fmt.Println("### EnsureServiceAccountAcl")
	if process.State.HasClusterAccess {
		return nil
	}

	clusterAccess, err := service.GetOrCreateClusterAccess(capabilityRootId, clusterId)
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

		return service.CreateAclEntry(clusterId, clusterAccess.ServiceAccountId, &nextEntry)
	}
}

func ensureServiceAccountApiKey(process *Process) error {
	service := process.service()
	capabilityRootId := process.State.CapabilityRootId
	clusterId := process.State.ClusterId

	fmt.Println("### EnsureServiceAccountApiKey")
	if process.State.HasApiKey {
		return nil
	}

	clusterAccess, err := service.GetClusterAccess(capabilityRootId, clusterId)
	if err != nil {
		return err
	}

	err2 := service.CreateApiKey(clusterAccess)
	if err2 != nil {
		return err2
	}

	process.State.HasApiKey = true
	return nil
}

func ensureServiceAccountApiKeyAreStoredInVault(process *Process) error {
	service := process.service()
	aws := process.Vault
	capabilityRootId := process.State.CapabilityRootId
	clusterId := process.State.ClusterId

	fmt.Println("### EnsureServiceAccountApiKeyAreStoredInVault")
	if process.State.HasApiKeyInVault {
		return nil
	}

	clusterAccess, err := service.GetClusterAccess(capabilityRootId, clusterId)
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
