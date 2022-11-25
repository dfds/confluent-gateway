package database

import (
	"context"
	"github.com/dfds/confluent-gateway/models"
	uuid "github.com/satori/go.uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
	return r.db.Model(&models.Process{}).Preload("Acl").WithContext(ctx)
}

func (r *processRepository) FindNextIncomplete(ctx context.Context) (*models.Process, error) {
	var process = models.Process{}

	err := r.query(ctx).Where("completed_at is null").Order("created_at asc").First(&process).Error
	if err != nil {
		return nil, err
	}

	return &process, nil
}

func NewProcessRepository(dsn string) (models.ProcessRepository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &processRepository{
		db: db,
	}, nil
}
