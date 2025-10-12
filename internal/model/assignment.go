package model

import (
	"time"

	"gorm.io/gorm"
)

// AssignmentStatus 分配状态枚举
type AssignmentStatus string

const (
	AssignmentStatusAssigned AssignmentStatus = "assigned" // 已分配
	AssignmentStatusAccepted AssignmentStatus = "accepted" // 已接受
	AssignmentStatusRejected AssignmentStatus = "rejected" // 已拒绝
)

// Assignment 司机分配模型
type Assignment struct {
	ID         uint             `json:"id" gorm:"primaryKey"`
	OrderID    uint             `json:"order_id" gorm:"not null;unique"`                   // 订单ID
	DriverID   uint             `json:"driver_id" gorm:"not null;index:idx_driver_status"` // 司机ID
	Status     AssignmentStatus `json:"status" gorm:"type:enum('assigned','accepted','rejected');default:'assigned';index:idx_driver_status"`
	AssignedAt time.Time        `json:"assigned_at"`
	AcceptedAt *time.Time       `json:"accepted_at"`
	RejectedAt *time.Time       `json:"rejected_at"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`

	// 关联
	Order  *PickupOrder `json:"order,omitempty" gorm:"foreignKey:OrderID"`
	Driver *Driver      `json:"driver,omitempty" gorm:"foreignKey:DriverID"`
}

// TableName 指定表名
func (Assignment) TableName() string {
	return "assignments"
}

// BeforeCreate GORM钩子
func (a *Assignment) BeforeCreate(tx *gorm.DB) error {
	if a.Status == "" {
		a.Status = AssignmentStatusAssigned
	}
	if a.AssignedAt.IsZero() {
		a.AssignedAt = time.Now()
	}
	return nil
}
