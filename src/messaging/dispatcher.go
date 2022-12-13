package messaging

import "context"

type RawMessage struct {
	Key     string
	Headers map[string]string
	Data    []byte
}

type Dispatcher interface {
	Dispatch(context.Context, RawMessage) error
}

type MessageHandlerRegistry interface {
	GetMessageHandler(messageType string) (MessageHandler, error)
}

func NewDispatcher(registry MessageHandlerRegistry, deserializer Deserializer) Dispatcher {
	return &dispatcher{
		registry:     registry,
		deserializer: deserializer,
	}
}

type MessageHandler interface {
	Handle(context.Context, MessageContext) error
}

type dispatcher struct {
	registry     MessageHandlerRegistry
	deserializer Deserializer
}

func (d *dispatcher) Dispatch(ctx context.Context, msg RawMessage) error {
	incomingMessage, err := d.deserializer.Deserialize(msg)
	if err != nil {
		return err
	}

	handler, err := d.registry.GetMessageHandler(incomingMessage.Type)
	if err != nil {
		return err
	}

	return handler.Handle(ctx, NewMessageContext(msg.Headers, incomingMessage.Message))
}
