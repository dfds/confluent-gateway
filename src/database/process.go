package database

import (
	"context"
	"errors"
	"fmt"
	"github.com/dfds/confluent-gateway/models"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type processRepository struct {
	db *gorm.DB
}

func (r *processRepository) FindById(ctx context.Context, id uuid.UUID) (*models.Process, error) {
	var process = models.Process{}

	err := r.query(ctx).Find(&process, id).Error
	if err != nil {
		return nil, err
	}

	return &process, nil
}

func (r *processRepository) query(ctx context.Context) *gorm.DB {
	return r.db.Model(&models.Process{}).Preload("Acl").Preload("ServiceAccount").WithContext(ctx)
}

func (r *processRepository) FindNextIncomplete(ctx context.Context) (*models.Process, error) {
	var process = models.Process{}

	err := r.query(ctx).Where("completed_at is null").Order("created_at asc").First(&process).Error
	if err != nil {
		return nil, err
	}

	return &process, nil
}

func (r *processRepository) Find(ctx context.Context, capabilityRootId models.CapabilityRootId, clusterId models.ClusterId, topicName string) (*models.Process, error) {
	var process = models.Process{}
	err := r.query(ctx).First(&process, "capability_root_id = ? and cluster_id = ? and topic_name = ?", capabilityRootId, clusterId, topicName).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &process, nil
}

func (r *processRepository) Save(ctx context.Context, process *models.Process) error {
	return r.db.WithContext(ctx).Create(process).Error
}

func (r *processRepository) Update(ctx context.Context, process *models.Process) error {
	fmt.Println("Update")
	return r.db.Debug().WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Clauses(clause.OnConflict{DoNothing: true}).Save(process).Error
}

func NewProcessRepository(db *gorm.DB) models.ProcessRepository {
	return &processRepository{db}
}
