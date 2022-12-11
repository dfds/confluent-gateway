package storage

import (
	"context"
	"github.com/dfds/confluent-gateway/logging"
	"github.com/dfds/confluent-gateway/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type database struct {
	db *gorm.DB
}

func NewDatabase(dsn string, logger logging.Logger) (*database, error) {
	config := gorm.Config{
		Logger: &databaseLogger{logger: logger},
	}
	if db, err := gorm.Open(postgres.Open(dsn), &config); err != nil {
		return nil, err
	} else {
		return &database{db}, nil
	}
}

func (d *database) NewSession(ctx context.Context) models.DataSession {
	return &database{d.db.Session(&gorm.Session{Context: ctx})}
}

func (d *database) Transaction(f func(models.DataSession) error) error {
	return d.db.Debug().Transaction(func(tx *gorm.DB) error {
		return f(&database{tx})
	})
}

func (d *database) ServiceAccounts() models.ServiceAccountRepository {
	return d
}

func (d *database) Processes() models.ProcessRepository {
	return d
}

func (d *database) Clusters() models.ClusterRepository {
	return d
}
