package config

import (
	"os"
	"strconv"
	"sync"

	"github.com/spf13/viper"
)

var (
	configOnce   sync.Once
	fileConfigV  *viper.Viper
	fileConfigOk bool
)

func loadFileConfig() {
	configOnce.Do(func() {
		v := viper.New()
		candidates := []string{
			"files/config.yaml",
			"./files/config.yaml",
		}

		for _, path := range candidates {
			if _, err := os.Stat(path); err != nil {
				continue
			}
			v.SetConfigFile(path)
			if err := v.ReadInConfig(); err == nil {
				fileConfigV = v
				fileConfigOk = true
				return
			}
		}

		fileConfigV = v
		fileConfigOk = false
	})
}

func getConfigString(key string) (string, bool) {
	loadFileConfig()
	if !fileConfigOk || !fileConfigV.IsSet(key) {
		return "", false
	}
	return fileConfigV.GetString(key), true
}

func getConfigInt(key string) (int, bool) {
	loadFileConfig()
	if !fileConfigOk || !fileConfigV.IsSet(key) {
		return 0, false
	}
	return fileConfigV.GetInt(key), true
}

func getEnvOrConfig(envKey, configKey, defaultValue string) string {
	if value := os.Getenv(envKey); value != "" {
		return value
	}
	if value, ok := getConfigString(configKey); ok && value != "" {
		return value
	}
	return defaultValue
}

func getEnvOrConfigInt(envKey, configKey string, defaultValue int) int {
	if value := os.Getenv(envKey); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	if value, ok := getConfigInt(configKey); ok {
		return value
	}
	return defaultValue
}
