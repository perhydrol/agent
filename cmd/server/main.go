package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/perhydrol/insurance-agent-backend/docs"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           InsurAI Platform API
// @version         1.0
// @description     这是保险电商与 AI 顾问系统的后端 API 文档。
// @host            localhost:8080
// @BasePath        /api/v1
func main() {
	r := gin.Default()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	v1 := r.Group("/api/v1")
	{
		v1.GET("/ping", Ping)
	}
	r.Run(":8080")
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
