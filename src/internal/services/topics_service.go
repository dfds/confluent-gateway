package services

import (
	"context"

	"github.com/dfds/confluent-gateway/internal/confluent"
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/logging"
)

type TopicServiceInterface interface {
	ListTopics(ctx context.Context, clusterId models.ClusterId) ([]models.ConfluentTopic, error)
}

type TopicService struct {
	Logger          logging.Logger
	ConfluentClient confluent.ConfluentClient
}

func NewTopicService(logger logging.Logger, confluentClient confluent.ConfluentClient) *TopicService {
	return &TopicService{
		Logger:          logger,
		ConfluentClient: confluentClient,
	}
}

func (s *TopicService) ListTopics(ctx context.Context, clusterId models.ClusterId) ([]models.ConfluentTopic, error) {
	topics, err := s.ConfluentClient.ListTopics(ctx, clusterId)

	if err != nil {
		s.Logger.Error(err, "failed to list topics")
		return nil, err
	}

	return topics, nil
}
