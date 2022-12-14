package models

import (
	"github.com/satori/go.uuid"
	"time"
)

type ProcessState struct {
	Id                uuid.UUID `gorm:"type:uuid;primarykey"`
	CapabilityRootId  CapabilityRootId
	ClusterId         ClusterId
	Topic             Topic `gorm:"embedded;embeddedPrefix:topic_"`
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
		Topic:             topic,
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
