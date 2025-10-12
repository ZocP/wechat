package model

import (
	"time"
)

// Notice 消息板/班次公告模型
type Notice struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	FlightNo       string    `json:"flight_no" gorm:"size:16;not null;index"` // 航班号
	Terminal       string    `json:"terminal" gorm:"size:32;not null"`        // 航站楼
	PickupBatch    string    `json:"pickup_batch" gorm:"size:32;not null"`    // 接机班次
	ArrivalAirport string    `json:"arrival_airport" gorm:"size:64;not null"` // 到达机场
	MeetingPoint   string    `json:"meeting_point" gorm:"size:128;not null"`  // 集合点
	GuideText      string    `json:"guide_text" gorm:"type:text"`             // 引导文字
	MapURL         string    `json:"map_url" gorm:"size:255"`                 // 地图URL
	ContactName    string    `json:"contact_name" gorm:"size:64"`             // 联系人姓名
	ContactPhone   string    `json:"contact_phone" gorm:"size:32"`            // 联系电话（脱敏显示）
	VisibleFrom    time.Time `json:"visible_from" gorm:"not null"`            // 可见开始时间
	VisibleTo      time.Time `json:"visible_to" gorm:"not null"`              // 可见结束时间
	CreatedBy      uint      `json:"created_by" gorm:"not null"`              // 创建人
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// 关联
	Creator *User `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`
}

// TableName 指定表名
func (Notice) TableName() string {
	return "notices"
}
