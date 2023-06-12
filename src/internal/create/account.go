package create

import (
	"context"
	"fmt"

	"github.com/dfds/confluent-gateway/internal/models"
)

type accountService struct {
	context context.Context
	repo    serviceAccountRepository
}

type serviceAccountRepository interface {
	GetServiceAccount(capabilityId models.CapabilityId) (*models.ServiceAccount, error)
}

func NewAccountService(ctx context.Context, repo serviceAccountRepository) *accountService {
	return &accountService{
		context: ctx,
		repo:    repo,
	}
}

func (h *accountService) GetServiceAccount(capabilityId models.CapabilityId) (*models.ServiceAccount, error) {
	return h.repo.GetServiceAccount(capabilityId)
}

func (h *accountService) HasClusterAccess(capabilityId models.CapabilityId, clusterId models.ClusterId) (bool, error) {
	serviceAccount, err := h.repo.GetServiceAccount(capabilityId)
	if err != nil {
		return false, err
	}
	if serviceAccount == nil {
		return false, fmt.Errorf("no service account for capability '%s' found", capabilityId)
	}

	_, hasClusterAccess := serviceAccount.TryGetClusterAccess(clusterId)

	if !hasClusterAccess {
		return false, fmt.Errorf("no cluster access for service account '%s' found", serviceAccount.Id)
	}
	return true, nil
}
