package messaging

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestNewMessageRegistry(t *testing.T) {
	sut := NewMessageRegistry()

	assert.IsType(t, &messageRegistry{}, sut)
}

func TestRegisterMessageHandler(t *testing.T) {
	tests := []struct {
		name    string
		message interface{}
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "pointer to struct",
			message: &dummyMessage{},
			wantErr: assert.NoError,
		},
		{
			name:    "struct",
			message: dummyMessage{},
			wantErr: assert.NoError,
		},
		{
			name:    "illegal type",
			message: make(map[string]string),
			wantErr: assert.Error,
		},
		{
			name:    "nil",
			message: nil,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := NewMessageRegistry()

			tt.wantErr(t, sut.RegisterMessageHandler("some_event", &dummyMessageHandler{}, tt.message), fmt.Sprintf("RegisterMessageHandler(%v, dummyMessage, %v)", "some_event", tt.message))
		})
	}
}

func TestRegisterMessageHandlerWithDuplicateRegistrations(t *testing.T) {

	sut := NewMessageRegistry()
	_ = sut.RegisterMessageHandler("some_event", &dummyMessageHandler{}, &dummyMessage{})
	err := sut.RegisterMessageHandler("some_event", &dummyMessageHandler{}, &dummyMessage{})

	assert.Error(t, err)
}

func TestGetMessageHandler(t *testing.T) {
	dummyHandler := &dummyMessageHandler{}

	sut := NewMessageRegistry()
	_ = sut.RegisterMessageHandler("some_event", dummyHandler, &dummyMessage{})

	handler, err := sut.GetMessageHandler("some_event")

	assert.NoError(t, err)
	assert.Same(t, dummyHandler, handler)
}

func TestGetMessageHandlerWithUnregisteredMessageType(t *testing.T) {
	dummyHandler := &dummyMessageHandler{}

	sut := NewMessageRegistry()
	_ = sut.RegisterMessageHandler("some_event", dummyHandler, &dummyMessage{})

	handler, err := sut.GetMessageHandler("another_event")

	assert.Error(t, err)
	assert.Nil(t, handler)
}

func TestGetMessageType(t *testing.T) {
	sut := NewMessageRegistry()
	_ = sut.RegisterMessageHandler("some_event", &dummyMessageHandler{}, &dummyMessage{})

	messageType, err := sut.GetMessageType("some_event")

	assert.NoError(t, err)
	assert.Equal(t, reflect.TypeOf(dummyMessage{}), messageType)
}

func TestGetMessageTypeWithUnregisteredMessageType(t *testing.T) {
	dummyHandler := &dummyMessageHandler{}

	sut := NewMessageRegistry()
	_ = sut.RegisterMessageHandler("some_event", dummyHandler, &dummyMessage{})

	messageType, err := sut.GetMessageType("another_event")

	assert.Error(t, err)
	assert.Nil(t, messageType)
}

// region Test Doubles

type dummyMessageHandler struct {
}

func (d *dummyMessageHandler) Handle(MessageContext) error {
	return nil
}

type dummyMessage struct {
}

// endregion
