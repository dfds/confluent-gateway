package create

import (
	"context"
	"errors"
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/messaging"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTopicRequestedHandler_Handle(t *testing.T) {
	tests := []struct {
		name             string
		process          *processStub
		msgContext       messaging.MessageContext
		wantCapabilityId models.CapabilityId
		wantClusterId    models.ClusterId
		wantTopicName    string
		wantErr          assert.ErrorAssertionFunc
	}{
		{
			name:    "process ok",
			process: &processStub{},
			msgContext: messaging.NewMessageContext(map[string]string{}, &TopicDeletionRequested{
				CapabilityId: string(someCapabilityId),
				ClusterId:    string(someClusterId),
				TopicName:    someTopicName,
			}),
			wantCapabilityId: someCapabilityId,
			wantClusterId:    someClusterId,
			wantTopicName:    someTopicName,
			wantErr:          assert.NoError,
		},
		{
			name:       "process fail",
			process:    &processStub{err: errors.New("fail")},
			msgContext: messaging.NewMessageContext(map[string]string{}, &TopicDeletionRequested{}),
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
			assert.Equal(t, tt.wantTopicName, tt.process.input.TopicName)
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
