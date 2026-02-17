package config

// WechatConfig 微信配置
type WechatConfig struct {
	AppID       string `yaml:"appId"`
	AppSecret   string `yaml:"appSecret"`
	MchID       string `yaml:"mchId"`
	MchKey      string `yaml:"mchKey"`
	NotifyURL   string `yaml:"notifyUrl"`
	AdminPhone  string `yaml:"adminPhone"`
	AdminOpenID string `yaml:"adminOpenId"`
}

// NewWechatConfig 创建微信配置
func NewWechatConfig() *WechatConfig {
	return &WechatConfig{
		AppID:       getEnvOrConfig("WECHAT_APPID", "wechat.appId", ""),
		AppSecret:   getEnvOrConfig("WECHAT_SECRET", "wechat.appSecret", ""),
		MchID:       getEnvOrConfig("WECHAT_MCH_ID", "wechat.mchId", ""),
		MchKey:      getEnvOrConfig("WECHAT_MCH_KEY", "wechat.mchKey", ""),
		NotifyURL:   getEnvOrConfig("WECHAT_NOTIFY_URL", "wechat.notifyUrl", ""),
		AdminPhone:  getEnvOrConfig("WECHAT_ADMIN_PHONE", "wechat.adminPhone", ""),
		AdminOpenID: getEnvOrConfig("WECHAT_ADMIN_OPEN_ID", "wechat.adminOpenId", ""),
	}
}
