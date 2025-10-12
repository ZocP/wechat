package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// ConsentScope 同意范围枚举
type ConsentScope string

const (
	ConsentScopePhone    ConsentScope = "phone"    // 手机号授权
	ConsentScopeWechat   ConsentScope = "wechat"   // 微信信息授权
	ConsentScopeLocation ConsentScope = "location" // 位置信息授权
)

// JSONB 自定义JSONB类型
type JSONB map[string]interface{}

// Scan 实现Scanner接口
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONB)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return gorm.ErrInvalidData
	}
	return json.Unmarshal(bytes, j)
}

// Value 实现Valuer接口
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// ConsentLog 同意记录模型
type ConsentLog struct {
	ID         uint         `json:"id" gorm:"primaryKey"`
	UserID     uint         `json:"user_id" gorm:"not null;index"`
	Scope      ConsentScope `json:"scope" gorm:"type:enum('phone','wechat','location');not null"`
	GrantedAt  time.Time    `json:"granted_at" gorm:"not null"`
	RawPayload JSONB        `json:"raw_payload" gorm:"type:json"` // 原始载荷
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`

	// 关联
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName 指定表名
func (ConsentLog) TableName() string {
	return "consent_logs"
}
