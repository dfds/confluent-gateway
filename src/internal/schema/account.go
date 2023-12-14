package schema

import (
	"context"
	"fmt"
	"github.com/dfds/confluent-gateway/internal/models"
)

type accountService struct {
	context   context.Context
	confluent Confluent
	repo      serviceAccountRepository
}

type serviceAccountRepository interface {
	GetServiceAccount(capabilityId models.CapabilityId) (*models.ServiceAccount, error)
}

func NewSchemaAccountService(ctx context.Context, confluent Confluent, repo serviceAccountRepository) *accountService {
	return &accountService{
		context:   ctx,
		confluent: confluent,
		repo:      repo,
	}
}

func (h *accountService) GetServiceAccount(capabilityId models.CapabilityId) (*models.ServiceAccount, error) {
	return h.repo.GetServiceAccount(capabilityId)
}

func (h *accountService) DeleteSchemaRegistryApiKey(clusterAccess *models.ClusterAccess) error {
	return h.confluent.DeleteSchemaRegistryApiKey(h.context, clusterAccess.ClusterId, clusterAccess.ServiceAccountId)
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

func (h *accountService) CountSchemaRegistryApiKeys(clusterAccess *models.ClusterAccess) (int, error) {
	keyCount, err := h.confluent.CountSchemaRegistryApiKeys(h.context, clusterAccess.ServiceAccountId, clusterAccess.ClusterId)
	if err != nil {
		return 0, err
	}
	return keyCount, nil
}

func (h *accountService) CreateSchemaRegistryApiKey(clusterAccess *models.ClusterAccess) (models.ApiKey, error) {
	return h.confluent.CreateSchemaRegistryApiKey(h.context, clusterAccess.ClusterId, clusterAccess.ServiceAccountId)

}

func (h *accountService) CreateServiceAccountRoleBinding(clusterAccess *models.ClusterAccess) error {
	err := h.confluent.CreateServiceAccountRoleBinding(h.context, clusterAccess.ServiceAccountId, clusterAccess.ClusterId)
	if err != nil {
		return err
	}
	return nil
}
