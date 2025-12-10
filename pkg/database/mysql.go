package database

import (
	"sync"
	"time"

	"github.com/perhydrol/insurance-agent-backend/pkg/config"
	"github.com/perhydrol/insurance-agent-backend/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

var once sync.Once

func InitDB(cfg config.DatabaseConfig) {
	once.Do(func() {
		gormLogger := NewGormLogger(logger.Log, cfg.SlowThreshold)

		var err error
		DB, err = gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
			Logger: &gormLogger,
		})

		if err != nil {
			logger.Log.Fatal("Database connection failed", zap.Error(err))
		}

		sqlDB, err := DB.DB()
		if err != nil {
			logger.Log.Fatal("Failed to retrieve sql.DB", zap.Error(err))
		}

		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
		sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Hour)

		logger.Log.Info("Database connection successful")
	})
}
