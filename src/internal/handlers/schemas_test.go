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

func TestListSchemas_Success(t *testing.T) {
	// Arrange
	mockSchemaService := new(mocks.MockSchemaService)
	mockTopicService := new(mocks.MockTopicService)
	mockLogger := new(mocks.MockLogger)
	handler := NewHandler(context.Background(), mockLogger, mockSchemaService, mockTopicService)

	schemas := []models.Schema{
		{ID: 1, Subject: "Schema1"},
		{ID: 2, Subject: "Schema2"},
	}

	mockSchemaService.On("ListSchemas", mock.Anything, "", mock.Anything).Return(schemas, nil)

	req, err := http.NewRequest(http.MethodGet, "/clusters/abc-1234/schemas", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	// Act
	ListSchemas(handler, rr, req, "", "")

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	expectedBody, _ := json.Marshal(schemas)
	assert.JSONEq(t, string(expectedBody), rr.Body.String())
	mockSchemaService.AssertExpectations(t)
}

func TestListSchemas_Error(t *testing.T) {
	// Arrange
	mockSchemaService := new(mocks.MockSchemaService)
	mockTopicService := new(mocks.MockTopicService)
	mockLogger := new(mocks.MockLogger)
	handler := NewHandler(context.Background(), mockLogger, mockSchemaService, mockTopicService)

	mockLogger.On("Error", mock.Anything, "failed to list schemas", mock.Anything).Return(nil)

	mockSchemaService.On("ListSchemas", mock.Anything, "", mock.Anything).Return(nil, errors.New("failed to list schemas"))

	req, err := http.NewRequest(http.MethodGet, "/clusters/abc-1234/schemas", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	// Act
	ListSchemas(handler, rr, req, "", "")

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockSchemaService.AssertExpectations(t)
}
