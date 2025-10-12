package zap

import (
	"os"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func NewLogger(config *Config) *zap.Logger {
	var coreArr []zapcore.Core

	//encoder for zap logger
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoderConfig.EncodeCaller = zapcore.FullCallerEncoder
	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	highPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev >= zap.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev < zap.ErrorLevel && lev >= zap.DebugLevel
	})
	var lv zapcore.Level

	//control whether to log to file based on config
	if config.LogToFile {
		infoFileWriteSyncer := zapcore.AddSync(&lumberjack.Logger{
			Filename:   config.LogPath + "info.log",
			MaxSize:    config.MaxSize,
			MaxBackups: 100,
			Compress:   config.Compress,
		})
		infoFileCore := zapcore.NewCore(encoder, infoFileWriteSyncer, lowPriority)
		errorFileWriteSyncer := zapcore.AddSync(&lumberjack.Logger{
			Filename:   config.LogPath + "svcerror.log",
			MaxSize:    config.MaxSize,
			MaxBackups: 100,
			Compress:   config.Compress,
		})
		errorFileCore := zapcore.NewCore(encoder, errorFileWriteSyncer, highPriority)

		coreArr = append(coreArr, infoFileCore)
		coreArr = append(coreArr, errorFileCore)
	}

	//control which level to log to stdout based on config
	if config.Developing {
		lv = zap.DebugLevel
	} else {
		lv = zap.InfoLevel
	}

	coreArr = append(coreArr, zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), lv))
	log := zap.New(zapcore.NewTee(coreArr...), zap.AddCaller())
	zap.ReplaceGlobals(log)
	return log
}

func FxLogger(log *zap.Logger) fxevent.Logger {
	lg := &fxevent.ZapLogger{
		Logger: log,
	}
	lg.UseLogLevel(zap.DebugLevel)
	return lg
}

func Provide() fx.Option {
	return fx.Options(fx.Provide(NewConfig, NewLogger), fx.WithLogger(FxLogger))
}
