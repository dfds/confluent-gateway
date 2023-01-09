package messaging

import (
	"errors"
	"fmt"
)

type OutgoingMessageRegistry interface {
	RegisterMessage(topicName string, eventType string, message OutgoingMessage) error
	GetRegistration(message OutgoingMessage) (*OutgoingMessageRegistration, error)
}

func NewOutgoingMessageRegistry() OutgoingMessageRegistry {
	return &outgoingMessageRegistry{
		registrations: make(map[string]OutgoingMessageRegistration),
	}
}

type OutgoingMessage interface {
	PartitionKey() string
}

type OutgoingMessageRegistration struct {
	eventType string
	topic     string
}

type outgoingMessageRegistry struct {
	registrations map[string]OutgoingMessageRegistration
}

func (r *outgoingMessageRegistry) RegisterMessage(topicName string, eventType string, message OutgoingMessage) error {
	if len(topicName) == 0 {
		return errors.New("topic name must be specified")
	}

	messageType, err := getMessageType(message)
	if err != nil {
		return err
	}

	messageTypeName := messageType.Name()

	if _, ok := r.registrations[messageTypeName]; ok {
		return fmt.Errorf("duplicate message already registered for message of type: %s", eventType)
	}

	r.registrations[messageTypeName] = OutgoingMessageRegistration{
		eventType: eventType,
		topic:     topicName,
	}

	return nil
}

func (r *outgoingMessageRegistry) GetRegistration(message OutgoingMessage) (*OutgoingMessageRegistration, error) {
	messageType, err := getMessageType(message)
	if err != nil {
		return nil, err
	}

	if registration, ok := r.registrations[messageType.Name()]; !ok {
		return nil, fmt.Errorf("unknown outgoing message of type %s", messageType)
	} else {
		return &registration, nil
	}
}
