package models

import "time"

// User 用户表
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	OpenID    string    `gorm:"column:open_id;type:varchar(64);not null;uniqueIndex:uk_users_open_id" json:"open_id"`
	Name      string    `gorm:"type:varchar(64);not null" json:"name"`
	Phone     string    `gorm:"type:varchar(20)" json:"phone"`
	Role      UserRole  `gorm:"type:enum('student','staff','admin');not null;default:'student';index:idx_users_role" json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}
