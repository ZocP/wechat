package config

import ()

// CryptoConfig 加密配置
type CryptoConfig struct {
	Key string `yaml:"key"`
}

// NewCryptoConfig 创建加密配置
func NewCryptoConfig() *CryptoConfig {
	return &CryptoConfig{
		Key: getEnv("CRYPTO_KEY", "pickup-crypto-key-32-characters-long"),
	}
}
