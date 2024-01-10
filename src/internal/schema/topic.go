package schema

import (
	"github.com/dfds/confluent-gateway/internal/models"
)

type topicService struct {
	repo topicRepository
}

type topicRepository interface {
	GetTopic(topicId string) (*models.Topic, error)
}

func NewTopicService(repo topicRepository) *topicService {
	return &topicService{repo: repo}
}

func (p *topicService) GetTopic(topicId string) (*models.Topic, error) {
	return p.repo.GetTopic(topicId)
}
