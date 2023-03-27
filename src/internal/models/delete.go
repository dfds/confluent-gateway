package models

import (
	"github.com/satori/go.uuid"
	"time"
)

type DeleteProcess struct {
	Id           uuid.UUID    `gorm:"type:uuid;primarykey"`
	CapabilityId CapabilityId `gorm:"column:capability_id"`
	ClusterId    ClusterId
	TopicId      string
	TopicName    string
	CreatedAt    time.Time
	CompletedAt  *time.Time
}

func NewDeleteProcess(capabilityId CapabilityId, clusterId ClusterId, topicId string, topicName string) *DeleteProcess {
	return &DeleteProcess{
		Id:           uuid.NewV4(),
		CapabilityId: capabilityId,
		ClusterId:    clusterId,
		TopicId:      topicId,
		TopicName:    topicName,
		CreatedAt:    time.Now(),
		CompletedAt:  nil,
	}
}

func (*DeleteProcess) TableName() string {
	return "delete_process"
}

func (p *DeleteProcess) IsCompleted() bool {
	return p.CompletedAt != nil
}

func (p *DeleteProcess) MarkAsCompleted() {
	if p.IsCompleted() {
		return
	}

	now := time.Now()
	p.CompletedAt = &now
}
