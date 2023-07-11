package mocks

import (
	"context"
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestVaultMock_PutAndGet(t *testing.T) {
	ctx := context.Background()
	capabilityId := models.CapabilityId("my-test-capability-id")
	clusterId := models.ClusterId("my-test-cluster-id")
	key := models.ApiKey{
		Username: "apikey-test-user",
		Password: "apikey-test-password",
	}

	mock := NewVaultMock()

	// Cluster api key
	err := mock.StoreClusterApiKey(ctx, capabilityId, clusterId, key)
	require.NoError(t, err)
	hasApiKey, err := mock.QueryClusterApiKey(ctx, capabilityId, clusterId)
	require.NoError(t, err)
	require.True(t, hasApiKey)
	err = mock.DeleteClusterApiKey(ctx, capabilityId, clusterId)
	require.NoError(t, err)
	hasApiKey, err = mock.QueryClusterApiKey(ctx, capabilityId, clusterId)
	require.NoError(t, err)
	require.False(t, hasApiKey)

	// Schema registry api key
	err = mock.StoreSchemaRegistryApiKey(ctx, capabilityId, clusterId, key)
	require.NoError(t, err)
	hasApiKey, err = mock.QuerySchemaRegistryApiKey(ctx, capabilityId, clusterId)
	require.NoError(t, err)
	require.True(t, hasApiKey)
	err = mock.DeleteSchemaRegistryApiKey(ctx, capabilityId, clusterId)
	require.NoError(t, err)
	hasApiKey, err = mock.QuerySchemaRegistryApiKey(ctx, capabilityId, clusterId)
	require.NoError(t, err)
}
