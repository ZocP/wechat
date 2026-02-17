package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pickup/internal/config"
	"pickup/internal/model"
	schedulercontrollers "pickup/internal/scheduler/controllers"
	"pickup/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ===== Mock Services =====

type mockAuthService struct {
	mock.Mock
}

func (m *mockAuthService) WechatLogin(code, phoneCode string) (*service.WechatLoginResponse, error) {
	args := m.Called(code, phoneCode)
	if v := args.Get(0); v != nil {
		return v.(*service.WechatLoginResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAuthService) GetUserInfo(userID uint) (*model.User, error) {
	args := m.Called(userID)
	if v := args.Get(0); v != nil {
		return v.(*model.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAuthService) UpdateLastLogin(userID uint) error {
	args := m.Called(userID)
	return args.Error(0)
}

type mockNoticeService struct {
	mock.Mock
}

func (m *mockNoticeService) CreateNotice(userID uint, req *service.CreateNoticeRequest) (*model.Notice, error) {
	args := m.Called(userID, req)
	if v := args.Get(0); v != nil {
		return v.(*model.Notice), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockNoticeService) GetNotice(id uint) (*model.Notice, error) {
	args := m.Called(id)
	if v := args.Get(0); v != nil {
		return v.(*model.Notice), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockNoticeService) GetVisibleNotices() ([]*model.Notice, error) {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*model.Notice), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockNoticeService) GetNoticesByFlightNo(flightNo string) ([]*model.Notice, error) {
	args := m.Called(flightNo)
	if v := args.Get(0); v != nil {
		return v.([]*model.Notice), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockNoticeService) UpdateNotice(id uint, userID uint, req *service.UpdateNoticeRequest) (*model.Notice, error) {
	args := m.Called(id, userID, req)
	if v := args.Get(0); v != nil {
		return v.(*model.Notice), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockNoticeService) DeleteNotice(id uint, userID uint) error {
	args := m.Called(id, userID)
	return args.Error(0)
}

type mockOrderService struct {
	mock.Mock
}

func (m *mockOrderService) CreateOrder(userID uint, req *service.CreateOrderRequest) (*model.PickupOrder, error) {
	args := m.Called(userID, req)
	if v := args.Get(0); v != nil {
		return v.(*model.PickupOrder), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockOrderService) GetOrder(id uint, userID uint) (*model.PickupOrder, error) {
	args := m.Called(id, userID)
	if v := args.Get(0); v != nil {
		return v.(*model.PickupOrder), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockOrderService) GetUserOrders(userID uint) ([]*model.PickupOrder, error) {
	args := m.Called(userID)
	if v := args.Get(0); v != nil {
		return v.([]*model.PickupOrder), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockOrderService) UpdateOrderStatus(orderID uint, status model.OrderStatus) error {
	args := m.Called(orderID, status)
	return args.Error(0)
}

type mockPaymentService struct {
	mock.Mock
}

func (m *mockPaymentService) PreparePayment(userID uint, orderID uint) (*service.PreparePaymentResponse, error) {
	args := m.Called(userID, orderID)
	if v := args.Get(0); v != nil {
		return v.(*service.PreparePaymentResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockPaymentService) HandlePaymentNotify(req *service.PaymentNotifyRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *mockPaymentService) GetPaymentByOrderID(orderID uint) (*model.PaymentOrder, error) {
	args := m.Called(orderID)
	if v := args.Get(0); v != nil {
		return v.(*model.PaymentOrder), args.Error(1)
	}
	return nil, args.Error(1)
}

type mockRegistrationService struct {
	mock.Mock
}

func (m *mockRegistrationService) CreateRegistration(userID uint, req *service.CreateRegistrationRequest) (*model.Registration, error) {
	args := m.Called(userID, req)
	if v := args.Get(0); v != nil {
		return v.(*model.Registration), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRegistrationService) GetRegistration(id uint, userID uint) (*model.Registration, error) {
	args := m.Called(id, userID)
	if v := args.Get(0); v != nil {
		return v.(*model.Registration), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRegistrationService) UpdateRegistration(id uint, userID uint, req *service.UpdateRegistrationRequest) (*model.Registration, error) {
	args := m.Called(id, userID, req)
	if v := args.Get(0); v != nil {
		return v.(*model.Registration), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRegistrationService) GetUserRegistrations(userID uint) ([]*model.Registration, error) {
	args := m.Called(userID)
	if v := args.Get(0); v != nil {
		return v.([]*model.Registration), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRegistrationService) DeleteRegistration(id uint, userID uint) error {
	args := m.Called(id, userID)
	return args.Error(0)
}

type mockSchemaService struct {
	mock.Mock
}

func (m *mockSchemaService) GetDatabaseSchema() ([]service.TableSchema, error) {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]service.TableSchema), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockSchemaService) ExportSchemaJSON() ([]byte, error) {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]byte), args.Error(1)
	}
	return nil, args.Error(1)
}

// ===== Helper =====

func setUserContext(c *gin.Context, userID uint, role string) {
	c.Set("user_id", userID)
	c.Set("user_role", role)
}

func newTestLogger() *zap.Logger {
	return zap.NewNop()
}

// ===== Auth Handler Tests =====

func TestAuthHandler_WechatLogin_Success(t *testing.T) {
	svc := new(mockAuthService)
	h := NewAuthHandler(svc, newTestLogger())

	resp := &service.WechatLoginResponse{
		Token: "jwt-token",
		User:  &model.User{ID: 1, Nickname: "test"},
	}
	svc.On("WechatLogin", "code123", "phone_code").Return(resp, nil).Once()

	router := gin.New()
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	body := `{"code":"code123","phone_code":"phone_code"}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/wechat/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestAuthHandler_WechatLogin_InvalidParams(t *testing.T) {
	svc := new(mockAuthService)
	h := NewAuthHandler(svc, newTestLogger())

	router := gin.New()
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	body := `{"code":""}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/wechat/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_WechatLogin_ServiceError(t *testing.T) {
	svc := new(mockAuthService)
	h := NewAuthHandler(svc, newTestLogger())

	svc.On("WechatLogin", "code123", "phone_code").Return(nil, errors.New("wechat error")).Once()

	router := gin.New()
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	body := `{"code":"code123","phone_code":"phone_code"}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/wechat/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

func TestAuthHandler_GetMe_Success(t *testing.T) {
	svc := new(mockAuthService)
	h := NewAuthHandler(svc, newTestLogger())

	user := &model.User{ID: 1, Nickname: "test"}
	svc.On("GetUserInfo", uint(1)).Return(user, nil).Once()

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "passenger"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestAuthHandler_GetMe_Unauthorized(t *testing.T) {
	svc := new(mockAuthService)
	h := NewAuthHandler(svc, newTestLogger())

	router := gin.New()
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_GetMe_ServiceError(t *testing.T) {
	svc := new(mockAuthService)
	h := NewAuthHandler(svc, newTestLogger())

	svc.On("GetUserInfo", uint(1)).Return(nil, errors.New("db error")).Once()

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "passenger"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

// ===== Notice Handler Tests =====

func TestNoticeHandler_GetVisibleNotices_Success(t *testing.T) {
	svc := new(mockNoticeService)
	h := NewNoticeHandler(svc, newTestLogger())

	notices := []*model.Notice{{ID: 1}, {ID: 2}}
	svc.On("GetVisibleNotices").Return(notices, nil).Once()

	router := gin.New()
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/notices", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestNoticeHandler_GetVisibleNotices_Error(t *testing.T) {
	svc := new(mockNoticeService)
	h := NewNoticeHandler(svc, newTestLogger())

	svc.On("GetVisibleNotices").Return(nil, errors.New("db error")).Once()

	router := gin.New()
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/notices", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

func TestNoticeHandler_GetNotice_Success(t *testing.T) {
	svc := new(mockNoticeService)
	h := NewNoticeHandler(svc, newTestLogger())

	notice := &model.Notice{ID: 1, FlightNo: "CA1234"}
	svc.On("GetNotice", uint(1)).Return(notice, nil).Once()

	router := gin.New()
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/notices/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestNoticeHandler_GetNotice_InvalidID(t *testing.T) {
	svc := new(mockNoticeService)
	h := NewNoticeHandler(svc, newTestLogger())

	router := gin.New()
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/notices/abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestNoticeHandler_GetNotice_NotFound(t *testing.T) {
	svc := new(mockNoticeService)
	h := NewNoticeHandler(svc, newTestLogger())

	svc.On("GetNotice", uint(999)).Return(nil, errors.New("not found")).Once()

	router := gin.New()
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/notices/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	svc.AssertExpectations(t)
}

func TestNoticeHandler_GetNoticesByFlightNo_Success(t *testing.T) {
	svc := new(mockNoticeService)
	h := NewNoticeHandler(svc, newTestLogger())

	notices := []*model.Notice{{ID: 1, FlightNo: "CA1234"}}
	svc.On("GetNoticesByFlightNo", "CA1234").Return(notices, nil).Once()

	router := gin.New()
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/notices/flight/CA1234", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestNoticeHandler_GetNoticesByFlightNo_Error(t *testing.T) {
	svc := new(mockNoticeService)
	h := NewNoticeHandler(svc, newTestLogger())

	svc.On("GetNoticesByFlightNo", "CA1234").Return(nil, errors.New("db error")).Once()

	router := gin.New()
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/notices/flight/CA1234", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

func TestNoticeHandler_CreateNotice_Success(t *testing.T) {
	svc := new(mockNoticeService)
	h := NewNoticeHandler(svc, newTestLogger())

	notice := &model.Notice{ID: 1, FlightNo: "CA1234"}
	svc.On("CreateNotice", uint(1), mock.AnythingOfType("*service.CreateNoticeRequest")).Return(notice, nil).Once()

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "admin"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	body := `{"flight_no":"CA1234","terminal":"T1","pickup_batch":"B1","arrival_airport":"PEK","meeting_point":"Gate A","visible_from":"2024-06-15T10:00:00Z","visible_to":"2024-06-15T12:00:00Z"}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/notices", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	svc.AssertExpectations(t)
}

func TestNoticeHandler_CreateNotice_Unauthorized(t *testing.T) {
	svc := new(mockNoticeService)
	h := NewNoticeHandler(svc, newTestLogger())

	router := gin.New()
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	body := `{"flight_no":"CA1234"}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/notices", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestNoticeHandler_CreateNotice_InvalidParams(t *testing.T) {
	svc := new(mockNoticeService)
	h := NewNoticeHandler(svc, newTestLogger())

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "admin"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	body := `{"flight_no":""}` // missing required fields
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/notices", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestNoticeHandler_CreateNotice_ServiceError(t *testing.T) {
	svc := new(mockNoticeService)
	h := NewNoticeHandler(svc, newTestLogger())

	svc.On("CreateNotice", uint(1), mock.Anything).Return(nil, errors.New("create error")).Once()

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "admin"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	body := `{"flight_no":"CA1234","terminal":"T1","pickup_batch":"B1","arrival_airport":"PEK","meeting_point":"Gate A","visible_from":"2024-06-15T10:00:00Z","visible_to":"2024-06-15T12:00:00Z"}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/notices", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

func TestNoticeHandler_UpdateNotice_Success(t *testing.T) {
	svc := new(mockNoticeService)
	h := NewNoticeHandler(svc, newTestLogger())

	notice := &model.Notice{ID: 1, FlightNo: "CA1234", Terminal: "T2"}
	svc.On("UpdateNotice", uint(1), uint(1), mock.AnythingOfType("*service.UpdateNoticeRequest")).Return(notice, nil).Once()

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "admin"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	body := `{"terminal":"T2"}`
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/admin/notices/1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestNoticeHandler_UpdateNotice_Unauthorized(t *testing.T) {
	svc := new(mockNoticeService)
	h := NewNoticeHandler(svc, newTestLogger())

	router := gin.New()
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	body := `{"terminal":"T2"}`
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/admin/notices/1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestNoticeHandler_UpdateNotice_InvalidID(t *testing.T) {
	svc := new(mockNoticeService)
	h := NewNoticeHandler(svc, newTestLogger())

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "admin"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	body := `{"terminal":"T2"}`
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/admin/notices/abc", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestNoticeHandler_UpdateNotice_InvalidBody(t *testing.T) {
	svc := new(mockNoticeService)
	h := NewNoticeHandler(svc, newTestLogger())

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "admin"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	body := `not json`
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/admin/notices/1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestNoticeHandler_UpdateNotice_ServiceError(t *testing.T) {
	svc := new(mockNoticeService)
	h := NewNoticeHandler(svc, newTestLogger())

	svc.On("UpdateNotice", uint(1), uint(1), mock.Anything).Return(nil, errors.New("update error")).Once()

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "admin"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	body := `{"terminal":"T2"}`
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/admin/notices/1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

func TestNoticeHandler_DeleteNotice_Success(t *testing.T) {
	svc := new(mockNoticeService)
	h := NewNoticeHandler(svc, newTestLogger())

	svc.On("DeleteNotice", uint(1), uint(1)).Return(nil).Once()

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "admin"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/admin/notices/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestNoticeHandler_DeleteNotice_Unauthorized(t *testing.T) {
	svc := new(mockNoticeService)
	h := NewNoticeHandler(svc, newTestLogger())

	router := gin.New()
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/admin/notices/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestNoticeHandler_DeleteNotice_InvalidID(t *testing.T) {
	svc := new(mockNoticeService)
	h := NewNoticeHandler(svc, newTestLogger())

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "admin"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/admin/notices/xyz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestNoticeHandler_DeleteNotice_ServiceError(t *testing.T) {
	svc := new(mockNoticeService)
	h := NewNoticeHandler(svc, newTestLogger())

	svc.On("DeleteNotice", uint(1), uint(1)).Return(errors.New("delete error")).Once()

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "admin"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/admin/notices/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

// ===== Order Handler Tests =====

func TestOrderHandler_CreateOrder_Success(t *testing.T) {
	svc := new(mockOrderService)
	h := NewOrderHandler(svc, newTestLogger())

	order := &model.PickupOrder{ID: 1, PassengerID: 1, PriceTotal: 5000}
	svc.On("CreateOrder", uint(1), mock.AnythingOfType("*service.CreateOrderRequest")).Return(order, nil).Once()

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "passenger"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	body := `{"registration_id":1,"price_total":5000,"currency":"CNY"}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	svc.AssertExpectations(t)
}

func TestOrderHandler_CreateOrder_Unauthorized(t *testing.T) {
	svc := new(mockOrderService)
	h := NewOrderHandler(svc, newTestLogger())

	router := gin.New()
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	body := `{"registration_id":1,"price_total":5000}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestOrderHandler_CreateOrder_InvalidParams(t *testing.T) {
	svc := new(mockOrderService)
	h := NewOrderHandler(svc, newTestLogger())

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "passenger"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	body := `{"invalid":"json"}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrderHandler_CreateOrder_ServiceError(t *testing.T) {
	svc := new(mockOrderService)
	h := NewOrderHandler(svc, newTestLogger())

	svc.On("CreateOrder", uint(1), mock.Anything).Return(nil, errors.New("service error")).Once()

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "passenger"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	body := `{"registration_id":1,"price_total":5000,"currency":"CNY"}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

func TestOrderHandler_GetOrder_Success(t *testing.T) {
	svc := new(mockOrderService)
	h := NewOrderHandler(svc, newTestLogger())

	order := &model.PickupOrder{ID: 1, PassengerID: 1}
	svc.On("GetOrder", uint(1), uint(1)).Return(order, nil).Once()

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "passenger"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/orders/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestOrderHandler_GetOrder_Unauthorized(t *testing.T) {
	svc := new(mockOrderService)
	h := NewOrderHandler(svc, newTestLogger())

	router := gin.New()
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/orders/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestOrderHandler_GetOrder_InvalidID(t *testing.T) {
	svc := new(mockOrderService)
	h := NewOrderHandler(svc, newTestLogger())

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "passenger"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/orders/abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrderHandler_GetOrder_NotFound(t *testing.T) {
	svc := new(mockOrderService)
	h := NewOrderHandler(svc, newTestLogger())

	svc.On("GetOrder", uint(999), uint(1)).Return(nil, errors.New("not found")).Once()

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "passenger"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/orders/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	svc.AssertExpectations(t)
}

func TestOrderHandler_GetMyOrders_Success(t *testing.T) {
	svc := new(mockOrderService)
	h := NewOrderHandler(svc, newTestLogger())

	orders := []*model.PickupOrder{{ID: 1}, {ID: 2}}
	svc.On("GetUserOrders", uint(1)).Return(orders, nil).Once()

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "passenger"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/orders", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestOrderHandler_GetMyOrders_Unauthorized(t *testing.T) {
	svc := new(mockOrderService)
	h := NewOrderHandler(svc, newTestLogger())

	router := gin.New()
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/orders", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestOrderHandler_GetMyOrders_ServiceError(t *testing.T) {
	svc := new(mockOrderService)
	h := NewOrderHandler(svc, newTestLogger())

	svc.On("GetUserOrders", uint(1)).Return(nil, errors.New("db error")).Once()

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "passenger"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/orders", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

func TestOrderHandler_NotifyOrder_Success(t *testing.T) {
	svc := new(mockOrderService)
	h := NewOrderHandler(svc, newTestLogger())

	svc.On("UpdateOrderStatus", uint(1), model.OrderStatusNotified).Return(nil).Once()

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "admin"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/orders/1/notify", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestOrderHandler_NotifyOrder_Forbidden(t *testing.T) {
	svc := new(mockOrderService)
	h := NewOrderHandler(svc, newTestLogger())

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "passenger"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/orders/1/notify", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestOrderHandler_NotifyOrder_InvalidID(t *testing.T) {
	svc := new(mockOrderService)
	h := NewOrderHandler(svc, newTestLogger())

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "admin"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/orders/abc/notify", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrderHandler_NotifyOrder_ServiceError(t *testing.T) {
	svc := new(mockOrderService)
	h := NewOrderHandler(svc, newTestLogger())

	svc.On("UpdateOrderStatus", uint(1), model.OrderStatusNotified).Return(errors.New("update error")).Once()

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "admin"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/orders/1/notify", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

func TestOrderHandler_NotifyOrder_DispatcherAllowed(t *testing.T) {
	svc := new(mockOrderService)
	h := NewOrderHandler(svc, newTestLogger())

	svc.On("UpdateOrderStatus", uint(1), model.OrderStatusNotified).Return(nil).Once()

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "dispatcher"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/orders/1/notify", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

// ===== Payment Handler Tests =====

func TestPaymentHandler_PreparePayment_Success(t *testing.T) {
	svc := new(mockPaymentService)
	h := NewPaymentHandler(svc, newTestLogger())

	resp := &service.PreparePaymentResponse{
		PrepayID:  "prepay_123",
		PayParams: map[string]string{"appId": "wx123"},
	}
	svc.On("PreparePayment", uint(1), uint(1)).Return(resp, nil).Once()

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "passenger"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	body := `{"order_id":1}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/pay/prepare", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestPaymentHandler_PreparePayment_Unauthorized(t *testing.T) {
	svc := new(mockPaymentService)
	h := NewPaymentHandler(svc, newTestLogger())

	router := gin.New()
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	body := `{"order_id":1}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/pay/prepare", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPaymentHandler_PreparePayment_InvalidParams(t *testing.T) {
	svc := new(mockPaymentService)
	h := NewPaymentHandler(svc, newTestLogger())

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "passenger"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	body := `{}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/pay/prepare", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPaymentHandler_PreparePayment_ServiceError(t *testing.T) {
	svc := new(mockPaymentService)
	h := NewPaymentHandler(svc, newTestLogger())

	svc.On("PreparePayment", uint(1), uint(1)).Return(nil, errors.New("pay error")).Once()

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "passenger"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	body := `{"order_id":1}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/pay/prepare", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

func TestPaymentHandler_PaymentNotify_Success(t *testing.T) {
	svc := new(mockPaymentService)
	h := NewPaymentHandler(svc, newTestLogger())

	svc.On("HandlePaymentNotify", mock.AnythingOfType("*service.PaymentNotifyRequest")).Return(nil).Once()

	router := gin.New()
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	body := `{"transaction_id":"tx123","order_id":1,"amount":5000,"status":"SUCCESS"}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/pay/notify", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "SUCCESS", resp["code"])
	svc.AssertExpectations(t)
}

func TestPaymentHandler_PaymentNotify_InvalidBody(t *testing.T) {
	svc := new(mockPaymentService)
	h := NewPaymentHandler(svc, newTestLogger())

	router := gin.New()
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	body := `not json`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/pay/notify", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPaymentHandler_PaymentNotify_ServiceError(t *testing.T) {
	svc := new(mockPaymentService)
	h := NewPaymentHandler(svc, newTestLogger())

	svc.On("HandlePaymentNotify", mock.Anything).Return(errors.New("notify error")).Once()

	router := gin.New()
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	body := `{"transaction_id":"tx123","order_id":1,"amount":5000,"status":"SUCCESS"}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/pay/notify", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

// ===== Registration Handler Tests =====

func TestRegistrationHandler_CreateRegistration_Success(t *testing.T) {
	svc := new(mockRegistrationService)
	h := NewRegistrationHandler(svc, newTestLogger())

	reg := &model.Registration{ID: 1, Name: "张三", UserID: 1}
	svc.On("CreateRegistration", uint(1), mock.AnythingOfType("*service.CreateRegistrationRequest")).Return(reg, nil).Once()

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "passenger"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	body := `{"name":"张三","phone":"13800138000","flight_no":"CA1234","arrival_date":"2024-06-15","arrival_time":"14:30","departure_city":"北京","pickup_method":"group"}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/registrations", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	svc.AssertExpectations(t)
}

func TestRegistrationHandler_CreateRegistration_Unauthorized(t *testing.T) {
	svc := new(mockRegistrationService)
	h := NewRegistrationHandler(svc, newTestLogger())

	router := gin.New()
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	body := `{"name":"张三"}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/registrations", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRegistrationHandler_CreateRegistration_InvalidParams(t *testing.T) {
	svc := new(mockRegistrationService)
	h := NewRegistrationHandler(svc, newTestLogger())

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "passenger"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	body := `{"name":""}` // missing required fields
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/registrations", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegistrationHandler_CreateRegistration_ServiceError(t *testing.T) {
	svc := new(mockRegistrationService)
	h := NewRegistrationHandler(svc, newTestLogger())

	svc.On("CreateRegistration", uint(1), mock.Anything).Return(nil, errors.New("create error")).Once()

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "passenger"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	body := `{"name":"张三","phone":"13800138000","flight_no":"CA1234","arrival_date":"2024-06-15","arrival_time":"14:30","departure_city":"北京","pickup_method":"group"}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/registrations", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

func TestRegistrationHandler_GetRegistration_Success(t *testing.T) {
	svc := new(mockRegistrationService)
	h := NewRegistrationHandler(svc, newTestLogger())

	reg := &model.Registration{ID: 1, UserID: 1}
	svc.On("GetRegistration", uint(1), uint(1)).Return(reg, nil).Once()

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "passenger"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/registrations/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestRegistrationHandler_GetRegistration_Unauthorized(t *testing.T) {
	svc := new(mockRegistrationService)
	h := NewRegistrationHandler(svc, newTestLogger())

	router := gin.New()
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/registrations/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRegistrationHandler_GetRegistration_InvalidID(t *testing.T) {
	svc := new(mockRegistrationService)
	h := NewRegistrationHandler(svc, newTestLogger())

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "passenger"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/registrations/abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegistrationHandler_GetRegistration_NotFound(t *testing.T) {
	svc := new(mockRegistrationService)
	h := NewRegistrationHandler(svc, newTestLogger())

	svc.On("GetRegistration", uint(999), uint(1)).Return(nil, errors.New("not found")).Once()

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "passenger"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/registrations/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	svc.AssertExpectations(t)
}

func TestRegistrationHandler_GetMyRegistrations_Success(t *testing.T) {
	svc := new(mockRegistrationService)
	h := NewRegistrationHandler(svc, newTestLogger())

	regs := []*model.Registration{{ID: 1}, {ID: 2}}
	svc.On("GetUserRegistrations", uint(1)).Return(regs, nil).Once()

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "passenger"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/registrations/my", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestRegistrationHandler_GetMyRegistrations_Unauthorized(t *testing.T) {
	svc := new(mockRegistrationService)
	h := NewRegistrationHandler(svc, newTestLogger())

	router := gin.New()
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/registrations/my", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRegistrationHandler_GetMyRegistrations_Error(t *testing.T) {
	svc := new(mockRegistrationService)
	h := NewRegistrationHandler(svc, newTestLogger())

	svc.On("GetUserRegistrations", uint(1)).Return(nil, errors.New("db error")).Once()

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "passenger"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/registrations/my", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

func TestRegistrationHandler_UpdateRegistration_Success(t *testing.T) {
	svc := new(mockRegistrationService)
	h := NewRegistrationHandler(svc, newTestLogger())

	reg := &model.Registration{ID: 1, UserID: 1, Name: "李四"}
	svc.On("UpdateRegistration", uint(1), uint(1), mock.AnythingOfType("*service.UpdateRegistrationRequest")).Return(reg, nil).Once()

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "passenger"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	body := `{"name":"李四"}`
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/registrations/1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestRegistrationHandler_UpdateRegistration_Unauthorized(t *testing.T) {
	svc := new(mockRegistrationService)
	h := NewRegistrationHandler(svc, newTestLogger())

	router := gin.New()
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	body := `{"name":"李四"}`
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/registrations/1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRegistrationHandler_UpdateRegistration_InvalidID(t *testing.T) {
	svc := new(mockRegistrationService)
	h := NewRegistrationHandler(svc, newTestLogger())

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "passenger"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	body := `{"name":"李四"}`
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/registrations/abc", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegistrationHandler_UpdateRegistration_InvalidBody(t *testing.T) {
	svc := new(mockRegistrationService)
	h := NewRegistrationHandler(svc, newTestLogger())

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "passenger"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	body := `not json`
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/registrations/1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegistrationHandler_UpdateRegistration_ServiceError(t *testing.T) {
	svc := new(mockRegistrationService)
	h := NewRegistrationHandler(svc, newTestLogger())

	svc.On("UpdateRegistration", uint(1), uint(1), mock.Anything).Return(nil, errors.New("update error")).Once()

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "passenger"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	body := `{"name":"李四"}`
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/registrations/1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

func TestRegistrationHandler_DeleteRegistration_Success(t *testing.T) {
	svc := new(mockRegistrationService)
	h := NewRegistrationHandler(svc, newTestLogger())

	svc.On("DeleteRegistration", uint(1), uint(1)).Return(nil).Once()

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "passenger"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/registrations/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestRegistrationHandler_DeleteRegistration_Unauthorized(t *testing.T) {
	svc := new(mockRegistrationService)
	h := NewRegistrationHandler(svc, newTestLogger())

	router := gin.New()
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/registrations/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRegistrationHandler_DeleteRegistration_InvalidID(t *testing.T) {
	svc := new(mockRegistrationService)
	h := NewRegistrationHandler(svc, newTestLogger())

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "passenger"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/registrations/abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegistrationHandler_DeleteRegistration_ServiceError(t *testing.T) {
	svc := new(mockRegistrationService)
	h := NewRegistrationHandler(svc, newTestLogger())

	svc.On("DeleteRegistration", uint(1), uint(1)).Return(errors.New("delete error")).Once()

	router := gin.New()
	router.Use(func(c *gin.Context) { setUserContext(c, 1, "passenger"); c.Next() })
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/registrations/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

// ===== Admin Handler Tests =====

func TestAdminHandler_ExportDatabaseFields_Success(t *testing.T) {
	svc := new(mockSchemaService)
	h := NewAdminHandler(svc, newTestLogger())

	jsonData := []byte(`[{"table_name":"users","columns":[]}]`)
	svc.On("ExportSchemaJSON").Return(jsonData, nil).Once()

	router := gin.New()
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/exports/database-fields", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Header().Get("Content-Disposition"), "database_fields.json")
	svc.AssertExpectations(t)
}

func TestAdminHandler_ExportDatabaseFields_Error(t *testing.T) {
	svc := new(mockSchemaService)
	h := NewAdminHandler(svc, newTestLogger())

	svc.On("ExportSchemaJSON").Return(nil, errors.New("export error")).Once()

	router := gin.New()
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/exports/database-fields", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

// ===== Router Config Tests =====

func TestNewRouterConfig(t *testing.T) {
	authCtl := schedulercontrollers.NewAuthController(nil)
	studentCtl := schedulercontrollers.NewStudentController(nil)
	adminCtl := schedulercontrollers.NewAdminController(nil)

	jwtCfg := &config.JWTConfig{
		Secret:     "test-secret",
		ExpireTime: 24 * time.Hour,
		Issuer:     "test",
	}

	rc := NewRouterConfig(authCtl, studentCtl, adminCtl, jwtCfg)
	require.NotNil(t, rc)
	assert.Equal(t, authCtl, rc.AuthController)
	assert.Equal(t, studentCtl, rc.StudentController)
	assert.Equal(t, adminCtl, rc.AdminController)
	assert.Equal(t, jwtCfg, rc.JWTConfig)
}

func TestSetupRoutes_HealthCheck(t *testing.T) {
	authCtl := schedulercontrollers.NewAuthController(nil)
	studentCtl := schedulercontrollers.NewStudentController(nil)
	adminCtl := schedulercontrollers.NewAdminController(nil)

	jwtCfg := &config.JWTConfig{
		Secret:     "test-secret",
		ExpireTime: 24 * time.Hour,
		Issuer:     "test",
	}

	rc := NewRouterConfig(authCtl, studentCtl, adminCtl, jwtCfg)

	router := gin.New()
	rc.SetupRoutes(router)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp["status"])
}
