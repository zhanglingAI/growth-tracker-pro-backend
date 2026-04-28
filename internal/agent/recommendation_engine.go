package agent

import (
	"fmt"
	"sort"
	"strings"

	"github.com/growth-tracker-pro/backend/internal/models"
)

// RecommendationEngine 推荐引擎 - 基于宝宝信息表生成个性化建议
type RecommendationEngine struct {
	profile *ChildProfile
	child   *models.Child
}

// Recommendation 推荐项
type Recommendation struct {
	Category    string   `json:"category"`    // nutrition/exercise/sleep/medical/lifestyle
	Priority    int      `json:"priority"`    // 1-5, 1最高
	Title       string   `json:"title"`       // 建议标题
	Description string   `json:"description"` // 详细说明
	Actions     []string `json:"actions"`     // 具体行动项
	Reason      string   `json:"reason"`      // 给出此建议的原因
	Timeframe   string   `json:"timeframe"`   // 实施时间框架
}

// DailyPlan 每日计划
type DailyPlan struct {
	Morning   []string `json:"morning"`
	Afternoon []string `json:"afternoon"`
	Evening   []string `json:"evening"`
}

// WeeklyPlan 每周计划
type WeeklyPlan struct {
	ExercisePlan    []string `json:"exercise_plan"`
	NutritionPlan   []string `json:"nutrition_plan"`
	Checklist       []string `json:"checklist"`
}

// NewRecommendationEngine 创建推荐引擎
func NewRecommendationEngine(profile *ChildProfile, child *models.Child) *RecommendationEngine {
	return &RecommendationEngine{
		profile: profile,
		child:   child,
	}
}

// GenerateRecommendations 生成个性化建议
func (e *RecommendationEngine) GenerateRecommendations() []Recommendation {
	var recommendations []Recommendation

	// 1. 基于优先级评分生成建议
	if e.profile.PriorityScores.Nutrition >= 60 {
		recommendations = append(recommendations, e.generateNutritionRecommendation())
	}

	if e.profile.PriorityScores.Exercise >= 60 {
		recommendations = append(recommendations, e.generateExerciseRecommendation())
	}

	if e.profile.PriorityScores.Sleep >= 60 {
		recommendations = append(recommendations, e.generateSleepRecommendation())
	}

	if e.profile.PriorityScores.Medical >= 70 {
		recommendations = append(recommendations, e.generateMedicalRecommendation())
	}

	// 2. 基于健康风险生成建议
	for _, risk := range e.profile.HealthRisks {
		recommendations = append(recommendations, e.generateRiskRecommendation(risk))
	}

	// 3. 基于生长趋势生成建议
	if e.profile.GrowthTrend.VelocityStatus == "slow" || e.profile.GrowthTrend.VelocityStatus == "very_slow" {
		recommendations = append(recommendations, e.generateGrowthVelocityRecommendation())
	}

	// 4. 基于生活方式评估生成建议
	if e.profile.LifestyleFactors.OverallScore < 60 {
		recommendations = append(recommendations, e.generateLifestyleRecommendation())
	}

	// 按优先级排序
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Priority < recommendations[j].Priority
	})

	return recommendations
}

// GenerateDailyPlan 生成每日计划
func (e *RecommendationEngine) GenerateDailyPlan() *DailyPlan {
	ageInDays := e.child.AgeInDays()
	years := ageInDays / 365

	plan := &DailyPlan{}

	// 根据年龄调整计划
	if years < 3 {
		plan.Morning = []string{
			"07:00-08:00 起床，早餐（牛奶+鸡蛋）",
			"09:00-11:00 户外活动，晒太阳",
		}
		plan.Afternoon = []string{
			"12:00-13:00 午餐（营养均衡）",
			"14:00-16:00 午睡",
			"16:00-18:00 亲子游戏",
		}
		plan.Evening = []string{
			"18:00-19:00 晚餐",
			"19:00-20:00 亲子互动",
			"20:00-21:00 洗漱，准备睡觉",
		}
	} else if years < 6 {
		plan.Morning = []string{
			"07:00-08:00 起床，早餐（牛奶+主食+鸡蛋）",
			"08:00-09:00 跳绳或户外跑跳",
		}
		plan.Afternoon = []string{
			"12:00-13:00 午餐（多吃蔬菜）",
			"13:00-14:30 午休",
			"15:00-17:30 运动时间（篮球/游泳/自行车）",
		}
		plan.Evening = []string{
			"18:00-19:00 晚餐",
			"19:00-20:00 亲子阅读",
			"20:00-21:00 洗漱，准备睡觉",
			"21:00 前入睡",
		}
	} else {
		plan.Morning = []string{
			"06:30-07:30 起床，早餐",
			"07:30-08:00 晨跑或跳绳500个",
		}
		plan.Afternoon = []string{
			"12:00-13:00 午餐",
			"13:00-14:00 午休",
			"16:00-18:30 运动时间（推荐：篮球、排球、游泳）",
		}
		plan.Evening = []string{
			"18:30-19:30 晚餐",
			"20:00-21:00 拉伸运动",
			"21:30-22:00 洗漱，准备睡觉",
			"22:00 前入睡",
		}
	}

	return plan
}

// GenerateWeeklyPlan 生成每周计划
func (e *RecommendationEngine) GenerateWeeklyPlan() *WeeklyPlan {
	plan := &WeeklyPlan{
		ExercisePlan: []string{
			"周一、周三、周五：跳绳（每天1000-2000个）",
			"周二、周四：篮球或游泳",
			"周六：户外活动（爬山、骑自行车）",
			"周日：休息或轻松活动",
		},
		NutritionPlan: []string{
			"每天：300-500ml牛奶",
			"每天：1-2个鸡蛋",
			"每天：50-100g肉类或鱼类",
			"每天：多种蔬菜水果",
			"每周：2-3次豆制品",
			"避免：甜饮料、油炸食品、零食",
		},
		Checklist: []string{
			"每周测量身高体重并记录",
			"观察孩子的食欲和精神状态",
			"确保充足的户外活动时间",
			"保持规律的作息时间",
			"关注睡眠质量和时长",
		},
	}

	return plan
}

// GenerateSummaryReport 生成综合评估报告
func (e *RecommendationEngine) GenerateSummaryReport() string {
	var sb strings.Builder

	// 基础信息
	sb.WriteString("📊 " + e.child.Name + " 生长发育综合评估报告\n\n")
	sb.WriteString("═══════════════════════════════════════\n\n")

	// 发育状态
	sb.WriteString("【发育状态】\n")
	if e.profile.GrowthAssessment.PercentileStatus == "normal" {
		sb.WriteString("✅ 当前身高处于正常百分位范围\n")
	} else {
		sb.WriteString("⚠️ 身高百分位需要关注\n")
	}
	sb.WriteString(fmt.Sprintf("- 当前身高百分位: P%d\n", e.profile.GrowthAssessment.CurrentPercentile))
	sb.WriteString(fmt.Sprintf("- 靶身高范围: %.1f-%.1f cm\n",
		e.profile.GrowthAssessment.TargetHeight.MinHeight,
		e.profile.GrowthAssessment.TargetHeight.MaxHeight))
	sb.WriteString(fmt.Sprintf("- 生长速度: %.1f cm/年 (%s)\n\n",
		e.profile.GrowthTrend.Velocity,
		e.translateVelocityStatus(e.profile.GrowthTrend.VelocityStatus)))

	// 营养状态
	sb.WriteString("【营养状态】\n")
	sb.WriteString(fmt.Sprintf("- 营养评分: %d分 (%s)\n",
		e.profile.NutritionStatus.Score,
		e.translateLevel(e.profile.NutritionStatus.Level)))
	if len(e.profile.NutritionStatus.Strengths) > 0 {
		sb.WriteString("✓ 做得好的: " + strings.Join(e.profile.NutritionStatus.Strengths, "、") + "\n")
	}
	if len(e.profile.NutritionStatus.Concerns) > 0 {
		sb.WriteString("⚠️ 需要改进: " + strings.Join(e.profile.NutritionStatus.Concerns, "、") + "\n")
	}
	sb.WriteString("\n")

	// 生活方式
	sb.WriteString("【生活方式】\n")
	sb.WriteString(fmt.Sprintf("- 运动状态: %d分 (%s)\n",
		e.profile.LifestyleFactors.ExerciseStatus.Score,
		e.translateLevel(e.profile.LifestyleFactors.ExerciseStatus.Level)))
	sb.WriteString(fmt.Sprintf("- 睡眠状态: %d分 (%s)\n",
		e.profile.LifestyleFactors.SleepStatus.Score,
		e.translateLevel(e.profile.LifestyleFactors.SleepStatus.Level)))
	sb.WriteString(fmt.Sprintf("- 推荐睡眠时长: %.0f小时\n\n",
		e.profile.LifestyleFactors.SleepStatus.RecommendedHours))

	// 健康风险
	if len(e.profile.HealthRisks) > 0 {
		sb.WriteString("【需要注意】\n")
		for _, risk := range e.profile.HealthRisks {
			level := "⚠️"
			if risk.Level == "high" || risk.Level == "critical" {
				level = "🔴"
			}
			sb.WriteString(fmt.Sprintf("%s %s\n", level, risk.Description))
			sb.WriteString(fmt.Sprintf("   建议: %s\n", risk.Action))
		}
		sb.WriteString("\n")
	}

	// 下一步建议
	sb.WriteString("【下一步行动建议】\n")
	recommendations := e.GenerateRecommendations()
	for i, rec := range recommendations {
		if i >= 3 {
			sb.WriteString("...\n")
			break
		}
		sb.WriteString(fmt.Sprintf("%d. %s (%s)\n", i+1, rec.Title, rec.Category))
	}

	sb.WriteString("\n═══════════════════════════════════════\n")
	sb.WriteString("⚠️ 本报告仅供参考，具体情况请咨询专业医生。\n")

	return sb.String()
}

// ========== 私有方法 ==========

func (e *RecommendationEngine) generateNutritionRecommendation() Recommendation {
	rec := Recommendation{
		Category: "nutrition",
		Priority: 1,
		Title:    "优化营养摄入",
		Actions:  []string{},
	}

	// 基于年龄给出具体建议
	years := e.child.AgeInDays() / 365

	if years < 3 {
		rec.Description = "这个年龄段的宝宝营养需求旺盛，合理的膳食结构对生长发育至关重要。"
		rec.Actions = []string{
			"保证每天300-500ml奶量",
			"辅食多样化，包括肉泥、鱼泥、蔬菜泥",
			"避免添加糖和盐",
		}
	} else if years < 6 {
		rec.Description = "学龄前儿童需要充足的蛋白质和钙质支持生长发育。"
		rec.Actions = []string{
			"每天1-2个鸡蛋",
			"适量红肉补铁",
			"多吃深色蔬菜",
			"水果作为加餐，避免果汁",
		}
	} else {
		rec.Description = "学龄期儿童应注重营养均衡，为青春期发育储备能量。"
		rec.Actions = []string{
			"每天300ml牛奶或等量奶制品",
			"保证优质蛋白质摄入",
			"多吃含钙食物（豆腐、虾皮等）",
			"控制零食和甜饮料",
		}
	}

	rec.Reason = fmt.Sprintf("当前营养评分为%d分，有提升空间。", e.profile.NutritionStatus.Score)
	rec.Timeframe = "立即开始，长期坚持"

	return rec
}

func (e *RecommendationEngine) generateExerciseRecommendation() Recommendation {
	years := e.child.AgeInDays() / 365

	rec := Recommendation{
		Category: "exercise",
		Priority: 2,
		Title:    "加强运动锻炼",
		Description: "适当的运动可以刺激生长激素分泌，促进骨骼生长。",
		Actions:  []string{},
		Reason:   fmt.Sprintf("当前运动评分为%d分。", e.profile.LifestyleFactors.ExerciseStatus.Score),
		Timeframe: "每天坚持",
	}

	if years < 3 {
		rec.Actions = []string{
			"每天2小时以上户外活动",
			"多进行跑跳、攀爬等游戏",
			"适当的球类玩具",
		}
	} else if years < 6 {
		rec.Actions = []string{
			"跳绳（从每天500个开始，逐步增加）",
			"骑自行车或滑板车",
			"游泳（全身运动，对关节刺激小）",
			"每天运动时间不少于1小时",
		}
	} else {
		rec.Actions = []string{
			"跳绳（每天1000-2000个）",
			"篮球（跳跃动作利于长高）",
			"游泳（伸展效果好）",
			"摸高跳（睡前10分钟）",
			"每次运动30-60分钟",
		}
	}

	return rec
}

func (e *RecommendationEngine) generateSleepRecommendation() Recommendation {
	rec := Recommendation{
		Category:    "sleep",
		Priority:    2,
		Title:       "保证充足睡眠",
		Description: "睡眠中生长激素分泌最旺盛，优质睡眠是长高的关键。",
		Actions: []string{
			"制定固定的作息时间",
			"睡前1小时避免使用电子设备",
			"保持卧室安静、黑暗",
			"睡前避免剧烈运动和情绪激动",
		},
		Reason:    fmt.Sprintf("推荐睡眠%.0f小时，实际应保证充足。", e.profile.LifestyleFactors.SleepStatus.RecommendedHours),
		Timeframe: "每天",
	}

	return rec
}

func (e *RecommendationEngine) generateMedicalRecommendation() Recommendation {
	rec := Recommendation{
		Category:    "medical",
		Priority:    1,
		Title:       "建议进行专业评估",
		Description: "基于当前数据，建议咨询专业医生进行进一步评估。",
		Actions: []string{
			"预约儿科或儿童内分泌科",
			"准备既往的生长记录",
			"如有化验单，带上相关报告",
		},
		Reason:    "存在需要关注的生长发育指标",
		Timeframe: "尽快（1-2周内）",
	}

	return rec
}

func (e *RecommendationEngine) generateRiskRecommendation(risk HealthRisk) Recommendation {
	rec := Recommendation{
		Category:    "lifestyle",
		Priority:    riskToPriority(risk.Level),
		Title:       "针对" + risk.Indicator + "的建议",
		Description: risk.Description,
		Actions:     []string{risk.Action},
		Reason:      fmt.Sprintf("检测到%s级别风险：%s", risk.Level, risk.Description),
		Timeframe:   "立即开始",
	}

	return rec
}

func (e *RecommendationEngine) generateGrowthVelocityRecommendation() Recommendation {
	rec := Recommendation{
		Category:    "medical",
		Priority:    1,
		Title:       "关注生长速度",
		Description: fmt.Sprintf("当前年生长速度为%.1f cm，低于正常水平。", e.profile.GrowthTrend.Velocity),
		Actions: []string{
			"增加营养摄入",
			"保证充足睡眠",
			"加强运动锻炼",
			"建议就医评估",
		},
		Reason:    "生长速度偏慢需要引起重视",
		Timeframe: "建议2周内就医",
	}

	return rec
}

func (e *RecommendationEngine) generateLifestyleRecommendation() Recommendation {
	rec := Recommendation{
		Category:    "lifestyle",
		Priority:    3,
		Title:       "改善整体生活方式",
		Description: "良好的生活习惯是健康成长的基础。",
		Actions: []string{
			"制定并执行规律的作息时间",
			"每天保证足够的户外活动时间",
			"保持均衡的饮食习惯",
			"家长以身作则",
		},
		Reason:    fmt.Sprintf("生活方式综合评分为%d分。", e.profile.LifestyleFactors.OverallScore),
		Timeframe: "逐步改善",
	}

	return rec
}

func (e *RecommendationEngine) translateLevel(level string) string {
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

func (e *RecommendationEngine) translateVelocityStatus(status string) string {
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

func riskToPriority(level string) int {
	switch level {
	case "critical":
		return 1
	case "high":
		return 2
	case "medium":
		return 3
	case "low":
		return 4
	default:
		return 5
	}
}
