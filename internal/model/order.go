package model

import (
	"time"

	"gorm.io/gorm"
)

// OrderStatus 订单状态枚举
type OrderStatus string

const (
	OrderStatusCreated   OrderStatus = "created"   // 已创建
	OrderStatusPaid      OrderStatus = "paid"      // 已支付
	OrderStatusAssigned  OrderStatus = "assigned"  // 已分配
	OrderStatusNotified  OrderStatus = "notified"  // 已通知
	OrderStatusCompleted OrderStatus = "completed" // 已完成
	OrderStatusCanceled  OrderStatus = "canceled"  // 已取消
)

// PickupOrder 接机订单模型
type PickupOrder struct {
	ID                   uint        `json:"id" gorm:"primaryKey"`
	PassengerID          uint        `json:"passenger_id" gorm:"not null;index"`
	RegistrationID       uint        `json:"registration_id" gorm:"not null;unique"` // 与报名一对一
	Status               OrderStatus `json:"status" gorm:"type:enum('created','paid','assigned','notified','completed','canceled');default:'created';index:idx_passenger_status"`
	PriceTotal           int64       `json:"price_total" gorm:"not null"`          // 总价格（分）
	Currency             string      `json:"currency" gorm:"size:8;default:'CNY'"` // 货币
	MeetingPoint         string      `json:"meeting_point" gorm:"size:128"`        // 集合点
	Terminal             string      `json:"terminal" gorm:"size:32"`              // 航站楼
	Gate                 string      `json:"gate" gorm:"size:16"`                  // 登机口
	ScheduledArrivalTime time.Time   `json:"scheduled_arrival_time"`               // 计划到达时间
	EstimatedArrivalTime *time.Time  `json:"estimated_arrival_time"`               // 预估到达时间
	CreatedAt            time.Time   `json:"created_at"`
	UpdatedAt            time.Time   `json:"updated_at"`

	// 关联
	Passenger    *User         `json:"passenger,omitempty" gorm:"foreignKey:PassengerID"`
	Registration *Registration `json:"registration,omitempty" gorm:"foreignKey:RegistrationID"`
	Assignment   *Assignment   `json:"assignment,omitempty" gorm:"foreignKey:OrderID"`
	PaymentOrder *PaymentOrder `json:"payment_order,omitempty" gorm:"foreignKey:OrderID"`
}

// TableName 指定表名
func (PickupOrder) TableName() string {
	return "pickup_orders"
}

// BeforeCreate GORM钩子
func (o *PickupOrder) BeforeCreate(tx *gorm.DB) error {
	if o.Status == "" {
		o.Status = OrderStatusCreated
	}
	if o.Currency == "" {
		o.Currency = "CNY"
	}
	return nil
}
