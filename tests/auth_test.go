package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"pickup/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService 模拟认证服务
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) WechatLogin(code, phoneCode string) (*model.WechatLoginResponse, error) {
	args := m.Called(code, phoneCode)
	return args.Get(0).(*model.WechatLoginResponse), args.Error(1)
}

func (m *MockAuthService) GetUserInfo(userID uint) (*model.User, error) {
	args := m.Called(userID)
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockAuthService) UpdateLastLogin(userID uint) error {
	args := m.Called(userID)
	return args.Error(0)
}

// TestWechatLogin 测试微信登录
func TestWechatLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建模拟服务
	mockAuthService := new(MockAuthService)

	// 设置期望的调用
	expectedResponse := &model.WechatLoginResponse{
		Token: "mock_token",
		User: &model.User{
			ID:     1,
			OpenID: "mock_openid",
			Phone:  "13800138000",
			Role:   model.RolePassenger,
			Status: model.UserStatusActive,
		},
	}

	mockAuthService.On("WechatLogin", "test_code", "test_phone_code").Return(expectedResponse, nil)

	// 创建测试请求
	requestBody := map[string]string{
		"code":       "test_code",
		"phone_code": "test_phone_code",
	}
	jsonBody, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/api/v1/auth/wechat/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// 创建响应记录器
	w := httptest.NewRecorder()

	// 创建路由
	router := gin.New()
	// 这里需要实际的处理器，暂时跳过
	// authHandler := handler.NewAuthHandler(mockAuthService, nil)
	// authHandler.RegisterRoutes(router.Group("/api/v1"))

	// 执行请求
	router.ServeHTTP(w, req)

	// 验证结果
	assert.Equal(t, http.StatusOK, w.Code)

	// 验证服务调用
	mockAuthService.AssertExpectations(t)
}

// TestGetMe 测试获取用户信息
func TestGetMe(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建模拟服务
	mockAuthService := new(MockAuthService)

	// 设置期望的调用
	expectedUser := &model.User{
		ID:     1,
		OpenID: "mock_openid",
		Phone:  "13800138000",
		Role:   model.RolePassenger,
		Status: model.UserStatusActive,
	}

	mockAuthService.On("GetUserInfo", uint(1)).Return(expectedUser, nil)

	// 创建测试请求
	req, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)
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
	mockAuthService.AssertExpectations(t)
}
