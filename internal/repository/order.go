package repository

import (
	"pickup/internal/model"

	"gorm.io/gorm"
)

// OrderRepository 订单仓储接口
type OrderRepository interface {
	Create(order *model.PickupOrder) error
	GetByID(id uint) (*model.PickupOrder, error)
	GetByUserID(userID uint) ([]*model.PickupOrder, error)
	GetByRegistrationID(registrationID uint) (*model.PickupOrder, error)
	Update(order *model.PickupOrder) error
	UpdateStatus(orderID uint, status model.OrderStatus) error
	Delete(id uint) error
}

// orderRepository 订单仓储实现
type orderRepository struct {
	db *gorm.DB
}

// NewOrderRepository 创建订单仓储
func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

// Create 创建订单
func (r *orderRepository) Create(order *model.PickupOrder) error {
	return r.db.Create(order).Error
}

// GetByID 根据ID获取订单
func (r *orderRepository) GetByID(id uint) (*model.PickupOrder, error) {
	var order model.PickupOrder
	err := r.db.Preload("Passenger").Preload("Registration").Preload("Assignment.Driver").Preload("PaymentOrder").First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// GetByUserID 根据用户ID获取订单列表
func (r *orderRepository) GetByUserID(userID uint) ([]*model.PickupOrder, error) {
	var orders []*model.PickupOrder
	err := r.db.Where("passenger_id = ?", userID).
		Preload("Registration").
		Preload("Assignment.Driver").
		Preload("PaymentOrder").
		Order("created_at DESC").
		Find(&orders).Error
	if err != nil {
		return nil, err
	}
	return orders, nil
}

// GetByRegistrationID 根据报名ID获取订单
func (r *orderRepository) GetByRegistrationID(registrationID uint) (*model.PickupOrder, error) {
	var order model.PickupOrder
	err := r.db.Where("registration_id = ?", registrationID).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// Update 更新订单
func (r *orderRepository) Update(order *model.PickupOrder) error {
	return r.db.Save(order).Error
}

// UpdateStatus 更新订单状态
func (r *orderRepository) UpdateStatus(orderID uint, status model.OrderStatus) error {
	return r.db.Model(&model.PickupOrder{}).Where("id = ?", orderID).Update("status", status).Error
}

// Delete 删除订单
func (r *orderRepository) Delete(id uint) error {
	return r.db.Delete(&model.PickupOrder{}, id).Error
}
