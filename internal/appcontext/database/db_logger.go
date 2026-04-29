package database

import (
	serviceLogger "awesomeProject/logger"
	"context"
	"time"

	"gorm.io/gorm/logger"
)

type Dblogger struct{}

func (d *Dblogger) LogMode(level logger.LogLevel) logger.Interface {
	return d
}
func (d *Dblogger) Info(ctx context.Context, s string, i ...interface{}) {
	serviceLogger.Get(ctx).Infof(s, i...)
}
func (d *Dblogger) Warn(ctx context.Context, s string, i ...interface{}) {
	serviceLogger.Get(ctx).Warnf(s, i...)
}
func (d *Dblogger) Error(ctx context.Context, s string, i ...interface{}) {
	serviceLogger.Get(ctx).Errorf(s, i...)
}
func (d *Dblogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	sql, rowsAffected := fc()
	serviceLogger.Get(ctx).
		WithField("sql", sql).
		WithField("rowsAffected", rowsAffected).
		WithField("begin", begin.Format(time.RFC3339)).
		WithField("duration(ms)", time.Now().Sub(begin).Milliseconds()).
		WithField("err", err).Infof("query info")
}
