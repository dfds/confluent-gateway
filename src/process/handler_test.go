package process

import (
	"context"
	"errors"
	"fmt"
	"github.com/dfds/confluent-gateway/messaging"
	"github.com/stretchr/testify/assert"
	"testing"
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
			msgContext: messaging.NewMessageContext(map[string]string{}, &TopicRequested{}),
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

type processStub struct {
	err error
}

func (t *processStub) Process(context.Context, CreateTopicProcessInput) error {
	return t.err
}
