package models

import (
	"fmt"
	uuid "github.com/satori/go.uuid"
	"time"
)

type ServiceAccountId string

type UserAccountId string

func MakeUserAccountId(id int) UserAccountId {
	return UserAccountId(fmt.Sprintf("User:%d", id))
}

type CapabilityRootId string

type ServiceAccount struct {
	Id               ServiceAccountId `gorm:"primarykey"`
	UserAccountId    UserAccountId
	CapabilityRootId CapabilityRootId
	ClusterAccesses  []ClusterAccess
	CreatedAt        time.Time
}

func (*ServiceAccount) TableName() string {
	return "service_account"
}

func (sa *ServiceAccount) TryGetClusterAccess(clusterId ClusterId) (*ClusterAccess, bool) {
	for _, clusterAccess := range sa.ClusterAccesses {
		if clusterAccess.ClusterId == clusterId {
			return &clusterAccess, true
		}
	}
	return nil, false
}

type ClusterAccess struct {
	Id               uuid.UUID `gorm:"primarykey"`
	ClusterId        ClusterId
	ServiceAccountId ServiceAccountId
	UserAccountId    UserAccountId
	ApiKey           ApiKey `gorm:"embedded;embeddedPrefix:api_key_"`
	Acl              []AclEntry
	CreatedAt        time.Time
}

func (*ClusterAccess) TableName() string {
	return "cluster_access"
}

func (ca *ClusterAccess) GetAclPendingCreation() []AclEntry {
	var pending []AclEntry

	for _, entry := range ca.Acl {
		if entry.CreatedAt == nil {
			pending = append(pending, entry)
		}
	}

	return pending
}

func NewClusterAccess(serviceAccountId ServiceAccountId, userAccountId UserAccountId, clusterId ClusterId, capabilityRootId CapabilityRootId) *ClusterAccess {
	clusterAccessId := uuid.NewV4()

	return &ClusterAccess{
		Id:               clusterAccessId,
		ServiceAccountId: serviceAccountId,
		UserAccountId:    userAccountId,
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
	CreatedAt       *time.Time `gorm:"autoCreateTime:false"`
	AclDefinition
}

func (*AclEntry) TableName() string {
	return "acl"
}

func (entry *AclEntry) Created() {
	now := time.Now()
	entry.CreatedAt = &now
}
