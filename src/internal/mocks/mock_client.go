package mocks

import (
	"context"

	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
}

func (m *MockClient) ListSchemas(ctx context.Context, subjectPrefix string, clusterId models.ClusterId) ([]models.Schema, error) {
	args := m.Called(ctx, subjectPrefix, clusterId)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]models.Schema), args.Error(1)
}

func (m *MockClient) CreateServiceAccount(ctx context.Context, name string, description string) (models.ServiceAccountId, error) {
	args := m.Called(ctx, name, description)
	return args.Get(0).(models.ServiceAccountId), args.Error(1)
}

func (m *MockClient) GetServiceAccount(ctx context.Context, displayName string) (models.ServiceAccountId, error) {
	args := m.Called(ctx, displayName)
	return args.Get(0).(models.ServiceAccountId), args.Error(1)
}

func (m *MockClient) CreateClusterApiKey(ctx context.Context, clusterId models.ClusterId, serviceAccountId models.ServiceAccountId) (models.ApiKey, error) {
	args := m.Called(ctx, clusterId, serviceAccountId)
	return args.Get(0).(models.ApiKey), args.Error(1)
}

func (m *MockClient) CreateSchemaRegistryApiKey(ctx context.Context, clusterId models.ClusterId, serviceAccountId models.ServiceAccountId) (models.ApiKey, error) {
	args := m.Called(ctx, clusterId, serviceAccountId)
	return args.Get(0).(models.ApiKey), args.Error(1)
}

func (m *MockClient) DeleteClusterApiKey(ctx context.Context, clusterId models.ClusterId, serviceAccountId models.ServiceAccountId) error {
	args := m.Called(ctx, clusterId, serviceAccountId)
	return args.Error(0)
}

func (m *MockClient) DeleteSchemaRegistryApiKey(ctx context.Context, clusterId models.ClusterId, serviceAccountId models.ServiceAccountId) error {
	args := m.Called(ctx, clusterId, serviceAccountId)
	return args.Error(0)
}

func (m *MockClient) CreateServiceAccountRoleBinding(ctx context.Context, serviceAccount models.ServiceAccountId, clusterId models.ClusterId) error {
	args := m.Called(ctx, serviceAccount, clusterId)
	return args.Error(0)
}

func (m *MockClient) CreateTopic(ctx context.Context, clusterId models.ClusterId, name string, partitions int, retention int64) error {
	args := m.Called(ctx, clusterId, name, partitions, retention)
	return args.Error(0)
}

func (m *MockClient) DeleteTopic(ctx context.Context, clusterId models.ClusterId, topicName string) error {
	args := m.Called(ctx, clusterId, topicName)
	return args.Error(0)
}

func (m *MockClient) GetConfluentInternalUsers(ctx context.Context) ([]models.ConfluentInternalUser, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.ConfluentInternalUser), args.Error(1)
}

func (m *MockClient) CountClusterApiKeys(ctx context.Context, serviceAccountId models.ServiceAccountId, clusterId models.ClusterId) (int, error) {
	args := m.Called(ctx, serviceAccountId, clusterId)
	return args.Int(0), args.Error(1)
}

func (m *MockClient) CountSchemaRegistryApiKeys(ctx context.Context, serviceAccountId models.ServiceAccountId, clusterId models.ClusterId) (int, error) {
	args := m.Called(ctx, serviceAccountId, clusterId)
	return args.Int(0), args.Error(1)
}

func (m *MockClient) RegisterSchema(ctx context.Context, clusterId models.ClusterId, subject string, schema string, version int32) error {
	args := m.Called(ctx, clusterId, subject, schema, version)
	return args.Error(0)
}

func (m *MockClient) DeleteSchema(ctx context.Context, clusterId models.ClusterId, subject string, schema string, version string) error {
	args := m.Called(ctx, clusterId, subject, schema, version)
	return args.Error(0)
}

func (m *MockClient) CreateACLEntry(ctx context.Context, clusterId models.ClusterId, userAccountId models.UserAccountId, entry models.AclDefinition) error {
	args := m.Called(ctx, clusterId, userAccountId, entry)
	return args.Error(0)
}
