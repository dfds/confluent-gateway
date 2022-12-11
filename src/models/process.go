package models

import (
	"context"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"time"
)

//region ProcessState

type ProcessState struct {
	Id                uuid.UUID `gorm:"type:uuid;primarykey"`
	CapabilityRootId  CapabilityRootId
	ClusterId         ClusterId
	Topic             Topic `gorm:"embedded;embeddedPrefix:topic_"`
	HasServiceAccount bool
	HasClusterAccess  bool
	HasApiKey         bool
	HasApiKeyInVault  bool
	CreatedAt         time.Time
	CompletedAt       *time.Time
}

func (*ProcessState) TableName() string {
	return "process"
}

func (p *ProcessState) IsCompleted() bool {
	return p.CompletedAt != nil
}

func (p *ProcessState) MarkAsCompleted() {
	if p.IsCompleted() {
		return
	}

	now := time.Now()
	p.CompletedAt = &now
}

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

	//1. Ensure capability has cluster access
	//  1.2. Ensure capability has service account
	if err := p.execute(ensureServiceAccount); err != nil {
		return err
	}
	//	1.3. Ensure service account has all acls
	if err := p.executeWhile(ensureServiceAccountAcl, p.HasPendingClusterAccess); err != nil {
		return err
	}
	//	1.4. Ensure service account has api keys
	if err := p.execute(ensureServiceAccountApiKey); err != nil {
		return err
	}
	//	1.5. Ensure api keys are stored in vault
	if err := p.execute(ensureServiceAccountApiKeyAreStoredInVault); err != nil {
		return err
	}
	//2. Ensure topic is created
	if err := p.execute(ensureTopicIsCreated); err != nil {
		return err
	}

	return nil
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

type process struct {
	Session DataSession
	State   *ProcessState
	Client  ConfluentClient
	Aws     VaultClient
}

func (p *process) HasPendingClusterAccess() bool {
	return !p.State.HasClusterAccess
}

func (p *process) execute(stepFunc func(*process) error) error {
	return p.Session.Transaction(func(session DataSession) error {
		process := &process{session, p.State, p.Client, p.Aws}

		if err := stepFunc(process); err != nil {
			return err
		}

		return session.Processes().UpdateProcessState(p.State)
	})
}

func (p *process) executeWhile(stepFunc func(*process) error, predicate func() bool) error {
	for predicate() {
		if err := p.execute(stepFunc); err != nil {
			return err
		}
	}
	return nil
}

// region Steps

func ensureServiceAccount(process *process) error {
	fmt.Println("### EnsureServiceAccount")

	if process.State.HasServiceAccount {
		return nil
	}

	serviceAccountId, err := process.Client.CreateServiceAccount(context.TODO(), "sa-some-name", "sa description")
	if err != nil {
		return err
	}

	newServiceAccount := &ServiceAccount{
		Id:               *serviceAccountId,
		CapabilityRootId: process.State.CapabilityRootId,
		ClusterAccesses:  []ClusterAccess{NewClusterAccess(*serviceAccountId, process.State.ClusterId, process.State.CapabilityRootId)},
		CreatedAt:        time.Now(),
	}

	process.State.HasServiceAccount = true

	return process.Session.ServiceAccounts().CreateServiceAccount(newServiceAccount)
}

func ensureServiceAccountAcl(process *process) error {
	fmt.Println("### EnsureServiceAccountAcl")
	if process.State.HasClusterAccess {
		return nil
	}

	serviceAccount, err := process.Session.ServiceAccounts().GetServiceAccount(process.State.CapabilityRootId)
	if err != nil {
		return err
	}

	clusterAccess, hasClusterAccess := serviceAccount.TryGetClusterAccess(process.State.ClusterId)

	if !hasClusterAccess {
		clusterAccess = NewClusterAccess(serviceAccount.Id, process.State.ClusterId, process.State.CapabilityRootId)
		serviceAccount.ClusterAccesses = append(serviceAccount.ClusterAccesses, clusterAccess)

		if err = process.Session.ServiceAccounts().CreateClusterAccess(clusterAccess); err != nil {
			return err
		}
	}

	entries := clusterAccess.GetAclPendingCreation()
	if len(entries) == 0 {
		// no acl entries left => mark as done
		process.State.HasClusterAccess = true
		return nil

	} else {
		nextEntry := entries[0]

		if err := process.Client.CreateACLEntry(context.TODO(), process.State.ClusterId, serviceAccount.Id, nextEntry.AclDefinition); err != nil {
			return err
		}

		now := time.Now()
		nextEntry.CreatedAt = &now

		return process.Session.ServiceAccounts().UpdateAclEntry(&nextEntry)
	}
}

func ensureServiceAccountApiKey(process *process) error {
	fmt.Println("### EnsureServiceAccountApiKey")
	if process.State.HasApiKey {
		return nil
	}

	serviceAccount, err := process.Session.ServiceAccounts().GetServiceAccount(process.State.CapabilityRootId)
	if err != nil {
		return err
	}

	clusterAccess, _ := serviceAccount.TryGetClusterAccess(process.State.ClusterId)

	key, err := process.Client.CreateApiKey(context.TODO(), clusterAccess.ClusterId, serviceAccount.Id)
	if err != nil {
		return err
	}

	clusterAccess.ApiKey = *key

	err = process.Session.ServiceAccounts().UpdateClusterAccess(clusterAccess)
	if err != nil {
		return err
	}

	process.State.HasApiKey = true
	return nil
}

func ensureServiceAccountApiKeyAreStoredInVault(process *process) error {
	fmt.Println("### EnsureServiceAccountApiKeyAreStoredInVault")
	if process.State.HasApiKeyInVault {
		return nil
	}

	serviceAccount, err := process.Session.ServiceAccounts().GetServiceAccount(process.State.CapabilityRootId)
	if err != nil {
		return err
	}

	clusterAccess, _ := serviceAccount.TryGetClusterAccess(process.State.ClusterId)

	if err := process.Aws.StoreApiKey(context.TODO(), process.State.CapabilityRootId, clusterAccess.ClusterId, clusterAccess.ApiKey); err != nil {
		return err
	}

	process.State.HasApiKeyInVault = true

	return nil
}

func ensureTopicIsCreated(process *process) error {
	fmt.Println("### EnsureTopicIsCreated")

	topic := process.State.Topic
	err := process.Client.CreateTopic(context.TODO(), process.State.ClusterId, topic.Name, topic.Partitions, topic.Retention)
	if err != nil {
		return err
	}

	process.State.MarkAsCompleted()

	return nil
}

// endregion
