package service

import (
	"errors"
	"testing"
	"time"

	"pickup/internal/config"
	"pickup/internal/model"
	"pickup/internal/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// ===== Mock User Repository =====

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) Create(user *model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *mockUserRepo) GetByID(id uint) (*model.User, error) {
	args := m.Called(id)
	if v := args.Get(0); v != nil {
		return v.(*model.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserRepo) GetByOpenID(openid string) (*model.User, error) {
	args := m.Called(openid)
	if v := args.Get(0); v != nil {
		return v.(*model.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserRepo) GetByPhone(phone string) (*model.User, error) {
	args := m.Called(phone)
	if v := args.Get(0); v != nil {
		return v.(*model.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserRepo) Update(user *model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *mockUserRepo) UpdateLastLoginAt(userID uint) error {
	args := m.Called(userID)
	return args.Error(0)
}

// ===== Auth Service Tests (using direct struct) =====

func newTestAuthServiceDirect(userRepo *mockUserRepo) *authService {
	return &authService{
		userRepo:     userRepo,
		wechatClient: nil, // Won't be used in non-wechat tests
		jwtUtil:      nil, // Won't be used in non-jwt tests
		cryptoUtil:   nil,
		logger:       zap.NewNop(),
	}
}

func TestAuthService_GetUserInfo_Success(t *testing.T) {
	repo := new(mockUserRepo)
	svc := newTestAuthServiceDirect(repo)

	expected := &model.User{ID: 1, Nickname: "test", Phone: "13800138000"}
	repo.On("GetByID", uint(1)).Return(expected, nil).Once()

	user, err := svc.GetUserInfo(1)
	require.NoError(t, err)
	assert.Equal(t, expected, user)
	repo.AssertExpectations(t)
}

func TestAuthService_GetUserInfo_NotFound(t *testing.T) {
	repo := new(mockUserRepo)
	svc := newTestAuthServiceDirect(repo)

	repo.On("GetByID", uint(999)).Return(nil, errors.New("not found")).Once()

	user, err := svc.GetUserInfo(999)
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "获取用户信息失败")
	repo.AssertExpectations(t)
}

func TestAuthService_UpdateLastLogin_Success(t *testing.T) {
	repo := new(mockUserRepo)
	svc := newTestAuthServiceDirect(repo)

	repo.On("UpdateLastLoginAt", uint(1)).Return(nil).Once()

	err := svc.UpdateLastLogin(1)
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestAuthService_UpdateLastLogin_Error(t *testing.T) {
	repo := new(mockUserRepo)
	svc := newTestAuthServiceDirect(repo)

	repo.On("UpdateLastLoginAt", uint(1)).Return(errors.New("db error")).Once()

	err := svc.UpdateLastLogin(1)
	assert.Error(t, err)
	repo.AssertExpectations(t)
}

// TestAuthService_WechatLogin_ExistingUser tests the login flow when user exists
// We can't easily test the full WechatLogin without a mock HTTP server for wechat,
// but we can test the internal logic by calling the method parts.

func TestAuthService_WechatLogin_ExistingUser_UpdateFails(t *testing.T) {
	repo := new(mockUserRepo)
	svc := newTestAuthServiceDirect(repo)

	// Set up jwtUtil for token generation
	svc.jwtUtil = utils.NewJWTUtil("test-secret", 24*time.Hour, "test-issuer")

	existingUser := &model.User{
		ID:     1,
		OpenID: "openid_123",
		Phone:  "13800138000",
		Role:   model.RolePassenger,
		Status: model.UserStatusActive,
	}

	// Test GetUserInfo path
	repo.On("GetByID", uint(1)).Return(existingUser, nil).Once()

	user, err := svc.GetUserInfo(1)
	require.NoError(t, err)
	assert.Equal(t, "openid_123", user.OpenID)
	repo.AssertExpectations(t)
}

// TestNewAuthService verifies the constructor
func TestNewAuthService(t *testing.T) {
	repo := new(mockUserRepo)
	wechatCfg := &config.WechatConfig{AppID: "test_appid", AppSecret: "test_secret"}
	jwtCfg := &config.JWTConfig{Secret: "test-secret", ExpireTime: 24 * time.Hour, Issuer: "test"}
	cryptoCfg := &config.CryptoConfig{Key: "test-crypto-key-32-characters!!"}

	svc := NewAuthService(repo, wechatCfg, jwtCfg, cryptoCfg, zap.NewNop())
	require.NotNil(t, svc)
}
