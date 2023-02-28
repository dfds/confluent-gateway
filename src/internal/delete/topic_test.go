package create

import (
	"context"
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTopicService_DeleteTopic(t *testing.T) {
	confluentSpy := &mocks.MockClient{}
	repoSpy := &topicRepositoryMock{}
	sut := NewTopicService(context.TODO(), confluentSpy, repoSpy)
	err := sut.DeleteTopic(someCapabilityRootId, someClusterId, someTopicName)

	assert.NoError(t, err)
	assert.Equal(t, string(someClusterId), confluentSpy.GotClusterId)
	assert.Equal(t, someTopicName, confluentSpy.GotName)
	assert.Equal(t, someCapabilityRootId, repoSpy.GotCapabilityRootId)
	assert.Equal(t, someClusterId, repoSpy.GotClusterId)
	assert.Equal(t, someTopicName, repoSpy.GotTopicName)
}

func TestTopicService_DeleteTopic_ConfluentError(t *testing.T) {
	spy := &mocks.MockClient{OnDeleteTopicError: serviceError}
	sut := NewTopicService(context.TODO(), spy, &topicRepositoryMock{})
	err := sut.DeleteTopic(someCapabilityRootId, someClusterId, someTopicName)

	assert.Error(t, err)
}

func TestTopicService_DeleteTopic_DatabaseError(t *testing.T) {
	sut := NewTopicService(context.TODO(), &mocks.MockClient{}, &topicRepositoryMock{OnDeleteTopicError: serviceError})
	err := sut.DeleteTopic(someCapabilityRootId, someClusterId, someTopicName)

	assert.Error(t, err)
}

type topicRepositoryMock struct {
	GotCapabilityRootId models.CapabilityRootId
	GotClusterId        models.ClusterId
	GotTopicName        string
	OnDeleteTopicError  error
}

func (m *topicRepositoryMock) DeleteTopic(capabilityRootId models.CapabilityRootId, clusterId models.ClusterId, topicName string) error {
	m.GotCapabilityRootId = capabilityRootId
	m.GotClusterId = clusterId
	m.GotTopicName = topicName
	return m.OnDeleteTopicError
}
