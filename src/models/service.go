package models

import (
	"context"
	"time"
)

type Service struct {
	confluent  Confluent
	repository ServiceAccountRepository
}

func NewService(confluent Confluent, repository ServiceAccountRepository) *Service {
	return &Service{confluent: confluent, repository: repository}
}

func (s *Service) CreateServiceAccount(capabilityRootId CapabilityRootId, clusterId ClusterId) error {
	serviceAccountId, err := s.confluent.CreateServiceAccount(context.TODO(), "sa-some-name", "sa description")
	if err != nil {
		return err
	}

	newServiceAccount := &ServiceAccount{
		Id:               *serviceAccountId,
		CapabilityRootId: capabilityRootId,
		ClusterAccesses:  []ClusterAccess{*NewClusterAccess(*serviceAccountId, clusterId, capabilityRootId)},
		CreatedAt:        time.Now(),
	}

	return s.repository.CreateServiceAccount(newServiceAccount)
}

func (s *Service) GetOrCreateClusterAccess(capabilityRootId CapabilityRootId, clusterId ClusterId) (*ClusterAccess, error) {
	serviceAccount, err := s.repository.GetServiceAccount(capabilityRootId)
	if err != nil {
		return nil, err
	}

	clusterAccess, hasClusterAccess := serviceAccount.TryGetClusterAccess(clusterId)

	if !hasClusterAccess {
		clusterAccess = NewClusterAccess(serviceAccount.Id, clusterId, capabilityRootId)
		serviceAccount.ClusterAccesses = append(serviceAccount.ClusterAccesses, *clusterAccess)

		if err = s.repository.CreateClusterAccess(clusterAccess); err != nil {
			return nil, err
		}
	}
	return clusterAccess, nil
}

func (s *Service) CreateAclEntry(clusterId ClusterId, clusterAccess *ClusterAccess, entry *AclEntry) error {
	if err := s.confluent.CreateACLEntry(context.TODO(), clusterId, clusterAccess.ServiceAccountId, entry.AclDefinition); err != nil {
		return err
	}

	entry.Created()

	return s.repository.UpdateAclEntry(entry)
}

func (s *Service) CreateApiKey(clusterAccess *ClusterAccess) error {
	key, err := s.confluent.CreateApiKey(context.TODO(), clusterAccess.ClusterId, clusterAccess.ServiceAccountId)
	if err != nil {
		return err
	}

	clusterAccess.ApiKey = *key

	return s.repository.UpdateClusterAccess(clusterAccess)
}

func (s *Service) CreateTopic(clusterId ClusterId, topic Topic) error {
	return s.confluent.CreateTopic(context.TODO(), clusterId, topic.Name, topic.Partitions, topic.Retention)
}
