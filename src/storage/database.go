package storage

import (
	"context"
	"errors"
	"github.com/dfds/confluent-gateway/logging"
	"github.com/dfds/confluent-gateway/models"
	"github.com/dfds/confluent-gateway/process"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	db *gorm.DB
}

func NewDatabase(dsn string, logger logging.Logger) (*Database, error) {
	config := gorm.Config{
		Logger: &databaseLogger{logger: logger},
	}
	if db, err := gorm.Open(postgres.Open(dsn), &config); err != nil {
		return nil, err
	} else {
		return &Database{db}, nil
	}
}

func (d *Database) WithContext(ctx context.Context) process.Database {
	return &Database{d.db.Session(&gorm.Session{Context: ctx})}
}

func (d *Database) Transaction(f func(process.Transaction) error) error {
	return d.db.Debug().Transaction(func(tx *gorm.DB) error {
		return f(&Database{tx})
	})
}

func (d *Database) GetClusterById(ctx context.Context, id models.ClusterId) (*models.Cluster, error) {
	var cluster models.Cluster

	err := d.db.WithContext(ctx).First(&cluster, id).Error
	if err != nil {
		return nil, err
	}

	return &cluster, nil
}

func (d *Database) GetProcessState(capabilityRootId models.CapabilityRootId, clusterId models.ClusterId, topicName string) (*models.ProcessState, error) {
	var state = models.ProcessState{}

	err := d.db.
		Model(&state).
		First(&state, "capability_root_id = ? and cluster_id = ? and topic_name = ?", capabilityRootId, clusterId, topicName).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &state, nil
}

func (d *Database) CreateProcessState(state *models.ProcessState) error {
	return d.db.Create(state).Error
}

func (d *Database) UpdateProcessState(state *models.ProcessState) error {
	return d.db.Save(state).Error
}

func (d *Database) GetServiceAccount(capabilityRootId models.CapabilityRootId) (*models.ServiceAccount, error) {
	var serviceAccount models.ServiceAccount

	err := d.db.
		Model(&serviceAccount).
		Preload("ClusterAccesses").
		Preload("ClusterAccesses.Acl").
		First(&serviceAccount, "capability_root_id = ?", capabilityRootId).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &serviceAccount, nil
}

func (d *Database) CreateServiceAccount(serviceAccount *models.ServiceAccount) error {
	return d.db.Create(serviceAccount).Error
}

func (d *Database) UpdateAclEntry(aclEntry *models.AclEntry) error {
	return d.db.Save(aclEntry).Error
}

func (d *Database) CreateClusterAccess(clusterAccess *models.ClusterAccess) error {
	return d.db.Create(clusterAccess).Error
}

func (d *Database) UpdateClusterAccess(clusterAccess *models.ClusterAccess) error {
	return d.db.Save(clusterAccess).Error
}
