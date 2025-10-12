package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// WechatClient 微信客户端
type WechatClient struct {
	appID     string
	appSecret string
	baseURL   string
}

// NewWechatClient 创建微信客户端
func NewWechatClient(appID, appSecret string) *WechatClient {
	return &WechatClient{
		appID:     appID,
		appSecret: appSecret,
		baseURL:   "https://api.weixin.qq.com",
	}
}

// JSCode2SessionResponse 微信登录响应
type JSCode2SessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// GetPhoneNumberResponse 获取手机号响应
type GetPhoneNumberResponse struct {
	ErrCode   int    `json:"errcode"`
	ErrMsg    string `json:"errmsg"`
	PhoneInfo struct {
		PhoneNumber     string `json:"phoneNumber"`
		PurePhoneNumber string `json:"purePhoneNumber"`
		CountryCode     string `json:"countryCode"`
	} `json:"phone_info"`
}

// JSCode2Session 微信登录
func (w *WechatClient) JSCode2Session(code string) (*JSCode2SessionResponse, error) {
	params := url.Values{}
	params.Set("appid", w.appID)
	params.Set("secret", w.appSecret)
	params.Set("js_code", code)
	params.Set("grant_type", "authorization_code")

	resp, err := http.Get(w.baseURL + "/sns/jscode2session?" + params.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result JSCode2SessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat api error: %d - %s", result.ErrCode, result.ErrMsg)
	}

	return &result, nil
}

// GetPhoneNumber 获取手机号（需要access_token）
func (w *WechatClient) GetPhoneNumber(accessToken, code string) (*GetPhoneNumberResponse, error) {
	reqBody := map[string]string{
		"code": code,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/wxa/business/getuserphonenumber?access_token=%s", w.baseURL, accessToken),
		"application/json",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result GetPhoneNumberResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat api error: %d - %s", result.ErrCode, result.ErrMsg)
	}

	return &result, nil
}
