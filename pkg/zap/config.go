package zap

import (
	"pickup/pkg/config"

	"github.com/spf13/viper"
)

type Config struct {
	Developing bool `yaml:"developing"`

	LogToFile bool `yaml:"logToFile"`

	//Below only works when LogToFile is true
	LogPath  string `yaml:"logPath"`
	MaxSize  int    `yaml:"maxSize"`
	Compress bool   `yaml:"compress"`
}

const (
	LogPath = "logs/"
)

func NewConfig(v *viper.Viper) *Config {
	cfg := new(Config)
	if !v.IsSet("logger") {
		return defaultConfig(v)
	}
	if err := v.UnmarshalKey("logger", cfg); err != nil {
		return defaultConfig(v)
	}
	return cfg
}

func defaultConfig(v *viper.Viper) *Config {
	cfg := &Config{
		Developing: false,
		LogToFile:  true,
		LogPath:    config.FileDirectory + LogPath,
		MaxSize:    2,
		Compress:   true,
	}
	v.SetDefault("logger", cfg)
	return cfg
}
