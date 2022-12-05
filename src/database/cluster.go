package database

import (
	"context"
	"github.com/dfds/confluent-gateway/models"
	"gorm.io/gorm"
)

type clusterRepository struct {
	db *gorm.DB
}

func (r *clusterRepository) Get(ctx context.Context, id models.ClusterId) (models.Cluster, error) {
	var cluster models.Cluster

	err := r.db.WithContext(ctx).Find(&cluster, id).Error
	if err != nil {
		return models.Cluster{}, err
	}

	return cluster, nil
}

func NewClusterRepository(db *gorm.DB) models.ClusterRepository {
	return &clusterRepository{
		db: db,
	}
}
