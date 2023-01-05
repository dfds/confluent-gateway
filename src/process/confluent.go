package process

import (
	"context"
	"github.com/dfds/confluent-gateway/models"
)

type Confluent interface {
	CreateServiceAccount(ctx context.Context, name string, description string) (models.ServiceAccountId, error)
	CreateACLEntry(ctx context.Context, clusterId models.ClusterId, userAccountId models.UserAccountId, entry models.AclDefinition) error
	CreateApiKey(ctx context.Context, clusterId models.ClusterId, serviceAccountId models.ServiceAccountId) (*models.ApiKey, error)
	CreateTopic(ctx context.Context, clusterId models.ClusterId, name string, partitions int, retention int64) error
	GetUsers(ctx context.Context) ([]models.User, error)
}
