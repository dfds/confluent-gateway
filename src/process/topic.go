package process

import (
	"context"
	"github.com/dfds/confluent-gateway/models"
)

type topicService struct {
	context   context.Context
	confluent Confluent
}

func NewTopicService(context context.Context, confluent Confluent) *topicService {
	return &topicService{context: context, confluent: confluent}
}

func (p *topicService) CreateTopic(clusterId models.ClusterId, topic models.TopicDescription) error {
	return p.confluent.CreateTopic(p.context, clusterId, topic.Name, topic.Partitions, topic.RetentionInMs())
}
