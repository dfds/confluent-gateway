package messaging

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewOutgoingMessageRegistry(t *testing.T) {
	sut := NewOutgoingMessageRegistry()

	assert.IsType(t, &outgoingMessageRegistry{}, sut)
}

func TestRegisterOutgoingMessageHandler(t *testing.T) {
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
			sut := NewOutgoingMessageRegistry()

			err := sut.RegisterMessage(tt.topic, "some_event", tt.message).Error

			tt.wantErr(t, err, fmt.Sprintf("RegisterMessage(%v, dummyMessage, %v)", "some_event", tt.message))
		})
	}
}

func TestRegisterOutgoingMessageHandlerWithDuplicateRegistrations(t *testing.T) {
	sut := NewOutgoingMessageRegistry()
	err := sut.
		RegisterMessage("some_topic", "some_event", &dummyMessage{}).
		RegisterMessage("some_topic", "some_event", &dummyMessage{}).
		Error

	assert.Error(t, err)
}

func TestOutgoingMessageRegistry_GetRegistration(t *testing.T) {
	sut := NewOutgoingMessageRegistry()
	_ = sut.RegisterMessage("some_topic", "some_event", &dummyMessage{})

	registration, err := sut.GetRegistration(&dummyMessage{})

	assert.NoError(t, err)
	assert.Equal(t, "some_event", registration.eventType)
	assert.Equal(t, "some_topic", registration.topic)
}

func TestOutgoingMessageRegistry_GetRegistrationWithUnregisteredMessageType(t *testing.T) {
	sut := NewOutgoingMessageRegistry()
	_ = sut.RegisterMessage("some_topic", "some_event", &dummyMessage{})

	registration, err := sut.GetRegistration(&unregisteredDummyMessage{})

	assert.Error(t, err)
	assert.Nil(t, registration)
}

type unregisteredDummyMessage struct {
}
