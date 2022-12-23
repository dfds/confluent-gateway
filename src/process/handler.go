package process

import (
	"context"
	"fmt"
	"github.com/dfds/confluent-gateway/messaging"
	"github.com/dfds/confluent-gateway/models"
	"time"
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
		retention, err := parseDuration(message.Retention)

		if err != nil {
			return fmt.Errorf("unable to parse duration: %w", err)
		}

		return h.process.Process(ctx, CreateTopicProcessInput{
			CapabilityRootId: models.CapabilityRootId(message.CapabilityRootId),
			ClusterId:        models.ClusterId(message.ClusterId),
			Topic:            models.NewTopic(message.TopicName, message.Partitions, retention),
		})

	default:
		return fmt.Errorf("unknown message %#v", message)
	}
}

func parseDuration(retention string) (time.Duration, error) {
	if retention == "-1" {
		return -1 * time.Millisecond, nil
	}

	duration, err := time.ParseDuration(retention)
	return duration, err
}
