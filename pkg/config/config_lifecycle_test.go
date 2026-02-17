package config

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

type fakeLifecycle struct {
	hooks []fx.Hook
}

func (f *fakeLifecycle) Append(h fx.Hook) {
	f.hooks = append(f.hooks, h)
}

func TestProvide_NotNil(t *testing.T) {
	assert.NotNil(t, Provide())
}

func TestLifecycle_OnStopWritesConfig(t *testing.T) {
	v := viper.New()
	dir := t.TempDir()
	v.SetConfigFile(filepath.Join(dir, "config.yaml"))
	v.Set("server.port", 8080)

	lcMock := &fakeLifecycle{}
	lc(lcMock, v)
	require.Len(t, lcMock.hooks, 1)
	require.NotNil(t, lcMock.hooks[0].OnStop)
	require.NoError(t, lcMock.hooks[0].OnStop(context.Background()))
}
