package messaging

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestNewDispatcher(t *testing.T) {
	sut := NewDispatcher(NewMessageRegistryStub(), NewDeserializerStub())

	assert.IsType(t, &dispatcher{}, sut)
}

func TestDispatch(t *testing.T) {
	spy := &messageHandlerMock{}
	d := NewDispatcher(NewMessageRegistryWithHandlerStub(spy), NewDeserializerStub())

	err := d.Dispatch(RawMessage{})

	assert.NoError(t, err)
	assert.True(t, spy.wasCalled)
}

func TestDispatchWithError(t *testing.T) {

	tests := []struct {
		name         string
		registry     MessageRegistry
		deserializer Deserializer
	}{
		{
			name:         "deserialization",
			registry:     NewMessageRegistryStub(),
			deserializer: NewDeserializerWithErrorStub(errors.New("deserialization error")),
		},
		{
			name:         "registry",
			registry:     NewMessageRegistryWithErrorStub(errors.New("get message handler error")),
			deserializer: NewDeserializerStub(),
		},
		{
			name:         "handler",
			registry:     NewMessageRegistryWithHandlerStub(NewMessageHandlerWithErrorStub(errors.New("handler error"))),
			deserializer: NewDeserializerStub(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDispatcher(tt.registry, tt.deserializer)

			assert.Error(t, d.Dispatch(RawMessage{}))
		})
	}
}

// region Test Doubles

type messageRegistryStub struct {
	handler MessageHandler
	error   error
}

func NewMessageRegistryStub() MessageRegistry {
	return &messageRegistryStub{}
}

func NewMessageRegistryWithHandlerStub(handler MessageHandler) MessageRegistry {
	return &messageRegistryStub{handler: handler}
}

func NewMessageRegistryWithErrorStub(error error) MessageRegistry {
	return &messageRegistryStub{error: error}
}

func (m *messageRegistryStub) RegisterMessageHandler(string, MessageHandler, interface{}) error {
	panic("implement me")
}

func (m *messageRegistryStub) GetMessageHandler(string) (MessageHandler, error) {
	return m.handler, m.error
}

func (m *messageRegistryStub) GetMessageType(string) (reflect.Type, error) {
	panic("implement me")
}

type messageHandlerMock struct {
	wasCalled bool
	error     error
}

func NewMessageHandlerWithErrorStub(err error) MessageHandler {
	return &messageHandlerMock{error: err}
}

func (h *messageHandlerMock) Handle(MessageContext) error {
	h.wasCalled = true
	return h.error
}

type deserializerStub struct {
	incomingMessage *IncomingMessage
	error           error
}

func NewDeserializerStub() Deserializer {
	return &deserializerStub{incomingMessage: &IncomingMessage{}}
}

func NewDeserializerWithErrorStub(err error) Deserializer {
	return &deserializerStub{error: err}
}

func (d *deserializerStub) Deserialize(RawMessage) (*IncomingMessage, error) {
	return d.incomingMessage, d.error
}

// endregion
