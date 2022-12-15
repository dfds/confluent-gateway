package mocks

import "github.com/dfds/confluent-gateway/models"

type AccountService struct {
	ReturnCapabilityRootId models.CapabilityRootId
	ReturnClusterId        models.ClusterId
	ReturnServiceAccountId models.ServiceAccountId
	ReturnServiceAccount   *models.ServiceAccount
	ReturnApiKey           models.ApiKey

	SavedServiceAccount *models.ServiceAccount
	SavedClusterAccess  *models.ClusterAccess

	OnCreateServiceAccountError error
	OnSaveServiceAccountError   error
	OnGetServiceAccountError    error
	OnCreateClusterAccessError  error
	OnUpdateClusterAccessError  error
	OnUpdateAclEntryError       error
	OnCreateAclEntryError       error
	OnCreateApiKeyError         error
}

func (m *AccountService) CapabilityRootId() models.CapabilityRootId {
	return m.ReturnCapabilityRootId
}

func (m *AccountService) ClusterId() models.ClusterId {
	return m.ReturnClusterId
}

func (m *AccountService) CreateServiceAccount() (*models.ServiceAccountId, error) {
	return &m.ReturnServiceAccountId, m.OnCreateServiceAccountError
}

func (m *AccountService) SaveServiceAccount(serviceAccount *models.ServiceAccount) error {
	m.SavedServiceAccount = serviceAccount
	return m.OnSaveServiceAccountError
}

func (m *AccountService) GetServiceAccount() (*models.ServiceAccount, error) {
	return m.ReturnServiceAccount, m.OnGetServiceAccountError
}

func (m *AccountService) CreateClusterAccess(clusterAccess *models.ClusterAccess) error {
	m.SavedClusterAccess = clusterAccess
	return m.OnCreateClusterAccessError
}

func (m *AccountService) UpdateClusterAccess(*models.ClusterAccess) error {
	return m.OnUpdateClusterAccessError
}

func (m *AccountService) CreateAclEntry(models.ServiceAccountId, *models.AclEntry) error {
	return m.OnCreateAclEntryError
}

func (m *AccountService) UpdateAclEntry(*models.AclEntry) error {
	return m.OnUpdateAclEntryError
}

func (m *AccountService) CreateApiKey(*models.ClusterAccess) (*models.ApiKey, error) {
	return &m.ReturnApiKey, m.OnCreateApiKeyError
}
