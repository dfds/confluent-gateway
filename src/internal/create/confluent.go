package create

import (
	"context"

	"github.com/dfds/confluent-gateway/internal/models"
)

type Confluent interface {
	CreateTopic(ctx context.Context, clusterId models.ClusterId, name string, partitions int, retention int64) error
}
