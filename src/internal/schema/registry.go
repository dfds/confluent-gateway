package schema

import (
	"context"
	"github.com/dfds/confluent-gateway/internal/models"
)

type SchemaRegistry interface {
	RegisterSchema(ctx context.Context, clusterId models.ClusterId, subject string, schema string, version int32) error
}
