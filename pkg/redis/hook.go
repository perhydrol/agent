package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/perhydrol/insurance-agent-backend/pkg/middleware"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func cutString(args string, maxLen int) string {
	strLen := len(args)
	if strLen >= maxLen {
		return args[:maxLen] + "...."
	} else {
		return args
	}
}

type ZapLogHook struct {
	zapLogger     *zap.Logger
	slowThreshold time.Duration
	logMaxLen     int
}

func NewZapLogHook(zaplogger *zap.Logger, slowThreshold time.Duration, logMaxLen int) ZapLogHook {
	return ZapLogHook{
		zapLogger:     zaplogger,
		slowThreshold: slowThreshold,
		logMaxLen:     logMaxLen,
	}
}

func (h *ZapLogHook) DialHook(next redis.DialHook) redis.DialHook {
	return next
}

func (h *ZapLogHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmd)
		cost := time.Since(start)

		fields := []zap.Field{
			zap.String(middleware.TraceIDKey, middleware.GetTraceID(ctx)),
			zap.String("cmd", cmd.Name()),
			zap.String("args", cutString(fmt.Sprintf("%v", cmd.Args()), h.logMaxLen)),
			zap.Duration("cost", cost),
		}
		if err != nil && !errors.Is(err, redis.Nil) {
			h.zapLogger.Error(
				"redis cmd failed", append(fields, zap.Error(err))...,
			)
			return err
		}

		if cost > h.slowThreshold {
			h.zapLogger.Warn("redis slow query", fields...)
			return err
		}

		if ce := h.zapLogger.Check(zap.DebugLevel, "redis Debug"); ce != nil {
			ce.Write(fields...)
		}
		return err
	}
}

func (h *ZapLogHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmds)
		cost := time.Since(start)

		// 为了日志简洁，这里只记录命令个数和摘要，详细内容仅在 Debug/Error 时打印
		fields := []zap.Field{
			zap.String(middleware.TraceIDKey, middleware.GetTraceID(ctx)),
			zap.String("type", "pipeline"),
			zap.Int("cmd_count", len(cmds)),
			zap.Duration("cost", cost),
		}

		if err != nil && !errors.Is(err, redis.Nil) {
			// Pipeline 中只要有一个失败，通常 err 就会非空
			h.zapLogger.Error("redis pipeline failed", append(fields, zap.Error(err))...)
			return err
		}

		if cost > h.slowThreshold {
			// 这里简单打印命令名，防止 Args 太大爆日志
			cmdNames := make([]string, 0, len(cmds))
			for _, cmd := range cmds {
				cmdNames = append(cmdNames, cmd.Name())
			}
			h.zapLogger.Warn("redis pipeline slow",
				append(fields, zap.Strings("cmds", cmdNames))...,
			)
			return err
		}

		if ce := h.zapLogger.Check(zap.DebugLevel, "redis pipeline debug"); ce != nil {
			ce.Write(fields...)
		}

		return err
	}
}
