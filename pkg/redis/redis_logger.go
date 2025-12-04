package redis

import (
	"context"

	"go.uber.org/zap"
)

type RedisLogger struct {
	zapLogger *zap.Logger
}

func (l *RedisLogger) Printf(ctx context.Context, format string, v ...interface{}) {
	l.zapLogger.Sugar().Infof(format, v...)
}

func NewRedisLogger(zaplogger *zap.Logger) RedisLogger {
	return RedisLogger{
		zapLogger: zaplogger,
	}
}
