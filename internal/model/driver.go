package model

import (
	"time"

	"gorm.io/gorm"
)

// DriverStatus 司机状态枚举
type DriverStatus string

const (
	DriverStatusAvailable DriverStatus = "available" // 可用
	DriverStatusBusy      DriverStatus = "busy"      // 忙碌
	DriverStatusOffline   DriverStatus = "offline"   // 离线
)

// Driver 司机模型
type Driver struct {
	ID         uint         `json:"id" gorm:"primaryKey"`
	UserID     uint         `json:"user_id" gorm:"not null;unique"`
	Name       string       `json:"name" gorm:"size:64;not null"`       // 姓名
	Phone      string       `json:"phone" gorm:"size:20;not null"`      // 手机号
	LicenseNo  string       `json:"license_no" gorm:"size:32;not null"` // 驾照号（加密存储）
	Status     DriverStatus `json:"status" gorm:"type:enum('available','busy','offline');default:'offline'"`
	VerifiedAt *time.Time   `json:"verified_at"` // 验证时间
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`

	// 关联
	User        *User        `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Vehicle     *Vehicle     `json:"vehicle,omitempty"`
	Assignments []Assignment `json:"assignments,omitempty" gorm:"foreignKey:DriverID"`
}

// TableName 指定表名
func (Driver) TableName() string {
	return "drivers"
}

// BeforeCreate GORM钩子
func (d *Driver) BeforeCreate(tx *gorm.DB) error {
	if d.Status == "" {
		d.Status = DriverStatusOffline
	}
	return nil
}
