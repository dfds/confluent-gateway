package models

import "context"

type Confluent interface {
	CreateServiceAccount(ctx context.Context, name string, description string) (*ServiceAccountId, error)
	CreateACLEntry(ctx context.Context, clusterId ClusterId, serviceAccountId ServiceAccountId, entry AclDefinition) error
	CreateApiKey(ctx context.Context, clusterId ClusterId, serviceAccountId ServiceAccountId) (*ApiKey, error)
	CreateTopic(ctx context.Context, clusterId ClusterId, name string, partitions int, retention int) error
}
