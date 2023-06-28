package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/dfds/confluent-gateway/logging"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

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

func formatMessage(rows int64, sql string, elapsed time.Duration) string {

	rowsString := "-"
	if rows != -1 {
		rowsString = fmt.Sprintf("%v", rows)
	}
	return fmt.Sprintf("[DB_LOGGER]: [%.3fms] [rows:%v] %s", float64(elapsed.Nanoseconds())/1e6, rowsString, sql)
}

func (l *databaseLogger) Trace(_ context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.level <= logger.Silent {
		return
	}

	sql, rows := fc()
	msg := formatMessage(rows, sql, time.Since(begin))
	switch {
	case err != nil && l.level >= logger.Error && (!errors.Is(err, gorm.ErrRecordNotFound) || !l.ignoreRecordNotFoundError):
		l.logger.Error(err, msg)
	case l.level == logger.Info:
		l.logger.Trace(msg)
	}
}
