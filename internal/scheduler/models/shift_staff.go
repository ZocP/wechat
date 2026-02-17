package models

// ShiftStaff 班次与志愿者中间表。
type ShiftStaff struct {
	ShiftID uint `gorm:"column:shift_id;primaryKey;not null" json:"shift_id"`
	StaffID uint `gorm:"column:staff_id;primaryKey;not null" json:"staff_id"`

	Shift *Shift `gorm:"foreignKey:ShiftID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	Staff *User  `gorm:"foreignKey:StaffID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}

func (ShiftStaff) TableName() string {
	return "shift_staffs"
}
