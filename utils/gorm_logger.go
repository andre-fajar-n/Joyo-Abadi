package utils

import (
	"context"
	"time"

	"gorm.io/gorm/logger"
)

type GormLogger struct{}

func (gl *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return gl
}

func (gl *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	Log.Infof(msg, data...)
}

func (gl *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	Log.Warnf(msg, data...)
}

func (gl *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	Log.Errorf(msg, data...)
}

func (gl *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	if err != nil {
		Log.WithError(err).Errorf("[%.3fms] [rows:%v] %s", float64(elapsed.Nanoseconds())/1e6, rows, sql)
	} else {
		Log.Infof("[%.3fms] [rows:%v] %s", float64(elapsed.Nanoseconds())/1e6, rows, sql)
	}
}
