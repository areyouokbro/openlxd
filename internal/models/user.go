package models

import (
	"time"
)

// User 用户模型
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:100;not null" json:"username"`
	Email        string    `gorm:"uniqueIndex;size:255;not null" json:"email"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"` // 不返回密码哈希
	APIKey       string    `gorm:"uniqueIndex;size:64;not null" json:"api_key"`
	Role         string    `gorm:"size:20;default:user" json:"role"` // admin, user
	Status       string    `gorm:"size:20;default:active" json:"status"` // active, suspended, deleted
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// IsAdmin 检查是否是管理员
func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}

// IsActive 检查用户是否活跃
func (u *User) IsActive() bool {
	return u.Status == "active"
}
