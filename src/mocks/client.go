package mocks

import (
	"context"
	"fmt"

	"github.com/dfds/confluent-gateway/internal/models"
)

type MockClient struct {
	ReturnServiceAccountId      models.ServiceAccountId
	ReturnApiKey                models.ApiKey
	ReturnUsers                 []models.ConfluentInternalUser
	GotClusterId                string
	GotName                     string
	GotPartitions               int
	GotRetention                int64
	OnCreateServiceAccountError error
	OnCreateAclEntryError       error
	OnCreateApiKeyError         error
	OnCreateTopicError          error
	OnGetUsersError             error
	OnDeleteTopicError          error
	OnDeleteSchemaError         error
}

func (m *MockClient) CreateServiceAccount(_ context.Context, name string, description string) (models.ServiceAccountId, error) {
	fmt.Printf("Creating Service Account %s (%s)\n", name, description)
	return m.ReturnServiceAccountId, m.OnCreateServiceAccountError
}

func (m *MockClient) CreateACLEntry(_ context.Context, clusterId models.ClusterId, userAccountId models.UserAccountId, entry models.AclDefinition) error {
	fmt.Printf("Creating ACL entry for %s on %s [%s]\n", clusterId, userAccountId, entry)
	return m.OnCreateAclEntryError
}

func (m *MockClient) CreateClusterApiKey(ctx context.Context, clusterId models.ClusterId, serviceAccountId models.ServiceAccountId) (*models.ApiKey, error) {
	fmt.Printf("Creating cluster API key for %s on %s\n", clusterId, serviceAccountId)
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

func (m *MockClient) GetUsers(ctx context.Context) ([]models.ConfluentInternalUser, error) {
	return m.ReturnUsers, m.OnGetUsersError
}

func (m *MockClient) DeleteTopic(ctx context.Context, clusterId models.ClusterId, topicName string) error {
	m.GotClusterId = string(clusterId)
	m.GotName = topicName
	return m.OnDeleteTopicError
}

func (m *MockClient) DeleteSchema(ctx context.Context, clusterId models.ClusterId, subject string, schema string, version string) error {
	fmt.Printf("Deleting Schema %s.%s/%s@%s\n", clusterId, subject, schema, version)
	return m.OnDeleteSchemaError
}
