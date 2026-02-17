package config

import (
	"fmt"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func TestProvide_NotNil(t *testing.T) {
	op := Provide()
	assert.NotNil(t, op)
}

func TestAutoMigrate_WithSQLite(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	err = autoMigrate(db)
	require.Error(t, err)
}

func TestNewDatabase_ErrorWhenDBUnavailable(t *testing.T) {
	cfg := &DatabaseConfig{
		Host:         "127.0.0.1",
		Port:         1,
		User:         "x",
		Password:     "x",
		Database:     "x",
		Charset:      "utf8mb4",
		ParseTime:    true,
		Loc:          "Local",
		MaxOpenConns: 1,
		MaxIdleConns: 1,
		MaxLifetime:  1,
	}
	db, err := NewDatabase(cfg, zap.NewNop())
	require.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, fmt.Sprintf("%v", err), "failed")
}
