package models

type ConfluentClient interface {
	CreateServiceAccount(name string, description string) ServiceAccountId
	CreateACLEntry(clusterId ClusterId, serviceAccountId ServiceAccountId, entry AclEntry)
	CreateApiKey(clusterId ClusterId, serviceAccountId ServiceAccountId) ApiKey
	CreateTopic(clusterId ClusterId, name string, partitions int, retention int)
}
