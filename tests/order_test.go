package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pickup/internal/handler"
	"pickup/internal/model"
	"pickup/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockOrderService 模拟订单服务
type MockOrderService struct {
	mock.Mock
}

func (m *MockOrderService) CreateOrder(userID uint, req *service.CreateOrderRequest) (*model.PickupOrder, error) {
	args := m.Called(userID, req)
	var order *model.PickupOrder
	if v := args.Get(0); v != nil {
		order = v.(*model.PickupOrder)
	}
	return order, args.Error(1)
}

func (m *MockOrderService) GetOrder(id uint, userID uint) (*model.PickupOrder, error) {
	args := m.Called(id, userID)
	var order *model.PickupOrder
	if v := args.Get(0); v != nil {
		order = v.(*model.PickupOrder)
	}
	return order, args.Error(1)
}

func (m *MockOrderService) GetUserOrders(userID uint) ([]*model.PickupOrder, error) {
	args := m.Called(userID)
	if v := args.Get(0); v != nil {
		return v.([]*model.PickupOrder), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockOrderService) UpdateOrderStatus(orderID uint, status model.OrderStatus) error {
	args := m.Called(orderID, status)
	return args.Error(0)
}

func setupOrderRouter(service service.OrderService, withUser bool, role string) *gin.Engine {
	router := gin.New()
	api := router.Group("/api/v1")
	if withUser {
		api.Use(func(c *gin.Context) {
			c.Set("user_id", uint(1))
			if role != "" {
				c.Set("user_role", role)
			}
		})
	} else if role != "" {
		api.Use(func(c *gin.Context) {
			c.Set("user_role", role)
		})
	}
	handler.NewOrderHandler(service, zap.NewNop()).RegisterRoutes(api)
	return router
}

func TestCreateOrder_Success(t *testing.T) {
	mockService := new(MockOrderService)
	expected := &model.PickupOrder{ID: 1, PassengerID: 1, RegistrationID: 2, CreatedAt: time.Now()}
	mockService.On("CreateOrder", uint(1), mock.MatchedBy(func(req *service.CreateOrderRequest) bool {
		return req.RegistrationID == 2 && req.PriceTotal == 5000
	})).Return(expected, nil).Once()
	router := setupOrderRouter(mockService, true, "")

	body := map[string]interface{}{"registration_id": 2, "price_total": 5000, "currency": "CNY"}
	data, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestCreateOrder_Unauthorized(t *testing.T) {
	mockService := new(MockOrderService)
	router := setupOrderRouter(mockService, false, "")

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertExpectations(t)
}

func TestCreateOrder_InvalidBody(t *testing.T) {
	mockService := new(MockOrderService)
	router := setupOrderRouter(mockService, true, "")

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader([]byte(`{"registration_id":"bad"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestCreateOrder_ServiceError(t *testing.T) {
	mockService := new(MockOrderService)
	mockService.On("CreateOrder", uint(1), mock.Anything).Return((*model.PickupOrder)(nil), errors.New("failed")).Once()
	router := setupOrderRouter(mockService, true, "")

	body := map[string]interface{}{"registration_id": 2, "price_total": 5000}
	data, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetOrder_Success(t *testing.T) {
	mockService := new(MockOrderService)
	expected := &model.PickupOrder{ID: 1, PassengerID: 1}
	mockService.On("GetOrder", uint(1), uint(1)).Return(expected, nil).Once()
	router := setupOrderRouter(mockService, true, "")

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/orders/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetOrder_InvalidID(t *testing.T) {
	mockService := new(MockOrderService)
	router := setupOrderRouter(mockService, true, "")

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/orders/abc", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetOrder_ServiceError(t *testing.T) {
	mockService := new(MockOrderService)
	mockService.On("GetOrder", uint(1), uint(1)).Return((*model.PickupOrder)(nil), errors.New("not found")).Once()
	router := setupOrderRouter(mockService, true, "")

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/orders/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetOrder_Unauthorized(t *testing.T) {
	mockService := new(MockOrderService)
	router := setupOrderRouter(mockService, false, "")

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/orders/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetOrders_Success(t *testing.T) {
	mockService := new(MockOrderService)
	mockService.On("GetUserOrders", uint(1)).Return([]*model.PickupOrder{{ID: 1}}, nil).Once()
	router := setupOrderRouter(mockService, true, "")

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/orders", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetOrders_ServiceError(t *testing.T) {
	mockService := new(MockOrderService)
	mockService.On("GetUserOrders", uint(1)).Return(([]*model.PickupOrder)(nil), errors.New("failed")).Once()
	router := setupOrderRouter(mockService, true, "")

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/orders", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestNotifyOrder_Success(t *testing.T) {
	mockService := new(MockOrderService)
	mockService.On("UpdateOrderStatus", uint(1), model.OrderStatusNotified).Return(nil).Once()
	router := setupOrderRouter(mockService, true, string(model.RoleAdmin))

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/orders/1/notify", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestNotifyOrder_InvalidID(t *testing.T) {
	mockService := new(MockOrderService)
	router := setupOrderRouter(mockService, true, string(model.RoleAdmin))

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/orders/abc/notify", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestNotifyOrder_Forbidden(t *testing.T) {
	mockService := new(MockOrderService)
	router := setupOrderRouter(mockService, true, string(model.RolePassenger))

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/orders/1/notify", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	mockService.AssertExpectations(t)
}

func TestNotifyOrder_ServiceError(t *testing.T) {
	mockService := new(MockOrderService)
	mockService.On("UpdateOrderStatus", uint(1), model.OrderStatusNotified).Return(errors.New("failed")).Once()
	router := setupOrderRouter(mockService, true, string(model.RoleAdmin))

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/orders/1/notify", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}
