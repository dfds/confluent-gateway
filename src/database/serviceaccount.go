package database

import (
	"errors"
	"github.com/dfds/confluent-gateway/models"
	"gorm.io/gorm"
)

func (s *dataSession) saQuery() *gorm.DB {
	return s.db.Model(&models.ServiceAccount{}).Preload("ClusterAccesses").Preload("ClusterAccesses.Acl")
}

func (s *dataSession) GetByCapabilityRootId(capabilityRootId models.CapabilityRootId) (*models.ServiceAccount, error) {
	var serviceAccount models.ServiceAccount

	err := s.saQuery().First(&serviceAccount, "capability_root_id = ?", capabilityRootId).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &serviceAccount, nil
}

func (s *dataSession) CreateServiceAccount(serviceAccount *models.ServiceAccount) error {
	return s.db.Create(serviceAccount).Error
}

func (s *dataSession) UpdateAclEntry(aclEntry *models.AclEntry) error {
	return s.db.Save(aclEntry).Error
}

func (s *dataSession) CreateClusterAccess(clusterAccess models.ClusterAccess) error {
	return s.db.Create(clusterAccess).Error
}

func (s *dataSession) UpdateClusterAccess(clusterAccess models.ClusterAccess) error {
	return s.db.Save(clusterAccess).Error
}
