package redis

import (
	"context"
	"sync"
	"time"

	"github.com/perhydrol/insurance-agent-backend/pkg/config"
	"github.com/perhydrol/insurance-agent-backend/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var RDB *redis.Client

var once sync.Once

func InitRedisDB(cfg config.RedisConfig) {
	once.Do(func() {
		rLogger := NewLogger(logger.Log)
		redis.SetLogger(&rLogger)
		RDB = redis.NewClient(&redis.Options{
			Addr:     cfg.Addr,
			Password: cfg.Password,
			DB:       cfg.DB,
		})

		zaphook := NewZapLogHook(logger.Log, time.Duration(cfg.SlowThreshold)*time.Millisecond, cfg.LogMaxLen)
		RDB.AddHook(&zaphook)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if _, err := RDB.Ping(ctx).Result(); err != nil {
			logger.Log.Fatal("failed to connect redis", zap.Error(err))
		}
	})
}
