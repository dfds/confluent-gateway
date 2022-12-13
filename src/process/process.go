package process

import (
	"context"
	"fmt"
	"github.com/dfds/confluent-gateway/models"
	uuid "github.com/satori/go.uuid"
	"time"
)

type TopicCreationProcess struct {
	database  Database
	confluent Confluent
	vault     Vault
}

func NewTopicCreationProcess(database Database, confluent Confluent, vault Vault) *TopicCreationProcess {
	return &TopicCreationProcess{database, confluent, vault}
}

type NewTopicHasBeenRequested struct {
	CapabilityRootId string // example => logistics-somecapability-abcd
	ClusterId        string
	TopicName        string // full name => pub.logistics-somecapability-abcd.foo
	Partitions       int
	Retention        int // in ms
}

func (tcp *TopicCreationProcess) ProcessLogic(ctx context.Context, request NewTopicHasBeenRequested) error {
	p, err := tcp.prepareProcess(ctx, request)
	if err != nil {
		return err
	}

	if p.State.IsCompleted() {
		// already completed => skip
		return nil
	}

	return PrepareSteps().
		Step(ensureServiceAccount).
		Step(ensureServiceAccountAcl).Until(func() bool { return p.State.HasClusterAccess }).
		Step(ensureServiceAccountApiKey).
		Step(ensureServiceAccountApiKeyAreStoredInVault).
		Step(ensureTopicIsCreated).
		Run(p)
}

func (tcp *TopicCreationProcess) prepareProcess(ctx context.Context, request NewTopicHasBeenRequested) (*Process, error) {
	capabilityRootId := models.CapabilityRootId(request.CapabilityRootId)
	clusterId := models.ClusterId(request.ClusterId)
	topic := models.NewTopic(request.TopicName, request.Partitions, request.Retention)

	session := tcp.database.NewSession(ctx)

	state, err := session.Processes().GetProcessState(capabilityRootId, clusterId, topic.Name)
	if err != nil {
		return nil, err
	}

	if state == nil {
		serviceAccount, err := session.ServiceAccounts().GetServiceAccount(capabilityRootId)
		if err != nil {
			return nil, err
		}

		hasServiceAccount := false
		hasClusterAccess := false

		if serviceAccount != nil {
			hasServiceAccount = true
			_, hasClusterAccess = serviceAccount.TryGetClusterAccess(clusterId)
		}

		state = &models.ProcessState{
			Id:                uuid.NewV4(),
			CapabilityRootId:  capabilityRootId,
			ClusterId:         clusterId,
			Topic:             topic,
			HasServiceAccount: hasServiceAccount,
			HasClusterAccess:  hasClusterAccess,
			HasApiKey:         hasClusterAccess,
			HasApiKeyInVault:  hasClusterAccess,
			CreatedAt:         time.Now(),
			CompletedAt:       nil,
		}

		if err := session.Processes().CreateProcessState(state); err != nil {
			return nil, err
		}
	}

	return &Process{
		Session: session,
		State:   state,
		Client:  tcp.confluent,
		Aws:     tcp.vault,
	}, nil
}

// region Process

type Process struct {
	Context context.Context
	Session DataSession
	State   *models.ProcessState
	Client  Confluent
	Aws     Vault
}

func (p *Process) NewSession(session DataSession) *Process {
	return &Process{
		Session: session,
		Context: p.Context,
		State:   p.State,
		Client:  p.Client,
		Aws:     p.Aws,
	}
}

func (p *Process) Execute(step Step) error {
	return p.Session.Transaction(func(session DataSession) error {
		np := p.NewSession(session)

		err := step(np)
		if err != nil {
			return err
		}

		return session.Processes().UpdateProcessState(p.State)
	})
}

func (p *Process) service() *Service {
	client := p.Client
	repository := p.Session.ServiceAccounts()
	service := NewService(client, repository)
	return service
}

// endregion

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

		return service.CreateAclEntry(clusterId, clusterAccess, &nextEntry)
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

	clusterAccess, err := service.GetOrCreateClusterAccess(capabilityRootId, clusterId)
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
	aws := process.Aws
	capabilityRootId := process.State.CapabilityRootId
	clusterId := process.State.ClusterId

	fmt.Println("### EnsureServiceAccountApiKeyAreStoredInVault")
	if process.State.HasApiKeyInVault {
		return nil
	}

	clusterAccess, err := service.GetOrCreateClusterAccess(capabilityRootId, clusterId)
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
	service := process.service()
	clusterId := process.State.ClusterId
	topic := process.State.Topic

	fmt.Println("### EnsureTopicIsCreated")
	if process.State.IsCompleted() {
		return nil
	}

	err := service.CreateTopic(clusterId, topic)
	if err != nil {
		return err
	}

	process.State.MarkAsCompleted()

	return nil
}

// endregion
