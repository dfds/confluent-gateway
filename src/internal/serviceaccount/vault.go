package serviceaccount

import (
	"context"

	"github.com/dfds/confluent-gateway/internal/models"
)

type Vault interface {
	StoreClusterApiKey(ctx context.Context, capabilityId models.CapabilityId, clusterId models.ClusterId, apiKey models.ApiKey) error
	StoreSchemaRegistryApiKey(ctx context.Context, capabilityId models.CapabilityId, clusterId models.ClusterId, apiKey models.ApiKey) error
	QueryClusterApiKey(ctx context.Context, capabilityId models.CapabilityId, clusterId models.ClusterId) (bool, error)
	QuerySchemaRegistryApiKey(ctx context.Context, capabilityId models.CapabilityId, clusterId models.ClusterId) (bool, error)
}

type vaultService struct {
	context context.Context
	vault   Vault
}

func NewVaultService(context context.Context, vault Vault) *vaultService {
	return &vaultService{context: context, vault: vault}
}

func (v *vaultService) StoreClusterApiKey(capabilityId models.CapabilityId, clusterAccess *models.ClusterAccess) error {
	return v.vault.StoreClusterApiKey(v.context, capabilityId, clusterAccess.ClusterId, clusterAccess.ApiKey)
}

func (v *vaultService) QueryClusterApiKey(capabilityId models.CapabilityId, clusterAccess *models.ClusterAccess) (bool, error) {
	return v.vault.QueryClusterApiKey(v.context, capabilityId, clusterAccess.ClusterId)
}

func (v *vaultService) StoreSchemaRegistryApiKey(capabilityId models.CapabilityId, clusterAccess *models.ClusterAccess) error {
	return v.vault.StoreSchemaRegistryApiKey(v.context, capabilityId, clusterAccess.ClusterId, clusterAccess.ApiKey)
}

func (v *vaultService) QuerySchemaRegistryApiKey(capabilityId models.CapabilityId, clusterAccess *models.ClusterAccess) (bool, error) {
	return v.vault.QuerySchemaRegistryApiKey(v.context, capabilityId, clusterAccess.ClusterId)
}
