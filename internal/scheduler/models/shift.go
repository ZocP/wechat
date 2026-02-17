package models

import "time"

// Shift 调度班次表
type Shift struct {
	ID            uint        `gorm:"primaryKey" json:"id"`
	DriverID      uint        `gorm:"column:driver_id;not null;index:idx_shifts_driver_id" json:"driver_id"`
	DepartureTime time.Time   `gorm:"column:departure_time;type:datetime;not null;index:idx_shifts_departure_time" json:"departure_time"`
	Status        ShiftStatus `gorm:"type:enum('draft','published');not null;default:'draft';index:idx_shifts_status" json:"status"`
	CreatedAt     time.Time   `json:"created_at"`

	Driver   *Driver   `gorm:"foreignKey:DriverID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"driver,omitempty"`
	Requests []Request `gorm:"many2many:shift_requests;joinForeignKey:ShiftID;joinReferences:RequestID" json:"requests,omitempty"`
	Staffs   []User    `gorm:"many2many:shift_staffs;joinForeignKey:ShiftID;joinReferences:StaffID" json:"staffs,omitempty"`
}

func (Shift) TableName() string {
	return "shifts"
}
