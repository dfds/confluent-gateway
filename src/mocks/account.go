package mocks

import "github.com/dfds/confluent-gateway/models"

type AccountRepository struct {
	ReturnServiceAccount        *models.ServiceAccount
	GotServiceAccount           *models.ServiceAccount
	GotClusterAccess            *models.ClusterAccess
	OnCreateServiceAccountError error
	OnGetServiceAccountError    error
	OnCreateClusterAccessError  error
	OnUpdateClusterAccessError  error
	OnUpdateAclEntryError       error
}

func (m *AccountRepository) GetServiceAccount(models.CapabilityRootId) (*models.ServiceAccount, error) {
	return m.ReturnServiceAccount, m.OnGetServiceAccountError
}
func (m *AccountRepository) CreateServiceAccount(serviceAccount *models.ServiceAccount) error {
	m.GotServiceAccount = serviceAccount
	return m.OnCreateServiceAccountError
}
func (m *AccountRepository) UpdateAclEntry(*models.AclEntry) error {
	return m.OnUpdateAclEntryError
}

func (m *AccountRepository) CreateClusterAccess(clusterAccess *models.ClusterAccess) error {
	m.GotClusterAccess = clusterAccess
	return m.OnCreateClusterAccessError
}

func (m *AccountRepository) UpdateClusterAccess(*models.ClusterAccess) error {
	return m.OnUpdateClusterAccessError
}
