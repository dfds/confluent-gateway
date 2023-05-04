package models

import (
	"github.com/satori/go.uuid"
	"time"
)

type DeleteProcess struct {
	Id          uuid.UUID `gorm:"type:uuid;primarykey"`
	TopicId     string
	CreatedAt   time.Time
	CompletedAt *time.Time
}

func NewDeleteProcess(topicId string) *DeleteProcess {
	return &DeleteProcess{
		Id:          uuid.NewV4(),
		TopicId:     topicId,
		CreatedAt:   time.Now(),
		CompletedAt: nil,
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
