package server

import (
	"fmt"
	"net/http"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type Server struct {
	router *gin.Engine
	server *http.Server
	logger *zap.Logger
}

func (s *Server) Start() error {
	go func() {
		s.logger.Info(fmt.Sprintf("http server start at %s", s.server.Addr))
		if err := s.server.ListenAndServe(); err != nil {
			s.logger.Fatal("http server svcerror", zap.Error(err))
		}
	}()
	return nil
}

func (s *Server) Stop() error {
	s.logger.Info("stopping http server")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "shutdown http server svcerror")
	}

	return nil
}

type InitRouter func(r *gin.Engine)

func NewRouter(cfg *Config, init InitRouter, logger *zap.Logger) *gin.Engine {
	router := gin.New()
	if cfg.AllowCORS {
		router.Use(Cors())
	}
	if cfg.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}
	router.Use(gin.Recovery())
	router.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	router.Use(ginzap.RecoveryWithZap(logger, true))
	init(router)
	return router
}

func NewServer(router *gin.Engine, logger *zap.Logger, config *Config) *Server {
	return &Server{
		router: router,
		server: &http.Server{
			Addr:    fmt.Sprintf("%s:%d", config.Addr, config.Port),
			Handler: router,
		},
		logger: logger,
	}
}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", "*") // 可将将 * 替换为指定的域名
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}

func Provide() fx.Option {
	return fx.Options(fx.Provide(NewServer, NewConfig, NewRouter), fx.Invoke(lc))
}

func lc(lifecycle fx.Lifecycle, s *Server) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return s.Start()
		},
		OnStop: func(ctx context.Context) error {
			return s.Stop()
		},
	})
}
