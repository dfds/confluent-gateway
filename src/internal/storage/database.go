package storage

import (
	"context"
	"errors"
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/logging"
	"github.com/dfds/confluent-gateway/messaging"
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

func (d *Database) NewSession(ctx context.Context) models.Session {
	return &Database{d.db.Session(&gorm.Session{Context: ctx})}
}

func (d *Database) Transaction(f func(models.Transaction) error) error {
	return d.db.Debug().Transaction(func(tx *gorm.DB) error {
		return f(&Database{tx})
	})
}

func (d *Database) GetClusters(ctx context.Context) ([]*models.Cluster, error) {
	var clusters []*models.Cluster

	err := d.db.WithContext(ctx).Find(&clusters).Error
	if err != nil {
		return nil, err
	}

	return clusters, nil
}

func (d *Database) GetCreateProcessState(capabilityId models.CapabilityId, clusterId models.ClusterId, topicName string) (*models.CreateProcess, error) {
	var state = models.CreateProcess{}

	err := d.db.
		Model(&state).
		First(&state, "capability_id = ? and cluster_id = ? and topic_name = ?", capabilityId, clusterId, topicName).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &state, nil
}

func (d *Database) SaveCreateProcessState(state *models.CreateProcess) error {
	return d.db.Create(state).Error
}

func (d *Database) UpdateCreateProcessState(state *models.CreateProcess) error {
	return d.db.Save(state).Error
}

func (d *Database) GetDeleteProcessState(topicId string) (*models.DeleteProcess, error) {
	var state = models.DeleteProcess{}

	err := d.db.
		Model(&state).
		First(&state, "topic_id = ?", topicId).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &state, nil
}

func (d *Database) SaveDeleteProcessState(state *models.DeleteProcess) error {
	return d.db.Create(state).Error
}

func (d *Database) UpdateDeleteProcessState(state *models.DeleteProcess) error {
	return d.db.Save(state).Error
}

func (d *Database) GetServiceAccount(capabilityId models.CapabilityId) (*models.ServiceAccount, error) {
	var serviceAccount models.ServiceAccount

	err := d.db.
		Model(&serviceAccount).
		Preload("ClusterAccesses").
		Preload("ClusterAccesses.Acl").
		First(&serviceAccount, "capability_id = ?", capabilityId).
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

func (d *Database) AddToOutbox(entry *messaging.OutboxEntry) error {
	return d.db.Create(entry).Error
}

func (d *Database) CreateTopic(topic *models.Topic) error {
	return d.db.Create(topic).Error
}

func (d *Database) GetTopic(topicId string) (*models.Topic, error) {
	var topic = &models.Topic{}

	err := d.db.First(topic, "id = ?", topicId).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return topic, nil
}

func (d *Database) DeleteTopic(topicId string) error {
	return d.db.Delete(&models.Topic{}, "id = ?", topicId).Error
}

func (d *Database) GetSchemaProcessState(messageContractId string) (*models.SchemaProcess, error) {
	var schema = &models.SchemaProcess{}

	err := d.db.
		Model(schema).
		First(schema, "message_contract_id = ?", messageContractId).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return schema, nil
}

func (d *Database) SaveSchemaProcessState(schema *models.SchemaProcess) error {
	return d.db.Create(schema).Error
}

func (d *Database) UpdateSchemaProcessState(schema *models.SchemaProcess) error {
	return d.db.Save(schema).Error
}
