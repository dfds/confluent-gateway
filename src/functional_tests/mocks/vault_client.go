package mocks

import (
	"context"
	"fmt"
	"github.com/dfds/confluent-gateway/internal/vault"
)

type vaultMock struct {
	keys map[string]string
}

func NewVaultMock() vault.Vault {
	return &vaultMock{
		keys: map[string]string{},
	}
}

// only for testing purposes
func getTestApiParameter(input vault.Input) string {
	switch input.OperationDestination {
	case vault.OperationDestinationCluster:
		return fmt.Sprintf("/capabilities/%s/kafka/%s/credentials", input.CapabilityId, input.ClusterId)
	case vault.OperationDestinationSchemaRegistry:
		return fmt.Sprintf("/capabilities/%s/kafka/%s/schemaregistry-credentials", input.CapabilityId, input.ClusterId)
	}
	return ""
}

func getTestVaultInput(input vault.Input) string {
	return fmt.Sprintf("%s-%s-%s", input.CapabilityId, input.ClusterId, input.OperationDestination)
}

func (v *vaultMock) StoreApiKey(ctx context.Context, input vault.Input) error {
	v.keys[getTestApiParameter(input)] = getTestVaultInput(input)
	return nil
}

func (v *vaultMock) QueryApiKey(ctx context.Context, input vault.Input) (bool, error) {
	_, ok := v.keys[getTestApiParameter(input)]
	return ok, nil
}

func (v *vaultMock) DeleteApiKey(ctx context.Context, input vault.Input) error {
	delete(v.keys, getTestApiParameter(input))
	return nil
}
