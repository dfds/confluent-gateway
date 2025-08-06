package mocks

import (
	"context"

	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/stretchr/testify/mock"
)

type MockTopicService struct {
	mock.Mock
}

func (m *MockTopicService) ListTopics(ctx context.Context, clusterId models.ClusterId) ([]models.ConfluentTopic, error) {
	args := m.Called(ctx, clusterId)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]models.ConfluentTopic), args.Error(1)
}
