package messaging

import (
	"encoding/json"
	"fmt"
	"github.com/dfds/confluent-gateway/logging"
	"github.com/satori/go.uuid"
	"time"
)

func ConfigureOutbox(logger logging.Logger, options ...OutboxOption) (OutboxFactory, error) {
	outgoingRegistry := NewOutgoingMessageRegistry()

	cfg := &outboxConfig{
		registry:  outgoingRegistry,
		generator: func() string { return uuid.NewV4().String() },
	}

	for _, option := range options {
		if err := option.apply(cfg); err != nil {
			return nil, err
		}
	}

	outboxFactory := func(repository OutboxRepository) *Outbox {
		return NewOutbox(logger, outgoingRegistry, repository, defaultMessageIdGenerator)
	}
	return outboxFactory, nil
}

type OutboxFactory = func(repository OutboxRepository) *Outbox

type outboxConfig struct {
	registry  OutgoingMessageRegistry
	generator MessageIdGenerator
}

type OutboxOption interface {
	apply(cfg *outboxConfig) error
}

type messageOption struct {
	topicName string
	eventType string
	message   OutgoingMessage
}

func (o messageOption) apply(cfg *outboxConfig) error {
	return cfg.registry.RegisterMessage(o.topicName, o.eventType, o.message)
}

func RegisterMessage(topicName string, eventType string, message OutgoingMessage) OutboxOption {
	return messageOption{
		topicName: topicName,
		eventType: eventType,
		message:   message,
	}
}

type Outbox struct {
	logger    logging.Logger
	registry  OutgoingMessageRegistry
	repo      OutboxRepository
	generator MessageIdGenerator
}

type MessageIdGenerator func() string

func defaultMessageIdGenerator() string {
	return uuid.NewV4().String()
}

type OutboxRepository interface {
	AddToOutbox(*OutboxEntry) error
}

func NewOutbox(logger logging.Logger, registry OutgoingMessageRegistry, repo OutboxRepository, generator MessageIdGenerator) *Outbox {
	return &Outbox{
		logger:    logger,
		registry:  registry,
		repo:      repo,
		generator: generator,
	}
}

func (p *Outbox) Produce(msg OutgoingMessage) error {
	registration, err := p.registry.GetRegistration(msg)
	if err != nil {
		return err
	}

	p.logger.Debug("Producing outgoing message {OutgoingMessage} ({EventType}) to {Topic}", fmt.Sprintf("%v", msg), registration.eventType, registration.topic)

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(envelope{
		MessageId: p.generator(),
		Type:      registration.eventType,
		Data:      data,
	})

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
