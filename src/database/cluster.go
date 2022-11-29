package database

import (
	"context"
	"github.com/dfds/confluent-gateway/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type clusterRepository struct {
	db *gorm.DB
}

func (r *clusterRepository) Get(ctx context.Context, id models.ClusterId) (models.Cluster, error) {
	//TODO implement me
	panic("implement me")
}

func (r *clusterRepository) GetAll(ctx context.Context) ([]models.Cluster, error) {

	var clusters []models.Cluster

	err := r.db.WithContext(ctx).Find(&clusters).Error
	if err != nil {
		return nil, err
	}

	return clusters, nil
}

func NewClusterRepository(dsn string) (models.ClusterRepository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &clusterRepository{
		db: db,
	}, nil
}
