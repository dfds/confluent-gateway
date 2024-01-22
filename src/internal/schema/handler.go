package schema

import (
	"context"
	"fmt"
	"github.com/dfds/confluent-gateway/messaging"
)

type handler struct {
	process Process
}

func NewSchemaAddedHandler(process Process) messaging.MessageHandler {
	return &handler{process: process}
}

type Process interface {
	Process(context.Context, ProcessInput) error
}

func (h *handler) Handle(ctx context.Context, msgContext messaging.MessageContext) error {
	switch message := msgContext.Message().(type) {

	case *MessageContractRequested:
		input := ProcessInput{
			MessageContractId: message.MessageContractId,
			TopicId:           message.TopicId,
			MessageType:       message.MessageType,
			Description:       message.Description,
			Schema:            message.Schema,
			SchemaVersion:     message.SchemaVersion,
		}
		return h.process.Process(ctx, input)

	default:
		return fmt.Errorf("unknown message %#v", message)
	}
}
