package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pickup/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRegistrationService 模拟报名服务
type MockRegistrationService struct {
	mock.Mock
}

func (m *MockRegistrationService) CreateRegistration(userID uint, req *model.CreateRegistrationRequest) (*model.Registration, error) {
	args := m.Called(userID, req)
	return args.Get(0).(*model.Registration), args.Error(1)
}

func (m *MockRegistrationService) GetRegistration(id uint, userID uint) (*model.Registration, error) {
	args := m.Called(id, userID)
	return args.Get(0).(*model.Registration), args.Error(1)
}

func (m *MockRegistrationService) UpdateRegistration(id uint, userID uint, req *model.UpdateRegistrationRequest) (*model.Registration, error) {
	args := m.Called(id, userID, req)
	return args.Get(0).(*model.Registration), args.Error(1)
}

func (m *MockRegistrationService) GetUserRegistrations(userID uint) ([]*model.Registration, error) {
	args := m.Called(userID)
	return args.Get(0).([]*model.Registration), args.Error(1)
}

func (m *MockRegistrationService) DeleteRegistration(id uint, userID uint) error {
	args := m.Called(id, userID)
	return args.Error(0)
}

// TestCreateRegistration 测试创建报名
func TestCreateRegistration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建模拟服务
	mockService := new(MockRegistrationService)

	// 设置期望的调用
	expectedRegistration := &model.Registration{
		ID:            1,
		UserID:        1,
		Name:          "张三",
		Phone:         "13800138000",
		FlightNo:      "CA1234",
		ArrivalDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		ArrivalTime:   "14:30:00",
		DepartureCity: "北京",
		Companions:    2,
		LuggageCount:  1,
		PickupMethod:  model.PickupMethodGroup,
		Status:        model.RegistrationStatusDraft,
	}

	request := &model.CreateRegistrationRequest{
		Name:          "张三",
		Phone:         "13800138000",
		FlightNo:      "CA1234",
		ArrivalDate:   "2024-01-01",
		ArrivalTime:   "14:30",
		DepartureCity: "北京",
		Companions:    2,
		LuggageCount:  1,
		PickupMethod:  model.PickupMethodGroup,
	}

	mockService.On("CreateRegistration", uint(1), request).Return(expectedRegistration, nil)

	// 创建测试请求
	jsonBody, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/api/v1/registrations", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer mock_token")

	// 创建响应记录器
	w := httptest.NewRecorder()

	// 创建路由
	router := gin.New()
	// 这里需要实际的处理器，暂时跳过

	// 执行请求
	router.ServeHTTP(w, req)

	// 验证结果
	assert.Equal(t, http.StatusCreated, w.Code)

	// 验证服务调用
	mockService.AssertExpectations(t)
}

// TestGetRegistration 测试获取报名
func TestGetRegistration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建模拟服务
	mockService := new(MockRegistrationService)

	// 设置期望的调用
	expectedRegistration := &model.Registration{
		ID:            1,
		UserID:        1,
		Name:          "张三",
		Phone:         "13800138000",
		FlightNo:      "CA1234",
		ArrivalDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		ArrivalTime:   "14:30:00",
		DepartureCity: "北京",
		Companions:    2,
		LuggageCount:  1,
		PickupMethod:  model.PickupMethodGroup,
		Status:        model.RegistrationStatusDraft,
	}

	mockService.On("GetRegistration", uint(1), uint(1)).Return(expectedRegistration, nil)

	// 创建测试请求
	req, _ := http.NewRequest("GET", "/api/v1/registrations/1", nil)
	req.Header.Set("Authorization", "Bearer mock_token")

	// 创建响应记录器
	w := httptest.NewRecorder()

	// 创建路由
	router := gin.New()
	// 这里需要实际的处理器，暂时跳过

	// 执行请求
	router.ServeHTTP(w, req)

	// 验证结果
	assert.Equal(t, http.StatusOK, w.Code)

	// 验证服务调用
	mockService.AssertExpectations(t)
}

// TestGetUserRegistrations 测试获取用户报名列表
func TestGetUserRegistrations(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建模拟服务
	mockService := new(MockRegistrationService)

	// 设置期望的调用
	expectedRegistrations := []*model.Registration{
		{
			ID:            1,
			UserID:        1,
			Name:          "张三",
			Phone:         "13800138000",
			FlightNo:      "CA1234",
			ArrivalDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			ArrivalTime:   "14:30:00",
			DepartureCity: "北京",
			Companions:    2,
			LuggageCount:  1,
			PickupMethod:  model.PickupMethodGroup,
			Status:        model.RegistrationStatusDraft,
		},
	}

	mockService.On("GetUserRegistrations", uint(1)).Return(expectedRegistrations, nil)

	// 创建测试请求
	req, _ := http.NewRequest("GET", "/api/v1/registrations/my", nil)
	req.Header.Set("Authorization", "Bearer mock_token")

	// 创建响应记录器
	w := httptest.NewRecorder()

	// 创建路由
	router := gin.New()
	// 这里需要实际的处理器，暂时跳过

	// 执行请求
	router.ServeHTTP(w, req)

	// 验证结果
	assert.Equal(t, http.StatusOK, w.Code)

	// 验证服务调用
	mockService.AssertExpectations(t)
}
