package alert

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/growth-tracker-pro-backend/internal/models"
	"gorm.io/gorm"
)

// Input 预警引擎输入
type Input struct {
	Child          *models.Child
	LatestRecord   *models.GrowthRecord
	AllRecords     []models.GrowthRecord
	TargetHeight   models.TargetHeightInfo
	CurrentPct     int
	RegionalPct    int
	Region         string
	BoneAgeRecords []models.GrowthRecord
}

// Engine 预警引擎
type Engine struct {
	db *gorm.DB
}

// NewEngine 创建预警引擎
func NewEngine(db *gorm.DB) *Engine {
	return &Engine{db: db}
}

// Evaluate 评估所有预警维度，返回应创建的预警列表
func (e *Engine) Evaluate(input *Input) []*models.HeightAlert {
	var alerts []*models.HeightAlert

	if input.Child == nil {
		return alerts
	}

	alerts = append(alerts, e.checkTargetGap(input)...)
	alerts = append(alerts, e.checkRegionalDeviation(input)...)
	alerts = append(alerts, e.checkBoneAgeDeviation(input)...)
	alerts = append(alerts, e.checkStagnation(input)...)
	alerts = append(alerts, e.checkVelocitySlow(input)...)
	alerts = append(alerts, e.checkPercentileDrop(input)...)

	return e.deduplicate(alerts)
}

// SaveAlerts 保存预警到数据库（按维度去重，同一维度已有未解决预警则更新）
func (e *Engine) SaveAlerts(ctx context.Context, childID, userID string, alerts []*models.HeightAlert) error {
	if len(alerts) == 0 {
		return nil
	}

	for _, alert := range alerts {
		alert.ChildID = childID
		alert.UserID = userID

		// 检查该维度是否已有未解决且未忽略的预警
		var existing models.HeightAlert
		err := e.db.WithContext(ctx).
			Where("child_id = ? AND alert_type = ? AND is_dismissed = ? AND resolved_at IS NULL",
				childID, alert.AlertType, false).
			Order("created_at DESC").
			First(&existing).Error

		if err == nil {
			// 已有同类型预警：如果级别没变则不重复创建；级别升高则创建新预警，旧预警标记为已解决
			if existing.AlertLevel == alert.AlertLevel {
				continue
			}
			// 旧预警标记解决
			now := time.Now()
			e.db.WithContext(ctx).Model(&existing).Update("resolved_at", now)
		}

		if err := e.db.WithContext(ctx).Create(alert).Error; err != nil {
			return err
		}
	}
	return nil
}

// GetActiveAlerts 获取宝宝当前有效的预警（未忽略且未解决）
func (e *Engine) GetActiveAlerts(ctx context.Context, childID string) ([]models.HeightAlert, error) {
	var alerts []models.HeightAlert
	err := e.db.WithContext(ctx).
		Where("child_id = ? AND is_dismissed = ? AND resolved_at IS NULL", childID, false).
		Order("CASE alert_level WHEN 'danger' THEN 1 WHEN 'warning' THEN 2 WHEN 'info' THEN 3 ELSE 4 END, created_at DESC").
		Find(&alerts).Error
	return alerts, err
}

// GetSummary 获取预警摘要
func (e *Engine) GetSummary(ctx context.Context, childID string) (*models.AlertSummaryResponse, error) {
	alerts, err := e.GetActiveAlerts(ctx, childID)
	if err != nil {
		return nil, err
	}

	if len(alerts) == 0 {
		return &models.AlertSummaryResponse{
			HasActiveAlert: false,
			HighestLevel:   "",
			TopAlerts:      []models.AlertResponse{},
			TotalActive:    0,
		}, nil
	}

	// 找出最高级别
	highestLevel := "info"
	for _, a := range alerts {
		if a.AlertLevel == "danger" {
			highestLevel = "danger"
			break
		}
		if a.AlertLevel == "warning" && highestLevel == "info" {
			highestLevel = "warning"
		}
	}

	// 取前3条
	topCount := 3
	if len(alerts) < topCount {
		topCount = len(alerts)
	}
	topAlerts := make([]models.AlertResponse, topCount)
	for i := 0; i < topCount; i++ {
		topAlerts[i] = models.AlertResponse{
			HeightAlert:  alerts[i],
			CreatedAtAgo: timeAgo(alerts[i].CreatedAt),
		}
	}

	return &models.AlertSummaryResponse{
		HasActiveAlert: true,
		HighestLevel:   highestLevel,
		TopAlerts:      topAlerts,
		TotalActive:    len(alerts),
	}, nil
}

// DismissAlert 忽略预警
func (e *Engine) DismissAlert(ctx context.Context, alertID, reason string) error {
	updates := map[string]interface{}{
		"is_dismissed": true,
	}
	if reason != "" {
		updates["description"] = gorm.Expr("CONCAT(description, '\n[忽略原因]: ', ?)", reason)
	}
	return e.db.WithContext(ctx).Model(&models.HeightAlert{}).
		Where("id = ?", alertID).Updates(updates).Error
}

// MarkAlertRead 标记预警已读
func (e *Engine) MarkAlertRead(ctx context.Context, alertID string) error {
	return e.db.WithContext(ctx).Model(&models.HeightAlert{}).
		Where("id = ?", alertID).Update("is_read", true).Error
}

// GetChildAlertList 分页获取预警列表
func (e *Engine) GetChildAlertList(ctx context.Context, childID string, req *models.AlertListRequest) (*models.AlertListResponse, error) {
	query := e.db.WithContext(ctx).Model(&models.HeightAlert{}).
		Where("child_id = ? AND is_dismissed = ? AND resolved_at IS NULL", childID, false)

	if req.Level != "" {
		query = query.Where("alert_level = ?", req.Level)
	}

	var total int64
	query.Count(&total)

	var alerts []models.HeightAlert
	offset := (req.Page - 1) * req.PageSize
	if err := query.Order("CASE alert_level WHEN 'danger' THEN 1 WHEN 'warning' THEN 2 WHEN 'info' THEN 3 ELSE 4 END, created_at DESC").
		Offset(offset).Limit(req.PageSize).Find(&alerts).Error; err != nil {
		return nil, err
	}

	items := make([]models.AlertResponse, len(alerts))
	unreadCount := 0
	for i, a := range alerts {
		items[i] = models.AlertResponse{
			HeightAlert:  a,
			CreatedAtAgo: timeAgo(a.CreatedAt),
		}
		if !a.IsRead {
			unreadCount++
		}
	}

	return &models.AlertListResponse{
		Items:       items,
		Total:       total,
		UnreadCount: unreadCount,
		ActiveCount: len(items),
	}, nil
}

// ==================== 6个预警维度检查 ====================

// 维度1: 靶身高差距预警
func (e *Engine) checkTargetGap(input *Input) []*models.HeightAlert {
	if input.LatestRecord == nil || input.LatestRecord.ID == "" {
		return nil
	}

	tp := input.TargetHeight
	targetPct := models.GetTargetHeightPercentile(input.LatestRecord.Height, tp)

	var alerts []*models.HeightAlert
	ageInMonths := ageInMonths(input.Child)

	triggerID := ""
	if input.LatestRecord != nil {
		triggerID = input.LatestRecord.ID
	}

	if targetPct < 15 && ageInMonths >= 72 {
		alerts = append(alerts, &models.HeightAlert{
			AlertType:       models.AlertTargetGapLow,
			AlertLevel:      "danger",
			Title:           "遗传潜力严重不足",
			Description:     fmt.Sprintf("当前身高仅为靶身高范围的%d%%位置（靶身高%.1fcm，范围%.1f-%.1fcm）。建议尽早就医排查内分泌、营养等影响生长的因素。", targetPct, tp.TargetHeight, tp.MinHeight, tp.MaxHeight),
			Dimension:       "target_gap",
			MetricValue:     floatPtr(float64(targetPct)),
			Threshold:       floatPtr(15),
			TriggerRecordID: &triggerID,
		})
	} else if targetPct < 30 {
		level := "warning"
		if ageInMonths >= 96 && targetPct < 20 {
			level = "danger"
		}
		alerts = append(alerts, &models.HeightAlert{
			AlertType:       models.AlertTargetGapLow,
			AlertLevel:      level,
			Title:           "靶身高差距偏大",
			Description:     fmt.Sprintf("当前身高在靶身高范围的%d%%位置，低于理想水平（靶身高%.1fcm）。建议关注睡眠、运动、营养等后天因素。", targetPct, tp.TargetHeight),
			Dimension:       "target_gap",
			MetricValue:     floatPtr(float64(targetPct)),
			Threshold:       floatPtr(30),
			TriggerRecordID: &triggerID,
		})
	} else if targetPct < 50 && len(input.AllRecords) >= 2 && hasDecliningTrend(input) {
		alerts = append(alerts, &models.HeightAlert{
			AlertType:       models.AlertTargetGapLow,
			AlertLevel:      "info",
			Title:           "遗传潜力有提升空间",
			Description:     fmt.Sprintf("当前身高在靶身高范围的%d%%位置，百分位呈下降趋势。优化生活方式可能帮助接近遗传上限。", targetPct),
			Dimension:       "target_gap",
			MetricValue:     floatPtr(float64(targetPct)),
			Threshold:       floatPtr(50),
			TriggerRecordID: &triggerID,
		})
	}

	return alerts
}

// 维度2: 区域修正后偏差
func (e *Engine) checkRegionalDeviation(input *Input) []*models.HeightAlert {
	if input.Region == "" || input.LatestRecord == nil || input.LatestRecord.ID == "" {
		return nil
	}

	corr := models.GetRegionCorrection(input.Region)
	if corr == nil {
		return nil
	}

	diff := input.CurrentPct - input.RegionalPct
	var alerts []*models.HeightAlert

	triggerID := ""
	if input.LatestRecord != nil {
		triggerID = input.LatestRecord.ID
	}

	corrSign := "+"
	if corr.Correction < 0 {
		corrSign = ""
	}

	if input.RegionalPct < 3 && diff >= 10 {
		alerts = append(alerts, &models.HeightAlert{
			AlertType:       models.AlertRegionalShort,
			AlertLevel:      "warning",
			Title:           fmt.Sprintf("在%s地区身高明显偏矮", corr.ProvinceName),
			Description:     fmt.Sprintf("按%s地区标准修正后（%s%.1fcm），当前身高仅处于P%d，而全国标准为P%d。在本地区同龄儿童中明显偏矮，建议就医评估。", corr.ProvinceName, corrSign, corr.Correction, input.RegionalPct, input.CurrentPct),
			Dimension:       "regional",
			MetricValue:     floatPtr(float64(input.RegionalPct)),
			Threshold:       floatPtr(3),
			TriggerRecordID: &triggerID,
		})
	} else if input.RegionalPct < 10 && diff >= 7 {
		alerts = append(alerts, &models.HeightAlert{
			AlertType:       models.AlertRegionalShort,
			AlertLevel:      "info",
			Title:           fmt.Sprintf("在%s地区身高中下", corr.ProvinceName),
			Description:     fmt.Sprintf("按%s地区标准（%s%.1fcm）修正后处于P%d。虽在全国范围正常，但在本地同龄儿童中偏低。", corr.ProvinceName, corrSign, corr.Correction, input.RegionalPct),
			Dimension:       "regional",
			MetricValue:     floatPtr(float64(input.RegionalPct)),
			Threshold:       floatPtr(10),
			TriggerRecordID: &triggerID,
		})
	}

	return alerts
}

// 维度3: 骨龄偏差预警
func (e *Engine) checkBoneAgeDeviation(input *Input) []*models.HeightAlert {
	if input.LatestRecord == nil || input.LatestRecord.BoneAge == nil {
		return nil
	}

	boneAge := *input.LatestRecord.BoneAge

	// 获取骨龄差值（优先用记录中的，没有则自动计算）
	var boneAgeDiff float64
	if input.LatestRecord.BoneAgeDiff != nil {
		boneAgeDiff = *input.LatestRecord.BoneAgeDiff
	} else {
		ageInMonths := ageInMonths(input.Child)
		boneAgeDiff = boneAge - float64(ageInMonths)/12.0
	}

	absDiff := boneAgeDiff
	if absDiff < 0 {
		absDiff = -absDiff
	}
	if absDiff <= 1.0 {
		return nil
	}

	// 用骨龄值查该骨龄对应的身高标准
	boneAgeMonths := int(boneAge * 12)
	std := models.GetGrowthStandard(boneAgeMonths, input.Child.Gender)
	if std == nil {
		return nil
	}

	isShort := input.LatestRecord.Height < std.P3
	var alerts []*models.HeightAlert

	// 根据骨龄差的正负判断是提前还是落后
	alertType := models.AlertBoneAgeAdvanced
	if boneAgeDiff < 0 {
		alertType = models.AlertBoneAgeDelayed
	}

	if isShort {
		level := "warning"
		if absDiff >= 2.0 {
			level = "danger"
		}
		alerts = append(alerts, &models.HeightAlert{
			AlertType:   alertType,
			AlertLevel:  level,
			Title:       "骨龄偏差伴身高不足",
			Description: fmt.Sprintf("骨龄%.1f岁，与实际年龄偏差%.1f岁。按骨龄标准身高仅%.1fcm（P3线），当前身高%.1fcm未达标。建议儿科内分泌科就诊。", boneAge, absDiff, std.P3, input.LatestRecord.Height),
			Dimension:       "bone_age",
			MetricValue:     floatPtr(absDiff),
			Threshold:       floatPtr(1.0),
			TriggerRecordID: &input.LatestRecord.ID,
		})
	} else {
		alerts = append(alerts, &models.HeightAlert{
			AlertType:   alertType,
			AlertLevel:  "info",
			Title:       "骨龄偏差需关注",
			Description: fmt.Sprintf("骨龄%.1f岁，与实际年龄偏差%.1f岁。当前身高%.1fcm在骨龄标准中达标，但需持续监测骨龄进展。", boneAge, absDiff, input.LatestRecord.Height),
			Dimension:       "bone_age",
			MetricValue:     floatPtr(absDiff),
			Threshold:       floatPtr(1.0),
			TriggerRecordID: &input.LatestRecord.ID,
		})
	}

	return alerts
}

// 维度4: 猛涨期停滞预警
func (e *Engine) checkStagnation(input *Input) []*models.HeightAlert {
	if input.Child.GrowthStage == nil || *input.Child.GrowthStage != "puberty" {
		return nil
	}
	if input.Child.LastHeightChangeDate == nil {
		return nil
	}

	weeksSinceChange := int(time.Since(*input.Child.LastHeightChangeDate).Hours() / 24 / 7)
	if weeksSinceChange < 4 {
		return nil
	}

	level := "warning"
	if weeksSinceChange >= 8 {
		level = "danger"
	}

	triggerID := ""
	if input.LatestRecord != nil {
		triggerID = input.LatestRecord.ID
	}

	return []*models.HeightAlert{{
		AlertType:       models.AlertStagnation,
		AlertLevel:      level,
		Title:           "猛涨期身高增长停滞",
		Description:     fmt.Sprintf("已进入青春期%d周，期间身高无变化。青春期是身高增长的最后关键窗口，建议尽早就医评估生长激素水平。", weeksSinceChange),
		Dimension:       "stagnation",
		MetricValue:     floatPtr(float64(weeksSinceChange)),
		Threshold:       floatPtr(4),
		TriggerRecordID: &triggerID,
	}}
}

// 维度5: 生长速度过慢
func (e *Engine) checkVelocitySlow(input *Input) []*models.HeightAlert {
	if len(input.AllRecords) < 2 {
		return nil
	}

	// 按日期排序
	records := make([]models.GrowthRecord, len(input.AllRecords))
	copy(records, input.AllRecords)
	sort.Slice(records, func(i, j int) bool {
		return records[i].MeasureDate.Before(records[j].MeasureDate)
	})

	// 取最近两次
	latest := records[len(records)-1]
	prev := records[len(records)-2]

	days := int(latest.MeasureDate.Sub(prev.MeasureDate).Hours() / 24)
	if days < 30 {
		return nil
	}

	heightDiff := latest.Height - prev.Height
	annualGrowth := heightDiff / float64(days) * 365

	ageInYears := ageInYears(input.Child)
	expectedMin := 5.0
	switch {
	case ageInYears < 2:
		expectedMin = 7.0
	case ageInYears < 10:
		expectedMin = 5.0
	default:
		if input.Child.Gender == "male" {
			expectedMin = 6.0
		} else {
			expectedMin = 5.0
		}
	}

	if annualGrowth < expectedMin {
		triggerID := ""
		if input.LatestRecord != nil {
			triggerID = input.LatestRecord.ID
		}
		return []*models.HeightAlert{{
			AlertType:       models.AlertVelocitySlow,
			AlertLevel:      "warning",
			Title:           "生长速度过慢",
			Description:     fmt.Sprintf("近%d天长高%.1fcm，折算年增速约%.1fcm，低于该年龄段最低期望%.1fcm/年。建议检查营养摄入和睡眠质量。", days, heightDiff, annualGrowth, expectedMin),
			Dimension:       "velocity",
			MetricValue:     floatPtr(annualGrowth),
			Threshold:       floatPtr(expectedMin),
			TriggerRecordID: &triggerID,
		}}
	}

	return nil
}

// 维度6: 百分位持续下降
func (e *Engine) checkPercentileDrop(input *Input) []*models.HeightAlert {
	if len(input.AllRecords) < 3 {
		return nil
	}

	// 按日期排序
	records := make([]models.GrowthRecord, len(input.AllRecords))
	copy(records, input.AllRecords)
	sort.Slice(records, func(i, j int) bool {
		return records[i].MeasureDate.Before(records[j].MeasureDate)
	})

	// 计算每条记录的百分位
	var percentiles []int
	for _, r := range records {
		ageInMonths := ageInMonthsAt(input.Child, r.MeasureDate)
		pct := models.CalculateHeightPercentile(r.Height, ageInMonths, input.Child.Gender)
		percentiles = append(percentiles, pct)
	}

	// 检查早期百分位是否高于后期
	earlyAvg := avgPercentile(percentiles[:len(percentiles)/2])
	lateAvg := avgPercentile(percentiles[len(percentiles)/2:])

	if earlyAvg > lateAvg+5 {
		triggerID := ""
		if input.LatestRecord != nil {
			triggerID = input.LatestRecord.ID
		}
		return []*models.HeightAlert{{
			AlertType:       models.AlertPercentileDrop,
			AlertLevel:      "info",
			Title:           "身高百分位持续下降",
			Description:     fmt.Sprintf("近期身高百分位（平均P%d）较前期（平均P%d）有所下降。建议回顾近期饮食、睡眠、运动是否有变化。", lateAvg, earlyAvg),
			Dimension:       "velocity",
			MetricValue:     floatPtr(float64(lateAvg)),
			Threshold:       floatPtr(float64(earlyAvg)),
			TriggerRecordID: &triggerID,
		}}
	}

	return nil
}

// ==================== 辅助函数 ====================

func (e *Engine) deduplicate(alerts []*models.HeightAlert) []*models.HeightAlert {
	// 同一维度同类型只保留最高级别
	seen := make(map[string]*models.HeightAlert)
	for _, a := range alerts {
		key := a.AlertType + "_" + a.Dimension
		if existing, ok := seen[key]; ok {
			if levelRank(a.AlertLevel) > levelRank(existing.AlertLevel) {
				seen[key] = a
			}
		} else {
			seen[key] = a
		}
	}

	result := make([]*models.HeightAlert, 0, len(seen))
	for _, a := range seen {
		result = append(result, a)
	}
	return result
}

func levelRank(level string) int {
	switch level {
	case "danger":
		return 3
	case "warning":
		return 2
	case "info":
		return 1
	default:
		return 0
	}
}

func ageInMonths(child *models.Child) int {
	years, months := child.CalculateAge(time.Now())
	return years*12 + months
}

func ageInYears(child *models.Child) int {
	years, _ := child.CalculateAge(time.Now())
	return years
}

func ageInMonthsAt(child *models.Child, at time.Time) int {
	years, months := child.CalculateAge(at)
	return years*12 + months
}

func floatPtr(v float64) *float64 {
	return &v
}

func hasDecliningTrend(input *Input) bool {
	if len(input.AllRecords) < 2 {
		return false
	}
	records := make([]models.GrowthRecord, len(input.AllRecords))
	copy(records, input.AllRecords)
	sort.Slice(records, func(i, j int) bool {
		return records[i].MeasureDate.Before(records[j].MeasureDate)
	})

	var percentiles []int
	for _, r := range records {
		ageInMonths := ageInMonthsAt(input.Child, r.MeasureDate)
		pct := models.CalculateHeightPercentile(r.Height, ageInMonths, input.Child.Gender)
		percentiles = append(percentiles, pct)
	}

	// 简单判断：后期平均低于前期平均
	mid := len(percentiles) / 2
	earlyAvg := avgPercentile(percentiles[:mid])
	lateAvg := avgPercentile(percentiles[mid:])
	return earlyAvg > lateAvg
}

func avgPercentile(pcts []int) int {
	if len(pcts) == 0 {
		return 50
	}
	sum := 0
	for _, p := range pcts {
		sum += p
	}
	return sum / len(pcts)
}

func timeAgo(t time.Time) string {
	d := time.Since(t)
	if d < time.Hour {
		return "刚刚"
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%d小时前", int(d.Hours()))
	} else if d < 30*24*time.Hour {
		return fmt.Sprintf("%d天前", int(d.Hours()/24))
	} else if d < 365*24*time.Hour {
		return fmt.Sprintf("%d个月前", int(d.Hours()/24/30))
	}
	return fmt.Sprintf("%d年前", int(d.Hours()/24/365))
}
