package config

import (
	"time"
)

// JWTConfig JWT配置
type JWTConfig struct {
	Secret     string        `yaml:"secret"`
	ExpireTime time.Duration `yaml:"expireTime"`
	Issuer     string        `yaml:"issuer"`
}

// NewJWTConfig 创建JWT配置
func NewJWTConfig() *JWTConfig {
	return &JWTConfig{
		Secret:     getEnv("JWT_SECRET", "pickup-secret-key"),
		ExpireTime: time.Duration(getEnvInt("JWT_EXPIRE_HOURS", 24)) * time.Hour,
		Issuer:     getEnv("JWT_ISSUER", "pickup"),
	}
}
