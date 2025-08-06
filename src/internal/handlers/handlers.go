package handlers

import (
	"context"

	"github.com/dfds/confluent-gateway/internal/services"
	"github.com/dfds/confluent-gateway/logging"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

type Handler struct {
	Ctx           context.Context
	Logger        logging.Logger
	SchemaService services.SchemaServiceInterface
	TopicService  services.TopicServiceInterface
}

func NewHandler(ctx context.Context, logger logging.Logger, schemaService services.SchemaServiceInterface, topicService services.TopicServiceInterface) *Handler {
	return &Handler{
		Ctx:           ctx,
		Logger:        logger,
		SchemaService: schemaService,
		TopicService:  topicService,
	}
}
