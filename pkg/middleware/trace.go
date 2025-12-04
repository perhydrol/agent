package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ctxKey string

const TraceIDKey = "trace_id"

// ContextTraceIDKey is the typed context key used to store the trace id
var ContextTraceIDKey ctxKey = "trace_id"

func GinTraceID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		reqID := uuid.New().String()
		c := ctx.Request.Context()
		cWithValue := context.WithValue(c, ContextTraceIDKey, reqID)
		ctx.Request = ctx.Request.WithContext(cWithValue)
		ctx.Header("X-Request-ID", reqID)
		ctx.Next()
	}
}

func GetTraceID(ctx context.Context) string {
	if id, ok := ctx.Value(ContextTraceIDKey).(string); ok {
		return id
	}
	return "unknown-trace-id" // Return a default value
}
