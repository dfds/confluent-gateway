package database

import (
	"context"
	"errors"
	"github.com/dfds/confluent-gateway/models"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

type processRepository struct {
	db *gorm.DB
}

func (r *processRepository) FindById(ctx context.Context, id uuid.UUID) (*models.ProcessState, error) {
	var process = models.ProcessState{}

	err := r.query(ctx).Find(&process, id).Error
	if err != nil {
		return nil, err
	}

	return &process, nil
}

func (r *processRepository) query(ctx context.Context) *gorm.DB {
	return r.db.Model(&models.ProcessState{}).Preload("Acl").Preload("ServiceAccount").WithContext(ctx)
}

func (r *processRepository) FindNextIncomplete(ctx context.Context) (*models.ProcessState, error) {
	var process = models.ProcessState{}

	err := r.query(ctx).Where("completed_at is null").Order("created_at asc").First(&process).Error
	if err != nil {
		return nil, err
	}

	return &process, nil
}

func (r *processRepository) Find(ctx context.Context, capabilityRootId models.CapabilityRootId, clusterId models.ClusterId, topicName string) (*models.ProcessState, error) {
	var process = models.ProcessState{}
	err := r.query(ctx).First(&process, "capability_root_id = ? and cluster_id = ? and topic_name = ?", capabilityRootId, clusterId, topicName).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &process, nil
}

func (r *processRepository) Create(ctx context.Context, process *models.ProcessState) error {
	return r.db.WithContext(ctx).Create(process).Error
}

func (r *processRepository) Update(ctx context.Context, process *models.ProcessState) error {
	return r.db.WithContext(ctx).Save(process).Error
}

func NewProcessRepository(db *gorm.DB) models.ProcessRepository {
	return &processRepository{db}
}
