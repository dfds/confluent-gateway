package models

import (
	"context"
	uuid "github.com/satori/go.uuid"
	"time"
)

type NewTopicHasBeenRequested struct {
	CapabilityRootId string // example => logistics-somecapability-abcd
	ClusterId        string
	TopicName        string // full name => pub.logistics-somecapability-abcd.foo
	Partitions       int
	Retention        int // in ms
}

type CapabilityRootId string
type ServiceAccountId string
type ClusterId string

type Cluster struct {
	ClusterId         ClusterId `gorm:"column:id;primarykey"`
	Name              string
	AdminApiEndpoint  string
	AdminApiKey       ApiKey `gorm:"embedded;embeddedPrefix:admin_api_key_"`
	BootstrapEndpoint string
}

func (*Cluster) TableName() string {
	return "cluster"
}

type ClusterRepository interface {
	Get(ctx context.Context, id ClusterId) (Cluster, error)
	GetAll(ctx context.Context) ([]Cluster, error)
}

type Acl struct {
	Entries []AclEntry
}

type CloudApiAccess struct {
	UserName    string
	Password    string
	ApiEndpoint string
}

type ServiceAccount struct {
	Id               ServiceAccountId
	CapabilityRootId CapabilityRootId
	ClusterAccess    []ClusterAccess
}

type ClusterAccess struct {
	ClusterId        ClusterId
	ServiceAccountId ServiceAccountId
	ApiKey           ApiKey
}

type Topic struct {
	Name       string
	Partitions int
	Retention  int
}

type AclEntry struct {
	Id             int `gorm:"primarykey"`
	ProcessId      uuid.UUID
	ResourceType   string
	ResourceName   string
	PatternType    string
	OperationType  string
	PermissionType string
}

func (*AclEntry) TableName() string {
	return "acl"
}

type ApiKey struct {
	UserName string
	Password string
}

type Process struct {
	Id               uuid.UUID `gorm:"type:uuid;primarykey"`
	CapabilityRootId CapabilityRootId
	ClusterId        ClusterId
	Topic            Topic `gorm:"embedded;embeddedPrefix:topic_"`
	ServiceAccountId ServiceAccountId
	Acl              []AclEntry
	ApiKey           ApiKey `gorm:"embedded;embeddedPrefix:api_key_"`
	ApiKeyCreatedAt  *time.Time
	CreatedAt        *time.Time
	CompletedAt      *time.Time
}

func (*Process) TableName() string {
	return "process"
}

func (p *Process) ProcessLogic() {
	//1. Ensure capability has cluster access
	//  1.2. Ensure capability has service account
	//	1.3. Ensure service account has all acls
	//	1.4. Ensure service account has api keys
	//	1.5. Ensure api keys are stored in vault
	//2. Ensure topic is created
	//3. Done!
}

type ProcessRepository interface {
	FindById(ctx context.Context, id uuid.UUID) (*Process, error)
	FindNextIncomplete(ctx context.Context) (*Process, error)
}
