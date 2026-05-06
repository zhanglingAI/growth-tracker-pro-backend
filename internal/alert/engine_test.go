package alert

import (
	"testing"
	"time"

	"github.com/growth-tracker-pro-backend/internal/models"
)

// ==================== 辅助构造函数 ====================

func strPtr(s string) *string { return &s }

// newChild 构造测试用的 Child
// ageYears: 当前年龄（岁），birthday 自动计算为 ageYears 年前
func newChild(ageYears int, gender string, fatherH, motherH float64, region *string) *models.Child {
	birthday := time.Now().AddDate(-ageYears, 0, 0)
	return &models.Child{
		BaseModel:    models.BaseModel{ID: "child-test-001"},
		Gender:       gender,
		Birthday:     birthday,
		FatherHeight: fatherH,
		MotherHeight: motherH,
		Region:       region,
	}
}

// newChildWithStage 构造带生长阶段的 Child
func newChildWithStage(ageYears int, gender string, fatherH, motherH float64, region *string, stage string, lastChangeDaysAgo int) *models.Child {
	child := newChild(ageYears, gender, fatherH, motherH, region)
	child.GrowthStage = strPtr(stage)
	if lastChangeDaysAgo >= 0 {
		t := time.Now().AddDate(0, 0, -lastChangeDaysAgo)
		child.LastHeightChangeDate = &t
	}
	return child
}

// newRecord 构造单条记录
func newRecord(height float64, ageMonths int, boneAge *float64, boneAgeDiff *float64) *models.GrowthRecord {
	recordDate := time.Now().AddDate(0, -ageMonths, 0)
	return &models.GrowthRecord{
		BaseModel:   models.BaseModel{ID: "rec-test-001"},
		ChildID:     "child-test-001",
		MeasureDate: recordDate,
		Height:      height,
		BoneAge:     boneAge,
		BoneAgeDiff: boneAgeDiff,
	}
}

// newRecords 构造多条记录（按时间倒序，最近的在前面）
func newRecords(heights []float64, intervals []int) []models.GrowthRecord {
	records := make([]models.GrowthRecord, len(heights))
	now := time.Now()
	for i := range heights {
		daysAgo := 0
		for j := i; j < len(intervals) && j < len(heights)-1; j++ {
			daysAgo += intervals[j]
		}
		records[len(heights)-1-i] = models.GrowthRecord{
			BaseModel:   models.BaseModel{ID: "rec-test-" + string(rune('0'+i))},
			ChildID:     "child-test-001",
			MeasureDate: now.AddDate(0, 0, -daysAgo),
			Height:      heights[i],
		}
	}
	return records
}

// newEngine 创建测试用的引擎（Evaluate 不依赖 db）
func newEngine() *Engine {
	return &Engine{db: nil}
}

// ==================== 测试：维度1 靶身高差距 ====================

func TestCheckTargetGap(t *testing.T) {
	e := newEngine()

	tests := []struct {
		name        string
		child       *models.Child
		latestH     float64
		records     []models.GrowthRecord
		wantType    string
		wantLevel   string
		wantAlert   bool
	}{
		// 靶身高 = (170+160+13)/2 = 171.5, 范围 163.5-179.5
		{
			name:      "靶身高9%_应触发danger",
			child:     newChild(16, "male", 170, 160, nil),
			latestH:   165.0, // 165在范围中占9.4%
			wantType:  models.AlertTargetGapLow,
			wantLevel: "danger",
			wantAlert: true,
		},
		{
			name:      "靶身高28%_应触发warning",
			child:     newChild(16, "male", 170, 160, nil),
			latestH:   168.0, // 168在范围中占28.1%
			wantType:  models.AlertTargetGapLow,
			wantLevel: "warning",
			wantAlert: true,
		},
		{
			name:      "靶身高16%_年龄大于8岁_应触发danger",
			child:     newChild(16, "male", 170, 160, nil),
			latestH:   166.0, // 166在范围中占15.6%, age>=96月且<20%→danger
			wantType:  models.AlertTargetGapLow,
			wantLevel: "danger",
			wantAlert: true,
		},
		{
			name:      "靶身高41%_无下降趋势_无预警",
			child:     newChild(16, "male", 170, 160, nil),
			latestH:   170.0, // 170在范围中占40.6%
			records:   newRecords([]float64{167.0, 168.0, 170.0}, []int{180, 180}),
			wantAlert: false,
		},
		{
			name:      "靶身高41%_有下降趋势_应触发info",
			child:     newChild(16, "male", 170, 160, nil),
			latestH:   170.0, // 170在范围中占40.6%
			records:   newRecords([]float64{173.0, 169.0, 170.0}, []int{180, 180}),
			wantType:  models.AlertTargetGapLow,
			wantLevel: "info",
			wantAlert: true,
		},
		{
			name:      "靶身高53%_无预警",
			child:     newChild(16, "male", 170, 160, nil),
			latestH:   172.0, // 172在范围中占53.1%
			wantAlert: false,
		},
		{
			name:      "无最新记录_无预警",
			child:     newChild(16, "male", 170, 160, nil),
			latestH:   0,
			wantAlert: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var latest *models.GrowthRecord
			if tt.latestH > 0 {
				years, months := tt.child.CalculateAge(time.Now())
				ageM := years*12 + months
				latest = newRecord(tt.latestH, ageM, nil, nil)
			}
			input := &Input{
				Child:        tt.child,
				LatestRecord: latest,
				AllRecords:   tt.records,
				TargetHeight: models.CalculateTargetHeight(tt.child.FatherHeight, tt.child.MotherHeight, tt.child.Gender),
			}
			if latest != nil {
				years, months := tt.child.CalculateAge(time.Now())
				ageM := years*12 + months
				input.CurrentPct = models.CalculateHeightPercentile(latest.Height, ageM, tt.child.Gender)
			}

			alerts := e.checkTargetGap(input)
			if !tt.wantAlert {
				if len(alerts) > 0 {
					t.Fatalf("expected no alert, got %d: %v", len(alerts), alerts[0].Title)
				}
				return
			}
			if len(alerts) == 0 {
				t.Fatalf("expected alert, got none")
			}
			if alerts[0].AlertType != tt.wantType {
				t.Errorf("alert type = %s, want %s", alerts[0].AlertType, tt.wantType)
			}
			if alerts[0].AlertLevel != tt.wantLevel {
				t.Errorf("alert level = %s, want %s", alerts[0].AlertLevel, tt.wantLevel)
			}
		})
	}
}

// ==================== 测试：维度2 区域修正偏差 ====================

func TestCheckRegionalDeviation(t *testing.T) {
	e := newEngine()

	tests := []struct {
		name        string
		region      string
		currentPct  int
		regionalPct int
		wantLevel   string
		wantAlert   bool
	}{
		{
			name:        "上海_区域P2_全国差12_应warning",
			region:      "shanghai",
			currentPct:  14,
			regionalPct: 2,
			wantLevel:   "warning",
			wantAlert:   true,
		},
		{
			name:        "上海_区域P8_全国差8_应info",
			region:      "shanghai",
			currentPct:  15,
			regionalPct: 8,
			wantLevel:   "info",
			wantAlert:   true,
		},
		{
			name:        "上海_区域P15_无预警",
			region:      "shanghai",
			currentPct:  20,
			regionalPct: 15,
			wantAlert:   false,
		},
		{
			name:        "无region_无预警",
			region:      "",
			currentPct:  5,
			regionalPct: 2,
			wantAlert:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			child := newChild(8, "male", 170, 160, strPtr(tt.region))
			latest := newRecord(130.0, 96, nil, nil)
			input := &Input{
				Child:        child,
				LatestRecord: latest,
				CurrentPct:   tt.currentPct,
				RegionalPct:  tt.regionalPct,
				Region:       tt.region,
			}

			alerts := e.checkRegionalDeviation(input)
			if !tt.wantAlert {
				if len(alerts) > 0 {
					t.Fatalf("expected no alert, got %d", len(alerts))
				}
				return
			}
			if len(alerts) == 0 {
				t.Fatalf("expected alert, got none")
			}
			if alerts[0].AlertLevel != tt.wantLevel {
				t.Errorf("alert level = %s, want %s", alerts[0].AlertLevel, tt.wantLevel)
			}
		})
	}
}

// ==================== 测试：维度3 骨龄偏差 ====================

func TestCheckBoneAgeDeviation(t *testing.T) {
	e := newEngine()

	tests := []struct {
		name        string
		ageYears    int
		gender      string
		height      float64
		boneAge     float64
		boneAgeDiff *float64
		wantType    string
		wantLevel   string
		wantAlert   bool
	}{
		{
			name:        "骨龄提前2岁_身高不足_应danger_advanced",
			ageYears:    10,
			gender:      "male",
			height:      133.0, // 低于12岁骨龄对应的P3(135.0)
			boneAge:     12.0,
			boneAgeDiff: floatPtr(2.0),
			wantType:    models.AlertBoneAgeAdvanced,
			wantLevel:   "danger",
			wantAlert:   true,
		},
		{
			name:        "骨龄落后1.5岁_身高不足_应warning_delayed",
			ageYears:    10,
			gender:      "male",
			height:      120.0, // 低于8.5岁骨龄对应的P3(约121.35)
			boneAge:     8.5,
			boneAgeDiff: floatPtr(-1.5),
			wantType:    models.AlertBoneAgeDelayed,
			wantLevel:   "warning",
			wantAlert:   true,
		},
		{
			name:        "骨龄提前2岁_身高达标_应info",
			ageYears:    10,
			gender:      "male",
			height:      155.0, // 高于12岁骨龄P3
			boneAge:     12.0,
			boneAgeDiff: floatPtr(2.0),
			wantType:    models.AlertBoneAgeAdvanced,
			wantLevel:   "info",
			wantAlert:   true,
		},
		{
			name:        "骨龄差0.5岁_无预警",
			ageYears:    10,
			gender:      "male",
			height:      140.0,
			boneAge:     10.5,
			boneAgeDiff: floatPtr(0.5),
			wantAlert:   false,
		},
		{
			name:      "无骨龄数据_无预警",
			ageYears:  10,
			gender:    "male",
			height:    140.0,
			boneAge:   0,
			wantAlert: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			child := newChild(tt.ageYears, tt.gender, 170, 160, nil)
			var latest *models.GrowthRecord
			if tt.boneAge > 0 {
				latest = newRecord(tt.height, tt.ageYears*12, &tt.boneAge, tt.boneAgeDiff)
			} else {
				latest = newRecord(tt.height, tt.ageYears*12, nil, nil)
			}
			input := &Input{
				Child:        child,
				LatestRecord: latest,
			}

			alerts := e.checkBoneAgeDeviation(input)
			if !tt.wantAlert {
				if len(alerts) > 0 {
					t.Fatalf("expected no alert, got %d", len(alerts))
				}
				return
			}
			if len(alerts) == 0 {
				t.Fatalf("expected alert, got none")
			}
			if alerts[0].AlertType != tt.wantType {
				t.Errorf("alert type = %s, want %s", alerts[0].AlertType, tt.wantType)
			}
			if alerts[0].AlertLevel != tt.wantLevel {
				t.Errorf("alert level = %s, want %s", alerts[0].AlertLevel, tt.wantLevel)
			}
		})
	}
}

// ==================== 测试：维度4 猛涨期停滞 ====================

func TestCheckStagnation(t *testing.T) {
	e := newEngine()

	tests := []struct {
		name              string
		stage             string
		lastChangeDaysAgo int
		wantLevel         string
		wantAlert         bool
	}{
		{
			name:              "青春期_8周无变化_应danger",
			stage:             "puberty",
			lastChangeDaysAgo: 60,
			wantLevel:         "danger",
			wantAlert:         true,
		},
		{
			name:              "青春期_4周无变化_应warning",
			stage:             "puberty",
			lastChangeDaysAgo: 30,
			wantLevel:         "warning",
			wantAlert:         true,
		},
		{
			name:              "青春期_2周无变化_无预警",
			stage:             "puberty",
			lastChangeDaysAgo: 14,
			wantAlert:         false,
		},
		{
			name:      "非青春期_无预警",
			stage:     "pre_puberty",
			wantAlert: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			child := newChildWithStage(12, "male", 170, 160, nil, tt.stage, tt.lastChangeDaysAgo)
			latest := newRecord(160.0, 144, nil, nil)
			input := &Input{
				Child:        child,
				LatestRecord: latest,
			}

			alerts := e.checkStagnation(input)
			if !tt.wantAlert {
				if len(alerts) > 0 {
					t.Fatalf("expected no alert, got %d", len(alerts))
				}
				return
			}
			if len(alerts) == 0 {
				t.Fatalf("expected alert, got none")
			}
			if alerts[0].AlertLevel != tt.wantLevel {
				t.Errorf("alert level = %s, want %s", alerts[0].AlertLevel, tt.wantLevel)
			}
		})
	}
}

// ==================== 测试：维度5 生长速度过慢 ====================

func TestCheckVelocitySlow(t *testing.T) {
	e := newEngine()

	tests := []struct {
		name      string
		ageYears  int
		gender    string
		records   []models.GrowthRecord
		wantAlert bool
	}{
		{
			name:     "5岁男孩_间隔60天长高0.5cm_年增速3cm_应warning",
			ageYears: 5,
			gender:   "male",
			records: newRecords(
				[]float64{108.0, 108.5}, // 最近两次：60天前108.0，现在108.5
				[]int{60},
			),
			wantAlert: true,
		},
		{
			name:     "5岁男孩_间隔60天长高2cm_年增速12cm_正常",
			ageYears: 5,
			gender:   "male",
			records: newRecords(
				[]float64{108.0, 110.0},
				[]int{60},
			),
			wantAlert: false,
		},
		{
			name:      "1条记录_无法计算_无预警",
			ageYears:  5,
			gender:    "male",
			records:   newRecords([]float64{110.0}, []int{}),
			wantAlert: false,
		},
		{
			name:     "间隔15天_无预警",
			ageYears: 5,
			gender:   "male",
			records: newRecords(
				[]float64{108.0, 108.2},
				[]int{15},
			),
			wantAlert: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			child := newChild(tt.ageYears, tt.gender, 170, 160, nil)
			var latest *models.GrowthRecord
			if len(tt.records) > 0 {
				latest = &tt.records[len(tt.records)-1]
			}
			input := &Input{
				Child:        child,
				LatestRecord: latest,
				AllRecords:   tt.records,
			}

			alerts := e.checkVelocitySlow(input)
			if !tt.wantAlert {
				if len(alerts) > 0 {
					t.Fatalf("expected no alert, got %d", len(alerts))
				}
				return
			}
			if len(alerts) == 0 {
				t.Fatalf("expected alert, got none")
			}
		})
	}
}

// ==================== 测试：维度6 百分位持续下降 ====================

func TestCheckPercentileDrop(t *testing.T) {
	e := newEngine()

	// 构造记录：早期高，后期低（百分位持续下降）
	// 男孩，5条记录，间隔6个月
	// 年龄从 8岁->8.5岁->9岁->9.5岁->10岁
	// 身高选择使早期百分位约P80，后期约P60
	dropRecords := func() []models.GrowthRecord {
		// 8岁男孩 P80 ≈ 130.8, P60 ≈ 128.0
		// 10岁男孩 P80 ≈ 142.1, P60 ≈ 139.0
		return newRecords(
			[]float64{131.0, 132.0, 135.0, 138.0, 139.0}, // 早期高后期低（相对同龄）
			[]int{180, 180, 180, 180},
		)
	}

	stableRecords := func() []models.GrowthRecord {
		return newRecords(
			[]float64{122.0, 126.0, 130.0, 134.0, 138.0}, // 正常增长，百分位稳定
			[]int{180, 180, 180, 180},
		)
	}

	tests := []struct {
		name      string
		ageYears  int
		records   []models.GrowthRecord
		wantAlert bool
	}{
		{
			name:      "5条记录_早期P80后期P60_应info",
			ageYears:  8,
			records:   dropRecords(),
			wantAlert: true,
		},
		{
			name:      "5条记录_百分位稳定_无预警",
			ageYears:  8,
			records:   stableRecords(),
			wantAlert: false,
		},
		{
			name:      "2条记录_不足3条_无预警",
			ageYears:  8,
			records:   newRecords([]float64{130.0, 132.0}, []int{180}),
			wantAlert: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			child := newChild(tt.ageYears, "male", 170, 160, nil)
			var latest *models.GrowthRecord
			if len(tt.records) > 0 {
				latest = &tt.records[len(tt.records)-1]
			}
			input := &Input{
				Child:        child,
				LatestRecord: latest,
				AllRecords:   tt.records,
			}

			alerts := e.checkPercentileDrop(input)
			if !tt.wantAlert {
				if len(alerts) > 0 {
					t.Fatalf("expected no alert, got %d", len(alerts))
				}
				return
			}
			if len(alerts) == 0 {
				t.Fatalf("expected alert, got none")
			}
		})
	}
}

// ==================== 测试：去重逻辑 ====================

func TestDeduplicate(t *testing.T) {
	e := newEngine()

	alerts := []*models.HeightAlert{
		{AlertType: models.AlertTargetGapLow, AlertLevel: "info", Dimension: "target_gap"},
		{AlertType: models.AlertTargetGapLow, AlertLevel: "warning", Dimension: "target_gap"},
		{AlertType: models.AlertRegionalShort, AlertLevel: "info", Dimension: "regional"},
	}

	result := e.deduplicate(alerts)
	if len(result) != 2 {
		t.Fatalf("expected 2 alerts after dedup, got %d", len(result))
	}

	// 验证同一类型保留了最高级别
	for _, a := range result {
		if a.AlertType == models.AlertTargetGapLow && a.AlertLevel != "warning" {
			t.Errorf("expected warning for target_gap, got %s", a.AlertLevel)
		}
	}
}

// ==================== 测试：用户给的真实案例 ====================

func TestEvaluate_RealCase_ShanghaiBoy12Y(t *testing.T) {
	e := newEngine()

	// 用户数据：上海男孩，12岁，161.5cm，父170，母168
	child := newChild(12, "male", 170, 168, strPtr("shanghai"))
	latest := newRecord(161.5, 144, nil, nil)

	years, months := child.CalculateAge(time.Now())
	ageM := years*12 + months

	currentPct := models.CalculateHeightPercentile(latest.Height, ageM, child.Gender)
	regionalPct := models.CalculateRegionalPercentile(latest.Height, ageM, child.Gender, *child.Region)

	input := &Input{
		Child:        child,
		LatestRecord: latest,
		AllRecords:   []models.GrowthRecord{*latest},
		TargetHeight: models.CalculateTargetHeight(child.FatherHeight, child.MotherHeight, child.Gender),
		CurrentPct:   currentPct,
		RegionalPct:  regionalPct,
		Region:       *child.Region,
	}

	alerts := e.Evaluate(input)

	// 应该触发 1 条 danger 级别靶身高预警
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d: %+v", len(alerts), alerts)
	}

	alert := alerts[0]
	if alert.AlertType != models.AlertTargetGapLow {
		t.Errorf("alert type = %s, want %s", alert.AlertType, models.AlertTargetGapLow)
	}
	if alert.AlertLevel != "danger" {
		t.Errorf("alert level = %s, want danger", alert.AlertLevel)
	}
	if alert.MetricValue == nil || *alert.MetricValue != 0 {
		t.Errorf("metric value = %v, want 0", alert.MetricValue)
	}

	t.Logf("Real case result: type=%s, level=%s, title=%s", alert.AlertType, alert.AlertLevel, alert.Title)
}
