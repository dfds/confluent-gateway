package create

import (
	"context"
	"github.com/dfds/confluent-gateway/internal/models"
)

type schemaService struct {
	context   context.Context
	confluent Confluent
	repo      schemaRepository
}

type schemaRepository interface {
	SelectSchemaProcessStatesByTopicId(topicId string) ([]models.SchemaProcess, error)
	DeleteSchemaProcessStateById(schemaId string) error
}

type SchemaService interface {
	DeleteSchemasByTopicId(string) error
}

func NewSchemaService(context context.Context, confluent Confluent, repo schemaRepository) SchemaService {
	return &schemaService{context: context, confluent: confluent, repo: repo}
}

func (p *schemaService) DeleteSchemasByTopicId(topicId string) error {
	schemaProcesses, err := p.repo.SelectSchemaProcessStatesByTopicId(topicId)
	if err != nil {
		return err
	}

	for _, schemaProcess := range schemaProcesses {
		err = p.confluent.DeleteSchema(p.context, schemaProcess.ClusterId, schemaProcess.Subject, schemaProcess.Schema, "latest")
		if err != nil {
			return err
		}
	}

	for _, schemaProcess := range schemaProcesses {
		err := p.repo.DeleteSchemaProcessStateById(schemaProcess.Id.String())
		if err != nil {
			return err
		}
	}

	return nil
}
