package utils

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
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

	// Create log fields
	fields := logrus.Fields{
		"duration_ms": float64(elapsed.Nanoseconds()) / 1e6,
		"rows":        rows,
		"sql":         sql,
	}

	// Log based on error and duration
	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		Log.WithFields(fields).WithError(err).Error("SQL error")
	case elapsed > time.Millisecond*500:
		Log.WithFields(fields).Warn("Slow SQL query (>500ms)")
	case elapsed > time.Millisecond*100:
		Log.WithFields(fields).Info("SQL query (>100ms)")
	case isImportantQuery(sql):
		Log.WithFields(fields).Info("SQL query")
	default:
		Log.WithFields(fields).Debug("SQL query")
	}
}

// isImportantQuery determines if a SQL query should be logged at INFO level
func isImportantQuery(sql string) bool {
	// Log important operations at INFO level even if they're fast
	importantKeywords := []string{
		"INSERT", "UPDATE", "DELETE",
		"CREATE TABLE", "ALTER TABLE", "DROP TABLE",
		"CREATE INDEX", "DROP INDEX",
	}

	sqlUpper := strings.ToUpper(sql)
	for _, keyword := range importantKeywords {
		if strings.Contains(sqlUpper, keyword) {
			return true
		}
	}
	return false
}
