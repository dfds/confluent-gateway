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
	mockService := new(mocks.MockSchemaService)
	mockLogger := new(mocks.MockLogger)
	handler := NewHandler(context.Background(), mockLogger, mockService)

	schemas := []models.Schema{
		{ID: 1, Subject: "Schema1"},
		{ID: 2, Subject: "Schema2"},
	}

	mockService.On("ListSchemas", mock.Anything, "", "").Return(schemas, nil)

	req, err := http.NewRequest(http.MethodGet, "/schemas", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	// Act
	ListSchemas(handler, rr, req, "", "")

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	expectedBody, _ := json.Marshal(schemas)
	assert.JSONEq(t, string(expectedBody), rr.Body.String())
	mockService.AssertExpectations(t)
}

func TestListSchemas_Error(t *testing.T) {
	// Arrange
	mockService := new(mocks.MockSchemaService)
	mockLogger := new(mocks.MockLogger)
	handler := NewHandler(context.Background(), mockLogger, mockService)

	mockLogger.On("Error", mock.Anything, "failed to list schemas", mock.Anything).Return(nil)

	mockService.On("ListSchemas", mock.Anything, "", "").Return(nil, errors.New("failed to list schemas"))

	req, err := http.NewRequest(http.MethodGet, "/schemas", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	// Act
	ListSchemas(handler, rr, req, "", "")

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}
