package models

import (
	"context"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"time"
)

type Process struct {
	Id               uuid.UUID `gorm:"type:uuid;primarykey"`
	CapabilityRootId CapabilityRootId
	ClusterId        ClusterId
	Topic            Topic `gorm:"embedded;embeddedPrefix:topic_"`
	ServiceAccountId *ServiceAccountId
	ServiceAccount   *ServiceAccount
	Acl              []ProcessAclEntry
	ApiKey           ApiKey `gorm:"embedded;embeddedPrefix:api_key_"`
	ApiKeyCreatedAt  *time.Time
	CreatedAt        *time.Time
	CompletedAt      *time.Time
}

func (*Process) TableName() string {
	return "process"
}

func (p *Process) IsCompleted() bool {
	return p.CompletedAt != nil
}

func (p *Process) MarkAsCompleted() {
	if p.IsCompleted() {
		return
	}

	now := time.Now()
	p.CompletedAt = &now
}

func (p *Process) SetServiceAccount(serviceAccount *ServiceAccount) {
	p.ServiceAccountId = &serviceAccount.Id
	p.ServiceAccount = serviceAccount
}

func NewProcess(capabilityRootId CapabilityRootId, clusterId ClusterId, topic Topic) *Process {
	return &Process{
		Id:               uuid.NewV4(),
		CapabilityRootId: capabilityRootId,
		ClusterId:        clusterId,
		Topic:            topic,
		ServiceAccountId: nil,
		ServiceAccount:   nil,
		Acl:              nil,
		ApiKey:           ApiKey{},
		ApiKeyCreatedAt:  nil,
		CreatedAt:        nil,
		CompletedAt:      nil,
	}
}

type ProcessAclEntry struct {
	Id        uuid.UUID `gorm:"primarykey"`
	ProcessId uuid.UUID
	CreatedAt *time.Time
	AclDefinition
}

func (*ProcessAclEntry) TableName() string {
	return "acl"
}

type TopicCreationProcess struct {
	r ProcessRepository
}

func NewTopicCreationProcess(r ProcessRepository) *TopicCreationProcess {
	return &TopicCreationProcess{r}
}

func (p *TopicCreationProcess) ProcessLogic(request NewTopicHasBeenRequested) {
	capabilityRootId := CapabilityRootId(request.CapabilityRootId)
	clusterId := ClusterId(request.ClusterId)
	topic := NewTopic(request.TopicName, request.Partitions, request.Retention)

	process, _ := p.findProcess(capabilityRootId, clusterId, topic.Name)
	if process == nil {
		process = NewProcess(capabilityRootId, clusterId, topic)

		p.save(process)
	}

	if process.IsCompleted() {
		// already completed => skip
		//return
	}

	//1. Ensure capability has cluster access

	//  1.2. Ensure capability has service account

	fmt.Println("### EnsureServiceAccount")
	p.ensureServiceAccount(process)
	p.update(process)

	//	1.3. Ensure service account has all acls
	fmt.Println("### EnsureServiceAccountAcl")
	p.ensureServiceAccountAcl(process)
	p.update(process)

	//	1.4. Ensure service account has api keys
	fmt.Println("### EnsureServiceAccountApiKey")
	p.ensureServiceAccountApiKey(process)
	p.update(process)

	//	1.5. Ensure api keys are stored in vault
	fmt.Println("### EnsureServiceAccountApiKeyAreStoredInVault")
	p.ensureServiceAccountApiKeyAreStoredInVault(process)
	p.update(process)

	//2. Ensure topic is created
	p.ensureTopicIsCreated(process)

	process.MarkAsCompleted()
	p.update(process)

}

func (p *TopicCreationProcess) findProcess(capabilityRootId CapabilityRootId, clusterId ClusterId, topicName string) (*Process, error) {
	// load from database
	process, err := p.r.Find(context.TODO(), capabilityRootId, clusterId, topicName)
	if err != nil {
		panic(err)
	}

	//serviceAccountId := ServiceAccountId("sa-some")

	return process, nil
}

func (p *TopicCreationProcess) save(process *Process) {
	err := p.r.Save(context.TODO(), process)
	if err != nil {
		panic(err)
	}
}

func (p *TopicCreationProcess) update(process *Process) {
	err := p.r.Update(context.TODO(), process)
	if err != nil {
		panic(err)
	}
}

func (p *TopicCreationProcess) ensureServiceAccount(process *Process) {
	if process.ServiceAccount != nil {
		// service account already exists
		return
	}

	serviceAccount := createServiceAccount(process.CapabilityRootId, process.ClusterId)
	process.SetServiceAccount(serviceAccount)
}

func (p *TopicCreationProcess) ensureServiceAccountAcl(process *Process) {
	if len(process.Acl) > 0 {
		return
	}

	allDefinitions := CreateAclDefinitions(process.CapabilityRootId)

	// TODO -- filter by existing Service Account ACLs

	acl := make([]ProcessAclEntry, len(allDefinitions))

	for i, definition := range allDefinitions {
		acl[i] = ProcessAclEntry{
			Id:            uuid.NewV4(),
			ProcessId:     process.Id,
			AclDefinition: definition,
		}
	}

	// TODO -- create ACLs in Confluent Cloud

	process.Acl = acl
}

func (p *TopicCreationProcess) ensureServiceAccountApiKey(process *Process) {
	if len(process.ServiceAccount.ApiKey.Username) > 0 && len(process.ServiceAccount.ApiKey.Password) > 0 {
		return
	}

	process.ServiceAccount.ApiKey.Username = "USERNAME"
	process.ServiceAccount.ApiKey.Password = "PA55W0RD"

	// TODO -- create API key in Confluent Cloud

}

func (p *TopicCreationProcess) ensureServiceAccountApiKeyAreStoredInVault(process *Process) {
	if process.ApiKeyCreatedAt != nil {
		return
	}

	now := time.Now()
	process.ApiKeyCreatedAt = &now
}

func (p *TopicCreationProcess) ensureTopicIsCreated(process *Process) {
}

func createServiceAccount(capabilityRootId CapabilityRootId, clusterId ClusterId) *ServiceAccount {
	// TODO -- create service account in confluent cloud
	return &ServiceAccount{
		Id:               "sa-some",
		CapabilityRootId: capabilityRootId,
		ClusterId:        clusterId,
		ApiKey: ApiKey{
			Username: "",
			Password: "",
		},
	}
}

type ProcessRepository interface {
	FindById(ctx context.Context, id uuid.UUID) (*Process, error)
	FindNextIncomplete(ctx context.Context) (*Process, error)
	Find(ctx context.Context, capabilityRootId CapabilityRootId, clusterId ClusterId, topicName string) (*Process, error)
	Save(ctx context.Context, process *Process) error
	Update(ctx context.Context, process *Process) error
}
