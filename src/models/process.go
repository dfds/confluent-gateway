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
	Create(process *ProcessState) error
	Update(process *ProcessState) error
}

//endregion

type TopicCreationProcess struct {
	data DataAccess
}

func NewTopicCreationProcess(data DataAccess) *TopicCreationProcess {
	return &TopicCreationProcess{data}
}

func (tcp *TopicCreationProcess) ProcessLogic(ctx context.Context, request NewTopicHasBeenRequested) {
	p, err := tcp.prepareProcess(ctx, request)
	if err != nil {
		panic(err)
	}

	if p.State.IsCompleted() {
		// already completed => skip
		return
	}

	//1. Ensure capability has cluster access
	//  1.2. Ensure capability has service account
	if err := p.execute(ensureServiceAccount); err != nil {
		panic(err)
	}
	//	1.3. Ensure service account has all acls
	if err := p.execute(ensureServiceAccountAcl); err != nil {
		panic(err)
	}
	//	1.4. Ensure service account has api keys
	if err := p.execute(ensureServiceAccountApiKey); err != nil {
		panic(err)
	}
	//	1.5. Ensure api keys are stored in vault
	if err := p.execute(ensureServiceAccountApiKeyAreStoredInVault); err != nil {
		panic(err)
	}
	//2. Ensure topic is created
	if err := p.execute(ensureTopicIsCreated); err != nil {
		panic(err)
	}
}

func (tcp *TopicCreationProcess) prepareProcess(ctx context.Context, request NewTopicHasBeenRequested) (*process, error) {
	capabilityRootId := CapabilityRootId(request.CapabilityRootId)
	clusterId := ClusterId(request.ClusterId)
	topic := NewTopic(request.TopicName, request.Partitions, request.Retention)

	session := tcp.data.NewSession(ctx)

	serviceAccount, err := session.ServiceAccounts().GetByCapabilityRootId(capabilityRootId)
	if err != nil {
		return nil, err
	}

	hasServiceAccount := false
	hasClusterAccess := false
	hasApiKey := false
	hasApiKeyInVault := false

	if serviceAccount != nil {
		hasServiceAccount = true
		_, hasClusterAccess = serviceAccount.TryGetClusterAccess(clusterId)
		hasApiKey = hasClusterAccess
		hasApiKeyInVault = hasClusterAccess
	}

	state := &ProcessState{
		Id:                uuid.NewV4(),
		CapabilityRootId:  capabilityRootId,
		ClusterId:         clusterId,
		Topic:             topic,
		HasServiceAccount: hasServiceAccount,
		HasClusterAccess:  hasClusterAccess,
		HasApiKey:         hasApiKey,
		HasApiKeyInVault:  hasApiKeyInVault,
		CreatedAt:         time.Now(),
		CompletedAt:       nil,
	}

	if err := session.Processes().Create(state); err != nil {
		return nil, err
	}

	return &process{session, state}, nil
}

type process struct {
	Session DataSession
	State   *ProcessState
}

func (p *process) execute(stepFunc func(*process) error) error {
	return p.Session.Transaction(func(session DataSession) error {
		process := &process{session, p.State}

		if err := stepFunc(process); err != nil {
			return err
		}

		return session.Processes().Update(p.State)
	})
}

// region Steps

func ensureServiceAccount(process *process) error {
	fmt.Println("### EnsureServiceAccount")

	if process.State.HasServiceAccount {
		return nil
	}

	// TODO -- create service account in confluent cloud
	serviceAccountId := ServiceAccountId("sa-some")

	newServiceAccount := &ServiceAccount{
		Id:               serviceAccountId,
		CapabilityRootId: process.State.CapabilityRootId,
		ClusterAccesses:  []ClusterAccess{NewClusterAccess(serviceAccountId, process.State.ClusterId, process.State.CapabilityRootId)},
		CreatedAt:        time.Now(),
	}

	process.State.HasServiceAccount = true

	return process.Session.ServiceAccounts().Create(newServiceAccount)
}

func ensureServiceAccountAcl(process *process) error {
	fmt.Println("### EnsureServiceAccountAcl")
	if process.State.HasClusterAccess {
		return nil
	}

	serviceAccount, err := process.Session.ServiceAccounts().GetByCapabilityRootId(process.State.CapabilityRootId)
	if err != nil {
		return err
	}

	clusterAccess, ok := serviceAccount.TryGetClusterAccess(process.State.ClusterId)

	if !ok {
		clusterAccess = NewClusterAccess(serviceAccount.Id, process.State.ClusterId, process.State.CapabilityRootId)
		serviceAccount.ClusterAccesses = append(serviceAccount.ClusterAccesses, clusterAccess)
		process.Session.ServiceAccounts().Save(serviceAccount)
	}

	for _, entry := range clusterAccess.Acl {
		if entry.CreatedAt != nil {
			continue
		}

		// TODO -- create ACLs in Confluent Cloud
		now := time.Now()
		entry.CreatedAt = &now

		err := process.Session.ServiceAccounts().Save(serviceAccount)
		if err != nil {
			return err
		}
	}

	process.State.HasClusterAccess = true
	return nil
}

func ensureServiceAccountApiKey(process *process) error {
	fmt.Println("### EnsureServiceAccountApiKey")
	if process.State.HasApiKey {
		return nil
	}

	serviceAccount, err := process.Session.ServiceAccounts().GetByCapabilityRootId(process.State.CapabilityRootId)
	if err != nil {
		return err
	}

	clusterAccess, _ := serviceAccount.TryGetClusterAccess(process.State.ClusterId)

	// TODO -- create API key in Confluent Cloud

	clusterAccess.ApiKey = ApiKey{"USERNAME", "PA55W0RD"}
	err = process.Session.ServiceAccounts().Save(serviceAccount)
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

	serviceAccount, err := process.Session.ServiceAccounts().GetByCapabilityRootId(process.State.CapabilityRootId)
	if err != nil {
		return err
	}

	clusterAccess, _ := serviceAccount.TryGetClusterAccess(process.State.ClusterId)

	// TODO -- save API key in vault
	_ = clusterAccess.ApiKey

	process.State.HasApiKeyInVault = true

	return nil
}

func ensureTopicIsCreated(process *process) error {
	// TODO -- create topic in Confluent Cloud

	process.State.MarkAsCompleted()

	return nil
}

// endregion
