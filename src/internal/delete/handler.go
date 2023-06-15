package delete

import (
	"context"
	"fmt"
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
			TopicId: message.TopicId,
		}
		return h.process.Process(ctx, input)

	default:
		return fmt.Errorf("unknown message %#v", message)
	}
}
