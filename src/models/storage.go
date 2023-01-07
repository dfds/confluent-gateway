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

	GetProcessState(CapabilityRootId, ClusterId, string) (*ProcessState, error)
	CreateProcessState(*ProcessState) error
	UpdateProcessState(*ProcessState) error

	CreateTopic(*Topic) error

	AddToOutbox(*messaging.OutboxEntry) error
}
