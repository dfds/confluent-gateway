package messaging

import (
	"encoding/json"
	"fmt"
	"github.com/dfds/confluent-gateway/logging"
	"github.com/satori/go.uuid"
	"time"
)

type Outbox struct {
	logger   logging.Logger
	registry OutgoingMessageRegistry
	repo     OutboxRepository
}

type OutboxRepository interface {
	AddToOutbox(*OutboxEntry) error
}

func NewOutbox(logger logging.Logger, registry OutgoingMessageRegistry, repo OutboxRepository) *Outbox {
	return &Outbox{
		logger:   logger,
		registry: registry,
		repo:     repo,
	}
}

func (p *Outbox) Produce(msg OutgoingMessage) error {
	registration, err := p.registry.GetRegistration(msg)
	if err != nil {
		return err
	}

	p.logger.Trace("Producing outgoing message {OutgoingMessage} ({EventType}) to {Topic}", fmt.Sprintf("%v", msg), registration.eventType, registration.topic)

	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	entry := &OutboxEntry{
		Id:          uuid.NewV4(),
		Topic:       registration.topic,
		Key:         msg.PartitionKey(),
		Payload:     string(payload),
		OccurredUtc: time.Now(),
	}

	return p.repo.AddToOutbox(entry)
}

type OutboxEntry struct {
	Id           uuid.UUID  `gorm:"type:uuid;primarykey;column:Id"`
	Topic        string     `gorm:"column:Topic"`
	Key          string     `gorm:"column:Key"`
	Payload      string     `gorm:"column:Payload"`
	OccurredUtc  time.Time  `gorm:"column:OccurredUtc"`
	ProcessedUtc *time.Time `gorm:"column:ProcessedUtc"`
}

func (*OutboxEntry) TableName() string {
	return "_outbox"
}

func (p *OutboxEntry) MarkAsProcessed() {
	if p.ProcessedUtc != nil {
		return
	}

	now := time.Now()
	p.ProcessedUtc = &now
}
