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

// ===== Mock Repositories =====

type mockOrderRepo struct {
	mock.Mock
}

func (m *mockOrderRepo) Create(order *model.PickupOrder) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *mockOrderRepo) GetByID(id uint) (*model.PickupOrder, error) {
	args := m.Called(id)
	if v := args.Get(0); v != nil {
		return v.(*model.PickupOrder), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockOrderRepo) GetByUserID(userID uint) ([]*model.PickupOrder, error) {
	args := m.Called(userID)
	if v := args.Get(0); v != nil {
		return v.([]*model.PickupOrder), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockOrderRepo) GetByRegistrationID(registrationID uint) (*model.PickupOrder, error) {
	args := m.Called(registrationID)
	if v := args.Get(0); v != nil {
		return v.(*model.PickupOrder), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockOrderRepo) Update(order *model.PickupOrder) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *mockOrderRepo) UpdateStatus(orderID uint, status model.OrderStatus) error {
	args := m.Called(orderID, status)
	return args.Error(0)
}

func (m *mockOrderRepo) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

type mockRegistrationRepo struct {
	mock.Mock
}

func (m *mockRegistrationRepo) Create(registration *model.Registration) error {
	args := m.Called(registration)
	return args.Error(0)
}

func (m *mockRegistrationRepo) GetByID(id uint) (*model.Registration, error) {
	args := m.Called(id)
	if v := args.Get(0); v != nil {
		return v.(*model.Registration), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRegistrationRepo) GetByUserID(userID uint) ([]*model.Registration, error) {
	args := m.Called(userID)
	if v := args.Get(0); v != nil {
		return v.([]*model.Registration), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRegistrationRepo) Update(registration *model.Registration) error {
	args := m.Called(registration)
	return args.Error(0)
}

func (m *mockRegistrationRepo) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

// ===== Order Service Tests =====

func newTestOrderService(orderRepo *mockOrderRepo, regRepo *mockRegistrationRepo) OrderService {
	return NewOrderService(orderRepo, regRepo, zap.NewNop())
}

func TestOrderService_CreateOrder_Success(t *testing.T) {
	orderRepo := new(mockOrderRepo)
	regRepo := new(mockRegistrationRepo)
	svc := newTestOrderService(orderRepo, regRepo)

	registration := &model.Registration{
		ID:          1,
		UserID:      1,
		ArrivalDate: time.Date(2024, 6, 15, 0, 0, 0, 0, time.Local),
		ArrivalTime: "14:30:00",
		Status:      model.RegistrationStatusDraft,
	}
	regRepo.On("GetByID", uint(1)).Return(registration, nil).Once()
	orderRepo.On("GetByRegistrationID", uint(1)).Return(nil, errors.New("not found")).Once()
	orderRepo.On("Create", mock.AnythingOfType("*model.PickupOrder")).Return(nil).Once()
	regRepo.On("Update", mock.AnythingOfType("*model.Registration")).Return(nil).Once()

	req := &CreateOrderRequest{
		RegistrationID: 1,
		PriceTotal:     5000,
		Currency:       "CNY",
	}

	order, err := svc.CreateOrder(1, req)
	require.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, uint(1), order.PassengerID)
	assert.Equal(t, uint(1), order.RegistrationID)
	assert.Equal(t, int64(5000), order.PriceTotal)
	assert.Equal(t, "CNY", order.Currency)
	assert.Equal(t, model.OrderStatusCreated, order.Status)
	orderRepo.AssertExpectations(t)
	regRepo.AssertExpectations(t)
}

func TestOrderService_CreateOrder_DefaultCurrency(t *testing.T) {
	orderRepo := new(mockOrderRepo)
	regRepo := new(mockRegistrationRepo)
	svc := newTestOrderService(orderRepo, regRepo)

	registration := &model.Registration{
		ID:          1,
		UserID:      1,
		ArrivalDate: time.Date(2024, 6, 15, 0, 0, 0, 0, time.Local),
		ArrivalTime: "14:30:00",
	}
	regRepo.On("GetByID", uint(1)).Return(registration, nil).Once()
	orderRepo.On("GetByRegistrationID", uint(1)).Return(nil, errors.New("not found")).Once()
	orderRepo.On("Create", mock.AnythingOfType("*model.PickupOrder")).Return(nil).Once()
	regRepo.On("Update", mock.AnythingOfType("*model.Registration")).Return(nil).Once()

	req := &CreateOrderRequest{
		RegistrationID: 1,
		PriceTotal:     3000,
		Currency:       "", // Empty should default to CNY
	}

	order, err := svc.CreateOrder(1, req)
	require.NoError(t, err)
	assert.Equal(t, "CNY", order.Currency)
}

func TestOrderService_CreateOrder_RegistrationNotFound(t *testing.T) {
	orderRepo := new(mockOrderRepo)
	regRepo := new(mockRegistrationRepo)
	svc := newTestOrderService(orderRepo, regRepo)

	regRepo.On("GetByID", uint(99)).Return(nil, errors.New("not found")).Once()

	req := &CreateOrderRequest{RegistrationID: 99, PriceTotal: 5000}
	order, err := svc.CreateOrder(1, req)
	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Contains(t, err.Error(), "获取报名信息失败")
}

func TestOrderService_CreateOrder_WrongUser(t *testing.T) {
	orderRepo := new(mockOrderRepo)
	regRepo := new(mockRegistrationRepo)
	svc := newTestOrderService(orderRepo, regRepo)

	registration := &model.Registration{ID: 1, UserID: 2} // Different user
	regRepo.On("GetByID", uint(1)).Return(registration, nil).Once()

	req := &CreateOrderRequest{RegistrationID: 1, PriceTotal: 5000}
	order, err := svc.CreateOrder(1, req) // User 1 tries to create for registration of user 2
	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Contains(t, err.Error(), "无权限")
}

func TestOrderService_CreateOrder_AlreadyExists(t *testing.T) {
	orderRepo := new(mockOrderRepo)
	regRepo := new(mockRegistrationRepo)
	svc := newTestOrderService(orderRepo, regRepo)

	registration := &model.Registration{ID: 1, UserID: 1}
	regRepo.On("GetByID", uint(1)).Return(registration, nil).Once()
	existingOrder := &model.PickupOrder{ID: 1}
	orderRepo.On("GetByRegistrationID", uint(1)).Return(existingOrder, nil).Once()

	req := &CreateOrderRequest{RegistrationID: 1, PriceTotal: 5000}
	order, err := svc.CreateOrder(1, req)
	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Contains(t, err.Error(), "已存在订单")
}

func TestOrderService_GetOrder_Success(t *testing.T) {
	orderRepo := new(mockOrderRepo)
	regRepo := new(mockRegistrationRepo)
	svc := newTestOrderService(orderRepo, regRepo)

	expected := &model.PickupOrder{ID: 1, PassengerID: 1}
	orderRepo.On("GetByID", uint(1)).Return(expected, nil).Once()

	order, err := svc.GetOrder(1, 1)
	require.NoError(t, err)
	assert.Equal(t, expected, order)
	orderRepo.AssertExpectations(t)
}

func TestOrderService_GetOrder_WrongUser(t *testing.T) {
	orderRepo := new(mockOrderRepo)
	regRepo := new(mockRegistrationRepo)
	svc := newTestOrderService(orderRepo, regRepo)

	existing := &model.PickupOrder{ID: 1, PassengerID: 2}
	orderRepo.On("GetByID", uint(1)).Return(existing, nil).Once()

	order, err := svc.GetOrder(1, 1) // User 1 tries to access user 2's order
	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Contains(t, err.Error(), "无权限")
}

func TestOrderService_GetOrder_NotFound(t *testing.T) {
	orderRepo := new(mockOrderRepo)
	regRepo := new(mockRegistrationRepo)
	svc := newTestOrderService(orderRepo, regRepo)

	orderRepo.On("GetByID", uint(999)).Return(nil, errors.New("not found")).Once()

	order, err := svc.GetOrder(999, 1)
	assert.Error(t, err)
	assert.Nil(t, order)
}

func TestOrderService_GetUserOrders_Success(t *testing.T) {
	orderRepo := new(mockOrderRepo)
	regRepo := new(mockRegistrationRepo)
	svc := newTestOrderService(orderRepo, regRepo)

	expected := []*model.PickupOrder{{ID: 1}, {ID: 2}}
	orderRepo.On("GetByUserID", uint(1)).Return(expected, nil).Once()

	orders, err := svc.GetUserOrders(1)
	require.NoError(t, err)
	assert.Len(t, orders, 2)
}

func TestOrderService_GetUserOrders_Error(t *testing.T) {
	orderRepo := new(mockOrderRepo)
	regRepo := new(mockRegistrationRepo)
	svc := newTestOrderService(orderRepo, regRepo)

	orderRepo.On("GetByUserID", uint(1)).Return(nil, errors.New("db error")).Once()

	orders, err := svc.GetUserOrders(1)
	assert.Error(t, err)
	assert.Nil(t, orders)
}

func TestOrderService_UpdateOrderStatus_Success(t *testing.T) {
	orderRepo := new(mockOrderRepo)
	regRepo := new(mockRegistrationRepo)
	svc := newTestOrderService(orderRepo, regRepo)

	// Order is in "created" status, transitioning to "paid" (valid)
	order := &model.PickupOrder{ID: 1, Status: model.OrderStatusCreated}
	orderRepo.On("GetByID", uint(1)).Return(order, nil).Once()
	orderRepo.On("UpdateStatus", uint(1), model.OrderStatusPaid).Return(nil).Once()

	err := svc.UpdateOrderStatus(1, model.OrderStatusPaid)
	require.NoError(t, err)
	orderRepo.AssertExpectations(t)
}

func TestOrderService_UpdateOrderStatus_InvalidTransition(t *testing.T) {
	orderRepo := new(mockOrderRepo)
	regRepo := new(mockRegistrationRepo)
	svc := newTestOrderService(orderRepo, regRepo)

	// Order is "completed", cannot transition to "created"
	order := &model.PickupOrder{ID: 1, Status: model.OrderStatusCompleted}
	orderRepo.On("GetByID", uint(1)).Return(order, nil).Once()

	err := svc.UpdateOrderStatus(1, model.OrderStatusCreated)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "不允许")
}

func TestOrderService_UpdateOrderStatus_Error(t *testing.T) {
	orderRepo := new(mockOrderRepo)
	regRepo := new(mockRegistrationRepo)
	svc := newTestOrderService(orderRepo, regRepo)

	order := &model.PickupOrder{ID: 1, Status: model.OrderStatusCreated}
	orderRepo.On("GetByID", uint(1)).Return(order, nil).Once()
	orderRepo.On("UpdateStatus", uint(1), model.OrderStatusPaid).Return(errors.New("db error")).Once()

	err := svc.UpdateOrderStatus(1, model.OrderStatusPaid)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "更新订单状态失败")
}

func TestOrderService_UpdateOrderStatus_OrderNotFound(t *testing.T) {
	orderRepo := new(mockOrderRepo)
	regRepo := new(mockRegistrationRepo)
	svc := newTestOrderService(orderRepo, regRepo)

	orderRepo.On("GetByID", uint(1)).Return(nil, errors.New("not found")).Once()

	err := svc.UpdateOrderStatus(1, model.OrderStatusPaid)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "获取订单失败")
}

func TestOrderService_CreateOrder_NegativePrice(t *testing.T) {
	orderRepo := new(mockOrderRepo)
	regRepo := new(mockRegistrationRepo)
	svc := newTestOrderService(orderRepo, regRepo)

	order, err := svc.CreateOrder(1, &CreateOrderRequest{
		RegistrationID: 1,
		PriceTotal:     -100,
		Currency:       "CNY",
	})
	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Contains(t, err.Error(), "负数")
}

func TestIsValidTransition(t *testing.T) {
	tests := []struct {
		from  model.OrderStatus
		to    model.OrderStatus
		valid bool
	}{
		{model.OrderStatusCreated, model.OrderStatusPaid, true},
		{model.OrderStatusCreated, model.OrderStatusCanceled, true},
		{model.OrderStatusCreated, model.OrderStatusCompleted, false},
		{model.OrderStatusPaid, model.OrderStatusAssigned, true},
		{model.OrderStatusPaid, model.OrderStatusCanceled, true},
		{model.OrderStatusPaid, model.OrderStatusCreated, false},
		{model.OrderStatusCompleted, model.OrderStatusCanceled, false},
		{model.OrderStatusCanceled, model.OrderStatusPaid, false},
	}

	for _, tc := range tests {
		t.Run(string(tc.from)+"->"+string(tc.to), func(t *testing.T) {
			assert.Equal(t, tc.valid, isValidTransition(tc.from, tc.to))
		})
	}
}
