package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger for structured JSON logging.
type Logger struct {
	zap *zap.Logger
}

// New creates a new Logger configured for production JSON output at the specified level.
// Valid levels: debug, info, warn, error, dpanic, panic, fatal
func New(level string) (*Logger, error) {
	lvl, err := zapcore.ParseLevel(level)
	if err != nil {
		return nil, err
	}

	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(lvl),
		Development:      false,
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	zapLogger, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}

	return &Logger{zap: zapLogger}, nil
}

// With returns a new Logger with the given fields added to all log entries.
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{zap: l.zap.With(fields...)}
}

// Debug logs a message at debug level.
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.zap.Debug(msg, fields...)
}

// Info logs a message at info level.
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.zap.Info(msg, fields...)
}

// Warn logs a message at warn level.
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.zap.Warn(msg, fields...)
}

// Error logs a message at error level.
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.zap.Error(msg, fields...)
}

// DPanic logs a message at dpanic level.
// In development, the logger panics after logging. In production, it logs at error level.
func (l *Logger) DPanic(msg string, fields ...zap.Field) {
	l.zap.DPanic(msg, fields...)
}

// Panic logs a message at panic level and then panics.
func (l *Logger) Panic(msg string, fields ...zap.Field) {
	l.zap.Panic(msg, fields...)
}

// Fatal logs a message at fatal level and then calls os.Exit(1).
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.zap.Fatal(msg, fields...)
}

// Sync flushes any buffered log entries.
func (l *Logger) Sync() error {
	return l.zap.Sync()
}

// Zap returns the underlying zap.Logger for use with middleware that requires it.
func (l *Logger) Zap() *zap.Logger {
	return l.zap
}
