package storage

import (
	"errors"
	"github.com/dfds/confluent-gateway/models"
	"gorm.io/gorm"
)

func (d *database) GetByCapabilityRootId(capabilityRootId models.CapabilityRootId) (*models.ServiceAccount, error) {
	var serviceAccount models.ServiceAccount

	err := d.db.
		Model(&serviceAccount).
		Preload("ClusterAccesses").
		Preload("ClusterAccesses.Acl").
		First(&serviceAccount, "capability_root_id = ?", capabilityRootId).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &serviceAccount, nil
}

func (d *database) CreateServiceAccount(serviceAccount *models.ServiceAccount) error {
	return d.db.Create(serviceAccount).Error
}

func (d *database) UpdateAclEntry(aclEntry *models.AclEntry) error {
	return d.db.Save(aclEntry).Error
}

func (d *database) CreateClusterAccess(clusterAccess models.ClusterAccess) error {
	return d.db.Create(clusterAccess).Error
}

func (d *database) UpdateClusterAccess(clusterAccess models.ClusterAccess) error {
	return d.db.Save(clusterAccess).Error
}
