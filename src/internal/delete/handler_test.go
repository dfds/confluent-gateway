package create

import (
	"context"
	"errors"
	"github.com/dfds/confluent-gateway/messaging"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTopicRequestedHandler_Handle(t *testing.T) {
	tests := []struct {
		name        string
		process     *processStub
		msgContext  messaging.MessageContext
		wantTopicId string
		wantErr     assert.ErrorAssertionFunc
	}{
		{
			name:    "process ok",
			process: &processStub{},
			msgContext: messaging.NewMessageContext(map[string]string{}, &TopicDeletionRequested{
				TopicId: someTopicId,
			}),
			wantTopicId: someTopicId,
			wantErr:     assert.NoError,
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
			assert.Equal(t, tt.wantTopicId, tt.process.input.TopicId)
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
