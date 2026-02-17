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
		Secret:     getEnvOrConfig("JWT_SECRET", "jwt.secret", "pickup-secret-key"),
		ExpireTime: time.Duration(getEnvOrConfigInt("JWT_EXPIRE_HOURS", "jwt.expireHours", 24)) * time.Hour,
		Issuer:     getEnvOrConfig("JWT_ISSUER", "jwt.issuer", "pickup"),
	}
}
