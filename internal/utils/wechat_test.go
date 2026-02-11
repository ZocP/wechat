package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWechatClient(t *testing.T) {
	client := NewWechatClient("app123", "secret456")
	assert.NotNil(t, client)
	assert.Equal(t, "app123", client.appID)
	assert.Equal(t, "secret456", client.appSecret)
	assert.Equal(t, "https://api.weixin.qq.com", client.baseURL)
}

func TestJSCode2Session_Success(t *testing.T) {
	// Mock WeChat API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/sns/jscode2session", r.URL.Path)
		assert.Equal(t, "test_appid", r.URL.Query().Get("appid"))
		assert.Equal(t, "test_secret", r.URL.Query().Get("secret"))
		assert.Equal(t, "test_code", r.URL.Query().Get("js_code"))
		assert.Equal(t, "authorization_code", r.URL.Query().Get("grant_type"))

		resp := JSCode2SessionResponse{
			OpenID:     "openid_123",
			SessionKey: "session_key_456",
			UnionID:    "unionid_789",
			ErrCode:    0,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &WechatClient{
		appID:     "test_appid",
		appSecret: "test_secret",
		baseURL:   server.URL,
	}

	result, err := client.JSCode2Session("test_code")
	require.NoError(t, err)
	assert.Equal(t, "openid_123", result.OpenID)
	assert.Equal(t, "session_key_456", result.SessionKey)
	assert.Equal(t, "unionid_789", result.UnionID)
}

func TestJSCode2Session_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := JSCode2SessionResponse{
			ErrCode: 40029,
			ErrMsg:  "invalid code",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &WechatClient{
		appID:     "test_appid",
		appSecret: "test_secret",
		baseURL:   server.URL,
	}

	_, err := client.JSCode2Session("bad_code")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "40029")
}

func TestJSCode2Session_NetworkError(t *testing.T) {
	client := &WechatClient{
		appID:     "test_appid",
		appSecret: "test_secret",
		baseURL:   "http://localhost:1", // Invalid port
	}

	_, err := client.JSCode2Session("code")
	assert.Error(t, err)
}

func TestJSCode2Session_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	client := &WechatClient{
		appID:     "test_appid",
		appSecret: "test_secret",
		baseURL:   server.URL,
	}

	_, err := client.JSCode2Session("code")
	assert.Error(t, err)
}

func TestGetPhoneNumber_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/wxa/business/getuserphonenumber", r.URL.Path)
		assert.Equal(t, "test_token", r.URL.Query().Get("access_token"))

		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "phone_code", body["code"])

		resp := GetPhoneNumberResponse{
			ErrCode: 0,
		}
		resp.PhoneInfo.PhoneNumber = "13800138000"
		resp.PhoneInfo.PurePhoneNumber = "13800138000"
		resp.PhoneInfo.CountryCode = "86"
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &WechatClient{
		appID:     "test_appid",
		appSecret: "test_secret",
		baseURL:   server.URL,
	}

	result, err := client.GetPhoneNumber("test_token", "phone_code")
	require.NoError(t, err)
	assert.Equal(t, "13800138000", result.PhoneInfo.PhoneNumber)
	assert.Equal(t, "86", result.PhoneInfo.CountryCode)
}

func TestGetPhoneNumber_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := GetPhoneNumberResponse{
			ErrCode: 40001,
			ErrMsg:  "invalid token",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &WechatClient{
		appID:     "test_appid",
		appSecret: "test_secret",
		baseURL:   server.URL,
	}

	_, err := client.GetPhoneNumber("bad_token", "code")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "40001")
}

func TestGetPhoneNumber_NetworkError(t *testing.T) {
	client := &WechatClient{
		appID:     "test_appid",
		appSecret: "test_secret",
		baseURL:   "http://localhost:1",
	}

	_, err := client.GetPhoneNumber("token", "code")
	assert.Error(t, err)
}
