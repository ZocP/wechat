package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pickup/internal/model"
	"pickup/internal/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newAuthSvcForWechatTest(repo *mockUserRepo, baseURL string) *authService {
	w := utils.NewWechatClient("appid", "secret")
	w.SetBaseURL(baseURL)
	return &authService{
		userRepo:     repo,
		wechatClient: w,
		jwtUtil:      utils.NewJWTUtil("secret", time.Hour, "issuer"),
		cryptoUtil:   nil,
		logger:       zap.NewNop(),
	}
}

func TestWechatLogin_CreateUserPath(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/sns/jscode2session", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"openid": "openid-1", "session_key": "sk", "unionid": "u1"})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	repo := new(mockUserRepo)
	repo.On("GetByOpenID", "openid-1").Return(nil, errors.New("not found")).Once()
	repo.On("Create", mock.AnythingOfType("*model.User")).Return(nil).Once()
	repo.On("UpdateLastLoginAt", uint(0)).Return(nil).Once()

	svc := newAuthSvcForWechatTest(repo, server.URL)
	resp, err := svc.WechatLogin("code", "13800138000")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, "openid-1", resp.User.OpenID)
	repo.AssertExpectations(t)
}

func TestWechatLogin_UpdateUserPath(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/sns/jscode2session", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"openid": "openid-2", "session_key": "sk", "unionid": "u2"})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	repo := new(mockUserRepo)
	existing := &model.User{ID: 7, OpenID: "openid-2", Role: model.RolePassenger, Status: model.UserStatusActive}
	repo.On("GetByOpenID", "openid-2").Return(existing, nil).Once()
	repo.On("Update", mock.AnythingOfType("*model.User")).Return(nil).Once()
	repo.On("UpdateLastLoginAt", uint(7)).Return(nil).Once()

	svc := newAuthSvcForWechatTest(repo, server.URL)
	resp, err := svc.WechatLogin("code", "13800138000")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, uint(7), resp.User.ID)
	repo.AssertExpectations(t)
}

func TestWechatLogin_ErrorBranches(t *testing.T) {
	t.Run("session error", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/sns/jscode2session", func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{"errcode": 40029, "errmsg": "bad code"})
		})
		server := httptest.NewServer(mux)
		defer server.Close()

		repo := new(mockUserRepo)
		svc := newAuthSvcForWechatTest(repo, server.URL)
		resp, err := svc.WechatLogin("bad", "1")
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("create user failed", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/sns/jscode2session", func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{"openid": "openid-3", "session_key": "sk"})
		})
		server := httptest.NewServer(mux)
		defer server.Close()

		repo := new(mockUserRepo)
		repo.On("GetByOpenID", "openid-3").Return(nil, errors.New("not found")).Once()
		repo.On("Create", mock.AnythingOfType("*model.User")).Return(errors.New("db err")).Once()

		svc := newAuthSvcForWechatTest(repo, server.URL)
		resp, err := svc.WechatLogin("code", "1")
		assert.Error(t, err)
		assert.Nil(t, resp)
		repo.AssertExpectations(t)
	})

	t.Run("update user failed", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/sns/jscode2session", func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{"openid": "openid-4", "session_key": "sk"})
		})
		server := httptest.NewServer(mux)
		defer server.Close()

		repo := new(mockUserRepo)
		existing := &model.User{ID: 4, OpenID: "openid-4", Role: model.RolePassenger, Status: model.UserStatusActive}
		repo.On("GetByOpenID", "openid-4").Return(existing, nil).Once()
		repo.On("Update", mock.AnythingOfType("*model.User")).Return(errors.New("db err")).Once()

		svc := newAuthSvcForWechatTest(repo, server.URL)
		resp, err := svc.WechatLogin("code", "1")
		assert.Error(t, err)
		assert.Nil(t, resp)
		repo.AssertExpectations(t)
	})
}
