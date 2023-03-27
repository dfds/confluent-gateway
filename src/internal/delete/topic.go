package create

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
	DeleteTopic(topicId string) error
}

func NewTopicService(context context.Context, confluent Confluent, repo topicRepository) *topicService {
	return &topicService{context: context, confluent: confluent, repo: repo}
}

func (p *topicService) DeleteTopic(clusterId models.ClusterId, topicId string, topicName string) error {
	err := p.confluent.DeleteTopic(p.context, clusterId, topicName)
	if err != nil {
		return err
	}

	return p.repo.DeleteTopic(topicId)
}
