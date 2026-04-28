package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/growth-tracker-pro/backend/internal/config"
	"github.com/growth-tracker-pro/backend/internal/models"
	"github.com/growth-tracker-pro/backend/internal/repository"
	"github.com/google/uuid"
)

// Service 服务接口
type Service interface {
	// 用户服务
	Login(ctx context.Context, code string) (*models.LoginResponse, error)
	GetUserInfo(ctx context.Context, userID string) (*models.User, error)
	UpdateUser(ctx context.Context, userID string, req *models.UpdateUserRequest) error

	// 宝宝服务
	GetChildren(ctx context.Context, userID string) ([]models.Child, error)
	GetActiveChild(ctx context.Context, userID string) (*models.Child, error)
	CreateChild(ctx context.Context, userID string, req *models.CreateChildRequest) (*models.Child, error)
	UpdateChild(ctx context.Context, userID, childID string, req *models.UpdateChildRequest) error
	DeleteChild(ctx context.Context, userID, childID string) error
	SwitchChild(ctx context.Context, userID, childID string) error
	GetChildDetail(ctx context.Context, userID, childID string) (*models.ChildResponse, error)

	// 记录服务
	GetRecords(ctx context.Context, childID string, req *models.RecordListRequest) (*models.PageResponse, error)
	CreateRecord(ctx context.Context, userID string, req *models.CreateRecordRequest) (*models.Record, error)
	UpdateRecord(ctx context.Context, userID, recordID string, req *models.UpdateRecordRequest) error
	DeleteRecord(ctx context.Context, userID, recordID string) error

	// 订阅服务
	GetSubscription(ctx context.Context, userID string) (*models.SubscriptionResponse, error)
	CreateOrder(ctx context.Context, userID string, req *models.CreateOrderRequest) (*models.OrderResponse, error)
	ProcessPayCallback(ctx context.Context, data map[string]string) error
	UpgradeToVip(ctx context.Context, userID, plan string) error

	// 家庭服务
	GetFamily(ctx context.Context, userID string) (*models.FamilyResponse, error)
	CreateFamily(ctx context.Context, userID string, req *models.CreateFamilyRequest) (*models.Family, error)
	JoinFamily(ctx context.Context, userID string, req *models.JoinFamilyRequest) error
	LeaveFamily(ctx context.Context, userID string) error
	UpdateMemberRole(ctx context.Context, userID, memberID, role string) error
	GenerateInviteCode(ctx context.Context, userID string) (*models.GenerateInviteCodeResponse, error)

	// AI服务
	Chat(ctx context.Context, userID string, req *models.AIChatRequest) (*models.AIChatResponse, error)
	ParseReport(ctx context.Context, userID string, req *models.ParseReportRequest) (*models.ParseReportResponse, error)

	// 首页数据
	GetHomeData(ctx context.Context, userID string) (*models.HomeDataResponse, error)

	// 靶身高计算
	CalculateTargetHeight(child *models.Child) models.TargetHeightInfo
	// 干预窗口计算
	CalculateInterventionWindow(child *models.Child) *models.InterventionInfo
	// 百分位计算
	CalculatePercentile(record *models.Record, child *models.Child) int
	// 生长状态判断
	DetermineGrowthStatus(percentile int, targetMin, targetMax, currentHeight float64) string
}

// ServiceImpl 服务实现
type ServiceImpl struct {
	repo   repository.Repository
	cache  *repository.RedisCache
	config *config.Config
}

// NewService 创建服务
func NewService(repo repository.Repository, cache *repository.RedisCache, cfg *config.Config) Service {
	return &ServiceImpl{
		repo:   repo,
		cache:  cache,
		config: cfg,
	}
}

// ========== 用户服务 ==========

func (s *ServiceImpl) Login(ctx context.Context, code string) (*models.LoginResponse, error) {
	// 微信登录 - 获取openid
	openID, err := s.getOpenIDFromWx(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("微信登录失败: %w", err)
	}

	// 查询或创建用户
	user, err := s.repo.GetUserByOpenID(ctx, openID)
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	isNew := false
	if user == nil {
		isNew = true
		user = &models.User{
			OpenID: openID,
		}
		if err := s.repo.CreateUser(ctx, user); err != nil {
			return nil, fmt.Errorf("创建用户失败: %w", err)
		}
	}

	// 生成JWT token
	token, expireAt, err := s.generateToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("生成token失败: %w", err)
	}

	// 如果是新用户，生成邀请码
	if isNew && user.Subscription == nil {
		user.Subscription = &models.Subscription{
			UserID:       user.ID,
			ReferralCode: s.generateReferralCode(),
			Status:       "active",
			AIQuota:      3, // 免费用户3次
		}
		if err := s.repo.CreateSubscription(ctx, user.Subscription); err != nil {
			// 不影响登录
		}
	}

	return &models.LoginResponse{
		Token:    token,
		ExpireAt: expireAt,
		User:     user,
	}, nil
}

func (s *ServiceImpl) GetUserInfo(ctx context.Context, userID string) (*models.User, error) {
	return s.repo.GetUserByID(ctx, userID)
}

func (s *ServiceImpl) UpdateUser(ctx context.Context, userID string, req *models.UpdateUserRequest) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil || user == nil {
		return fmt.Errorf("用户不存在")
	}

	if req.NickName != "" {
		user.NickName = req.NickName
	}
	if req.AvatarURL != "" {
		user.AvatarURL = req.AvatarURL
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}

	return s.repo.UpdateUser(ctx, user)
}

// ========== 宝宝服务 ==========

func (s *ServiceImpl) GetChildren(ctx context.Context, userID string) ([]models.Child, error) {
	return s.repo.GetChildrenByUserID(ctx, userID)
}

func (s *ServiceImpl) GetActiveChild(ctx context.Context, userID string) (*models.Child, error) {
	return s.repo.GetActiveChild(ctx, userID)
}

func (s *ServiceImpl) CreateChild(ctx context.Context, userID string, req *models.CreateChildRequest) (*models.Child, error) {
	birthday, err := time.Parse("2006-01-02", req.Birthday)
	if err != nil {
		return nil, fmt.Errorf("无效的出生日期格式")
	}

	// 如果是第一个宝宝，自动设为激活
	children, _ := s.repo.GetChildrenByUserID(ctx, userID)
	isActive := len(children) == 0

	child := &models.Child{
		UserID:       userID,
		Name:         req.Name,
		Gender:       req.Gender,
		Birthday:     birthday,
		FatherHeight: req.FatherHeight,
		MotherHeight: req.MotherHeight,
		IsActive:     isActive,
	}

	if err := s.repo.CreateChild(ctx, child); err != nil {
		return nil, fmt.Errorf("创建宝宝失败: %w", err)
	}

	return child, nil
}

func (s *ServiceImpl) UpdateChild(ctx context.Context, userID, childID string, req *models.UpdateChildRequest) error {
	child, err := s.repo.GetChildByID(ctx, childID)
	if err != nil || child == nil || child.UserID != userID {
		return fmt.Errorf("宝宝不存在")
	}

	if req.Name != "" {
		child.Name = req.Name
	}
	if req.Gender != "" {
		child.Gender = req.Gender
	}
	if req.Birthday != "" {
		birthday, err := time.Parse("2006-01-02", req.Birthday)
		if err == nil {
			child.Birthday = birthday
		}
	}
	if req.FatherHeight > 0 {
		child.FatherHeight = req.FatherHeight
	}
	if req.MotherHeight > 0 {
		child.MotherHeight = req.MotherHeight
	}

	return s.repo.UpdateChild(ctx, child)
}

func (s *ServiceImpl) DeleteChild(ctx context.Context, userID, childID string) error {
	child, err := s.repo.GetChildByID(ctx, childID)
	if err != nil || child == nil || child.UserID != userID {
		return fmt.Errorf("宝宝不存在")
	}

	return s.repo.DeleteChild(ctx, childID)
}

func (s *ServiceImpl) SwitchChild(ctx context.Context, userID, childID string) error {
	child, err := s.repo.GetChildByID(ctx, childID)
	if err != nil || child == nil || child.UserID != userID {
		return fmt.Errorf("宝宝不存在")
	}

	return s.repo.SetActiveChild(ctx, userID, childID)
}

func (s *ServiceImpl) GetChildDetail(ctx context.Context, userID, childID string) (*models.ChildResponse, error) {
	child, err := s.repo.GetChildByID(ctx, childID)
	if err != nil || child == nil || child.UserID != userID {
		return nil, fmt.Errorf("宝宝不存在")
	}

	response := &models.ChildResponse{
		Child:         *child,
		TargetHeight:  s.CalculateTargetHeight(child),
	}

	// 计算年龄
	now := time.Now()
	years, months := child.CalculateAge(now)
	response.AgeStr = fmt.Sprintf("%d岁%d月", years, months)

	// 获取最新记录
	records, _, _ := s.repo.GetRecordsByChildID(ctx, childID, "", "", 1, 1)
	if len(records) > 0 {
		response.LatestRecord = &records[0]
		response.Percentile = s.CalculatePercentile(&records[0], child)
		response.GrowthStatus = s.DetermineGrowthStatus(response.Percentile, response.TargetHeight.MinHeight, response.TargetHeight.MaxHeight, records[0].Height)
	}

	// 计算干预窗口
	response.InterventionWindow = s.CalculateInterventionWindow(child)

	return response, nil
}

// ========== 记录服务 ==========

func (s *ServiceImpl) GetRecords(ctx context.Context, childID string, req *models.RecordListRequest) (*models.PageResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	records, total, err := s.repo.GetRecordsByChildID(ctx, childID, req.StartDate, req.EndDate, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	return &models.PageResponse{
		Items:    records,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (s *ServiceImpl) CreateRecord(ctx context.Context, userID string, req *models.CreateRecordRequest) (*models.Record, error) {
	child, err := s.repo.GetChildByID(ctx, req.ChildID)
	if err != nil || child == nil {
		return nil, fmt.Errorf("宝宝不存在")
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, fmt.Errorf("无效的日期格式")
	}

	// 计算年龄
	ageInDays := int(date.Sub(child.Birthday).Hours() / 24)
	years := ageInDays / 365
	months := (ageInDays % 365) / 30
	ageStr := fmt.Sprintf("%d.%d", years, months)

	record := &models.Record{
		ChildID:   req.ChildID,
		Height:    req.Height,
		Weight:    req.Weight,
		Date:      date,
		AgeStr:    ageStr,
		AgeInDays: ageInDays,
		Note:      req.Note,
		Photo:     req.Photo,
		CreatorID: userID,
	}

	if err := s.repo.CreateRecord(ctx, record); err != nil {
		return nil, fmt.Errorf("创建记录失败: %w", err)
	}

	return record, nil
}

func (s *ServiceImpl) UpdateRecord(ctx context.Context, userID, recordID string, req *models.UpdateRecordRequest) error {
	record, err := s.repo.GetRecordByID(ctx, recordID)
	if err != nil || record == nil {
		return fmt.Errorf("记录不存在")
	}

	if req.Height > 0 {
		record.Height = req.Height
	}
	if req.Weight > 0 {
		record.Weight = req.Weight
	}
	if req.Date != "" {
		date, err := time.Parse("2006-01-02", req.Date)
		if err == nil {
			record.Date = date
		}
	}
	if req.Note != "" {
		record.Note = req.Note
	}
	if req.Photo != "" {
		record.Photo = req.Photo
	}

	return s.repo.UpdateRecord(ctx, record)
}

func (s *ServiceImpl) DeleteRecord(ctx context.Context, userID, recordID string) error {
	record, err := s.repo.GetRecordByID(ctx, recordID)
	if err != nil || record == nil {
		return fmt.Errorf("记录不存在")
	}

	return s.repo.DeleteRecord(ctx, recordID)
}

// ========== 订阅服务 ==========

func (s *ServiceImpl) GetSubscription(ctx context.Context, userID string) (*models.SubscriptionResponse, error) {
	sub, err := s.repo.GetSubscriptionByUserID(ctx, userID)
	if err != nil || sub == nil {
		return &models.SubscriptionResponse{
			Subscription: &models.Subscription{
				Status:   "inactive",
				AIQuota:  3,
				AIUsed:   0,
			},
			RemainingQuota: 3,
			IsActive:       false,
			MemberBenefits: []models.MemberBenefit{},
		}, nil
	}

	benefits := s.getMemberBenefits(sub)

	return &models.SubscriptionResponse{
		Subscription:   sub,
		RemainingQuota: sub.GetRemainingQuota(),
		IsActive:       sub.IsActive(),
		MemberBenefits: benefits,
	}, nil
}

func (s *ServiceImpl) CreateOrder(ctx context.Context, userID string, req *models.CreateOrderRequest) (*models.OrderResponse, error) {
	// 生成订单号
	orderID := fmt.Sprintf("GT%d%s", time.Now().Unix(), uuid.New().String()[:8])

	// 获取价格配置
	fee := s.getPlanPrice(req.PlanID)
	if req.TotalFee != fee {
		req.TotalFee = fee
	}

	// 微信支付统一下单 (这里简化处理，实际需要调用微信支付API)
	// 实际实现需要调用微信支付接口
	payParams := &models.OrderResponse{
		TimeStamp: fmt.Sprintf("%d", time.Now().Unix()),
		NonceStr:  uuid.New().String(),
		Package:   fmt.Sprintf("prepay_id=%s", orderID),
		SignType:  "MD5",
		PaySign:   "", // 实际需要签名
		OrderID:   orderID,
	}

	// 缓存订单信息
	orderKey := fmt.Sprintf("order:%s", orderID)
	orderData := map[string]interface{}{
		"user_id":    userID,
		"plan_id":    req.PlanID,
		"total_fee":  req.TotalFee,
		"status":     "pending",
		"created_at": time.Now(),
	}
	s.cache.Set(ctx, orderKey, orderData, 2*time.Hour)

	return payParams, nil
}

func (s *ServiceImpl) ProcessPayCallback(ctx context.Context, data map[string]string) error {
	// 验证签名 (实际需要完整验证)
	// transactionID := data["transaction_id"]
	orderID := data["out_trade_no"]
	// resultCode := data["result_code"]

	// 更新订单状态
	orderKey := fmt.Sprintf("order:%s", orderID)
	var orderData map[string]interface{}
	if err := s.cache.Get(ctx, orderKey, &orderData); err != nil {
		return err
	}

	userID := orderData["user_id"].(string)
	planID := orderData["plan_id"].(string)

	// 升级用户为VIP
	return s.UpgradeToVip(ctx, userID, planID)
}

func (s *ServiceImpl) UpgradeToVip(ctx context.Context, userID, plan string) error {
	now := time.Now()
	var startDate, endDate time.Time
	var aiQuota int

	switch plan {
	case "monthly":
		startDate = now
		endDate = now.AddDate(0, 1, 0)
		aiQuota = 30
	case "quarterly":
		startDate = now
		endDate = now.AddDate(0, 3, 0)
		aiQuota = 100
	case "yearly":
		startDate = now
		endDate = now.AddDate(1, 0, 0)
		aiQuota = 400
	default:
		return fmt.Errorf("无效的订阅方案")
	}

	sub := &models.Subscription{
		UserID:       userID,
		Plan:         plan,
		StartDate:    startDate,
		EndDate:      endDate,
		AIQuota:      aiQuota,
		AIUsed:       0,
		Status:       "active",
		ReferralCode: s.generateReferralCode(),
	}

	existing, _ := s.repo.GetSubscriptionByUserID(ctx, userID)
	if existing != nil {
		sub.ID = existing.ID
		// 累加时间
		if existing.IsActive() {
			sub.StartDate = existing.StartDate
			sub.EndDate = existing.EndDate.Add(endDate.Sub(startDate))
		}
	}

	if existing == nil {
		return s.repo.CreateSubscription(ctx, sub)
	}
	return s.repo.UpdateSubscription(ctx, sub)
}

// ========== 家庭服务 ==========

func (s *ServiceImpl) GetFamily(ctx context.Context, userID string) (*models.FamilyResponse, error) {
	family, err := s.repo.GetFamilyByUserID(ctx, userID)
	if err != nil || family == nil {
		return nil, fmt.Errorf("未找到家庭")
	}

	return &models.FamilyResponse{
		Family:      *family,
		MemberCount: len(family.Members),
		ChildCount:  len(family.Children),
	}, nil
}

func (s *ServiceImpl) CreateFamily(ctx context.Context, userID string, req *models.CreateFamilyRequest) (*models.Family, error) {
	// 检查是否已有家庭
	existing, _ := s.repo.GetFamilyByUserID(ctx, userID)
	if existing != nil {
		return nil, fmt.Errorf("您已在家庭中")
	}

	family := &models.Family{
		FamilyID:   uuid.New().String(),
		CreatorID:  userID,
		Name:       req.Name,
		InviteCode: s.generateInviteCode(),
		MaxMembers: 10,
	}

	// 添加创建者为owner
	member := &models.FamilyMember{
		ID:       uuid.New().String(),
		FamilyID: family.FamilyID,
		UserID:   userID,
		Role:     "owner",
	}
	family.Members = append(family.Members, *member)

	if err := s.repo.CreateFamily(ctx, family); err != nil {
		return nil, fmt.Errorf("创建家庭失败: %w", err)
	}

	// 更新用户的家庭ID
	user, _ := s.repo.GetUserByID(ctx, userID)
	if user != nil {
		user.FamilyID = &family.FamilyID
		s.repo.UpdateUser(ctx, user)
	}

	return family, nil
}

func (s *ServiceImpl) JoinFamily(ctx context.Context, userID string, req *models.JoinFamilyRequest) error {
	// 检查是否已有家庭
	existing, _ := s.repo.GetFamilyByUserID(ctx, userID)
	if existing != nil {
		return fmt.Errorf("您已在家庭中")
	}

	family, err := s.repo.GetFamilyByInviteCode(ctx, req.InviteCode)
	if err != nil || family == nil {
		return fmt.Errorf("邀请码无效")
	}

	// 检查是否满员
	if len(family.Members) >= family.MaxMembers {
		return fmt.Errorf("家庭成员已满")
	}

	role := "viewer"
	if req.Role != "" {
		role = req.Role
	}

	member := &models.FamilyMember{
		ID:       uuid.New().String(),
		FamilyID: family.FamilyID,
		UserID:   userID,
		Role:     role,
	}

	if err := s.repo.AddFamilyMember(ctx, member); err != nil {
		return fmt.Errorf("加入家庭失败: %w", err)
	}

	// 更新用户的家庭ID
	user, _ := s.repo.GetUserByID(ctx, userID)
	if user != nil {
		user.FamilyID = &family.FamilyID
		s.repo.UpdateUser(ctx, user)
	}

	return nil
}

func (s *ServiceImpl) LeaveFamily(ctx context.Context, userID string) error {
	family, err := s.repo.GetFamilyByUserID(ctx, userID)
	if err != nil || family == nil {
		return fmt.Errorf("未找到家庭")
	}

	// 找到当前用户的成员
	var memberID string
	for _, m := range family.Members {
		if m.UserID == userID {
			memberID = m.ID
			break
		}
	}

	if memberID == "" {
		return fmt.Errorf("未找到成员")
	}

	// owner不能退出，只能删除家庭
	for _, m := range family.Members {
		if m.UserID == userID && m.Role == "owner" {
			return fmt.Errorf("管理员不能退出，请先删除家庭")
		}
	}

	if err := s.repo.RemoveFamilyMember(ctx, family.FamilyID, memberID); err != nil {
		return fmt.Errorf("退出家庭失败: %w", err)
	}

	// 更新用户的家庭ID
	user, _ := s.repo.GetUserByID(ctx, userID)
	if user != nil {
		user.FamilyID = nil
		s.repo.UpdateUser(ctx, user)
	}

	return nil
}

func (s *ServiceImpl) UpdateMemberRole(ctx context.Context, userID, memberID, role string) error {
	// 验证权限 (owner才能修改角色)
	family, err := s.repo.GetFamilyByUserID(ctx, userID)
	if err != nil || family == nil {
		return fmt.Errorf("未找到家庭")
	}

	isOwner := false
	for _, m := range family.Members {
		if m.UserID == userID && m.Role == "owner" {
			isOwner = true
			break
		}
	}

	if !isOwner {
		return fmt.Errorf("无权限修改成员角色")
	}

	return s.repo.UpdateMemberRole(ctx, memberID, role)
}

func (s *ServiceImpl) GenerateInviteCode(ctx context.Context, userID string) (*models.GenerateInviteCodeResponse, error) {
	family, err := s.repo.GetFamilyByUserID(ctx, userID)
	if err != nil || family == nil {
		return nil, fmt.Errorf("未找到家庭")
	}

	return &models.GenerateInviteCodeResponse{
		InviteCode: family.InviteCode,
		ShareURL:   fmt.Sprintf("pages/family/join?code=%s", family.InviteCode),
	}, nil
}

// ========== AI服务 ==========

func (s *ServiceImpl) Chat(ctx context.Context, userID string, req *models.AIChatRequest) (*models.AIChatResponse, error) {
	// 检查配额
	sub, _ := s.repo.GetSubscriptionByUserID(ctx, userID)
	isVip := sub != nil && sub.IsActive()

	if !isVip {
		if sub == nil || sub.AIUsed >= sub.AIQuota {
			return nil, fmt.Errorf("AI额度已用完，请升级会员")
		}
		// 增加使用次数
		s.repo.IncrementAIUsage(ctx, userID)
	}

	// 构建上下文
	var systemPrompt string
	if req.ChildID != "" {
		child, _ := s.repo.GetChildByID(ctx, req.ChildID)
		if child != nil {
			systemPrompt = fmt.Sprintf(`你是专业的儿童生长发育AI助理。根据以下信息回答用户问题：
宝宝姓名: %s
性别: %s
出生日期: %s
父亲身高: %.1fcm
母亲身高: %.1fcm
靶身高范围: %.1f-%.1fcm

请用专业、温和的语气回答儿童生长发育相关问题。涉及医疗建议时，请提醒用户咨询专业医生。`,
				child.Name,
				map[string]string{"male": "男孩", "female": "女孩"}[child.Gender],
				child.Birthday.Format("2006-01-02"),
				child.FatherHeight,
				child.MotherHeight,
				s.CalculateTargetHeight(child).MinHeight,
				s.CalculateTargetHeight(child).MaxHeight,
			)
		}
	} else {
		systemPrompt = "你是专业的儿童生长发育AI助理。请用专业、温和的语气回答关于儿童身高、体重、营养、运动、睡眠等方面的问题。涉及医疗建议时，请提醒用户咨询专业医生。"
	}

	// 调用AI (这里简化处理，实际需要调用AI API)
	response := s.mockAIResponse(systemPrompt, req.Message, req.Context)

	// 记录对话
	sessionID := fmt.Sprintf("session_%s", userID)
	var messages []models.AIChatMessage
	if req.Context != nil {
		messages = append(messages, req.Context...)
	}
	messages = append(messages, models.AIChatMessage{Role: "user", Content: req.Message})
	messages = append(messages, models.AIChatMessage{Role: "assistant", Content: response})

	messagesJSON, _ := json.Marshal(messages)
	conv := &models.AIConversation{
		UserID:    userID,
		ChildID:   req.ChildID,
		SessionID: sessionID,
		Messages:  string(messagesJSON),
	}

	existing, _ := s.repo.GetConversationBySessionID(ctx, sessionID)
	if existing != nil {
		conv.ID = existing.ID
		s.repo.UpdateConversation(ctx, conv)
	} else {
		s.repo.CreateConversation(ctx, conv)
	}

	return &models.AIChatResponse{
		Response: response,
		Tokens:   len(req.Message) / 4, // 粗略估算
	}, nil
}

func (s *ServiceImpl) ParseReport(ctx context.Context, userID string, req *models.ParseReportRequest) (*models.ParseReportResponse, error) {
	// OCR识别 (这里简化处理，实际需要调用腾讯云OCR)
	ocrText := "模拟OCR识别结果:\n骨龄: 8岁\n骨骺: 未闭合\n评估: 骨龄与实际年龄相符"

	// AI解析
	analysis := "根据化验单分析：宝宝的骨龄评估显示骨骼发育正常，与实际年龄相符。这意味着宝宝的骨骼生长处于正常轨道。建议继续保持均衡营养和适量运动。"

	suggestions := []string{
		"继续保持均衡的饮食习惯",
		"每天保证8-10小时睡眠",
		"适量进行跳绳、篮球等纵向运动",
		"每3-6个月复查骨龄",
	}

	result := &models.AIReportResult{
		KeyIndicators: []models.KeyIndicator{
			{Name: "骨龄", Value: "8岁", Status: "normal"},
			{Name: "骨骺状态", Value: "未闭合", Status: "normal"},
		},
		NormalRanges: map[string]string{
			"骨龄": "与实际年龄相差±1岁为正常",
		},
		Analysis:   analysis,
		Suggestions: suggestions,
	}

	// 保存化验单记录
	report := &models.LabReport{
		ChildID:    req.ChildID,
		UserID:     userID,
		ImageURL:   req.ImageURL,
		OCRText:    ocrText,
		ReportType: req.ReportType,
	}
	resultJSON, _ := json.Marshal(result)
	report.AIResult = string(resultJSON)

	s.repo.CreateLabReport(ctx, report)

	return &models.ParseReportResponse{
		OCRText:  ocrText,
		AIResult: result,
	}, nil
}

// ========== 首页数据 ==========

func (s *ServiceImpl) GetHomeData(ctx context.Context, userID string) (*models.HomeDataResponse, error) {
	response := &models.HomeDataResponse{}

	// 获取激活的宝宝
	child, err := s.repo.GetActiveChild(ctx, userID)
	if err != nil || child == nil {
		response.HasBaby = false
		return response, nil
	}

	response.HasBaby = true

	// 构建宝宝详情
	childResponse := &models.ChildResponse{
		Child:        *child,
		TargetHeight: s.CalculateTargetHeight(child),
	}

	now := time.Now()
	years, months := child.CalculateAge(now)
	childResponse.AgeStr = fmt.Sprintf("%d岁%d月", years, months)
	childResponse.InterventionWindow = s.CalculateInterventionWindow(child)

	response.Baby = childResponse
	response.TargetHeight = childResponse.TargetHeight.TargetHeight
	response.TargetHeightMin = childResponse.TargetHeight.MinHeight
	response.TargetHeightMax = childResponse.TargetHeight.MaxHeight

	// 获取记录
	records, _, _ := s.repo.GetRecordsByChildID(ctx, child.ID, "", "", 1, 100)
	if len(records) > 0 {
		// 按日期排序
		sort.Slice(records, func(i, j int) bool {
			return records[i].Date.Before(records[j].Date)
		})

		response.Records = records
		response.LatestRecord = &records[len(records)-1]

		// 计算百分位和生长状态
		response.Percentile = s.CalculatePercentile(response.LatestRecord, child)
		response.GrowthStatus = s.DetermineGrowthStatus(
			response.Percentile,
			response.TargetHeightMin,
			response.TargetHeightMax,
			response.LatestRecord.Height,
		)

		// 生成图表数据
		response.ChartData = &models.ChartData{
			Categories: make([]string, len(records)),
			Series: []models.ChartSeries{
				{Name: "身高", Data: make([]float64, len(records))},
			},
		}
		for i, r := range records {
			response.ChartData.Categories[i] = r.Date.Format("01-02")
			response.ChartData.Series[0].Data[i] = r.Height
		}
	}

	// 会员状态
	sub, _ := s.repo.GetSubscriptionByUserID(ctx, userID)
	if sub != nil && sub.IsActive() {
		response.IsVip = true
		response.AIRemaining = sub.GetRemainingQuota()
	} else {
		response.IsVip = false
		if sub != nil {
			response.AIRemaining = sub.AIQuota - sub.AIUsed
		} else {
			response.AIRemaining = 3
		}
	}

	return response, nil
}

// ========== 科学计算方法 ==========

// CalculateTargetHeight 计算靶身高 (Khamis-Roche简化版)
func (s *ServiceImpl) CalculateTargetHeight(child *models.Child) models.TargetHeightInfo {
	father := child.FatherHeight
	mother := child.MotherHeight

	var targetHeight float64
	if child.Gender == "male" {
		targetHeight = (father + mother + 13) / 2
	} else {
		targetHeight = (father + mother - 13) / 2
	}

	// ±8cm误差范围
	return models.TargetHeightInfo{
		TargetHeight: math.Round(targetHeight*10) / 10,
		MinHeight:    math.Round((targetHeight-8)*10) / 10,
		MaxHeight:    math.Round((targetHeight+8)*10) / 10,
	}
}

// CalculateInterventionWindow 计算干预窗口
func (s *ServiceImpl) CalculateInterventionWindow(child *models.Child) *models.InterventionInfo {
	now := time.Now()
	birth := child.Birthday

	// 男孩: 10-15岁, 女孩: 8-13岁
	var windowStart, windowEnd time.Time
	if child.Gender == "male" {
		windowStart = birth.AddDate(10, 0, 0)
		windowEnd = birth.AddDate(15, 0, 0)
	} else {
		windowStart = birth.AddDate(8, 0, 0)
		windowEnd = birth.AddDate(13, 0, 0)
	}

	remainingDays := int(now.Sub(windowEnd).Hours() / 24)
	if remainingDays > 0 {
		remainingDays = 0
	}

	isInWindow := now.After(windowStart) && now.Before(windowEnd)

	return &models.InterventionInfo{
		Start:         windowStart.Format("2006-01-02"),
		End:           windowEnd.Format("2006-01-02"),
		RemainingDays: -remainingDays,
		IsInWindow:    isInWindow,
	}
}

// CalculatePercentile 计算百分位 (简化版WHO标准)
func (s *ServiceImpl) CalculatePercentile(record *models.Record, child *models.Child) int {
	if record == nil {
		return 0
	}

	height := record.Height
	ageInMonths := float64(record.AgeInDays) / 30.44

	// 简化版WHO参考值 (50百分位)
	var median50 float64
	if child.Gender == "male" {
		median50 = 76 + ageInMonths*0.65
	} else {
		median50 = 75 + ageInMonths*0.62
	}

	// 计算偏离程度
	ratio := height / median50

	if ratio >= 1.15 {
		return 97
	} else if ratio >= 1.10 {
		return 85
	} else if ratio >= 1.05 {
		return 75
	} else if ratio >= 1.0 {
		return 50
	} else if ratio >= 0.95 {
		return 25
	} else if ratio >= 0.90 {
		return 10
	} else if ratio >= 0.85 {
		return 5
	}
	return 3
}

// DetermineGrowthStatus 判断生长状态
func (s *ServiceImpl) DetermineGrowthStatus(percentile int, targetMin, targetMax, currentHeight float64) string {
	// 正常: 百分位在15-85之间
	if percentile >= 15 && percentile <= 85 {
		return "normal"
	}

	// 低于靶身高通道
	if currentHeight < targetMin {
		return "warning"
	}

	// 需要关注
	return "attention"
}

// ========== 辅助方法 ==========

func (s *ServiceImpl) getOpenIDFromWx(ctx context.Context, code string) (string, error) {
	// 实际实现需要调用微信API
	// 这里简化处理
	return fmt.Sprintf("mock_openid_%s", code), nil
}

func (s *ServiceImpl) generateToken(userID string) (string, int64, error) {
	expireAt := time.Now().Add(time.Duration(s.config.JWT.ExpireTime) * time.Second).Unix()
	// 实际实现需要生成JWT token
	token := fmt.Sprintf("jwt_token_%s_%d", userID, expireAt)
	return token, expireAt, nil
}

func (s *ServiceImpl) generateReferralCode() string {
	return fmt.Sprintf("GT%s", uuid.New().String()[:6])
}

func (s *ServiceImpl) generateInviteCode() string {
	chars := "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	code := make([]byte, 6)
	for i := range code {
		idx, _ := strconv.ParseInt(uuid.New().String()[:8], 16, 64)
		code[i] = chars[idx%int64(len(chars))]
	}
	return string(code)
}

func (s *ServiceImpl) getPlanPrice(planID string) int {
	prices := map[string]int{
		"monthly":  2990,
		"quarterly": 6990,
		"yearly":   19990,
	}
	if price, ok := prices[planID]; ok {
		return price
	}
	return 0
}

func (s *ServiceImpl) getMemberBenefits(sub *models.Subscription) []models.MemberBenefit {
	if sub == nil || !sub.IsActive() {
		return []models.MemberBenefit{}
	}

	benefits := []models.MemberBenefit{
		{Icon: "infinity", Text: "无限次AI分析"},
		{Icon: "history", Text: "历史记录永久保存"},
		{Icon: "priority", Text: "优先使用新功能"},
	}

	switch sub.Plan {
	case "yearly":
		benefits = append(benefits,
			models.MemberBenefit{Icon: "shield", Text: "专属客服支持"},
			models.MemberBenefit{Icon: "star", Text: "医生咨询预约"},
		)
	case "quarterly", "monthly":
		benefits = append(benefits,
			models.MemberBenefit{Icon: "support", Text: "在线客服支持"},
		)
	}

	return benefits
}

func (s *ServiceImpl) mockAIResponse(systemPrompt, userMessage string, context []models.AIChatMessage) string {
	// 简单的模拟AI响应
	msg := strings.ToLower(userMessage)

	if strings.Contains(msg, "正常") || strings.Contains(msg, "发育") {
		return "根据您提供的信息，宝宝目前的生长发育处于正常范围。建议继续保持均衡营养，保证充足睡眠，每天进行适量的户外运动。每3个月记录一次身高体重数据，持续关注发育趋势。"
	}

	if strings.Contains(msg, "靶身高") || strings.Contains(msg, "预测") {
		return "靶身高是根据父母身高使用Khamis-Roche公式计算的遗传潜力身高，仅供参考。实际身高会受到营养、运动、睡眠、疾病等多种因素影响。建议关注宝宝的生长速度，如果每年增长少于5cm，建议咨询专业医生。"
	}

	if strings.Contains(msg, "营养") || strings.Contains(msg, "吃什么") {
		return "促进长高的营养建议：\n1. 每天300-500ml牛奶（补钙）\n2. 每天1-2个鸡蛋（优质蛋白）\n3. 适量瘦肉、鱼虾（蛋白质+锌）\n4. 多吃蔬菜水果（维生素）\n5. 避免过多甜食和碳酸饮料"
	}

	if strings.Contains(msg, "睡眠") {
		return "睡眠对身高发育很重要！生长激素在深睡眠时分泌最旺盛。建议：\n- 2-6岁: 10-12小时/天\n- 7-12岁: 10-11小时/天\n- 养成规律作息的习惯\n- 睡前1小时避免使用电子设备"
	}

	if strings.Contains(msg, "运动") {
		return "推荐的长高运动：\n1. 跳绳（每天1000-2000个）\n2. 篮球（跳跃动作多）\n3. 游泳（全身伸展）\n4. 摸高跳（刺激骨骼生长）\n每次运动30-60分钟，以有氧运动为主，运动后做好拉伸。"
	}

	return "感谢您的提问！关于儿童生长发育，我建议您：\n1. 定期记录身高体重数据\n2. 关注生长速度而非单次数值\n3. 保持均衡营养和充足睡眠\n4. 每天适量运动\n如有具体问题，可以继续咨询，或升级Pro会员获得更详细的个性化分析。"
}
