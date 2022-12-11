package storage

import (
	"context"
	"github.com/dfds/confluent-gateway/logging"
	"github.com/dfds/confluent-gateway/models"
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

func (d *Database) NewSession(ctx context.Context) models.DataSession {
	return &Database{d.db.Session(&gorm.Session{Context: ctx})}
}

func (d *Database) Transaction(f func(models.DataSession) error) error {
	return d.db.Debug().Transaction(func(tx *gorm.DB) error {
		return f(&Database{tx})
	})
}

func (d *Database) ServiceAccounts() models.ServiceAccountRepository {
	return d
}

func (d *Database) Processes() models.ProcessRepository {
	return d
}

func (d *Database) Clusters() models.ClusterRepository {
	return d
}
