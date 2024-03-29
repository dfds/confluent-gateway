package models

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type CreateProcess struct {
	Id              uuid.UUID    `gorm:"type:uuid;primarykey"`
	CapabilityId    CapabilityId `gorm:"column:capability_id"`
	ClusterId       ClusterId
	TopicId         string
	TopicName       string
	TopicPartitions int
	TopicRetention  int64
	CreatedAt       time.Time
	CompletedAt     *time.Time
}

func NewCreateProcess(capabilityId CapabilityId, clusterId ClusterId, topicId string, topic TopicDescription) *CreateProcess {
	return &CreateProcess{
		Id:              uuid.NewV4(),
		CapabilityId:    capabilityId,
		ClusterId:       clusterId,
		TopicId:         topicId,
		TopicName:       topic.Name,
		TopicPartitions: topic.Partitions,
		TopicRetention:  topic.RetentionInMs(),
		CreatedAt:       time.Now(),
		CompletedAt:     nil,
	}
}

func (*CreateProcess) TableName() string {
	return "create_process"
}

func (p *CreateProcess) IsCompleted() bool {
	return p.CompletedAt != nil
}

func (p *CreateProcess) MarkAsCompleted() {
	if p.IsCompleted() {
		return
	}

	now := time.Now()
	p.CompletedAt = &now
}

func (p *CreateProcess) TopicDescription() TopicDescription {
	topic, _ := NewTopicDescription(p.TopicName, p.TopicPartitions, RetentionFromMs(p.TopicRetention))
	return topic
}
