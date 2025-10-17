package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"pickup/internal/handler"
	"pickup/internal/model"
	"pickup/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockPaymentService 模拟支付服务
type MockPaymentService struct {
	mock.Mock
}

func (m *MockPaymentService) PreparePayment(userID uint, orderID uint) (*service.PreparePaymentResponse, error) {
	args := m.Called(userID, orderID)
	var resp *service.PreparePaymentResponse
	if v := args.Get(0); v != nil {
		resp = v.(*service.PreparePaymentResponse)
	}
	return resp, args.Error(1)
}

func (m *MockPaymentService) HandlePaymentNotify(req *service.PaymentNotifyRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockPaymentService) GetPaymentByOrderID(orderID uint) (*model.PaymentOrder, error) {
	args := m.Called(orderID)
	var payment *model.PaymentOrder
	if v := args.Get(0); v != nil {
		payment = v.(*model.PaymentOrder)
	}
	return payment, args.Error(1)
}

func setupPaymentRouter(service service.PaymentService, withUser bool) *gin.Engine {
	router := gin.New()
	api := router.Group("/api/v1")
	if withUser {
		api.Use(func(c *gin.Context) {
			c.Set("user_id", uint(1))
		})
	}
	handler.NewPaymentHandler(service, zap.NewNop()).RegisterRoutes(api)
	return router
}

func TestPreparePayment_Success(t *testing.T) {
	mockService := new(MockPaymentService)
	expected := &service.PreparePaymentResponse{PrepayID: "prepay", PayParams: map[string]string{"package": "prepay_id=prepay"}}
	mockService.On("PreparePayment", uint(1), uint(2)).Return(expected, nil).Once()
	router := setupPaymentRouter(mockService, true)

	body := map[string]uint{"order_id": 2}
	data, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/pay/prepare", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestPreparePayment_Unauthorized(t *testing.T) {
	mockService := new(MockPaymentService)
	router := setupPaymentRouter(mockService, false)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/pay/prepare", bytes.NewReader([]byte(`{"order_id":1}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertExpectations(t)
}

func TestPreparePayment_InvalidBody(t *testing.T) {
	mockService := new(MockPaymentService)
	router := setupPaymentRouter(mockService, true)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/pay/prepare", bytes.NewReader([]byte(`{"order_id":"bad"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestPreparePayment_ServiceError(t *testing.T) {
	mockService := new(MockPaymentService)
	mockService.On("PreparePayment", uint(1), uint(2)).Return((*service.PreparePaymentResponse)(nil), errors.New("failed")).Once()
	router := setupPaymentRouter(mockService, true)

	reqBody := map[string]uint{"order_id": 2}
	data, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/pay/prepare", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestPaymentNotify_Success(t *testing.T) {
	mockService := new(MockPaymentService)
	mockService.On("HandlePaymentNotify", mock.AnythingOfType("*service.PaymentNotifyRequest")).Return(nil).Once()
	router := setupPaymentRouter(mockService, false)

	reqBody := service.PaymentNotifyRequest{OrderID: 1, TransactionID: "tx", Amount: 100, Status: "SUCCESS"}
	data, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/pay/notify", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "SUCCESS", resp["code"])
	mockService.AssertExpectations(t)
}

func TestPaymentNotify_InvalidBody(t *testing.T) {
	mockService := new(MockPaymentService)
	router := setupPaymentRouter(mockService, false)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/pay/notify", bytes.NewReader([]byte(`{"order_id":"bad"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestPaymentNotify_ServiceError(t *testing.T) {
	mockService := new(MockPaymentService)
	mockService.On("HandlePaymentNotify", mock.AnythingOfType("*service.PaymentNotifyRequest")).Return(errors.New("failed")).Once()
	router := setupPaymentRouter(mockService, false)

	reqBody := service.PaymentNotifyRequest{OrderID: 1, TransactionID: "tx", Amount: 100, Status: "SUCCESS"}
	data, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/pay/notify", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}
