package agent

import (
	"strings"
	"testing"
	"time"

	"github.com/growth-tracker-pro-backend/internal/models"
)

func TestProfileBuilder_Build(t *testing.T) {
	builder := NewProfileBuilder()

	child := &models.Child{
		Name:         "小明",
		Gender:       "male",
		Birthday:     time.Now().AddDate(-5, 0, 0), // 5岁
		FatherHeight: 175,
		MotherHeight: 160,
	}

	records := []models.Record{
		{
			Height:    100,
			Weight:    16,
			Date:      time.Now().AddDate(-1, 0, 0),
			AgeInDays: 1460,
		},
		{
			Height:    105,
			Weight:    17,
			Date:      time.Now(),
			AgeInDays: 1825,
		},
	}

	profile := builder.Build(child, records, nil)

	// 验证基础信息
	if profile.BasicInfo.Name != "小明" {
		t.Errorf("Expected name '小明', got '%s'", profile.BasicInfo.Name)
	}
	if profile.BasicInfo.GenderLabel != "男孩" {
		t.Errorf("Expected gender '男孩', got '%s'", profile.BasicInfo.GenderLabel)
	}

	// 验证发育评估
	if profile.GrowthAssessment.TargetHeight.TargetHeight == 0 {
		t.Error("Target height should be calculated")
	}

	// 验证生长趋势
	if profile.GrowthTrend.Velocity <= 0 {
		t.Error("Growth velocity should be positive")
	}
}

func TestProfileBuilder_CalculatePercentile(t *testing.T) {
	builder := NewProfileBuilder()

	child := &models.Child{
		Gender: "male",
	}

	// 测试正常百分位
	record := &models.Record{
		Height:    110,
		AgeInDays: 1825, // 约5岁
	}

	percentile := builder.calculatePercentile(record, child)
	if percentile <= 0 || percentile > 100 {
		t.Errorf("Percentile should be between 1-100, got %d", percentile)
	}
}

func TestMedicalGuard_CheckResponse(t *testing.T) {
	guard := NewMedicalGuard()

	tests := []struct {
		name     string
		response string
		hasAlert bool
	}{
		{
			name:     "正常回复",
			response: "宝宝发育正常，建议继续保持均衡营养",
			hasAlert: false,
		},
		{
			name:     "包含禁止词汇",
			response: "建议确诊后使用生长激素治疗",
			hasAlert: true,
		},
		{
			name:     "包含医疗告警词",
			response: "可能存在矮小症，建议就医",
			hasAlert: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, isOK := guard.CheckResponse(tt.response)
			// isOK=false means has alert, isOK=true means no alert
			hasAlert := !isOK
			if hasAlert != tt.hasAlert {
				t.Errorf("Expected alert=%v, got %v", tt.hasAlert, hasAlert)
			}
		})
	}
}

func TestMedicalGuard_CheckAndSanitize(t *testing.T) {
	guard := NewMedicalGuard()

	// 测试替换禁止词汇
	response := "建议确诊后使用生长激素治疗"
	cleaned, alert := guard.CheckAndSanitize(response)

	if alert == nil {
		t.Error("Expected medical alert")
	}

	// 验证医生提示被追加
	if !strings.Contains(cleaned, "医生") {
		t.Error("Doctor consultation notice should be added")
	}
}

func TestMedicalGuard_IsQuestionable(t *testing.T) {
	guard := NewMedicalGuard()

	tests := []struct {
		question   string
		questionable bool
	}{
		{"孩子是不是发育有问题", true},
		{"需要打生长激素吗", true},
		{"要不要去医院看看", true},
		{"吃什么能长高", false},
		{"运动建议有哪些", false},
	}

	for _, tt := range tests {
		t.Run(tt.question, func(t *testing.T) {
			result := guard.IsQuestionable(tt.question)
			if result != tt.questionable {
				t.Errorf("Expected questionable=%v for '%s'", tt.questionable, tt.question)
			}
		})
	}
}

func TestRecommendationEngine_GenerateRecommendations(t *testing.T) {
	profile := &ChildProfile{
		NutritionStatus: NutritionStatus{
			Score: 60,
			Level: "average",
		},
		LifestyleFactors: LifestyleFactors{
			ExerciseStatus: ExerciseStatus{
				Score: 50,
				Level: "poor",
			},
			SleepStatus: SleepStatus{
				Score: 70,
				Level: "good",
			},
		},
		PriorityScores: PriorityScores{
			Nutrition: 70,
			Exercise:  80,
			Sleep:     50,
			Medical:   30,
		},
		HealthRisks: []HealthRisk{},
	}

	child := &models.Child{
		Name:         "测试宝宝",
		Gender:       "male",
		Birthday:     time.Now().AddDate(-6, 0, 0),
		FatherHeight: 170,
		MotherHeight: 160,
	}

	engine := NewRecommendationEngine(profile, child)
	recommendations := engine.GenerateRecommendations()

	if len(recommendations) == 0 {
		t.Error("Should generate at least one recommendation")
	}

	// 验证优先级排序
	for i := 1; i < len(recommendations); i++ {
		if recommendations[i].Priority < recommendations[i-1].Priority {
			t.Error("Recommendations should be sorted by priority")
		}
	}
}

func TestRecommendationEngine_GenerateDailyPlan(t *testing.T) {
	profile := &ChildProfile{}
	child := &models.Child{
		Name:         "测试宝宝",
		Gender:       "male",
		Birthday:     time.Now().AddDate(-4, 0, 0), // 4岁
		FatherHeight: 170,
		MotherHeight: 160,
	}

	engine := NewRecommendationEngine(profile, child)
	plan := engine.GenerateDailyPlan()

	if plan == nil {
		t.Fatal("Daily plan should not be nil")
	}

	if len(plan.Morning) == 0 {
		t.Error("Morning plan should not be empty")
	}

	if len(plan.Afternoon) == 0 {
		t.Error("Afternoon plan should not be empty")
	}

	if len(plan.Evening) == 0 {
		t.Error("Evening plan should not be empty")
	}
}

func TestRecommendationEngine_GenerateWeeklyPlan(t *testing.T) {
	profile := &ChildProfile{}
	child := &models.Child{
		Name:         "测试宝宝",
		Gender:       "female",
		Birthday:     time.Now().AddDate(-8, 0, 0), // 8岁
		FatherHeight: 175,
		MotherHeight: 165,
	}

	engine := NewRecommendationEngine(profile, child)
	plan := engine.GenerateWeeklyPlan()

	if plan == nil {
		t.Fatal("Weekly plan should not be nil")
	}

	if len(plan.ExercisePlan) == 0 {
		t.Error("Exercise plan should not be empty")
	}

	if len(plan.NutritionPlan) == 0 {
		t.Error("Nutrition plan should not be empty")
	}
}

func TestRecommendationEngine_GenerateSummaryReport(t *testing.T) {
	profile := &ChildProfile{
		BasicInfo: BasicInfo{
			Name:   "测试宝宝",
			AgeStr: "5岁",
		},
		GrowthAssessment: GrowthAssessment{
			CurrentPercentile: 50,
			TargetHeight: models.TargetHeightInfo{
				TargetHeight: 170,
				MinHeight:    162,
				MaxHeight:    178,
			},
		},
		GrowthTrend: GrowthTrend{
			Velocity:       5.5,
			VelocityStatus: "normal",
		},
		NutritionStatus: NutritionStatus{
			Score: 70,
			Level: "good",
		},
		LifestyleFactors: LifestyleFactors{
			ExerciseStatus: ExerciseStatus{
				Score: 60,
			},
			SleepStatus: SleepStatus{
				RecommendedHours: 10,
			},
		},
	}
	child := &models.Child{Name: "测试宝宝"}

	engine := NewRecommendationEngine(profile, child)
	report := engine.GenerateSummaryReport()

	if report == "" {
		t.Error("Summary report should not be empty")
	}

	if !strings.Contains(report, "测试宝宝") {
		t.Error("Report should contain child name")
	}

	if !strings.Contains(report, "生长发育") {
		t.Error("Report should contain key content")
	}
}

func TestHealthRisk_LevelClassification(t *testing.T) {
	risks := []HealthRisk{
		{Type: "test", Level: "critical", Description: "Critical risk"},
		{Type: "test", Level: "high", Description: "High risk"},
		{Type: "test", Level: "medium", Description: "Medium risk"},
		{Type: "test", Level: "low", Description: "Low risk"},
	}

	for _, risk := range risks {
		priority := riskToPriority(risk.Level)
		if priority <= 0 || priority > 5 {
			t.Errorf("Invalid priority %d for level %s", priority, risk.Level)
		}
	}
}

func TestFormatAge(t *testing.T) {
	tests := []struct {
		years   int
		months  int
		expected string
	}{
		{0, 6, "6个月"},
		{1, 3, "1岁3个月"},
		{5, 0, "5岁0个月"},
	}

	for _, tt := range tests {
		result := formatAge(tt.years, tt.months)
		if result != tt.expected {
			t.Errorf("formatAge(%d, %d) = %s, expected %s", tt.years, tt.months, result, tt.expected)
		}
	}
}
