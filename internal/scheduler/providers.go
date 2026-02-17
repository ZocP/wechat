package scheduler

import (
	"pickup/internal/scheduler/controllers"
	"pickup/internal/scheduler/cron"
	"pickup/internal/scheduler/service"

	"go.uber.org/fx"
)

// Provide 注册调度域依赖。
func Provide() fx.Option {
	return fx.Options(
		fx.Provide(
			service.NewShiftAssignmentService,
			service.NewAuthService,
			service.NewStudentService,
			service.NewAdminService,
			controllers.NewAuthController,
			controllers.NewStudentController,
			controllers.NewAdminController,
			cron.NewSyncFlightService,
		),
		fx.Invoke(cron.RegisterCron),
	)
}
