package delete

import (
	"context"
	"github.com/dfds/confluent-gateway/internal/models"
)

type Confluent interface {
	DeleteTopic(ctx context.Context, clusterId models.ClusterId, topicName string) error
	DeleteSchema(ctx context.Context, clusterId models.ClusterId, subject string, schema string, version string) error
}
