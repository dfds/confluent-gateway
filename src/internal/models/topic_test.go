package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const someTopicName = "some-name"
const somePartitions = 1

var days = 24 * time.Hour

func TestNewTopicDescription_WithRetentionFromDuration(t *testing.T) {
	topic, err := NewTopicDescription(someTopicName, somePartitions, RetentionFromDuration(1*time.Millisecond))

	assert.NoError(t, err)
	assert.Equal(t, someTopicName, topic.Name)
	assert.Equal(t, somePartitions, topic.Partitions)
	assert.Equal(t, time.Duration(1000000), topic.Retention)
}

func TestNewTopicDescription_WithRetentionFromString(t *testing.T) {
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
			retention:     foreverString,
			wantErr:       assert.NoError,
			wantRetention: -1 * time.Millisecond,
		},
		{
			name:          "7 days",
			retention:     "604800000ms",
			wantErr:       assert.NoError,
			wantRetention: 7 * days,
		},
		{
			name:          "31 days",
			retention:     "2678400000ms",
			wantErr:       assert.NoError,
			wantRetention: 31 * days,
		},
		{
			name:          "365 days",
			retention:     "31536000000ms",
			wantErr:       assert.NoError,
			wantRetention: 365 * days,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			topic, err := NewTopicDescription(someTopicName, somePartitions, RetentionFromString(tt.retention))

			if !tt.wantErr(t, err) {
				return
			}

			assert.Equal(t, someTopicName, topic.Name)
			assert.Equal(t, somePartitions, topic.Partitions)
			assert.Equal(t, tt.wantRetention, topic.Retention)
		})
	}
}

func TestNewTopicDescription_WithRetentionFromMs(t *testing.T) {
	tests := []struct {
		name          string
		retention     int64
		wantRetention time.Duration
	}{
		{
			name:          "forever",
			retention:     -1,
			wantRetention: -1 * time.Millisecond,
		},
		{
			name:          "7 days",
			retention:     604800000,
			wantRetention: 7 * days,
		},
		{
			name:          "31 days",
			retention:     2678400000,
			wantRetention: 31 * days,
		},
		{
			name:          "365 days",
			retention:     31536000000,
			wantRetention: 365 * days,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			topic, err := NewTopicDescription(someTopicName, 1, RetentionFromMs(tt.retention))

			assert.NoError(t, err)
			assert.Equal(t, someTopicName, topic.Name)
			assert.Equal(t, 1, topic.Partitions)
			assert.Equal(t, tt.wantRetention, topic.Retention)
		})
	}
}

func TestTopicDescription_RetentionInMs(t *testing.T) {
	tests := []struct {
		name          string
		retention     time.Duration
		wantRetention int64
	}{
		{
			name:          "forever",
			retention:     -1 * time.Millisecond,
			wantRetention: -1,
		},
		{
			name:          "7 days",
			retention:     7 * days,
			wantRetention: 604800000,
		},
		{
			name:          "31 days",
			retention:     31 * days,
			wantRetention: 2678400000,
		},
		{
			name:          "365 days",
			retention:     365 * days,
			wantRetention: 31536000000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			topic, err := NewTopicDescription(someTopicName, 1, RetentionFromDuration(tt.retention))

			assert.NoError(t, err)
			assert.Equal(t, someTopicName, topic.Name)
			assert.Equal(t, 1, topic.Partitions)
			assert.Equal(t, tt.wantRetention, topic.RetentionInMs())
		})
	}
}
