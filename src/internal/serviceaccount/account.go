package serviceaccount

import (
	"context"
	"errors"
	"fmt"
	"github.com/dfds/confluent-gateway/internal/confluent"
	"time"

	"github.com/dfds/confluent-gateway/internal/models"
)

type accountService struct {
	context   context.Context
	confluent Confluent
	repo      serviceAccountRepository
}

type serviceAccountRepository interface {
	GetServiceAccount(capabilityId models.CapabilityId) (*models.ServiceAccount, error)
	CreateServiceAccount(serviceAccount *models.ServiceAccount) error
	UpdateAclEntry(aclEntry *models.AclEntry) error
	CreateClusterAccess(clusterAccess *models.ClusterAccess) error
	UpdateClusterAccess(clusterAccess *models.ClusterAccess) error
}

func NewAccountService(ctx context.Context, confluent Confluent, repo serviceAccountRepository) *accountService {
	return &accountService{
		context:   ctx,
		confluent: confluent,
		repo:      repo,
	}
}

func (h *accountService) GetServiceAccount(capabilityId models.CapabilityId) (*models.ServiceAccount, error) {
	return h.repo.GetServiceAccount(capabilityId)
}

func (h *accountService) CreateServiceAccount(capabilityId models.CapabilityId, clusterId models.ClusterId) error {
	serviceAccountId, err := h.confluent.CreateServiceAccount(h.context, string(capabilityId), "Created by Confluent Gateway")
	if err != nil {
		return err
	}

	users, err := h.confluent.GetConfluentInternalUsers(h.context)
	if err != nil {
		return err
	}

	userMap := make(map[string]models.ConfluentInternalUser)

	for _, user := range users {
		userMap[user.ResourceID] = user
	}

	user, found := userMap[string(serviceAccountId)]
	if !found {
		return fmt.Errorf("unable to find matching user account for %q", serviceAccountId)
	}
	if user.Deactivated {
		return fmt.Errorf("found matching user account for %q, but user is deactivated", serviceAccountId)
	}

	userAccountId := models.MakeUserAccountId(user.Id)

	newServiceAccount := &models.ServiceAccount{
		Id:              serviceAccountId,
		UserAccountId:   userAccountId,
		CapabilityId:    capabilityId,
		ClusterAccesses: []models.ClusterAccess{*models.NewClusterAccess(serviceAccountId, userAccountId, clusterId, capabilityId)},
		CreatedAt:       time.Now(),
	}

	return h.repo.CreateServiceAccount(newServiceAccount)
}

func (h *accountService) GetOrCreateClusterAccess(capabilityId models.CapabilityId, clusterId models.ClusterId) (*models.ClusterAccess, error) {
	serviceAccount, err := h.repo.GetServiceAccount(capabilityId)
	if err != nil {
		return nil, err
	}
	if serviceAccount == nil {
		return nil, fmt.Errorf("no service account for capability '%s' found", capabilityId)
	}

	clusterAccess, hasClusterAccess := serviceAccount.TryGetClusterAccess(clusterId)

	if !hasClusterAccess {
		clusterAccess = models.NewClusterAccess(serviceAccount.Id, serviceAccount.UserAccountId, clusterId, capabilityId)
		serviceAccount.ClusterAccesses = append(serviceAccount.ClusterAccesses, *clusterAccess)

		if err = h.repo.CreateClusterAccess(clusterAccess); err != nil {
			return nil, err
		}
	}
	return clusterAccess, nil
}

func (h *accountService) GetClusterAccess(capabilityId models.CapabilityId, clusterId models.ClusterId) (*models.ClusterAccess, error) {
	serviceAccount, err := h.repo.GetServiceAccount(capabilityId)
	if err != nil {
		return nil, err
	}
	if serviceAccount == nil {
		return nil, fmt.Errorf("no service account for capability '%s' found", capabilityId)
	}

	clusterAccess, hasClusterAccess := serviceAccount.TryGetClusterAccess(clusterId)

	if !hasClusterAccess {
		return nil, fmt.Errorf("no cluster access for service account '%s' found", serviceAccount.Id)
	}
	return clusterAccess, nil
}

func (h *accountService) CreateAclEntry(clusterId models.ClusterId, userAccountId models.UserAccountId, entry *models.AclEntry) error {
	if err := h.confluent.CreateACLEntry(h.context, clusterId, userAccountId, entry.AclDefinition); err != nil {
		return err
	}

	entry.Created()

	return h.repo.UpdateAclEntry(entry)
}

func (h *accountService) CreateClusterApiKey(clusterAccess *models.ClusterAccess) error {
	key, err := h.confluent.CreateClusterApiKey(h.context, clusterAccess.ClusterId, clusterAccess.ServiceAccountId)
	if err != nil {
		return err
	}

	clusterAccess.ApiKey = *key

	return h.repo.UpdateClusterAccess(clusterAccess)
}

func (h *accountService) CountClusterApiKeys(clusterAccess *models.ClusterAccess) (int, error) {
	keyCount, err := h.confluent.CountClusterApiKeys(h.context, clusterAccess.ServiceAccountId, clusterAccess.ClusterId)
	if err != nil {
		return 0, err
	}
	return keyCount, nil
}

func (h *accountService) CountSchemaRegistryApiKeys(clusterAccess *models.ClusterAccess) (int, error) {
	keyCount, err := h.confluent.CountSchemaRegistryApiKeys(h.context, clusterAccess.ServiceAccountId, clusterAccess.ClusterId)
	if err != nil {
		return 0, err
	}
	return keyCount, nil
}

func (h *accountService) RecreateSchemaRegistryApiKeyAndStoreInDb(clusterAccess *models.ClusterAccess) error {

	err := h.confluent.DeleteSchemaRegistryApiKey(h.context, clusterAccess.ClusterId, clusterAccess.ServiceAccountId)
	if err != nil && !errors.Is(err, confluent.ErrSchemaRegistryApiKeyNotFoundForDeletion) {
		return err
	}

	if errors.Is(err, confluent.ErrSchemaRegistryApiKeyNotFoundForDeletion) {
		// TODO: log?
	}

	return h.CreateSchemaRegistryApiKey(clusterAccess)
}

func (h *accountService) CreateSchemaRegistryApiKey(clusterAccess *models.ClusterAccess) error {
	key, err := h.confluent.CreateSchemaRegistryApiKey(h.context, clusterAccess.ClusterId, clusterAccess.ServiceAccountId)
	if err != nil {
		return err
	}

	clusterAccess.SchemaRegistryApiKey = *key

	return h.repo.UpdateClusterAccess(clusterAccess)
}

func (h *accountService) CreateServiceAccountRoleBinding(clusterAccess *models.ClusterAccess) error {
	err := h.confluent.CreateServiceAccountRoleBinding(h.context, clusterAccess.ServiceAccountId, clusterAccess.ClusterId)
	if err != nil {
		return err
	}
	return nil
}
