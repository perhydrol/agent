package logger

import (
	"os"
	"sync"

	"github.com/natefinch/lumberjack"
	"github.com/perhydrol/insurance-agent-backend/pkg/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

var once sync.Once

func InitLogger(cfg config.LogConfig) {
	once.Do(func() {
		writeSyncer := getLogWriter(cfg)
		encoder := getEncoder()

		var level zapcore.Level
		if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
			level = zapcore.InfoLevel // 默认 Info
		}
		core := zapcore.NewTee(
			zapcore.NewCore(encoder, writeSyncer, level),                            // 写入文件
			zapcore.NewCore(getConsoleEncoder(), zapcore.AddSync(os.Stdout), level), // 写入控制台
		)
		Log = zap.New(core, zap.AddCaller())
		zap.ReplaceGlobals(Log) // 替换全局的log
	})
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder   // 时间格式: 2023-10-01T12:00:00.000Z
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder // 级别大写: INFO, ERROR
	return zapcore.NewJSONEncoder(encoderConfig)
}

func getConsoleEncoder() zapcore.Encoder {
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // 带颜色
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter(cfg config.LogConfig) zapcore.WriteSyncer {
	lumberjackLogger := &lumberjack.Logger{
		Filename:   cfg.Filename,
		MaxSize:    cfg.MaxSize,
		MaxAge:     cfg.MaxAge,
		MaxBackups: cfg.MaxBackups,
		Compress:   cfg.Compress,
	}
	return zapcore.AddSync(lumberjackLogger)
}
