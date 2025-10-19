package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLoggerConfig(t *testing.T) {
	cfg := NewLoggerConfig()
	assert.Equal(t, "timestamp", cfg.Config.EncoderConfig.TimeKey)
}

func TestBuildLogger(t *testing.T) {
	t.Run("Valid logger config", func(t *testing.T) {
		loggerCfg := NewLoggerConfig()
		logger, err := loggerCfg.BuildLogger()
		defer logger.Sync()

		assert.NoError(t, err)
		assert.NotNil(t, logger)
	})

	t.Run("Logger build failure", func(t *testing.T) {
		loggerCfg := NewLoggerConfig()
		loggerCfg.Config.OutputPaths = []string{"/invalid/path/to/logs"}
		logger, err := loggerCfg.BuildLogger()

		assert.Error(t, err)
		assert.Nil(t, logger)
	})
}
