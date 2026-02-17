package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pickup/internal/config"
	"pickup/internal/scheduler/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAuthService_LoginWithWechatCode_CreateAndExisting(t *testing.T) {
	db := newTestDB(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/sns/jscode2session", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"openid":      "openid-1",
			"session_key": "sk",
			"errcode":     0,
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	svc := NewAuthService(
		db,
		&config.WechatConfig{AppID: "app", AppSecret: "sec"},
		&config.JWTConfig{Secret: "secret", ExpireTime: time.Hour, Issuer: "pickup"},
		zap.NewNop(),
	)
	svc.wechatClient.SetBaseURL(server.URL)

	res1, err := svc.LoginWithWechatCode("code-1")
	require.NoError(t, err)
	require.NotEmpty(t, res1.Token)
	assert.Equal(t, "openid-1", res1.User.OpenID)
	assert.Equal(t, models.UserRoleStudent, res1.User.Role)

	res2, err := svc.LoginWithWechatCode("code-1")
	require.NoError(t, err)
	assert.Equal(t, res1.User.ID, res2.User.ID)
}

func TestAuthService_BindPhone_SuccessAndEmptyPhone(t *testing.T) {
	db := newTestDB(t)
	require.NoError(t, db.Create(&models.User{OpenID: "openid-2", Name: "u2", Role: models.UserRoleStudent}).Error)

	phoneNumber := "13800138000"
	mux := http.NewServeMux()
	mux.HandleFunc("/cgi-bin/token", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "token-1", "expires_in": 7200})
	})
	mux.HandleFunc("/wxa/business/getuserphonenumber", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode": 0,
			"phone_info": map[string]any{
				"purePhoneNumber": phoneNumber,
			},
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	svc := NewAuthService(
		db,
		&config.WechatConfig{AppID: "app", AppSecret: "sec"},
		&config.JWTConfig{Secret: "secret", ExpireTime: time.Hour, Issuer: "pickup"},
		zap.NewNop(),
	)
	svc.wechatClient.SetBaseURL(server.URL)

	var user models.User
	require.NoError(t, db.Where("open_id = ?", "openid-2").First(&user).Error)
	require.NoError(t, svc.BindPhone(user.ID, "phone-code"))

	require.NoError(t, db.First(&user, user.ID).Error)
	assert.Equal(t, phoneNumber, user.Phone)

	muxEmpty := http.NewServeMux()
	muxEmpty.HandleFunc("/cgi-bin/token", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "token-2", "expires_in": 7200})
	})
	muxEmpty.HandleFunc("/wxa/business/getuserphonenumber", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"errcode": 0, "phone_info": map[string]any{}})
	})
	serverEmpty := httptest.NewServer(muxEmpty)
	defer serverEmpty.Close()
	svc.wechatClient.SetBaseURL(serverEmpty.URL)

	err := svc.BindPhone(user.ID, "phone-code")
	assert.ErrorContains(t, err, "wechat phone number is empty")
}
