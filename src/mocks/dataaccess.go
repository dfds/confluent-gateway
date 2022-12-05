package mocks

import (
	"context"
	"github.com/dfds/confluent-gateway/models"
)

type DataAccess struct {
	ClusterRepo models.ClusterRepository
}

func (d *DataAccess) NewSession(context.Context) models.DataSession {
	return d
}

func (d *DataAccess) Transaction(f func(models.DataSession) error) error {
	return f(d)
}

func (d *DataAccess) ServiceAccounts() models.ServiceAccountRepository {
	//TODO implement me
	panic("implement me")
}

func (d *DataAccess) Processes() models.ProcessRepository {
	//TODO implement me
	panic("implement me")
}

func (d *DataAccess) Clusters() models.ClusterRepository {
	return d.ClusterRepo
}

type ClusterRepositoryStub struct {
	Cluster models.Cluster
}

func (s *ClusterRepositoryStub) Get(context.Context, models.ClusterId) (models.Cluster, error) {
	return s.Cluster, nil
}
