package create

import (
	"context"
	"fmt"
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/messaging"
)

type handler struct {
	process Process
}

func NewTopicRequestedHandler(process Process) messaging.MessageHandler {
	return &handler{process: process}
}

type Process interface {
	Process(context.Context, ProcessInput) error
}

func (h *handler) Handle(ctx context.Context, msgContext messaging.MessageContext) error {
	switch message := msgContext.Message().(type) {

	case *TopicRequested:
		topic, err := models.NewTopicDescription(message.GetTopicName(), message.Partitions, models.RetentionFromString(message.Retention))

		if err != nil {
			return err
		}

		input := ProcessInput{
			CapabilityId: models.CapabilityId(message.GetCapabilityId()),
			ClusterId:    models.ClusterId(message.GetClusterId()),
			Topic:        topic,
		}
		return h.process.Process(ctx, input)

	default:
		return fmt.Errorf("unknown message %#v", message)
	}
}
