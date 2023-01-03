package messaging

import (
	"errors"
	"fmt"
)

type OutgoingMessageRegistry interface {
	RegisterMessage(topicName string, eventType string, message OutgoingMessage) *OutgoingRegistration
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

type OutgoingRegistration struct {
	registry *outgoingMessageRegistry
	Error    error
}

func (r *OutgoingRegistration) RegisterMessage(topicName string, eventType string, message OutgoingMessage) *OutgoingRegistration {
	if r.Error != nil {
		return r
	}

	return r.registry.RegisterMessage(topicName, eventType, message)
}

func (r *outgoingMessageRegistry) RegisterMessage(topicName string, eventType string, message OutgoingMessage) *OutgoingRegistration {
	if len(topicName) == 0 {
		return r.error(errors.New("topic name must be specified"))
	}

	messageType, err := getMessageType(message)
	if err != nil {
		return r.error(err)
	}

	messageTypeName := messageType.Name()

	if _, ok := r.registrations[messageTypeName]; ok {
		return r.error(fmt.Errorf("duplicate message already registered for message of type: %s", eventType))
	}

	r.registrations[messageTypeName] = OutgoingMessageRegistration{
		eventType: eventType,
		topic:     topicName,
	}

	return r.ok()
}

func (r *outgoingMessageRegistry) ok() *OutgoingRegistration {
	return &OutgoingRegistration{
		registry: r,
		Error:    nil,
	}
}

func (r *outgoingMessageRegistry) error(err error) *OutgoingRegistration {
	return &OutgoingRegistration{
		registry: r,
		Error:    err,
	}
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
