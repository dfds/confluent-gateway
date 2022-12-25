package process

import (
	"context"
	"fmt"
	"github.com/dfds/confluent-gateway/mocks"
	"github.com/dfds/confluent-gateway/models"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestProcess_createTopic(t *testing.T) {
	const clusterId = "some-cluster"
	const topicName = "some-topic"
	tests := []struct {
		name           string
		retention      time.Duration
		wantErr        assert.ErrorAssertionFunc
		wantClusterId  string
		wantName       string
		wantPartitions int
		wantRetention  int64
	}{
		{
			name:           "retention forever",
			retention:      -1 * time.Millisecond,
			wantErr:        assert.NoError,
			wantClusterId:  clusterId,
			wantName:       topicName,
			wantPartitions: 3,
			wantRetention:  -1,
		},
		{
			name:           "retention 7 days",
			retention:      7 * 24 * time.Hour,
			wantErr:        assert.NoError,
			wantClusterId:  clusterId,
			wantName:       topicName,
			wantPartitions: 3,
			wantRetention:  604800000,
		},
		{
			name:           "retention 31 days",
			retention:      31 * 24 * time.Hour,
			wantErr:        assert.NoError,
			wantClusterId:  clusterId,
			wantName:       topicName,
			wantPartitions: 3,
			wantRetention:  2678400000,
		},
		{
			name:           "retention 365 days",
			retention:      365 * 24 * time.Hour,
			wantErr:        assert.NoError,
			wantClusterId:  clusterId,
			wantName:       topicName,
			wantPartitions: 3,
			wantRetention:  31536000000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spy := &mocks.MockClient{}
			p := &Process{
				Context:   context.TODO(),
				Confluent: spy,
				State: models.NewProcessState(
					"some-capability-root-id",
					clusterId,
					models.NewTopic(topicName, 3, tt.retention),
					false,
					false,
				),
			}

			tt.wantErr(t, p.createTopic(), fmt.Sprintf("createTopic()"))

			assert.Equal(t, tt.wantClusterId, spy.GotClusterId)
			assert.Equal(t, tt.wantName, spy.GotName)
			assert.Equal(t, tt.wantPartitions, spy.GotPartitions)
			assert.Equal(t, tt.wantRetention, spy.GotRetention)

		})
	}
}
