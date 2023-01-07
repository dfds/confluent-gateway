package create

import (
	"context"
	"github.com/dfds/confluent-gateway/models"
)

type Confluent interface {
	DeleteTopic(ctx context.Context, clusterId models.ClusterId, topicName string) error
}
