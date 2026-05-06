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

// Child 宝宝模型
type Child struct {
	BaseModel    `gorm:"embedded"`
	UserID       string    `gorm:"type:varchar(36);index;not null" json:"user_id"`
	FamilyID     *uint     `gorm:"index" json:"family_id,omitempty"`
	Nickname     string    `gorm:"type:varchar(50);not null" json:"nickname"`
	Gender       string    `gorm:"type:varchar(10);not null" json:"gender"` // male/female
	Birthday     time.Time `gorm:"type:date;not null" json:"birthday"`
	InitialHeight float64  `gorm:"type:decimal(5,1)" json:"initial_height"`
	InitialWeight *float64 `gorm:"type:decimal(5,1)" json:"initial_weight"`
	FatherHeight float64   `gorm:"type:decimal(5,1)" json:"father_height"`
	MotherHeight float64   `gorm:"type:decimal(5,1)" json:"mother_height"`
	StandardType string    `gorm:"type:varchar(10);default:'cn'" json:"standard_type"`
	Region       *string   `gorm:"type:varchar(20)" json:"region,omitempty"`
	GrowthStage  *string   `gorm:"type:varchar(20);default:''" json:"growth_stage,omitempty"`
	StageConfirmedAt *time.Time `gorm:"type:datetime" json:"stage_confirmed_at,omitempty"`
	LastHeightChangeDate *time.Time `gorm:"type:date" json:"last_height_change_date,omitempty"`
	Records      []GrowthRecord `gorm:"foreignKey:ChildID" json:"records,omitempty"`
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
	months = int(at.Month()) - int(c.Birthday.Month())

	if months < 0 {
		years--
		months += 12
	}

	if at.Day() < c.Birthday.Day() {
		months--
		if months < 0 {
			years--
			months = 11
		}
	}

	if years < 0 {
		years = 0
		months = 0
	}
	return
}

// GrowthRecord 生长记录模型 (兼容 Record)
type GrowthRecord struct {
	BaseModel  `gorm:"embedded"`
	ChildID    string    `gorm:"type:varchar(36);index;uniqueIndex:idx_child_date;not null" json:"child_id"`
	MeasureDate time.Time `gorm:"type:date;index:idx_measure_date;uniqueIndex:idx_child_date;not null" json:"measure_date"`
	Height     float64   `gorm:"type:decimal(5,1);not null" json:"height"`     // 身高 cm
	Weight     *float64  `gorm:"type:decimal(5,1)" json:"weight"`             // 体重 kg
	HeightPercentile *float64 `gorm:"type:decimal(5,2)" json:"height_percentile"`
	WeightPercentile *float64 `gorm:"type:decimal(5,2)" json:"weight_percentile"`
	HeightZScore     *float64 `gorm:"type:decimal(5,3)" json:"height_zscore"`
	WeightZScore     *float64 `gorm:"type:decimal(5,3)" json:"weight_zscore"`
	HeightStatus     string   `gorm:"type:varchar(20);default:'normal'" json:"height_status"`
	WeightStatus     string   `gorm:"type:varchar(20);default:'normal'" json:"weight_status"`
	BoneAge         *float64 `gorm:"type:decimal(5,2)" json:"bone_age,omitempty"`
	BoneAgeSource   *string  `gorm:"type:varchar(30)" json:"bone_age_source,omitempty"`
	BoneAgeDiff     *float64 `gorm:"type:decimal(5,2)" json:"bone_age_diff,omitempty"`
	Remarks         string   `gorm:"type:text" json:"remarks"`
}

// TableName 表名
func (GrowthRecord) TableName() string {
	return "growth_records"
}

// Record 类型别名，保持向后兼容
type Record = GrowthRecord

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

// FamilyChild 家庭关联的宝宝
type FamilyChild struct {
	ID        string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	FamilyID  string    `gorm:"type:varchar(36);index;not null" json:"family_id"`
	ChildID   string    `gorm:"type:varchar(36);index;not null" json:"child_id"`
	AddedBy   string    `gorm:"type:varchar(36)" json:"added_by"`
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

// ========== 预警相关常量 ==========
const (
	AlertTargetGapLow    = "target_gap_low"
	AlertRegionalShort   = "regional_short"
	AlertBoneAgeAdvanced = "bone_age_advanced"
	AlertBoneAgeDelayed  = "bone_age_delayed"
	AlertStagnation      = "growth_stagnation"
	AlertVelocitySlow    = "velocity_slow"
	AlertPercentileDrop  = "percentile_drop"
)

// HeightAlert 身高异常预警记录
type HeightAlert struct {
	BaseModel        `gorm:"embedded"`
	ChildID          string     `gorm:"type:varchar(36);index;not null" json:"child_id"`
	UserID           string     `gorm:"type:varchar(36);index" json:"user_id"`
	AlertType        string     `gorm:"type:varchar(50);not null;index" json:"alert_type"`
	AlertLevel       string     `gorm:"type:varchar(20);not null" json:"alert_level"`
	Title            string     `gorm:"type:varchar(200);not null" json:"title"`
	Description      string     `gorm:"type:text" json:"description"`
	Dimension        string     `gorm:"type:varchar(30);not null" json:"dimension"`
	MetricValue      *float64   `gorm:"type:decimal(8,3)" json:"metric_value,omitempty"`
	Threshold        *float64   `gorm:"type:decimal(8,3)" json:"threshold,omitempty"`
	TriggerRecordID  *string    `gorm:"type:varchar(36)" json:"trigger_record_id,omitempty"`
	IsRead           bool       `gorm:"default:false" json:"is_read"`
	IsDismissed      bool       `gorm:"default:false" json:"is_dismissed"`
	ResolvedAt       *time.Time `gorm:"type:datetime" json:"resolved_at,omitempty"`
}

func (HeightAlert) TableName() string {
	return "height_alerts"
}

// ========== 环境问卷评估记录 ==========

// EnvironmentAssessment 环境问卷评估记录（存储用户提交的问卷和计算结果）
type EnvironmentAssessment struct {
	BaseModel    `gorm:"embedded"`
	ChildID      string  `gorm:"type:varchar(36);index;not null" json:"child_id"`
	UserID       string  `gorm:"type:varchar(36);index;not null" json:"user_id"`
	AssessmentDate time.Time `gorm:"type:date;not null" json:"assessment_date"`

	// 营养模块原始答案 (JSON存储)
	NutritionRaw string `gorm:"type:text" json:"-"` // 原始JSON

	// 睡眠模块原始答案
	SleepRaw string `gorm:"type:text" json:"-"`

	// 运动模块原始答案
	ExerciseRaw string `gorm:"type:text" json:"-"`

	// 健康模块原始答案
	HealthRaw string `gorm:"type:text" json:"-"`

	// 心理模块原始答案
	MentalRaw string `gorm:"type:text" json:"-"`

	// 计算得分
	NutritionScore  float64 `gorm:"type:decimal(5,2)" json:"nutrition_score"`
	SleepScore      float64 `gorm:"type:decimal(5,2)" json:"sleep_score"`
	ExerciseScore   float64 `gorm:"type:decimal(5,2)" json:"exercise_score"`
	HealthScore     float64 `gorm:"type:decimal(5,2)" json:"health_score"`
	MentalScore     float64 `gorm:"type:decimal(5,2)" json:"mental_score"`
	TotalScore      float64 `gorm:"type:decimal(5,2)" json:"total_score"`

	// 预测结果
	GeneticTargetHeight float64 `gorm:"type:decimal(5,1)" json:"genetic_target_height"`
	EnvironmentIncrement float64 `gorm:"type:decimal(5,1)" json:"environment_increment"`
	PredictedHeight     float64 `gorm:"type:decimal(5,1)" json:"predicted_height"`

	// 分区
	InterventionZone string `gorm:"type:varchar(20)" json:"intervention_zone"` // high/medium/low

	// 行动计划 (JSON存储)
	ActionPlan string `gorm:"type:text" json:"-"`
}

func (EnvironmentAssessment) TableName() string {
	return "environment_assessments"
}
