package agent

import (
	"strings"
)

// MedicalGuard 医疗合规守护器 - 确保AI回复不越界
type MedicalGuard struct {
	// 禁止词汇列表 (出现这些词汇需要强制提示)
	prohibitedTerms []ProhibitedTerm

	// 需要引导就医的关键词
	medicalAlertKeywords []string
}

// ProhibitedTerm 禁止词汇
type ProhibitedTerm struct {
	Term        string // 禁止词汇
	Replacement string // 替换建议
	Level       string // critical/high/medium
}

// MedicalAlert 医疗告警
type MedicalAlert struct {
	Type          string `json:"type"`           // medical_term/diagnosis/prescription/urgent
	Term          string `json:"term"`           // 触发的关键词
	Suggestion    string `json:"suggestion"`     // 建议的回复
	RequireDoctor bool   `json:"require_doctor"` // 是否强制提示"请咨询医生"
}

// NewMedicalGuard 创建医疗合规守护器
func NewMedicalGuard() *MedicalGuard {
	return &MedicalGuard{
		prohibitedTerms: []ProhibitedTerm{
			// 诊断类 - 绝对禁止
			{Term: "确诊", Replacement: "建议进行专业检查", Level: "critical"},
			{Term: "诊断为", Replacement: "建议咨询专业医生", Level: "critical"},
			{Term: "确定是", Replacement: "建议进行专业评估", Level: "critical"},
			{Term: "就是", Replacement: "可能需要专业确认", Level: "high"},

			// 处方类 - 绝对禁止
			{Term: "开药", Replacement: "建议就医咨询用药", Level: "critical"},
			{Term: "吃药", Replacement: "请遵医嘱用药", Level: "critical"},
			{Term: "服用", Replacement: "请遵医嘱", Level: "high"},
			{Term: "注射", Replacement: "请在医生指导下进行", Level: "critical"},
			{Term: "打针", Replacement: "请在医生指导下进行", Level: "high"},
			{Term: "处方", Replacement: "请咨询专业医生", Level: "critical"},
			{Term: "用药方案", Replacement: "请遵医嘱", Level: "critical"},

			// 治疗类 - 需要谨慎
			{Term: "生长激素治疗", Replacement: "建议咨询专业医生", Level: "critical"},
			{Term: "打生长激素", Replacement: "需经专业医生评估", Level: "critical"},
			{Term: "生长激素", Replacement: "需在医生指导下使用", Level: "high"},
			{Term: "手术", Replacement: "需经专业医生评估", Level: "critical"},
			{Term: "治疗方案", Replacement: "建议咨询专业医生", Level: "high"},

			// 绝对禁止的医疗行为
			{Term: "自己买药", Replacement: "请咨询医生后购买", Level: "critical"},
			{Term: "自行用药", Replacement: "请遵医嘱", Level: "critical"},
			{Term: "不需要看医生", Replacement: "建议咨询专业医生", Level: "critical"},
			{Term: "不用去医院", Replacement: "建议咨询专业医生", Level: "critical"},
		},

		medicalAlertKeywords: []string{
			// 身高相关严重情况
			"矮小症", "侏儒症", "巨人症", "性早熟", "生长激素缺乏",
			"甲状腺功能减退", "Turner综合征", "Klinefelter综合征",

			// 骨骼相关
			"骨龄提前超过2年", "骨骺闭合", "骨骼发育异常",

			// 严重营养问题
			"严重营养不良", "蛋白质-能量营养不良", "重度贫血",

			// 其他需要紧急就医的情况
			"持续发热", "体重持续下降", "发育停滞",
		},
	}
}

// CheckResponse 检查AI回复是否合规
func (g *MedicalGuard) CheckResponse(response string) (*MedicalAlert, bool) {
	// 检查禁止词汇
	for _, pt := range g.prohibitedTerms {
		if strings.Contains(response, pt.Term) {
			return &MedicalAlert{
				Type:          "prohibited_term",
				Term:          pt.Term,
				Suggestion:    pt.Replacement,
				RequireDoctor: pt.Level == "critical",
			}, false
		}
	}

	// 检查需要就医的关键词
	for _, keyword := range g.medicalAlertKeywords {
		if strings.Contains(response, keyword) {
			return &MedicalAlert{
				Type:          "medical_alert",
				Term:          keyword,
				Suggestion:    "建议尽快咨询专业医生",
				RequireDoctor: true,
			}, false
		}
	}

	return nil, true
}

// AppendDoctorConsultation 追加医生咨询提示
func (g *MedicalGuard) AppendDoctorConsultation(response string, require bool) string {
	doctorNotice := "\n\n⚠️ 以上内容仅供参考，不能替代专业医生的诊断和建议。如有疑问，请及时咨询儿科医生或内分泌科医生。"

	if require {
		return response + doctorNotice
	}

	// 检查是否已经包含医生提示
	if strings.Contains(response, "请咨询医生") || strings.Contains(response, "专业医生") {
		return response
	}

	return response + "\n\n💡 如有生长发育方面的疑问，建议咨询专业儿科医生或儿童内分泌科医生。"
}

// CheckAndSanitize 检查并清理回复
func (g *MedicalGuard) CheckAndSanitize(response string) (string, *MedicalAlert) {
	alert, ok := g.CheckResponse(response)
	if ok {
		return response, nil
	}

	// 替换禁止词汇
	cleaned := response
	for _, pt := range g.prohibitedTerms {
		if strings.Contains(cleaned, pt.Term) {
			cleaned = strings.ReplaceAll(cleaned, pt.Term, pt.Replacement)
		}
	}

	// 追加医生提示
	cleaned = g.AppendDoctorConsultation(cleaned, alert.RequireDoctor)

	return cleaned, alert
}

// IsQuestionable 判定问题是否涉及需要医疗介入的情况
func (g *MedicalGuard) IsQuestionable(question string) bool {
	question = strings.ToLower(question)

	questionablePatterns := []string{
		"是不是", "是不是有", "是不是得了",
		"需要打", "需要吃药", "需要治疗",
		"要不要去医院", "要不要看医生",
		"能否用药", "能否打针",
	}

	for _, pattern := range questionablePatterns {
		if strings.Contains(question, pattern) {
			return true
		}
	}

	return false
}

// GetMedicalDisclaimer 获取医疗免责声明
func (g *MedicalGuard) GetMedicalDisclaimer() string {
	return `📋 医疗免责声明

1. 本应用提供的AI分析仅供参考，不能替代专业医疗诊断。
2. 所有涉及疾病诊断、治疗方案、用药建议等问题，请务必咨询专业医生。
3. 儿童的生长发育情况因个体差异较大，专业医生的面诊和检查是必要的。
4. 如发现孩子生长发育异常，请及时就医，不要仅依赖AI分析结果。
5. 本应用不承担因用户自行判断或延误就医而产生的任何责任。

感谢您的理解与配合。`
}
