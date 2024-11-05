package mocks

import (
	"context"

	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/stretchr/testify/mock"
)

type MockSchemaService struct {
	mock.Mock
}

func (m *MockSchemaService) ListSchemas(ctx context.Context, subjectPrefix string, clusterId models.ClusterId) ([]models.Schema, error) {
	args := m.Called(ctx, subjectPrefix, clusterId)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]models.Schema), args.Error(1)
}
