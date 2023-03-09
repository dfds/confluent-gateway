package create

import (
	"context"
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/logging"
	"github.com/dfds/confluent-gateway/messaging"
)

type StepContext struct {
	logger   logging.Logger
	ctx      context.Context
	state    *models.SchemaProcess
	registry SchemaRegistry
	outbox   Outbox
}

func NewStepContext(logger logging.Logger, ctx context.Context, schema *models.SchemaProcess, registry SchemaRegistry, outbox Outbox) *StepContext {
	return &StepContext{logger: logger, ctx: ctx, state: schema, registry: registry, outbox: outbox}
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

func (c *StepContext) RegisterSchema() error {
	return c.registry.RegisterSchema(c.ctx, c.state.ClusterId, c.state.Subject, c.state.Schema)
}

func (c *StepContext) MarkAsCompleted() {
	c.state.MarkAsCompleted()
}

func (c *StepContext) RaiseSchemaRegisteredEvent() error {
	event := &SchemaRegistered{
		MessageContractId: c.state.MessageContractId,
	}
	return c.outbox.Produce(event)
}
