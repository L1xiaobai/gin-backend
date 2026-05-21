package db

import (
	"context"
	"errors"
	"time"

	"go-test/internal/global"
	"go-test/pkg/xcontext"

	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type GormLogger struct {
	SlowThreshold time.Duration
	LogLevel      gormlogger.LogLevel
}

func NewGormLogger(slowThreshold time.Duration, logLevel gormlogger.LogLevel) *GormLogger {
	return &GormLogger{
		SlowThreshold: slowThreshold,
		LogLevel:      logLevel,
	}
}

func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= gormlogger.Info {
		global.Logger.Info("gorm info",
			zap.String("request_id", xcontext.GetRequestID(ctx)),
			zap.String("msg", msg),
			zap.Any("data", data),
		)
	}
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= gormlogger.Warn {
		global.Logger.Warn("gorm warn",
			zap.String("request_id", xcontext.GetRequestID(ctx)),
			zap.String("msg", msg),
			zap.Any("data", data),
		)
	}
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= gormlogger.Error {
		global.Logger.Error("gorm error",
			zap.String("request_id", xcontext.GetRequestID(ctx)),
			zap.String("msg", msg),
			zap.Any("data", data),
		)
	}
}

func (l *GormLogger) Trace(
	ctx context.Context,
	begin time.Time,
	fc func() (sql string, rowsAffected int64),
	err error,
) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	requestID := xcontext.GetRequestID(ctx)

	switch {
	case err != nil && l.LogLevel >= gormlogger.Error && !errors.Is(err, gorm.ErrRecordNotFound):
		global.Logger.Error("gorm sql error",
			zap.String("request_id", requestID),
			zap.Error(err),
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)

	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= gormlogger.Warn:
		global.Logger.Warn("gorm slow sql",
			zap.String("request_id", requestID),
			zap.Duration("elapsed", elapsed),
			zap.Duration("slow_threshold", l.SlowThreshold),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)

	case l.LogLevel >= gormlogger.Info:
		global.Logger.Info("gorm sql",
			zap.String("request_id", requestID),
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
	}
}