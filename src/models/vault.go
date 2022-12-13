package models

import "context"

type Vault interface {
	StoreApiKey(ctx context.Context, capabilityRootId CapabilityRootId, clusterId ClusterId, apiKey ApiKey) error
}
