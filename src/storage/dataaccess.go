package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/dfds/confluent-gateway/logging"
	"github.com/dfds/confluent-gateway/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

type dataAccess struct {
	db *gorm.DB
}

func (da *dataAccess) NewSession(ctx context.Context) models.DataSession {
	return &dataSession{da.db.Session(&gorm.Session{Context: ctx})}
}

func NewDatabase(dsn string, log logging.Logger) (models.DataAccess, error) {
	config := gorm.Config{
		Logger: &databaseLogger{logger: log},
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

func (s *dataSession) Transaction(f func(models.DataSession) error) error {
	return s.db.Debug().Transaction(func(tx *gorm.DB) error {
		return f(&dataSession{tx})
	})
}

func (s *dataSession) ServiceAccounts() models.ServiceAccountRepository {
	return s
}

func (s *dataSession) Processes() models.ProcessRepository {
	return s
}

func (s *dataSession) Clusters() models.ClusterRepository {
	return s
}

// region logging

type databaseLogger struct {
	logger                    logging.Logger
	level                     logger.LogLevel
	ignoreRecordNotFoundError bool
}

func (l *databaseLogger) LogMode(level logger.LogLevel) logger.Interface {
	l.level = level
	return l
}

func (l *databaseLogger) Info(_ context.Context, s string, i ...interface{}) {
	if l.level >= logger.Info {
		l.logger.Information(fmt.Sprintf(s, i...))
	}
}

func (l *databaseLogger) Warn(_ context.Context, s string, i ...interface{}) {
	if l.level >= logger.Warn {
		l.logger.Warning(fmt.Sprintf(s, i...))
	}
}

func (l *databaseLogger) Error(_ context.Context, s string, i ...interface{}) {
	if l.level >= logger.Error {
		l.logger.Error(nil, fmt.Sprintf(s, i...))
	}
}

func (l *databaseLogger) Trace(_ context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.level <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)

	switch {
	case err != nil && l.level >= logger.Error && (!errors.Is(err, gorm.ErrRecordNotFound) || !l.ignoreRecordNotFoundError):
		sql, rows := fc()

		if rows == -1 {
			s := fmt.Sprintf("[%.3fms] [rows:%v] %s", float64(elapsed.Nanoseconds())/1e6, "-", sql)
			l.logger.Error(&err, s)

		} else {
			s := fmt.Sprintf("[%.3fms] [rows:%v] %s", float64(elapsed.Nanoseconds())/1e6, rows, sql)
			l.logger.Error(&err, s)
		}
	case l.level == logger.Info:
		sql, rows := fc()
		if rows == -1 {
			s := fmt.Sprintf("[%.3fms] [rows:%v] %s", float64(elapsed.Nanoseconds())/1e6, "-", sql)
			l.logger.Trace(s)
		} else {
			s := fmt.Sprintf("[%.3fms] [rows:%v] %s", float64(elapsed.Nanoseconds())/1e6, rows, sql)
			l.logger.Trace(s)
		}
	}
}

// endregion
