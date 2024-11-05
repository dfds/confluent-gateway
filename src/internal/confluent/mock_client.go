// confluent_mock.go
package confluent

import (
	"context"

	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/stretchr/testify/mock"
)

// MockClient is a mock implementation of the Confluent Client.
type MockClient struct {
	mock.Mock
}

func (m *MockClient) ListSchemas(ctx context.Context, subjectPrefix string, clusterId models.ClusterId) ([]models.Schema, error) {
	args := m.Called(ctx, subjectPrefix, clusterId)
	return args.Get(0).([]models.Schema), args.Error(1)
}
