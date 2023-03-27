package create

import (
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/logging"
	"github.com/dfds/confluent-gateway/messaging"
)

type StepContext struct {
	logger logging.Logger
	state  *models.DeleteProcess
	topic  TopicService
	outbox Outbox
}

func NewStepContext(logger logging.Logger, state *models.DeleteProcess, topic TopicService, outbox Outbox) *StepContext {
	return &StepContext{logger: logger, state: state, topic: topic, outbox: outbox}
}

type TopicService interface {
	DeleteTopic(models.ClusterId, string, string) error
}

type Outbox interface {
	Produce(msg messaging.OutgoingMessage) error
}

type OutboxRepository interface {
	AddToOutbox(entry *messaging.OutboxEntry) error
}

type OutboxFactory func(repository OutboxRepository) Outbox

func (c *StepContext) IsCompleted() bool {
	return c.state.IsCompleted()
}

func (c *StepContext) DeleteTopic() error {
	return c.topic.DeleteTopic(c.state.ClusterId, c.state.TopicId, c.state.TopicName)
}

func (c *StepContext) MarkAsCompleted() {
	c.state.MarkAsCompleted()
}

func (c *StepContext) RaiseTopicDeletedEvent() error {
	event := &TopicDeleted{
		TopicId:      c.state.TopicId,
		CapabilityId: string(c.state.CapabilityId),
		ClusterId:    string(c.state.ClusterId),
		TopicName:    c.state.TopicName,
	}
	return c.outbox.Produce(event)
}
