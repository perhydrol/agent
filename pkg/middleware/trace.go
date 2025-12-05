package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	traceid "github.com/perhydrol/insurance-agent-backend/pkg/traceID"
)

func GinTraceID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		reqID := uuid.New().String()
		c := ctx.Request.Context()
		cWithValue := context.WithValue(c, traceid.ContextTraceIDKey, reqID)
		ctx.Request = ctx.Request.WithContext(cWithValue)
		ctx.Header("X-Request-ID", reqID)
		ctx.Next()
	}
}
