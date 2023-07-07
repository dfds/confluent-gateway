package mocks

import (
	"github.com/dfds/confluent-gateway/functional_tests/helpers"
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/internal/storage"
	"github.com/dfds/confluent-gateway/messaging"
	"gorm.io/gorm"
	"time"
)

type Database struct {
	*storage.Database
	rawDb *gorm.DB
}

func NewDatabase(database *storage.Database, rawDb *gorm.DB) *Database {
	return &Database{Database: database, rawDb: rawDb}
}

func (d *Database) GetAllOutboxEntries() ([]*messaging.OutboxEntry, error) {
	var outboxEntries []*messaging.OutboxEntry

	err := d.rawDb.Find(&outboxEntries).Error
	if err != nil {
		return nil, err
	}

	return outboxEntries, nil
}

func (d *Database) AddMockServiceAccountWithClusterAccess(serviceAccountId models.ServiceAccountId,
	userAccountId models.UserAccountId,
	clusterId models.ClusterId,
	capabilityId models.CapabilityId) (*models.ServiceAccount, error) {

	clusterAccess := []models.ClusterAccess{
		*models.NewClusterAccess(serviceAccountId, userAccountId, clusterId, capabilityId),
	}

	newServiceAccount := &models.ServiceAccount{
		Id:              serviceAccountId,
		UserAccountId:   userAccountId,
		CapabilityId:    capabilityId,
		ClusterAccesses: clusterAccess,
		CreatedAt:       time.Now(),
	}

	err := d.CreateServiceAccount(newServiceAccount)
	if err != nil {
		return nil, err
	}
	account, err := d.GetServiceAccount(newServiceAccount.CapabilityId)
	if err != nil {
		return nil, err
	}
	for _, access := range account.ClusterAccesses {
		for _, entry := range access.Acl {
			err = d.UpdateAclEntry(&entry)
			if err != nil {
				return nil, err
			}
		}
	}
	return newServiceAccount, nil
}

// RemoveServiceAccount removes service account, attached ACLs and cluster accesses
func (d *Database) RemoveServiceAccount(serviceAccount *models.ServiceAccount) error {
	if serviceAccount == nil {
		return nil
	}

	for _, access := range serviceAccount.ClusterAccesses {
		for _, entry := range access.Acl {
			d.rawDb.Delete(entry)
		}
		d.rawDb.Delete(access)
	}

	return d.rawDb.Delete(serviceAccount).Error
}

func (d *Database) RemoveCreateProcessesWithTopicId(topicId string) error {
	return d.rawDb.Delete(&models.CreateProcess{}, "topic_id = ?", topicId).Error
}

func (d *Database) RemoveDeleteProcessesWithTopicId(topicId string) error {
	return d.rawDb.Delete(&models.DeleteProcess{}, "topic_id = ?", topicId).Error

}

func (d *Database) RemoveSchemaProcessWithTopicId(topicId string) error {
	return d.rawDb.Delete(&models.SchemaProcess{}, "topic_id = ?", topicId).Error
}

// Full teardown functions

func (d *Database) RemoveAllCreateProcesses() error {
	return d.rawDb.Exec("DELETE FROM create_process").Error
}

func (d *Database) RemoveAllDeleteProcesses() error {
	return d.rawDb.Exec("DELETE FROM delete_process").Error
}

func (d *Database) RemoveAllTopics() error {
	return d.rawDb.Exec("DELETE FROM topic").Error
}

func (d *Database) RemoveAllOutboxEntries() error {
	return d.rawDb.Exec("DELETE FROM _outbox").Error
}

func (d *Database) RemoveAllServiceAccounts() helpers.ErrorList {

	var errlist helpers.ErrorList
	errlist.AppendIfErr(d.rawDb.Exec("DELETE FROM acl").Error)
	errlist.AppendIfErr(d.rawDb.Exec("DELETE FROM cluster_access").Error)
	errlist.AppendIfErr(d.rawDb.Exec("DELETE FROM service_account").Error)
	return errlist
}
