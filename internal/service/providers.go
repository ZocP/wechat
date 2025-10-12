package service

import (
	"go.uber.org/fx"
)

// Provide 提供服务依赖注入
func Provide() fx.Option {
	return fx.Options(
		fx.Provide(
			// 认证服务
			fx.Annotate(
				NewAuthService,
				fx.ParamTags(`name:"userRepo"`),
			),
			// 报名服务
			fx.Annotate(
				NewRegistrationService,
				fx.ParamTags(`name:"registrationRepo"`),
			),
			// 订单服务
			fx.Annotate(
				NewOrderService,
				fx.ParamTags(`name:"orderRepo"`, `name:"registrationRepo"`),
			),
			// 支付服务
			fx.Annotate(
				NewPaymentService,
				fx.ParamTags(`name:"paymentRepo"`, `name:"orderRepo"`),
			),
			// 消息板服务
			fx.Annotate(
				NewNoticeService,
				fx.ParamTags(`name:"noticeRepo"`),
			),
		),
	)
}
