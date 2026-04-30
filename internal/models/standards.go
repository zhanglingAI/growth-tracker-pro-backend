package models

import "math"

// GrowthStandard 生长标准数据 (cm)
// 数据来源: 国家卫生健康委员会2025年3月1日最新发布
//         《中国0-18岁儿童青少年生长标准》
type GrowthStandard struct {
	AgeMonths int     // 月龄
	Sex       int     // 0=女, 1=男
	P3        float64 // 第3百分位 - 最低正常值
	P10       float64 // 第10百分位
	P25       float64 // 第25百分位
	P50       float64 // 第50百分位 - 中位数
	P75       float64 // 第75百分位
	P90       float64 // 第90百分位
	P97       float64 // 第97百分位 - 最高正常值
}

// ChinaHeightStandards 中国0-18岁儿童青少年身高标准 (cm)
// 数据来源: 国家卫健委2025年3月最新发布标准
var ChinaHeightStandards = []GrowthStandard{
	// ===== 男童 0-216月龄 =====
	{0, 1, 47.1, 48.3, 49.7, 50.4, 52.1, 53.2, 54.3},    // 出生
	{1, 1, 50.8, 52.1, 53.6, 54.8, 56.3, 57.5, 58.7},    // 1月
	{3, 1, 57.3, 58.7, 60.2, 61.6, 63.2, 64.5, 65.8},    // 3月
	{6, 1, 63.5, 65.1, 66.6, 67.8, 69.3, 70.6, 71.9},    // 6月
	{12, 1, 71.0, 72.7, 74.3, 76.5, 78.6, 80.5, 82.4},   // 1岁
	{24, 1, 81.3, 83.3, 85.4, 88.5, 91.2, 93.8, 96.5},   // 2岁
	{36, 1, 88.7, 91.0, 93.7, 96.8, 99.9, 102.5, 105.3}, // 3岁
	{48, 1, 95.5, 98.7, 102.1, 106.0, 109.8, 113.1, 116.5}, // 4岁
	{60, 1, 101.4, 105.3, 109.3, 113.6, 117.9, 121.6, 125.4}, // 5岁
	{72, 1, 108.1, 112.4, 116.8, 121.7, 126.8, 131.0, 135.4}, // 6岁
	{84, 1, 113.6, 118.4, 123.3, 128.6, 134.1, 139.0, 144.1}, // 7岁
	{96, 1, 118.8, 124.1, 129.6, 135.5, 141.6, 147.0, 152.6}, // 8岁
	{108, 1, 123.9, 129.6, 135.7, 142.1, 148.6, 154.5, 160.6}, // 9岁
	{120, 1, 128.0, 133.9, 140.0, 146.4, 153.0, 158.8, 165.0}, // 10岁
	{132, 1, 130.0, 134.5, 139.7, 145.3, 151.0, 156.3, 161.6}, // 11岁 ✅卫健委2025标准
	{144, 1, 135.0, 140.5, 146.6, 153.0, 159.8, 166.0, 172.5}, // 12岁
	{156, 1, 142.0, 149.5, 157.0, 164.5, 171.5, 177.5, 183.5}, // 13岁
	{168, 1, 150.0, 157.5, 164.5, 171.0, 177.0, 181.5, 186.0}, // 14岁
	{180, 1, 156.0, 163.0, 169.0, 174.0, 178.8, 182.8, 186.8}, // 15岁
	{192, 1, 159.0, 165.5, 170.8, 175.4, 179.9, 183.7, 187.5}, // 16岁
	{204, 1, 160.5, 166.8, 171.8, 176.2, 180.5, 184.2, 187.9}, // 17岁
	{216, 1, 161.0, 167.2, 172.2, 176.5, 180.8, 184.5, 188.2}, // 18岁(成人)

	// ===== 女童 0-216月龄 =====
	{0, 0, 46.6, 47.8, 49.0, 49.7, 51.2, 52.2, 53.2},    // 出生
	{1, 0, 49.8, 51.2, 52.7, 53.7, 55.0, 56.2, 57.4},    // 1月
	{3, 0, 56.0, 57.4, 58.8, 60.1, 61.6, 62.8, 64.1},    // 3月
	{6, 0, 62.2, 63.7, 65.1, 66.4, 67.9, 69.1, 70.3},    // 6月
	{12, 0, 69.0, 70.7, 72.3, 75.0, 77.6, 79.9, 82.2},   // 1岁
	{24, 0, 80.5, 82.5, 84.8, 87.2, 89.9, 92.4, 94.9},   // 2岁
	{36, 0, 87.4, 89.8, 92.5, 95.6, 98.8, 101.6, 104.5}, // 3岁
	{48, 0, 94.6, 97.8, 101.2, 104.9, 108.7, 112.1, 115.6}, // 4岁
	{60, 0, 100.4, 104.0, 107.7, 111.7, 115.8, 119.5, 123.3}, // 5岁
	{72, 0, 106.5, 111.0, 115.6, 120.6, 125.7, 130.3, 135.1}, // 6岁
	{84, 0, 111.8, 116.8, 122.0, 127.5, 133.1, 138.2, 143.4}, // 7岁
	{96, 0, 117.0, 122.5, 128.3, 134.5, 140.8, 146.6, 152.5}, // 8岁
	{108, 0, 122.5, 128.5, 134.8, 141.5, 148.3, 154.5, 160.8}, // 9岁
	{120, 0, 127.5, 133.8, 140.5, 147.5, 154.6, 161.0, 167.5}, // 10岁
	{132, 0, 131.5, 138.1, 145.0, 152.5, 160.0, 166.5, 173.1}, // 11岁
	{144, 0, 138.5, 144.9, 151.4, 157.5, 163.7, 169.1, 174.6}, // 12岁
	{156, 0, 145.0, 150.8, 156.5, 161.8, 167.2, 171.9, 176.7}, // 13岁
	{168, 0, 149.0, 154.2, 159.3, 164.0, 168.7, 172.8, 177.0}, // 14岁
	{180, 0, 150.8, 155.7, 160.5, 165.0, 169.4, 173.3, 177.3}, // 15岁
	{192, 0, 151.2, 156.0, 160.8, 165.2, 169.6, 173.5, 177.5}, // 16岁
	{204, 0, 151.3, 156.1, 160.9, 165.3, 169.7, 173.6, 177.6}, // 17岁
	{216, 0, 151.4, 156.2, 161.0, 165.4, 169.8, 173.7, 177.7}, // 18岁(成人)
}

// 标准类型常量
const (
	StandardCN  = "cn"  // 中国标准 (卫健委2025)
	StandardWHO = "who" // WHO标准
)

// GetGrowthStandard 获取指定年龄和性别的生长标准数据
func GetGrowthStandard(ageInMonths int, gender string) *GrowthStandard {
	sex := 0
	if gender == "male" {
		sex = 1
	}

	// 范围限制: 0-18岁
	if ageInMonths < 0 {
		ageInMonths = 0
	}
	if ageInMonths > 216 {
		ageInMonths = 216
	}

	// 查找精确匹配
	for i := range ChinaHeightStandards {
		if ChinaHeightStandards[i].AgeMonths == ageInMonths && ChinaHeightStandards[i].Sex == sex {
			return &ChinaHeightStandards[i]
		}
	}

	// 线性插值: 查找前后两个点
	var lower, upper *GrowthStandard
	for i := range ChinaHeightStandards {
		data := &ChinaHeightStandards[i]
		if data.Sex != sex {
			continue
		}
		if data.AgeMonths < ageInMonths {
			lower = data
		} else if data.AgeMonths > ageInMonths && upper == nil {
			upper = data
		}
	}

	if lower != nil && upper != nil {
		ratio := float64(ageInMonths-lower.AgeMonths) / float64(upper.AgeMonths-lower.AgeMonths)
		return &GrowthStandard{
			AgeMonths: ageInMonths,
			Sex:       sex,
			P3:  interpolate(lower.P3, upper.P3, ratio),
			P10: interpolate(lower.P10, upper.P10, ratio),
			P25: interpolate(lower.P25, upper.P25, ratio),
			P50: interpolate(lower.P50, upper.P50, ratio),
			P75: interpolate(lower.P75, upper.P75, ratio),
			P90: interpolate(lower.P90, upper.P90, ratio),
			P97: interpolate(lower.P97, upper.P97, ratio),
		}
	}

	// 返回最接近的数据
	if lower != nil {
		return lower
	}
	return upper
}

// 线性插值
func interpolate(a, b, ratio float64) float64 {
	return a + (b-a)*ratio
}

// CalculateHeightPercentile 计算身高百分位
// 返回值: 1-99 的百分位数值
func CalculateHeightPercentile(height float64, ageInMonths int, gender string) int {
	std := GetGrowthStandard(ageInMonths, gender)
	if std == nil {
		return 50
	}

	// 基于各百分位区间计算
	if height <= std.P3 {
		// P3以下: 1-3
		p3Height := std.P3
		belowP3 := p3Height * 0.92 // 假设P1约为P3的92%
		if height <= belowP3 {
			return 1
		}
		ratio := (height - belowP3) / (p3Height - belowP3)
		return int(1 + ratio*2)
	} else if height <= std.P10 {
		ratio := (height - std.P3) / (std.P10 - std.P3)
		return int(3 + ratio*7)
	} else if height <= std.P25 {
		ratio := (height - std.P10) / (std.P25 - std.P10)
		return int(10 + ratio*15)
	} else if height <= std.P50 {
		ratio := (height - std.P25) / (std.P50 - std.P25)
		return int(25 + ratio*25)
	} else if height <= std.P75 {
		ratio := (height - std.P50) / (std.P75 - std.P50)
		return int(50 + ratio*25)
	} else if height <= std.P90 {
		ratio := (height - std.P75) / (std.P90 - std.P75)
		return int(75 + ratio*15)
	} else if height <= std.P97 {
		ratio := (height - std.P90) / (std.P97 - std.P90)
		return int(90 + ratio*7)
	} else {
		// P97以上: 97-99
		aboveP97 := std.P97 * 1.07 // 假设P99约为P97的107%
		if height >= aboveP97 {
			return 99
		}
		ratio := (height - std.P97) / (aboveP97 - std.P97)
		return int(97 + ratio*2)
	}
}

// GetHeightPercentileStatus 获取身高百分位状态描述
// 按照卫健委标准: <P3=偏矮, P3-P97=正常, >P97=超高
func GetHeightPercentileStatus(percentile int) string {
	switch {
	case percentile < 3:
		return "身高偏矮"
	case percentile < 10:
		return "身高偏下"
	case percentile < 25:
		return "身高中下"
	case percentile < 75:
		return "身高中等"
	case percentile < 90:
		return "身高中上"
	case percentile < 97:
		return "身高偏高"
	default:
		return "身高超高"
	}
}

// IsHeightNormal 判断身高是否在正常范围内 (P3-P97)
func IsHeightNormal(height float64, ageInMonths int, gender string) bool {
	percentile := CalculateHeightPercentile(height, ageInMonths, gender)
	return percentile >= 3 && percentile <= 97
}

// GetHeightPercentileLevel 获取身高水平等级
func GetHeightPercentileLevel(percentile int) string {
	switch {
	case percentile >= 97:
		return "excellent"
	case percentile >= 75:
		return "good"
	case percentile >= 25:
		return "normal"
	case percentile >= 10:
		return "below"
	default:
		return "warning"
	}
}

// CalculateZScore 计算身高Z评分 (标准差评分)
func CalculateZScore(height float64, ageInMonths int, gender string) float64 {
	std := GetGrowthStandard(ageInMonths, gender)
	if std == nil {
		return 0
	}

	// 使用四分位距估算标准差
	sd := (std.P75 - std.P25) / 0.6745
	if sd <= 0 {
		sd = 5.0
	}

	return (height - std.P50) / sd
}

// Round 保留n位小数
func Round(x float64, n int) float64 {
	mult := math.Pow10(n)
	return math.Round(x*mult) / mult
}
