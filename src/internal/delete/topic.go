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
	DeleteTopic(capabilityId models.CapabilityId, clusterId models.ClusterId, topicName string) error
}

func NewTopicService(context context.Context, confluent Confluent, repo topicRepository) *topicService {
	return &topicService{context: context, confluent: confluent, repo: repo}
}

func (p *topicService) DeleteTopic(capabilityId models.CapabilityId, clusterId models.ClusterId, topicName string) error {
	err := p.confluent.DeleteTopic(p.context, clusterId, topicName)
	if err != nil {
		return err
	}

	return p.repo.DeleteTopic(capabilityId, clusterId, topicName)
}
