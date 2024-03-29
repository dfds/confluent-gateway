package serviceaccount

import (
	"context"

	"github.com/dfds/confluent-gateway/internal/models"
)

type Confluent interface {
	CreateServiceAccount(ctx context.Context, name string, description string) (models.ServiceAccountId, error)
	GetServiceAccount(ctx context.Context, displayName string) (models.ServiceAccountId, error)
	CreateACLEntry(ctx context.Context, clusterId models.ClusterId, userAccountId models.UserAccountId, entry models.AclDefinition) error
	CreateClusterApiKey(ctx context.Context, clusterId models.ClusterId, serviceAccountId models.ServiceAccountId) (models.ApiKey, error)
	CreateSchemaRegistryApiKey(ctx context.Context, clusterId models.ClusterId, serviceAccountId models.ServiceAccountId) (models.ApiKey, error)
	CreateServiceAccountRoleBinding(ctx context.Context, serviceAccount models.ServiceAccountId, clusterId models.ClusterId) error
	GetConfluentInternalUsers(ctx context.Context) ([]models.ConfluentInternalUser, error)
	CountClusterApiKeys(ctx context.Context, clusterAccess models.ServiceAccountId, clusterId models.ClusterId) (int, error)
	CountSchemaRegistryApiKeys(ctx context.Context, clusterAccess models.ServiceAccountId, clusterId models.ClusterId) (int, error)
	DeleteClusterApiKey(ctx context.Context, clusterId models.ClusterId, serviceAccountId models.ServiceAccountId) error
	DeleteSchemaRegistryApiKey(ctx context.Context, clusterId models.ClusterId, serviceAccountId models.ServiceAccountId) error
}
