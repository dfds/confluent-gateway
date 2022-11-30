package messaging

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type Deserializer interface {
	Deserialize(RawMessage) (*IncomingMessage, error)
}

type IncomingMessage struct {
	MessageId string
	Type      string
	Headers   map[string]string
	Message   interface{}
}

func NewDefaultDeserializer(registry MessageRegistry) Deserializer {
	return &deserializer{registry}
}

type deserializer struct {
	registry MessageRegistry
}

func (d *deserializer) Deserialize(msg RawMessage) (*IncomingMessage, error) {
	var envelope envelope

	headers, err := deserializerHeaders(msg)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(msg.Data, &envelope); err != nil {
		return nil, err
	}

	messageType, err := d.registry.GetMessageType(envelope.Type)
	if err != nil {
		return nil, err
	}

	message := reflect.New(messageType).Interface()

	if err := json.Unmarshal(envelope.Data, message); err != nil {
		return nil, err
	}

	return &IncomingMessage{
		MessageId: envelope.MessageId,
		Type:      envelope.Type,
		Headers:   headers,
		Message:   message,
	}, nil
}

var ignoreHeaders = [...]string{"messageId", "type", "data"}

func deserializerHeaders(msg RawMessage) (map[string]string, error) {
	var envelope map[string]interface{}

	if err := json.Unmarshal(msg.Data, &envelope); err != nil {
		return nil, err
	}

	for _, header := range ignoreHeaders {
		delete(envelope, header)
	}

	headers := make(map[string]string)

	for k, v := range msg.Headers {
		headers[k] = v
	}

	for k, v := range envelope {
		headers[k] = fmt.Sprint(v)
	}

	return headers, nil
}

type envelope struct {
	MessageId string          `json:"messageId"`
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data"`
}
