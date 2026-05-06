package models

// API请求/响应结构体

// BaseResponse 基础响应
type BaseResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// PageResponse 分页响应
type PageResponse struct {
	Items    interface{} `json:"items"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// ========== 认证相关 ==========

// LoginRequest 登录请求
type LoginRequest struct {
	Code string `json:"code" binding:"required"` // 微信登录code
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token    string `json:"token"`
	ExpireAt int64  `json:"expire_at"`
	User     *User  `json:"user,omitempty"`
}

// ========== 用户相关 ==========

// UpdateUserRequest 更新用户信息
type UpdateUserRequest struct {
	NickName  string `json:"nick_name"`
	AvatarURL string `json:"avatar_url"`
	Phone     string `json:"phone"`
}

// UpdateUserSettingsRequest 更新设置
type UpdateUserSettingsRequest struct {
	Settings map[string]interface{} `json:"settings"`
}

// ========== 宝宝相关 ==========

// CreateChildRequest 创建宝宝
type CreateChildRequest struct {
	Nickname     string  `json:"nickname" binding:"required,max=64"` // 宝宝昵称，与响应字段统一
	Gender       string  `json:"gender" binding:"required,oneof=male female"`
	Birthday     string  `json:"birthday" binding:"required"`
	FatherHeight float64 `json:"father_height" binding:"required,min=100,max=250"`
	MotherHeight float64 `json:"mother_height" binding:"required,min=100,max=250"`
	Region       *string `json:"region,omitempty"`
}

// UpdateChildRequest 更新宝宝信息
type UpdateChildRequest struct {
	Nickname     string  `json:"nickname"` // 宝宝昵称
	Gender       string  `json:"gender"`
	Birthday     string  `json:"birthday"`
	FatherHeight float64 `json:"father_height"`
	MotherHeight float64 `json:"mother_height"`
	Region       *string `json:"region,omitempty"`
	GrowthStage  *string `json:"growth_stage" binding:"omitempty,oneof=pre_puberty puberty post_puberty"`
}

// SwitchChildRequest 切换当前宝宝
type SwitchChildRequest struct {
	ChildID string `json:"child_id" binding:"required"`
}

// ChildResponse 宝宝响应(包含计算数据)
type ChildResponse struct {
	Child
	AgeStr                string                  `json:"age_str"`
	LatestRecord          *Record                 `json:"latest_record,omitempty"`
	TargetHeight          TargetHeightInfo        `json:"target_height"`
	TargetPercentile      int                     `json:"target_percentile"`
	PotentialStatus       string                  `json:"potential_status"`
	Percentile            int                     `json:"percentile"`
	GrowthStatus          string                  `json:"growth_status"`
	IsNormalRange         bool                    `json:"is_normal_range"`
	InterventionWindow    *InterventionInfo       `json:"intervention_window,omitempty"`
	RegionalCorrection    *RegionalCorrectionInfo `json:"regional_correction,omitempty"`
	AdjustedPercentile    int                     `json:"adjusted_percentile"`
	Alerts                *AlertSummaryResponse   `json:"alerts,omitempty"`
	BoneAgeInfo           *BoneAgeSummary         `json:"bone_age_info,omitempty"`
}

// InterventionInfo 干预窗口信息
type InterventionInfo struct {
	Start         string `json:"start"`
	End           string `json:"end"`
	RemainingDays int    `json:"remaining_days"`
	IsInWindow    bool   `json:"is_in_window"`
}

// RegionalCorrectionInfo 区域修正信息
type RegionalCorrectionInfo struct {
	Region       string  `json:"region"`
	CorrectionCM float64 `json:"correction_cm"`
	AdjustedP50  float64 `json:"adjusted_p50"`
	OriginalP50  float64 `json:"original_p50"`
}

// BoneAgeSummary 骨龄摘要
type BoneAgeSummary struct {
	LatestBoneAge *float64 `json:"latest_bone_age,omitempty"`
	BoneAgeDiff   *float64 `json:"bone_age_diff,omitempty"`
	BoneAgeSource string   `json:"bone_age_source"`
	AssessmentDate string  `json:"assessment_date"`
	IsAbnormal    bool     `json:"is_abnormal"`
}

// AlertResponse 预警响应
type AlertResponse struct {
	HeightAlert
	CreatedAtAgo string `json:"created_at_ago"`
}

// AlertListRequest 预警列表请求
type AlertListRequest struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
	Level    string `form:"level" binding:"omitempty,oneof=info warning danger"`
}

// AlertListResponse 预警列表响应
type AlertListResponse struct {
	Items       []AlertResponse `json:"items"`
	Total       int64           `json:"total"`
	UnreadCount int             `json:"unread_count"`
	ActiveCount int             `json:"active_count"`
}

// AlertSummaryResponse 预警摘要(嵌入首页和详情页)
type AlertSummaryResponse struct {
	HasActiveAlert bool            `json:"has_active_alert"`
	HighestLevel   string          `json:"highest_level"`
	TopAlerts      []AlertResponse `json:"top_alerts"`
	TotalActive    int             `json:"total_active"`
}

// SetGrowthStageRequest 设置生长阶段请求
type SetGrowthStageRequest struct {
	GrowthStage string `json:"growth_stage" binding:"required,oneof=pre_puberty puberty post_puberty"`
	Source      string `json:"source" binding:"required,oneof=self_assessment doctor_visit"`
}

// DismissAlertRequest 忽略预警请求
type DismissAlertRequest struct {
	AlertID string `json:"alert_id" binding:"required"`
	Reason  string `json:"reason"`
}

// ========== 记录相关 ==========

// CreateRecordRequest 创建记录
type CreateRecordRequest struct {
	ChildID string   `json:"child_id" binding:"required"`
	Height  float64  `json:"height" binding:"required,min=30,max=200"`
	Weight  float64  `json:"weight" binding:"omitempty,min=1,max=100"`
	Date    string   `json:"date" binding:"required"`
	Note    string   `json:"note"`
	Photo   string   `json:"photo"`
	BoneAge *float64 `json:"bone_age,omitempty" binding:"omitempty,min=0,max=25"`
}

// UpdateRecordRequest 更新记录
type UpdateRecordRequest struct {
	Height  float64 `json:"height"`
	Weight  float64 `json:"weight"`
	Date    string  `json:"date"`
	Note    string  `json:"note"`
	Photo   string  `json:"photo"`
	BoneAge *float64 `json:"bone_age,omitempty" binding:"omitempty,min=0,max=25"`
}

// RecordResponse 记录响应(含年龄)
type RecordResponse struct {
	GrowthRecord
	AgeStr string `json:"age_str"` // 测量时的年龄: "3岁2个月"
}

// RecordListRequest 记录列表请求
type RecordListRequest struct {
	ChildID   string `form:"child_id"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
}

// ========== 订阅相关 ==========

// CreateOrderRequest 创建订单
type CreateOrderRequest struct {
	Code      string `json:"code" binding:"required"`
	PlanID    string `json:"plan_id" binding:"required,oneof=monthly quarterly yearly"`
	ProductID string `json:"product_id" binding:"required"`
	TotalFee  int    `json:"total_fee" binding:"required,min=1"` // 分
}

// OrderResponse 订单响应
type OrderResponse struct {
	TimeStamp string `json:"timeStamp"`
	NonceStr  string `json:"nonceStr"`
	Package   string `json:"package"`
	SignType  string `json:"signType"`
	PaySign   string `json:"paySign"`
	OrderID   string `json:"order_id"`
}

// PayCallbackRequest 支付回调
type PayCallbackRequest struct {
	XMLName string `xml:"xml"`
}

// SubscriptionResponse 订阅响应
type SubscriptionResponse struct {
	*Subscription
	RemainingQuota int  `json:"remaining_quota"`
	IsActive       bool `json:"is_active"`
	MemberBenefits []MemberBenefit `json:"member_benefits"`
}

// MemberBenefit 会员权益
type MemberBenefit struct {
	Icon string `json:"icon"`
	Text string `json:"text"`
}

// ========== 家庭相关 ==========

// CreateFamilyRequest 创建家庭
type CreateFamilyRequest struct {
	Name string `json:"name"`
}

// JoinFamilyRequest 加入家庭
type JoinFamilyRequest struct {
	InviteCode string `json:"invite_code" binding:"required"`
	Role       string `json:"role" binding:"omitempty,oneof=editor viewer"`
}

// UpdateMemberRoleRequest 更新成员角色
type UpdateMemberRoleRequest struct {
	MemberID string `json:"member_id" binding:"required"`
	Role     string `json:"role" binding:"required,oneof=editor viewer"`
}

// FamilyResponse 家庭响应
type FamilyResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	InviteCode  string `json:"invite_code"`
	MemberCount int    `json:"member_count"`
	ChildCount  int    `json:"child_count"`
}

// GenerateInviteCodeResponse 生成邀请码响应
type GenerateInviteCodeResponse struct {
	InviteCode string `json:"invite_code"`
	ShareURL   string `json:"share_url"`
}

// ========== AI相关 ==========

// AIChatRequest AI对话请求
type AIChatRequest struct {
	ChildID string             `json:"child_id"`
	Message string             `json:"message" binding:"required"`
	Context []AIChatMessage    `json:"context"`
}

// AIChatMessage AI消息
type AIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AIChatResponse AI对话响应
type AIChatResponse struct {
	Response string `json:"response"`
	Tokens   int    `json:"tokens"`
}

// ParseReportRequest 解析化验单请求
type ParseReportRequest struct {
	ChildID    string `json:"child_id" binding:"required"`
	ImageURL   string `json:"image_url" binding:"required"`
	ReportType string `json:"report_type" binding:"required"`
}

// ParseReportResponse 解析化验单响应
type ParseReportResponse struct {
	OCRText  string         `json:"ocr_text"`
	AIResult *AIReportResult `json:"ai_result"`
}

// AIReportResult AI报告解析结果
type AIReportResult struct {
	KeyIndicators []KeyIndicator   `json:"key_indicators"`
	NormalRanges  map[string]string `json:"normal_ranges"`
	Analysis      string            `json:"analysis"`
	Suggestions   []string          `json:"suggestions"`
}

// KeyIndicator 关键指标
type KeyIndicator struct {
	Name   string  `json:"name"`
	Value  string  `json:"value"`
	Status string  `json:"status"` // normal, high, low
}

// ========== 首页数据 ==========

// SubscriptionInfo 订阅信息
type SubscriptionInfo struct {
	IsActive        bool   `json:"is_active"`
	RemainingQuota  int    `json:"remaining_quota"`
	ExpireTime     string `json:"expire_time,omitempty"`
}

// HomeRecordsResponse 首页记录列表
type HomeRecordsResponse struct {
	Items    []RecordResponse `json:"items"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

// HomeDataResponse 首页数据响应
type HomeDataResponse struct {
	Baby         *ChildResponse      `json:"baby,omitempty"`
	HasBaby      bool                `json:"has_baby"`
	Records      *HomeRecordsResponse `json:"records,omitempty"`
	Subscription *SubscriptionInfo   `json:"subscription,omitempty"`
	IsVip        bool                `json:"is_vip"`
	AIRemaining  int                 `json:"ai_remaining"`
}

// ChartData 图表数据
type ChartData struct {
	Categories []string  `json:"categories"`
	Series    []ChartSeries `json:"series"`
}

// ChartSeries 图表系列
type ChartSeries struct {
	Name string    `json:"name"`
	Data []float64 `json:"data"`
}

// ========== 环境问卷评估相关 ==========

// CreateEnvironmentAssessmentRequest 创建环境评估请求
type CreateEnvironmentAssessmentRequest struct {
	// 基础信息（用于计算靶身高和Khamis-Roche）
	CurrentHeight float64 `json:"current_height" binding:"required,min=30,max=200"`
	CurrentWeight float64 `json:"current_weight" binding:"omitempty,min=1,max=150"`

	// 营养模块
	Nutrition NutritionRequest `json:"nutrition" binding:"required"`
	// 睡眠模块
	Sleep SleepRequest `json:"sleep" binding:"required"`
	// 运动模块
	Exercise ExerciseRequest `json:"exercise" binding:"required"`
	// 健康模块
	Health HealthRequest `json:"health" binding:"required"`
	// 心理模块
	Mental MentalRequest `json:"mental" binding:"required"`
}

// NutritionRequest 营养模块请求
type NutritionRequest struct {
	DietDiversity     int `json:"diet_diversity" binding:"min=0,max=2"`      // 膳食多样性 0-2
	ProteinAdequacy   int `json:"protein_adequacy" binding:"min=0,max=2"`    // 蛋白质充足度 0-2
	CalciumIntake     int `json:"calcium_intake" binding:"min=0,max=2"`      // 钙质摄入 0-2
	VitaminDStatus    int `json:"vitamin_d_status" binding:"min=0,max=1"`    // 维生素D 0-1
	BadEatingBehavior int `json:"bad_eating_behavior" binding:"min=0,max=2"` // 不良饮食行为 0-2
	WeightManagement  int `json:"weight_management" binding:"min=0,max=1"`   // 体重管理 0-1
}

// SleepRequest 睡眠模块请求
type SleepRequest struct {
	Duration          float64 `json:"duration" binding:"required,min=0,max=24"` // 睡眠时长(小时)
	BedtimeRegularity int     `json:"bedtime_regularity" binding:"min=0,max=2"` // 入睡规律性 0-2
	DeepSleepCover    int     `json:"deep_sleep_cover" binding:"min=0,max=2"`   // 深睡眠覆盖 0-2
	SleepContinuity   int     `json:"sleep_continuity" binding:"min=0,max=2"`   // 睡眠连续性 0-2
	SleepEnvironment  int     `json:"sleep_environment" binding:"min=0,max=2"`  // 睡眠环境 0-2
}

// ExerciseRequest 运动模块请求
type ExerciseRequest struct {
	Frequency       int `json:"frequency" binding:"min=0,max=3"`       // 运动频率 0-3
	TypeSuitability int `json:"type_suitability" binding:"min=0,max=3"` // 类型适宜性 0-3
	Duration        int `json:"duration" binding:"min=0,max=3"`        // 时长适中性 0-3
	Intensity       int `json:"intensity" binding:"min=0,max=3"`       // 强度分级 0-3
}

// HealthRequest 健康模块请求
type HealthRequest struct {
	DiseaseControl    int `json:"disease_control" binding:"min=0,max=2"`    // 疾病控制 0-2
	CheckupCompliance int `json:"checkup_compliance" binding:"min=0,max=2"` // 体检依从性 0-2
	MedicationSafety  int `json:"medication_safety" binding:"min=0,max=1"`  // 用药安全 0-1
}

// MentalRequest 心理模块请求
type MentalRequest struct {
	EmotionRegulation int `json:"emotion_regulation" binding:"min=0,max=2"` // 情绪调节 0-2
	FamilySupport     int `json:"family_support" binding:"min=0,max=2"`     // 家庭支持 0-2
	StressManagement  int `json:"stress_management" binding:"min=0,max=1"`  // 压力管理 0-1
}

// EnvironmentAssessmentResponse 环境评估响应
type EnvironmentAssessmentResponse struct {
	ID                    string                 `json:"id"`
	ChildID               string                 `json:"child_id"`
	AssessmentDate        string                 `json:"assessment_date"`
	ModuleScores          map[string]ModuleScore `json:"module_scores"`
	TotalScore            float64                `json:"total_score"`
	MaxPossibleScore      float64                `json:"max_possible_score"`
	InterventionZone      string                 `json:"intervention_zone"`
	ZoneLabel             string                 `json:"zone_label"`
	Interpretation        string                 `json:"interpretation"`
	GeneticTargetHeight   float64                `json:"genetic_target_height"`
	EnvironmentIncrement  float64                `json:"environment_increment"`
	PredictedHeight       float64                `json:"predicted_height"`
	PredictionMethod      string                 `json:"prediction_method"`
	ErrorRange            float64                `json:"error_range"`
	KhamisRoche           *KhamisRocheResult     `json:"khamis_roche,omitempty"`
	AgeWeights            AgeLayeredWeights      `json:"age_weights"`
	ActionPlan            *WeeklyActionPlan      `json:"action_plan,omitempty"`
	ClinicalInterpretation string                `json:"clinical_interpretation"`
}

// WeeklyActionPlan 个性化行动计划
type WeeklyActionPlan struct {
	TopPriorities []ActionPlanItem `json:"top_priorities"` // 本周3个最优先行动
	NutritionPlan []string         `json:"nutrition_plan"`
	SleepPlan     []string         `json:"sleep_plan"`
	ExercisePlan  []string         `json:"exercise_plan"`
	HealthPlan    []string         `json:"health_plan"`
	MentalPlan    []string         `json:"mental_plan"`
	TrackReminder string           `json:"track_reminder"`
}

// ActionPlanItem 行动计划项
type ActionPlanItem struct {
	Priority    int    `json:"priority"`     // 1=第1优先, 2=第2优先, 3=第3优先
	Module      string `json:"module"`       // 所属模块
	Title       string `json:"title"`        // 行动标题
	Description string `json:"description"`  // 具体说明
	Why         string `json:"why"`          // 为什么重要
	HowToStart  string `json:"how_to_start"` // 如何开始
	Difficulty  string `json:"difficulty"`   // easy/medium/hard
}

// EnvironmentAssessmentHistoryResponse 评估历史响应
type EnvironmentAssessmentHistoryResponse struct {
	Items    []EnvironmentAssessmentSummary `json:"items"`
	Total    int64                          `json:"total"`
	Page     int                            `json:"page"`
	PageSize int                            `json:"page_size"`
}

// EnvironmentAssessmentSummary 评估摘要（列表用）
type EnvironmentAssessmentSummary struct {
	ID                   string  `json:"id"`
	AssessmentDate       string  `json:"assessment_date"`
	TotalScore           float64 `json:"total_score"`
	InterventionZone     string  `json:"intervention_zone"`
	PredictedHeight      float64 `json:"predicted_height"`
	EnvironmentIncrement float64 `json:"environment_increment"`
}

// ========== 生长速度监测相关 ==========

// GrowthVelocityResponse 生长速度响应
type GrowthVelocityResponse struct {
	Velocity       float64 `json:"velocity"`        // 年生长速度 cm/年
	MonthsBack     int     `json:"months_back"`     // 计算使用的月数
	LatestHeight   float64 `json:"latest_height"`
	PreviousHeight float64 `json:"previous_height"`
	ExpectedMin    float64 `json:"expected_min"`    // 该年龄段最低期望
	Status         string  `json:"status"`          // 正常/偏慢/过慢
	AlertLevel     string  `json:"alert_level"`     // yellow/orange/red/空
	Action         string  `json:"action"`          // 建议行动
	Deviation      float64 `json:"deviation"`       // 与期望值的偏差
}

// ========== 靶身高对比相关 ==========

// TargetHeightComparisonResponse 靶身高对比响应
type TargetHeightComparisonResponse struct {
	GeneticTargetHeight   float64                `json:"genetic_target_height"`    // MPH
	QuantitativeGenetics  *QuantitativeGeneticsResult `json:"quantitative_genetics,omitempty"`
	KhamisRoche           *KhamisRocheResult     `json:"khamis_roche,omitempty"`
	EnvironmentPrediction *ComprehensivePredictionResult `json:"environment_prediction,omitempty"`
	CurrentHeight         float64                `json:"current_height"`
	CurrentPercentile     int                    `json:"current_percentile"`
	TargetPercentile      int                    `json:"target_percentile"`       // 当前身高在靶身高范围中的位置
	PotentialStatus       string                 `json:"potential_status"`        // 遗传潜力达成状态
	GrowthVelocity        *GrowthVelocityResponse `json:"growth_velocity,omitempty"`
}

// ========== 错误码 ==========

const (
	CodeSuccess           = 0
	CodeParamError        = 400
	CodeUnauthorized      = 401
	CodeForbidden         = 403
	CodeNotFound          = 404
	CodeServerError       = 500
	CodeQuotaExhausted    = 1001
	CodeNotVip            = 1002
	CodeInvalidInviteCode = 2001
	CodeFamilyFull        = 2002
)
