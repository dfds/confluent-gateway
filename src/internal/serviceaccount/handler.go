package serviceaccount

import (
	"context"
	"fmt"

	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/messaging"
)

type handler struct {
	process Process
}

func ServiceAccountAccessRequestedHandler(process Process) messaging.MessageHandler {
	return &handler{process: process}
}

type Process interface {
	Process(context.Context, ProcessInput) error
}

func (h *handler) Handle(ctx context.Context, msgContext messaging.MessageContext) error {
	switch message := msgContext.Message().(type) {

	case *ServiceAccountAccessRequested:
		input := ProcessInput{
			CapabilityId: models.CapabilityId(message.GetCapabilityId()),
			ClusterId:    models.ClusterId(message.GetClusterId()),
		}
		return h.process.Process(ctx, input)

	default:
		return fmt.Errorf("unknown message %#v", message)
	}
}
