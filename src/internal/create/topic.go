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
	CreateTopic(topic *models.Topic) error
}

func NewTopicService(context context.Context, confluent Confluent, repo topicRepository) *topicService {
	return &topicService{context: context, confluent: confluent, repo: repo}
}

func (p *topicService) CreateTopic(capabilityId models.CapabilityId, clusterId models.ClusterId, topicId string, topic models.TopicDescription) error {
	err := p.confluent.CreateTopic(p.context, clusterId, topic.Name, topic.Partitions, topic.RetentionInMs())
	if err != nil {
		return err
	}

	return p.repo.CreateTopic(models.NewTopic(capabilityId, clusterId, topicId, topic))
}
