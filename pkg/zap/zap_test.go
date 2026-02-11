package zap

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestNewConfig_WithLoggerSet(t *testing.T) {
	v := viper.New()
	v.Set("logger.developing", true)
	v.Set("logger.logToFile", false)
	v.Set("logger.logPath", "/tmp/logs/")
	v.Set("logger.maxSize", 5)
	v.Set("logger.compress", false)

	cfg := NewConfig(v)
	assert.NotNil(t, cfg)
	assert.True(t, cfg.Developing)
	assert.False(t, cfg.LogToFile)
	assert.Equal(t, "/tmp/logs/", cfg.LogPath)
	assert.Equal(t, 5, cfg.MaxSize)
	assert.False(t, cfg.Compress)
}

func TestNewConfig_WithoutLoggerSet(t *testing.T) {
	v := viper.New()

	cfg := NewConfig(v)
	assert.NotNil(t, cfg)
	// Should return default config
	assert.False(t, cfg.Developing)
	assert.True(t, cfg.LogToFile)
	assert.Equal(t, 2, cfg.MaxSize)
	assert.True(t, cfg.Compress)
}

func TestNewConfig_InvalidLoggerConfig(t *testing.T) {
	v := viper.New()
	// Set logger to an invalid type that can't unmarshal to Config struct
	v.Set("logger", "invalid_string")

	cfg := NewConfig(v)
	assert.NotNil(t, cfg)
	// Should fall back to default config
	assert.False(t, cfg.Developing)
	assert.True(t, cfg.LogToFile)
}

func TestDefaultConfig(t *testing.T) {
	v := viper.New()
	cfg := defaultConfig(v)
	assert.NotNil(t, cfg)
	assert.False(t, cfg.Developing)
	assert.True(t, cfg.LogToFile)
	assert.Equal(t, 2, cfg.MaxSize)
	assert.True(t, cfg.Compress)
	// Verify viper default was set
	assert.True(t, v.IsSet("logger"))
}

func TestNewLogger_NoFile(t *testing.T) {
	cfg := &Config{
		Developing: true,
		LogToFile:  false,
	}

	logger := NewLogger(cfg)
	assert.NotNil(t, logger)
	// Logger should still work with console-only output
	logger.Info("test log message from TestNewLogger_NoFile")
}

func TestNewLogger_WithFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zap_test_*")
	assert.NoError(t, err)
	// Don't defer cleanup - lumberjack keeps file handles open

	cfg := &Config{
		Developing: false,
		LogToFile:  true,
		LogPath:    tmpDir + "/",
		MaxSize:    1,
		Compress:   false,
	}

	logger := NewLogger(cfg)
	assert.NotNil(t, logger)
	logger.Info("test log message from TestNewLogger_WithFile")
}

func TestNewLogger_DevelopingMode(t *testing.T) {
	cfg := &Config{
		Developing: true,
		LogToFile:  false,
	}

	logger := NewLogger(cfg)
	assert.NotNil(t, logger)
	// In developing mode, debug level logs should be enabled
	logger.Debug("this should appear in developing mode")
}

func TestFxLogger(t *testing.T) {
	cfg := &Config{
		Developing: true,
		LogToFile:  false,
	}
	logger := NewLogger(cfg)

	fxLogger := FxLogger(logger)
	assert.NotNil(t, fxLogger)
}

func TestProvide(t *testing.T) {
	opt := Provide()
	assert.NotNil(t, opt)
}
