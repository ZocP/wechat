package config

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"os"
)

const (
	ConfigName    = "config"
	FileDirectory = "./files/"
	ConfigType    = "yaml"
)

func NewViper() *viper.Viper {
	v := viper.New()
	v.SetConfigName(ConfigName)
	v.SetConfigType(ConfigType)
	v.AddConfigPath(FileDirectory)
	if err := v.ReadInConfig(); err != nil {
		fmt.Println("Load config failed. Regenerate config.")
		_ = os.MkdirAll(FileDirectory, os.ModeDir)
		v.SetConfigFile(FileDirectory + ConfigName + "." + ConfigType)
	}
	return v
}

func Provide() fx.Option {
	return fx.Options(fx.Provide(NewViper), fx.Invoke(lc))
}

// Viper does not support struct tags, so we have to use the same name as the yaml file. However, camelCase works.
// lc allows a default config file to be generated if it does not exist when first exit the program. the program has to end properly to generate the config file.
func lc(lifecycle fx.Lifecycle, v *viper.Viper) {
	lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return v.WriteConfig()
		},
	})
}
