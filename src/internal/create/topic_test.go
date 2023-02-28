package create

import (
	"context"
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTopicService_CreateTopic(t *testing.T) {
	confluentSpy := &mocks.MockClient{}
	repoSpy := &topicRepositoryMock{}
	sut := NewTopicService(context.TODO(), confluentSpy, repoSpy)
	err := sut.CreateTopic(someCapabilityRootId, someClusterId, models.TopicDescription{
		Name:       someTopicName,
		Partitions: 1,
		Retention:  -1 * time.Millisecond,
	})

	assert.NoError(t, err)
	assert.Equal(t, string(someClusterId), confluentSpy.GotClusterId)
	assert.Equal(t, someTopicName, confluentSpy.GotName)
	assert.Equal(t, 1, confluentSpy.GotPartitions)
	assert.Equal(t, int64(-1), confluentSpy.GotRetention)
	assert.Equal(t, someTopicName, repoSpy.GotTopic.Name)
	assert.Equal(t, 1, repoSpy.GotTopic.Partitions)
	assert.Equal(t, int64(-1), repoSpy.GotTopic.Retention)
}

func TestTopicService_CreateTopic_ConfluentError(t *testing.T) {
	spy := &mocks.MockClient{OnCreateTopicError: serviceError}
	sut := NewTopicService(context.TODO(), spy, &topicRepositoryMock{})
	err := sut.CreateTopic(someCapabilityRootId, someClusterId, models.TopicDescription{})

	assert.Error(t, err)
}

func TestTopicService_CreateTopic_DatabaseError(t *testing.T) {
	sut := NewTopicService(context.TODO(), &mocks.MockClient{}, &topicRepositoryMock{OnCreateTopicError: serviceError})
	err := sut.CreateTopic(someCapabilityRootId, someClusterId, models.TopicDescription{})

	assert.Error(t, err)
}

type topicRepositoryMock struct {
	GotTopic           *models.Topic
	OnCreateTopicError error
}

func (m *topicRepositoryMock) CreateTopic(topic *models.Topic) error {
	m.GotTopic = topic
	return m.OnCreateTopicError
}
