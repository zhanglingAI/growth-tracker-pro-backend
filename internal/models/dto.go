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
}

// UpdateChildRequest 更新宝宝信息
type UpdateChildRequest struct {
	Nickname     string  `json:"nickname"` // 宝宝昵称
	Gender       string  `json:"gender"`
	Birthday     string  `json:"birthday"`
	FatherHeight float64 `json:"father_height"`
	MotherHeight float64 `json:"mother_height"`
}

// SwitchChildRequest 切换当前宝宝
type SwitchChildRequest struct {
	ChildID string `json:"child_id" binding:"required"`
}

// ChildResponse 宝宝响应(包含计算数据)
type ChildResponse struct {
	Child
	AgeStr           string             `json:"age_str"`
	LatestRecord     *Record            `json:"latest_record,omitempty"`
	TargetHeight     TargetHeightInfo   `json:"target_height"`
	Percentile       int                `json:"percentile"`
	GrowthStatus     string             `json:"growth_status"`
	InterventionWindow *InterventionInfo `json:"intervention_window,omitempty"`
}

// TargetHeightInfo 靶身高信息
type TargetHeightInfo struct {
	TargetHeight   float64 `json:"target_height"`
	MinHeight      float64 `json:"min_height"`
	MaxHeight      float64 `json:"max_height"`
}

// InterventionInfo 干预窗口信息
type InterventionInfo struct {
	Start         string `json:"start"`
	End           string `json:"end"`
	RemainingDays int    `json:"remaining_days"`
	IsInWindow    bool   `json:"is_in_window"`
}

// ========== 记录相关 ==========

// CreateRecordRequest 创建记录
type CreateRecordRequest struct {
	ChildID string  `json:"child_id" binding:"required"`
	Height  float64 `json:"height" binding:"required,min=30,max=200"`
	Weight  float64 `json:"weight" binding:"required,min=1,max=100"`
	Date    string  `json:"date" binding:"required"`
	Note    string  `json:"note"`
	Photo   string  `json:"photo"`
}

// UpdateRecordRequest 更新记录
type UpdateRecordRequest struct {
	Height float64 `json:"height"`
	Weight float64 `json:"weight"`
	Date   string  `json:"date"`
	Note   string  `json:"note"`
	Photo  string  `json:"photo"`
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
