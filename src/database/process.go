package database

import (
	"errors"
	"github.com/dfds/confluent-gateway/models"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

func (s *dataSession) FindById(id uuid.UUID) (*models.ProcessState, error) {
	var process = models.ProcessState{}

	err := s.query().Find(&process, id).Error
	if err != nil {
		return nil, err
	}

	return &process, nil
}

func (s *dataSession) query() *gorm.DB {
	return s.db.Model(&models.ProcessState{})
}

func (s *dataSession) FindNextIncomplete() (*models.ProcessState, error) {
	var process = models.ProcessState{}

	err := s.query().Where("completed_at is null").Order("created_at asc").First(&process).Error
	if err != nil {
		return nil, err
	}

	return &process, nil
}

func (s *dataSession) Find(capabilityRootId models.CapabilityRootId, clusterId models.ClusterId, topicName string) (*models.ProcessState, error) {
	var process = models.ProcessState{}
	err := s.query().First(&process, "capability_root_id = ? and cluster_id = ? and topic_name = ?", capabilityRootId, clusterId, topicName).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &process, nil
}

func (s *dataSession) Create(process *models.ProcessState) error {
	return s.db.Create(process).Error
}

func (s *dataSession) Update(process *models.ProcessState) error {
	return s.db.Save(process).Error
}
