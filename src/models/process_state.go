package models

import (
	"github.com/satori/go.uuid"
	"time"
)

type ProcessState struct {
	Id                uuid.UUID `gorm:"type:uuid;primarykey"`
	CapabilityRootId  CapabilityRootId
	ClusterId         ClusterId
	TopicName         string
	TopicPartitions   int
	TopicRetention    int64
	HasServiceAccount bool
	HasClusterAccess  bool
	HasApiKey         bool
	HasApiKeyInVault  bool
	CreatedAt         time.Time
	CompletedAt       *time.Time
}

func NewProcessState(capabilityRootId CapabilityRootId, clusterId ClusterId, topic Topic, hasServiceAccount bool, hasClusterAccess bool) *ProcessState {
	return &ProcessState{
		Id:                uuid.NewV4(),
		CapabilityRootId:  capabilityRootId,
		ClusterId:         clusterId,
		TopicName:         topic.Name,
		TopicPartitions:   topic.Partitions,
		TopicRetention:    int64(topic.Retention / time.Millisecond),
		HasServiceAccount: hasServiceAccount,
		HasClusterAccess:  hasClusterAccess,
		HasApiKey:         hasClusterAccess,
		HasApiKeyInVault:  hasClusterAccess,
		CreatedAt:         time.Now(),
		CompletedAt:       nil,
	}
}

func (*ProcessState) TableName() string {
	return "process"
}

func (p *ProcessState) IsCompleted() bool {
	return p.CompletedAt != nil
}

func (p *ProcessState) MarkAsCompleted() {
	if p.IsCompleted() {
		return
	}

	now := time.Now()
	p.CompletedAt = &now
}

func (p *ProcessState) Topic() Topic {
	return NewTopic(p.TopicName, p.TopicPartitions, time.Duration(p.TopicRetention)*time.Millisecond)
}
