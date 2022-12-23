package process

import (
	"context"
	"errors"
	"fmt"
	"github.com/dfds/confluent-gateway/messaging"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTopicRequestedHandler_Handle(t *testing.T) {
	tests := []struct {
		name       string
		process    CreateTopicProcess
		msgContext messaging.MessageContext
		wantErr    assert.ErrorAssertionFunc
	}{
		{
			name:       "process ok",
			process:    &processStub{},
			msgContext: messaging.NewMessageContext(map[string]string{}, &TopicRequested{Retention: "-1"}),
			wantErr:    assert.NoError,
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
			tt.wantErr(t, h.Handle(context.TODO(), tt.msgContext), fmt.Sprintf("Handle(context.TODO(), %v)", tt.msgContext))
		})
	}
}

func TestTopicRequestedHandler_Handle_Retention(t *testing.T) {
	tests := []struct {
		name          string
		retention     string
		wantErr       assert.ErrorAssertionFunc
		wantRetention time.Duration
	}{
		{
			name:          "no retention",
			retention:     "",
			wantErr:       assert.Error,
			wantRetention: 0,
		},
		{
			name:          "1 year",
			retention:     "1y",
			wantErr:       assert.Error,
			wantRetention: 0,
		},
		{
			name:          "365 days",
			retention:     "365d",
			wantErr:       assert.Error,
			wantRetention: 0,
		},
		{
			name:          "forever",
			retention:     "-1",
			wantErr:       assert.NoError,
			wantRetention: -1 * time.Millisecond,
		},
		{
			name:          "7 days",
			retention:     "604800000ms",
			wantErr:       assert.NoError,
			wantRetention: 1 * time.Hour * 24 * 7,
		},
		{
			name:          "31 days",
			retention:     "2678400000ms",
			wantErr:       assert.NoError,
			wantRetention: 1 * time.Hour * 24 * 31,
		},
		{
			name:          "365 days",
			retention:     "31536000000ms",
			wantErr:       assert.NoError,
			wantRetention: 1 * time.Hour * 24 * 365,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &processStub{}
			h := NewTopicRequestedHandler(stub)
			msgContext := messaging.NewMessageContext(map[string]string{}, &TopicRequested{Retention: tt.retention})
			tt.wantErr(t, h.Handle(context.TODO(), msgContext))

			assert.Equal(t, tt.wantRetention, stub.input.Topic.Retention)
		})
	}
}

type processStub struct {
	input CreateTopicProcessInput
	err   error
}

func (t *processStub) Process(_ context.Context, input CreateTopicProcessInput) error {
	t.input = input
	return t.err
}
