package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// newTestLogger creates a logger that writes to a buffer for testing.
func newTestLogger(buf *bytes.Buffer, level zapcore.Level) *Logger {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = ""

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.AddSync(buf),
		level,
	)

	return &Logger{zap: zap.New(core)}
}

func TestNew_ValidLevels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error", "dpanic", "panic", "fatal"}

	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			logger, err := New(level)
			if err != nil {
				t.Fatalf("New(%q) returned error: %v", level, err)
			}
			if logger == nil {
				t.Fatalf("New(%q) returned nil logger", level)
			}
			_ = logger.Sync()
		})
	}
}

func TestNew_InvalidLevel(t *testing.T) {
	_, err := New("invalid")
	if err == nil {
		t.Fatal("New(\"invalid\") should return error")
	}
}

func TestLogger_Info(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf, zapcore.InfoLevel)

	logger.Info("test message", zap.String("key", "value"))

	var logEntry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}

	if logEntry["msg"] != "test message" {
		t.Errorf("msg = %q, want %q", logEntry["msg"], "test message")
	}
	if logEntry["level"] != "info" {
		t.Errorf("level = %q, want %q", logEntry["level"], "info")
	}
	if logEntry["key"] != "value" {
		t.Errorf("key = %q, want %q", logEntry["key"], "value")
	}
}

func TestLogger_Debug(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf, zapcore.DebugLevel)

	logger.Debug("debug message")

	var logEntry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}

	if logEntry["level"] != "debug" {
		t.Errorf("level = %q, want %q", logEntry["level"], "debug")
	}
}

func TestLogger_Warn(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf, zapcore.WarnLevel)

	logger.Warn("warning message")

	var logEntry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}

	if logEntry["level"] != "warn" {
		t.Errorf("level = %q, want %q", logEntry["level"], "warn")
	}
}

func TestLogger_Error(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf, zapcore.ErrorLevel)

	logger.Error("error message", zap.Error(context.DeadlineExceeded))

	var logEntry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}

	if logEntry["level"] != "error" {
		t.Errorf("level = %q, want %q", logEntry["level"], "error")
	}
	if logEntry["error"] != "context deadline exceeded" {
		t.Errorf("error = %q, want %q", logEntry["error"], "context deadline exceeded")
	}
}

func TestLogger_With(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf, zapcore.InfoLevel)

	childLogger := logger.With(zap.String("service", "test-service"))
	childLogger.Info("test message")

	var logEntry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}

	if logEntry["service"] != "test-service" {
		t.Errorf("service = %q, want %q", logEntry["service"], "test-service")
	}
}

func TestWithLogger_FromContext(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf, zapcore.InfoLevel)

	ctx := WithLogger(context.Background(), logger)

	retrieved := FromContext(ctx)
	if retrieved == nil {
		t.Fatal("FromContext returned nil")
	}

	retrieved.Info("from context")

	if buf.Len() == 0 {
		t.Error("expected log output from context logger")
	}
}

func TestFromContext_NoLogger(t *testing.T) {
	ctx := context.Background()

	retrieved := FromContext(ctx)
	if retrieved != nil {
		t.Error("FromContext should return nil when no logger in context")
	}
}

func TestWithTraceID_TraceIDFromContext(t *testing.T) {
	traceID := "abc-123-xyz"
	ctx := WithTraceID(context.Background(), traceID)

	retrieved := TraceIDFromContext(ctx)
	if retrieved != traceID {
		t.Errorf("TraceIDFromContext = %q, want %q", retrieved, traceID)
	}
}

func TestTraceIDFromContext_NoTraceID(t *testing.T) {
	ctx := context.Background()

	retrieved := TraceIDFromContext(ctx)
	if retrieved != "" {
		t.Errorf("TraceIDFromContext should return empty string, got %q", retrieved)
	}
}

func TestLoggerFromContext_WithTraceID(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf, zapcore.InfoLevel)

	ctx := WithLogger(context.Background(), logger)
	ctx = WithTraceID(ctx, "trace-123")

	enrichedLogger := LoggerFromContext(ctx)
	if enrichedLogger == nil {
		t.Fatal("LoggerFromContext returned nil")
	}

	enrichedLogger.Info("traced message")

	var logEntry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}

	if logEntry["trace_id"] != "trace-123" {
		t.Errorf("trace_id = %q, want %q", logEntry["trace_id"], "trace-123")
	}
}

func TestLoggerFromContext_WithoutTraceID(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf, zapcore.InfoLevel)

	ctx := WithLogger(context.Background(), logger)

	retrieved := LoggerFromContext(ctx)
	if retrieved == nil {
		t.Fatal("LoggerFromContext returned nil")
	}

	retrieved.Info("no trace message")

	var logEntry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}

	if _, exists := logEntry["trace_id"]; exists {
		t.Error("trace_id should not be present when not set in context")
	}
}

func TestLoggerFromContext_NoLogger(t *testing.T) {
	ctx := context.Background()

	retrieved := LoggerFromContext(ctx)
	if retrieved != nil {
		t.Error("LoggerFromContext should return nil when no logger in context")
	}
}
