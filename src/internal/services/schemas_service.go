package services

import (
	"context"

	"github.com/dfds/confluent-gateway/internal/confluent"
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/logging"
)

type SchemaServiceInterface interface {
	ListSchemas(ctx context.Context, subjectPrefix string) ([]models.Schema, error)
}

type SchemaService struct {
	Logger          logging.Logger
	ConfluentClient confluent.ConfluentClient
}

func NewSchemaService(logger logging.Logger, confluentClient confluent.ConfluentClient) *SchemaService {
	return &SchemaService{
		ConfluentClient: confluentClient,
	}
}

func (s *SchemaService) ListSchemas(ctx context.Context, subjectPrefix string) ([]models.Schema, error) {
	schemas, err := s.ConfluentClient.ListSchemas(ctx, subjectPrefix)

	if err != nil {
		s.Logger.Error(err, "failed to list schemas")
		return nil, err
	}

	return schemas, nil
}
