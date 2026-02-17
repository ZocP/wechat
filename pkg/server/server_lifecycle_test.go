package server

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type fakeLifecycle struct {
	hooks []fx.Hook
}

func (f *fakeLifecycle) Append(h fx.Hook) {
	f.hooks = append(f.hooks, h)
}

func safeFatalLogger() *zap.Logger {
	core := zapcore.NewCore(zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()), zapcore.AddSync(io.Discard), zap.DebugLevel)
	return zap.New(core, zap.WithFatalHook(zapcore.WriteThenNoop))
}

func TestProvide_NotNil(t *testing.T) {
	assert.NotNil(t, Provide())
}

func TestLC_AppendsHook(t *testing.T) {
	l := &fakeLifecycle{}
	s := &Server{server: &http.Server{}, logger: zap.NewNop()}
	lc(l, s)
	require.Len(t, l.hooks, 1)
}

func TestServerStopFromServing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	httpServer := &http.Server{Handler: gin.New()}
	s := &Server{server: httpServer, logger: safeFatalLogger()}

	go func() {
		_ = httpServer.Serve(listener)
	}()

	time.Sleep(30 * time.Millisecond)
	require.NoError(t, s.Stop())
}

func TestServerStopError(t *testing.T) {
	s := &Server{server: &http.Server{}, logger: zap.NewNop()}
	err := s.Stop()
	require.NoError(t, err)
}

func TestLC_HookCallbacks(t *testing.T) {
	l := &fakeLifecycle{}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	httpServer := &http.Server{Handler: gin.New()}
	s := &Server{server: httpServer, logger: safeFatalLogger()}

	go func() {
		_ = httpServer.Serve(listener)
	}()

	lc(l, s)
	require.Len(t, l.hooks, 1)
	time.Sleep(30 * time.Millisecond)
	require.NoError(t, l.hooks[0].OnStop(context.Background()))
}

func TestServerStartOnly(t *testing.T) {
	cfg := &Config{Addr: "127.0.0.1", Port: 0}
	s := NewServer(gin.New(), safeFatalLogger(), cfg)
	require.NoError(t, s.Start())
	time.Sleep(20 * time.Millisecond)
}

func TestLC_OnStart(t *testing.T) {
	l := &fakeLifecycle{}
	cfg := &Config{Addr: "127.0.0.1", Port: 0}
	s := NewServer(gin.New(), safeFatalLogger(), cfg)
	lc(l, s)
	require.Len(t, l.hooks, 1)
	require.NoError(t, l.hooks[0].OnStart(context.Background()))
	time.Sleep(20 * time.Millisecond)
}
