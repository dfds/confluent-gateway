package storage

import (
	"errors"
	"github.com/dfds/confluent-gateway/models"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

func (d *database) FindById(id uuid.UUID) (*models.ProcessState, error) {
	var process = models.ProcessState{}

	err := d.query().Find(&process, id).Error
	if err != nil {
		return nil, err
	}

	return &process, nil
}

func (d *database) query() *gorm.DB {
	return d.db.Model(&models.ProcessState{})
}

func (d *database) FindNextIncomplete() (*models.ProcessState, error) {
	var process = models.ProcessState{}

	err := d.query().Where("completed_at is null").Order("created_at asc").First(&process).Error
	if err != nil {
		return nil, err
	}

	return &process, nil
}

func (d *database) Find(capabilityRootId models.CapabilityRootId, clusterId models.ClusterId, topicName string) (*models.ProcessState, error) {
	var process = models.ProcessState{}
	err := d.query().First(&process, "capability_root_id = ? and cluster_id = ? and topic_name = ?", capabilityRootId, clusterId, topicName).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &process, nil
}

func (d *database) Create(process *models.ProcessState) error {
	return d.db.Create(process).Error
}

func (d *database) Update(process *models.ProcessState) error {
	return d.db.Save(process).Error
}
