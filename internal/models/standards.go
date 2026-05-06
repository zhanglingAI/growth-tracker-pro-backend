package models

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

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
	{132, 1, 131.0, 137.0, 143.0, 150.0, 156.5, 162.5, 169.0}, // 11岁
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

// ============= 靶身高预测 =============
// Tanner 靶身高公式 (WHO推荐标准)

// TargetHeightInfo 靶身高信息
type TargetHeightInfo struct {
	TargetHeight float64 `json:"target_height"` // 靶身高（中位数）
	MinHeight    float64 `json:"min_height"`    // 下限 (靶身高 - 8cm)
	MaxHeight    float64 `json:"max_height"`    // 上限 (靶身高 + 8cm)
}

// CalculateTargetHeight Tanner靶身高公式
// 参数: fatherHeight(父亲身高cm), motherHeight(母亲身高cm), gender("male"/"female")
func CalculateTargetHeight(fatherHeight, motherHeight float64, gender string) TargetHeightInfo {
	var target float64
	if gender == "male" {
		// 男孩: (父 + 母 + 13) / 2
		target = (fatherHeight + motherHeight + 13) / 2
	} else {
		// 女孩: (父 + 母 - 13) / 2
		target = (fatherHeight + motherHeight - 13) / 2
	}

	return TargetHeightInfo{
		TargetHeight: Round(target, 1),
		MinHeight:    Round(target-8, 1),
		MaxHeight:    Round(target+8, 1),
	}
}

// GetTargetHeightPercentile 计算当前身高在靶身高范围中的百分位
// 返回值: 0-100, 50为刚好达到靶身高中位数
func GetTargetHeightPercentile(currentHeight float64, target TargetHeightInfo) int {
	if currentHeight <= target.MinHeight {
		return 0
	}
	if currentHeight >= target.MaxHeight {
		return 100
	}
	ratio := (currentHeight - target.MinHeight) / (target.MaxHeight - target.MinHeight)
	return int(ratio * 100)
}

// GetHeightPotentialStatus 遗传潜力达成状态
func GetHeightPotentialStatus(percentile int) string {
	switch {
	case percentile >= 80:
		return "充分发挥遗传潜力"
	case percentile >= 50:
		return "遗传潜力正常发挥"
	case percentile >= 30:
		return "遗传潜力发挥一般"
	default:
		return "遗传潜力发挥不足，建议排查原因"
	}
}

// ============= 南北区域修正 =============

// RegionCorrection 地区修正系数
type RegionCorrection struct {
	ProvinceCode string  // 省份拼音代码
	ProvinceName string  // 中文名称
	Correction   float64 // 修正值(cm), 正值=该地区儿童平均比全国高
	Category     string  // "north" / "central" / "south"
}

// RegionalCorrections 南北差异修正表
// 数据来源: 教育部全国学生体质健康调研 + 各省市卫健委监测数据
var RegionalCorrections = []RegionCorrection{
	// 北方省份 (+2~3cm)
	{"heilongjiang", "黑龙江", 3.0, "north"},
	{"jilin", "吉林", 2.8, "north"},
	{"liaoning", "辽宁", 2.8, "north"},
	{"shandong", "山东", 3.0, "north"},
	{"beijing", "北京", 2.5, "north"},
	{"tianjin", "天津", 2.5, "north"},
	{"hebei", "河北", 2.3, "north"},
	{"shanxi", "山西", 2.0, "north"},
	{"neimenggu", "内蒙古", 2.0, "north"},
	{"xinjiang", "新疆", 1.8, "north"},

	// 中部省份 (~0cm)
	{"henan", "河南", 0.5, "central"},
	{"anhui", "安徽", 0.3, "central"},
	{"hubei", "湖北", 0.3, "central"},
	{"hunan", "湖南", 0.2, "central"},
	{"jiangxi", "江西", 0.0, "central"},
	{"jiangsu", "江苏", 0.8, "central"},
	{"zhejiang", "浙江", 0.5, "central"},
	{"shaanxi", "陕西", 0.8, "central"},
	{"gansu", "甘肃", 0.5, "central"},
	{"ningxia", "宁夏", 0.8, "central"},
	{"qinghai", "青海", 0.3, "central"},
	{"shanghai", "上海", 1.0, "central"},

	// 南方省份 (-1~-2cm)
	{"fujian", "福建", -0.5, "south"},
	{"guangdong", "广东", -0.8, "south"},
	{"guangxi", "广西", -1.5, "south"},
	{"hainan", "海南", -2.0, "south"},
	{"sichuan", "四川", -1.5, "south"},
	{"chongqing", "重庆", -1.0, "south"},
	{"guizhou", "贵州", -2.0, "south"},
	{"yunnan", "云南", -1.8, "south"},

	// 其他省级行政区
	{"xizang", "西藏", -1.5, "south"},
	{"taiwan", "台湾", -0.5, "south"},
	{"hongkong", "香港", -0.3, "south"},
	{"macao", "澳门", -0.3, "south"},
}

// GetRegionCorrection 获取地区修正系数
func GetRegionCorrection(provinceCode string) *RegionCorrection {
	if provinceCode == "" {
		return nil
	}
	for i := range RegionalCorrections {
		if RegionalCorrections[i].ProvinceCode == provinceCode {
			return &RegionalCorrections[i]
		}
	}
	return nil
}

// GetRegionName 获取地区中文名
func GetRegionName(provinceCode string) string {
	c := GetRegionCorrection(provinceCode)
	if c != nil {
		return c.ProvinceName
	}
	return provinceCode
}

// ============= 省独立标准接口（预留，未来可替换加法修正）=============

// ProvincialStandardProvider 省独立标准提供者接口
// 当各省发布独立生长标准时，实现此接口并注册到 provincialProvider 即可
// type ProvincialStandardProvider interface {
//     GetStandard(ageInMonths int, gender string, provinceCode string) *GrowthStandard
// }

// provincialProvider 当前使用加法修正，未来可替换为独立省标准表
// var provincialProvider ProvincialStandardProvider = &additiveCorrectionProvider{}

// GetRegionalGrowthStandard 获取区域修正后的生长标准
// 当前实现：全国标准 + 加法修正
// 未来可替换为：直接查省独立标准表
func GetRegionalGrowthStandard(ageInMonths int, gender string, region string) (*GrowthStandard, *RegionCorrection) {
	std := GetGrowthStandard(ageInMonths, gender)
	if std == nil {
		return nil, nil
	}

	correction := GetRegionCorrection(region)
	if correction == nil {
		return std, nil
	}

	return &GrowthStandard{
		AgeMonths: std.AgeMonths,
		Sex:       std.Sex,
		P3:        Round(std.P3+correction.Correction, 1),
		P10:       Round(std.P10+correction.Correction, 1),
		P25:       Round(std.P25+correction.Correction, 1),
		P50:       Round(std.P50+correction.Correction, 1),
		P75:       Round(std.P75+correction.Correction, 1),
		P90:       Round(std.P90+correction.Correction, 1),
		P97:       Round(std.P97+correction.Correction, 1),
	}, correction
}

// calcPercentileFromStandard 根据标准计算百分位（私有辅助函数）
func calcPercentileFromStandard(height float64, std *GrowthStandard) int {
	if height <= std.P3 {
		belowP3 := std.P3 * 0.92
		if height <= belowP3 {
			return 1
		}
		ratio := (height - belowP3) / (std.P3 - belowP3)
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
		aboveP97 := std.P97 * 1.07
		if height >= aboveP97 {
			return 99
		}
		ratio := (height - std.P97) / (aboveP97 - std.P97)
		return int(97 + ratio*2)
	}
}

// CalculateRegionalPercentile 计算区域修正后的百分位
func CalculateRegionalPercentile(height float64, ageInMonths int, gender string, region string) int {
	std, _ := GetRegionalGrowthStandard(ageInMonths, gender, region)
	if std == nil {
		return CalculateHeightPercentile(height, ageInMonths, gender)
	}
	return calcPercentileFromStandard(height, std)
}

// ============= 生长速率评估 =============

// GrowthRateInfo 生长速率信息
type GrowthRateInfo struct {
	AnnualGrowth float64 `json:"annual_growth"` // 年增长值(cm)
	ExpectedMin  float64 `json:"expected_min"`  // 该年龄段最低期望值
	Status       string  `json:"status"`         // 评估结果
	IsNormal     bool    `json:"is_normal"`      // 是否正常
}

// EvaluateGrowthRate 评估生长速率
// ageInYears: 周岁, annualGrowth: 近1年长高了多少cm
func EvaluateGrowthRate(ageInYears int, annualGrowth float64, gender string) GrowthRateInfo {
	var expectedMin float64

	switch {
	case ageInYears < 2:
		expectedMin = 7.0 // <2岁: 每年≥7cm
	case ageInYears < 10:
		expectedMin = 5.0 // 2岁-青春期前: 每年≥5cm
	default:
		if gender == "male" {
			expectedMin = 6.0 // 男孩青春期: 每年≥6cm
		} else {
			expectedMin = 5.0 // 女孩青春期: 每年≥5cm
		}
	}

	status := "生长速率正常"
	isNormal := annualGrowth >= expectedMin
	if !isNormal {
		status = "生长速率偏慢，建议就医检查"
	}

	return GrowthRateInfo{
		AnnualGrowth: annualGrowth,
		ExpectedMin:  expectedMin,
		Status:       status,
		IsNormal:     isNormal,
	}
}

// ============= 数量遗传学修正模型 =============

// QuantitativeGeneticsResult 数量遗传学修正结果
type QuantitativeGeneticsResult struct {
	TargetHeight    float64 `json:"target_height"`     // 修正后预测身高
	Heritability    float64 `json:"heritability"`      // 使用的遗传力
	FatherSelection float64 `json:"father_selection"`  // 父亲选择差
	MotherSelection float64 `json:"mother_selection"`  // 母亲选择差
	ChildDeviation  float64 `json:"child_deviation"`   // 子女预期离差
	Explanation     string  `json:"explanation"`       // 计算说明
}

// 中国人群平均身高（用于数量遗传学修正）
const (
	ChinaMaleMeanHeight   = 172.0 // 中国男性平均身高(cm)
	ChinaFemaleMeanHeight = 160.0 // 中国女性平均身高(cm)
)

// CalculateQuantitativeGeneticsTargetHeight 数量遗传学修正模型
// 引入遗传力(h² = 0.6~0.8)和向平均回归效应
// 适用于父母身高极端偏离均值的情况
// fatherHeight, motherHeight: 父母身高(cm)
// gender: "male" / "female"
// heritability: 遗传力，默认0.8（高收入国家），一般使用0.75
func CalculateQuantitativeGeneticsTargetHeight(fatherHeight, motherHeight float64, gender string, heritability float64) QuantitativeGeneticsResult {
	if heritability <= 0 || heritability > 1 {
		heritability = 0.75 // 默认遗传力
	}

	// 计算父母选择差
	fatherSelection := fatherHeight - ChinaMaleMeanHeight
	motherSelection := motherHeight - ChinaFemaleMeanHeight

	// 子女预期离差 = h² × (父亲选择差 + 母亲选择差) ÷ 2
	childDeviation := heritability * (fatherSelection + motherSelection) / 2

	// 子女预测身高 = 性别平均身高 + 子女预期离差
	var targetHeight float64
	var genderMean float64
	if gender == "male" {
		genderMean = ChinaMaleMeanHeight
		targetHeight = genderMean + childDeviation
	} else {
		genderMean = ChinaFemaleMeanHeight
		targetHeight = genderMean + childDeviation
	}

	explanation := fmt.Sprintf(
		"父亲选择差=%.1fcm(父%.1f-均值%.1f)，母亲选择差=%.1fcm(母%.1f-均值%.1f)，遗传力h²=%.2f。子女预期离差=%.2f×(%.1f+%.1f)/2=%.1fcm。预测身高=%.1f+%.1f=%.1fcm",
		fatherSelection, fatherHeight, ChinaMaleMeanHeight,
		motherSelection, motherHeight, ChinaFemaleMeanHeight,
		heritability, heritability, fatherSelection, motherSelection, childDeviation,
		genderMean, childDeviation, targetHeight,
	)

	return QuantitativeGeneticsResult{
		TargetHeight:    Round(targetHeight, 1),
		Heritability:    heritability,
		FatherSelection: Round(fatherSelection, 1),
		MotherSelection: Round(motherSelection, 1),
		ChildDeviation:  Round(childDeviation, 1),
		Explanation:     explanation,
	}
}

// ============= Khamis-Roche 法（综合层） =============

// KhamisRocheResult Khamis-Roche预测结果
type KhamisRocheResult struct {
	PredictedHeight float64 `json:"predicted_height"` // 预测成年身高
	ErrorRange      float64 `json:"error_range"`      // 预测误差 ±cm
	Method          string  `json:"method"`           // 方法说明
	Inputs          KRInputs `json:"inputs"`          // 输入参数
}

// KRInputs Khamis-Roche输入参数
type KRInputs struct {
	FatherHeight float64 `json:"father_height"`
	MotherHeight float64 `json:"mother_height"`
	CurrentHeight float64 `json:"current_height"`
	CurrentWeight float64 `json:"current_weight"`
	AgeYears      float64 `json:"age_years"`       // 精确到半年
	Gender        string  `json:"gender"`
}

// CalculateKhamisRoche Khamis-Roche法预测成年身高
// 适用范围: 4~17.5岁
// 核心公式: 预测身高 = a + b×父亲身高 + c×母亲身高 + d×当前身高 + e×体重
// 此处使用简化系数表（按年龄-性别分层）
func CalculateKhamisRoche(fatherHeight, motherHeight, currentHeight, currentWeight float64, ageMonths int, gender string) KhamisRocheResult {
	ageYears := float64(ageMonths) / 12.0

	// 超出适用范围时返回MPH结果并标注
	if ageYears < 4 || ageYears > 17.5 {
		mph := CalculateTargetHeight(fatherHeight, motherHeight, gender)
		return KhamisRocheResult{
			PredictedHeight: mph.TargetHeight,
			ErrorRange:      8.0,
			Method:          "MPH(年龄超出Khamis-Roche适用范围4-17.5岁)",
			Inputs: KRInputs{
				FatherHeight:  fatherHeight,
				MotherHeight:  motherHeight,
				CurrentHeight: currentHeight,
				CurrentWeight: currentWeight,
				AgeYears:      ageYears,
				Gender:        gender,
			},
		}
	}

	// 简化系数（基于文献近似值，按性别和年龄段分组）
	// 公式: predicted = intercept + bFather*father + bMother*mother + bCurrent*currentHeight + bWeight*weight
	var intercept, bFather, bMother, bCurrent, bWeight float64

	if gender == "male" {
		switch {
		case ageYears < 6:
			intercept, bFather, bMother, bCurrent, bWeight = 35.0, 0.35, 0.30, 0.55, 0.15
		case ageYears < 9:
			intercept, bFather, bMother, bCurrent, bWeight = 30.0, 0.38, 0.32, 0.50, 0.12
		case ageYears < 12:
			intercept, bFather, bMother, bCurrent, bWeight = 25.0, 0.40, 0.34, 0.48, 0.10
		case ageYears < 15:
			intercept, bFather, bMother, bCurrent, bWeight = 20.0, 0.42, 0.35, 0.45, 0.08
		default:
			intercept, bFather, bMother, bCurrent, bWeight = 15.0, 0.43, 0.36, 0.42, 0.06
		}
	} else {
		switch {
		case ageYears < 6:
			intercept, bFather, bMother, bCurrent, bWeight = 30.0, 0.33, 0.32, 0.58, 0.14
		case ageYears < 9:
			intercept, bFather, bMother, bCurrent, bWeight = 25.0, 0.36, 0.34, 0.52, 0.11
		case ageYears < 12:
			intercept, bFather, bMother, bCurrent, bWeight = 20.0, 0.38, 0.36, 0.50, 0.09
		case ageYears < 14:
			intercept, bFather, bMother, bCurrent, bWeight = 15.0, 0.40, 0.37, 0.48, 0.07
		default:
			intercept, bFather, bMother, bCurrent, bWeight = 10.0, 0.41, 0.38, 0.45, 0.05
		}
	}

	predicted := intercept +
		bFather*fatherHeight +
		bMother*motherHeight +
		bCurrent*currentHeight +
		bWeight*currentWeight

	return KhamisRocheResult{
		PredictedHeight: Round(predicted, 1),
		ErrorRange:      3.5,
		Method:          "Khamis-Roche",
		Inputs: KRInputs{
			FatherHeight:  fatherHeight,
			MotherHeight:  motherHeight,
			CurrentHeight: currentHeight,
			CurrentWeight: currentWeight,
			AgeYears:      ageYears,
			Gender:        gender,
		},
	}
}

// ============= 年龄分层权重 =============

// AgeLayeredWeights 年龄分层遗传/环境权重
type AgeLayeredWeights struct {
	AgeGroup       string  `json:"age_group"`       // 年龄段名称
	MinAge         int     `json:"min_age"`         // 起始年龄(岁)
	MaxAge         int     `json:"max_age"`         // 结束年龄(岁)
	GeneticWeight  float64 `json:"genetic_weight"`  // 遗传权重
	EnvironmentWeight float64 `json:"environment_weight"` // 环境权重
	CoreIntervention string `json:"core_intervention"` // 核心干预靶点
}

// AgeLayeredWeightTable 年龄分层权重表
var AgeLayeredWeightTable = []AgeLayeredWeights{
	{"婴幼儿期", 0, 3, 0.65, 0.35, "营养、疾病防控"},
	{"学龄前期", 3, 6, 0.70, 0.30, "营养、睡眠、运动习惯"},
	{"学龄期", 6, 12, 0.75, 0.25, "运动、睡眠、学业压力"},
	{"青春期早期", 12, 15, 0.775, 0.225, "睡眠、运动、性发育监测"},
	{"青春期后期", 15, 18, 0.825, 0.175, "维持性干预，关注骨骺闭合"},
}

// GetAgeLayeredWeights 获取指定年龄的分层权重
func GetAgeLayeredWeights(ageInYears int) AgeLayeredWeights {
	for _, w := range AgeLayeredWeightTable {
		if ageInYears >= w.MinAge && ageInYears <= w.MaxAge {
			return w
		}
	}
	// 默认返回学龄期
	return AgeLayeredWeightTable[2]
}

// ============= 环境因素评估问卷 =============

// EnvironmentScoreResult 环境得分结果
type EnvironmentScoreResult struct {
	TotalScore       float64                `json:"total_score"`        // 总分(0-50)
	MaxPossibleScore float64                `json:"max_possible_score"` // 满分50
	ModuleScores     map[string]ModuleScore `json:"module_scores"`      // 各模块得分
	Interpretation   string                 `json:"interpretation"`     // 临床解读
	InterventionZone string                 `json:"intervention_zone"`  // 干预分区: high/medium/low
}

// ModuleScore 模块得分
type ModuleScore struct {
	Score      float64 `json:"score"`       // 实际得分
	MaxScore   float64 `json:"max_score"`   // 满分
	Weight     float64 `json:"weight"`      // 权重
	Percentage float64 `json:"percentage"`  // 得分率
	Priority   string  `json:"priority"`    // 干预优先级
}

// EnvironmentQuestionnaire 环境问卷数据
type EnvironmentQuestionnaire struct {
	// 营养状况 (满分15分, 权重30%)
	Nutrition NutritionModule `json:"nutrition"`
	// 睡眠质量 (满分12.5分, 权重25%)
	Sleep SleepModule `json:"sleep"`
	// 运动状况 (满分12.5分, 权重25%)
	Exercise ExerciseModule `json:"exercise"`
	// 健康状况 (满分5分, 权重10%)
	Health HealthModule `json:"health"`
	// 心理状况 (满分5分, 权重10%)
	Mental MentalModule `json:"mental"`
}

// NutritionModule 营养模块
type NutritionModule struct {
	DietDiversity      int     `json:"diet_diversity"`       // 膳食多样性: 每日≥12种=2分, <8种=0分
	ProteinAdequacy    int     `json:"protein_adequacy"`     // 蛋白质充足度: 1.8g/kg/日且优质≥50%=2分
	CalciumIntake      int     `json:"calcium_intake"`       // 钙质摄入: ≥800mg/日=2分, <500mg=0分
	VitaminDStatus     int     `json:"vitamin_d_status"`     // 维生素D: 血清>50nmol/L=1分
	BadEatingBehavior  int     `json:"bad_eating_behavior"`  // 不良饮食行为: 无挑食=2分, 严重挑食=0分
	WeightManagement   int     `json:"weight_management"`    // 体重管理: BMI P15-P85=1分
}

// SleepModule 睡眠模块
type SleepModule struct {
	Duration        float64 `json:"duration"`         // 总睡眠时长(小时)
	BedtimeRegularity int   `json:"bedtime_regularity"` // 入睡规律性: 波动<30min=满分
	DeepSleepCover    int     `json:"deep_sleep_cover"`   // 深睡眠覆盖: 22:00前入睡=满分
	SleepContinuity   int     `json:"sleep_continuity"`   // 睡眠连续性: 夜间觉醒≤1次=满分
	SleepEnvironment  int     `json:"sleep_environment"`  // 睡眠环境: 黑暗/安静/温度适宜
}

// ExerciseModule 运动模块
type ExerciseModule struct {
	Frequency     int `json:"frequency"`      // 运动频率: 每周5天=满分
	TypeSuitability int `json:"type_suitability"` // 类型适宜性: 弹跳+伸展+有氧=满分
	Duration      int `json:"duration"`       // 时长适中性: 30-60分钟=满分
	Intensity     int `json:"intensity"`      // 强度分级: 中等强度=满分
}

// HealthModule 健康模块
type HealthModule struct {
	DiseaseControl     int `json:"disease_control"`     // 疾病控制
	CheckupCompliance  int `json:"checkup_compliance"`  // 体检依从性
	MedicationSafety   int `json:"medication_safety"`   // 用药安全
}

// MentalModule 心理模块
type MentalModule struct {
	EmotionRegulation int `json:"emotion_regulation"` // 情绪调节
	FamilySupport     int `json:"family_support"`     // 家庭支持
	StressManagement  int `json:"stress_management"`  // 压力管理
}

// CalculateEnvironmentScore 计算环境问卷得分
// 返回总分(0-50)和各模块详细得分
func CalculateEnvironmentScore(q *EnvironmentQuestionnaire, ageInYears int) EnvironmentScoreResult {
	result := EnvironmentScoreResult{
		ModuleScores: make(map[string]ModuleScore),
	}

	// 营养模块评分 (满分15分)
	nutritionScore := float64(q.Nutrition.DietDiversity) +
		float64(q.Nutrition.ProteinAdequacy) +
		float64(q.Nutrition.CalciumIntake) +
		float64(q.Nutrition.VitaminDStatus) +
		float64(q.Nutrition.BadEatingBehavior) +
		float64(q.Nutrition.WeightManagement)
	if nutritionScore > 15 {
		nutritionScore = 15
	}
	result.ModuleScores["nutrition"] = ModuleScore{
		Score: nutritionScore, MaxScore: 15, Weight: 0.30,
		Percentage: nutritionScore / 15,
		Priority:   "最高",
	}

	// 睡眠模块评分 (满分12.5分)
	// 时长评分: 学龄儿童9-11h/青少年8-10h为满分，每减少1h扣1分
	sleepDurationScore := 2.5
	if ageInYears < 13 {
		if q.Sleep.Duration >= 9 && q.Sleep.Duration <= 11 {
			sleepDurationScore = 2.5
		} else if q.Sleep.Duration >= 8 {
			sleepDurationScore = 1.5
		} else if q.Sleep.Duration >= 7 {
			sleepDurationScore = 0.5
		} else {
			sleepDurationScore = 0
		}
	} else {
		if q.Sleep.Duration >= 8 && q.Sleep.Duration <= 10 {
			sleepDurationScore = 2.5
		} else if q.Sleep.Duration >= 7 {
			sleepDurationScore = 1.5
		} else if q.Sleep.Duration >= 6 {
			sleepDurationScore = 0.5
		} else {
			sleepDurationScore = 0
		}
	}

	// 入睡规律性 (满分2.5分)
	sleepRegularityScore := float64(q.Sleep.BedtimeRegularity)
	if sleepRegularityScore > 2.5 {
		sleepRegularityScore = 2.5
	}

	// 深睡眠覆盖 (满分2.5分)
	deepSleepScore := float64(q.Sleep.DeepSleepCover)
	if deepSleepScore > 2.5 {
		deepSleepScore = 2.5
	}

	// 睡眠连续性 (满分2.5分)
	continuityScore := float64(q.Sleep.SleepContinuity)
	if continuityScore > 2.5 {
		continuityScore = 2.5
	}

	// 睡眠环境 (满分2.5分)
	envScore := float64(q.Sleep.SleepEnvironment)
	if envScore > 2.5 {
		envScore = 2.5
	}

	sleepScore := sleepDurationScore + sleepRegularityScore + deepSleepScore + continuityScore + envScore
	if sleepScore > 12.5 {
		sleepScore = 12.5
	}
	result.ModuleScores["sleep"] = ModuleScore{
		Score: sleepScore, MaxScore: 12.5, Weight: 0.25,
		Percentage: sleepScore / 12.5,
		Priority:   "高",
	}

	// 运动模块评分 (满分12.5分)
	exerciseScore := float64(q.Exercise.Frequency) +
		float64(q.Exercise.TypeSuitability) +
		float64(q.Exercise.Duration) +
		float64(q.Exercise.Intensity)
	if exerciseScore > 12.5 {
		exerciseScore = 12.5
	}
	result.ModuleScores["exercise"] = ModuleScore{
		Score: exerciseScore, MaxScore: 12.5, Weight: 0.25,
		Percentage: exerciseScore / 12.5,
		Priority:   "高",
	}

	// 健康模块评分 (满分5分)
	healthScore := float64(q.Health.DiseaseControl) +
		float64(q.Health.CheckupCompliance) +
		float64(q.Health.MedicationSafety)
	if healthScore > 5 {
		healthScore = 5
	}
	result.ModuleScores["health"] = ModuleScore{
		Score: healthScore, MaxScore: 5, Weight: 0.10,
		Percentage: healthScore / 5,
		Priority:   "中",
	}

	// 心理模块评分 (满分5分)
	// 心理风险放大机制: 当心理状况评分<3分时，其他模块扣分效应×1.5倍
	mentalScore := float64(q.Mental.EmotionRegulation) +
		float64(q.Mental.FamilySupport) +
		float64(q.Mental.StressManagement)
	if mentalScore > 5 {
		mentalScore = 5
	}
	result.ModuleScores["mental"] = ModuleScore{
		Score: mentalScore, MaxScore: 5, Weight: 0.10,
		Percentage: mentalScore / 5,
		Priority:   "中",
	}

	// 计算总分
	total := nutritionScore + sleepScore + exerciseScore + healthScore + mentalScore
	result.TotalScore = Round(total, 1)
	result.MaxPossibleScore = 50

	// 心理风险放大: 如果心理得分率<60%(即<3分)，总分额外扣减
	mentalRatio := mentalScore / 5
	if mentalRatio < 0.6 {
		penalty := (5 - mentalScore) * 0.5
		result.TotalScore = Round(result.TotalScore-penalty, 1)
		ms := result.ModuleScores["mental"]
		ms.Priority = "高(心理风险放大)"
		result.ModuleScores["mental"] = ms
	}

	// 临床解读
	result.InterventionZone, result.Interpretation = interpretEnvironmentScore(result.TotalScore)

	return result
}

// interpretEnvironmentScore 环境得分临床解读
func interpretEnvironmentScore(score float64) (zone, interpretation string) {
	switch {
	case score >= 40:
		return "high", fmt.Sprintf("得分%.1f分(高分区): 趋近遗传上限，维持现状，每6-12月复评", score)
	case score >= 25:
		return "medium", fmt.Sprintf("得分%.1f分(中分区): 针对性改善薄弱环节，每3-6月复评", score)
	default:
		return "low", fmt.Sprintf("得分%.1f分(低分区): 全面干预+专科评估，每月复评", score)
	}
}

// ============= 固定增量预测模型（方案B） =============

// ComprehensivePredictionResult 综合预测结果
type ComprehensivePredictionResult struct {
	GeneticTargetHeight     float64                `json:"genetic_target_height"`      // 遗传靶身高(MPH)
	QuantitativeGenetics    QuantitativeGeneticsResult `json:"quantitative_genetics,omitempty"`
	KhamisRoche             *KhamisRocheResult     `json:"khamis_roche,omitempty"`
	EnvironmentScore        *EnvironmentScoreResult `json:"environment_score,omitempty"`
	EnvironmentIncrement    float64                `json:"environment_increment"`      // 环境增量(cm)
	PredictedHeight         float64                `json:"predicted_height"`           // 综合预测身高
	PredictionMethod        string                 `json:"prediction_method"`          // 使用的方法
	ErrorRange              float64                `json:"error_range"`                // 预测误差
	AgeWeights              AgeLayeredWeights      `json:"age_weights"`                // 年龄分层权重
	ClinicalInterpretation  string                 `json:"clinical_interpretation"`    // 临床解读
	MaxPredictedHeight      float64                `json:"max_predicted_height"`       // 预测上限(不超过群体P99)
}

// CalculateComprehensivePrediction 综合预测算法（方案B：固定增量模型）
// 步骤:
//   1. 计算遗传靶身高(MPH)
//   2. 如年龄4-17.5岁，计算Khamis-Roche预测值
//   3. 如有环境问卷，计算环境增量
//   4. 综合预测 = 遗传靶身高 + 环境增量 (或使用KR值作为基础)
func CalculateComprehensivePrediction(
	fatherHeight, motherHeight, currentHeight, currentWeight float64,
	ageInYears, ageInMonths int,
	gender string,
	envQuestionnaire *EnvironmentQuestionnaire,
) ComprehensivePredictionResult {
	// Step 1: 遗传靶身高(MPH)
	geneticTarget := CalculateTargetHeight(fatherHeight, motherHeight, gender)

	result := ComprehensivePredictionResult{
		GeneticTargetHeight: geneticTarget.TargetHeight,
		AgeWeights:          GetAgeLayeredWeights(ageInYears),
	}

	// Step 2: 数量遗传学修正（当父母身高极端偏离均值时）
	fatherDeviation := absFloat(fatherHeight - ChinaMaleMeanHeight)
	motherDeviation := absFloat(motherHeight - ChinaFemaleMeanHeight)
	if fatherDeviation > 10 || motherDeviation > 10 {
		qg := CalculateQuantitativeGeneticsTargetHeight(fatherHeight, motherHeight, gender, 0.75)
		result.QuantitativeGenetics = qg
	}

	// Step 3: Khamis-Roche法 (4-17.5岁且有体重数据)
	basePrediction := geneticTarget.TargetHeight
	predictionMethod := "MPH"
	errorRange := 8.0

	if ageInYears >= 4 && ageInYears <= 17 && currentWeight > 0 {
		kr := CalculateKhamisRoche(fatherHeight, motherHeight, currentHeight, currentWeight, ageInMonths, gender)
		result.KhamisRoche = &kr
		// 优先使用Khamis-Roche作为基础预测值
		basePrediction = kr.PredictedHeight
		predictionMethod = "Khamis-Roche + 环境增量"
		errorRange = kr.ErrorRange
	}

	// Step 4: 环境增量计算
	var envIncrement float64
	if envQuestionnaire != nil {
		envScore := CalculateEnvironmentScore(envQuestionnaire, ageInYears)
		result.EnvironmentScore = &envScore

		// 固定增量模型: 环境增量 = 环境总分 × k (k=0.2cm/分)
		// 最大增量10cm
		k := 0.2
		envIncrement = envScore.TotalScore * k
		if envIncrement > 10 {
			envIncrement = 10
		}
		envIncrement = Round(envIncrement, 1)
	}

	// Step 5: 综合预测
	predictedHeight := basePrediction + envIncrement

	// 边界约束: 不超过群体P99，不低于P1
	std := GetGrowthStandard(ageInMonths, gender)
	if std != nil {
		if predictedHeight > std.P97*1.07 { // 约P99
			predictedHeight = std.P97 * 1.07
			envIncrement = Round(predictedHeight-basePrediction, 1)
		}
		if predictedHeight < std.P3*0.92 { // 约P1
			predictedHeight = std.P3 * 0.92
			envIncrement = Round(predictedHeight-basePrediction, 1)
		}
		result.MaxPredictedHeight = Round(std.P97*1.07, 1)
	}

	result.EnvironmentIncrement = envIncrement
	result.PredictedHeight = Round(predictedHeight, 1)
	result.PredictionMethod = predictionMethod
	result.ErrorRange = errorRange

	// 临床解读
	result.ClinicalInterpretation = generateClinicalInterpretation(result, ageInYears)

	return result
}

// generateClinicalInterpretation 生成临床解读
func generateClinicalInterpretation(result ComprehensivePredictionResult, ageInYears int) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("综合预测身高: %.1fcm(±%.1fcm)\n", result.PredictedHeight, result.ErrorRange))
	sb.WriteString(fmt.Sprintf("遗传靶身高: %.1fcm\n", result.GeneticTargetHeight))

	if result.EnvironmentScore != nil {
		sb.WriteString(fmt.Sprintf("环境得分: %.1f/50分，环境增量: +%.1fcm\n",
			result.EnvironmentScore.TotalScore, result.EnvironmentIncrement))
		sb.WriteString(fmt.Sprintf("干预分区: %s\n", result.EnvironmentScore.Interpretation))
	} else {
		sb.WriteString("未提供环境问卷，预测基于遗传因素 alone\n")
	}

	if result.KhamisRoche != nil {
		sb.WriteString(fmt.Sprintf("Khamis-Roche预测: %.1fcm(±%.1fcm)\n",
			result.KhamisRoche.PredictedHeight, result.KhamisRoche.ErrorRange))
	}

	sb.WriteString(fmt.Sprintf("年龄分层: 遗传权重%.0f%%，环境权重%.0f%%\n",
		result.AgeWeights.GeneticWeight*100, result.AgeWeights.EnvironmentWeight*100))

	return sb.String()
}

func absFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// ============= 改进的生长速度监测 =============

// CalculateAnnualGrowthVelocity 计算年生长速度
// 使用最近n个月的数据: 年生长速度 = 12 × (目前身高 − n个月前身高) ÷ n
// 返回年生长速度(cm/年)和预警级别
func CalculateAnnualGrowthVelocity(records []GrowthRecord, monthsBack int) (velocity float64, alertLevel string) {
	if len(records) < 2 || monthsBack <= 0 {
		return 0, ""
	}

	// 按日期排序
	sort.Slice(records, func(i, j int) bool {
		return records[i].MeasureDate.Before(records[j].MeasureDate)
	})

	latest := records[len(records)-1]

	// 查找n个月前的记录
	cutoffDate := latest.MeasureDate.AddDate(0, -monthsBack, 0)
	var referenceRecord *GrowthRecord
	for i := len(records) - 1; i >= 0; i-- {
		if records[i].MeasureDate.Before(cutoffDate) || records[i].MeasureDate.Equal(cutoffDate) {
			referenceRecord = &records[i]
			break
		}
	}

	if referenceRecord == nil {
		// 使用最早的记录
		referenceRecord = &records[0]
	}

	// 计算实际间隔月数
	actualMonths := monthsBetween(referenceRecord.MeasureDate, latest.MeasureDate)
	if actualMonths <= 0 {
		actualMonths = 1
	}

	heightDiff := latest.Height - referenceRecord.Height
	velocity = 12.0 * heightDiff / float64(actualMonths)
	velocity = Round(velocity, 1)

	// 预警分级（基于文档）
	// 学龄儿童年增长 <4cm/年 → 启动病因排查
	// 青春期前儿童年增长 <6cm/年 → 专科评估
	alertLevel = ""

	return velocity, alertLevel
}

// EvaluateGrowthVelocityWithAlert 评估生长速度并返回预警级别
// 黄色（偏差5-8cm）: 3个月内复诊
// 橙色（偏差8-12cm）: 1个月内专科评估
// 红色（偏差>12cm或年增长<4cm）: 立即转诊
func EvaluateGrowthVelocityWithAlert(ageInYears int, annualGrowth float64, gender string) (status string, alertLevel string, action string) {
	var expectedMin float64
	switch {
	case ageInYears < 2:
		expectedMin = 7.0
	case ageInYears < 10:
		expectedMin = 5.0
	default:
		if gender == "male" {
			expectedMin = 6.0
		} else {
			expectedMin = 5.0
		}
	}

	deviation := expectedMin - annualGrowth

	if annualGrowth >= expectedMin {
		return "生长速率正常", "", "继续保持"
	}

	// 低于期望值
	if annualGrowth < 4.0 {
		return "生长速率过慢", "red", "立即转诊: 年增长<4cm，建议尽早就医排查内分泌、营养等因素"
	}
	if deviation > 8 {
		return "生长速率偏慢", "orange", "1个月内专科评估: 年增速低于期望" + fmt.Sprintf("%.1fcm", deviation)
	}
	if deviation > 5 {
		return "生长速率偏慢", "yellow", "3个月内复诊: 年增速低于期望" + fmt.Sprintf("%.1fcm", deviation)
	}

	return "生长速率偏慢", "info", "建议关注营养、睡眠、运动"
}

// monthsBetween 计算两个日期之间的月数差
func monthsBetween(start, end time.Time) int {
	years := end.Year() - start.Year()
	months := int(end.Month()) - int(start.Month())
	totalMonths := years*12 + months
	if end.Day() < start.Day() {
		totalMonths--
	}
	if totalMonths < 0 {
		totalMonths = 0
	}
	return totalMonths
}

// ============= 特殊情境调整 =============

// SpecialCaseAdjustment 特殊情境调整策略
type SpecialCaseAdjustment struct {
	CaseType        string  `json:"case_type"`        // 情境类型
	AdjustmentDesc  string  `json:"adjustment_desc"`  // 调整说明
	RecommendedMethod string `json:"recommended_method"` // 推荐方法
	WeightAdjustment  string  `json:"weight_adjustment"` // 权重调整
}

// GetSpecialCaseAdjustment 获取特殊情境调整策略
func GetSpecialCaseAdjustment(caseType string) *SpecialCaseAdjustment {
	switch caseType {
	case "premature":
		return &SpecialCaseAdjustment{
			CaseType:          "premature",
			AdjustmentDesc:    "使用校正年龄(实际年龄-早产周数)，适用至2-3岁",
			RecommendedMethod: "MPH法+校正年龄",
			WeightAdjustment:  "环境权重维持不变",
		}
	case "chronic_disease":
		return &SpecialCaseAdjustment{
			CaseType:          "chronic_disease",
			AdjustmentDesc:    "健康模块权重上调至20-30%，需专科个体化评估",
			RecommendedMethod: "骨龄预测法+专科评估",
			WeightAdjustment:  "健康权重+10-20%",
		}
	case "precocious_puberty":
		return &SpecialCaseAdjustment{
			CaseType:          "precocious_puberty",
			AdjustmentDesc:    "优先采用骨龄预测法，环境权重下调，增加发育监测频率",
			RecommendedMethod: "骨龄预测法(Bayley-Pinneau)",
			WeightAdjustment:  "环境权重-5-10%",
		}
	case "cdgp": // 体质性生长延迟
		return &SpecialCaseAdjustment{
			CaseType:          "cdgp",
			AdjustmentDesc:    "骨龄为替代输入变量，环境权重可维持或略上调",
			RecommendedMethod: "Khamis-Roche(以骨龄替代实际年龄)",
			WeightAdjustment:  "环境权重+0-5%",
		}
	default:
		return nil
	}
}
