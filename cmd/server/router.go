package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/perhydrol/insurance-agent-backend/pkg/config"
	"github.com/perhydrol/insurance-agent-backend/pkg/logger"
	"github.com/perhydrol/insurance-agent-backend/pkg/middleware"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(cfg config.ServerConfig) *gin.Engine {
	gin.SetMode(cfg.Mode)
	r := gin.New()
	r.Use(middleware.GinTraceID(), logger.GinLogger(), logger.GinRecovery(cfg.Stack))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	v1 := r.Group("/api/v1")
	{
		v1.GET("/ping", Ping)
	}

	return r
}

// PingHandler
// @Summary      健康检查
// @Description  检测服务是否存活
// @Tags         System
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /ping [get]
func Ping(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
