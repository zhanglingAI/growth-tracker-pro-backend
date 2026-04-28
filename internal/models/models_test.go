package models

import (
	"testing"
	"time"
)

func TestUserTableName(t *testing.T) {
	u := User{}
	if u.TableName() != "users" {
		t.Errorf("Expected table name 'users', got '%s'", u.TableName())
	}
}

func TestChildTableName(t *testing.T) {
	c := Child{}
	if c.TableName() != "children" {
		t.Errorf("Expected table name 'children', got '%s'", c.TableName())
	}
}

func TestRecordTableName(t *testing.T) {
	r := Record{}
	if r.TableName() != "growth_records" {
		t.Errorf("Expected table name 'growth_records', got '%s'", r.TableName())
	}
}

func TestSubscriptionTableName(t *testing.T) {
	s := Subscription{}
	if s.TableName() != "subscriptions" {
		t.Errorf("Expected table name 'subscriptions', got '%s'", s.TableName())
	}
}

func TestFamilyTableName(t *testing.T) {
	f := Family{}
	if f.TableName() != "families" {
		t.Errorf("Expected table name 'families', got '%s'", f.TableName())
	}
}

func TestFamilyMemberTableName(t *testing.T) {
	fm := FamilyMember{}
	if fm.TableName() != "family_members" {
		t.Errorf("Expected table name 'family_members', got '%s'", fm.TableName())
	}
}

func TestChildCalculateAge(t *testing.T) {
	birthday := time.Date(2022, 1, 15, 0, 0, 0, 0, time.UTC)
	child := &Child{Birthday: birthday}

	tests := []struct {
		name     string
		checkAt  time.Time
		wantY    int
		wantM    int
	}{
		{
			name:    "刚出生",
			checkAt: time.Date(2022, 1, 15, 0, 0, 0, 0, time.UTC),
			wantY:   0,
			wantM:   0,
		},
		{
			name:    "1岁",
			checkAt: time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
			wantY:   1,
			wantM:   0,
		},
		{
			name:    "超过12个月",
			checkAt: time.Date(2025, 4, 20, 0, 0, 0, 0, time.UTC),
			wantY:   4,
			wantM:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			years, months := child.CalculateAge(tt.checkAt)
			if years != tt.wantY {
				t.Errorf("Years = %d, want %d", years, tt.wantY)
			}
			if months != tt.wantM {
				t.Errorf("Months = %d, want %d", months, tt.wantM)
			}
		})
	}
}

func TestSubscriptionIsActive(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		sub      *Subscription
		expected bool
	}{
		{
			name: "有效订阅",
			sub: &Subscription{
				Status:  "active",
				EndDate: now.AddDate(0, 1, 0),
			},
			expected: true,
		},
		{
			name: "今天过期",
			sub: &Subscription{
				Status:  "active",
				EndDate: now,
			},
			expected: false,
		},
		{
			name: "已过期",
			sub: &Subscription{
				Status:  "active",
				EndDate: now.AddDate(0, -1, 0),
			},
			expected: false,
		},
		{
			name: "已取消",
			sub: &Subscription{
				Status:  "cancelled",
				EndDate: now.AddDate(0, 1, 0),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sub.IsActive(); got != tt.expected {
				t.Errorf("IsActive() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSubscriptionGetRemainingQuota(t *testing.T) {
	tests := []struct {
		name     string
		sub      *Subscription
		expected int
	}{
		{
			name: "无限额度",
			sub: &Subscription{
				AIQuota: 0,
				AIUsed:  100,
			},
			expected: -1,
		},
		{
			name: "有剩余",
			sub: &Subscription{
				AIQuota: 30,
				AIUsed:  10,
			},
			expected: 20,
		},
		{
			name: "刚好用完",
			sub: &Subscription{
				AIQuota: 30,
				AIUsed:  30,
			},
			expected: 0,
		},
		{
			name: "超额使用",
			sub: &Subscription{
				AIQuota: 30,
				AIUsed:  35,
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sub.GetRemainingQuota(); got != tt.expected {
				t.Errorf("GetRemainingQuota() = %v, want %v", got, tt.expected)
			}
		})
	}
}
