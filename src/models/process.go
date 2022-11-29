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
	Create(ctx context.Context, process *ProcessState) error
	Update(ctx context.Context, process *ProcessState) error
}

//endregion

type TopicCreationProcess struct {
	data DataAccess
}

func NewTopicCreationProcess(data DataAccess) *TopicCreationProcess {
	return &TopicCreationProcess{data}
}

func (tcp *TopicCreationProcess) ProcessLogic(request NewTopicHasBeenRequested) {
	p, err := tcp.prepareProcess(request)
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

func (tcp *TopicCreationProcess) prepareProcess(request NewTopicHasBeenRequested) (*process, error) {
	capabilityRootId := CapabilityRootId(request.CapabilityRootId)
	clusterId := ClusterId(request.ClusterId)
	topic := NewTopic(request.TopicName, request.Partitions, request.Retention)

	serviceAccount, err := tcp.data.ServiceAccounts().GetByCapabilityRootId(context.TODO(), capabilityRootId)
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

	if err := tcp.data.Processes().Create(context.TODO(), state); err != nil {
		return nil, err
	}

	return &process{tcp.data, state}, nil
}

type process struct {
	Data  DataAccess
	State *ProcessState
}

func (p *process) execute(stepFunc func(*process) error) error {
	return p.Data.Transaction(func(data DataAccess) error {
		session := &process{data, p.State}

		if err := stepFunc(session); err != nil {
			return err
		}

		return data.Processes().Update(context.TODO(), p.State)
	})
}

// region Steps

func ensureServiceAccount(session *process) error {
	fmt.Println("### EnsureServiceAccount")

	if session.State.HasServiceAccount {
		return nil
	}

	// TODO -- create service account in confluent cloud
	serviceAccountId := ServiceAccountId("sa-some")

	newServiceAccount := &ServiceAccount{
		Id:               serviceAccountId,
		CapabilityRootId: session.State.CapabilityRootId,
		ClusterAccesses:  []ClusterAccess{NewClusterAccess(serviceAccountId, session.State.ClusterId, session.State.CapabilityRootId)},
		CreatedAt:        time.Now(),
	}

	session.State.HasServiceAccount = true

	return session.Data.ServiceAccounts().Create(context.TODO(), newServiceAccount)
}

func ensureServiceAccountAcl(session *process) error {
	fmt.Println("### EnsureServiceAccountAcl")
	if session.State.HasClusterAccess {
		return nil
	}

	serviceAccount, err := session.Data.ServiceAccounts().GetByCapabilityRootId(context.TODO(), session.State.CapabilityRootId)
	if err != nil {
		return err
	}

	clusterAccess, ok := serviceAccount.TryGetClusterAccess(session.State.ClusterId)

	if !ok {
		clusterAccess = NewClusterAccess(serviceAccount.Id, session.State.ClusterId, session.State.CapabilityRootId)
		serviceAccount.ClusterAccesses = append(serviceAccount.ClusterAccesses, clusterAccess)
		session.Data.ServiceAccounts().Save(context.TODO(), serviceAccount)
	}

	for _, entry := range clusterAccess.Acl {
		if entry.CreatedAt != nil {
			continue
		}

		// TODO -- create ACLs in Confluent Cloud
		now := time.Now()
		entry.CreatedAt = &now

		err := session.Data.ServiceAccounts().Save(context.TODO(), serviceAccount)
		if err != nil {
			return err
		}
	}

	session.State.HasClusterAccess = true
	return nil
}

func ensureServiceAccountApiKey(session *process) error {
	fmt.Println("### EnsureServiceAccountApiKey")
	if session.State.HasApiKey {
		return nil
	}

	serviceAccount, err := session.Data.ServiceAccounts().GetByCapabilityRootId(context.TODO(), session.State.CapabilityRootId)
	if err != nil {
		return err
	}

	clusterAccess, _ := serviceAccount.TryGetClusterAccess(session.State.ClusterId)

	// TODO -- create API key in Confluent Cloud

	clusterAccess.ApiKey = ApiKey{"USERNAME", "PA55W0RD"}
	err = session.Data.ServiceAccounts().Save(context.TODO(), serviceAccount)
	if err != nil {
		return err
	}

	session.State.HasApiKey = true
	return nil
}

func ensureServiceAccountApiKeyAreStoredInVault(session *process) error {
	fmt.Println("### EnsureServiceAccountApiKeyAreStoredInVault")
	if session.State.HasApiKeyInVault {
		return nil
	}

	serviceAccount, err := session.Data.ServiceAccounts().GetByCapabilityRootId(context.TODO(), session.State.CapabilityRootId)
	if err != nil {
		return err
	}

	clusterAccess, _ := serviceAccount.TryGetClusterAccess(session.State.ClusterId)

	// TODO -- save API key in vault
	_ = clusterAccess.ApiKey

	session.State.HasApiKeyInVault = true

	return nil
}

func ensureTopicIsCreated(session *process) error {
	// TODO -- create topic in Confluent Cloud

	session.State.MarkAsCompleted()

	return nil
}

// endregion
