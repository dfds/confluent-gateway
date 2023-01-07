package create

import (
	"context"
	"github.com/dfds/confluent-gateway/models"
)

type topicService struct {
	context   context.Context
	confluent Confluent
	repo      topicRepository
}

type topicRepository interface {
	DeleteTopic(capabilityRootId models.CapabilityRootId, clusterId models.ClusterId, topicName string) error
}

func NewTopicService(context context.Context, confluent Confluent, repo topicRepository) *topicService {
	return &topicService{context: context, confluent: confluent, repo: repo}
}

func (p *topicService) DeleteTopic(capabilityRootId models.CapabilityRootId, clusterId models.ClusterId, topicName string) error {
	err := p.confluent.DeleteTopic(p.context, clusterId, topicName)
	if err != nil {
		return err
	}

	return p.repo.DeleteTopic(capabilityRootId, clusterId, topicName)
}
