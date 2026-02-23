package logging

import (
	"context"

	"go.uber.org/zap"
)

type loggerKey struct{}
type traceIDKey struct{}

// WithLogger returns a new context with the logger attached.
func WithLogger(ctx context.Context, l *Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

// FromContext retrieves the logger from the context.
// Returns nil if no logger is found.
func FromContext(ctx context.Context) *Logger {
	l, ok := ctx.Value(loggerKey{}).(*Logger)
	if !ok {
		return nil
	}
	return l
}

// WithTraceID returns a new context with the trace ID attached.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey{}, traceID)
}

// TraceIDFromContext retrieves the trace ID from the context.
// Returns empty string if no trace ID is found.
func TraceIDFromContext(ctx context.Context) string {
	traceID, ok := ctx.Value(traceIDKey{}).(string)
	if !ok {
		return ""
	}
	return traceID
}

// LoggerFromContext retrieves the logger from context and enriches it with the trace ID if present.
// Returns nil if no logger is found in the context.
func LoggerFromContext(ctx context.Context) *Logger {
	l := FromContext(ctx)
	if l == nil {
		return nil
	}

	traceID := TraceIDFromContext(ctx)
	if traceID != "" {
		return l.With(zap.String("trace_id", traceID))
	}

	return l
}
