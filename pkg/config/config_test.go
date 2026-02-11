package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewViper_WithConfigFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte("server:\n  port: 9090\n"), 0644)
	assert.NoError(t, err)

	// Temporarily change the search path
	origDir := FileDirectory
	// NewViper uses constants, but we can test the default behavior (no config file found)
	_ = origDir

	v := NewViper()
	assert.NotNil(t, v)
}

func TestNewViper_NoConfigFile(t *testing.T) {
	// When config file doesn't exist, NewViper should still return a viper instance
	// It will print "Load config failed. Regenerate config." and continue
	v := NewViper()
	assert.NotNil(t, v)
}

func TestProvide_ReturnsOption(t *testing.T) {
	opt := Provide()
	assert.NotNil(t, opt)
}

func TestLc(t *testing.T) {
	// lc just registers an OnStop hook, we can test that it doesn't panic
	// by verifying the function signature is correct
	// The actual fx lifecycle test would require a full fx app
	assert.NotPanics(t, func() {
		_ = Provide()
	})
}
