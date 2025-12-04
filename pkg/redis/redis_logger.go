package redis

import (
	"context"

	"go.uber.org/zap"
)

// Logger 日志适配器，用于将 Redis 客户端日志输出到 Zap
type Logger struct {
	zapLogger *zap.Logger
}

// Printf 按格式输出日志
func (l *Logger) Printf(ctx context.Context, format string, v ...interface{}) {
	l.zapLogger.Sugar().Infof(format, v...)
}

// NewLogger 创建 Logger
func NewLogger(zaplogger *zap.Logger) Logger {
	return Logger{
		zapLogger: zaplogger,
	}
}
