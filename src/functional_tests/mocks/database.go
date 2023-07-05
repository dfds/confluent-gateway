package mocks

import (
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/internal/storage"
	"github.com/dfds/confluent-gateway/messaging"
	"gorm.io/gorm"
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

func (d *Database) RemoveAllOutboxEntries() error {

	outboxEntries, err := d.GetAllOutboxEntries()
	if err != nil {
		return err
	}
	if len(outboxEntries) > 0 {
		err = d.rawDb.Delete(outboxEntries).Error
		if err != nil {
			return err
		}
	}

	return nil
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
