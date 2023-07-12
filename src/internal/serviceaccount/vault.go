package serviceaccount

import (
	"context"
	"github.com/dfds/confluent-gateway/internal/vault"

	"github.com/dfds/confluent-gateway/internal/models"
)

type VaultService interface {
	StoreClusterApiKey(capabilityId models.CapabilityId, clusterId models.ClusterId, apiKey models.ApiKey, shouldOverwrite bool) error
	QueryClusterApiKey(capabilityId models.CapabilityId, clusterId models.ClusterId) (bool, error)
	DeleteClusterApiKey(capabilityId models.CapabilityId, clusterId models.ClusterId) error
	StoreSchemaRegistryApiKey(capabilityId models.CapabilityId, clusterId models.ClusterId, apiKey models.ApiKey, shouldOverwrite bool) error
	QuerySchemaRegistryApiKey(capabilityId models.CapabilityId, clusterId models.ClusterId) (bool, error)
	DeleteSchemaRegistryApiKey(capabilityId models.CapabilityId, clusterId models.ClusterId) error
}

type vaultService struct {
	context context.Context
	vault   vault.Vault
}

func NewVaultService(context context.Context, vault vault.Vault) *vaultService {
	return &vaultService{context: context, vault: vault}
}

func (v *vaultService) StoreClusterApiKey(capabilityId models.CapabilityId, clusterId models.ClusterId, apiKey models.ApiKey, shouldOverwrite bool) error {
	return v.vault.StoreApiKey(v.context, vault.Input{
		OperationDestination: vault.OperationDestinationCluster,
		CapabilityId:         capabilityId,
		ClusterId:            clusterId,
		StoringInput:         &vault.StoringInput{ApiKey: apiKey, Overwrite: shouldOverwrite},
	})
}

func (v *vaultService) QueryClusterApiKey(capabilityId models.CapabilityId, clusterId models.ClusterId) (bool, error) {
	return v.vault.QueryApiKey(v.context, vault.Input{
		OperationDestination: vault.OperationDestinationCluster,
		CapabilityId:         capabilityId,
		ClusterId:            clusterId,
	})
}

func (v *vaultService) DeleteClusterApiKey(capabilityId models.CapabilityId, clusterId models.ClusterId) error {
	return v.vault.DeleteApiKey(v.context, vault.Input{
		OperationDestination: vault.OperationDestinationCluster,
		CapabilityId:         capabilityId,
		ClusterId:            clusterId,
	})
}

func (v *vaultService) StoreSchemaRegistryApiKey(capabilityId models.CapabilityId, clusterId models.ClusterId, apiKey models.ApiKey, shouldOverwrite bool) error {
	return v.vault.StoreApiKey(v.context, vault.Input{
		OperationDestination: vault.OperationDestinationSchemaRegistry,
		CapabilityId:         capabilityId,
		ClusterId:            clusterId,
		StoringInput:         &vault.StoringInput{ApiKey: apiKey, Overwrite: shouldOverwrite},
	})

}

func (v *vaultService) QuerySchemaRegistryApiKey(capabilityId models.CapabilityId, clusterId models.ClusterId) (bool, error) {
	return v.vault.QueryApiKey(v.context, vault.Input{
		OperationDestination: vault.OperationDestinationSchemaRegistry,
		CapabilityId:         capabilityId,
		ClusterId:            clusterId,
	})

}

func (v *vaultService) DeleteSchemaRegistryApiKey(capabilityId models.CapabilityId, clusterId models.ClusterId) error {
	return v.vault.DeleteApiKey(v.context, vault.Input{
		OperationDestination: vault.OperationDestinationSchemaRegistry,
		CapabilityId:         capabilityId,
		ClusterId:            clusterId,
	})
}
