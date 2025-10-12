package service

import (
	"fmt"
	"time"

	"pickup/internal/model"
	"pickup/internal/repository"

	"go.uber.org/zap"
)

// OrderService 订单服务接口
type OrderService interface {
	CreateOrder(userID uint, req *CreateOrderRequest) (*model.PickupOrder, error)
	GetOrder(id uint, userID uint) (*model.PickupOrder, error)
	GetUserOrders(userID uint) ([]*model.PickupOrder, error)
	UpdateOrderStatus(orderID uint, status model.OrderStatus) error
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	RegistrationID uint   `json:"registration_id" binding:"required"`
	PriceTotal     int64  `json:"price_total" binding:"required"`
	Currency       string `json:"currency"`
}

// orderService 订单服务实现
type orderService struct {
	orderRepo        repository.OrderRepository
	registrationRepo repository.RegistrationRepository
	logger           *zap.Logger
}

// NewOrderService 创建订单服务
func NewOrderService(
	orderRepo repository.OrderRepository,
	registrationRepo repository.RegistrationRepository,
	logger *zap.Logger,
) OrderService {
	return &orderService{
		orderRepo:        orderRepo,
		registrationRepo: registrationRepo,
		logger:           logger,
	}
}

// CreateOrder 创建订单
func (s *orderService) CreateOrder(userID uint, req *CreateOrderRequest) (*model.PickupOrder, error) {
	// 检查报名是否存在且属于当前用户
	registration, err := s.registrationRepo.GetByID(req.RegistrationID)
	if err != nil {
		return nil, fmt.Errorf("获取报名信息失败: %w", err)
	}

	if registration.UserID != userID {
		return nil, fmt.Errorf("无权限为该报名创建订单")
	}

	// 检查是否已经存在订单
	existingOrder, err := s.orderRepo.GetByRegistrationID(req.RegistrationID)
	if err == nil && existingOrder != nil {
		return nil, fmt.Errorf("该报名已存在订单")
	}

	// 计算计划到达时间
	scheduledArrivalTime := time.Date(
		registration.ArrivalDate.Year(),
		registration.ArrivalDate.Month(),
		registration.ArrivalDate.Day(),
		0, 0, 0, 0, time.Local,
	)

	// 解析时间部分
	arrivalTime, err := time.Parse("15:04:05", registration.ArrivalTime)
	if err == nil {
		scheduledArrivalTime = scheduledArrivalTime.Add(
			time.Hour*time.Duration(arrivalTime.Hour()) +
				time.Minute*time.Duration(arrivalTime.Minute()),
		)
	}

	// 设置默认货币
	currency := req.Currency
	if currency == "" {
		currency = "CNY"
	}

	order := &model.PickupOrder{
		PassengerID:          userID,
		RegistrationID:       req.RegistrationID,
		Status:               model.OrderStatusCreated,
		PriceTotal:           req.PriceTotal,
		Currency:             currency,
		ScheduledArrivalTime: scheduledArrivalTime,
	}

	if err := s.orderRepo.Create(order); err != nil {
		s.logger.Error("failed to create order", zap.Error(err))
		return nil, fmt.Errorf("创建订单失败: %w", err)
	}

	// 更新报名状态为已提交
	registration.Status = model.RegistrationStatusSubmitted
	if err := s.registrationRepo.Update(registration); err != nil {
		s.logger.Warn("failed to update registration status", zap.Error(err))
	}

	s.logger.Info("order created", zap.Uint("order_id", order.ID))
	return order, nil
}

// GetOrder 获取订单信息
func (s *orderService) GetOrder(id uint, userID uint) (*model.PickupOrder, error) {
	order, err := s.orderRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("获取订单信息失败: %w", err)
	}

	// 检查权限：只能查看自己的订单
	if order.PassengerID != userID {
		return nil, fmt.Errorf("无权限访问该订单")
	}

	return order, nil
}

// GetUserOrders 获取用户订单列表
func (s *orderService) GetUserOrders(userID uint) ([]*model.PickupOrder, error) {
	orders, err := s.orderRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("获取订单列表失败: %w", err)
	}
	return orders, nil
}

// UpdateOrderStatus 更新订单状态
func (s *orderService) UpdateOrderStatus(orderID uint, status model.OrderStatus) error {
	if err := s.orderRepo.UpdateStatus(orderID, status); err != nil {
		s.logger.Error("failed to update order status", zap.Error(err))
		return fmt.Errorf("更新订单状态失败: %w", err)
	}

	s.logger.Info("order status updated",
		zap.Uint("order_id", orderID),
		zap.String("status", string(status)))
	return nil
}
