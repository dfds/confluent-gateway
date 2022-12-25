package process

import (
	"context"
	"github.com/dfds/confluent-gateway/models"
)

type Database interface {
	NewSession(context.Context) DataSession
}

type DataSession interface {
	Transaction(func(DataSession) error) error

	serviceAccountRepository
	stateRepository
	UpdateProcessState(state *models.ProcessState) error
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
	CreateProcessState(state *models.ProcessState) error
	GetServiceAccount(capabilityRootId models.CapabilityRootId) (*models.ServiceAccount, error)
}
