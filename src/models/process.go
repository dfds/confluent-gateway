package models

import (
	"context"
	uuid "github.com/satori/go.uuid"
	"time"
)

type Process struct {
	Id               uuid.UUID `gorm:"type:uuid;primarykey"`
	CapabilityRootId CapabilityRootId
	ClusterId        ClusterId
	Topic            Topic `gorm:"embedded;embeddedPrefix:topic_"`
	ServiceAccountId *ServiceAccountId
	Acl              []AclEntry
	ApiKey           ApiKey `gorm:"embedded;embeddedPrefix:api_key_"`
	ApiKeyCreatedAt  *time.Time
	CreatedAt        *time.Time
	CompletedAt      *time.Time
}

func (*Process) TableName() string {
	return "process"
}

func (p *Process) ProcessLogic(request NewTopicHasBeenRequested) {
	//1. Ensure capability has cluster access
	//EnsureCapabilityHasClusterAccess(capabilityRootId, clusterId, request.TopicName)
	//  1.2. Ensure capability has service account
	//	1.3. Ensure service account has all acls
	//	1.4. Ensure service account has api keys
	//	1.5. Ensure api keys are stored in vault
	//2. Ensure topic is created
	//3. Done!

}

type processLoginInput struct {
	CapabilityRootId string // example => logistics-somecapability-abcd
	ClusterId        string
	TopicName        string // full name => pub.logistics-somecapability-abcd.foo
	Partitions       int
	Retention        int // in ms
}

type ProcessRepository interface {
	FindById(ctx context.Context, id uuid.UUID) (*Process, error)
	FindNextIncomplete(ctx context.Context) (*Process, error)
}
