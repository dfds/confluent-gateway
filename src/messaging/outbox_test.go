package messaging

import (
	"github.com/dfds/confluent-gateway/logging"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOutbox_Produce(t *testing.T) {
	tests := []struct {
		name     string
		registry OutgoingMessageRegistry
		msg      interface{}
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			name:    "ok",
			msg:     nil,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spy := &outgoingRepositoryMock{}

			registry := NewOutgoingMessageRegistry()
			registry.RegisterMessage("some-topic", "some-event-type", &dummyOutgoingMessage{})

			p := NewOutbox(logging.NilLogger(), registry, spy, func() string { return "some-message-id" })
			tt.wantErr(t, p.Produce(&dummyOutgoingMessage{
				Id:   1,
				Name: "a-name",
			}))

			assert.Equal(t, "some-topic", spy.entry.Topic)
			assert.Equal(t, "some-partition-key", spy.entry.Key)
			assert.Equal(t, "{\"messageId\":\"some-message-id\",\"type\":\"some-event-type\",\"data\":{\"id\":1,\"name\":\"a-name\"}}", spy.entry.Payload)
			assert.Nil(t, spy.entry.ProcessedUtc)

		})
	}
}

type outgoingRepositoryMock struct {
	entry *OutboxEntry
}

func (m *outgoingRepositoryMock) AddToOutbox(entry *OutboxEntry) error {
	m.entry = entry
	return nil
}
