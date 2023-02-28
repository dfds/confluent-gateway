package create

import (
	"context"
	"github.com/dfds/confluent-gateway/internal/models"
)

type Vault interface {
	StoreApiKey(ctx context.Context, capabilityRootId models.CapabilityRootId, clusterId models.ClusterId, apiKey models.ApiKey) error
}

type vaultService struct {
	context context.Context
	vault   Vault
}

func NewVaultService(context context.Context, vault Vault) *vaultService {
	return &vaultService{context: context, vault: vault}
}

func (v *vaultService) StoreApiKey(capabilityRootId models.CapabilityRootId, clusterAccess *models.ClusterAccess) error {
	return v.vault.StoreApiKey(v.context, capabilityRootId, clusterAccess.ClusterId, clusterAccess.ApiKey)
}
