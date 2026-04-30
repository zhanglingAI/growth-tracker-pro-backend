package models

// User 用户模型
type User struct {
	BaseModel `gorm:"embedded"`
	OpenID    string `gorm:"uniqueIndex;size:64;not null" json:"openid"`      // 微信 openid
	NickName  string `gorm:"size:50;column:nick_name" json:"nick_name"`        // 昵称
	AvatarURL string `gorm:"size:500;column:avatar_url" json:"avatar_url"`     // 头像URL
	Phone     string `gorm:"size:20" json:"phone"`                             // 手机号(可选)
	Settings  string `gorm:"type:text" json:"settings"`                         // JSON设置
}

// TableName 表名
func (User) TableName() string {
	return "users"
}
