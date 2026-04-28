package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/growth-tracker-pro-backend/internal/models"
	"github.com/growth-tracker-pro-backend/internal/repository"
)

// GrowthAgent 生长追踪AI Agent - 核心智能体
type GrowthAgent struct {
	profileBuilder    *ProfileBuilder
	medicalGuard      *MedicalGuard
	recommendationEng *RecommendationEngine
	repo              repository.Repository
}

// NewGrowthAgent 创建生长追踪AI Agent
func NewGrowthAgent(repo repository.Repository) *GrowthAgent {
	return &GrowthAgent{
		profileBuilder: NewProfileBuilder(),
		medicalGuard:   NewMedicalGuard(),
		repo:           repo,
	}
}

// AgentResponse Agent响应
type AgentResponse struct {
	Content      string            `json:"content"`
	Profile      *ChildProfile     `json:"profile,omitempty"`
	Recommendations []Recommendation `json:"recommendations,omitempty"`
	MedicalAlert *MedicalAlert     `json:"medical_alert,omitempty"`
	Tokens       int               `json:"tokens"`
	SessionID    string            `json:"session_id"`
}

// ChatRequest 聊天请求
type ChatRequest struct {
	UserID     string
	ChildID    string
	Message    string
	Context    []models.AIChatMessage
	Mode       string // "auto" / "profile" / "recommendation" / "report"
}

// Chat 执行AI对话
func (a *GrowthAgent) Chat(ctx context.Context, req *ChatRequest) (*AgentResponse, error) {
	// 1. 获取宝宝数据并构建立体信息表
	profile, err := a.buildChildProfile(ctx, req.UserID, req.ChildID)
	if err != nil {
		return nil, fmt.Errorf("构建宝宝信息表失败: %w", err)
	}

	// 2. 确定对话模式
	mode := a.determineMode(req.Message)

	// 3. 根据模式生成响应
	var content string
	var recommendations []Recommendation

	switch mode {
	case "profile":
		content = a.generateProfileAnalysis(profile)
	case "recommendation":
		engine := NewRecommendationEngine(profile, &models.Child{})
		recommendations = engine.GenerateRecommendations()
		content = a.formatRecommendations(recommendations)
	case "report":
		engine := NewRecommendationEngine(profile, &models.Child{})
		content = engine.GenerateSummaryReport()
	case "auto":
		fallthrough
	default:
		content = a.generateContextualResponse(req.Message, profile)
	}

	// 4. 医疗合规检查
	cleanContent, alert := a.medicalGuard.CheckAndSanitize(content)

	// 5. 保存对话历史
	sessionID := fmt.Sprintf("session_%s", req.UserID)
	a.saveConversation(ctx, req.UserID, req.ChildID, sessionID, req.Message, cleanContent)

	return &AgentResponse{
		Content:        cleanContent,
		Profile:        profile,
		Recommendations: recommendations,
		MedicalAlert:   alert,
		Tokens:         len(req.Message) / 4,
		SessionID:      sessionID,
	}, nil
}

// GetProfile 获取宝宝立体信息表
func (a *GrowthAgent) GetProfile(ctx context.Context, userID, childID string) (*ChildProfile, error) {
	return a.buildChildProfile(ctx, userID, childID)
}

// GetRecommendations 获取个性化建议
func (a *GrowthAgent) GetRecommendations(ctx context.Context, userID, childID string) ([]Recommendation, error) {
	profile, err := a.buildChildProfile(ctx, userID, childID)
	if err != nil {
		return nil, err
	}

	engine := NewRecommendationEngine(profile, &models.Child{})
	return engine.GenerateRecommendations(), nil
}

// GetDailyPlan 获取每日计划
func (a *GrowthAgent) GetDailyPlan(ctx context.Context, userID, childID string) (*DailyPlan, error) {
	profile, err := a.buildChildProfile(ctx, userID, childID)
	if err != nil {
		return nil, err
	}

	child, _ := a.repo.GetChildByID(ctx, childID)
	if child == nil {
		child = &models.Child{}
	}

	engine := NewRecommendationEngine(profile, child)
	return engine.GenerateDailyPlan(), nil
}

// GetWeeklyPlan 获取每周计划
func (a *GrowthAgent) GetWeeklyPlan(ctx context.Context, userID, childID string) (*WeeklyPlan, error) {
	profile, err := a.buildChildProfile(ctx, userID, childID)
	if err != nil {
		return nil, err
	}

	child, _ := a.repo.GetChildByID(ctx, childID)
	if child == nil {
		child = &models.Child{}
	}

	engine := NewRecommendationEngine(profile, child)
	return engine.GenerateWeeklyPlan(), nil
}

// GetSummaryReport 获取综合评估报告
func (a *GrowthAgent) GetSummaryReport(ctx context.Context, userID, childID string) (string, error) {
	profile, err := a.buildChildProfile(ctx, userID, childID)
	if err != nil {
		return "", err
	}

	child, _ := a.repo.GetChildByID(ctx, childID)
	if child == nil {
		child = &models.Child{}
	}

	engine := NewRecommendationEngine(profile, child)
	return engine.GenerateSummaryReport(), nil
}

// ParseLabReport 解析化验单
func (a *GrowthAgent) ParseLabReport(ctx context.Context, userID, childID, ocrText, reportType string) (*AIReportResult, error) {
	// 1. 获取宝宝信息
	profile, err := a.buildChildProfile(ctx, userID, childID)
	if err != nil {
		return nil, err
	}

	// 2. 模拟AI解析（实际需要调用AI API）
	result := a.mockLabReportAnalysis(ocrText, reportType, profile)

	// 3. 保存化验单记录
	report := &models.LabReport{
		ChildID:    childID,
		UserID:     userID,
		OCRText:    ocrText,
		ReportType: reportType,
	}
	resultJSON, _ := json.Marshal(result)
	report.AIResult = string(resultJSON)
	a.repo.CreateLabReport(ctx, report)

	return result, nil
}

// ========== 私有方法 ==========

func (a *GrowthAgent) buildChildProfile(ctx context.Context, userID, childID string) (*ChildProfile, error) {
	// 获取宝宝信息
	child, err := a.repo.GetChildByID(ctx, childID)
	if err != nil || child == nil {
		return nil, fmt.Errorf("宝宝不存在")
	}

	// 验证用户权限
	if child.UserID != userID {
		// 检查是否在家庭中
		family, _ := a.repo.GetFamilyByUserID(ctx, userID)
		if family == nil {
			return nil, fmt.Errorf("无权限访问该宝宝信息")
		}
	}

	// 获取生长记录
	records, _, _ := a.repo.GetRecordsByChildID(ctx, childID, "", "", 1, 100)

	// 获取化验单
	var reports []models.LabReport
	// TODO: 添加获取化验单的方法

	// 构建立体信息表
	return a.profileBuilder.Build(child, records, reports), nil
}

func (a *GrowthAgent) determineMode(message string) string {
	msg := strings.ToLower(message)

	// 关键词检测
	if strings.Contains(msg, "信息表") || strings.Contains(msg, "档案") || strings.Contains(msg, "评估") {
		return "profile"
	}
	if strings.Contains(msg, "建议") || strings.Contains(msg, "怎么做") || strings.Contains(msg, "怎么办") {
		return "recommendation"
	}
	if strings.Contains(msg, "报告") || strings.Contains(msg, "总结") || strings.Contains(msg, "分析") {
		return "report"
	}

	return "auto"
}

func (a *GrowthAgent) generateProfileAnalysis(profile *ChildProfile) string {
	var sb strings.Builder

	sb.WriteString("📋 " + profile.BasicInfo.Name + " 的生长发育档案\n\n")
	sb.WriteString("═══════════════════════════════════════\n\n")

	// 基础信息
	sb.WriteString("【基础信息】\n")
	sb.WriteString(fmt.Sprintf("- 性别: %s\n", profile.BasicInfo.GenderLabel))
	sb.WriteString(fmt.Sprintf("- 年龄: %s\n", profile.BasicInfo.AgeStr))
	sb.WriteString(fmt.Sprintf("- 父亲身高: %.1f cm\n", profile.BasicInfo.FatherHeight))
	sb.WriteString(fmt.Sprintf("- 母亲身高: %.1f cm\n\n", profile.BasicInfo.MotherHeight))

	// 发育评估
	sb.WriteString("【发育评估】\n")
	sb.WriteString(fmt.Sprintf("- 当前百分位: P%d\n", profile.GrowthAssessment.CurrentPercentile))
	sb.WriteString(fmt.Sprintf("- 百分位状态: %s\n", translateStatus(profile.GrowthAssessment.PercentileStatus)))
	sb.WriteString(fmt.Sprintf("- 靶身高范围: %.1f - %.1f cm\n",
		profile.GrowthAssessment.TargetHeight.MinHeight,
		profile.GrowthAssessment.TargetHeight.MaxHeight))
	sb.WriteString(fmt.Sprintf("- 测量记录: %d条 (%s)\n\n",
		profile.GrowthAssessment.RecordsCount,
		profile.GrowthAssessment.MeasurementFrequency))

	// 生长趋势
	sb.WriteString("【生长趋势】\n")
	sb.WriteString(fmt.Sprintf("- 年生长速度: %.1f cm/年\n", profile.GrowthTrend.Velocity))
	sb.WriteString(fmt.Sprintf("- 速度状态: %s\n", translateVelocity(profile.GrowthTrend.VelocityStatus)))
	sb.WriteString(fmt.Sprintf("- 趋势方向: %s\n\n", translateTrend(profile.GrowthTrend.TrendDirection)))

	// 营养状态
	sb.WriteString("【营养状态】\n")
	sb.WriteString(fmt.Sprintf("- 评分: %d分 (%s)\n",
		profile.NutritionStatus.Score,
		translateLevel(profile.NutritionStatus.Level)))
	if len(profile.NutritionStatus.Strengths) > 0 {
		sb.WriteString("✓ " + strings.Join(profile.NutritionStatus.Strengths, "\n✓ ") + "\n")
	}
	if len(profile.NutritionStatus.Concerns) > 0 {
		sb.WriteString("⚠️ " + strings.Join(profile.NutritionStatus.Concerns, "\n⚠️ ") + "\n")
	}
	sb.WriteString("\n")

	// 生活方式
	sb.WriteString("【生活方式】\n")
	sb.WriteString(fmt.Sprintf("- 运动评分: %d分\n", profile.LifestyleFactors.ExerciseStatus.Score))
	sb.WriteString(fmt.Sprintf("- 睡眠推荐: %.0f小时/天\n\n", profile.LifestyleFactors.SleepStatus.RecommendedHours))

	// 健康风险
	if len(profile.HealthRisks) > 0 {
		sb.WriteString("【健康关注点】\n")
		for _, risk := range profile.HealthRisks {
			icon := "⚠️"
			if risk.Level == "high" || risk.Level == "critical" {
				icon = "🔴"
			}
			sb.WriteString(fmt.Sprintf("%s %s: %s\n", icon, risk.Indicator, risk.Description))
		}
		sb.WriteString("\n")
	}

	// 干预优先级
	sb.WriteString("【干预优先级】\n")
	sb.WriteString(fmt.Sprintf("- 营养: %d分\n", profile.PriorityScores.Nutrition))
	sb.WriteString(fmt.Sprintf("- 运动: %d分\n", profile.PriorityScores.Exercise))
	sb.WriteString(fmt.Sprintf("- 睡眠: %d分\n", profile.PriorityScores.Sleep))
	sb.WriteString(fmt.Sprintf("- 医学检查: %d分\n\n", profile.PriorityScores.Medical))

	sb.WriteString("═══════════════════════════════════════\n")
	sb.WriteString(fmt.Sprintf("📅 报告生成时间: %s\n", profile.GeneratedAt.Format("2006-01-02 15:04")))

	return sb.String()
}

func (a *GrowthAgent) generateContextualResponse(message string, profile *ChildProfile) string {
	msg := strings.ToLower(message)

	// 问题类型识别
	if strings.Contains(msg, "正常") || strings.Contains(msg, "发育") {
		return a.answerAboutDevelopment(profile)
	}
	if strings.Contains(msg, "矮") || strings.Contains(msg, "偏矮") {
		return a.answerAboutShortStature(profile)
	}
	if strings.Contains(msg, "营养") || strings.Contains(msg, "吃") {
		return a.answerAboutNutrition(profile)
	}
	if strings.Contains(msg, "运动") || strings.Contains(msg, "锻炼") {
		return a.answerAboutExercise(profile)
	}
	if strings.Contains(msg, "睡眠") {
		return a.answerAboutSleep(profile)
	}
	if strings.Contains(msg, "靶身高") || strings.Contains(msg, "预测") {
		return a.answerAboutTargetHeight(profile)
	}
	if strings.Contains(msg, "骨龄") || strings.Contains(msg, "化验") {
		return a.answerAboutLabReports(profile)
	}

	// 默认回复
	return a.generateDefaultResponse(profile)
}

func (a *GrowthAgent) answerAboutDevelopment(profile *ChildProfile) string {
	var sb strings.Builder

	sb.WriteString("根据宝宝的生长发育数据分析：\n\n")

	if profile.GrowthAssessment.PercentileStatus == "normal" {
		sb.WriteString("✅ 宝宝目前的生长发育处于正常范围。\n")
		sb.WriteString(fmt.Sprintf("- 当前身高百分位为 P%d，在正常范围内\n", profile.GrowthAssessment.CurrentPercentile))
	} else {
		sb.WriteString("⚠️ 宝宝目前的身高百分位偏低，建议关注。\n")
	}

	sb.WriteString(fmt.Sprintf("- 生长速度为 %.1f cm/年，处于%s\n",
		profile.GrowthTrend.Velocity,
		translateVelocity(profile.GrowthTrend.VelocityStatus)))
	sb.WriteString(fmt.Sprintf("- 建议继续保持定期测量，关注生长趋势\n"))

	return sb.String()
}

func (a *GrowthAgent) answerAboutShortStature(profile *ChildProfile) string {
	var sb strings.Builder

	sb.WriteString("关于身高问题，我来帮您分析：\n\n")

	// 靶身高分析
	targetMin := profile.GrowthAssessment.TargetHeight.MinHeight
	targetMax := profile.GrowthAssessment.TargetHeight.MaxHeight
	sb.WriteString(fmt.Sprintf("📊 遗传靶身高范围: %.1f - %.1f cm\n", targetMin, targetMax))

	// 当前状态
	if profile.GrowthAssessment.CurrentPercentile >= 15 {
		sb.WriteString("\n✅ 宝宝目前身高处于正常范围。\n")
		sb.WriteString("继续保持均衡营养、充足睡眠和适量运动即可。\n")
	} else if profile.GrowthAssessment.CurrentPercentile >= 3 {
		sb.WriteString("\n⚠️ 宝宝身高处于偏低水平。\n")
		sb.WriteString("建议从以下几个方面改善：\n")
		sb.WriteString("1. 营养：保证每天300-500ml牛奶\n")
		sb.WriteString("2. 运动：每天跳绳1000-2000个\n")
		sb.WriteString("3. 睡眠：保证充足睡眠\n")
	} else {
		sb.WriteString("\n🔴 宝宝身高明显偏低，建议尽早就医评估。\n")
		sb.WriteString("可能需要进行以下检查：\n")
		sb.WriteString("- 骨龄评估\n")
		sb.WriteString("- 生长激素检测\n")
		sb.WriteString("- 甲状腺功能检查\n")
	}

	return sb.String()
}

func (a *GrowthAgent) answerAboutNutrition(profile *ChildProfile) string {
	var sb strings.Builder

	sb.WriteString("🍽️ 关于营养建议：\n\n")

	// 推荐食物
	sb.WriteString("推荐食物：\n")
	for i, food := range profile.NutritionStatus.RecommendedFoods {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, food))
	}

	sb.WriteString("\n需要控制的食物：\n")
	for i, food := range profile.NutritionStatus.FoodsToLimit {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, food))
	}

	// 如果有担忧
	if len(profile.NutritionStatus.Concerns) > 0 {
		sb.WriteString("\n⚠️ 需要注意：\n")
		for _, concern := range profile.NutritionStatus.Concerns {
			sb.WriteString("- " + concern + "\n")
		}
	}

	return sb.String()
}

func (a *GrowthAgent) answerAboutExercise(profile *ChildProfile) string {
	var sb strings.Builder

	sb.WriteString("🏃 关于运动建议：\n\n")

	years := profile.BasicInfo.AgeInDays / 365

	if years < 3 {
		sb.WriteString("推荐运动（3岁以下）：\n")
		sb.WriteString("1. 户外活动，每天2小时以上\n")
		sb.WriteString("2. 跑跳游戏、攀爬\n")
		sb.WriteString("3. 球类玩具\n")
	} else if years < 6 {
		sb.WriteString("推荐运动（3-6岁）：\n")
		sb.WriteString("1. 跳绳（每天500-1000个）\n")
		sb.WriteString("2. 骑自行车\n")
		sb.WriteString("3. 游泳\n")
		sb.WriteString("4. 篮球\n")
	} else {
		sb.WriteString("推荐运动（6岁以上）：\n")
		sb.WriteString("1. 跳绳（每天1000-2000个）\n")
		sb.WriteString("2. 篮球、排球\n")
		sb.WriteString("3. 游泳\n")
		sb.WriteString("4. 摸高跳\n")
	}

	sb.WriteString("\n💡 运动可以刺激生长激素分泌，建议每天运动30-60分钟。\n")

	return sb.String()
}

func (a *GrowthAgent) answerAboutSleep(profile *ChildProfile) string {
	var sb strings.Builder

	sb.WriteString("😴 关于睡眠建议：\n\n")

	sb.WriteString(fmt.Sprintf("推荐睡眠时长：%.0f 小时/天\n\n", profile.LifestyleFactors.SleepStatus.RecommendedHours))

	sb.WriteString("睡眠对长高非常重要，因为生长激素在深睡眠时分泌最旺盛。\n\n")

	sb.WriteString("建议：\n")
	sb.WriteString("1. 制定固定的作息时间\n")
	sb.WriteString("2. 睡前1小时避免使用电子设备\n")
	sb.WriteString("3. 保持卧室安静、光线暗淡\n")
	sb.WriteString("4. 避免睡前剧烈运动\n")

	return sb.String()
}

func (a *GrowthAgent) answerAboutTargetHeight(profile *ChildProfile) string {
	var sb strings.Builder

	target := profile.GrowthAssessment.TargetHeight

	sb.WriteString("📏 关于靶身高：\n\n")

	sb.WriteString("靶身高是根据父母身高使用医学公式计算的遗传潜力身高。\n\n")

	sb.WriteString(fmt.Sprintf("根据父母身高计算：\n"))
	sb.WriteString(fmt.Sprintf("- 靶身高: %.1f cm\n", target.TargetHeight))
	sb.WriteString(fmt.Sprintf("- 遗传范围: %.1f - %.1f cm\n\n", target.MinHeight, target.MaxHeight))

	sb.WriteString("⚠️ 靶身高仅供参考，实际身高会受到以下因素影响：\n")
	sb.WriteString("1. 营养状况\n")
	sb.WriteString("2. 运动习惯\n")
	sb.WriteString("3. 睡眠质量\n")
	sb.WriteString("4. 疾病因素\n")
	sb.WriteString("5. 心理状态\n\n")

	sb.WriteString("建议关注生长速度，而非单次身高数值。\n")

	return sb.String()
}

func (a *GrowthAgent) answerAboutLabReports(profile *ChildProfile) string {
	var sb strings.Builder

	sb.WriteString("📋 关于化验单：\n\n")

	sb.WriteString("如果您有化验单需要分析，可以上传图片，我会帮您：\n\n")

	sb.WriteString("1. OCR识别化验单内容\n")
	sb.WriteString("2. AI智能解读关键指标\n")
	sb.WriteString("3. 结合宝宝情况给出建议\n")
	sb.WriteString("4. 标注需要关注的异常项\n\n")

	sb.WriteString("支持的化验单类型：\n")
	sb.WriteString("- 骨龄片\n")
	sb.WriteString("- 生长激素激发试验\n")
	sb.WriteString("- 甲状腺功能\n")
	sb.WriteString("- 性激素六项\n")
	sb.WriteString("- 血常规\n")
	sb.WriteString("- IGF-1\n")

	return sb.String()
}

func (a *GrowthAgent) generateDefaultResponse(profile *ChildProfile) string {
	var sb strings.Builder

	sb.WriteString("感谢您的提问！\n\n")

	sb.WriteString(fmt.Sprintf("%s 目前的生长发育状况：\n", profile.BasicInfo.Name))
	sb.WriteString(fmt.Sprintf("- 年龄: %s\n", profile.BasicInfo.AgeStr))
	sb.WriteString(fmt.Sprintf("- 身高百分位: P%d\n", profile.GrowthAssessment.CurrentPercentile))
	sb.WriteString(fmt.Sprintf("- 生长速度: %.1f cm/年\n", profile.GrowthTrend.Velocity))
	sb.WriteString(fmt.Sprintf("- 靶身高范围: %.1f-%.1f cm\n\n",
		profile.GrowthAssessment.TargetHeight.MinHeight,
		profile.GrowthAssessment.TargetHeight.MaxHeight))

	sb.WriteString("您可以问我以下问题：\n")
	sb.WriteString("1. 发育是否正常？\n")
	sb.WriteString("2. 身高偏矮怎么办？\n")
	sb.WriteString("3. 营养建议\n")
	sb.WriteString("4. 运动建议\n")
	sb.WriteString("5. 睡眠建议\n")
	sb.WriteString("6. 靶身高解读\n")
	sb.WriteString("7. 如何上传化验单？\n")

	return sb.String()
}

func (a *GrowthAgent) formatRecommendations(recommendations []Recommendation) string {
	var sb strings.Builder

	sb.WriteString("💡 基于宝宝情况，我给您以下建议：\n\n")

	for i, rec := range recommendations {
		if i >= 5 {
			sb.WriteString("...\n")
			break
		}

		icon := getCategoryIcon(rec.Category)
		sb.WriteString(fmt.Sprintf("%s %s\n", icon, rec.Title))
		sb.WriteString(fmt.Sprintf("   %s\n", rec.Description))
		sb.WriteString("   具体建议：\n")
		for j, action := range rec.Actions {
			if j >= 3 {
				break
			}
			sb.WriteString(fmt.Sprintf("   • %s\n", action))
		}
		sb.WriteString(fmt.Sprintf("   ⏰ %s\n\n", rec.Timeframe))
	}

	return sb.String()
}

func (a *GrowthAgent) saveConversation(ctx context.Context, userID, childID, sessionID, userMsg, assistantMsg string) {
	var messages []models.AIChatMessage
	messages = append(messages, models.AIChatMessage{Role: "user", Content: userMsg})
	messages = append(messages, models.AIChatMessage{Role: "assistant", Content: assistantMsg})

	messagesJSON, _ := json.Marshal(messages)
	conv := &models.AIConversation{
		UserID:    userID,
		ChildID:   childID,
		SessionID: sessionID,
		Messages:  string(messagesJSON),
	}

	existing, _ := a.repo.GetConversationBySessionID(ctx, sessionID)
	if existing != nil {
		conv.ID = existing.ID
		a.repo.UpdateConversation(ctx, conv)
	} else {
		a.repo.CreateConversation(ctx, conv)
	}
}

func (a *GrowthAgent) mockLabReportAnalysis(ocrText, reportType string, profile *ChildProfile) *AIReportResult {
	// 模拟AI解析（实际需要调用AI API）
	result := &AIReportResult{
		KeyIndicators: []KeyIndicator{},
		Analysis:      "根据化验单分析，结果显示：\n",
		Suggestions:   []string{},
	}

	// 根据报告类型生成不同内容
	switch reportType {
	case "bone_age":
		result.KeyIndicators = append(result.KeyIndicators,
			KeyIndicator{Name: "骨龄", Value: "与实际年龄相符", Status: "normal"})
		result.Analysis += "骨骼发育处于正常轨道。\n"
		result.Suggestions = append(result.Suggestions, "继续保持均衡营养和适量运动")

	case "hormone":
		result.KeyIndicators = append(result.KeyIndicators,
			KeyIndicator{Name: "生长激素", Value: "正常范围", Status: "normal"},
			KeyIndicator{Name: "甲状腺激素", Value: "正常", Status: "normal"})
		result.Analysis += "激素水平正常。\n"
		result.Suggestions = append(result.Suggestions, "建议定期复查")

	case "blood_routine":
		result.KeyIndicators = append(result.KeyIndicators,
			KeyIndicator{Name: "血红蛋白", Value: "正常", Status: "normal"})
		result.Analysis += "血常规检查无明显异常。\n"
		result.Suggestions = append(result.Suggestions, "继续保持均衡饮食")

	default:
		result.Analysis += "建议咨询专业医生获取详细解读。\n"
		result.Suggestions = append(result.Suggestions, "咨询医生")
	}

	// 添加通用建议
	result.Suggestions = append(result.Suggestions,
		"每3-6个月复查相关指标",
		"持续关注生长发育趋势")

	result.NormalRanges = map[string]string{
		"骨龄": "与实际年龄相差±1岁为正常",
	}

	return result
}

// ========== 辅助函数 ==========

func translateStatus(status string) string {
	switch status {
	case "normal":
		return "正常"
	case "attention":
		return "需要关注"
	case "warning":
		return "需要警惕"
	default:
		return status
	}
}

func translateLevel(level string) string {
	switch level {
	case "excellent":
		return "优秀"
	case "good":
		return "良好"
	case "average":
		return "一般"
	case "poor":
		return "较差"
	default:
		return level
	}
}

func translateVelocity(status string) string {
	switch status {
	case "optimal":
		return "理想"
	case "normal":
		return "正常"
	case "slow":
		return "偏慢"
	case "very_slow":
		return "过慢"
	default:
		return status
	}
}

func translateTrend(direction string) string {
	switch direction {
	case "accelerating":
		return "加速增长"
	case "stable":
		return "稳定增长"
	case "decelerating":
		return "增长放缓"
	default:
		return direction
	}
}

func getCategoryIcon(category string) string {
	switch category {
	case "nutrition":
		return "🍽️"
	case "exercise":
		return "🏃"
	case "sleep":
		return "😴"
	case "medical":
		return "🏥"
	case "lifestyle":
		return "💡"
	default:
		return "📌"
	}
}

// AIReportResult AI化验单解析结果
type AIReportResult struct {
	KeyIndicators []KeyIndicator        `json:"key_indicators"`
	NormalRanges  map[string]string     `json:"normal_ranges"`
	Analysis      string                `json:"analysis"`
	Suggestions   []string             `json:"suggestions"`
}

// KeyIndicator 关键指标
type KeyIndicator struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Status string `json:"status"` // normal/abnormal/critical
}
