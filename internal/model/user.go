package model

import (
	"time"

	"gorm.io/gorm"
)

// UserRole 用户角色枚举
type UserRole string

const (
	RolePassenger  UserRole = "passenger"  // 乘客
	RoleDriver     UserRole = "driver"     // 司机
	RoleDispatcher UserRole = "dispatcher" // 调度员
	RoleAdmin      UserRole = "admin"      // 管理员
)

// UserStatus 用户状态枚举
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"   // 激活
	UserStatusInactive UserStatus = "inactive" // 未激活
	UserStatusBlocked  UserStatus = "blocked"  // 封禁
)

// User 用户模型
type User struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	OpenID      string     `json:"openid" gorm:"uniqueIndex;size:64;not null"`                                           // 微信OpenID
	UnionID     string     `json:"unionid" gorm:"size:64"`                                                               // 微信UnionID
	Phone       string     `json:"phone" gorm:"uniqueIndex;size:20;not null"`                                            // 手机号
	Nickname    string     `json:"nickname" gorm:"size:64"`                                                              // 昵称
	AvatarURL   string     `json:"avatar_url" gorm:"size:255"`                                                           // 头像URL
	Role        UserRole   `json:"role" gorm:"type:enum('passenger','driver','dispatcher','admin');default:'passenger'"` // 角色
	Status      UserStatus `json:"status" gorm:"type:enum('active','inactive','blocked');default:'active'"`              // 状态
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	LastLoginAt *time.Time `json:"last_login_at"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// BeforeCreate GORM钩子
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.Role == "" {
		u.Role = RolePassenger
	}
	if u.Status == "" {
		u.Status = UserStatusActive
	}
	return nil
}
