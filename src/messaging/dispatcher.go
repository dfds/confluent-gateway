package messaging

type RawMessage struct {
	Key     string
	Headers map[string]string
	Data    []byte
}

type Dispatcher interface {
	Dispatch(RawMessage) error
}

func NewDispatcher(registry MessageRegistry, deserializer Deserializer) Dispatcher {
	return &dispatcher{
		registry:     registry,
		deserializer: deserializer,
	}
}

type MessageHandler interface {
	Handle(MessageContext) error
}

type dispatcher struct {
	registry     MessageRegistry
	deserializer Deserializer
}

func (d *dispatcher) Dispatch(msg RawMessage) error {
	incomingMessage, err := d.deserializer.Deserialize(msg)
	if err != nil {
		return err
	}

	handler, err := d.registry.GetMessageHandler(incomingMessage.Type)
	if err != nil {
		return err
	}

	return handler.Handle(NewMessageContext(msg.Headers, incomingMessage.Message))
}
