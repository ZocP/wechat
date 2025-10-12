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

type apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// MockAuthService 模拟认证服务
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) WechatLogin(code, phoneCode string) (*service.WechatLoginResponse, error) {
	args := m.Called(code, phoneCode)
	var resp *service.WechatLoginResponse
	if v := args.Get(0); v != nil {
		resp = v.(*service.WechatLoginResponse)
	}
	return resp, args.Error(1)
}

func (m *MockAuthService) GetUserInfo(userID uint) (*model.User, error) {
	args := m.Called(userID)
	var user *model.User
	if v := args.Get(0); v != nil {
		user = v.(*model.User)
	}
	return user, args.Error(1)
}

func (m *MockAuthService) UpdateLastLogin(userID uint) error {
	args := m.Called(userID)
	return args.Error(0)
}

func setupAuthRouter(service service.AuthService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api/v1")
	handler.NewAuthHandler(service, zap.NewNop()).RegisterRoutes(api)
	return router
}

func setupAuthRouterWithUser(service service.AuthService, userID uint) *gin.Engine {
	router := gin.New()
	api := router.Group("/api/v1")
	api.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
	})
	handler.NewAuthHandler(service, zap.NewNop()).RegisterRoutes(api)
	return router
}

func TestWechatLogin_Success(t *testing.T) {
	mockService := new(MockAuthService)
	response := &service.WechatLoginResponse{
		Token: "mock_token",
		User: &model.User{
			ID:     1,
			OpenID: "mock_openid",
			Phone:  "13800138000",
			Role:   model.RolePassenger,
			Status: model.UserStatusActive,
		},
	}
	mockService.On("WechatLogin", "code", "phone").Return(response, nil).Once()

	router := setupAuthRouter(mockService)

	body := map[string]string{"code": "code", "phone_code": "phone"}
	data, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/wechat/login", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp apiResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, model.CodeSuccess, resp.Code)

	var loginResp service.WechatLoginResponse
	err = json.Unmarshal(resp.Data, &loginResp)
	assert.NoError(t, err)
	assert.Equal(t, response.Token, loginResp.Token)
	assert.Equal(t, response.User.ID, loginResp.User.ID)
	mockService.AssertExpectations(t)
}

func TestWechatLogin_InvalidRequest(t *testing.T) {
	mockService := new(MockAuthService)
	router := setupAuthRouter(mockService)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/wechat/login", bytes.NewReader([]byte(`{"code":"only"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp apiResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, model.CodeInvalidParams, resp.Code)
	mockService.AssertExpectations(t)
}

func TestWechatLogin_ServiceError(t *testing.T) {
	mockService := new(MockAuthService)
	mockService.On("WechatLogin", "code", "phone").Return((*service.WechatLoginResponse)(nil), errors.New("failed")).Once()
	router := setupAuthRouter(mockService)

	body := map[string]string{"code": "code", "phone_code": "phone"}
	data, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/wechat/login", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp apiResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, model.CodeWechatAuthFailed, resp.Code)
	mockService.AssertExpectations(t)
}

func TestGetMe_Unauthorized(t *testing.T) {
	mockService := new(MockAuthService)
	router := setupAuthRouter(mockService)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp apiResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, model.CodeUnauthorized, resp.Code)
	mockService.AssertExpectations(t)
}

func TestGetMe_Success(t *testing.T) {
	mockService := new(MockAuthService)
	expected := &model.User{ID: 1, OpenID: "openid"}
	mockService.On("GetUserInfo", uint(1)).Return(expected, nil).Once()

	router := setupAuthRouterWithUser(mockService, 1)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp apiResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, model.CodeSuccess, resp.Code)
	var user model.User
	err = json.Unmarshal(resp.Data, &user)
	assert.NoError(t, err)
	assert.Equal(t, expected.ID, user.ID)
	mockService.AssertExpectations(t)
}

func TestGetMe_ServiceError(t *testing.T) {
	mockService := new(MockAuthService)
	mockService.On("GetUserInfo", uint(1)).Return((*model.User)(nil), errors.New("failed")).Once()

	router := setupAuthRouterWithUser(mockService, 1)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp apiResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, model.CodeInternalError, resp.Code)
	mockService.AssertExpectations(t)
}
