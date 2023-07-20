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

type CapabilityId string

type ServiceAccount struct {
	Id              ServiceAccountId `gorm:"primarykey"`
	UserAccountId   UserAccountId
	CapabilityId    CapabilityId
	ClusterAccesses []ClusterAccess
	CreatedAt       time.Time
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

func NewClusterAccess(serviceAccountId ServiceAccountId, userAccountId UserAccountId, clusterId ClusterId, capabilityId CapabilityId) *ClusterAccess {
	clusterAccessId := uuid.NewV4()

	return &ClusterAccess{
		Id:               clusterAccessId,
		ServiceAccountId: serviceAccountId,
		UserAccountId:    userAccountId,
		ClusterId:        clusterId,
		Acl:              createAclEntries(capabilityId, clusterAccessId),
		CreatedAt:        time.Now(),
	}
}

func createAclEntries(capabilityId CapabilityId, clusterAccessId uuid.UUID) []AclEntry {
	allDefinitions := CreateAclDefinitions(capabilityId)

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

func (e *AclEntry) Created() {
	now := time.Now()
	e.CreatedAt = &now
}

func (e *AclEntry) IsValid() bool {
	return e.CreatedAt != nil
}
