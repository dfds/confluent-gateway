package schema

import (
	"context"

	"github.com/dfds/confluent-gateway/internal/models"
)

type Confluent interface {
	CreateServiceAccount(ctx context.Context, name string, description string) (models.ServiceAccountId, error)
	CreateSchemaRegistryApiKey(ctx context.Context, clusterId models.ClusterId, serviceAccountId models.ServiceAccountId) (models.ApiKey, error)
	CreateServiceAccountRoleBinding(ctx context.Context, serviceAccount models.ServiceAccountId, clusterId models.ClusterId) error
	CountSchemaRegistryApiKeys(ctx context.Context, clusterAccess models.ServiceAccountId, clusterId models.ClusterId) (int, error)
	DeleteSchemaRegistryApiKey(ctx context.Context, clusterId models.ClusterId, serviceAccountId models.ServiceAccountId) error
	RegisterSchema(ctx context.Context, clusterId models.ClusterId, subject string, schema string) error
}
