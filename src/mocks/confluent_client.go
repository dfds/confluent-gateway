package mocks

import (
	"context"
	"fmt"
	"github.com/dfds/confluent-gateway/models"
)

type MockClient struct {
	ReturnServiceAccountId      models.ServiceAccountId
	ReturnApiKey                models.ApiKey
	GotClusterId                string
	GotName                     string
	GotPartitions               int
	GotRetention                int64
	OnCreateServiceAccountError error
	OnCreateAclEntryError       error
	OnCreateApiKeyError         error
	OnCreateTopicError          error
}

func (m *MockClient) CreateServiceAccount(_ context.Context, name string, description string) (*models.ServiceAccountId, error) {
	fmt.Printf("Creating Service Account %s (%s)\n", name, description)
	return &m.ReturnServiceAccountId, m.OnCreateServiceAccountError
}

func (m *MockClient) CreateACLEntry(_ context.Context, clusterId models.ClusterId, serviceAccountId models.ServiceAccountId, entry models.AclDefinition) error {
	fmt.Printf("Creating ACL entry for %s on %s [%s]\n", clusterId, serviceAccountId, entry)
	return m.OnCreateAclEntryError
}

func (m *MockClient) CreateApiKey(_ context.Context, clusterId models.ClusterId, serviceAccountId models.ServiceAccountId) (*models.ApiKey, error) {
	fmt.Printf("Creating API key for %s on %s\n", clusterId, serviceAccountId)
	return &m.ReturnApiKey, m.OnCreateApiKeyError
}

func (m *MockClient) CreateTopic(_ context.Context, clusterId models.ClusterId, name string, partitions int, retention int64) error {
	fmt.Printf("Creating topic %s on %s (Partitions=%d, Retention=%d\n", name, clusterId, partitions, retention)
	m.GotClusterId = string(clusterId)
	m.GotName = name
	m.GotPartitions = partitions
	m.GotRetention = retention
	return m.OnCreateTopicError
}
