package service

import (
	"fmt"
	"time"

	"pickup/internal/model"
	"pickup/internal/repository"

	"go.uber.org/zap"
)

// PaymentService 支付服务接口
type PaymentService interface {
	PreparePayment(userID uint, orderID uint) (*PreparePaymentResponse, error)
	HandlePaymentNotify(req *PaymentNotifyRequest) error
	GetPaymentByOrderID(orderID uint) (*model.PaymentOrder, error)
}

// PreparePaymentResponse 准备支付响应
type PreparePaymentResponse struct {
	PrepayID  string            `json:"prepay_id"`
	PayParams map[string]string `json:"pay_params"`
}

// PaymentNotifyRequest 支付通知请求
type PaymentNotifyRequest struct {
	TransactionID string `json:"transaction_id"`
	OrderID       uint   `json:"order_id"`
	Amount        int64  `json:"amount"`
	Status        string `json:"status"`
}

// paymentService 支付服务实现
type paymentService struct {
	paymentRepo repository.PaymentRepository
	orderRepo   repository.OrderRepository
	logger      *zap.Logger
}

// NewPaymentService 创建支付服务
func NewPaymentService(
	paymentRepo repository.PaymentRepository,
	orderRepo repository.OrderRepository,
	logger *zap.Logger,
) PaymentService {
	return &paymentService{
		paymentRepo: paymentRepo,
		orderRepo:   orderRepo,
		logger:      logger,
	}
}

// PreparePayment 准备支付
func (s *paymentService) PreparePayment(userID uint, orderID uint) (*PreparePaymentResponse, error) {
	// 检查订单是否存在且属于当前用户
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return nil, fmt.Errorf("获取订单信息失败: %w", err)
	}

	if order.PassengerID != userID {
		return nil, fmt.Errorf("无权限支付该订单")
	}

	// 检查订单状态
	if order.Status != model.OrderStatusCreated {
		return nil, fmt.Errorf("订单状态不允许支付")
	}

	// 检查是否已存在支付订单
	existingPayment, err := s.paymentRepo.GetByOrderID(orderID)
	if err == nil && existingPayment != nil {
		// 如果支付订单已存在且状态为pending，返回现有信息
		if existingPayment.State == model.PaymentStatePending {
			return &PreparePaymentResponse{
				PrepayID: existingPayment.WxPrepayID,
				PayParams: map[string]string{
					"appId":     "wx1234567890", // 实际应从配置获取
					"timeStamp": fmt.Sprintf("%d", time.Now().Unix()),
					"nonceStr":  "random_string",
					"package":   fmt.Sprintf("prepay_id=%s", existingPayment.WxPrepayID),
					"signType":  "MD5",
					"paySign":   "mock_sign",
				},
			}, nil
		}
		return nil, fmt.Errorf("订单已支付")
	}

	// 创建支付订单
	payment := &model.PaymentOrder{
		OrderID:    orderID,
		UserID:     userID,
		Amount:     order.PriceTotal,
		Currency:   order.Currency,
		State:      model.PaymentStatePending,
		WxPrepayID: fmt.Sprintf("prepay_%d_%d", orderID, time.Now().Unix()),
	}

	if err := s.paymentRepo.Create(payment); err != nil {
		s.logger.Error("failed to create payment order", zap.Error(err))
		return nil, fmt.Errorf("创建支付订单失败: %w", err)
	}

	// 返回支付参数（这里是模拟数据，实际需要调用微信支付API）
	response := &PreparePaymentResponse{
		PrepayID: payment.WxPrepayID,
		PayParams: map[string]string{
			"appId":     "wx1234567890", // 实际应从配置获取
			"timeStamp": fmt.Sprintf("%d", time.Now().Unix()),
			"nonceStr":  "random_string",
			"package":   fmt.Sprintf("prepay_id=%s", payment.WxPrepayID),
			"signType":  "MD5",
			"paySign":   "mock_sign",
		},
	}

	s.logger.Info("payment prepared", zap.Uint("payment_id", payment.ID))
	return response, nil
}

// HandlePaymentNotify 处理支付通知
func (s *paymentService) HandlePaymentNotify(req *PaymentNotifyRequest) error {
	// 幂等性检查：根据微信交易ID查找支付订单
	payment, err := s.paymentRepo.GetByTransactionID(req.TransactionID)
	if err != nil {
		// 如果找不到，可能是第一次收到通知，根据订单ID查找
		payment, err = s.paymentRepo.GetByOrderID(req.OrderID)
		if err != nil {
			return fmt.Errorf("找不到支付订单: %w", err)
		}
	}

	// 检查支付订单状态，避免重复处理
	if payment.State == model.PaymentStatePaid {
		s.logger.Info("payment already processed", zap.String("transaction_id", req.TransactionID))
		return nil
	}

	// 验证金额
	if payment.Amount != req.Amount {
		s.logger.Error("payment amount mismatch",
			zap.Int64("expected", payment.Amount),
			zap.Int64("actual", req.Amount))
		return fmt.Errorf("支付金额不匹配")
	}

	// 更新支付订单状态
	payment.WxTransactionID = req.TransactionID
	payment.State = model.PaymentStatePaid
	now := time.Now()
	payment.PaidAt = &now

	if err := s.paymentRepo.Update(payment); err != nil {
		s.logger.Error("failed to update payment order", zap.Error(err))
		return fmt.Errorf("更新支付订单失败: %w", err)
	}

	// 更新订单状态为已支付
	if err := s.orderRepo.UpdateStatus(payment.OrderID, model.OrderStatusPaid); err != nil {
		s.logger.Error("failed to update order status", zap.Error(err))
		return fmt.Errorf("更新订单状态失败: %w", err)
	}

	s.logger.Info("payment processed successfully",
		zap.String("transaction_id", req.TransactionID),
		zap.Uint("order_id", payment.OrderID))
	return nil
}

// GetPaymentByOrderID 根据订单ID获取支付信息
func (s *paymentService) GetPaymentByOrderID(orderID uint) (*model.PaymentOrder, error) {
	payment, err := s.paymentRepo.GetByOrderID(orderID)
	if err != nil {
		return nil, fmt.Errorf("获取支付信息失败: %w", err)
	}
	return payment, nil
}
