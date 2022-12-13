package models

import "context"

type Database interface {
	NewSession(context.Context) DataSession
}

type DataSession interface {
	Transaction(func(DataSession) error) error
	ServiceAccounts() ServiceAccountRepository
	Processes() ProcessRepository
}

type ServiceAccountRepository interface {
	GetServiceAccount(capabilityRootId CapabilityRootId) (*ServiceAccount, error)
	CreateServiceAccount(serviceAccount *ServiceAccount) error
	UpdateAclEntry(aclEntry *AclEntry) error
	CreateClusterAccess(clusterAccess *ClusterAccess) error
	UpdateClusterAccess(clusterAccess *ClusterAccess) error
}

type ProcessRepository interface {
	CreateProcessState(state *ProcessState) error
	UpdateProcessState(state *ProcessState) error
	GetProcessState(capabilityRootId CapabilityRootId, clusterId ClusterId, topicName string) (*ProcessState, error)
}
