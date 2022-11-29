package database

import (
	"context"
	"github.com/dfds/confluent-gateway/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type dataAccess struct {
	db *gorm.DB
}

func (da *dataAccess) Transaction(f func(models.DataAccess) error) error {
	return da.db.Debug().Transaction(func(tx *gorm.DB) error {
		return f(&dataAccess{tx})
	})
}

func (da *dataAccess) ServiceAccounts() models.ServiceAccountRepository {
	return NewServiceAccountRepository(da.db)
}

func (da *dataAccess) Processes() models.ProcessRepository {
	return NewProcessRepository(da.db)
}

func (da *dataAccess) NewSession(ctx context.Context) models.DataSession {
	return &dataSession{da.db.Session(&gorm.Session{Context: ctx})}
}

func NewDatabase(dsn string) (models.DataAccess, error) {
	config := gorm.Config{
		//	//Logger: logger.Default.LogMode(logger.Silent),
	}
	if db, err := gorm.Open(postgres.Open(dsn), &config); err != nil {
		return nil, err
	} else {
		return &dataAccess{db}, nil
	}
}

type dataSession struct {
	db *gorm.DB
}

func (da *dataSession) Transaction(f func(models.DataSession) error) error {
	return da.db.Debug().Transaction(func(tx *gorm.DB) error {
		return f(&dataSession{tx})
	})
}

func (da *dataSession) ServiceAccounts() models.ServiceAccountRepository {
	return NewServiceAccountRepository(da.db)
}

func (da *dataSession) Processes() models.ProcessRepository {
	return NewProcessRepository(da.db)
}
