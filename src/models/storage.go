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
	GetServiceAccount(CapabilityRootId) (*ServiceAccount, error)
	CreateServiceAccount(*ServiceAccount) error
	UpdateAclEntry(*AclEntry) error
	CreateClusterAccess(*ClusterAccess) error
	UpdateClusterAccess(*ClusterAccess) error

	GetCreateProcessState(CapabilityRootId, ClusterId, string) (*CreateProcess, error)
	SaveCreateProcessState(*CreateProcess) error
	UpdateCreateProcessState(*CreateProcess) error

	GetDeleteProcessState(CapabilityRootId, ClusterId, string) (*DeleteProcess, error)
	SaveDeleteProcessState(*DeleteProcess) error
	UpdateDeleteProcessState(*DeleteProcess) error

	GetTopic(CapabilityRootId, ClusterId, string) (*Topic, error)
	CreateTopic(*Topic) error
	DeleteTopic(CapabilityRootId, ClusterId, string) error

	AddToOutbox(*messaging.OutboxEntry) error
}
