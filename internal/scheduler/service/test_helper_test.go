package service

import (
	"fmt"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	schema := []string{
		`CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			open_id TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			phone TEXT,
			role TEXT NOT NULL DEFAULT 'student',
			created_at DATETIME,
			updated_at DATETIME
		);`,
		`CREATE TABLE drivers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			car_model TEXT NOT NULL,
			max_seats INTEGER NOT NULL,
			max_checked INTEGER NOT NULL,
			max_carry_on INTEGER NOT NULL
		);`,
		`CREATE TABLE requests (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			flight_no TEXT NOT NULL,
			arrival_date DATETIME NOT NULL,
			terminal TEXT NOT NULL,
			checked_bags INTEGER NOT NULL DEFAULT 0,
			carry_on_bags INTEGER NOT NULL DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'pending',
			arrival_time_api DATETIME,
			pickup_buffer INTEGER NOT NULL DEFAULT 45,
			calc_pickup_time DATETIME,
			created_at DATETIME,
			updated_at DATETIME
		);`,
		`CREATE TABLE shifts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			driver_id INTEGER NOT NULL,
			departure_time DATETIME NOT NULL,
			status TEXT NOT NULL DEFAULT 'draft',
			created_at DATETIME
		);`,
		`CREATE TABLE shift_requests (
			shift_id INTEGER NOT NULL,
			request_id INTEGER NOT NULL UNIQUE,
			PRIMARY KEY (shift_id, request_id)
		);`,
		`CREATE TABLE shift_staffs (
			shift_id INTEGER NOT NULL,
			staff_id INTEGER NOT NULL,
			PRIMARY KEY (shift_id, staff_id)
		);`,
	}

	for _, ddl := range schema {
		require.NoError(t, db.Exec(ddl).Error)
	}

	return db
}
