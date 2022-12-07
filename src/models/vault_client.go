package models

import "context"

type VaultClient interface {
	StoreApiKey(ctx context.Context, capabilityRootId CapabilityRootId, clusterId ClusterId, apiKey ApiKey) error
}
