package model

import (
	"time"
)

// Vehicle 车辆模型
type Vehicle struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	DriverID  uint      `json:"driver_id" gorm:"not null;unique"` // 司机ID，一对一关系
	PlateNo   string    `json:"plate_no" gorm:"size:16;not null"` // 车牌号（加密存储）
	Brand     string    `json:"brand" gorm:"size:32;not null"`    // 品牌
	Model     string    `json:"model" gorm:"size:32;not null"`    // 型号
	Seats     int       `json:"seats" gorm:"not null"`            // 座位数
	Color     string    `json:"color" gorm:"size:16;not null"`    // 颜色
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联
	Driver *Driver `json:"driver,omitempty" gorm:"foreignKey:DriverID"`
}

// TableName 指定表名
func (Vehicle) TableName() string {
	return "vehicles"
}
