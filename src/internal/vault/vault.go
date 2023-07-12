package vault

import (
	"context"
)

type Vault interface {
	StoreApiKey(ctx context.Context, input Input) error
	QueryApiKey(ctx context.Context, input Input) (bool, error)
	DeleteApiKey(ctx context.Context, input Input) error
}
