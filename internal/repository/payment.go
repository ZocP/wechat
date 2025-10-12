package repository

import (
	"pickup/internal/model"

	"gorm.io/gorm"
)

// PaymentRepository 支付仓储接口
type PaymentRepository interface {
	Create(payment *model.PaymentOrder) error
	GetByID(id uint) (*model.PaymentOrder, error)
	GetByOrderID(orderID uint) (*model.PaymentOrder, error)
	GetByTransactionID(transactionID string) (*model.PaymentOrder, error)
	Update(payment *model.PaymentOrder) error
	UpdateState(paymentID uint, state model.PaymentState) error
}

// paymentRepository 支付仓储实现
type paymentRepository struct {
	db *gorm.DB
}

// NewPaymentRepository 创建支付仓储
func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

// Create 创建支付订单
func (r *paymentRepository) Create(payment *model.PaymentOrder) error {
	return r.db.Create(payment).Error
}

// GetByID 根据ID获取支付订单
func (r *paymentRepository) GetByID(id uint) (*model.PaymentOrder, error) {
	var payment model.PaymentOrder
	err := r.db.Preload("Order").Preload("User").First(&payment, id).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// GetByOrderID 根据订单ID获取支付订单
func (r *paymentRepository) GetByOrderID(orderID uint) (*model.PaymentOrder, error) {
	var payment model.PaymentOrder
	err := r.db.Where("order_id = ?", orderID).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// GetByTransactionID 根据微信交易ID获取支付订单
func (r *paymentRepository) GetByTransactionID(transactionID string) (*model.PaymentOrder, error) {
	var payment model.PaymentOrder
	err := r.db.Where("wx_transaction_id = ?", transactionID).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// Update 更新支付订单
func (r *paymentRepository) Update(payment *model.PaymentOrder) error {
	return r.db.Save(payment).Error
}

// UpdateState 更新支付状态
func (r *paymentRepository) UpdateState(paymentID uint, state model.PaymentState) error {
	updates := map[string]interface{}{"state": state}
	if state == model.PaymentStatePaid {
		updates["paid_at"] = gorm.Expr("NOW()")
	}
	return r.db.Model(&model.PaymentOrder{}).Where("id = ?", paymentID).Updates(updates).Error
}
