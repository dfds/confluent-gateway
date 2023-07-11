package mocks

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/internal/vault"
)

type vaultMock struct {
	keys map[string]*ssm.PutParameterInput
}

func NewVaultMock() vault.Vault {
	return &vaultMock{
		keys: map[string]*ssm.PutParameterInput{},
	}
}

func (v *vaultMock) StoreClusterApiKey(ctx context.Context, capabilityId models.CapabilityId, clusterId models.ClusterId, apiKey models.ApiKey) error {
	parameter := vault.GetClusterApiParameter(capabilityId, clusterId)
	v.keys[parameter] = vault.CreateApiKeyInput(string(capabilityId), parameter, apiKey)
	return nil
}

func (v *vaultMock) QueryClusterApiKey(ctx context.Context, capabilityId models.CapabilityId, clusterId models.ClusterId) (bool, error) {
	parameter := vault.GetClusterApiParameter(capabilityId, clusterId)
	_, ok := v.keys[parameter]
	return ok, nil
}

func (v *vaultMock) StoreSchemaRegistryApiKey(ctx context.Context, capabilityId models.CapabilityId, clusterId models.ClusterId, apiKey models.ApiKey) error {
	parameter := vault.GetSchemaRegistryApiParameter(capabilityId, clusterId)
	v.keys[parameter] = vault.CreateApiKeyInput(string(capabilityId), parameter, apiKey)
	return nil
}

func (v *vaultMock) QuerySchemaRegistryApiKey(ctx context.Context, capabilityId models.CapabilityId, clusterId models.ClusterId) (bool, error) {
	parameter := vault.GetSchemaRegistryApiParameter(capabilityId, clusterId)
	_, ok := v.keys[parameter]
	return ok, nil
}

func (v *vaultMock) DeleteClusterApiKey(ctx context.Context, capabilityId models.CapabilityId, clusterId models.ClusterId) error {
	delete(v.keys, vault.GetClusterApiParameter(capabilityId, clusterId))
	return nil
}

func (v *vaultMock) DeleteSchemaRegistryApiKey(ctx context.Context, capabilityId models.CapabilityId, clusterId models.ClusterId) error {
	delete(v.keys, vault.GetSchemaRegistryApiParameter(capabilityId, clusterId))
	return nil
}
