package server

import (
	"github.com/spf13/viper"
)

type Config struct {
	Addr        string `yaml:"addr"`
	Port        int    `yaml:"port"`
	AllowCORS   bool   `yaml:"allowCORS"`
	ReleaseMode bool   `yaml:"releaseMode"`
}

func NewConfig(v *viper.Viper) (*Config, error) {
	cfg := new(Config)
	if !v.IsSet("server") {
		return defaultConfig(v), nil
	}
	if err := v.UnmarshalKey("server", cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func defaultConfig(viper *viper.Viper) *Config {
	cfg := &Config{
		Port:        8080,
		AllowCORS:   true,
		ReleaseMode: false,
	}
	viper.SetDefault("server", cfg)
	return cfg
}
