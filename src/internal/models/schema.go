package models

import (
	"github.com/satori/go.uuid"
	"time"
)

type SchemaProcess struct {
	Id                uuid.UUID `gorm:"type:uuid;primarykey"`
	ClusterId         ClusterId
	MessageContractId string
	TopicId           string
	MessageType       string
	Description       string
	Subject           string
	Schema            string
	CreatedAt         time.Time
	CompletedAt       *time.Time
	SchemaVersion     int32
}

func NewSchemaProcess(clusterId ClusterId, messageContractId string, topicId string, messageType string, description string, subject string, schema string, schemaVersion int32) *SchemaProcess {
	return &SchemaProcess{
		Id:                uuid.NewV4(),
		ClusterId:         clusterId,
		MessageContractId: messageContractId,
		TopicId:           topicId,
		MessageType:       messageType,
		Description:       description,
		Subject:           subject,
		Schema:            schema,
		CreatedAt:         time.Now(),
		CompletedAt:       nil,
		SchemaVersion:     schemaVersion,
	}
}

func (*SchemaProcess) TableName() string {
	return "schema_process"
}

func (p *SchemaProcess) IsCompleted() bool {
	return p.CompletedAt != nil
}

func (p *SchemaProcess) MarkAsCompleted() {
	if p.IsCompleted() {
		return
	}

	now := time.Now()
	p.CompletedAt = &now
}
