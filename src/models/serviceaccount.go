package models

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

type ServiceAccountId string

type ServiceAccount struct {
	Id               ServiceAccountId `gorm:"primarykey"`
	CapabilityRootId CapabilityRootId
	ClusterAccesses  []ClusterAccess
	CreatedAt        time.Time
}

func (*ServiceAccount) TableName() string {
	return "service_account"
}

func (sa *ServiceAccount) TryGetClusterAccess(clusterId ClusterId) (ClusterAccess, bool) {
	for _, clusterAccess := range sa.ClusterAccesses {
		if clusterAccess.ClusterId == clusterId {
			return clusterAccess, true
		}
	}
	return ClusterAccess{}, false
}

type ClusterAccess struct {
	Id               uuid.UUID `gorm:"primarykey"`
	ClusterId        ClusterId
	ServiceAccountId ServiceAccountId
	ApiKey           ApiKey `gorm:"embedded;embeddedPrefix:api_key_"`
	Acl              []AclEntry
	CreatedAt        time.Time
}

func (*ClusterAccess) TableName() string {
	return "cluster_access"
}

func (ca *ClusterAccess) GetAclPendingCreation() []*AclEntry {
	var pending []*AclEntry

	for _, entry := range ca.Acl {
		if entry.CreatedAt == nil {
			pending = append(pending, &entry)
		}
	}

	return pending
}

func NewClusterAccess(serviceAccountId ServiceAccountId, clusterId ClusterId, capabilityRootId CapabilityRootId) ClusterAccess {
	clusterAccessId := uuid.NewV4()

	return ClusterAccess{
		Id:               clusterAccessId,
		ServiceAccountId: serviceAccountId,
		ClusterId:        clusterId,
		ApiKey:           ApiKey{},
		Acl:              createAclEntries(capabilityRootId, clusterAccessId),
		CreatedAt:        time.Now(),
	}
}

func createAclEntries(capabilityRootId CapabilityRootId, clusterAccessId uuid.UUID) []AclEntry {
	allDefinitions := CreateAclDefinitions(capabilityRootId)

	acl := make([]AclEntry, len(allDefinitions))

	for i, definition := range allDefinitions {
		acl[i] = AclEntry{
			Id:              uuid.NewV4(),
			ClusterAccessId: clusterAccessId,
			CreatedAt:       nil,
			AclDefinition:   definition,
		}
	}
	return acl
}

type AclEntry struct {
	Id              uuid.UUID `gorm:"primarykey"`
	ClusterAccessId uuid.UUID
	CreatedAt       *time.Time
	AclDefinition
}

func (*AclEntry) TableName() string {
	return "acl"
}

type ServiceAccountRepository interface {
	GetByCapabilityRootId(capabilityRootId CapabilityRootId) (*ServiceAccount, error)
	Create(serviceAccount *ServiceAccount) error
	Save(serviceAccount *ServiceAccount) error
}
