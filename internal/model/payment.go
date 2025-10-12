package model

import (
	"time"

	"gorm.io/gorm"
)

// PaymentState 支付状态枚举
type PaymentState string

const (
	PaymentStatePending  PaymentState = "pending"  // 待支付
	PaymentStatePaid     PaymentState = "paid"     // 已支付
	PaymentStateRefunded PaymentState = "refunded" // 已退款
	PaymentStateClosed   PaymentState = "closed"   // 已关闭
)

// PaymentOrder 支付订单模型
type PaymentOrder struct {
	ID              uint         `json:"id" gorm:"primaryKey"`
	OrderID         uint         `json:"order_id" gorm:"not null;unique"`                                                // 关联的订单ID
	UserID          uint         `json:"user_id" gorm:"not null"`                                                        // 用户ID
	WxPrepayID      string       `json:"wx_prepay_id" gorm:"size:64"`                                                    // 微信预支付ID
	WxTransactionID string       `json:"wx_transaction_id" gorm:"size:64"`                                               // 微信交易ID
	Amount          int64        `json:"amount" gorm:"not null"`                                                         // 支付金额（分）
	Currency        string       `json:"currency" gorm:"size:8;default:'CNY'"`                                           // 货币
	State           PaymentState `json:"state" gorm:"type:enum('pending','paid','refunded','closed');default:'pending'"` // 支付状态
	PaidAt          *time.Time   `json:"paid_at"`                                                                        // 支付时间
	RefundID        string       `json:"refund_id" gorm:"size:64"`                                                       // 退款ID
	RefundedAt      *time.Time   `json:"refunded_at"`                                                                    // 退款时间
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`

	// 关联
	Order *PickupOrder `json:"order,omitempty" gorm:"foreignKey:OrderID"`
	User  *User        `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName 指定表名
func (PaymentOrder) TableName() string {
	return "payment_orders"
}

// BeforeCreate GORM钩子
func (p *PaymentOrder) BeforeCreate(tx *gorm.DB) error {
	if p.State == "" {
		p.State = PaymentStatePending
	}
	if p.Currency == "" {
		p.Currency = "CNY"
	}
	return nil
}
