package create

import (
	"context"
	"errors"
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTopicService_DeleteTopic(t *testing.T) {
	confluentSpy := &mocks.MockClient{}
	repoSpy := &topicRepositoryMock{Topic: &models.Topic{
		Id:           someTopicId,
		CapabilityId: someCapabilityId,
		ClusterId:    someClusterId,
		Name:         someTopicName,
	}}
	sut := NewTopicService(context.TODO(), confluentSpy, repoSpy)
	err := sut.DeleteTopic(someTopicId)

	assert.NoError(t, err)
	assert.Equal(t, string(someClusterId), confluentSpy.GotClusterId)
	assert.Equal(t, someTopicName, confluentSpy.GotName)
	assert.Equal(t, someTopicId, repoSpy.GotTopicId)
}

func TestTopicService_DeleteTopic_TopicNotFound(t *testing.T) {
	sut := NewTopicService(context.TODO(), &mocks.MockClient{}, &topicRepositoryMock{OnGetTopicError: errors.New("not found")})
	err := sut.DeleteTopic(someTopicId)

	assert.Equal(t, errors.New("not found"), err)
}

func TestTopicService_DeleteTopic_ConfluentError(t *testing.T) {
	spy := &mocks.MockClient{OnDeleteTopicError: serviceError}
	sut := NewTopicService(context.TODO(), spy, &topicRepositoryMock{Topic: &models.Topic{
		Id:           someTopicId,
		CapabilityId: someCapabilityId,
		ClusterId:    someClusterId,
		Name:         someTopicName,
	}})
	err := sut.DeleteTopic(someTopicId)

	assert.Equal(t, serviceError, err)
}

func TestTopicService_DeleteTopic_DatabaseError(t *testing.T) {
	sut := NewTopicService(context.TODO(), &mocks.MockClient{}, &topicRepositoryMock{
		Topic: &models.Topic{
			Id:           someTopicId,
			CapabilityId: someCapabilityId,
			ClusterId:    someClusterId,
			Name:         someTopicName,
		},
		OnDeleteTopicError: serviceError})
	err := sut.DeleteTopic(someTopicId)

	assert.Equal(t, serviceError, err)
}

type topicRepositoryMock struct {
	Topic              *models.Topic
	GotTopicId         string
	OnDeleteTopicError error
	OnGetTopicError    error
}

func (m *topicRepositoryMock) GetTopic(topicId string) (*models.Topic, error) {
	return m.Topic, m.OnGetTopicError
}

func (m *topicRepositoryMock) DeleteTopic(topicId string) error {
	m.GotTopicId = topicId
	return m.OnDeleteTopicError
}
