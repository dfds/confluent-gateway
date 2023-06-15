package models

import (
	"context"
	"github.com/dfds/confluent-gateway/messaging"
)

type Database interface {
	NewSession(context.Context) Session
}

type Session interface {
	Transaction(func(Transaction) error) error
}

type Transaction interface {
	GetServiceAccount(CapabilityId) (*ServiceAccount, error)
	CreateServiceAccount(*ServiceAccount) error
	UpdateAclEntry(*AclEntry) error
	CreateClusterAccess(*ClusterAccess) error
	UpdateClusterAccess(*ClusterAccess) error

	GetCreateProcessState(CapabilityId, ClusterId, string) (*CreateProcess, error)
	SaveCreateProcessState(*CreateProcess) error
	UpdateCreateProcessState(*CreateProcess) error

	GetDeleteProcessState(string) (*DeleteProcess, error)
	SaveDeleteProcessState(*DeleteProcess) error
	UpdateDeleteProcessState(*DeleteProcess) error

	GetTopic(string) (*Topic, error)
	CreateTopic(*Topic) error
	DeleteTopic(string) error

	GetSchemaProcessState(string) (*SchemaProcess, error)
	SaveSchemaProcessState(*SchemaProcess) error
	UpdateSchemaProcessState(*SchemaProcess) error

	SelectSchemaProcessStatesByTopicId(string) ([]SchemaProcess, error)
	DeleteSchemaProcessStateById(string) error

	AddToOutbox(*messaging.OutboxEntry) error
}
