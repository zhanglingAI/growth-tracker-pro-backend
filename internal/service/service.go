package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/growth-tracker-pro-backend/internal/alert"
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
	GetChildren(ctx context.Context, userID string) (*models.PageResponse, error)
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

	// 预警系统
	SetGrowthStage(ctx context.Context, userID, childID string, req *models.SetGrowthStageRequest) error
	GetChildAlerts(ctx context.Context, userID, childID string, req *models.AlertListRequest) (*models.AlertListResponse, error)
	MarkAlertRead(ctx context.Context, userID, alertID string) error
	DismissAlert(ctx context.Context, userID string, req *models.DismissAlertRequest) error
	GetAlertsSummary(ctx context.Context, userID string) (*models.AlertSummaryResponse, error)

	// 环境问卷评估
	CreateEnvironmentAssessment(ctx context.Context, userID, childID string, req *models.CreateEnvironmentAssessmentRequest) (*models.EnvironmentAssessmentResponse, error)
	GetLatestEnvironmentAssessment(ctx context.Context, userID, childID string) (*models.EnvironmentAssessmentResponse, error)
	GetEnvironmentAssessmentHistory(ctx context.Context, userID, childID string, page, pageSize int) (*models.EnvironmentAssessmentHistoryResponse, error)

	// 靶身高与生长速度
	GetTargetHeightComparison(ctx context.Context, userID, childID string) (*models.TargetHeightComparisonResponse, error)
	GetGrowthVelocity(ctx context.Context, userID, childID string, monthsBack int) (*models.GrowthVelocityResponse, error)
}

// growthService 生长服务实现
type growthService struct {
	db          *gorm.DB
	alertEngine *alert.Engine
}

// NewService 创建服务实例
func NewService(db *gorm.DB) Service {
	return &growthService{
		db:          db,
		alertEngine: alert.NewEngine(db),
	}
}

// ========== 认证 ==========

func (s *growthService) Login(ctx context.Context, code string) (*models.LoginResponse, error) {
	// 简化实现：使用code作为userID
	user := &models.User{}
	result := s.db.WithContext(ctx).Where("open_id = ?", code).First(user)
	if result.Error != nil {
		// 创建新用户
		user = &models.User{
			OpenID:   code,
			NickName: "新用户",
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
		updates["nick_name"] = req.NickName
	}
	if req.AvatarURL != "" {
		updates["avatar_url"] = req.AvatarURL
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

func (s *growthService) GetChildren(ctx context.Context, userID string) (*models.PageResponse, error) {
	var children []models.Child
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at desc").Find(&children).Error; err != nil {
		return nil, err
	}

	result := make([]models.ChildResponse, len(children))
	for i, child := range children {
		// 计算年龄
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

		// 获取最新记录
		var latestRecord models.GrowthRecord
		s.db.WithContext(ctx).Where("child_id = ?", child.ID).Order("measure_date desc").First(&latestRecord)

		result[i] = models.ChildResponse{
			Child:        child,
			AgeStr:       ageStr,
			LatestRecord: &latestRecord,
		}
		if latestRecord.ID == "" {
			result[i].LatestRecord = nil
		} else {
			// 计算身高百分位 (全国同龄)
			ageInMonths := years*12 + months
			percentile := models.CalculateHeightPercentile(latestRecord.Height, ageInMonths, child.Gender)
			result[i].Percentile = percentile
			result[i].GrowthStatus = models.GetHeightPercentileStatus(percentile)
			result[i].IsNormalRange = models.IsHeightNormal(latestRecord.Height, ageInMonths, child.Gender)

			// Tanner靶身高 + 遗传潜力评估
			result[i].TargetHeight = models.CalculateTargetHeight(child.FatherHeight, child.MotherHeight, child.Gender)
			result[i].TargetPercentile = models.GetTargetHeightPercentile(latestRecord.Height, result[i].TargetHeight)
			result[i].PotentialStatus = models.GetHeightPotentialStatus(result[i].TargetPercentile)

			// 区域修正
			if child.Region != nil && *child.Region != "" {
				regionalPct := models.CalculateRegionalPercentile(latestRecord.Height, ageInMonths, child.Gender, *child.Region)
				result[i].AdjustedPercentile = regionalPct
				std, corr := models.GetRegionalGrowthStandard(ageInMonths, child.Gender, *child.Region)
				if corr != nil && std != nil {
					origStd := models.GetGrowthStandard(ageInMonths, child.Gender)
					result[i].RegionalCorrection = &models.RegionalCorrectionInfo{
						Region:       corr.ProvinceName,
						CorrectionCM: corr.Correction,
						AdjustedP50:  std.P50,
						OriginalP50:  origStd.P50,
					}
				}
			}

			// 骨龄信息
			result[i].BoneAgeInfo = s.buildBoneAgeSummary(ctx, child.ID)

			// 预警摘要
			alertSummary, _ := s.alertEngine.GetSummary(ctx, child.ID)
			result[i].Alerts = alertSummary
		}
	}

	return &models.PageResponse{
		Items:    result,
		Total:    int64(len(result)),
		Page:     1,
		PageSize: len(result),
	}, nil
}

func (s *growthService) CreateChild(ctx context.Context, userID string, req *models.CreateChildRequest) (*models.Child, error) {
	birthday, err := time.Parse("2006-01-02", req.Birthday)
	if err != nil {
		return nil, err
	}

	child := &models.Child{
		UserID:        userID,
		Nickname:      req.Nickname,
		Gender:        req.Gender,
		Birthday:      birthday,
		FatherHeight:  req.FatherHeight,
		MotherHeight:  req.MotherHeight,
		StandardType:  "cn",
		Region:        req.Region,
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
		ageStr = strconv.Itoa(years) + "岁"
	}
	if months > 0 {
		ageStr += strconv.Itoa(months) + "个月"
	}
	if ageStr == "" {
		ageStr = "0个月"
	}

	resp := &models.ChildResponse{
		Child:        *child,
		AgeStr:       ageStr,
		LatestRecord: &latestRecord,
	}
	if latestRecord.ID == "" {
		resp.LatestRecord = nil
	} else {
		// 计算身高百分位 (全国同龄)
		ageInMonths := years*12 + months
		percentile := models.CalculateHeightPercentile(latestRecord.Height, ageInMonths, child.Gender)
		resp.Percentile = percentile
		resp.GrowthStatus = models.GetHeightPercentileStatus(percentile)
		resp.IsNormalRange = models.IsHeightNormal(latestRecord.Height, ageInMonths, child.Gender)

		// Tanner靶身高 + 遗传潜力评估
		resp.TargetHeight = models.CalculateTargetHeight(child.FatherHeight, child.MotherHeight, child.Gender)
		resp.TargetPercentile = models.GetTargetHeightPercentile(latestRecord.Height, resp.TargetHeight)
		resp.PotentialStatus = models.GetHeightPotentialStatus(resp.TargetPercentile)

		// 区域修正
		if child.Region != nil && *child.Region != "" {
			regionalPct := models.CalculateRegionalPercentile(latestRecord.Height, ageInMonths, child.Gender, *child.Region)
			resp.AdjustedPercentile = regionalPct
			std, corr := models.GetRegionalGrowthStandard(ageInMonths, child.Gender, *child.Region)
			if corr != nil && std != nil {
				origStd := models.GetGrowthStandard(ageInMonths, child.Gender)
				resp.RegionalCorrection = &models.RegionalCorrectionInfo{
					Region:       corr.ProvinceName,
					CorrectionCM: corr.Correction,
					AdjustedP50:  std.P50,
					OriginalP50:  origStd.P50,
				}
			}
		}

		// 骨龄信息
		resp.BoneAgeInfo = s.buildBoneAgeSummary(ctx, child.ID)

		// 预警摘要
		alertSummary, _ := s.alertEngine.GetSummary(ctx, child.ID)
		resp.Alerts = alertSummary
	}

	return resp, nil
}

func (s *growthService) UpdateChild(ctx context.Context, userID, childID string, req *models.UpdateChildRequest) error {
	updates := map[string]interface{}{}
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
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
	if req.Region != nil {
		updates["region"] = *req.Region
	}
	if req.GrowthStage != nil {
		updates["growth_stage"] = *req.GrowthStage
		updates["stage_confirmed_at"] = time.Now()
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
	// 获取宝宝生日
	var child models.Child
	if err := s.db.WithContext(ctx).Select("birthday").Where("id = ?", childID).First(&child).Error; err != nil {
		return nil, err
	}

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

	// 为每条记录计算 age_str
	recordResponses := make([]models.RecordResponse, len(records))
	for i, record := range records {
		ageStr := calculateAgeString(child.Birthday, record.MeasureDate)
		recordResponses[i] = models.RecordResponse{
			GrowthRecord: record,
			AgeStr:       ageStr,
		}
	}

	return &models.PageResponse{
		Items:    recordResponses,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// calculateAgeString 计算两个日期之间的年龄字符串: "X岁X个月" / "X个月" / "X天"
func calculateAgeString(birthday, measureDate time.Time) string {
	years := measureDate.Year() - birthday.Year()
	months := int(measureDate.Month()) - int(birthday.Month())

	if months < 0 {
		years--
		months += 12
	}

	if measureDate.Day() < birthday.Day() {
		months--
		if months < 0 {
			years--
			months += 12
		}
	}

	// 小于1个月，显示天数
	if years == 0 && months == 0 {
		days := int(measureDate.Sub(birthday).Hours() / 24)
		if days <= 0 {
			days = 1
		}
		return strconv.Itoa(days) + "天"
	}

	ageStr := ""
	if years > 0 {
		ageStr = strconv.Itoa(years) + "岁"
	}
	if months > 0 {
		ageStr += strconv.Itoa(months) + "个月"
	}
	return ageStr
}

func (s *growthService) CreateRecord(ctx context.Context, userID string, req *models.CreateRecordRequest) (*models.GrowthRecord, error) {
	// 验证宝宝属于该用户
	var child models.Child
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", req.ChildID, userID).First(&child).Error; err != nil {
		return nil, errors.New("宝宝不存在或无权限")
	}

	measureDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, err
	}

	// 幂等性检查：同一天不能有两条记录
	var existingCount int64
	if err := s.db.WithContext(ctx).Model(&models.GrowthRecord{}).
		Where("child_id = ? AND measure_date = ?", req.ChildID, measureDate).
		Count(&existingCount).Error; err != nil {
		return nil, err
	}
	if existingCount > 0 {
		return nil, errors.New("该日期已有记录，请选择其他日期或更新现有记录")
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

	// 计算记录时的年龄（月）
	recordAgeYears, recordAgeMonths := child.CalculateAge(measureDate)
	recordAgeInMonths := recordAgeYears*12 + recordAgeMonths

	// 处理骨龄
	if req.BoneAge != nil {
		record.BoneAge = req.BoneAge
		source := "manual"
		record.BoneAgeSource = &source
		diff := *req.BoneAge - float64(recordAgeInMonths)/12.0
		record.BoneAgeDiff = &diff
	}

	// 计算百分位、Z分、状态
	pct := models.CalculateHeightPercentile(req.Height, recordAgeInMonths, child.Gender)
	zscore := models.CalculateZScore(req.Height, recordAgeInMonths, child.Gender)
	pctF := float64(pct)
	zscoreF := zscore
	record.HeightPercentile = &pctF
	record.HeightZScore = &zscoreF
	record.HeightStatus = models.GetHeightPercentileStatus(pct)

	if err := s.db.WithContext(ctx).Create(record).Error; err != nil {
		return nil, err
	}

	// 更新最后身高变化日期
	s.db.WithContext(ctx).Model(&models.Child{}).Where("id = ?", req.ChildID).
		Update("last_height_change_date", measureDate)

	// 触发预警评估
	_ = s.evaluateAndSaveAlerts(ctx, userID, &child, record)

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
	if req.BoneAge != nil {
		updates["bone_age"] = *req.BoneAge
		updates["bone_age_source"] = "manual"
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

	// 如果更新了身高或骨龄，重新计算 bone_age_diff 和百分位
	if req.BoneAge != nil || req.Height > 0 {
		measureDate := record.MeasureDate
		if req.Date != "" {
			if d, err := time.Parse("2006-01-02", req.Date); err == nil {
				measureDate = d
			}
		}
		years, months := child.CalculateAge(measureDate)
		ageInMonths := years*12 + months

		if req.BoneAge != nil {
			diff := *req.BoneAge - float64(ageInMonths)/12.0
			updates["bone_age_diff"] = diff
		}
		if req.Height > 0 {
			pct := models.CalculateHeightPercentile(req.Height, ageInMonths, child.Gender)
			zscore := models.CalculateZScore(req.Height, ageInMonths, child.Gender)
			updates["height_percentile"] = float64(pct)
			updates["height_zscore"] = zscore
			updates["height_status"] = models.GetHeightPercentileStatus(pct)
		}
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
		HasBaby:     len(children) > 0,
		IsVip:       false,
		AIRemaining: 3,
	}

	// 订阅信息
	resp.Subscription = &models.SubscriptionInfo{
		IsActive:       false,
		RemainingQuota: 3,
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

		// 计算靶身高和百分位
		targetHeight := (child.FatherHeight + child.MotherHeight) / 2
		if child.Gender == "male" {
			targetHeight += 6.5
		} else {
			targetHeight -= 6.5
		}

		resp.Baby = &models.ChildResponse{
			Child:    *child,
			AgeStr:   ageStr,
			TargetHeight: models.TargetHeightInfo{
				TargetHeight: models.Round(targetHeight, 1),
				MinHeight:    models.Round(targetHeight-8, 1),
				MaxHeight:    models.Round(targetHeight+8, 1),
			},
		}

		// 获取最近20条记录（带age_str）
		var records []models.GrowthRecord
		var total int64
		s.db.WithContext(ctx).Model(&models.GrowthRecord{}).Where("child_id = ?", child.ID).Count(&total)
		s.db.WithContext(ctx).Where("child_id = ?", child.ID).Order("measure_date desc").Limit(20).Find(&records)

		recordResponses := make([]models.RecordResponse, len(records))
		for i, record := range records {
			ageStr := calculateAgeString(child.Birthday, record.MeasureDate)
			recordResponses[i] = models.RecordResponse{
				GrowthRecord: record,
				AgeStr:       ageStr,
			}
		}

		resp.Records = &models.HomeRecordsResponse{
			Items:    recordResponses,
			Total:    total,
			Page:     1,
			PageSize: 20,
		}

		// 计算最新记录的百分位
		if len(records) > 0 {
			ageInMonths := years*12 + months
			percentile := models.CalculateHeightPercentile(records[0].Height, ageInMonths, child.Gender)
			resp.Baby.Percentile = percentile
			resp.Baby.GrowthStatus = models.GetHeightPercentileStatus(percentile)

			// 区域修正
			if child.Region != nil && *child.Region != "" {
				regionalPct := models.CalculateRegionalPercentile(records[0].Height, ageInMonths, child.Gender, *child.Region)
				resp.Baby.AdjustedPercentile = regionalPct
				std, corr := models.GetRegionalGrowthStandard(ageInMonths, child.Gender, *child.Region)
				if corr != nil && std != nil {
					origStd := models.GetGrowthStandard(ageInMonths, child.Gender)
					resp.Baby.RegionalCorrection = &models.RegionalCorrectionInfo{
						Region:       corr.ProvinceName,
						CorrectionCM: corr.Correction,
						AdjustedP50:  std.P50,
						OriginalP50:  origStd.P50,
					}
				}
			}

			// 骨龄信息
			resp.Baby.BoneAgeInfo = s.buildBoneAgeSummary(ctx, child.ID)

			// 预警摘要
			alertSummary, _ := s.alertEngine.GetSummary(ctx, child.ID)
			resp.Baby.Alerts = alertSummary
		}
	}

	return resp, nil
}

// ========== 预警相关 ==========

func (s *growthService) SetGrowthStage(ctx context.Context, userID, childID string, req *models.SetGrowthStageRequest) error {
	var child models.Child
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", childID, userID).First(&child).Error; err != nil {
		return errors.New("宝宝不存在")
	}
	return s.db.WithContext(ctx).Model(&models.Child{}).Where("id = ?", childID).Updates(map[string]interface{}{
		"growth_stage":         req.GrowthStage,
		"stage_confirmed_at":   time.Now(),
		"last_height_change_date": time.Now(),
	}).Error
}

func (s *growthService) GetChildAlerts(ctx context.Context, userID, childID string, req *models.AlertListRequest) (*models.AlertListResponse, error) {
	var child models.Child
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", childID, userID).First(&child).Error; err != nil {
		return nil, errors.New("宝宝不存在")
	}
	return s.alertEngine.GetChildAlertList(ctx, childID, req)
}

func (s *growthService) MarkAlertRead(ctx context.Context, userID, alertID string) error {
	// 验证预警属于该用户
	var alert models.HeightAlert
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", alertID, userID).First(&alert).Error; err != nil {
		return errors.New("预警不存在")
	}
	return s.alertEngine.MarkAlertRead(ctx, alertID)
}

func (s *growthService) DismissAlert(ctx context.Context, userID string, req *models.DismissAlertRequest) error {
	var alert models.HeightAlert
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", req.AlertID, userID).First(&alert).Error; err != nil {
		return errors.New("预警不存在")
	}
	return s.alertEngine.DismissAlert(ctx, req.AlertID, req.Reason)
}

func (s *growthService) GetAlertsSummary(ctx context.Context, userID string) (*models.AlertSummaryResponse, error) {
	// 获取用户的所有宝宝
	var children []models.Child
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Find(&children).Error; err != nil {
		return nil, err
	}

	if len(children) == 0 {
		return &models.AlertSummaryResponse{
			HasActiveAlert: false,
			TopAlerts:      []models.AlertResponse{},
		}, nil
	}

	// 取第一个宝宝的预警摘要作为全局摘要（简化处理）
	return s.alertEngine.GetSummary(ctx, children[0].ID)
}

// buildBoneAgeSummary 构建骨龄摘要
func (s *growthService) buildBoneAgeSummary(ctx context.Context, childID string) *models.BoneAgeSummary {
	var record models.GrowthRecord
	if err := s.db.WithContext(ctx).Where("child_id = ? AND bone_age IS NOT NULL", childID).
		Order("measure_date DESC").First(&record).Error; err != nil {
		return nil
	}

	summary := &models.BoneAgeSummary{
		LatestBoneAge:  record.BoneAge,
		BoneAgeDiff:    record.BoneAgeDiff,
		AssessmentDate: record.MeasureDate.Format("2006-01-02"),
		IsAbnormal:     false,
	}
	if record.BoneAgeSource != nil {
		summary.BoneAgeSource = *record.BoneAgeSource
	}
	if record.BoneAgeDiff != nil && (*record.BoneAgeDiff > 1.0 || *record.BoneAgeDiff < -1.0) {
		summary.IsAbnormal = true
	}
	return summary
}

// evaluateAndSaveAlerts 运行预警引擎并保存结果
func (s *growthService) evaluateAndSaveAlerts(ctx context.Context, userID string, child *models.Child, latestRecord *models.GrowthRecord) error {
	// 重新加载完整 child 数据（确保 Region 等字段完整）
	var fullChild models.Child
	if err := s.db.WithContext(ctx).First(&fullChild, "id = ?", child.ID).Error; err != nil {
		return err
	}

	// 获取所有记录
	var allRecords []models.GrowthRecord
	s.db.WithContext(ctx).Where("child_id = ?", child.ID).Order("measure_date asc").Find(&allRecords)

	// 获取有骨龄的记录
	var boneAgeRecords []models.GrowthRecord
	s.db.WithContext(ctx).Where("child_id = ? AND bone_age IS NOT NULL", child.ID).Order("measure_date desc").Find(&boneAgeRecords)

	// 计算年龄（月）
	years, months := fullChild.CalculateAge(time.Now())
	ageInMonths := years*12 + months

	currentPct := models.CalculateHeightPercentile(latestRecord.Height, ageInMonths, fullChild.Gender)
	regionalPct := currentPct
	if fullChild.Region != nil && *fullChild.Region != "" {
		regionalPct = models.CalculateRegionalPercentile(latestRecord.Height, ageInMonths, fullChild.Gender, *fullChild.Region)
	}

	input := &alert.Input{
		Child:          &fullChild,
		LatestRecord:   latestRecord,
		AllRecords:     allRecords,
		TargetHeight:   models.CalculateTargetHeight(fullChild.FatherHeight, fullChild.MotherHeight, fullChild.Gender),
		CurrentPct:     currentPct,
		RegionalPct:    regionalPct,
		Region:         "",
		BoneAgeRecords: boneAgeRecords,
	}
	if fullChild.Region != nil {
		input.Region = *fullChild.Region
	}

	alerts := s.alertEngine.Evaluate(input)
	return s.alertEngine.SaveAlerts(ctx, child.ID, userID, alerts)
}

// ========== 环境问卷评估 ==========

func (s *growthService) CreateEnvironmentAssessment(ctx context.Context, userID, childID string, req *models.CreateEnvironmentAssessmentRequest) (*models.EnvironmentAssessmentResponse, error) {
	var child models.Child
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", childID, userID).First(&child).Error; err != nil {
		return nil, errors.New("宝宝不存在")
	}

	ageYears, ageMonths := child.CalculateAge(time.Now())
	ageInMonths := ageYears*12 + ageMonths

	// 构建问卷结构
	questionnaire := &models.EnvironmentQuestionnaire{
		Nutrition: models.NutritionModule{
			DietDiversity:     req.Nutrition.DietDiversity,
			ProteinAdequacy:   req.Nutrition.ProteinAdequacy,
			CalciumIntake:     req.Nutrition.CalciumIntake,
			VitaminDStatus:    req.Nutrition.VitaminDStatus,
			BadEatingBehavior: req.Nutrition.BadEatingBehavior,
			WeightManagement:  req.Nutrition.WeightManagement,
		},
		Sleep: models.SleepModule{
			Duration:          req.Sleep.Duration,
			BedtimeRegularity: req.Sleep.BedtimeRegularity,
			DeepSleepCover:    req.Sleep.DeepSleepCover,
			SleepContinuity:   req.Sleep.SleepContinuity,
			SleepEnvironment:  req.Sleep.SleepEnvironment,
		},
		Exercise: models.ExerciseModule{
			Frequency:       req.Exercise.Frequency,
			TypeSuitability: req.Exercise.TypeSuitability,
			Duration:        req.Exercise.Duration,
			Intensity:       req.Exercise.Intensity,
		},
		Health: models.HealthModule{
			DiseaseControl:    req.Health.DiseaseControl,
			CheckupCompliance: req.Health.CheckupCompliance,
			MedicationSafety:  req.Health.MedicationSafety,
		},
		Mental: models.MentalModule{
			EmotionRegulation: req.Mental.EmotionRegulation,
			FamilySupport:     req.Mental.FamilySupport,
			StressManagement:  req.Mental.StressManagement,
		},
	}

	// 计算环境得分
	envScore := models.CalculateEnvironmentScore(questionnaire, ageYears)

	// 计算综合预测
	prediction := models.CalculateComprehensivePrediction(
		child.FatherHeight, child.MotherHeight,
		req.CurrentHeight, req.CurrentWeight,
		ageYears, ageInMonths,
		child.Gender, questionnaire,
	)

	// 序列化原始答案
	nutritionRaw, _ := json.Marshal(req.Nutrition)
	sleepRaw, _ := json.Marshal(req.Sleep)
	exerciseRaw, _ := json.Marshal(req.Exercise)
	healthRaw, _ := json.Marshal(req.Health)
	mentalRaw, _ := json.Marshal(req.Mental)

	// 生成行动计划
	actionPlan := generateWeeklyActionPlan(envScore.ModuleScores)
	actionPlanRaw, _ := json.Marshal(actionPlan)

	// 保存到数据库
	assessment := &models.EnvironmentAssessment{
		ChildID:              childID,
		UserID:               userID,
		AssessmentDate:       time.Now(),
		NutritionRaw:         string(nutritionRaw),
		SleepRaw:             string(sleepRaw),
		ExerciseRaw:          string(exerciseRaw),
		HealthRaw:            string(healthRaw),
		MentalRaw:            string(mentalRaw),
		NutritionScore:       envScore.ModuleScores["nutrition"].Score,
		SleepScore:           envScore.ModuleScores["sleep"].Score,
		ExerciseScore:        envScore.ModuleScores["exercise"].Score,
		HealthScore:          envScore.ModuleScores["health"].Score,
		MentalScore:          envScore.ModuleScores["mental"].Score,
		TotalScore:           envScore.TotalScore,
		GeneticTargetHeight:  prediction.GeneticTargetHeight,
		EnvironmentIncrement: prediction.EnvironmentIncrement,
		PredictedHeight:      prediction.PredictedHeight,
		InterventionZone:     envScore.InterventionZone,
		ActionPlan:           string(actionPlanRaw),
	}

	if err := s.db.WithContext(ctx).Create(assessment).Error; err != nil {
		return nil, err
	}

	return buildEnvironmentAssessmentResponse(assessment, &envScore, &prediction, actionPlan), nil
}

func (s *growthService) GetLatestEnvironmentAssessment(ctx context.Context, userID, childID string) (*models.EnvironmentAssessmentResponse, error) {
	var child models.Child
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", childID, userID).First(&child).Error; err != nil {
		return nil, errors.New("宝宝不存在")
	}

	var assessment models.EnvironmentAssessment
	if err := s.db.WithContext(ctx).Where("child_id = ?", childID).Order("assessment_date DESC").First(&assessment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("暂无评估记录")
		}
		return nil, err
	}

	// 反序列化行动计划
	var actionPlan *models.WeeklyActionPlan
	if assessment.ActionPlan != "" {
		_ = json.Unmarshal([]byte(assessment.ActionPlan), &actionPlan)
	}

	// 重新构建模块得分
	moduleScores := map[string]models.ModuleScore{
		"nutrition": {Score: assessment.NutritionScore, MaxScore: 15, Weight: 0.30},
		"sleep":     {Score: assessment.SleepScore, MaxScore: 12.5, Weight: 0.25},
		"exercise":  {Score: assessment.ExerciseScore, MaxScore: 12.5, Weight: 0.25},
		"health":    {Score: assessment.HealthScore, MaxScore: 5, Weight: 0.10},
		"mental":    {Score: assessment.MentalScore, MaxScore: 5, Weight: 0.10},
	}

	resp := &models.EnvironmentAssessmentResponse{
		ID:                    assessment.ID,
		ChildID:               assessment.ChildID,
		AssessmentDate:        assessment.AssessmentDate.Format("2006-01-02"),
		ModuleScores:          moduleScores,
		TotalScore:            assessment.TotalScore,
		MaxPossibleScore:      50,
		InterventionZone:      assessment.InterventionZone,
		ZoneLabel:             getZoneLabel(assessment.InterventionZone),
		Interpretation:        getZoneInterpretation(assessment.InterventionZone, assessment.TotalScore),
		GeneticTargetHeight:   assessment.GeneticTargetHeight,
		EnvironmentIncrement:  assessment.EnvironmentIncrement,
		PredictedHeight:       assessment.PredictedHeight,
		PredictionMethod:      "Khamis-Roche + 环境增量",
		ErrorRange:            8.0,
		ActionPlan:            actionPlan,
		ClinicalInterpretation: getZoneInterpretation(assessment.InterventionZone, assessment.TotalScore),
	}

	return resp, nil
}

func (s *growthService) GetEnvironmentAssessmentHistory(ctx context.Context, userID, childID string, page, pageSize int) (*models.EnvironmentAssessmentHistoryResponse, error) {
	var child models.Child
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", childID, userID).First(&child).Error; err != nil {
		return nil, errors.New("宝宝不存在")
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	var total int64
	s.db.WithContext(ctx).Model(&models.EnvironmentAssessment{}).Where("child_id = ?", childID).Count(&total)

	var assessments []models.EnvironmentAssessment
	if err := s.db.WithContext(ctx).Where("child_id = ?", childID).Order("assessment_date DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&assessments).Error; err != nil {
		return nil, err
	}

	items := make([]models.EnvironmentAssessmentSummary, len(assessments))
	for i, a := range assessments {
		items[i] = models.EnvironmentAssessmentSummary{
			ID:                   a.ID,
			AssessmentDate:       a.AssessmentDate.Format("2006-01-02"),
			TotalScore:           a.TotalScore,
			InterventionZone:     a.InterventionZone,
			PredictedHeight:      a.PredictedHeight,
			EnvironmentIncrement: a.EnvironmentIncrement,
		}
	}

	return &models.EnvironmentAssessmentHistoryResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// ========== 靶身高与生长速度 ==========

func (s *growthService) GetTargetHeightComparison(ctx context.Context, userID, childID string) (*models.TargetHeightComparisonResponse, error) {
	var child models.Child
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", childID, userID).First(&child).Error; err != nil {
		return nil, errors.New("宝宝不存在")
	}

	// 获取最新记录
	var latestRecord models.GrowthRecord
	if err := s.db.WithContext(ctx).Where("child_id = ?", childID).Order("measure_date DESC").First(&latestRecord).Error; err != nil {
		return nil, errors.New("暂无生长记录")
	}

	ageYears, ageMonths := child.CalculateAge(time.Now())
	ageInMonths := ageYears*12 + ageMonths

	// 遗传靶身高
	targetHeight := models.CalculateTargetHeight(child.FatherHeight, child.MotherHeight, child.Gender)

	// 定量遗传学（父母身高极端时）
	var qgResult *models.QuantitativeGeneticsResult
	fatherDev := abs(child.FatherHeight - 172.0)
	motherDev := abs(child.MotherHeight - 160.0)
	if fatherDev > 10 || motherDev > 10 {
		qg := models.CalculateQuantitativeGeneticsTargetHeight(child.FatherHeight, child.MotherHeight, child.Gender, 0.75)
		qgResult = &qg
	}

	// Khamis-Roche
	var krResult *models.KhamisRocheResult
	if ageYears >= 4 && ageYears <= 17 && latestRecord.Weight != nil && *latestRecord.Weight > 0 {
		kr := models.CalculateKhamisRoche(child.FatherHeight, child.MotherHeight, latestRecord.Height, *latestRecord.Weight, ageInMonths, child.Gender)
		krResult = &kr
	}

	// 当前百分位
	currentPct := models.CalculateHeightPercentile(latestRecord.Height, ageInMonths, child.Gender)

	// 获取最新环境评估
	var envAssessment models.EnvironmentAssessment
	var envPrediction *models.ComprehensivePredictionResult
	if err := s.db.WithContext(ctx).Where("child_id = ?", childID).Order("assessment_date DESC").First(&envAssessment).Error; err == nil {
		// 重新构建问卷并计算
		questionnaire := &models.EnvironmentQuestionnaire{}
		_ = json.Unmarshal([]byte(envAssessment.NutritionRaw), &questionnaire.Nutrition)
		_ = json.Unmarshal([]byte(envAssessment.SleepRaw), &questionnaire.Sleep)
		_ = json.Unmarshal([]byte(envAssessment.ExerciseRaw), &questionnaire.Exercise)
		_ = json.Unmarshal([]byte(envAssessment.HealthRaw), &questionnaire.Health)
		_ = json.Unmarshal([]byte(envAssessment.MentalRaw), &questionnaire.Mental)
		weight := 0.0
		if latestRecord.Weight != nil {
			weight = *latestRecord.Weight
		}
		pred := models.CalculateComprehensivePrediction(
			child.FatherHeight, child.MotherHeight,
			latestRecord.Height, weight,
			ageYears, ageInMonths,
			child.Gender, questionnaire,
		)
		envPrediction = &pred
	}

	// 生长速度
	var records []models.GrowthRecord
	s.db.WithContext(ctx).Where("child_id = ?", childID).Order("measure_date asc").Find(&records)
	var velocityResp *models.GrowthVelocityResponse
	if len(records) >= 2 {
		velocity, _ := models.CalculateAnnualGrowthVelocity(records, 12)
		status, level, action := models.EvaluateGrowthVelocityWithAlert(ageYears, velocity, child.Gender)
		velocityResp = &models.GrowthVelocityResponse{
			Velocity:       velocity,
			MonthsBack:     12,
			LatestHeight:   records[len(records)-1].Height,
			PreviousHeight: records[0].Height,
			Status:         status,
			AlertLevel:     level,
			Action:         action,
		}
	}

	return &models.TargetHeightComparisonResponse{
		GeneticTargetHeight:   targetHeight.TargetHeight,
		QuantitativeGenetics:  qgResult,
		KhamisRoche:           krResult,
		EnvironmentPrediction: envPrediction,
		CurrentHeight:         latestRecord.Height,
		CurrentPercentile:     currentPct,
		PotentialStatus:       evaluatePotentialStatus(latestRecord.Height, targetHeight.TargetHeight, currentPct),
		GrowthVelocity:        velocityResp,
	}, nil
}

func (s *growthService) GetGrowthVelocity(ctx context.Context, userID, childID string, monthsBack int) (*models.GrowthVelocityResponse, error) {
	var child models.Child
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", childID, userID).First(&child).Error; err != nil {
		return nil, errors.New("宝宝不存在")
	}

	var records []models.GrowthRecord
	if err := s.db.WithContext(ctx).Where("child_id = ?", childID).Order("measure_date asc").Find(&records).Error; err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return nil, errors.New("记录不足，无法计算生长速度")
	}

	if monthsBack <= 0 {
		monthsBack = 12
	}

	ageYears, _ := child.CalculateAge(time.Now())
	velocity, _ := models.CalculateAnnualGrowthVelocity(records, monthsBack)
	status, alertLevel, action := models.EvaluateGrowthVelocityWithAlert(ageYears, velocity, child.Gender)

	// 查找参考记录
	latest := records[len(records)-1]
	cutoffDate := latest.MeasureDate.AddDate(0, -monthsBack, 0)
	var prevHeight float64
	for i := len(records) - 1; i >= 0; i-- {
		if records[i].MeasureDate.Before(cutoffDate) || records[i].MeasureDate.Equal(cutoffDate) {
			prevHeight = records[i].Height
			break
		}
	}
	if prevHeight == 0 {
		prevHeight = records[0].Height
	}

	// 期望值
	var expectedMin float64
	switch {
	case ageYears < 2:
		expectedMin = 7.0
	case ageYears < 10:
		expectedMin = 5.0
	default:
		if child.Gender == "male" {
			expectedMin = 6.0
		} else {
			expectedMin = 5.0
		}
	}

	return &models.GrowthVelocityResponse{
		Velocity:       velocity,
		MonthsBack:     monthsBack,
		LatestHeight:   latest.Height,
		PreviousHeight: prevHeight,
		ExpectedMin:    expectedMin,
		Status:         status,
		AlertLevel:     alertLevel,
		Action:         action,
		Deviation:      models.Round(expectedMin-velocity, 1),
	}, nil
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

func buildEnvironmentAssessmentResponse(assessment *models.EnvironmentAssessment, envScore *models.EnvironmentScoreResult, prediction *models.ComprehensivePredictionResult, actionPlan *models.WeeklyActionPlan) *models.EnvironmentAssessmentResponse {
	return &models.EnvironmentAssessmentResponse{
		ID:                    assessment.ID,
		ChildID:               assessment.ChildID,
		AssessmentDate:        assessment.AssessmentDate.Format("2006-01-02"),
		ModuleScores:          envScore.ModuleScores,
		TotalScore:            envScore.TotalScore,
		MaxPossibleScore:      envScore.MaxPossibleScore,
		InterventionZone:      envScore.InterventionZone,
		ZoneLabel:             getZoneLabel(envScore.InterventionZone),
		Interpretation:        envScore.Interpretation,
		GeneticTargetHeight:   prediction.GeneticTargetHeight,
		EnvironmentIncrement:  prediction.EnvironmentIncrement,
		PredictedHeight:       prediction.PredictedHeight,
		PredictionMethod:      prediction.PredictionMethod,
		ErrorRange:            prediction.ErrorRange,
		KhamisRoche:           prediction.KhamisRoche,
		AgeWeights:            prediction.AgeWeights,
		ActionPlan:            actionPlan,
		ClinicalInterpretation: prediction.ClinicalInterpretation,
	}
}

func getZoneLabel(zone string) string {
	switch zone {
	case "high":
		return "优秀"
	case "medium":
		return "良好"
	case "low":
		return "需关注"
	default:
		return "待评估"
	}
}

func getZoneInterpretation(zone string, score float64) string {
	switch zone {
	case "high":
		return fmt.Sprintf("得分%.1f分(高分区): 趋近遗传上限，维持现状，每6-12月复评", score)
	case "medium":
		return fmt.Sprintf("得分%.1f分(中分区): 针对性改善薄弱环节，每3-6月复评", score)
	case "low":
		return fmt.Sprintf("得分%.1f分(低分区): 全面干预+专科评估，每月复评", score)
	default:
		return ""
	}
}

func evaluatePotentialStatus(currentHeight, targetHeight float64, currentPct int) string {
	diff := currentHeight - targetHeight
	switch {
	case diff >= -3 && diff <= 3:
		return "遗传潜力正常发挥"
	case diff > 3:
		return "超越遗传潜力"
	case diff > -5:
		return "接近遗传潜力下限"
	default:
		return "遗传潜力未充分发挥"
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func generateWeeklyActionPlan(moduleScores map[string]models.ModuleScore) *models.WeeklyActionPlan {
	// 按得分率排序，找出最薄弱的模块
	type moduleInfo struct {
		name       string
		percentage float64
		score      float64
		maxScore   float64
	}

	modules := []moduleInfo{
		{"nutrition", moduleScores["nutrition"].Percentage, moduleScores["nutrition"].Score, moduleScores["nutrition"].MaxScore},
		{"sleep", moduleScores["sleep"].Percentage, moduleScores["sleep"].Score, moduleScores["sleep"].MaxScore},
		{"exercise", moduleScores["exercise"].Percentage, moduleScores["exercise"].Score, moduleScores["exercise"].MaxScore},
		{"health", moduleScores["health"].Percentage, moduleScores["health"].Score, moduleScores["health"].MaxScore},
		{"mental", moduleScores["mental"].Percentage, moduleScores["mental"].Score, moduleScores["mental"].MaxScore},
	}

	// 按得分率升序排序（最差的在前）
	for i := 0; i < len(modules); i++ {
		for j := i + 1; j < len(modules); j++ {
			if modules[j].percentage < modules[i].percentage {
				modules[i], modules[j] = modules[j], modules[i]
			}
		}
	}

	plan := &models.WeeklyActionPlan{
		TopPriorities: make([]models.ActionPlanItem, 0, 3),
	}

	// 生成前3个优先行动
	for i := 0; i < 3 && i < len(modules); i++ {
		m := modules[i]
		if m.percentage >= 0.8 {
			continue // 得分率80%以上不列为优先
		}
		item := buildActionPlanItem(i+1, m.name, m.score, m.maxScore)
		plan.TopPriorities = append(plan.TopPriorities, item)
	}

	// 各模块计划
	plan.NutritionPlan = getModulePlan("nutrition", moduleScores["nutrition"].Percentage)
	plan.SleepPlan = getModulePlan("sleep", moduleScores["sleep"].Percentage)
	plan.ExercisePlan = getModulePlan("exercise", moduleScores["exercise"].Percentage)
	plan.HealthPlan = getModulePlan("health", moduleScores["health"].Percentage)
	plan.MentalPlan = getModulePlan("mental", moduleScores["mental"].Percentage)
	plan.TrackReminder = "建议每月复评一次，持续追踪改善效果"

	return plan
}

func buildActionPlanItem(priority int, module string, score, maxScore float64) models.ActionPlanItem {
	items := map[string][]models.ActionPlanItem{
		"nutrition": {
			{Priority: 1, Module: "nutrition", Title: "每天早餐加一杯牛奶+一个鸡蛋", Description: "确保蛋白质和钙质摄入，这是最容易执行的起点", Why: "蛋白质是生长激素合成的原料，钙是骨骼矿化的基础", HowToStart: "明天早餐开始，固定200ml牛奶+1个鸡蛋", Difficulty: "easy"},
			{Priority: 2, Module: "nutrition", Title: "每天保证5种颜色的食物", Description: "白(米/奶)、绿(蔬菜)、红(肉/番茄)、黄(蛋/玉米)、紫/深色", Why: "食物多样性确保微量营养素全面覆盖", HowToStart: "记录今天吃了几种颜色，明天补缺少的颜色", Difficulty: "easy"},
			{Priority: 3, Module: "nutrition", Title: "减少高糖零食，替换为水果和坚果", Description: "肥胖会加速骨龄进展，压缩生长时间", Why: "高糖高脂零食导致胰岛素抵抗，影响生长激素分泌", HowToStart: "本周不买薯片和奶茶，准备苹果和核桃", Difficulty: "medium"},
		},
		"sleep": {
			{Priority: 1, Module: "sleep", Title: "今晚21:30前上床熄灯", Description: "22:00-02:00是生长激素分泌黄金窗口", Why: "错过这个窗口无法弥补，深睡眠时分泌量占全天50-70%", HowToStart: "21:00关电视/手机，21:30准时熄灯", Difficulty: "medium"},
			{Priority: 2, Module: "sleep", Title: "建立睡前固定仪式", Description: "洗澡→关灯→轻音乐/故事书，固定流程帮助大脑进入睡眠模式", Why: "条件反射帮助大脑快速进入睡眠状态，缩短入睡时间", HowToStart: "今晚开始执行，连续7天养成习惯", Difficulty: "easy"},
			{Priority: 3, Module: "sleep", Title: "卧室调整为黑暗、安静、18-22℃", Description: "全遮光窗帘+静音环境+适宜温度", Why: "光线和噪音会中断深睡眠，温度不适影响睡眠质量", HowToStart: "今晚检查卧室环境，添置遮光窗帘", Difficulty: "easy"},
		},
		"exercise": {
			{Priority: 1, Module: "exercise", Title: "每天跳绳10分钟", Description: "跳绳是性价比最高的增高运动，10分钟=30分钟跑步的骨骼刺激", Why: "纵向弹跳对骨骺板的机械刺激促进软骨细胞增殖", HowToStart: "今晚开始，从100个起步，逐步增加", Difficulty: "easy"},
			{Priority: 2, Module: "exercise", Title: "每周3次摸高或吊单杠", Description: "每次3组×30秒悬挂，伸展脊柱间隙", Why: "改善姿势性身高损失，拉伸脊柱椎间盘", HowToStart: "找门框或单杠，每天放学做3组", Difficulty: "easy"},
			{Priority: 3, Module: "exercise", Title: "周末增加1次游泳或篮球", Description: "全身伸展+弹跳结合，比单一运动效果更好", Why: "游泳伸展全身，篮球结合弹跳和伸展", HowToStart: "本周六安排1小时游泳或篮球", Difficulty: "medium"},
		},
		"health": {
			{Priority: 1, Module: "health", Title: "预约年度体检", Description: "必查：身高体重(绘制生长曲线)、血常规(排贫血)、甲状腺功能", Why: "慢性病和贫血是生长的隐形杀手，早发现早干预", HowToStart: "本周内预约儿科体检", Difficulty: "easy"},
			{Priority: 2, Module: "health", Title: "记录身高，观察生长趋势", Description: "每月固定日期测量身高，绘制生长曲线", Why: "生长曲线比单次测量更有价值，可及早发现生长减速", HowToStart: "每月1号早晨测量并记录", Difficulty: "easy"},
			{Priority: 3, Module: "health", Title: "如长期服药，咨询医生对生长的影响", Description: "激素类药物可能抑制生长", Why: "地塞米松、泼尼松等药物可能抑制软骨细胞增殖", HowToStart: "下次复诊时主动询问医生", Difficulty: "easy"},
		},
		"mental": {
			{Priority: 1, Module: "mental", Title: "每天15分钟专注陪伴", Description: "不看手机，全身心陪孩子做他喜欢的事", Why: "安全感降低皮质醇，皮质醇直接拮抗生长激素", HowToStart: "今晚开始，饭后15分钟亲子游戏或聊天", Difficulty: "easy"},
			{Priority: 2, Module: "mental", Title: "减少学业压力，优先保证睡眠", Description: "成绩可以补，但身高窗口不可逆", Why: "长期压力导致慢性皮质醇升高，抑制生长激素分泌", HowToStart: "重新评估课外班数量，砍掉1-2个", Difficulty: "medium"},
			{Priority: 3, Module: "mental", Title: "如情绪问题持续2周以上，寻求专业帮助", Description: "联系学校心理老师或儿童心理咨询", Why: "临床焦虑/抑郁需要专业干预，不是'想开点'就能解决的", HowToStart: "本周联系学校心理老师或预约儿童心理门诊", Difficulty: "medium"},
		},
	}

	list, ok := items[module]
	if !ok || len(list) == 0 {
		return models.ActionPlanItem{}
	}

	// 根据得分选择不同优先级的建议
	percentage := score / maxScore
	idx := 0
	if percentage < 0.4 && len(list) > 2 {
		idx = 2 // 得分很低，给最难但最有效的
	} else if percentage < 0.6 && len(list) > 1 {
		idx = 1 // 得分中等，给中等难度的
	}

	item := list[idx]
	item.Priority = priority
	return item
}

func getModulePlan(module string, percentage float64) []string {
	if percentage >= 0.8 {
		return []string{"当前表现优秀，继续保持即可"}
	}

	plans := map[string][]string{
		"nutrition": {
			"每天早餐固定一杯200ml牛奶",
			"每天保证2种蔬菜+1种水果",
			"每周吃2次鱼或虾",
			"减少炸鸡、薯条、奶茶等高糖高脂零食",
		},
		"sleep": {
			"21:30前上床，22:00前入睡",
			"睡前1小时不用电子设备",
			"卧室保持黑暗、安静、18-22℃",
			"午睡不超过30分钟",
		},
		"exercise": {
			"每天跳绳10分钟或摸高50次",
			"每周2次游泳或篮球",
			"每周设1-2天完全休息",
			"运动后30分钟内补充牛奶+香蕉",
		},
		"health": {
			"每年至少1次体检(身高体重+血常规)",
			"每月固定日期测量身高",
			"如有慢性病，定期复查控制情况",
		},
		"mental": {
			"每天15分钟专注陪伴",
			"学业压力大时优先保证睡眠",
			"家庭冲突避免当着孩子面发生",
			"情绪持续低落超过2周寻求专业帮助",
		},
	}

	list, ok := plans[module]
	if !ok {
		return []string{}
	}
	return list
}
