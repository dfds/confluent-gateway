package mocks

import (
	"context"
	"fmt"
	"github.com/dfds/confluent-gateway/models"
)

type MockClient struct {
	ServiceAccountId models.ServiceAccountId
	ApiKey           models.ApiKey
	Error            error
	ClusterId        string
	Name             string
	Partitions       int
	Retention        int64
}

func (m *MockClient) CreateServiceAccount(ctx context.Context, name string, description string) (*models.ServiceAccountId, error) {
	fmt.Printf("Creating Service Account %s (%s)\n", name, description)
	return &m.ServiceAccountId, m.Error
}

func (m *MockClient) CreateACLEntry(ctx context.Context, clusterId models.ClusterId, serviceAccountId models.ServiceAccountId, entry models.AclDefinition) error {
	fmt.Printf("Creating ACL entry for %s on %s [%s]\n", clusterId, serviceAccountId, entry)
	return m.Error
}

func (m *MockClient) CreateApiKey(ctx context.Context, clusterId models.ClusterId, serviceAccountId models.ServiceAccountId) (*models.ApiKey, error) {
	fmt.Printf("Creating API key for %s on %s\n", clusterId, serviceAccountId)
	return &m.ApiKey, m.Error
}

func (m *MockClient) CreateTopic(ctx context.Context, clusterId models.ClusterId, name string, partitions int, retention int64) error {
	fmt.Printf("Creating topic %s on %s (Partitions=%d, Retention=%d\n", name, clusterId, partitions, retention)
	m.ClusterId = string(clusterId)
	m.Name = name
	m.Partitions = partitions
	m.Retention = retention
	return m.Error
}
