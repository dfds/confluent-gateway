package handlers

import (
	"context"
	"testing"

	"github.com/dfds/confluent-gateway/internal/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewHandler(t *testing.T) {
	// Arrange: Create mock objects for the dependencies
	ctx := context.Background()
	mockLogger := &mocks.MockLogger{}
	mockSchemaService := &mocks.MockSchemaService{}

	// Act: Initialize a new Handler
	handler := NewHandler(ctx, mockLogger, mockSchemaService)

	// Assert: Check if the handler is correctly initialized
	assert.NotNil(t, handler)
	assert.Equal(t, ctx, handler.Ctx)
	assert.Equal(t, mockLogger, handler.Logger)
	assert.Equal(t, mockSchemaService, handler.SchemaService)
}
