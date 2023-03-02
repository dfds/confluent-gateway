package create

import (
	"context"
	"github.com/dfds/confluent-gateway/internal/models"
)

type Vault interface {
	StoreApiKey(ctx context.Context, capabilityId models.CapabilityId, clusterId models.ClusterId, apiKey models.ApiKey) error
}

type vaultService struct {
	context context.Context
	vault   Vault
}

func NewVaultService(context context.Context, vault Vault) *vaultService {
	return &vaultService{context: context, vault: vault}
}

func (v *vaultService) StoreApiKey(capabilityId models.CapabilityId, clusterAccess *models.ClusterAccess) error {
	return v.vault.StoreApiKey(v.context, capabilityId, clusterAccess.ClusterId, clusterAccess.ApiKey)
}
