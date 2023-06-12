package create

import (
	"fmt"

	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/logging"
	"github.com/dfds/confluent-gateway/messaging"
)

type StepContext struct {
	logger  logging.Logger
	state   *models.CreateProcess
	account AccountService
	topic   TopicService
	outbox  Outbox
}

func NewStepContext(logger logging.Logger, state *models.CreateProcess, account AccountService, topic TopicService, outbox Outbox) *StepContext {
	return &StepContext{logger: logger, state: state, account: account, topic: topic, outbox: outbox}
}

type AccountService interface {
	HasClusterAccess(capabilityId models.CapabilityId, clusterId models.ClusterId) (bool, error)
}

type TopicService interface {
	CreateTopic(models.CapabilityId, models.ClusterId, string, models.TopicDescription) error
}

type Outbox interface {
	Produce(msg messaging.OutgoingMessage) error
}

type OutboxRepository interface {
	AddToOutbox(entry *messaging.OutboxEntry) error
}

type OutboxFactory func(repository OutboxRepository) Outbox

func (c *StepContext) HasClusterAccess() bool {
	exists, err := c.account.HasClusterAccess(c.state.CapabilityId, c.state.ClusterId)
	fmt.Printf("Service account for CapabilityId: %s\n\tFound: %t\n\tError: %t\n", c.state.CapabilityId, exists, err != nil)
	if err != nil {
		return false
	}
	return exists
}

func (c *StepContext) IsCompleted() bool {
	return c.state.IsCompleted()
}

func (c *StepContext) CreateTopic() error {
	return c.topic.CreateTopic(c.state.CapabilityId, c.state.ClusterId, c.state.TopicId, c.state.TopicDescription())
}

func (c *StepContext) MarkAsCompleted() {
	c.state.MarkAsCompleted()
}

func (c *StepContext) RaiseTopicProvisionedEvent() error {
	event := &TopicProvisioned{
		TopicId:      c.state.TopicId,
		CapabilityId: string(c.state.CapabilityId),
		ClusterId:    string(c.state.ClusterId),
		TopicName:    c.state.TopicName,
	}
	return c.outbox.Produce(event)
}
