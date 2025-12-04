package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/perhydrol/insurance-agent-backend/cmd/server"
	_ "github.com/perhydrol/insurance-agent-backend/docs"
	"github.com/perhydrol/insurance-agent-backend/pkg/config"
	"github.com/perhydrol/insurance-agent-backend/pkg/database"
	"github.com/perhydrol/insurance-agent-backend/pkg/domain"
	"github.com/perhydrol/insurance-agent-backend/pkg/logger"
	"github.com/perhydrol/insurance-agent-backend/pkg/redis"
	"go.uber.org/zap"
)

// @title           InsurAI Platform API
// @version         1.0
// @description     这是保险电商与 AI 顾问系统的后端 API 文档。
// @host            localhost:8080
// @BasePath        /api/v1
func main() {
	config.InitConfig()
	logger.InitLogger(config.AppConfig.Log)
	domain.InitIDGenerator(config.AppConfig.Snowflake.NodeID)
	database.InitDB(config.AppConfig.Database)
	redis.InitRedisDB(config.AppConfig.Redis)

	r := server.SetupRouter(config.AppConfig.Server)

	port := config.AppConfig.Server.Port
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Log.Info("Server is running", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("listen error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // 阻塞直到收到信号
	logger.Log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Log.Info("Server exiting")
}
