package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetEnv_Default(t *testing.T) {
	val := getEnv("NONEXISTENT_TEST_VAR_123456", "default_value")
	assert.Equal(t, "default_value", val)
}

func TestGetEnv_Set(t *testing.T) {
	os.Setenv("TEST_GET_ENV_VAR", "custom_value")
	defer os.Unsetenv("TEST_GET_ENV_VAR")

	val := getEnv("TEST_GET_ENV_VAR", "default_value")
	assert.Equal(t, "custom_value", val)
}

func TestGetEnv_Empty(t *testing.T) {
	os.Setenv("TEST_EMPTY_VAR", "")
	defer os.Unsetenv("TEST_EMPTY_VAR")

	val := getEnv("TEST_EMPTY_VAR", "default_value")
	assert.Equal(t, "default_value", val)
}

func TestGetEnvInt_Default(t *testing.T) {
	val := getEnvInt("NONEXISTENT_INT_VAR_123456", 42)
	assert.Equal(t, 42, val)
}

func TestGetEnvInt_Set(t *testing.T) {
	os.Setenv("TEST_INT_VAR", "100")
	defer os.Unsetenv("TEST_INT_VAR")

	val := getEnvInt("TEST_INT_VAR", 42)
	assert.Equal(t, 100, val)
}

func TestGetEnvInt_Invalid(t *testing.T) {
	os.Setenv("TEST_BAD_INT", "not_a_number")
	defer os.Unsetenv("TEST_BAD_INT")

	val := getEnvInt("TEST_BAD_INT", 42)
	assert.Equal(t, 42, val)
}

func TestGetEnvInt_Empty(t *testing.T) {
	os.Setenv("TEST_EMPTY_INT", "")
	defer os.Unsetenv("TEST_EMPTY_INT")

	val := getEnvInt("TEST_EMPTY_INT", 42)
	assert.Equal(t, 42, val)
}

// ===== Database Config Tests =====

func TestNewDatabaseConfig_Defaults(t *testing.T) {
	cfg := NewDatabaseConfig()
	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 3306, cfg.Port)
	assert.Equal(t, "root", cfg.User)
	assert.Equal(t, "", cfg.Password)
	assert.Equal(t, "pickup", cfg.Database)
	assert.Equal(t, "utf8mb4", cfg.Charset)
	assert.True(t, cfg.ParseTime)
	assert.Equal(t, "Local", cfg.Loc)
	assert.Equal(t, 100, cfg.MaxOpenConns)
	assert.Equal(t, 10, cfg.MaxIdleConns)
	assert.Equal(t, 3600, cfg.MaxLifetime)
}

func TestNewDatabaseConfig_CustomEnv(t *testing.T) {
	os.Setenv("DB_HOST", "custom-host")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
	}()

	cfg := NewDatabaseConfig()
	assert.Equal(t, "custom-host", cfg.Host)
	assert.Equal(t, 5432, cfg.Port)
	assert.Equal(t, "testuser", cfg.User)
	assert.Equal(t, "testpass", cfg.Password)
	assert.Equal(t, "testdb", cfg.Database)
}

// ===== JWT Config Tests =====

func TestNewJWTConfig_Defaults(t *testing.T) {
	cfg := NewJWTConfig()
	assert.Equal(t, "pickup-secret-key", cfg.Secret)
	assert.Equal(t, 24*time.Hour, cfg.ExpireTime)
	assert.Equal(t, "pickup", cfg.Issuer)
}

func TestNewJWTConfig_CustomEnv(t *testing.T) {
	os.Setenv("JWT_SECRET", "custom-secret")
	os.Setenv("JWT_EXPIRE_HOURS", "48")
	os.Setenv("JWT_ISSUER", "custom-issuer")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("JWT_EXPIRE_HOURS")
		os.Unsetenv("JWT_ISSUER")
	}()

	cfg := NewJWTConfig()
	assert.Equal(t, "custom-secret", cfg.Secret)
	assert.Equal(t, 48*time.Hour, cfg.ExpireTime)
	assert.Equal(t, "custom-issuer", cfg.Issuer)
}

// ===== Wechat Config Tests =====

func TestNewWechatConfig_Defaults(t *testing.T) {
	cfg := NewWechatConfig()
	assert.Empty(t, cfg.AppID)
	assert.Empty(t, cfg.AppSecret)
	assert.Empty(t, cfg.MchID)
	assert.Empty(t, cfg.MchKey)
	assert.Empty(t, cfg.NotifyURL)
	assert.Empty(t, cfg.AdminPhone)
	assert.Empty(t, cfg.AdminOpenID)
}

func TestNewWechatConfig_CustomEnv(t *testing.T) {
	os.Setenv("WECHAT_APPID", "wx_test_appid")
	os.Setenv("WECHAT_SECRET", "wx_test_secret")
	os.Setenv("WECHAT_MCH_ID", "wx_test_mch")
	os.Setenv("WECHAT_MCH_KEY", "wx_test_key")
	os.Setenv("WECHAT_NOTIFY_URL", "https://example.com/notify")
	os.Setenv("WECHAT_ADMIN_PHONE", "13928998540")
	os.Setenv("WECHAT_ADMIN_OPEN_ID", "openid_admin")
	defer func() {
		os.Unsetenv("WECHAT_APPID")
		os.Unsetenv("WECHAT_SECRET")
		os.Unsetenv("WECHAT_MCH_ID")
		os.Unsetenv("WECHAT_MCH_KEY")
		os.Unsetenv("WECHAT_NOTIFY_URL")
		os.Unsetenv("WECHAT_ADMIN_PHONE")
		os.Unsetenv("WECHAT_ADMIN_OPEN_ID")
	}()

	cfg := NewWechatConfig()
	assert.Equal(t, "wx_test_appid", cfg.AppID)
	assert.Equal(t, "wx_test_secret", cfg.AppSecret)
	assert.Equal(t, "wx_test_mch", cfg.MchID)
	assert.Equal(t, "wx_test_key", cfg.MchKey)
	assert.Equal(t, "https://example.com/notify", cfg.NotifyURL)
	assert.Equal(t, "13928998540", cfg.AdminPhone)
	assert.Equal(t, "openid_admin", cfg.AdminOpenID)
}

// ===== Crypto Config Tests =====

func TestNewCryptoConfig_Defaults(t *testing.T) {
	cfg := NewCryptoConfig()
	require.NotEmpty(t, cfg.Key)
	assert.Equal(t, "pickup-crypto-key-32-characters-long", cfg.Key)
}

func TestNewCryptoConfig_CustomEnv(t *testing.T) {
	os.Setenv("CRYPTO_KEY", "custom-crypto-key-for-testing!!!")
	defer os.Unsetenv("CRYPTO_KEY")

	cfg := NewCryptoConfig()
	assert.Equal(t, "custom-crypto-key-for-testing!!!", cfg.Key)
}
