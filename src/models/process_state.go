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
