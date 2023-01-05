package process

import (
	"context"
	"fmt"
	"github.com/dfds/confluent-gateway/models"
	"time"
)

type accountService struct {
	context    context.Context
	confluent  Confluent
	repository serviceAccountRepository
}

func NewAccountService(ctx context.Context, confluent Confluent, repository serviceAccountRepository) *accountService {
	return &accountService{
		context:    ctx,
		confluent:  confluent,
		repository: repository,
	}
}

func (h *accountService) CreateServiceAccount(capabilityRootId models.CapabilityRootId, clusterId models.ClusterId) error {
	serviceAccountId, err := h.confluent.CreateServiceAccount(h.context, string(capabilityRootId), "Created by Confluent Gateway")
	if err != nil {
		return err
	}

	users, err := h.confluent.GetUsers(h.context)
	if err != nil {
		return err
	}

	userMap := make(map[string]models.User)

	for _, user := range users {
		userMap[user.ResourceID] = user
	}

	user, found := userMap[string(serviceAccountId)]
	if !found || user.Deactivated {
		return fmt.Errorf("unable to find matching user account for %q", serviceAccountId)
	}

	userAccountId := user.Id

	newServiceAccount := &models.ServiceAccount{
		Id:               serviceAccountId,
		UserAccountId:    userAccountId,
		CapabilityRootId: capabilityRootId,
		ClusterAccesses:  []models.ClusterAccess{*models.NewClusterAccess(serviceAccountId, userAccountId, clusterId, capabilityRootId)},
		CreatedAt:        time.Now(),
	}

	return h.repository.CreateServiceAccount(newServiceAccount)
}

func (h *accountService) GetOrCreateClusterAccess(capabilityRootId models.CapabilityRootId, clusterId models.ClusterId) (*models.ClusterAccess, error) {
	serviceAccount, err := h.repository.GetServiceAccount(capabilityRootId)
	if err != nil {
		return nil, err
	}
	if serviceAccount == nil {
		return nil, fmt.Errorf("no service account for capability '%s' found", capabilityRootId)
	}

	clusterAccess, hasClusterAccess := serviceAccount.TryGetClusterAccess(clusterId)

	if !hasClusterAccess {
		clusterAccess = models.NewClusterAccess(serviceAccount.Id, serviceAccount.UserAccountId, clusterId, capabilityRootId)
		serviceAccount.ClusterAccesses = append(serviceAccount.ClusterAccesses, *clusterAccess)

		if err = h.repository.CreateClusterAccess(clusterAccess); err != nil {
			return nil, err
		}
	}
	return clusterAccess, nil
}

func (h *accountService) GetClusterAccess(capabilityRootId models.CapabilityRootId, clusterId models.ClusterId) (*models.ClusterAccess, error) {
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

func (h *accountService) CreateAclEntry(clusterId models.ClusterId, userAccountId int, entry *models.AclEntry) error {
	if err := h.confluent.CreateACLEntry(h.context, clusterId, userAccountId, entry.AclDefinition); err != nil {
		return err
	}

	entry.Created()

	return h.repository.UpdateAclEntry(entry)
}

func (h *accountService) CreateApiKey(clusterAccess *models.ClusterAccess) error {
	key, err := h.confluent.CreateApiKey(h.context, clusterAccess.ClusterId, clusterAccess.ServiceAccountId)
	if err != nil {
		return err
	}

	clusterAccess.ApiKey = *key

	return h.repository.UpdateClusterAccess(clusterAccess)
}
