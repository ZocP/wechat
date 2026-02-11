package service

import (
	"errors"
	"testing"
	"time"

	"pickup/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// ===== Mock Payment Repository =====

type mockPaymentRepo struct {
	mock.Mock
}

func (m *mockPaymentRepo) Create(payment *model.PaymentOrder) error {
	args := m.Called(payment)
	return args.Error(0)
}

func (m *mockPaymentRepo) GetByID(id uint) (*model.PaymentOrder, error) {
	args := m.Called(id)
	if v := args.Get(0); v != nil {
		return v.(*model.PaymentOrder), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockPaymentRepo) GetByOrderID(orderID uint) (*model.PaymentOrder, error) {
	args := m.Called(orderID)
	if v := args.Get(0); v != nil {
		return v.(*model.PaymentOrder), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockPaymentRepo) GetByTransactionID(transactionID string) (*model.PaymentOrder, error) {
	args := m.Called(transactionID)
	if v := args.Get(0); v != nil {
		return v.(*model.PaymentOrder), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockPaymentRepo) Update(payment *model.PaymentOrder) error {
	args := m.Called(payment)
	return args.Error(0)
}

func (m *mockPaymentRepo) UpdateState(paymentID uint, state model.PaymentState) error {
	args := m.Called(paymentID, state)
	return args.Error(0)
}

// ===== Payment Service Tests =====

func newTestPaymentService(paymentRepo *mockPaymentRepo, orderRepo *mockOrderRepo) PaymentService {
	return NewPaymentService(paymentRepo, orderRepo, zap.NewNop())
}

func TestPaymentService_PreparePayment_Success(t *testing.T) {
	paymentRepo := new(mockPaymentRepo)
	orderRepo := new(mockOrderRepo)
	svc := newTestPaymentService(paymentRepo, orderRepo)

	order := &model.PickupOrder{
		ID:          1,
		PassengerID: 1,
		Status:      model.OrderStatusCreated,
		PriceTotal:  5000,
		Currency:    "CNY",
	}
	orderRepo.On("GetByID", uint(1)).Return(order, nil).Once()
	paymentRepo.On("GetByOrderID", uint(1)).Return(nil, errors.New("not found")).Once()
	paymentRepo.On("Create", mock.AnythingOfType("*model.PaymentOrder")).Return(nil).Once()

	resp, err := svc.PreparePayment(1, 1)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.PrepayID)
	assert.NotEmpty(t, resp.PayParams)
	paymentRepo.AssertExpectations(t)
	orderRepo.AssertExpectations(t)
}

func TestPaymentService_PreparePayment_WrongUser(t *testing.T) {
	paymentRepo := new(mockPaymentRepo)
	orderRepo := new(mockOrderRepo)
	svc := newTestPaymentService(paymentRepo, orderRepo)

	order := &model.PickupOrder{ID: 1, PassengerID: 2, Status: model.OrderStatusCreated}
	orderRepo.On("GetByID", uint(1)).Return(order, nil).Once()

	resp, err := svc.PreparePayment(1, 1) // User 1 tries to pay for user 2's order
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "无权限")
}

func TestPaymentService_PreparePayment_WrongStatus(t *testing.T) {
	paymentRepo := new(mockPaymentRepo)
	orderRepo := new(mockOrderRepo)
	svc := newTestPaymentService(paymentRepo, orderRepo)

	order := &model.PickupOrder{ID: 1, PassengerID: 1, Status: model.OrderStatusPaid}
	orderRepo.On("GetByID", uint(1)).Return(order, nil).Once()

	resp, err := svc.PreparePayment(1, 1)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "订单状态不允许支付")
}

func TestPaymentService_PreparePayment_ExistingPending(t *testing.T) {
	paymentRepo := new(mockPaymentRepo)
	orderRepo := new(mockOrderRepo)
	svc := newTestPaymentService(paymentRepo, orderRepo)

	order := &model.PickupOrder{ID: 1, PassengerID: 1, Status: model.OrderStatusCreated}
	orderRepo.On("GetByID", uint(1)).Return(order, nil).Once()

	existingPayment := &model.PaymentOrder{
		ID:         1,
		OrderID:    1,
		State:      model.PaymentStatePending,
		WxPrepayID: "existing_prepay",
	}
	paymentRepo.On("GetByOrderID", uint(1)).Return(existingPayment, nil).Once()

	resp, err := svc.PreparePayment(1, 1)
	require.NoError(t, err)
	assert.Equal(t, "existing_prepay", resp.PrepayID)
}

func TestPaymentService_PreparePayment_AlreadyPaid(t *testing.T) {
	paymentRepo := new(mockPaymentRepo)
	orderRepo := new(mockOrderRepo)
	svc := newTestPaymentService(paymentRepo, orderRepo)

	order := &model.PickupOrder{ID: 1, PassengerID: 1, Status: model.OrderStatusCreated}
	orderRepo.On("GetByID", uint(1)).Return(order, nil).Once()

	existingPayment := &model.PaymentOrder{
		ID:      1,
		OrderID: 1,
		State:   model.PaymentStatePaid,
	}
	paymentRepo.On("GetByOrderID", uint(1)).Return(existingPayment, nil).Once()

	resp, err := svc.PreparePayment(1, 1)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "已支付")
}

func TestPaymentService_PreparePayment_OrderNotFound(t *testing.T) {
	paymentRepo := new(mockPaymentRepo)
	orderRepo := new(mockOrderRepo)
	svc := newTestPaymentService(paymentRepo, orderRepo)

	orderRepo.On("GetByID", uint(999)).Return(nil, errors.New("not found")).Once()

	resp, err := svc.PreparePayment(1, 999)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestPaymentService_HandlePaymentNotify_Success(t *testing.T) {
	paymentRepo := new(mockPaymentRepo)
	orderRepo := new(mockOrderRepo)
	svc := newTestPaymentService(paymentRepo, orderRepo)

	payment := &model.PaymentOrder{
		ID:      1,
		OrderID: 1,
		Amount:  5000,
		State:   model.PaymentStatePending,
	}
	paymentRepo.On("GetByTransactionID", "tx123").Return(nil, errors.New("not found")).Once()
	paymentRepo.On("GetByOrderID", uint(1)).Return(payment, nil).Once()
	paymentRepo.On("Update", mock.AnythingOfType("*model.PaymentOrder")).Return(nil).Once()
	orderRepo.On("UpdateStatus", uint(1), model.OrderStatusPaid).Return(nil).Once()

	req := &PaymentNotifyRequest{
		TransactionID: "tx123",
		OrderID:       1,
		Amount:        5000,
		Status:        "SUCCESS",
	}

	err := svc.HandlePaymentNotify(req)
	require.NoError(t, err)

	// Verify state was updated
	assert.Equal(t, "tx123", payment.WxTransactionID)
	assert.Equal(t, model.PaymentStatePaid, payment.State)
	assert.NotNil(t, payment.PaidAt)
	paymentRepo.AssertExpectations(t)
	orderRepo.AssertExpectations(t)
}

func TestPaymentService_HandlePaymentNotify_AlreadyPaid(t *testing.T) {
	paymentRepo := new(mockPaymentRepo)
	orderRepo := new(mockOrderRepo)
	svc := newTestPaymentService(paymentRepo, orderRepo)

	payment := &model.PaymentOrder{
		ID:      1,
		OrderID: 1,
		Amount:  5000,
		State:   model.PaymentStatePaid, // Already paid
	}
	paymentRepo.On("GetByTransactionID", "tx123").Return(payment, nil).Once()

	req := &PaymentNotifyRequest{
		TransactionID: "tx123",
		OrderID:       1,
		Amount:        5000,
		Status:        "SUCCESS",
	}

	err := svc.HandlePaymentNotify(req)
	require.NoError(t, err) // Idempotent - no error
	paymentRepo.AssertExpectations(t)
}

func TestPaymentService_HandlePaymentNotify_AmountMismatch(t *testing.T) {
	paymentRepo := new(mockPaymentRepo)
	orderRepo := new(mockOrderRepo)
	svc := newTestPaymentService(paymentRepo, orderRepo)

	payment := &model.PaymentOrder{
		ID:      1,
		OrderID: 1,
		Amount:  5000,
		State:   model.PaymentStatePending,
	}
	paymentRepo.On("GetByTransactionID", "tx123").Return(payment, nil).Once()

	req := &PaymentNotifyRequest{
		TransactionID: "tx123",
		OrderID:       1,
		Amount:        3000, // Wrong amount
		Status:        "SUCCESS",
	}

	err := svc.HandlePaymentNotify(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "金额不匹配")
}

func TestPaymentService_HandlePaymentNotify_PaymentNotFound(t *testing.T) {
	paymentRepo := new(mockPaymentRepo)
	orderRepo := new(mockOrderRepo)
	svc := newTestPaymentService(paymentRepo, orderRepo)

	paymentRepo.On("GetByTransactionID", "tx123").Return(nil, errors.New("not found")).Once()
	paymentRepo.On("GetByOrderID", uint(1)).Return(nil, errors.New("not found")).Once()

	req := &PaymentNotifyRequest{
		TransactionID: "tx123",
		OrderID:       1,
		Amount:        5000,
	}

	err := svc.HandlePaymentNotify(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "找不到支付订单")
}

func TestPaymentService_GetPaymentByOrderID_Success(t *testing.T) {
	paymentRepo := new(mockPaymentRepo)
	orderRepo := new(mockOrderRepo)
	svc := newTestPaymentService(paymentRepo, orderRepo)

	now := time.Now()
	expected := &model.PaymentOrder{ID: 1, OrderID: 1, Amount: 5000, PaidAt: &now}
	paymentRepo.On("GetByOrderID", uint(1)).Return(expected, nil).Once()

	payment, err := svc.GetPaymentByOrderID(1)
	require.NoError(t, err)
	assert.Equal(t, expected, payment)
}

func TestPaymentService_GetPaymentByOrderID_NotFound(t *testing.T) {
	paymentRepo := new(mockPaymentRepo)
	orderRepo := new(mockOrderRepo)
	svc := newTestPaymentService(paymentRepo, orderRepo)

	paymentRepo.On("GetByOrderID", uint(999)).Return(nil, errors.New("not found")).Once()

	payment, err := svc.GetPaymentByOrderID(999)
	assert.Error(t, err)
	assert.Nil(t, payment)
}
