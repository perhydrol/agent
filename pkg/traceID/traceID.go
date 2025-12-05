package traceid

import "context"

type ctxKey string

const TraceIDKey = "trace_id"

// ContextTraceIDKey is the typed context key used to store the trace id
var ContextTraceIDKey ctxKey = "trace_id"

func GetTraceID(ctx context.Context) string {
	if id, ok := ctx.Value(ContextTraceIDKey).(string); ok {
		return id
	}
	return "unknown-trace-id" // Return a default value
}
