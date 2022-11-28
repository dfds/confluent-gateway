package models

import "context"

type ConfluentClient interface {
	CreateServiceAccount(ctx context.Context, name string, description string) ServiceAccountId
	CreateACLEntry(ctx context.Context, clusterId ClusterId, serviceAccountId ServiceAccountId, entry AclDefinition)
	CreateApiKey(ctx context.Context, clusterId ClusterId, serviceAccountId ServiceAccountId) ApiKey
	CreateTopic(ctx context.Context, clusterId ClusterId, name string, partitions int, retention int)
}
