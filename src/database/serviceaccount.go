package database

import (
	"errors"
	"github.com/dfds/confluent-gateway/models"
	"gorm.io/gorm"
)

type serviceAccountRepository struct {
	db *gorm.DB
}

func (r *serviceAccountRepository) query() *gorm.DB {
	return r.db.Model(&models.ServiceAccount{}).Preload("ClusterAccesses").Preload("ClusterAccesses.Acl")
}

func (r *serviceAccountRepository) GetByCapabilityRootId(capabilityRootId models.CapabilityRootId) (*models.ServiceAccount, error) {
	var serviceAccount models.ServiceAccount

	err := r.query().First(&serviceAccount, "capability_root_id = ?", capabilityRootId).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &serviceAccount, nil
}

func (r *serviceAccountRepository) Create(serviceAccount *models.ServiceAccount) error {
	return r.db.Create(serviceAccount).Error
}

func (r *serviceAccountRepository) Save(serviceAccount *models.ServiceAccount) error {
	return r.db.Save(serviceAccount).Error
}

func (r *serviceAccountRepository) UpdateAclEntry(aclEntry *models.AclEntry) error {
	return r.db.Save(aclEntry).Error
}

func (r *serviceAccountRepository) CreateClusterAccess(clusterAccess models.ClusterAccess) error {
	return r.db.Create(clusterAccess).Error
}

func (r *serviceAccountRepository) UpdateClusterAccess(clusterAccess models.ClusterAccess) error {
	return r.db.Save(clusterAccess).Error
}

func NewServiceAccountRepository(db *gorm.DB) models.ServiceAccountRepository {
	return &serviceAccountRepository{db}
}
