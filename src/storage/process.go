package storage

import (
	"errors"
	"github.com/dfds/confluent-gateway/models"
	"gorm.io/gorm"
)

func (d *database) Find(capabilityRootId models.CapabilityRootId, clusterId models.ClusterId, topicName string) (*models.ProcessState, error) {
	var process = models.ProcessState{}
	err := d.db.
		Model(&models.ProcessState{}).
		First(&process, "capability_root_id = ? and cluster_id = ? and topic_name = ?", capabilityRootId, clusterId, topicName).
		Error

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
