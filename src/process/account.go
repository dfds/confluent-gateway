package process

import (
	"fmt"
	"github.com/dfds/confluent-gateway/models"
	"time"
)

type AccountService interface {
	CapabilityRootId() models.CapabilityRootId
	ClusterId() models.ClusterId
	CreateServiceAccount() (*models.ServiceAccountId, error)
	SaveServiceAccount(*models.ServiceAccount) error
	GetServiceAccount() (*models.ServiceAccount, error)
	CreateClusterAccess(*models.ClusterAccess) error
	CreateAclEntry(models.ServiceAccountId, *models.AclEntry) error
	UpdateAclEntry(*models.AclEntry) error
	CreateApiKey(*models.ClusterAccess) (*models.ApiKey, error)
	UpdateClusterAccess(*models.ClusterAccess) error
}

type AccountHelper struct {
	service AccountService
}

func NewAccountHelper(service AccountService) *AccountHelper {
	return &AccountHelper{service}
}

func (h *AccountHelper) CreateServiceAccount() error {
	serviceAccountId, err := h.service.CreateServiceAccount()
	if err != nil {
		return err
	}

	newServiceAccount := &models.ServiceAccount{
		Id:               *serviceAccountId,
		CapabilityRootId: h.service.CapabilityRootId(),
		ClusterAccesses:  []models.ClusterAccess{*models.NewClusterAccess(*serviceAccountId, h.service.ClusterId(), h.service.CapabilityRootId())},
		CreatedAt:        time.Now(),
	}

	return h.service.SaveServiceAccount(newServiceAccount)
}

func (h *AccountHelper) GetOrCreateClusterAccess() (*models.ClusterAccess, error) {
	s := h.service
	serviceAccount, err := s.GetServiceAccount()
	if err != nil {
		return nil, err
	}
	if serviceAccount == nil {
		return nil, fmt.Errorf("no service account for capability '%s' found", s.CapabilityRootId())
	}

	clusterAccess, hasClusterAccess := serviceAccount.TryGetClusterAccess(s.ClusterId())

	if !hasClusterAccess {
		clusterAccess = models.NewClusterAccess(serviceAccount.Id, s.ClusterId(), s.CapabilityRootId())
		serviceAccount.ClusterAccesses = append(serviceAccount.ClusterAccesses, *clusterAccess)

		if err = s.CreateClusterAccess(clusterAccess); err != nil {
			return nil, err
		}
	}
	return clusterAccess, nil
}

func (h *AccountHelper) CreateAclEntry(serviceAccountId models.ServiceAccountId, entry *models.AclEntry) error {
	s := h.service
	if err := s.CreateAclEntry(serviceAccountId, entry); err != nil {
		return err
	}

	entry.Created()

	return s.UpdateAclEntry(entry)
}

func (h *AccountHelper) CreateApiKey(clusterAccess *models.ClusterAccess) error {
	s := h.service
	key, err := s.CreateApiKey(clusterAccess)
	if err != nil {
		return err
	}

	clusterAccess.ApiKey = *key

	return s.UpdateClusterAccess(clusterAccess)
}
