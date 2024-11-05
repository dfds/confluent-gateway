package services

import (
	"context"
	"testing"

	"github.com/dfds/confluent-gateway/internal/mocks"
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListSchemas(t *testing.T) {
	mockClient := new(mocks.MockClient)
	mockLogger := new(mocks.MockLogger)

	// Initialize the service with the mocked client and logger
	schemaService := &SchemaService{
		Logger:          mockLogger,
		ConfluentClient: mockClient,
	}

	// Define expected schemas
	expectedSchemas := []models.Schema{
		{
			ID:         1,
			Subject:    "subject-1",
			Version:    1,
			SchemaType: "AVRO",
			Schema:     `{"type": "record", "name": "TestRecord", "fields": [{"name": "field1", "type": "string"}]}`,
		},
	}

	// Test 1: Successful listing of schemas
	mockClient.On("ListSchemas", mock.Anything, "", "").Return(expectedSchemas, nil)

	schemas, err := schemaService.ListSchemas(context.TODO(), "", "")
	assert.Equal(t, expectedSchemas, schemas)
	assert.NoError(t, err)

	// Assert that all expectations were met
	mockClient.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}
