package storage

import (
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewClusterCache(t *testing.T) {

	type testCase struct {
		name     string
		clusters []*models.Cluster
		key      models.ClusterId
		wantErr  bool
	}
	tests := []testCase{
		{
			name:     "empty",
			clusters: nil,
			key:      "cluster-1",
			wantErr:  true,
		},
		{
			name:     "ok",
			clusters: []*models.Cluster{{ClusterId: "cluster-1"}},
			key:      "cluster-1",
			wantErr:  false,
		},
		{
			name:     "not found",
			clusters: []*models.Cluster{{ClusterId: "cluster-1"}},
			key:      "cluster-2",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := NewClusterCache(tt.clusters)

			_, err := sut.Get(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestErrKeyNotFound_Error(t *testing.T) {
	err := ErrKeyNotFound("what")

	assert.Error(t, err)
	assert.Equal(t, err.Error(), "key \"what\" not found")
}
