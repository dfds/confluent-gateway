package create

import (
	"context"
	"errors"
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/messaging"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTopicRequestedHandler_Handle(t *testing.T) {
	tests := []struct {
		name             string
		process          *processStub
		msgContext       messaging.MessageContext
		wantCapabilityId models.CapabilityId
		wantClusterId    models.ClusterId
		wantTopicName    string
		wantPartition    int
		wantRetention    time.Duration
		wantErr          assert.ErrorAssertionFunc
	}{
		{
			name:    "process ok",
			process: &processStub{},
			msgContext: messaging.NewMessageContext(map[string]string{}, &TopicRequested{
				CapabilityId: string(someCapabilityId),
				ClusterId:    string(someClusterId),
				TopicName:    someTopicName,
				Partitions:   1,
				Retention:    "-1",
			}),
			wantCapabilityId: someCapabilityId,
			wantClusterId:    someClusterId,
			wantTopicName:    someTopicName,
			wantPartition:    1,
			wantRetention:    -1 * time.Millisecond,
			wantErr:          assert.NoError,
		},
		{
			name:       "bad retention",
			process:    &processStub{},
			msgContext: messaging.NewMessageContext(map[string]string{}, &TopicRequested{Retention: "1y"}),
			wantErr:    assert.Error,
		},
		{
			name:       "process fail",
			process:    &processStub{err: errors.New("fail")},
			msgContext: messaging.NewMessageContext(map[string]string{}, &TopicRequested{}),
			wantErr:    assert.Error,
		},
		{
			name:       "unknown message",
			process:    &processStub{},
			msgContext: messaging.NewMessageContext(map[string]string{}, "bad message"),
			wantErr:    assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTopicRequestedHandler(tt.process)
			tt.wantErr(t, h.Handle(context.TODO(), tt.msgContext))
			assert.Equal(t, tt.wantCapabilityId, tt.process.input.CapabilityId)
			assert.Equal(t, tt.wantClusterId, tt.process.input.ClusterId)
			assert.Equal(t, tt.wantTopicName, tt.process.input.Topic.Name)
			assert.Equal(t, tt.wantPartition, tt.process.input.Topic.Partitions)
			assert.Equal(t, tt.wantRetention, tt.process.input.Topic.Retention)
		})
	}
}

type processStub struct {
	input ProcessInput
	err   error
}

func (t *processStub) Process(_ context.Context, input ProcessInput) error {
	t.input = input
	return t.err
}
