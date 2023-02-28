package models

import (
	"github.com/satori/go.uuid"
	"time"
)

type CreateProcess struct {
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

func NewCreateProcess(capabilityRootId CapabilityRootId, clusterId ClusterId, topic TopicDescription, hasServiceAccount bool, hasClusterAccess bool) *CreateProcess {
	return &CreateProcess{
		Id:                uuid.NewV4(),
		CapabilityRootId:  capabilityRootId,
		ClusterId:         clusterId,
		TopicName:         topic.Name,
		TopicPartitions:   topic.Partitions,
		TopicRetention:    topic.RetentionInMs(),
		HasServiceAccount: hasServiceAccount,
		HasClusterAccess:  hasClusterAccess,
		HasApiKey:         hasClusterAccess,
		HasApiKeyInVault:  hasClusterAccess,
		CreatedAt:         time.Now(),
		CompletedAt:       nil,
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
