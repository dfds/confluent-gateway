package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dfds/confluent-gateway/internal/mocks"
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListTopics_Success(t *testing.T) {
	// Arrange
	mockSchemaService := new(mocks.MockSchemaService)
	mockTopicService := new(mocks.MockTopicService)
	mockLogger := new(mocks.MockLogger)
	handler := NewHandler(context.Background(), mockLogger, mockSchemaService, mockTopicService)

	topics := []models.ConfluentTopic{
		{
			Kind: "Topic",
			Metadata: struct {
				Self         string `json:"self"`
				ResourceName string `json:"resource_name"`
			}{
				Self:         "/clusters/abc-1234/topics/topic1",
				ResourceName: "topic1",
			},
			ClusterId:         models.ClusterId("abc-1234"), // assuming ClusterId is a string alias
			TopicName:         "topic1",
			IsInternal:        false,
			ReplicationFactor: 3,
			PartitionsCount:   5,
			Partitions: []struct {
				Related string `json:"related"`
			}{
				{Related: "/clusters/abc-1234/topics/topic1/partitions/0"},
			},
			Configs: []struct {
				Related string `json:"related"`
			}{
				{Related: "/clusters/abc-1234/topics/topic1/configs/0"},
			},
			PartitionReasignments: []struct {
				Related string `json:"related"`
			}{
				{Related: "/clusters/abc-1234/topics/topic1/partition_reasignments"},
			},
		},
	}

	mockTopicService.On("ListTopics", mock.Anything, mock.Anything).Return(topics, nil)

	req, err := http.NewRequest(http.MethodGet, "/clusters/abc-1234/topics", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	// Act
	ListTopics(handler, rr, req, "")

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	expectedBody, _ := json.Marshal(topics)
	assert.JSONEq(t, string(expectedBody), rr.Body.String())
	mockSchemaService.AssertExpectations(t)
}

func TestListTopics_Error(t *testing.T) {
	// Arrange
	mockSchemaService := new(mocks.MockSchemaService)
	mockTopicService := new(mocks.MockTopicService)
	mockLogger := new(mocks.MockLogger)
	handler := NewHandler(context.Background(), mockLogger, mockSchemaService, mockTopicService)

	mockLogger.On("Error", mock.Anything, "failed to list topics", mock.Anything).Return(nil)

	mockTopicService.On("ListTopics", mock.Anything, mock.Anything).Return(nil, errors.New("failed to list schemas"))

	req, err := http.NewRequest(http.MethodGet, "/clusters/abc-1234/topics", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	// Act
	ListTopics(handler, rr, req, "")

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockSchemaService.AssertExpectations(t)
}
