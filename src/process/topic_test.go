package process

import (
	"context"
	"github.com/dfds/confluent-gateway/mocks"
	"github.com/dfds/confluent-gateway/models"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTopicService_CreateTopic(t *testing.T) {
	const clusterId = "some-cluster"
	const topicName = "some-topic"
	const partitions = 3
	tests := []struct {
		name          string
		retention     time.Duration
		wantErr       assert.ErrorAssertionFunc
		wantRetention int64
	}{
		{
			name:          "retention forever",
			retention:     -1 * time.Millisecond,
			wantErr:       assert.NoError,
			wantRetention: -1,
		},
		{
			name:          "retention 7 days",
			retention:     7 * 24 * time.Hour,
			wantErr:       assert.NoError,
			wantRetention: 604800000,
		},
		{
			name:          "retention 31 days",
			retention:     31 * 24 * time.Hour,
			wantErr:       assert.NoError,
			wantRetention: 2678400000,
		},
		{
			name:          "retention 365 days",
			retention:     365 * 24 * time.Hour,
			wantErr:       assert.NoError,
			wantRetention: 31536000000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spy := &mocks.MockClient{}
			sut := NewTopicService(context.TODO(), spy)
			topic, err := models.NewTopicDescription(topicName, partitions, models.RetentionFromDuration(tt.retention))

			tt.wantErr(t, sut.CreateTopic(clusterId, topic))

			assert.NoError(t, err)
			assert.Equal(t, clusterId, spy.GotClusterId)
			assert.Equal(t, topicName, spy.GotName)
			assert.Equal(t, partitions, spy.GotPartitions)
			assert.Equal(t, tt.wantRetention, spy.GotRetention)
		})
	}
}
