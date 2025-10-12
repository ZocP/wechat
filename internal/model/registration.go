package model

import (
	"time"
)

// PickupMethod 接机方式枚举
type PickupMethod string

const (
	PickupMethodGroup   PickupMethod = "group"   // 团体接机
	PickupMethodPrivate PickupMethod = "private" // 私人接机
	PickupMethodShuttle PickupMethod = "shuttle" // 班车接机
)

// RegistrationStatus 报名状态枚举
type RegistrationStatus string

const (
	RegistrationStatusDraft     RegistrationStatus = "draft"     // 草稿
	RegistrationStatusSubmitted RegistrationStatus = "submitted" // 已提交
	RegistrationStatusConfirmed RegistrationStatus = "confirmed" // 已确认
)

// Registration 报名模型
type Registration struct {
	ID            uint               `json:"id" gorm:"primaryKey"`
	UserID        uint               `json:"user_id" gorm:"not null;index"`
	Name          string             `json:"name" gorm:"size:64;not null"`                                             // 姓名
	Phone         string             `json:"phone" gorm:"size:20;not null"`                                            // 手机号
	WechatID      string             `json:"wechat_id" gorm:"size:64"`                                                 // 微信号
	FlightNo      string             `json:"flight_no" gorm:"size:16;not null"`                                        // 航班号
	ArrivalDate   time.Time          `json:"arrival_date" gorm:"type:date;not null"`                                   // 到达日期
	ArrivalTime   string             `json:"arrival_time" gorm:"type:time;not null"`                                   // 到达时间
	DepartureCity string             `json:"departure_city" gorm:"size:64;not null"`                                   // 出发城市
	Companions    int                `json:"companions" gorm:"default:0"`                                              // 随行人数
	LuggageCount  int                `json:"luggage_count" gorm:"default:0"`                                           // 行李件数
	PickupMethod  PickupMethod       `json:"pickup_method" gorm:"type:enum('group','private','shuttle');not null"`     // 接机方式
	Notes         string             `json:"notes" gorm:"type:text"`                                                   // 备注
	Status        RegistrationStatus `json:"status" gorm:"type:enum('draft','submitted','confirmed');default:'draft'"` // 状态
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`

	// 关联
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName 指定表名
func (Registration) TableName() string {
	return "registrations"
}
