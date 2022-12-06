package messaging

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewDispatcher(t *testing.T) {
	sut := NewDispatcher(&messageHandlerRegistryStub{}, NewDeserializerStub())

	assert.IsType(t, &dispatcher{}, sut)
}

func TestDispatch(t *testing.T) {
	spy := &messageHandlerSpy{}

	d := NewDispatcher(&messageHandlerRegistryStub{spy}, NewDeserializerStub())

	err := d.Dispatch(RawMessage{})

	assert.NoError(t, err)
	assert.True(t, spy.wasCalled)
}

func TestDispatchWithError(t *testing.T) {

	tests := []struct {
		name         string
		registry     MessageHandlerRegistry
		deserializer Deserializer
	}{
		{
			name:         "deserialization",
			registry:     &messageHandlerRegistryStub{},
			deserializer: &errorStub{error: errors.New("deserialization error")},
		},
		{
			name:         "registry",
			registry:     &errorStub{errors.New("get message handler error")},
			deserializer: NewDeserializerStub(),
		},
		{
			name:         "handler",
			registry:     &messageHandlerRegistryStub{&errorStub{errors.New("handler error")}},
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

type messageHandlerRegistryStub struct {
	handler MessageHandler
}

func (m *messageHandlerRegistryStub) GetMessageHandler(string) (MessageHandler, error) {
	return m.handler, nil
}

type errorStub struct {
	error error
}

func (s *errorStub) GetMessageHandler(string) (MessageHandler, error) {
	return nil, s.error
}

func (s *errorStub) Deserialize(RawMessage) (*IncomingMessage, error) {
	return nil, s.error
}

func (s *errorStub) Handle(MessageContext) error {
	return s.error
}

type messageHandlerSpy struct {
	wasCalled bool
}

func (h *messageHandlerSpy) Handle(MessageContext) error {
	h.wasCalled = true
	return nil
}

type deserializerStub struct {
	incomingMessage *IncomingMessage
}

func NewDeserializerStub() Deserializer {
	return &deserializerStub{&IncomingMessage{}}
}

func (d *deserializerStub) Deserialize(RawMessage) (*IncomingMessage, error) {
	return d.incomingMessage, nil
}

// endregion
