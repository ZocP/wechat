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

// MockRegistrationService 模拟报名服务
type MockRegistrationService struct {
	mock.Mock
}

func (m *MockRegistrationService) CreateRegistration(userID uint, req *service.CreateRegistrationRequest) (*model.Registration, error) {
	args := m.Called(userID, req)
	var registration *model.Registration
	if v := args.Get(0); v != nil {
		registration = v.(*model.Registration)
	}
	return registration, args.Error(1)
}

func (m *MockRegistrationService) GetRegistration(id uint, userID uint) (*model.Registration, error) {
	args := m.Called(id, userID)
	var registration *model.Registration
	if v := args.Get(0); v != nil {
		registration = v.(*model.Registration)
	}
	return registration, args.Error(1)
}

func (m *MockRegistrationService) UpdateRegistration(id uint, userID uint, req *service.UpdateRegistrationRequest) (*model.Registration, error) {
	args := m.Called(id, userID, req)
	var registration *model.Registration
	if v := args.Get(0); v != nil {
		registration = v.(*model.Registration)
	}
	return registration, args.Error(1)
}

func (m *MockRegistrationService) GetUserRegistrations(userID uint) ([]*model.Registration, error) {
	args := m.Called(userID)
	if v := args.Get(0); v != nil {
		return v.([]*model.Registration), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRegistrationService) DeleteRegistration(id uint, userID uint) error {
	args := m.Called(id, userID)
	return args.Error(0)
}

func setupRegistrationRouter(service service.RegistrationService, withUser bool) *gin.Engine {
	router := gin.New()
	api := router.Group("/api/v1")
	if withUser {
		api.Use(func(c *gin.Context) {
			c.Set("user_id", uint(1))
		})
	}
	handler.NewRegistrationHandler(service, zap.NewNop()).RegisterRoutes(api)
	return router
}

func TestCreateRegistration_Success(t *testing.T) {
	mockService := new(MockRegistrationService)
	expected := &model.Registration{ID: 1, UserID: 1, Name: "张三"}
	mockService.On("CreateRegistration", uint(1), mock.MatchedBy(func(req *service.CreateRegistrationRequest) bool {
		return req.Name == "张三" && req.Phone == "13800138000"
	})).Return(expected, nil).Once()

	router := setupRegistrationRouter(mockService, true)

	body := map[string]interface{}{
		"name":           "张三",
		"phone":          "13800138000",
		"flight_no":      "CA1234",
		"arrival_date":   "2024-01-01",
		"arrival_time":   "14:30",
		"departure_city": "北京",
		"companions":     1,
		"luggage_count":  1,
		"pickup_method":  "group",
	}
	data, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/registrations", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp apiResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, model.CodeSuccess, resp.Code)
	mockService.AssertExpectations(t)
}

func TestCreateRegistration_Unauthorized(t *testing.T) {
	mockService := new(MockRegistrationService)
	router := setupRegistrationRouter(mockService, false)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/registrations", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertExpectations(t)
}

func TestCreateRegistration_InvalidBody(t *testing.T) {
	mockService := new(MockRegistrationService)
	router := setupRegistrationRouter(mockService, true)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/registrations", bytes.NewReader([]byte(`{"name":123}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestCreateRegistration_ServiceError(t *testing.T) {
	mockService := new(MockRegistrationService)
	mockService.On("CreateRegistration", uint(1), mock.Anything).Return((*model.Registration)(nil), errors.New("failed")).Once()
	router := setupRegistrationRouter(mockService, true)

	body := map[string]interface{}{
		"name":           "张三",
		"phone":          "13800138000",
		"flight_no":      "CA1234",
		"arrival_date":   "2024-01-01",
		"arrival_time":   "14:30",
		"departure_city": "北京",
		"pickup_method":  "group",
	}
	data, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/registrations", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetRegistration_Success(t *testing.T) {
	mockService := new(MockRegistrationService)
	expected := &model.Registration{ID: 1, UserID: 1, Name: "张三"}
	mockService.On("GetRegistration", uint(1), uint(1)).Return(expected, nil).Once()
	router := setupRegistrationRouter(mockService, true)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/registrations/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetRegistration_InvalidID(t *testing.T) {
	mockService := new(MockRegistrationService)
	router := setupRegistrationRouter(mockService, true)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/registrations/abc", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetRegistration_ServiceError(t *testing.T) {
	mockService := new(MockRegistrationService)
	mockService.On("GetRegistration", uint(1), uint(1)).Return((*model.Registration)(nil), errors.New("not found")).Once()
	router := setupRegistrationRouter(mockService, true)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/registrations/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetRegistration_Unauthorized(t *testing.T) {
	mockService := new(MockRegistrationService)
	router := setupRegistrationRouter(mockService, false)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/registrations/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetMyRegistrations_Success(t *testing.T) {
	mockService := new(MockRegistrationService)
	expected := []*model.Registration{{ID: 1, UserID: 1}}
	mockService.On("GetUserRegistrations", uint(1)).Return(expected, nil).Once()
	router := setupRegistrationRouter(mockService, true)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/registrations", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetMyRegistrations_ServiceError(t *testing.T) {
	mockService := new(MockRegistrationService)
	mockService.On("GetUserRegistrations", uint(1)).Return(([]*model.Registration)(nil), errors.New("failed")).Once()
	router := setupRegistrationRouter(mockService, true)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/registrations", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestUpdateRegistration_Success(t *testing.T) {
	mockService := new(MockRegistrationService)
	expected := &model.Registration{ID: 1, UserID: 1, Name: "李四"}
	mockService.On("UpdateRegistration", uint(1), uint(1), mock.MatchedBy(func(req *service.UpdateRegistrationRequest) bool {
		return req.Name != nil && *req.Name == "李四"
	})).Return(expected, nil).Once()
	router := setupRegistrationRouter(mockService, true)

	body := map[string]string{"name": "李四"}
	data, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/registrations/1", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestUpdateRegistration_InvalidID(t *testing.T) {
	mockService := new(MockRegistrationService)
	router := setupRegistrationRouter(mockService, true)

	req, _ := http.NewRequest(http.MethodPut, "/api/v1/registrations/abc", bytes.NewReader([]byte(`{"name":"李四"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestUpdateRegistration_InvalidBody(t *testing.T) {
	mockService := new(MockRegistrationService)
	router := setupRegistrationRouter(mockService, true)

	req, _ := http.NewRequest(http.MethodPut, "/api/v1/registrations/1", bytes.NewReader([]byte(`{"name":123}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestUpdateRegistration_ServiceError(t *testing.T) {
	mockService := new(MockRegistrationService)
	mockService.On("UpdateRegistration", uint(1), uint(1), mock.Anything).Return((*model.Registration)(nil), errors.New("failed")).Once()
	router := setupRegistrationRouter(mockService, true)

	body := map[string]string{"name": "李四"}
	data, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/registrations/1", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestUpdateRegistration_Unauthorized(t *testing.T) {
	mockService := new(MockRegistrationService)
	router := setupRegistrationRouter(mockService, false)

	req, _ := http.NewRequest(http.MethodPut, "/api/v1/registrations/1", bytes.NewReader([]byte(`{"name":"李四"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertExpectations(t)
}

func TestDeleteRegistration_Success(t *testing.T) {
	mockService := new(MockRegistrationService)
	mockService.On("DeleteRegistration", uint(1), uint(1)).Return(nil).Once()
	router := setupRegistrationRouter(mockService, true)

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/registrations/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestDeleteRegistration_InvalidID(t *testing.T) {
	mockService := new(MockRegistrationService)
	router := setupRegistrationRouter(mockService, true)

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/registrations/abc", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestDeleteRegistration_ServiceError(t *testing.T) {
	mockService := new(MockRegistrationService)
	mockService.On("DeleteRegistration", uint(1), uint(1)).Return(errors.New("failed")).Once()
	router := setupRegistrationRouter(mockService, true)

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/registrations/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestDeleteRegistration_Unauthorized(t *testing.T) {
	mockService := new(MockRegistrationService)
	router := setupRegistrationRouter(mockService, false)

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/registrations/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertExpectations(t)
}
