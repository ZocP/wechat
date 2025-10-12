package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"pickup/internal/model"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	Database     string `yaml:"database"`
	Charset      string `yaml:"charset"`
	ParseTime    bool   `yaml:"parseTime"`
	Loc          string `yaml:"loc"`
	MaxOpenConns int    `yaml:"maxOpenConns"`
	MaxIdleConns int    `yaml:"maxIdleConns"`
	MaxLifetime  int    `yaml:"maxLifetime"`
}

// NewDatabaseConfig 创建数据库配置
func NewDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:         getEnv("DB_HOST", "localhost"),
		Port:         getEnvInt("DB_PORT", 3306),
		User:         getEnv("DB_USER", "root"),
		Password:     getEnv("DB_PASSWORD", ""),
		Database:     getEnv("DB_NAME", "pickup"),
		Charset:      "utf8mb4",
		ParseTime:    true,
		Loc:          "Local",
		MaxOpenConns: 100,
		MaxIdleConns: 10,
		MaxLifetime:  3600,
	}
}

// NewDatabase 创建数据库连接
func NewDatabase(cfg *DatabaseConfig, logger *zap.Logger) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.Charset, cfg.ParseTime, cfg.Loc)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second)

	// 自动迁移数据库表
	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	logger.Info("database connected successfully")
	return db, nil
}

// autoMigrate 自动迁移数据库表
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Registration{},
		&model.PickupOrder{},
		&model.Assignment{},
		&model.Driver{},
		&model.Vehicle{},
		&model.Notice{},
		&model.PaymentOrder{},
		&model.ConsentLog{},
	)
}

// Provide 提供依赖注入
func Provide() fx.Option {
	return fx.Options(
		fx.Provide(NewDatabaseConfig),
		fx.Provide(NewJWTConfig),
		fx.Provide(NewWechatConfig),
		fx.Provide(NewCryptoConfig),
		fx.Provide(NewDatabase),
	)
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt 获取环境变量并转换为int，如果不存在或转换失败则返回默认值
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
