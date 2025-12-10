package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/perhydrol/insurance-agent-backend/internal/infrastructure/queue"
	"github.com/perhydrol/insurance-agent-backend/internal/infrastructure/repository"
	"github.com/perhydrol/insurance-agent-backend/internal/modules/order"
	"github.com/perhydrol/insurance-agent-backend/internal/modules/product"
	"github.com/perhydrol/insurance-agent-backend/internal/modules/user"
	"github.com/perhydrol/insurance-agent-backend/pkg/config"
	"github.com/perhydrol/insurance-agent-backend/pkg/logger"
	"github.com/perhydrol/insurance-agent-backend/pkg/middleware"
	"github.com/perhydrol/insurance-agent-backend/pkg/redis"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(cfg config.ServerConfig) *gin.Engine {
	gin.SetMode(cfg.Mode)
	r := gin.New()
	r.Use(middleware.GinTraceID(), logger.GinLogger(), logger.GinRecovery(cfg.Stack))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	q := queue.NewRedisQueue(redis.RDB)

	// Services
	userSvc := user.NewUserService(repository.Repository.UserRepo)
	productSvc := product.NewProductService(repository.Repository.ProductRepo)
	// Order service starts a background processor
	orderSvc := order.NewService(
		context.Background(),
		repository.Repository.OrderRepo,
		repository.Repository.ProductRepo,
		q,
	)

	// Handlers
	userH := user.NewUHandler(userSvc)
	productH := product.NewPHandler(productSvc)
	orderH := order.NewOHandler(orderSvc)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/ping", Ping)

		// Auth
		auth := v1.Group("/auth")
		{
			auth.POST("/register", userH.Register)
			auth.POST("/login", userH.Login)
		}

		// Products
		products := v1.Group("/products")
		{
			products.GET("", productH.ListProducts)
			products.GET("/:id", productH.GetProduct)
		}

		// Orders
		orders := v1.Group("/orders")
		orders.Use(middleware.JWTAuth())
		{
			orders.POST("", orderH.CreateOrder)
			orders.POST("/:id/pay", orderH.PayOrder)
		}
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
