package create

import (
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/logging"
	"github.com/dfds/confluent-gateway/messaging"
)

type StepContext struct {
	logger        logging.Logger
	state         *models.DeleteProcess
	topicService  TopicService
	schemaService SchemaService
	outbox        Outbox
}

func NewStepContext(logger logging.Logger, state *models.DeleteProcess, topicService TopicService, schemaService SchemaService, outbox Outbox) *StepContext {
	return &StepContext{logger: logger, state: state, topicService: topicService, schemaService: schemaService, outbox: outbox}
}

type TopicService interface {
	DeleteTopic(string) error
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
	return c.topicService.DeleteTopic(c.state.TopicId)
}

func (c *StepContext) DeleteSchemasByTopicId() error {
	return c.schemaService.DeleteSchemasByTopicId(c.state.TopicId)
}
func (c *StepContext) MarkSchemasAsDeleted() {
	c.state.MarkSchemasAsDeleted()
}
func (c *StepContext) AreSchemasDeleted() bool {
	return c.state.AreSchemasDeleted()
}

func (c *StepContext) MarkAsCompleted() {
	c.state.MarkAsCompleted()
}

func (c *StepContext) RaiseTopicDeletedEvent() error {
	event := &TopicDeleted{ // TODO: Figure out if we want this event
		TopicId: c.state.TopicId,
	}
	return c.outbox.Produce(event)
}
