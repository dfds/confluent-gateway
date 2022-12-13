package messaging

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type MessageRegistry interface {
	RegisterMessageHandler(topicName string, eventType string, handler MessageHandler, message interface{}) *Registration
	GetMessageHandler(messageType string) (MessageHandler, error)
	GetMessageType(messageType string) (reflect.Type, error)
	GetTopics() []string
}

func NewMessageRegistry() MessageRegistry {
	return &messageRegistry{
		registrations: make(map[string]MessageRegistration),
		topics:        make(map[string]string),
	}
}

type MessageRegistration struct {
	messageHandler MessageHandler
	message        interface{}
	messageType    reflect.Type
}

type messageRegistry struct {
	registrations map[string]MessageRegistration
	topics        map[string]string
}

type Registration struct {
	registry *messageRegistry
	Error    error
}

func (r *Registration) RegisterMessageHandler(topicName string, eventType string, handler MessageHandler, message interface{}) *Registration {
	if r.Error != nil {
		return r
	}

	return r.registry.RegisterMessageHandler(topicName, eventType, handler, message)
}

func (r *messageRegistry) RegisterMessageHandler(topicName string, eventType string, handler MessageHandler, message interface{}) *Registration {
	if len(topicName) == 0 {
		return r.error(errors.New("topic name must be specified"))
	}

	if _, ok := r.registrations[eventType]; ok {
		return r.error(fmt.Errorf("duplicate message handler registration for message of type: %s", eventType))
	}

	messageType, err := getMessageType(message)
	if err != nil {
		return r.error(err)
	}

	r.registrations[eventType] = MessageRegistration{
		messageHandler: handler,
		messageType:    messageType,
		message:        message,
	}

	r.topics[strings.ToUpper(topicName)] = topicName

	return r.ok()
}

func (r *messageRegistry) ok() *Registration {
	return &Registration{
		registry: r,
		Error:    nil,
	}
}

func (r *messageRegistry) error(err error) *Registration {
	return &Registration{
		registry: r,
		Error:    err,
	}
}

func getMessageType(example interface{}) (reflect.Type, error) {
	if example == nil {
		return nil, fmt.Errorf("example cannot be nil")
	}

	typeOf := reflect.TypeOf(example)
	kind := typeOf.Kind()

	if kind == reflect.Ptr {
		typeOf = typeOf.Elem()
		kind = typeOf.Kind()
	}

	switch kind {
	case reflect.Struct:
		return typeOf, nil
	default:
		return nil, fmt.Errorf("example '%#v' is not a struct or pointer to a struct", example)
	}
}

func (r *messageRegistry) GetMessageHandler(messageType string) (MessageHandler, error) {
	if registration, err := r.getMessageRegistration(messageType); err != nil {
		return nil, err
	} else {
		return registration.messageHandler, nil
	}
}

func (r *messageRegistry) getMessageRegistration(messageType string) (*MessageRegistration, error) {
	if registration, ok := r.registrations[messageType]; ok {
		return &registration, nil
	}
	return nil, fmt.Errorf("unknown message of type %s", messageType)
}

func (r *messageRegistry) GetMessageType(messageType string) (reflect.Type, error) {
	if registration, err := r.getMessageRegistration(messageType); err != nil {
		return nil, err
	} else {
		return registration.messageType, nil
	}
}

func (r *messageRegistry) GetTopics() []string {
	var topics []string

	for _, t := range r.topics {
		topics = append(topics, t)
	}

	return topics
}
