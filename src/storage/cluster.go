package storage

import (
	"context"
	"github.com/dfds/confluent-gateway/models"
)

func (s *dataSession) Get(ctx context.Context, id models.ClusterId) (models.Cluster, error) {
	var cluster models.Cluster

	err := s.db.WithContext(ctx).Find(&cluster, id).Error
	if err != nil {
		return models.Cluster{}, err
	}

	return cluster, nil
}
