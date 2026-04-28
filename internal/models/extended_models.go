package models

import (
	"time"
)

// Family 家庭组表
type Family struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	FamilyID   string `gorm:"size:36;uniqueIndex" json:"family_id"`
	Name       string `gorm:"size:50" json:"name"`                  // 家庭名称
	InviteCode string `gorm:"uniqueIndex;size:6" json:"invite_code"` // 6位邀请码
	CreatorID  string `gorm:"size:36" json:"creator_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Members    []FamilyMember `gorm:"foreignKey:FamilyID" json:"members,omitempty"`
}

// TableName 表名
func (Family) TableName() string {
	return "families"
}

// FamilyMember 家庭成员表
type FamilyMember struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	FamilyID  string `gorm:"size:36;index;not null" json:"family_id"`
	UserID    string `gorm:"size:36;index;not null" json:"user_id"`
	Name      string `gorm:"size:64" json:"name"`
	Phone     string `gorm:"size:20" json:"phone"`
	Role      string `gorm:"size:20;default:'viewer'" json:"role"` // creator/editor/viewer
	JoinedAt  time.Time `json:"joined_at"`
}

// TableName 表名
func (FamilyMember) TableName() string {
	return "family_members"
}

// Hospital 医院表
type Hospital struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	Name          string `gorm:"size:200;not null" json:"name"`          // 医院名称
	Level         string `gorm:"size:20" json:"level"`                   // 三级甲等/二级甲等
	Address       string `gorm:"size:500" json:"address"`                // 地址
	Latitude      float64 `json:"latitude"`                              // 纬度
	Longitude     float64 `json:"longitude"`                             // 经度
	Phone         string `gorm:"size:20" json:"phone"`                  // 电话
	Logo          string `gorm:"size:500" json:"logo"`                  // 医院Logo
	PediatricEndo bool   `gorm:"default:true" json:"pediatric_endo"`   // 有儿童内分泌科
	EstimatedFee string `gorm:"size:50" json:"estimated_fee"`          // 预估费用
	City          string `gorm:"size:50" json:"city"`                  // 城市
	District      string `gorm:"size:50" json:"district"`              // 区县
	Departments   []HospitalDepartment `gorm:"foreignKey:HospitalID" json:"departments,omitempty"`
}

// TableName 表名
func (Hospital) TableName() string {
	return "hospitals"
}

// HospitalDepartment 医院科室表
type HospitalDepartment struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	HospitalID  string `gorm:"size:36;index;not null" json:"hospital_id"`
	Name        string `gorm:"size:50;not null" json:"name"`            // 科室名称
	Description string `gorm:"size:500" json:"description"`           // 科室描述
}

// TableName 表名
func (HospitalDepartment) TableName() string {
	return "hospital_departments"
}

// Membership 会员表
type Membership struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"size:36;uniqueIndex;not null" json:"user_id"`
	PlanType  string    `gorm:"size:20" json:"plan_type"`     // monthly/quarterly/yearly
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Status    string    `gorm:"size:20;default:'active'" json:"status"` // active/expired/cancelled
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 表名
func (Membership) TableName() string {
	return "memberships"
}

// UsageQuota 额度使用表
type UsageQuota struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"size:36;index;not null" json:"user_id"`
	Year      int       `json:"year"`
	Month     int       `json:"month"`
	UsedCount int       `gorm:"default:0" json:"used_count"` // 已使用次数
	FreeQuota int       `gorm:"default:3" json:"free_quota"` // 免费额度
	PaidQuota int       `gorm:"default:20" json:"paid_quota"` // 付费额度
}

// TableName 表名
func (UsageQuota) TableName() string {
	return "usage_quotas"
}

// Report 化验单报告表
type Report struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        string    `gorm:"size:36;index;not null" json:"user_id"`
	ChildID       string    `gorm:"size:36;index" json:"child_id"`
	ReportType    string    `gorm:"size:50" json:"report_type"`     // 报告类型
	ImageURL      string    `gorm:"size:500" json:"image_url"`      // 图片URL
	Hospital      string    `gorm:"size:100" json:"hospital"`       // 医院名称
	ReportDate    *time.Time `json:"report_date"`                   // 报告日期
	AnalyzeResult string    `gorm:"type:text" json:"analyze_result"` // 解析结果JSON
	AIResponse    string    `gorm:"type:text" json:"ai_response"`   // AI解读文本
	CreatedAt     time.Time `json:"created_at"`
}

// TableName 表名
func (Report) TableName() string {
	return "reports"
}

// SubscriptionReminder 订阅提醒表
type SubscriptionReminder struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"size:36;index;not null" json:"user_id"`
	ChildID   string    `gorm:"size:36;index" json:"child_id"`
	DayOfWeek int       `json:"day_of_week"` // 0=周日, 1=周一...
	Time      string    `gorm:"size:10;default:'09:00'" json:"time"` // 提醒时间
	Enabled   bool      `gorm:"default:true" json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 表名
func (SubscriptionReminder) TableName() string {
	return "subscription_reminders"
}
