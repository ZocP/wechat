package handler

import (
	"pickup/internal/config"
	"pickup/internal/service"
	"pickup/pkg/server"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// RouterConfig 路由配置
type RouterConfig struct {
	AuthService         service.AuthService
	RegistrationService service.RegistrationService
	OrderService        service.OrderService
	PaymentService      service.PaymentService
	NoticeService       service.NoticeService
	SchemaService       service.SchemaService
	JWTConfig           *config.JWTConfig
	Logger              *zap.Logger
}

// NewRouterConfig 创建路由配置
func NewRouterConfig(
	authService service.AuthService,
	registrationService service.RegistrationService,
	orderService service.OrderService,
	paymentService service.PaymentService,
	noticeService service.NoticeService,
	schemaService service.SchemaService,
	jwtConfig *config.JWTConfig,
	logger *zap.Logger,
) *RouterConfig {
	return &RouterConfig{
		AuthService:         authService,
		RegistrationService: registrationService,
		OrderService:        orderService,
		PaymentService:      paymentService,
		NoticeService:       noticeService,
		SchemaService:       schemaService,
		JWTConfig:           jwtConfig,
		Logger:              logger,
	}
}

// SetupRoutes 设置路由
func (rc *RouterConfig) SetupRoutes(r *gin.Engine) {
	// API v1 路由组
	api := r.Group("/api/v1")

	// 创建处理器
	authHandler := NewAuthHandler(rc.AuthService, rc.Logger)
	registrationHandler := NewRegistrationHandler(rc.RegistrationService, rc.Logger)
	orderHandler := NewOrderHandler(rc.OrderService, rc.Logger)
	paymentHandler := NewPaymentHandler(rc.PaymentService, rc.Logger)
	noticeHandler := NewNoticeHandler(rc.NoticeService, rc.Logger)
	adminHandler := NewAdminHandler(rc.SchemaService, rc.Logger)

	// 创建JWT工具（这里需要实际的JWT工具实例）
	// jwtUtil := utils.NewJWTUtil(rc.JWTConfig.Secret, rc.JWTConfig.ExpireTime, rc.JWTConfig.Issuer)

	// 注册路由
	authHandler.RegisterRoutes(api)
	registrationHandler.RegisterRoutes(api)
	orderHandler.RegisterRoutes(api)
	paymentHandler.RegisterRoutes(api)
	noticeHandler.RegisterRoutes(api)
	adminHandler.RegisterRoutes(api)

	// 健康检查
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}

// Provide 提供依赖注入
func Provide() fx.Option {
	return fx.Options(
		fx.Provide(NewRouterConfig),
		// Provide an initializer compatible with server.InitRouter
		fx.Provide(func(rc *RouterConfig) server.InitRouter {
			return func(r *gin.Engine) {
				rc.SetupRoutes(r)
			}
		}),
	)
}
