package messaging

import (
	"context"
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
		topic   string
		message interface{}
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "pointer to struct",
			topic:   "some_topic",
			message: &dummyMessage{},
			wantErr: assert.NoError,
		},
		{
			name:    "struct",
			topic:   "some_topic",
			message: dummyMessage{},
			wantErr: assert.NoError,
		},
		{
			name:    "illegal type",
			topic:   "some_topic",
			message: make(map[string]string),
			wantErr: assert.Error,
		},
		{
			name:    "no topic",
			topic:   "",
			message: &dummyMessage{},
			wantErr: assert.Error,
		},
		{
			name:    "nil",
			topic:   "some_topic",
			message: nil,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := NewMessageRegistry()

			err := sut.RegisterMessageHandler(tt.topic, "some_event", &dummyMessageHandler{}, tt.message).Error

			tt.wantErr(t, err, fmt.Sprintf("RegisterMessageHandler(%v, dummyMessage, %v)", "some_event", tt.message))
		})
	}
}

func TestRegisterMessageHandlerWithDuplicateRegistrations(t *testing.T) {
	sut := NewMessageRegistry()
	err := sut.
		RegisterMessageHandler("some_topic", "some_event", &dummyMessageHandler{}, &dummyMessage{}).
		RegisterMessageHandler("some_topic", "some_event", &dummyMessageHandler{}, &dummyMessage{}).
		Error

	assert.Error(t, err)
}

func TestGetMessageHandler(t *testing.T) {
	dummyHandler := &dummyMessageHandler{}

	sut := NewMessageRegistry()
	_ = sut.RegisterMessageHandler("some_topic", "some_event", dummyHandler, &dummyMessage{})

	handler, err := sut.GetMessageHandler("some_event")

	assert.NoError(t, err)
	assert.Same(t, dummyHandler, handler)
}

func TestGetMessageHandlerWithUnregisteredMessageType(t *testing.T) {
	sut := NewMessageRegistry()
	_ = sut.RegisterMessageHandler("some_topic", "some_event", &dummyMessageHandler{}, &dummyMessage{})

	handler, err := sut.GetMessageHandler("another_event")

	assert.Error(t, err)
	assert.Nil(t, handler)
}

func TestGetMessageType(t *testing.T) {
	sut := NewMessageRegistry()
	_ = sut.RegisterMessageHandler("some_topic", "some_event", &dummyMessageHandler{}, &dummyMessage{})

	messageType, err := sut.GetMessageType("some_event")

	assert.NoError(t, err)
	assert.Equal(t, reflect.TypeOf(dummyMessage{}), messageType)
}

func TestGetMessageTypeWithUnregisteredMessageType(t *testing.T) {
	sut := NewMessageRegistry()
	_ = sut.RegisterMessageHandler("some_topic", "some_event", &dummyMessageHandler{}, &dummyMessage{})

	messageType, err := sut.GetMessageType("another_event")

	assert.Error(t, err)
	assert.Nil(t, messageType)
}

func TestGetTopics(t *testing.T) {
	sut := NewMessageRegistry()
	_ = sut.RegisterMessageHandler("topicA", "some_event", &dummyMessageHandler{}, &dummyMessage{})
	_ = sut.RegisterMessageHandler("topicB", "another_event", &dummyMessageHandler{}, &dummyMessage{})

	topics := sut.GetTopics()

	assert.ElementsMatch(t, []string{"topicA", "topicB"}, topics)
}

// region Test Doubles

type dummyMessageHandler struct {
}

func (d *dummyMessageHandler) Handle(context.Context, MessageContext) error {
	return nil
}

type dummyMessage struct {
}

// endregion
