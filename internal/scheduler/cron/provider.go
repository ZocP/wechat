package cron

import (
	"context"

	rcron "github.com/robfig/cron/v3"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func RegisterCron(lc fx.Lifecycle, syncSvc *SyncFlightService, logger *zap.Logger) {
	c := rcron.New()

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			_, err := c.AddFunc("@every 30m", func() {
				if syncErr := syncSvc.SyncFlightData(context.Background()); syncErr != nil {
					logger.Error("sync flight data failed", zap.Error(syncErr))
				}
			})
			if err != nil {
				return err
			}
			c.Start()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			stopCtx := c.Stop()
			select {
			case <-stopCtx.Done():
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
	})
}
