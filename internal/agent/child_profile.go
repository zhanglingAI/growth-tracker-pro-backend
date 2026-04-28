package agent

import (
	"fmt"
	"time"

	"github.com/growth-tracker-pro-backend/internal/models"
)

// ChildProfile 宝宝立体信息表 - 动态构建的宝宝全景画像
type ChildProfile struct {
	// 基础信息
	BasicInfo BasicInfo `json:"basic_info"`

	// 发育评估
	GrowthAssessment GrowthAssessment `json:"growth_assessment"`

	// 营养状况
	NutritionStatus NutritionStatus `json:"nutrition_status"`

	// 生活方式
	LifestyleFactors LifestyleFactors `json:"lifestyle_factors"`

	// 健康风险
	HealthRisks []HealthRisk `json:"health_risks"`

	// 生长趋势
	GrowthTrend GrowthTrend `json:"growth_trend"`

	// 干预建议优先级
	PriorityScores PriorityScores `json:"priority_scores"`

	// 生成时间
	GeneratedAt time.Time `json:"generated_at"`
}

// BasicInfo 基础信息
type BasicInfo struct {
	Name         string    `json:"name"`
	Gender       string    `json:"gender"`
	Birthday     time.Time `json:"birthday"`
	AgeInDays    int       `json:"age_in_days"`
	AgeStr       string    `json:"age_str"`
	GenderLabel  string    `json:"gender_label"`
	FatherHeight float64   `json:"father_height"`
	MotherHeight float64   `json:"mother_height"`
}

// GrowthAssessment 发育评估
type GrowthAssessment struct {
	TargetHeight         models.TargetHeightInfo `json:"target_height"`
	CurrentPercentile    int                     `json:"current_percentile"`
	PercentileStatus     string                  `json:"percentile_status"` // normal/attention/warning
	GrowthStatus         string                  `json:"growth_status"`     // optimal/good/normal/slow
	LatestRecord         *models.Record          `json:"latest_record,omitempty"`
	RecordsCount         int                     `json:"records_count"`
	MeasurementFrequency string                  `json:"measurement_frequency"` // 测量频率评估
}

// NutritionStatus 营养状况评估
type NutritionStatus struct {
	Score            int      `json:"score"` // 0-100
	Level            string   `json:"level"` // excellent/good/average/poor
	Strengths        []string `json:"strengths"`
	Concerns         []string `json:"concerns"`
	RecommendedFoods []string `json:"recommended_foods"`
	FoodsToLimit     []string `json:"foods_to_limit"`
}

// LifestyleFactors 生活方式因素
type LifestyleFactors struct {
	ExerciseStatus ExerciseStatus `json:"exercise_status"`
	SleepStatus    SleepStatus   `json:"sleep_status"`
	SunlightStatus SunlightStatus `json:"sunlight_status"`
	OverallScore   int           `json:"overall_score"` // 0-100
}

// ExerciseStatus 运动状态
type ExerciseStatus struct {
	Score           int      `json:"score"` // 0-100
	Level           string   `json:"level"` // excellent/good/average/poor/none
	Frequency       string   `json:"frequency"`
	RecommendedTypes []string `json:"recommended_types"`
}

// SleepStatus 睡眠状态
type SleepStatus struct {
	Score            int     `json:"score"` // 0-100
	Level            string  `json:"level"`
	RecommendedHours float64 `json:"recommended_hours"`
	ActualHours      float64 `json:"actual_hours"`
	SleepQuality     string  `json:"sleep_quality"` // good/normal/poor
}

// SunlightStatus 日照状态
type SunlightStatus struct {
	Score    int    `json:"score"` // 0-100
	Level    string `json:"level"`
	Duration string `json:"recommended_duration"`
}

// HealthRisk 健康风险
type HealthRisk struct {
	Type        string `json:"type"`        // risk type
	Level       string `json:"level"`       // low/medium/high/critical
	Indicator   string `json:"indicator"`   // related indicator
	Trend       string `json:"trend"`       // improving/stable/worsening
	Description string `json:"description"`
	Action      string `json:"action"`      // recommended action
}

// GrowthTrend 生长趋势
type GrowthTrend struct {
	Velocity         float64 `json:"velocity"`          // 年生长速度 cm/year
	VelocityStatus   string  `json:"velocity_status"`  // optimal/normal/slow
	TrendDirection   string  `json:"trend_direction"`  // accelerating/stable/decelerating
	ComparedToTarget bool    `json:"compared_to_target"` // 是否在靶身高通道内
}

// PriorityScores 干预建议优先级评分
type PriorityScores struct {
	Nutrition int `json:"nutrition"`  // 营养优先级
	Exercise  int `json:"exercise"`   // 运动优先级
	Sleep     int `json:"sleep"`      // 睡眠优先级
	Lifestyle int `json:"lifestyle"`  // 整体生活方式
	Medical   int `json:"medical"`    // 医学检查优先级
}

// LabReportSummary 化验单摘要
type LabReportSummary struct {
	ReportType     string    `json:"report_type"`
	Date           time.Time `json:"date"`
	KeyFindings    []string  `json:"key_findings"`
	OverallStatus  string    `json:"overall_status"` // normal/abnormal/critical
}

// ProfileBuilder 信息表构建器
type ProfileBuilder struct{}

// NewProfileBuilder 创建信息表构建器
func NewProfileBuilder() *ProfileBuilder {
	return &ProfileBuilder{}
}

// Build 构建宝宝立体信息表
func (b *ProfileBuilder) Build(
	child *models.Child,
	records []models.Record,
	reports []models.LabReport,
) *ChildProfile {
	profile := &ChildProfile{
		GeneratedAt: time.Now(),
	}

	// 构建基础信息
	profile.BasicInfo = b.buildBasicInfo(child)

	// 构建发育评估
	profile.GrowthAssessment = b.buildGrowthAssessment(child, records)

	// 构建营养状况 (基于现有数据评估)
	profile.NutritionStatus = b.buildNutritionStatus(child, records)

	// 构建生活方式因素
	profile.LifestyleFactors = b.buildLifestyleFactors(child)

	// 评估健康风险
	profile.HealthRisks = b.assessHealthRisks(child, records, reports)

	// 计算生长趋势
	profile.GrowthTrend = b.calculateGrowthTrend(records)

	// 计算优先级评分
	profile.PriorityScores = b.calculatePriorityScores(profile)

	return profile
}

func (b *ProfileBuilder) buildBasicInfo(child *models.Child) BasicInfo {
	genderLabel := "未知"
	if child.Gender == "male" {
		genderLabel = "男孩"
	} else if child.Gender == "female" {
		genderLabel = "女孩"
	}

	now := time.Now()
	ageInDays := int(now.Sub(child.Birthday).Hours() / 24)
	years := ageInDays / 365
	months := (ageInDays % 365) / 30

	return BasicInfo{
		Name:         child.Nickname,
		Gender:       child.Gender,
		Birthday:     child.Birthday,
		AgeInDays:    ageInDays,
		AgeStr:       formatAge(years, months),
		GenderLabel:  genderLabel,
		FatherHeight: child.FatherHeight,
		MotherHeight: child.MotherHeight,
	}
}

func (b *ProfileBuilder) buildGrowthAssessment(child *models.Child, records []models.Record) GrowthAssessment {
	assessment := GrowthAssessment{
		RecordsCount: len(records),
	}

	// 计算靶身高
	targetHeight := b.calculateTargetHeight(child)
	assessment.TargetHeight = targetHeight

	// 计算最新百分位
	if len(records) > 0 {
		latest := records[len(records)-1]
		assessment.LatestRecord = &latest
		assessment.CurrentPercentile = b.calculatePercentile(&latest, child)
		assessment.PercentileStatus = b.determinePercentileStatus(assessment.CurrentPercentile, targetHeight, latest.Height)
		assessment.GrowthStatus = b.determineGrowthStatus(assessment.CurrentPercentile, latest.Height, targetHeight)
	}

	// 评估测量频率
	assessment.MeasurementFrequency = b.assessMeasurementFrequency(records)

	return assessment
}

func (b *ProfileBuilder) buildNutritionStatus(child *models.Child, records []models.Record) NutritionStatus {
	status := NutritionStatus{
		Score:     70, // 默认中等
		Level:     "average",
		Strengths: []string{},
		Concerns:  []string{},
	}

	// 基于年龄和发育状态评估
	ageInDays := int(time.Now().Sub(child.Birthday).Hours() / 24)
	years := ageInDays / 365

	// 营养建议基于年龄
	if years < 3 {
		status.RecommendedFoods = []string{
			"母乳或配方奶",
			"高铁米粉",
			"蔬菜泥、果泥",
			"肉泥、鱼肉",
		}
	} else if years < 6 {
		status.RecommendedFoods = []string{
			"每天300-500ml奶",
			"1-2个鸡蛋",
			"适量肉类、鱼类",
			"多样化的蔬菜水果",
		}
	} else {
		status.RecommendedFoods = []string{
			"每天300ml牛奶或等量奶制品",
			"1-2个鸡蛋",
			"50-100g瘦肉或鱼类",
			"充足的蔬菜水果",
			"适量的全谷物",
		}
	}

	status.FoodsToLimit = []string{
		"糖果和甜饮料",
		"油炸食品",
		"过咸的食物",
		"碳酸饮料",
	}

	// 基于体重评估营养状态
	if len(records) > 0 {
		latest := records[len(records)-1]
		if latest.Weight != nil {
			bmi := *latest.Weight / ((latest.Height / 100) * (latest.Height / 100))
			if bmi < 18.5 {
				status.Concerns = append(status.Concerns, "可能存在体重偏低或营养不足")
				status.Score -= 15
			} else if bmi > 24 {
				status.Concerns = append(status.Concerns, "注意控制体重增长")
				status.Score -= 10
			} else {
				status.Strengths = append(status.Strengths, "体重在正常范围内")
			}
		}
	}

	// 调整评分等级
	if status.Score >= 85 {
		status.Level = "excellent"
	} else if status.Score >= 70 {
		status.Level = "good"
	} else if status.Score >= 50 {
		status.Level = "average"
	} else {
		status.Level = "poor"
	}

	return status
}

func (b *ProfileBuilder) buildLifestyleFactors(child *models.Child) LifestyleFactors {
	factors := LifestyleFactors{}

	ageInDays := int(time.Now().Sub(child.Birthday).Hours() / 24)
	years := ageInDays / 365

	// 运动状态
	exercise := ExerciseStatus{
		Score: 60,
		Level: "average",
	}
	if years < 3 {
		exercise.RecommendedTypes = []string{"户外活动", "攀爬", "跑跳游戏"}
	} else {
		exercise.RecommendedTypes = []string{"跳绳", "篮球", "游泳", "跑步"}
	}
	factors.ExerciseStatus = exercise

	// 睡眠状态
	sleep := SleepStatus{
		Score: 70,
		Level: "good",
	}
	if years < 2 {
		sleep.RecommendedHours = 12
	} else if years < 6 {
		sleep.RecommendedHours = 11
	} else if years < 13 {
		sleep.RecommendedHours = 10
	} else {
		sleep.RecommendedHours = 9
	}
	sleep.ActualHours = sleep.RecommendedHours // 假设正常
	factors.SleepStatus = sleep

	// 日照状态
	sunlight := SunlightStatus{
		Score:    70,
		Level:    "average",
		Duration: "建议每天2小时户外活动",
	}
	factors.SunlightStatus = sunlight

	// 综合评分
	factors.OverallScore = (exercise.Score + sleep.Score + sunlight.Score) / 3

	return factors
}

func (b *ProfileBuilder) assessHealthRisks(child *models.Child, records []models.Record, reports []models.LabReport) []HealthRisk {
	risks := []HealthRisk{}

	// 基于生长数据分析风险
	if len(records) >= 2 {
		latest := records[len(records)-1]
		prev := records[len(records)-2]

		// 生长速度评估
		daysDiff := int(latest.MeasureDate.Sub(prev.MeasureDate).Hours() / 24)
		if daysDiff > 0 {
			heightDiff := latest.Height - prev.Height
			annualVelocity := (heightDiff / float64(daysDiff)) * 365

			if child.Gender == "male" {
				if childAge := float64(child.AgeInDays()) / 365; childAge >= 4 && childAge < 10 && annualVelocity < 4.5 {
					risks = append(risks, HealthRisk{
						Type:        "growth_velocity_low",
						Level:       "medium",
						Indicator:   "年生长速度",
						Trend:       "需要关注",
						Description: "年生长速度低于正常范围",
						Action:      "建议增加营养摄入和运动，并咨询医生",
					})
				}
			} else {
				if childAge := float64(child.AgeInDays()) / 365; childAge >= 4 && childAge < 8 && annualVelocity < 4.5 {
					risks = append(risks, HealthRisk{
						Type:        "growth_velocity_low",
						Level:       "medium",
						Indicator:   "年生长速度",
						Trend:       "需要关注",
						Description: "年生长速度低于正常范围",
						Action:      "建议增加营养摄入和运动，并咨询医生",
					})
				}
			}
		}
	}

	// 基于百分位评估风险
	if len(records) > 0 {
		latest := records[len(records)-1]
		percentile := b.calculatePercentile(&latest, child)

		if percentile < 3 {
			risks = append(risks, HealthRisk{
				Type:        "short_stature",
				Level:       "high",
				Indicator:   "身高百分位",
				Trend:       "持续偏低",
				Description: "身高明显低于同龄人正常范围",
				Action:      "建议尽早就医，进行全面检查",
			})
		} else if percentile < 10 {
			risks = append(risks, HealthRisk{
				Type:        "short_stature_risk",
				Level:       "medium",
				Indicator:   "身高百分位",
				Trend:       "偏低",
				Description: "身高处于偏低的百分位",
				Action:      "建议密切关注，必要时咨询医生",
			})
		}
	}

	// 基于化验单评估风险
	_ = reports // 预留，后续解析AIResult

	return risks
}

func (b *ProfileBuilder) calculateGrowthTrend(records []models.Record) GrowthTrend {
	trend := GrowthTrend{
		Velocity:       5.0, // 默认正常速度
		VelocityStatus: "normal",
		TrendDirection: "stable",
	}

	if len(records) < 2 {
		return trend
	}

	// 计算年生长速度
	latest := records[len(records)-1]
	oldest := records[0]

	daysDiff := int(latest.MeasureDate.Sub(oldest.MeasureDate).Hours() / 24)
	if daysDiff > 30 {
		heightDiff := latest.Height - oldest.Height
		trend.Velocity = (heightDiff / float64(daysDiff)) * 365
		trend.Velocity = float64(int(trend.Velocity*10)) / 10 // 保留一位小数
	}

	// 评估速度状态
	if trend.Velocity >= 6 {
		trend.VelocityStatus = "optimal"
	} else if trend.Velocity >= 5 {
		trend.VelocityStatus = "normal"
	} else if trend.Velocity >= 4 {
		trend.VelocityStatus = "slow"
	} else {
		trend.VelocityStatus = "very_slow"
	}

	// 评估趋势方向 (如果有3个以上数据点)
	if len(records) >= 3 {
		firstToMid := records[len(records)/2].Height - records[0].Height
		midToLast := records[len(records)-1].Height - records[len(records)/2].Height

		if midToLast > firstToMid {
			trend.TrendDirection = "accelerating"
		} else if midToLast < firstToMid {
			trend.TrendDirection = "decelerating"
		}
	}

	return trend
}

func (b *ProfileBuilder) calculatePriorityScores(profile *ChildProfile) PriorityScores {
	scores := PriorityScores{}

	// 营养优先级
	scores.Nutrition = 50 + (100-profile.NutritionStatus.Score)/2
	if len(profile.HealthRisks) > 0 {
		for _, risk := range profile.HealthRisks {
			if risk.Type == "short_stature" || risk.Type == "growth_velocity_low" {
				scores.Nutrition += 20
			}
		}
	}

	// 运动优先级
	scores.Exercise = 50 + (100-profile.LifestyleFactors.ExerciseStatus.Score)/2

	// 睡眠优先级
	scores.Sleep = 50 + (100-profile.LifestyleFactors.SleepStatus.Score)/2

	// 生活方式综合优先级
	scores.Lifestyle = (scores.Nutrition + scores.Exercise + scores.Sleep) / 3

	// 医学检查优先级
	scores.Medical = 30
	if profile.GrowthAssessment.PercentileStatus == "warning" {
		scores.Medical += 40
	}
	if profile.GrowthTrend.VelocityStatus == "slow" || profile.GrowthTrend.VelocityStatus == "very_slow" {
		scores.Medical += 30
	}
	for _, risk := range profile.HealthRisks {
		if risk.Level == "high" || risk.Level == "critical" {
			scores.Medical += 20
		}
	}

	// 确保不超过100
	if scores.Nutrition > 100 {
		scores.Nutrition = 100
	}
	if scores.Exercise > 100 {
		scores.Exercise = 100
	}
	if scores.Sleep > 100 {
		scores.Sleep = 100
	}
	if scores.Lifestyle > 100 {
		scores.Lifestyle = 100
	}
	if scores.Medical > 100 {
		scores.Medical = 100
	}

	return scores
}

// ========== 辅助计算方法 ==========

func (b *ProfileBuilder) calculateTargetHeight(child *models.Child) models.TargetHeightInfo {
	var targetHeight float64
	if child.Gender == "male" {
		targetHeight = (child.FatherHeight + child.MotherHeight + 13) / 2
	} else {
		targetHeight = (child.FatherHeight + child.MotherHeight - 13) / 2
	}

	return models.TargetHeightInfo{
		TargetHeight: float64(int(targetHeight*10)) / 10,
		MinHeight:    float64(int((targetHeight-8)*10)) / 10,
		MaxHeight:    float64(int((targetHeight+8)*10)) / 10,
	}
}

func (b *ProfileBuilder) calculatePercentile(record *models.Record, child *models.Child) int {
	height := record.Height
	ageInDays := int(time.Since(child.Birthday).Hours() / 24)
	ageInMonths := float64(ageInDays) / 30.44

	var median50 float64
	if child.Gender == "male" {
		median50 = 76 + ageInMonths*0.65
	} else {
		median50 = 75 + ageInMonths*0.62
	}

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

func (b *ProfileBuilder) determinePercentileStatus(percentile int, target models.TargetHeightInfo, currentHeight float64) string {
	_ = target
	_ = currentHeight
	if percentile >= 3 && percentile <= 97 {
		return "normal"
	} else if percentile < 3 {
		return "warning"
	}
	return "attention"
}

func (b *ProfileBuilder) determineGrowthStatus(percentile int, currentHeight float64, target models.TargetHeightInfo) string {
	if percentile >= 15 && percentile <= 85 {
		return "normal"
	}
	if currentHeight < target.MinHeight {
		return "slow"
	}
	return "attention"
}

func (b *ProfileBuilder) assessMeasurementFrequency(records []models.Record) string {
	if len(records) < 2 {
		return "建议定期测量"
	}

	// 计算平均测量间隔
	totalDays := 0
	for i := 1; i < len(records); i++ {
		totalDays += int(records[i].MeasureDate.Sub(records[i-1].MeasureDate).Hours() / 24)
	}
	avgDays := totalDays / (len(records) - 1)

	if avgDays <= 30 {
		return "测量频率良好"
	} else if avgDays <= 90 {
		return "测量频率适中"
	}
	return "建议增加测量频率"
}

func formatAge(years, months int) string {
	if years == 0 {
		return fmt.Sprintf("%d个月", months)
	}
	return fmt.Sprintf("%d岁%d个月", years, months)
}
