package database

import (
	"context"
	"errors"
	"github.com/dfds/confluent-gateway/models"
	"gorm.io/gorm"
)

type serviceAccountRepository struct {
	db *gorm.DB
}

func (r *serviceAccountRepository) prepare(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *serviceAccountRepository) query(ctx context.Context) *gorm.DB {
	return r.prepare(ctx).Model(&models.ServiceAccount{}).Preload("ClusterAccesses").Preload("ClusterAccesses.Acl")
}

func (r *serviceAccountRepository) GetByCapabilityRootId(ctx context.Context, capabilityRootId models.CapabilityRootId) (*models.ServiceAccount, error) {
	var serviceAccount models.ServiceAccount

	err := r.query(ctx).First(&serviceAccount, "capability_root_id = ?", capabilityRootId).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &serviceAccount, nil
}

func (r *serviceAccountRepository) Create(ctx context.Context, serviceAccount *models.ServiceAccount) error {
	return r.prepare(ctx).Create(serviceAccount).Error
}

func (r *serviceAccountRepository) Save(ctx context.Context, serviceAccount *models.ServiceAccount) error {
	return r.prepare(ctx).Save(serviceAccount).Error
}

func NewServiceAccountRepository(db *gorm.DB) models.ServiceAccountRepository {
	return &serviceAccountRepository{db}
}
