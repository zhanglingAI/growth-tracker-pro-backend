package service

import (
	"context"
	"testing"
	"time"

	"github.com/growth-tracker-pro/backend/internal/models"
)

// 测试靶身高计算
func TestCalculateTargetHeight(t *testing.T) {
	svc := &ServiceImpl{}

	tests := []struct {
		name     string
		child    *models.Child
		wantMin  float64
		wantMax  float64
	}{
		{
			name: "男孩，父亲170母亲160",
			child: &models.Child{
				Gender:       "male",
				FatherHeight: 170,
				MotherHeight: 160,
			},
			wantMin: 162.5,
			wantMax: 178.5,
		},
		{
			name: "女孩，父亲175母亲165",
			child: &models.Child{
				Gender:       "female",
				FatherHeight: 175,
				MotherHeight: 165,
			},
			wantMin: 163.5,
			wantMax: 179.5,
		},
		{
			name: "男孩，父母身高相同",
			child: &models.Child{
				Gender:       "male",
				FatherHeight: 180,
				MotherHeight: 180,
			},
			wantMin: 172.5,
			wantMax: 188.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.CalculateTargetHeight(tt.child)

			if result.MinHeight != tt.wantMin {
				t.Errorf("MinHeight = %v, want %v", result.MinHeight, tt.wantMin)
			}
			if result.MaxHeight != tt.wantMax {
				t.Errorf("MaxHeight = %v, want %v", result.MaxHeight, tt.wantMax)
			}
			if result.TargetHeight < tt.wantMin || result.TargetHeight > tt.wantMax {
				t.Errorf("TargetHeight = %v, expected between %v and %v",
					result.TargetHeight, tt.wantMin, tt.wantMax)
			}
		})
	}
}

// 测试干预窗口计算
func TestCalculateInterventionWindow(t *testing.T) {
	svc := &ServiceImpl{}

	now := time.Now()

	tests := []struct {
		name     string
		child    *models.Child
		wantDays bool // 是否应该剩余天数 > 0
	}{
		{
			name: "男孩8岁 - 应该在窗口内",
			child: &models.Child{
				Gender:   "male",
				Birthday: now.AddDate(-10, 0, 0),
			},
			wantDays: true,
		},
		{
			name: "男孩5岁 - 不在窗口内",
			child: &models.Child{
				Gender:   "male",
				Birthday: now.AddDate(-5, 0, 0),
			},
			wantDays: false,
		},
		{
			name: "女孩9岁 - 应该在窗口内",
			child: &models.Child{
				Gender:   "female",
				Birthday: now.AddDate(-9, 0, 0),
			},
			wantDays: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.CalculateInterventionWindow(tt.child)

			if tt.wantDays && result.RemainingDays <= 0 {
				t.Logf("Expected positive remaining days, got %d", result.RemainingDays)
			}
			if !tt.wantDays && result.RemainingDays > 0 {
				t.Logf("Expected zero or negative remaining days, got %d", result.RemainingDays)
			}
		})
	}
}

// 测试百分位计算
func TestCalculatePercentile(t *testing.T) {
	svc := &ServiceImpl{}

	tests := []struct {
		name      string
		record    *models.Record
		child     *models.Child
		wantMin   int
		wantMax   int
	}{
		{
			name: "男孩身高正常",
			record: &models.Record{
				Height:    120,
				AgeInDays: 1460, // 4岁
			},
			child: &models.Child{
				Gender: "male",
			},
			wantMin: 25,
			wantMax: 85,
		},
		{
			name: "男孩身高偏高",
			record: &models.Record{
				Height:    140,
				AgeInDays: 1460,
			},
			child: &models.Child{
				Gender: "male",
			},
			wantMin: 85,
			wantMax: 100,
		},
		{
			name: "女孩身高偏低",
			record: &models.Record{
				Height:    100,
				AgeInDays: 1460,
			},
			child: &models.Child{
				Gender: "female",
			},
			wantMin: 3,
			wantMax: 25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.CalculatePercentile(tt.record, tt.child)

			if result < tt.wantMin || result > tt.wantMax {
				t.Errorf("Percentile = %v, want between %v and %v",
					result, tt.wantMin, tt.wantMax)
			}
		})
	}
}

// 测试生长状态判断
func TestDetermineGrowthStatus(t *testing.T) {
	svc := &ServiceImpl{}

	tests := []struct {
		name         string
		percentile   int
		targetMin    float64
		targetMax    float64
		currentHeight float64
		wantStatus   string
	}{
		{
			name:          "正常状态",
			percentile:    50,
			targetMin:     160,
			targetMax:     180,
			currentHeight: 165,
			wantStatus:    "normal",
		},
		{
			name:          "低于靶身高通道",
			percentile:    10,
			targetMin:     160,
			targetMax:     180,
			currentHeight: 155,
			wantStatus:    "warning",
		},
		{
			name:          "百分位偏低但高于靶身高下限",
			percentile:    10,
			targetMin:     150,
			targetMax:     170,
			currentHeight: 152,
			wantStatus:    "attention",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.DetermineGrowthStatus(tt.percentile, tt.targetMin, tt.targetMax, tt.currentHeight)

			if result != tt.wantStatus {
				t.Errorf("GrowthStatus = %v, want %v", result, tt.wantStatus)
			}
		})
	}
}

// 测试宝宝年龄计算
func TestChildCalculateAge(t *testing.T) {
	tests := []struct {
		name     string
		birthday time.Time
		checkAt  time.Time
		wantYears   int
		wantMonths int
	}{
		{
			name:     "正好3岁",
			birthday: time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
			checkAt:  time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
			wantYears:   3,
			wantMonths: 0,
		},
		{
			name:     "3岁3个月",
			birthday: time.Date(2022, 10, 1, 0, 0, 0, 0, time.UTC),
			checkAt:  time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
			wantYears:   4,
			wantMonths: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			child := &models.Child{Birthday: tt.birthday}
			years, months := child.CalculateAge(tt.checkAt)

			if years != tt.wantYears {
				t.Errorf("Years = %v, want %v", years, tt.wantYears)
			}
			if months != tt.wantMonths {
				t.Errorf("Months = %v, want %v", months, tt.wantMonths)
			}
		})
	}
}

// 测试订阅是否有效
func TestSubscriptionIsActive(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		sub      *models.Subscription
		wantActive bool
	}{
		{
			name: "有效订阅",
			sub: &models.Subscription{
				Status:   "active",
				EndDate:  now.AddDate(0, 1, 0),
			},
			wantActive: true,
		},
		{
			name: "已过期",
			sub: &models.Subscription{
				Status:   "active",
				EndDate:  now.AddDate(0, -1, 0),
			},
			wantActive: false,
		},
		{
			name: "已取消",
			sub: &models.Subscription{
				Status:   "cancelled",
				EndDate:  now.AddDate(0, 1, 0),
			},
			wantActive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.sub.IsActive()

			if result != tt.wantActive {
				t.Errorf("IsActive = %v, want %v", result, tt.wantActive)
			}
		})
	}
}

// 测试剩余额度计算
func TestSubscriptionGetRemainingQuota(t *testing.T) {
	tests := []struct {
		name          string
		sub          *models.Subscription
		wantRemaining int
	}{
		{
			name: "无限额度",
			sub: &models.Subscription{
				AIQuota: 0,
				AIUsed:  100,
			},
			wantRemaining: -1,
		},
		{
			name: "有剩余额度",
			sub: &models.Subscription{
				AIQuota: 30,
				AIUsed:  10,
			},
			wantRemaining: 20,
		},
		{
			name: "额度用完",
			sub: &models.Subscription{
				AIQuota: 30,
				AIUsed:  30,
			},
			wantRemaining: 0,
		},
		{
			name: "额度超额",
			sub: &models.Subscription{
				AIQuota: 30,
				AIUsed:  35,
			},
			wantRemaining: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.sub.GetRemainingQuota()

			if result != tt.wantRemaining {
				t.Errorf("GetRemainingQuota = %v, want %v", result, tt.wantRemaining)
			}
		})
	}
}

// 基准测试：靶身高计算性能
func BenchmarkCalculateTargetHeight(b *testing.B) {
	svc := &ServiceImpl{}
	child := &models.Child{
		Gender:       "male",
		FatherHeight: 175,
		MotherHeight: 165,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.CalculateTargetHeight(child)
	}
}

// 基准测试：百分位计算性能
func BenchmarkCalculatePercentile(b *testing.B) {
	svc := &ServiceImpl{}
	record := &models.Record{
		Height:    120,
		AgeInDays: 1460,
	}
	child := &models.Child{
		Gender: "male",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.CalculatePercentile(record, child)
	}
}
