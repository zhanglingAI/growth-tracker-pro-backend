package models

import (
	"time"
)

// User 用户模型
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	OpenID    string    `gorm:"uniqueIndex;size:64;not null" json:"openid"` // 微信 openid
	Nickname  string    `gorm:"size:50" json:"nickname"`                    // 昵称
	Avatar    string    `gorm:"size:500" json:"avatar"`                    // 头像URL
	Phone     string    `gorm:"size:20" json:"phone"`                      // 手机号(可选)
	Settings  string    `gorm:"type:text" json:"settings"`                  // JSON设置
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 表名
func (User) TableName() string {
	return "users"
}
