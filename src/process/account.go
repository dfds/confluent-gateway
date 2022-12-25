package process

import (
	"context"
	"fmt"
	"github.com/dfds/confluent-gateway/models"
	"time"
)

type AccountService struct {
	context    context.Context
	confluent  Confluent
	repository serviceAccountRepository
}

func NewAccountService(ctx context.Context, confluent Confluent, repository serviceAccountRepository) *AccountService {
	return &AccountService{
		context:    ctx,
		confluent:  confluent,
		repository: repository,
	}
}

func (h *AccountService) CreateServiceAccount(capabilityRootId models.CapabilityRootId, clusterId models.ClusterId) error {
	serviceAccountId, err := h.confluent.CreateServiceAccount(h.context, "sa-some-name", "sa description")
	if err != nil {
		return err
	}

	newServiceAccount := &models.ServiceAccount{
		Id:               *serviceAccountId,
		CapabilityRootId: capabilityRootId,
		ClusterAccesses:  []models.ClusterAccess{*models.NewClusterAccess(*serviceAccountId, clusterId, capabilityRootId)},
		CreatedAt:        time.Now(),
	}

	return h.repository.CreateServiceAccount(newServiceAccount)
}

func (h *AccountService) GetOrCreateClusterAccess(capabilityRootId models.CapabilityRootId, clusterId models.ClusterId) (*models.ClusterAccess, error) {
	serviceAccount, err := h.repository.GetServiceAccount(capabilityRootId)
	if err != nil {
		return nil, err
	}
	if serviceAccount == nil {
		return nil, fmt.Errorf("no service account for capability '%s' found", capabilityRootId)
	}

	clusterAccess, hasClusterAccess := serviceAccount.TryGetClusterAccess(clusterId)

	if !hasClusterAccess {
		clusterAccess = models.NewClusterAccess(serviceAccount.Id, clusterId, capabilityRootId)
		serviceAccount.ClusterAccesses = append(serviceAccount.ClusterAccesses, *clusterAccess)

		if err = h.repository.CreateClusterAccess(clusterAccess); err != nil {
			return nil, err
		}
	}
	return clusterAccess, nil
}

func (h *AccountService) GetClusterAccess(capabilityRootId models.CapabilityRootId, clusterId models.ClusterId) (*models.ClusterAccess, error) {
	serviceAccount, err := h.repository.GetServiceAccount(capabilityRootId)
	if err != nil {
		return nil, err
	}
	if serviceAccount == nil {
		return nil, fmt.Errorf("no service account for capability '%s' found", capabilityRootId)
	}

	clusterAccess, hasClusterAccess := serviceAccount.TryGetClusterAccess(clusterId)

	if !hasClusterAccess {
		return nil, fmt.Errorf("no cluster access for service account '%s' found", serviceAccount.Id)
	}
	return clusterAccess, nil
}

func (h *AccountService) CreateAclEntry(clusterId models.ClusterId, serviceAccountId models.ServiceAccountId, entry *models.AclEntry) error {
	if err := h.confluent.CreateACLEntry(h.context, clusterId, serviceAccountId, entry.AclDefinition); err != nil {
		return err
	}

	entry.Created()

	return h.repository.UpdateAclEntry(entry)
}

func (h *AccountService) CreateApiKey(clusterAccess *models.ClusterAccess) error {
	key, err := h.confluent.CreateApiKey(h.context, clusterAccess.ClusterId, clusterAccess.ServiceAccountId)
	if err != nil {
		return err
	}

	clusterAccess.ApiKey = *key

	return h.repository.UpdateClusterAccess(clusterAccess)
}
