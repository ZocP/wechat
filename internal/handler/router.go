package handler

import (
	"pickup/internal/config"
	"pickup/internal/scheduler/controllers"
	"pickup/internal/scheduler/routes"
	"pickup/internal/utils"
	"pickup/pkg/server"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

// RouterConfig 路由配置
type RouterConfig struct {
	AuthController    *controllers.AuthController
	StudentController *controllers.StudentController
	AdminController   *controllers.AdminController
	JWTConfig         *config.JWTConfig
}

// NewRouterConfig 创建路由配置
func NewRouterConfig(
	authController *controllers.AuthController,
	studentController *controllers.StudentController,
	adminController *controllers.AdminController,
	jwtConfig *config.JWTConfig,
) *RouterConfig {
	return &RouterConfig{
		AuthController:    authController,
		StudentController: studentController,
		AdminController:   adminController,
		JWTConfig:         jwtConfig,
	}
}

// SetupRoutes 设置路由
func (rc *RouterConfig) SetupRoutes(r *gin.Engine) {
	// 创建JWT工具
	jwtUtil := utils.NewJWTUtil(rc.JWTConfig.Secret, rc.JWTConfig.ExpireTime, rc.JWTConfig.Issuer)
	routes.RegisterRoutes(r, rc.AuthController, rc.StudentController, rc.AdminController, jwtUtil)
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
