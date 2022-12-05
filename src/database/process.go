package database

import (
	"errors"
	"github.com/dfds/confluent-gateway/models"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

type processRepository struct {
	db *gorm.DB
}

func (r *processRepository) FindById(id uuid.UUID) (*models.ProcessState, error) {
	var process = models.ProcessState{}

	err := r.query().Find(&process, id).Error
	if err != nil {
		return nil, err
	}

	return &process, nil
}

func (r *processRepository) query() *gorm.DB {
	return r.db.Model(&models.ProcessState{})
}

func (r *processRepository) FindNextIncomplete() (*models.ProcessState, error) {
	var process = models.ProcessState{}

	err := r.query().Where("completed_at is null").Order("created_at asc").First(&process).Error
	if err != nil {
		return nil, err
	}

	return &process, nil
}

func (r *processRepository) Find(capabilityRootId models.CapabilityRootId, clusterId models.ClusterId, topicName string) (*models.ProcessState, error) {
	var process = models.ProcessState{}
	err := r.query().First(&process, "capability_root_id = ? and cluster_id = ? and topic_name = ?", capabilityRootId, clusterId, topicName).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &process, nil
}

func (r *processRepository) Create(process *models.ProcessState) error {
	return r.db.Create(process).Error
}

func (r *processRepository) Update(process *models.ProcessState) error {
	return r.db.Save(process).Error
}

func NewProcessRepository(db *gorm.DB) models.ProcessRepository {
	return &processRepository{db}
}
