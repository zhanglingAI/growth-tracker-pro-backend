package service

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/growth-tracker-pro-backend/internal/models"
	"gorm.io/gorm"
)

// Service 服务接口
type Service interface {
	// 认证
	Login(ctx context.Context, code string) (*models.LoginResponse, error)

	// 用户
	GetUserInfo(ctx context.Context, userID string) (*models.User, error)
	UpdateUser(ctx context.Context, userID string, req *models.UpdateUserRequest) error

	// 宝宝
	GetChildren(ctx context.Context, userID string) ([]models.Child, error)
	CreateChild(ctx context.Context, userID string, req *models.CreateChildRequest) (*models.Child, error)
	GetChildDetail(ctx context.Context, userID, childID string) (*models.ChildResponse, error)
	UpdateChild(ctx context.Context, userID, childID string, req *models.UpdateChildRequest) error
	DeleteChild(ctx context.Context, userID, childID string) error
	SwitchChild(ctx context.Context, userID, childID string) error

	// 记录
	GetRecords(ctx context.Context, childID string, req *models.RecordListRequest) (*models.PageResponse, error)
	CreateRecord(ctx context.Context, userID string, req *models.CreateRecordRequest) (*models.GrowthRecord, error)
	UpdateRecord(ctx context.Context, userID, recordID string, req *models.UpdateRecordRequest) error
	DeleteRecord(ctx context.Context, userID, recordID string) error

	// 订阅
	GetSubscription(ctx context.Context, userID string) (*models.SubscriptionResponse, error)
	CreateOrder(ctx context.Context, userID string, req *models.CreateOrderRequest) (*models.OrderResponse, error)
	ProcessPayCallback(ctx context.Context, xmlData map[string]string) error

	// 家庭
	GetFamily(ctx context.Context, userID string) (*models.FamilyResponse, error)
	CreateFamily(ctx context.Context, userID string, req *models.CreateFamilyRequest) (*models.Family, error)
	JoinFamily(ctx context.Context, userID string, req *models.JoinFamilyRequest) error
	LeaveFamily(ctx context.Context, userID string) error
	UpdateMemberRole(ctx context.Context, userID, memberID, role string) error
	GenerateInviteCode(ctx context.Context, userID string) (*models.GenerateInviteCodeResponse, error)

	// AI
	Chat(ctx context.Context, userID string, req *models.AIChatRequest) (*models.AIChatResponse, error)
	ParseReport(ctx context.Context, userID string, req *models.ParseReportRequest) (*models.ParseReportResponse, error)

	// 首页
	GetHomeData(ctx context.Context, userID string) (*models.HomeDataResponse, error)
}

// growthService 生长服务实现
type growthService struct {
	db *gorm.DB
}

// NewService 创建服务实例
func NewService(db *gorm.DB) Service {
	return &growthService{db: db}
}

// ========== 认证 ==========

func (s *growthService) Login(ctx context.Context, code string) (*models.LoginResponse, error) {
	// 简化实现：使用code作为userID
	user := &models.User{}
	result := s.db.WithContext(ctx).Where("open_id = ?", code).First(user)
	if result.Error != nil {
		// 创建新用户
		user = &models.User{
			OpenID:    code,
			Nickname:  "新用户",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := s.db.WithContext(ctx).Create(user).Error; err != nil {
			return nil, err
		}
	}

	token := "jwt_token_" + user.OpenID + "_" + strconv.FormatInt(time.Now().Unix(), 10)
	return &models.LoginResponse{
		Token:    token,
		ExpireAt: time.Now().Add(7 * 24 * time.Hour).Unix(),
		User:     user,
	}, nil
}

// ========== 用户 ==========

func (s *growthService) GetUserInfo(ctx context.Context, userID string) (*models.User, error) {
	user := &models.User{}
	if err := s.db.WithContext(ctx).Where("open_id = ?", userID).First(user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (s *growthService) UpdateUser(ctx context.Context, userID string, req *models.UpdateUserRequest) error {
	updates := map[string]interface{}{}
	if req.NickName != "" {
		updates["nickname"] = req.NickName
	}
	if req.AvatarURL != "" {
		updates["avatar"] = req.AvatarURL
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if len(updates) == 0 {
		return nil
	}
	return s.db.WithContext(ctx).Model(&models.User{}).Where("open_id = ?", userID).Updates(updates).Error
}

// ========== 宝宝 ==========

func (s *growthService) GetChildren(ctx context.Context, userID string) ([]models.Child, error) {
	var children []models.Child
	err := s.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at desc").Find(&children).Error
	return children, err
}

func (s *growthService) CreateChild(ctx context.Context, userID string, req *models.CreateChildRequest) (*models.Child, error) {
	birthday, err := time.Parse("2006-01-02", req.Birthday)
	if err != nil {
		return nil, err
	}

	child := &models.Child{
		UserID:        userID,
		Nickname:      req.Name,
		Gender:        req.Gender,
		Birthday:      birthday,
		FatherHeight:  req.FatherHeight,
		MotherHeight:  req.MotherHeight,
		StandardType:  "cn",
	}

	if err := s.db.WithContext(ctx).Create(child).Error; err != nil {
		return nil, err
	}
	return child, nil
}

func (s *growthService) GetChildDetail(ctx context.Context, userID, childID string) (*models.ChildResponse, error) {
	child := &models.Child{}
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", childID, userID).First(child).Error; err != nil {
		return nil, err
	}

	// 获取最新记录
	var latestRecord models.GrowthRecord
	s.db.WithContext(ctx).Where("child_id = ?", childID).Order("measure_date desc").First(&latestRecord)

	// 计算年龄
	years, months := child.CalculateAge(time.Now())
	ageStr := ""
	if years > 0 {
		ageStr = string(rune('0'+years)) + "岁"
	}
	if months > 0 {
		ageStr += string(rune('0'+months)) + "个月"
	}

	return &models.ChildResponse{
		Child:       *child,
		AgeStr:      ageStr,
		LatestRecord: &latestRecord,
	}, nil
}

func (s *growthService) UpdateChild(ctx context.Context, userID, childID string, req *models.UpdateChildRequest) error {
	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["nickname"] = req.Name
	}
	if req.Gender != "" {
		updates["gender"] = req.Gender
	}
	if req.Birthday != "" {
		if birthday, err := time.Parse("2006-01-02", req.Birthday); err == nil {
			updates["birthday"] = birthday
		}
	}
	if req.FatherHeight > 0 {
		updates["father_height"] = req.FatherHeight
	}
	if req.MotherHeight > 0 {
		updates["mother_height"] = req.MotherHeight
	}
	if len(updates) == 0 {
		return nil
	}
	return s.db.WithContext(ctx).Model(&models.Child{}).Where("id = ? AND user_id = ?", childID, userID).Updates(updates).Error
}

func (s *growthService) DeleteChild(ctx context.Context, userID, childID string) error {
	return s.db.WithContext(ctx).Where("id = ? AND user_id = ?", childID, userID).Delete(&models.Child{}).Error
}

func (s *growthService) SwitchChild(ctx context.Context, userID, childID string) error {
	// 简化实现：更新用户设置
	return s.db.WithContext(ctx).Model(&models.User{}).Where("open_id = ?", userID).Update("settings", `{"current_child_id":"`+childID+`"}`).Error
}

// ========== 记录 ==========

func (s *growthService) GetRecords(ctx context.Context, childID string, req *models.RecordListRequest) (*models.PageResponse, error) {
	query := s.db.WithContext(ctx).Model(&models.GrowthRecord{}).Where("child_id = ?", childID)

	if req.StartDate != "" {
		if startDate, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			query = query.Where("measure_date >= ?", startDate)
		}
	}
	if req.EndDate != "" {
		if endDate, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			query = query.Where("measure_date <= ?", endDate)
		}
	}

	var total int64
	query.Count(&total)

	var records []models.GrowthRecord
	offset := (req.Page - 1) * req.PageSize
	if err := query.Order("measure_date desc").Offset(offset).Limit(req.PageSize).Find(&records).Error; err != nil {
		return nil, err
	}

	return &models.PageResponse{
		Items:    records,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (s *growthService) CreateRecord(ctx context.Context, userID string, req *models.CreateRecordRequest) (*models.GrowthRecord, error) {
	measureDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, err
	}

	record := &models.GrowthRecord{
		ChildID:     req.ChildID,
		MeasureDate: measureDate,
		Height:      req.Height,
		Remarks:     req.Note,
	}

	if req.Weight > 0 {
		weight := req.Weight
		record.Weight = &weight
	}

	if err := s.db.WithContext(ctx).Create(record).Error; err != nil {
		return nil, err
	}
	return record, nil
}

func (s *growthService) UpdateRecord(ctx context.Context, userID, recordID string, req *models.UpdateRecordRequest) error {
	updates := map[string]interface{}{}
	if req.Height > 0 {
		updates["height"] = req.Height
	}
	if req.Weight > 0 {
		weight := req.Weight
		updates["weight"] = &weight
	}
	if req.Date != "" {
		if date, err := time.Parse("2006-01-02", req.Date); err == nil {
			updates["measure_date"] = date
		}
	}
	if req.Note != "" {
		updates["remarks"] = req.Note
	}
	if len(updates) == 0 {
		return nil
	}

	// 验证记录属于该用户
	var record models.GrowthRecord
	if err := s.db.WithContext(ctx).Where("id = ?", recordID).First(&record).Error; err != nil {
		return err
	}

	var child models.Child
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", record.ChildID, userID).First(&child).Error; err != nil {
		return errors.New("记录不存在")
	}

	return s.db.WithContext(ctx).Model(&models.GrowthRecord{}).Where("id = ?", recordID).Updates(updates).Error
}

func (s *growthService) DeleteRecord(ctx context.Context, userID, recordID string) error {
	// 验证记录属于该用户
	var record models.GrowthRecord
	if err := s.db.WithContext(ctx).Where("id = ?", recordID).First(&record).Error; err != nil {
		return err
	}

	var child models.Child
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", record.ChildID, userID).First(&child).Error; err != nil {
		return errors.New("记录不存在")
	}

	return s.db.WithContext(ctx).Delete(&models.GrowthRecord{}, "id = ?", recordID).Error
}

// ========== 订阅 ==========

func (s *growthService) GetSubscription(ctx context.Context, userID string) (*models.SubscriptionResponse, error) {
	var sub models.Subscription
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&sub).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &models.SubscriptionResponse{
				IsActive:       false,
				RemainingQuota: 3,
				MemberBenefits: []models.MemberBenefit{
					{Icon: "chart", Text: "生长曲线分析"},
					{Icon: "ai", Text: "AI健康助手"},
					{Icon: "report", Text: "化验单解析"},
				},
			}, nil
		}
		return nil, err
	}

	remaining := sub.GetRemainingQuota()
	return &models.SubscriptionResponse{
		Subscription:  &sub,
		IsActive:      sub.IsActive(),
		RemainingQuota: remaining,
		MemberBenefits: []models.MemberBenefit{
			{Icon: "chart", Text: "生长曲线分析"},
			{Icon: "ai", Text: "AI健康助手"},
			{Icon: "report", Text: "化验单解析"},
			{Icon: "hospital", Text: "医院推荐"},
		},
	}, nil
}

func (s *growthService) CreateOrder(ctx context.Context, userID string, req *models.CreateOrderRequest) (*models.OrderResponse, error) {
	// 简化实现：返回模拟支付参数
	return &models.OrderResponse{
		TimeStamp: string(rune(time.Now().Unix())),
		NonceStr:  "mock_nonce",
		Package:   "prepay_id=mock",
		SignType:  "MD5",
		PaySign:   "mock_sign",
		OrderID:   "order_" + userID + "_" + time.Now().Format("20060102150405"),
	}, nil
}

func (s *growthService) ProcessPayCallback(ctx context.Context, xmlData map[string]string) error {
	// 简化实现
	return nil
}

// ========== 家庭 ==========

func (s *growthService) GetFamily(ctx context.Context, userID string) (*models.FamilyResponse, error) {
	var member models.FamilyMember
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	var family models.Family
	if err := s.db.WithContext(ctx).Where("family_id = ?", member.FamilyID).First(&family).Error; err != nil {
		return nil, err
	}

	var memberCount int64
	s.db.WithContext(ctx).Model(&models.FamilyMember{}).Where("family_id = ?", member.FamilyID).Count(&memberCount)

	return &models.FamilyResponse{
		Name:        family.Name,
		InviteCode:  family.InviteCode,
		MemberCount: int(memberCount),
	}, nil
}

func (s *growthService) CreateFamily(ctx context.Context, userID string, req *models.CreateFamilyRequest) (*models.Family, error) {
	family := &models.Family{
		FamilyID:   "fam_" + time.Now().Format("20060102150405"),
		Name:       req.Name,
		InviteCode: generateInviteCode(),
		CreatorID:  userID,
	}

	if err := s.db.WithContext(ctx).Create(family).Error; err != nil {
		return nil, err
	}

	// 添加创建者为成员
	member := &models.FamilyMember{
		FamilyID: family.FamilyID,
		UserID:   userID,
		Role:     "creator",
	}
	s.db.WithContext(ctx).Create(member)

	return family, nil
}

func (s *growthService) JoinFamily(ctx context.Context, userID string, req *models.JoinFamilyRequest) error {
	var family models.Family
	if err := s.db.WithContext(ctx).Where("invite_code = ?", req.InviteCode).First(&family).Error; err != nil {
		return errors.New("邀请码无效")
	}

	role := "viewer"
	if req.Role != "" {
		role = req.Role
	}

	member := &models.FamilyMember{
		FamilyID: family.FamilyID,
		UserID:   userID,
		Role:     role,
	}
	return s.db.WithContext(ctx).Create(member).Error
}

func (s *growthService) LeaveFamily(ctx context.Context, userID string) error {
	return s.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.FamilyMember{}).Error
}

func (s *growthService) UpdateMemberRole(ctx context.Context, userID, memberID, role string) error {
	// 验证权限
	var member models.FamilyMember
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&member).Error; err != nil {
		return errors.New("你不是家庭成员")
	}
	if member.Role != "creator" {
		return errors.New("只有创建者可以修改成员角色")
	}

	return s.db.WithContext(ctx).Model(&models.FamilyMember{}).Where("id = ?", memberID).Update("role", role).Error
}

func (s *growthService) GenerateInviteCode(ctx context.Context, userID string) (*models.GenerateInviteCodeResponse, error) {
	var member models.FamilyMember
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&member).Error; err != nil {
		return nil, errors.New("你还没有加入家庭")
	}

	var family models.Family
	if err := s.db.WithContext(ctx).Where("family_id = ?", member.FamilyID).First(&family).Error; err != nil {
		return nil, err
	}

	return &models.GenerateInviteCodeResponse{
		InviteCode: family.InviteCode,
		ShareURL:   "growthtracker://join?code=" + family.InviteCode,
	}, nil
}

// ========== AI ==========

func (s *growthService) Chat(ctx context.Context, userID string, req *models.AIChatRequest) (*models.AIChatResponse, error) {
	// 简化实现：返回模拟响应
	return &models.AIChatResponse{
		Response: "感谢您的提问。我会根据您孩子的生长数据提供专业建议。",
		Tokens:   50,
	}, nil
}

func (s *growthService) ParseReport(ctx context.Context, userID string, req *models.ParseReportRequest) (*models.ParseReportResponse, error) {
	// 简化实现：返回模拟解析结果
	return &models.ParseReportResponse{
		OCRText: "模拟OCR识别文本",
		AIResult: &models.AIReportResult{
			KeyIndicators: []models.KeyIndicator{
				{Name: "身高", Value: "120cm", Status: "normal"},
				{Name: "体重", Value: "22kg", Status: "normal"},
			},
			NormalRanges: map[string]string{
				"身高": "110-130cm",
				"体重": "18-26kg",
			},
			Analysis:   "各项指标均在正常范围内。",
			Suggestions: []string{"保持均衡饮食", "适当运动"},
		},
	}, nil
}

// ========== 首页 ==========

func (s *growthService) GetHomeData(ctx context.Context, userID string) (*models.HomeDataResponse, error) {
	// 获取第一个宝宝
	var children []models.Child
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at desc").Limit(1).Find(&children).Error; err != nil {
		return nil, err
	}

	resp := &models.HomeDataResponse{
		HasBaby:    len(children) > 0,
		IsVip:      false,
		AIRemaining: 3,
	}

	if len(children) > 0 {
		child := &children[0]
		years, months := child.CalculateAge(time.Now())
		ageStr := ""
		if years > 0 {
			ageStr = strconv.Itoa(years) + "岁"
		}
		if months > 0 {
			ageStr += strconv.Itoa(months) + "个月"
		}
		if ageStr == "" {
			ageStr = "0个月"
		}

		resp.Baby = &models.ChildResponse{
			Child:  *child,
			AgeStr: ageStr,
		}

		// 获取最新记录
		var latestRecord models.GrowthRecord
		s.db.WithContext(ctx).Where("child_id = ?", child.ID).Order("measure_date desc").First(&latestRecord)
		if latestRecord.ID != "" {
			resp.LatestRecord = &latestRecord
		}
	}

	return resp, nil
}

// ========== 辅助函数 ==========

func generateInviteCode() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	code := make([]byte, 6)
	for i := range code {
		code[i] = chars[time.Now().UnixNano()%int64(len(chars))]
		time.Sleep(time.Nanosecond)
	}
	return string(code)
}
