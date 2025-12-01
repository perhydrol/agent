package database

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type GormLogger struct {
	ZapLogger     *zap.Logger
	SlowThreshold time.Duration
}

func NewGormLogger(zapLogger *zap.Logger, slowThreshold int) GormLogger {
	return GormLogger{
		ZapLogger:     zapLogger,
		SlowThreshold: time.Duration(slowThreshold) * time.Millisecond,
	}
}

func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.ZapLogger.Sugar().Infof(msg, data...)
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.ZapLogger.Sugar().Warnf(msg, data...)
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.ZapLogger.Sugar().Errorf(msg, data...)
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowAffected int64), err error) {
	elapsed := time.Since(begin)

	sql, row := fc()
	fields := []zap.Field{
		zap.String("sql", sql),
		zap.Int64("rows", row),
		zap.Duration("elapsed", elapsed),
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		l.ZapLogger.Error("GORM Error", append(fields, zap.Error(err))...)
		return
	}

	if elapsed > l.SlowThreshold {
		l.ZapLogger.Warn("GORM Slow Query", fields...)
		return
	}

	// zap.logger.Check 是一个高性能的实现，能够在不需要输出时快速返回nil，避免内存分配
	if ce := l.ZapLogger.Check(zap.DebugLevel, "GORM Debug"); ce != nil {
		ce.Write(fields...)
	}
}
