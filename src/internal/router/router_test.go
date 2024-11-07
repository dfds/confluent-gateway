// internal/router/routers_test.go

package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dfds/confluent-gateway/internal/handlers"
	"github.com/dfds/confluent-gateway/internal/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSetupRoutes(t *testing.T) {
	// Create a mock logger
	mockLogger := &mocks.MockLogger{}

	// Create a mock SchemaService
	mockSchemaService := &mocks.MockSchemaService{}

	// Mock the ListSchemas method in the SchemaService
	mockSchemaService.On("ListSchemas", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(nil, nil)

	// Create a new handler with the mock services
	handler := handlers.NewHandler(context.Background(), mockLogger, mockSchemaService)

	// Initialize the routes with the handler
	router := SetupRoutes(handler)

	// Create a new HTTP request to the /schemas route
	req, err := http.NewRequest("GET", "/clusters/abc-1234/schemas", nil)
	require.NoError(t, err)

	// Record the response
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Assert that the status code is as expected (200 OK in this case)
	require.Equal(t, http.StatusOK, rr.Code)

	// Verify that the mock methods were called as expected
	mockLogger.AssertExpectations(t)
	mockSchemaService.AssertExpectations(t)
}
