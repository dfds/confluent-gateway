package delete

import (
	"context"
	"github.com/dfds/confluent-gateway/internal/models"
)

type topicService struct {
	context   context.Context
	confluent Confluent
	repo      topicRepository
}

type topicRepository interface {
	GetTopic(topicId string) (*models.Topic, error)
	DeleteTopic(topicId string) error
}

func NewTopicService(context context.Context, confluent Confluent, repo topicRepository) *topicService {
	return &topicService{context: context, confluent: confluent, repo: repo}
}

func (p *topicService) DeleteTopic(topicId string) error {
	topic, err := p.repo.GetTopic(topicId)
	if err != nil {
		return err
	}

	err = p.confluent.DeleteTopic(p.context, topic.ClusterId, topic.Name)
	if err != nil {
		return err
	}

	return p.repo.DeleteTopic(topicId)
}
