package repository

import (
	"go.uber.org/fx"
)

// Provide 提供仓储依赖注入
func Provide() fx.Option {
	return fx.Options(
		fx.Provide(
			// 用户仓储
			fx.Annotate(
				NewUserRepository,
				fx.ResultTags(`name:"userRepo"`),
			),
			// 报名仓储
			fx.Annotate(
				NewRegistrationRepository,
				fx.ResultTags(`name:"registrationRepo"`),
			),
			// 订单仓储
			fx.Annotate(
				NewOrderRepository,
				fx.ResultTags(`name:"orderRepo"`),
			),
			// 支付仓储
			fx.Annotate(
				NewPaymentRepository,
				fx.ResultTags(`name:"paymentRepo"`),
			),
			// 分配仓储
			fx.Annotate(
				NewAssignmentRepository,
				fx.ResultTags(`name:"assignmentRepo"`),
			),
			// 消息仓储
			fx.Annotate(
				NewNoticeRepository,
				fx.ResultTags(`name:"noticeRepo"`),
			),
			// 数据库结构仓储
			fx.Annotate(
				NewSchemaRepository,
				fx.ResultTags(`name:"schemaRepo"`),
			),
		),
	)
}
