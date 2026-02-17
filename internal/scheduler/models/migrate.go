package models

import "gorm.io/gorm"

// AutoMigrate 自动迁移调度域表。
func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&User{},
		&Driver{},
		&Request{},
		&Shift{},
	); err != nil {
		return err
	}

	if err := db.SetupJoinTable(&Shift{}, "Requests", &ShiftRequest{}); err != nil {
		return err
	}
	if err := db.SetupJoinTable(&Shift{}, "Staffs", &ShiftStaff{}); err != nil {
		return err
	}

	return db.AutoMigrate(
		&ShiftRequest{},
		&ShiftStaff{},
	)
}
