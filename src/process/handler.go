package process

import (
	"context"
	"fmt"
	"github.com/dfds/confluent-gateway/messaging"
	"github.com/dfds/confluent-gateway/models"
)

type handler struct {
	process CreateTopicProcess
}

func NewTopicRequestedHandler(process CreateTopicProcess) messaging.MessageHandler {
	return &handler{process: process}
}

type CreateTopicProcess interface {
	Process(context.Context, CreateTopicProcessInput) error
}

func (h *handler) Handle(ctx context.Context, msgContext messaging.MessageContext) error {
	switch message := msgContext.Message().(type) {

	case *TopicRequested:
		return h.process.Process(ctx, CreateTopicProcessInput{
			CapabilityRootId: models.CapabilityRootId(message.CapabilityRootId),
			ClusterId:        models.ClusterId(message.ClusterId),
			Topic:            models.NewTopic(message.TopicName, message.Partitions, message.Retention),
		})

	default:
		return fmt.Errorf("unknown message %#v", message)
	}
}
