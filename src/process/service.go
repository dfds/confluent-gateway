package process

import (
	"context"
	"github.com/dfds/confluent-gateway/models"
	"time"
)

type Service struct {
	context    context.Context
	confluent  Confluent
	repository serviceAccountRepository
}

func NewService(ctx context.Context, confluent Confluent, repository serviceAccountRepository) *Service {
	return &Service{
		context:    ctx,
		confluent:  confluent,
		repository: repository,
	}
}

func (s *Service) CreateServiceAccount(capabilityRootId models.CapabilityRootId, clusterId models.ClusterId) error {
	serviceAccountId, err := s.confluent.CreateServiceAccount(s.context, "sa-some-name", "sa description")
	if err != nil {
		return err
	}

	newServiceAccount := &models.ServiceAccount{
		Id:               *serviceAccountId,
		CapabilityRootId: capabilityRootId,
		ClusterAccesses:  []models.ClusterAccess{*models.NewClusterAccess(*serviceAccountId, clusterId, capabilityRootId)},
		CreatedAt:        time.Now(),
	}

	return s.repository.CreateServiceAccount(newServiceAccount)
}

func (s *Service) GetOrCreateClusterAccess(capabilityRootId models.CapabilityRootId, clusterId models.ClusterId) (*models.ClusterAccess, error) {
	serviceAccount, err := s.repository.GetServiceAccount(capabilityRootId)
	if err != nil {
		return nil, err
	}

	clusterAccess, hasClusterAccess := serviceAccount.TryGetClusterAccess(clusterId)

	if !hasClusterAccess {
		clusterAccess = models.NewClusterAccess(serviceAccount.Id, clusterId, capabilityRootId)
		serviceAccount.ClusterAccesses = append(serviceAccount.ClusterAccesses, *clusterAccess)

		if err = s.repository.CreateClusterAccess(clusterAccess); err != nil {
			return nil, err
		}
	}
	return clusterAccess, nil
}

func (s *Service) CreateAclEntry(clusterId models.ClusterId, clusterAccess *models.ClusterAccess, entry *models.AclEntry) error {
	if err := s.confluent.CreateACLEntry(s.context, clusterId, clusterAccess.ServiceAccountId, entry.AclDefinition); err != nil {
		return err
	}

	entry.Created()

	return s.repository.UpdateAclEntry(entry)
}

func (s *Service) CreateApiKey(clusterAccess *models.ClusterAccess) error {
	key, err := s.confluent.CreateApiKey(s.context, clusterAccess.ClusterId, clusterAccess.ServiceAccountId)
	if err != nil {
		return err
	}

	clusterAccess.ApiKey = *key

	return s.repository.UpdateClusterAccess(clusterAccess)
}

func (s *Service) CreateTopic(clusterId models.ClusterId, topic models.Topic) error {
	return s.confluent.CreateTopic(s.context, clusterId, topic.Name, topic.Partitions, topic.Retention)
}
