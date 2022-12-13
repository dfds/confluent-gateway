package models

import (
	"context"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"time"
)

//region ProcessState

type ProcessRepository interface {
	CreateProcessState(state *ProcessState) error
	UpdateProcessState(state *ProcessState) error
	GetProcessState(capabilityRootId CapabilityRootId, clusterId ClusterId, topicName string) (*ProcessState, error)
}

//endregion

type TopicCreationProcess struct {
	data   Database
	client ConfluentClient
	aws    VaultClient
}

func NewTopicCreationProcess(data Database, client ConfluentClient, aws VaultClient) *TopicCreationProcess {
	return &TopicCreationProcess{data, client, aws}
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

func (tcp *TopicCreationProcess) prepareProcess(ctx context.Context, request NewTopicHasBeenRequested) (*process, error) {
	capabilityRootId := CapabilityRootId(request.CapabilityRootId)
	clusterId := ClusterId(request.ClusterId)
	topic := NewTopic(request.TopicName, request.Partitions, request.Retention)

	session := tcp.data.NewSession(ctx)

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

		state = &ProcessState{
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

	return &process{
		Session: session,
		State:   state,
		Client:  tcp.client,
		Aws:     tcp.aws,
	}, nil
}

// region process

type process struct {
	Context context.Context
	Session DataSession
	State   *ProcessState
	Client  ConfluentClient
	Aws     VaultClient
}

func (p *process) NewSession(session DataSession) *process {
	return &process{
		Session: session,
		Context: p.Context,
		State:   p.State,
		Client:  p.Client,
		Aws:     p.Aws,
	}
}

func (p *process) Execute(step Step) error {
	return p.Session.Transaction(func(session DataSession) error {
		np := p.NewSession(session)

		err := step(np)
		if err != nil {
			return err
		}

		return session.Processes().UpdateProcessState(p.State)
	})
}

func (p *process) service() *Service {
	client := p.Client
	repository := p.Session.ServiceAccounts()
	service := NewService(client, repository)
	return service
}

// endregion

// region Steps

func ensureServiceAccount(process *process) error {
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

func ensureServiceAccountAcl(process *process) error {
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

func ensureServiceAccountApiKey(process *process) error {
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

func ensureServiceAccountApiKeyAreStoredInVault(process *process) error {
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

func ensureTopicIsCreated(process *process) error {
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
