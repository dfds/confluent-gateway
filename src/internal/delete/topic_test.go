package create

import (
	"context"
	"github.com/dfds/confluent-gateway/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTopicService_DeleteTopic(t *testing.T) {
	confluentSpy := &mocks.MockClient{}
	repoSpy := &topicRepositoryMock{}
	sut := NewTopicService(context.TODO(), confluentSpy, repoSpy)
	err := sut.DeleteTopic(someClusterId, someTopicId, someTopicName)

	assert.NoError(t, err)
	assert.Equal(t, string(someClusterId), confluentSpy.GotClusterId)
	assert.Equal(t, someTopicName, confluentSpy.GotName)
	assert.Equal(t, someTopicId, repoSpy.GotTopicId)
}

func TestTopicService_DeleteTopic_ConfluentError(t *testing.T) {
	spy := &mocks.MockClient{OnDeleteTopicError: serviceError}
	sut := NewTopicService(context.TODO(), spy, &topicRepositoryMock{})
	err := sut.DeleteTopic(someClusterId, someTopicId, someTopicName)

	assert.Error(t, err)
}

func TestTopicService_DeleteTopic_DatabaseError(t *testing.T) {
	sut := NewTopicService(context.TODO(), &mocks.MockClient{}, &topicRepositoryMock{OnDeleteTopicError: serviceError})
	err := sut.DeleteTopic(someClusterId, someTopicId, someTopicName)

	assert.Error(t, err)
}

type topicRepositoryMock struct {
	GotTopicId         string
	OnDeleteTopicError error
}

func (m *topicRepositoryMock) DeleteTopic(topicId string) error {
	m.GotTopicId = topicId
	return m.OnDeleteTopicError
}
