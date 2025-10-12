package config

import ()

// WechatConfig 微信配置
type WechatConfig struct {
	AppID     string `yaml:"appId"`
	AppSecret string `yaml:"appSecret"`
	MchID     string `yaml:"mchId"`
	MchKey    string `yaml:"mchKey"`
	NotifyURL string `yaml:"notifyUrl"`
}

// NewWechatConfig 创建微信配置
func NewWechatConfig() *WechatConfig {
	return &WechatConfig{
		AppID:     getEnv("WECHAT_APPID", ""),
		AppSecret: getEnv("WECHAT_SECRET", ""),
		MchID:     getEnv("WECHAT_MCH_ID", ""),
		MchKey:    getEnv("WECHAT_MCH_KEY", ""),
		NotifyURL: getEnv("WECHAT_NOTIFY_URL", ""),
	}
}
