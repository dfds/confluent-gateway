package serviceaccount

import (
	"context"

	"github.com/dfds/confluent-gateway/internal/models"
)

type Vault interface {
	DeleteClusterApiKey(ctx context.Context, capabilityId models.CapabilityId, clusterId models.ClusterId) error
	DeleteSchemaRegistryApiKey(ctx context.Context, capabilityId models.CapabilityId, clusterId models.ClusterId) error
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

func (v *vaultService) StoreClusterApiKey(capabilityId models.CapabilityId, clusterAccess *models.ClusterAccess, apiKey models.ApiKey) error {
	return v.vault.StoreClusterApiKey(v.context, capabilityId, clusterAccess.ClusterId, apiKey)
}

func (v *vaultService) QueryClusterApiKey(capabilityId models.CapabilityId, clusterAccess *models.ClusterAccess) (bool, error) {
	return v.vault.QueryClusterApiKey(v.context, capabilityId, clusterAccess.ClusterId)
}

func (v *vaultService) StoreSchemaRegistryApiKey(capabilityId models.CapabilityId, clusterAccess *models.ClusterAccess, apiKey models.ApiKey) error {
	return v.vault.StoreSchemaRegistryApiKey(v.context, capabilityId, clusterAccess.ClusterId, apiKey)
}

func (v *vaultService) QuerySchemaRegistryApiKey(capabilityId models.CapabilityId, clusterAccess *models.ClusterAccess) (bool, error) {
	return v.vault.QuerySchemaRegistryApiKey(v.context, capabilityId, clusterAccess.ClusterId)
}

func (v *vaultService) DeleteClusterApiKey(capabilityId models.CapabilityId, clusterAccess *models.ClusterAccess) error {
	return v.vault.DeleteClusterApiKey(v.context, capabilityId, clusterAccess.ClusterId)
}

func (v *vaultService) DeleteSchemaRegistryApiKey(capabilityId models.CapabilityId, clusterAccess *models.ClusterAccess) error {
	return v.vault.DeleteSchemaRegistryApiKey(v.context, capabilityId, clusterAccess.ClusterId)
}
