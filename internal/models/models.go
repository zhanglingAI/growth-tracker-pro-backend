package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel 基础模型
type BaseModel struct {
	ID        string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// BeforeCreate 创建前回调
func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return nil
}

// User 用户模型
type User struct {
	BaseModel   `gorm:"embedded"`
	OpenID      string    `gorm:"type:varchar(128);uniqueIndex;not null" json:"open_id"` // 微信OpenID
	NickName    string    `gorm:"type:varchar(64)" json:"nick_name"`
	AvatarURL   string    `gorm:"type:varchar(512)" json:"avatar_url"`
	Phone       string    `gorm:"type:varchar(20)" json:"phone"`
	Settings    string    `gorm:"type:text" json:"settings"` // JSON字符串
	Children    []Child   `gorm:"foreignKey:UserID" json:"children,omitempty"`
	FamilyID    *string   `gorm:"type:varchar(36)" json:"family_id,omitempty"`
	Family      *Family   `gorm:"foreignKey:FamilyID" json:"family,omitempty"`
	Subscription *Subscription `gorm:"foreignKey:UserID" json:"subscription,omitempty"`
}

// TableName 表名
func (User) TableName() string {
	return "users"
}

// Child 宝宝模型
type Child struct {
	BaseModel    `gorm:"embedded"`
	UserID       string    `gorm:"type:varchar(36);index;not null" json:"user_id"`
	Name         string    `gorm:"type:varchar(64);not null" json:"name"`
	Gender       string    `gorm:"type:varchar(10);not null" json:"gender"` // male/female
	Birthday     time.Time `gorm:"type:date;not null" json:"birthday"`
	FatherHeight float64   `gorm:"type:decimal(5,1)" json:"father_height"`
	MotherHeight float64   `gorm:"type:decimal(5,1)" json:"mother_height"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	Records      []Record  `gorm:"foreignKey:ChildID" json:"records,omitempty"`
}

// AgeInDays 计算年龄天数
func (c *Child) AgeInDays() int {
	return int(time.Since(c.Birthday).Hours() / 24)
}

// TableName 表名
func (Child) TableName() string {
	return "children"
}

// CalculateAge 计算宝宝年龄
func (c *Child) CalculateAge(at time.Time) (years, months int) {
	years = at.Year() - c.Birthday.Year()
	months = int(at.Sub(c.Birthday).Hours() / (24 * 30))
	if months >= 12 {
		years++
		months -= 12
	}
	return
}

// Record 生长记录模型
type Record struct {
	BaseModel  `gorm:"embedded"`
	ChildID    string    `gorm:"type:varchar(36);index;not null" json:"child_id"`
	Height     float64   `gorm:"type:decimal(5,1);not null" json:"height"`     // 身高 cm
	Weight     float64   `gorm:"type:decimal(5,1);not null" json:"weight"`    // 体重 kg
	Date       time.Time `gorm:"type:date;index;not null" json:"date"`
	AgeStr     string    `gorm:"type:varchar(20)" json:"age_str"`    // 年龄字符串
	AgeInDays  int       `gorm:"type:int" json:"age_in_days"`        // 年龄天数
	Note       string    `gorm:"type:text" json:"note"`             // 备注
	Photo      string    `gorm:"type:varchar(512)" json:"photo"`    // 照片URL
	CreatorID  string    `gorm:"type:varchar(36)" json:"creator_id"` // 创建者ID
}

// TableName 表名
func (Record) TableName() string {
	return "growth_records"
}

// Subscription 订阅模型
type Subscription struct {
	BaseModel            `gorm:"embedded"`
	UserID                string    `gorm:"type:varchar(36);uniqueIndex;not null" json:"user_id"`
	Plan                  string    `gorm:"type:varchar(20);not null" json:"plan"` // monthly/quarterly/yearly
	StartDate             time.Time `gorm:"type:date" json:"start_date"`
	EndDate               time.Time `gorm:"type:date" json:"end_date"`
	AIQuota               int       `gorm:"default:0" json:"ai_quota"`    // AI额度
	AIUsed                int       `gorm:"default:0" json:"ai_used"`     // 已使用次数
	ReferredBy            string    `gorm:"type:varchar(36)" json:"referred_by"` // 推荐人ID
	ReferralCode          string    `gorm:"type:varchar(20);uniqueIndex" json:"referral_code"`
	SubscriptionCancelled bool      `gorm:"default:false" json:"subscription_cancelled"`
	Status                string    `gorm:"type:varchar(20);default:active" json:"status"` // active/expired/cancelled
}

// TableName 表名
func (Subscription) TableName() string {
	return "subscriptions"
}

// IsActive 检查订阅是否有效
func (s *Subscription) IsActive() bool {
	return s.Status == "active" && time.Now().Before(s.EndDate)
}

// GetRemainingQuota 获取剩余AI额度
func (s *Subscription) GetRemainingQuota() int {
	if s.AIQuota == 0 {
		return -1 // 无限
	}
	remaining := s.AIQuota - s.AIUsed
	if remaining < 0 {
		return 0
	}
	return remaining
}

// Family 家庭模型
type Family struct {
	BaseModel  `gorm:"embedded"`
	FamilyID   string    `gorm:"type:varchar(36);uniqueIndex;not null" json:"family_id"`
	CreatorID  string    `gorm:"type:varchar(36);index" json:"creator_id"`
	Name       string    `gorm:"type:varchar(64)" json:"name"`
	InviteCode string    `gorm:"type:varchar(20);uniqueIndex" json:"invite_code"`
	MaxMembers int       `gorm:"default:10" json:"max_members"`
	Members    []FamilyMember `gorm:"foreignKey:FamilyID" json:"members,omitempty"`
	Children   []FamilyChild  `gorm:"foreignKey:FamilyID" json:"children,omitempty"`
}

// TableName 表名
func (Family) TableName() string {
	return "families"
}

// FamilyMember 家庭成员
type FamilyMember struct {
	ID        string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	FamilyID  string    `gorm:"type:varchar(36);index;not null" json:"family_id"`
	UserID    string    `gorm:"type:varchar(36);index;not null" json:"user_id"`
	Name      string    `gorm:"type:varchar(64)" json:"name"`
	Phone     string    `gorm:"type:varchar(20)" json:"phone"`
	Role      string    `gorm:"type:varchar(20);default:viewer" json:"role"` // owner/editor/viewer
	JoinedAt  time.Time `gorm:"autoCreateTime" json:"joined_at"`
}

// TableName 表名
func (FamilyMember) TableName() string {
	return "family_members"
}

// FamilyChild 家庭关联的宝宝
type FamilyChild struct {
	ID        string `gorm:"type:varchar(36);primaryKey" json:"id"`
	FamilyID  string `gorm:"type:varchar(36);index;not null" json:"family_id"`
	ChildID   string `gorm:"type:varchar(36);index;not null" json:"child_id"`
	AddedBy   string `gorm:"type:varchar(36)" json:"added_by"`
	AddedAt   time.Time `gorm:"autoCreateTime" json:"added_at"`
}

// TableName 表名
func (FamilyChild) TableName() string {
	return "family_children"
}

// LabReport 化验单报告
type LabReport struct {
	BaseModel  `gorm:"embedded"`
	ChildID    string    `gorm:"type:varchar(36);index" json:"child_id"`
	UserID     string    `gorm:"type:varchar(36);index" json:"user_id"`
	ImageURL   string    `gorm:"type:varchar(512);not null" json:"image_url"`
	OCRText    string    `gorm:"type:text" json:"ocr_text"`
	AIResult   string    `gorm:"type:text" json:"ai_result"` // JSON字符串
	ReportType string    `gorm:"type:varchar(50)" json:"report_type"`
}

// TableName 表名
func (LabReport) TableName() string {
	return "lab_reports"
}

// AIConversation AI对话
type AIConversation struct {
	BaseModel  `gorm:"embedded"`
	UserID     string    `gorm:"type:varchar(36);index" json:"user_id"`
	ChildID    string    `gorm:"type:varchar(36);index" json:"child_id"`
	SessionID  string    `gorm:"type:varchar(36);index" json:"session_id"`
	Messages   string    `gorm:"type:text" json:"messages"` // JSON数组字符串
	CreatedAt  time.Time `gorm:"index" json:"created_at"`
}

// TableName 表名
func (AIConversation) TableName() string {
	return "ai_conversations"
}
