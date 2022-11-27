package models

import (
	uuid "github.com/satori/go.uuid"
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
