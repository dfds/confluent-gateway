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

	case *TopicDeletionRequested:
		input := ProcessInput{
			CapabilityId: models.CapabilityId(message.CapabilityId),
			ClusterId:    models.ClusterId(message.ClusterId),
			TopicId:      message.TopicId,
			TopicName:    message.TopicName,
		}
		return h.process.Process(ctx, input)

	default:
		return fmt.Errorf("unknown message %#v", message)
	}
}
