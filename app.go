package main

import (
	internalcfg "pickup/internal/config"
	"pickup/internal/handler"
	"pickup/internal/repository"
	"pickup/internal/scheduler"
	"pickup/internal/service"
	pkgcfg "pickup/pkg/config"
	"pickup/pkg/server"
	"pickup/pkg/zap"

	"go.uber.org/fx"
)

func main() {
	fx.New(fx.Options(
		// 基础组件
		server.Provide(),
		zap.Provide(),
		pkgcfg.Provide(),

		// 数据库配置
		internalcfg.Provide(),

		// 仓储层
		repository.Provide(),

		// 服务层
		service.Provide(),

		// 处理器层
		handler.Provide(),

		// 新调度域
		scheduler.Provide(),
	)).Run()
}
