package config

// CryptoConfig 加密配置
type CryptoConfig struct {
	Key string `yaml:"key"`
}

// NewCryptoConfig 创建加密配置
func NewCryptoConfig() *CryptoConfig {
	return &CryptoConfig{
		Key: getEnvOrConfig("CRYPTO_KEY", "crypto.key", "pickup-crypto-key-32-characters-long"),
	}
}
