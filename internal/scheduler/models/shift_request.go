package models

// ShiftRequest 班次与需求中间表。
// 约束：request_id 全局唯一，确保一个需求只能在一个班次中。
type ShiftRequest struct {
	ShiftID   uint `gorm:"column:shift_id;primaryKey;not null" json:"shift_id"`
	RequestID uint `gorm:"column:request_id;primaryKey;not null;uniqueIndex:uk_shift_requests_request_id" json:"request_id"`

	Shift   *Shift   `gorm:"foreignKey:ShiftID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	Request *Request `gorm:"foreignKey:RequestID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}

func (ShiftRequest) TableName() string {
	return "shift_requests"
}
