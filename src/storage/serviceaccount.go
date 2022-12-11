package storage

import (
	"errors"
	"github.com/dfds/confluent-gateway/models"
	"gorm.io/gorm"
)

func (d *Database) GetByCapabilityRootId(capabilityRootId models.CapabilityRootId) (*models.ServiceAccount, error) {
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

func (d *Database) CreateServiceAccount(serviceAccount *models.ServiceAccount) error {
	return d.db.Create(serviceAccount).Error
}

func (d *Database) UpdateAclEntry(aclEntry *models.AclEntry) error {
	return d.db.Save(aclEntry).Error
}

func (d *Database) CreateClusterAccess(clusterAccess models.ClusterAccess) error {
	return d.db.Create(clusterAccess).Error
}

func (d *Database) UpdateClusterAccess(clusterAccess models.ClusterAccess) error {
	return d.db.Save(clusterAccess).Error
}
