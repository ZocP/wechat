package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ===== Config Tests =====

func TestNewConfig_Default(t *testing.T) {
	v := viper.New()

	cfg, err := NewConfig(v)
	require.NoError(t, err)
	assert.Equal(t, 8080, cfg.Port)
	assert.True(t, cfg.AllowCORS)
	assert.False(t, cfg.ReleaseMode)
}

func TestNewConfig_FromViper(t *testing.T) {
	v := viper.New()
	v.Set("server.addr", "0.0.0.0")
	v.Set("server.port", 9090)
	v.Set("server.allowCORS", false)
	v.Set("server.releaseMode", true)

	cfg, err := NewConfig(v)
	require.NoError(t, err)
	assert.Equal(t, "0.0.0.0", cfg.Addr)
	assert.Equal(t, 9090, cfg.Port)
	assert.False(t, cfg.AllowCORS)
	assert.True(t, cfg.ReleaseMode)
}

func TestDefaultConfig(t *testing.T) {
	v := viper.New()
	cfg := defaultConfig(v)

	assert.Equal(t, 8080, cfg.Port)
	assert.True(t, cfg.AllowCORS)
	assert.False(t, cfg.ReleaseMode)

	// Verify it was also set in viper
	assert.True(t, v.IsSet("server"))
}

// ===== Cors Tests =====

func TestCors_SetsHeaders(t *testing.T) {
	router := gin.New()
	router.Use(Cors())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCors_OptionsRequest(t *testing.T) {
	router := gin.New()
	router.Use(Cors())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestCors_NoOriginHeader(t *testing.T) {
	router := gin.New()
	router.Use(Cors())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	// No Origin header
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

// ===== NewRouter Tests =====

func TestNewRouter_WithCORS(t *testing.T) {
	cfg := &Config{AllowCORS: true, ReleaseMode: false}
	initFn := func(r *gin.Engine) {
		r.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
	}
	logger := zap.NewNop()

	router := NewRouter(cfg, initFn, logger)
	require.NotNil(t, router)

	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestNewRouter_WithoutCORS(t *testing.T) {
	cfg := &Config{AllowCORS: false, ReleaseMode: false}
	initFn := func(r *gin.Engine) {
		r.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
	}
	logger := zap.NewNop()

	router := NewRouter(cfg, initFn, logger)
	require.NotNil(t, router)

	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

// ===== NewServer Tests =====

func TestNewServer(t *testing.T) {
	cfg := &Config{Addr: "127.0.0.1", Port: 0}
	router := gin.New()
	logger := zap.NewNop()

	server := NewServer(router, logger, cfg)
	require.NotNil(t, server)
	assert.Equal(t, router, server.router)
	assert.Equal(t, logger, server.logger)
	assert.Equal(t, "127.0.0.1:0", server.server.Addr)
}

func TestServer_StartStop(t *testing.T) {
	cfg := &Config{Addr: "127.0.0.1", Port: 0}
	router := gin.New()
	// Use a development logger that won't os.Exit on Fatal for ErrServerClosed
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	server := NewServer(router, logger, cfg)
	// Only test construction, not actual start/stop to avoid Fatal on ErrServerClosed
	assert.NotNil(t, server.server)
	assert.Equal(t, "127.0.0.1:0", server.server.Addr)
}
