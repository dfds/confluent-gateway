package vault

import (
	"context"
	"github.com/dfds/confluent-gateway/internal/models"
)

type Vault interface {
	StoreClusterApiKey(ctx context.Context, capabilityId models.CapabilityId, clusterId models.ClusterId, apiKey models.ApiKey) error
	StoreSchemaRegistryApiKey(ctx context.Context, capabilityId models.CapabilityId, clusterId models.ClusterId, apiKey models.ApiKey) error
	QuerySchemaRegistryApiKey(ctx context.Context, capabilityId models.CapabilityId, clusterId models.ClusterId) (bool, error)
	QueryClusterApiKey(ctx context.Context, capabilityId models.CapabilityId, clusterId models.ClusterId) (bool, error)
	DeleteClusterApiKey(ctx context.Context, capabilityId models.CapabilityId, clusterId models.ClusterId) error
	DeleteSchemaRegistryApiKey(ctx context.Context, capabilityId models.CapabilityId, clusterId models.ClusterId) error
}
