package serviceaccount

import (
	"context"

	"github.com/dfds/confluent-gateway/internal/models"
)

type Confluent interface {
	CreateServiceAccount(ctx context.Context, name string, description string) (models.ServiceAccountId, error)
	CreateACLEntry(ctx context.Context, clusterId models.ClusterId, userAccountId models.UserAccountId, entry models.AclDefinition) error
	CreateApiKey(ctx context.Context, clusterId models.ClusterId, serviceAccountId models.ServiceAccountId) (*models.ApiKey, error)
	GetUsers(ctx context.Context) ([]models.User, error)
	CountApiKeys(ctx context.Context, clusterAccess models.ServiceAccountId, clusterId models.ClusterId) (int, error)
}