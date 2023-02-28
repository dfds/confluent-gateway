package mocks

import "github.com/dfds/confluent-gateway/internal/models"

type StepContextMock struct {
	ReturnClusterAccess *models.ClusterAccess

	MarkServiceAccountAsReadyWasCalled bool
	MarkClusterAccessAsReadyWasCalled  bool
	MarkApiKeyAsReadyWasCalled         bool
	MarkApiKeyInVaultAsReadyWasCalled  bool
	MarkAsCompletedWasCalled           bool
	TopicProvisionedEventWasRaised     bool
	TopicDeletedEventWasRaised         bool

	OnCreateServiceAccountError     error
	OnGetOrCreateClusterAccessError error
	OnCreateAclEntryError           error
	OnGetClusterAccessError         error
	OnCreateApiKeyError             error
	OnStoreApiKeyError              error
	OnCreateTopicError              error
	OnDeleteTopicError              error
}

func (m *StepContextMock) HasServiceAccount() bool {
	return false
}

func (m *StepContextMock) CreateServiceAccount() error {
	return m.OnCreateServiceAccountError
}

func (m *StepContextMock) MarkServiceAccountAsReady() {
	m.MarkServiceAccountAsReadyWasCalled = true
}

func (m *StepContextMock) HasClusterAccess() bool {
	return false
}

func (m *StepContextMock) GetOrCreateClusterAccess() (*models.ClusterAccess, error) {
	return m.ReturnClusterAccess, m.OnGetOrCreateClusterAccessError
}

func (m *StepContextMock) CreateAclEntry(*models.ClusterAccess, models.AclEntry) error {
	return m.OnCreateAclEntryError
}

func (m *StepContextMock) MarkClusterAccessAsReady() {
	m.MarkClusterAccessAsReadyWasCalled = true
}

func (m *StepContextMock) HasApiKey() bool {
	return false
}

func (m *StepContextMock) GetClusterAccess() (*models.ClusterAccess, error) {
	return m.ReturnClusterAccess, m.OnGetClusterAccessError
}

func (m *StepContextMock) CreateApiKey(*models.ClusterAccess) error {
	return m.OnCreateApiKeyError
}

func (m *StepContextMock) MarkApiKeyAsReady() {
	m.MarkApiKeyAsReadyWasCalled = true
}

func (m *StepContextMock) HasApiKeyInVault() bool {
	return false
}

func (m *StepContextMock) StoreApiKey(*models.ClusterAccess) error {
	return m.OnStoreApiKeyError
}

func (m *StepContextMock) MarkApiKeyInVaultAsReady() {
	m.MarkApiKeyInVaultAsReadyWasCalled = true
}

func (m *StepContextMock) IsCompleted() bool {
	return false
}

func (m *StepContextMock) CreateTopic() error {
	return m.OnCreateTopicError
}

func (m *StepContextMock) MarkAsCompleted() {
	m.MarkAsCompletedWasCalled = true
}

func (m *StepContextMock) RaiseTopicProvisionedEvent() error {
	m.TopicProvisionedEventWasRaised = true
	return nil
}

func (m *StepContextMock) DeleteTopic() error {
	return m.OnDeleteTopicError
}

func (m *StepContextMock) RaiseTopicDeletedEvent() error {
	m.TopicDeletedEventWasRaised = true
	return nil
}
