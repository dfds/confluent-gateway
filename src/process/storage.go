package process

import (
	"context"
	"github.com/dfds/confluent-gateway/messaging"
	"github.com/dfds/confluent-gateway/models"
)

type Database interface {
	NewSession(context.Context) Session
}

type Session interface {
	Transaction(func(Transaction) error) error
}

type Transaction interface {
	serviceAccountRepository
	stateRepository
	topicRepository
	UpdateProcessState(state *models.ProcessState) error
	AddToOutbox(entry *messaging.OutboxEntry) error
}

type serviceAccountRepository interface {
	GetServiceAccount(capabilityRootId models.CapabilityRootId) (*models.ServiceAccount, error)
	CreateServiceAccount(serviceAccount *models.ServiceAccount) error
	UpdateAclEntry(aclEntry *models.AclEntry) error
	CreateClusterAccess(clusterAccess *models.ClusterAccess) error
	UpdateClusterAccess(clusterAccess *models.ClusterAccess) error
}

type stateRepository interface {
	GetProcessState(capabilityRootId models.CapabilityRootId, clusterId models.ClusterId, topicName string) (*models.ProcessState, error)
	GetServiceAccount(capabilityRootId models.CapabilityRootId) (*models.ServiceAccount, error)
	CreateProcessState(state *models.ProcessState) error
}

type topicRepository interface {
	CreateTopic(topic *models.Topic) error
}
