package process

import (
	"context"
	"github.com/dfds/confluent-gateway/models"
)

type Vault interface {
	StoreApiKey(ctx context.Context, capabilityRootId models.CapabilityRootId, clusterId models.ClusterId, apiKey models.ApiKey) error
}
