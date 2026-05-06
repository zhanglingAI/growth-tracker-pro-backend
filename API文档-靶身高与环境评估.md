# 靶身高 App API 接口文档

> 本文档描述靶身高与环境评估模块的所有后端接口，供前端同学对接参考。
>
> **Base URL**: `http://localhost:8080/api/v1`
> **认证方式**: JWT Token，`Authorization: Bearer jwt_token_{user_id}_{timestamp}`
> **响应格式**: 统一使用 `BaseResponse { code, msg, data }`

---

## 目录

1. [环境问卷评估](#一环境问卷评估)
2. [靶身高与生长速度](#二靶身高与生长速度)
3. [预警系统](#三预警系统)
4. [业务流程图](#四业务流程图)
5. [时序图](#五时序图)
6. [错误码汇总](#六错误码汇总)

---

## 一、环境问卷评估

### 1. 提交环境问卷评估

提交孩子的环境问卷，系统计算各模块得分、遗传靶身高、预测身高，并生成个性化行动计划。

```
POST /children/{child_id}/environment-assessment
```

**请求参数** (Body):

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| current_height | float64 | 是 | 当前身高(cm) |
| current_weight | float64 | 否 | 当前体重(kg) |
| nutrition | object | 是 | 营养模块评分 |
| nutrition.diet_diversity | int | 是 | 膳食多样性 0-2 |
| nutrition.protein_adequacy | int | 是 | 蛋白质充足度 0-2 |
| nutrition.calcium_intake | int | 是 | 钙质摄入 0-2 |
| nutrition.vitamin_d_status | int | 是 | 维生素D 0-1 |
| nutrition.bad_eating_behavior | int | 是 | 不良饮食行为 0-2 |
| nutrition.weight_management | int | 是 | 体重管理 0-1 |
| sleep | object | 是 | 睡眠模块评分 |
| sleep.duration | float64 | 是 | 睡眠时长(小时) |
| sleep.bedtime_regularity | int | 是 | 入睡规律性 0-2 |
| sleep.deep_sleep_cover | int | 是 | 深睡眠覆盖 0-2 |
| sleep.sleep_continuity | int | 是 | 睡眠连续性 0-2 |
| sleep.sleep_environment | int | 是 | 睡眠环境 0-2 |
| exercise | object | 是 | 运动模块评分 |
| exercise.frequency | int | 是 | 运动频率 0-3 |
| exercise.type_suitability | int | 是 | 类型适宜性 0-3 |
| exercise.duration | int | 是 | 时长适中性 0-3 |
| exercise.intensity | int | 是 | 强度分级 0-3 |
| health | object | 是 | 健康模块评分 |
| health.disease_control | int | 是 | 疾病控制 0-2 |
| health.checkup_compliance | int | 是 | 体检依从性 0-2 |
| health.medication_safety | int | 是 | 用药安全 0-1 |
| mental | object | 是 | 心理模块评分 |
| mental.emotion_regulation | int | 是 | 情绪调节 0-2 |
| mental.family_support | int | 是 | 家庭支持 0-2 |
| mental.stress_management | int | 是 | 压力管理 0-1 |

**请求示例**:

```json
{
  "current_height": 135.5,
  "current_weight": 32.0,
  "nutrition": {
    "diet_diversity": 1,
    "protein_adequacy": 2,
    "calcium_intake": 1,
    "vitamin_d_status": 0,
    "bad_eating_behavior": 1,
    "weight_management": 1
  },
  "sleep": {
    "duration": 8.5,
    "bedtime_regularity": 1,
    "deep_sleep_cover": 1,
    "sleep_continuity": 2,
    "sleep_environment": 2
  },
  "exercise": {
    "frequency": 2,
    "type_suitability": 2,
    "duration": 2,
    "intensity": 2
  },
  "health": {
    "disease_control": 2,
    "checkup_compliance": 1,
    "medication_safety": 1
  },
  "mental": {
    "emotion_regulation": 2,
    "family_support": 1,
    "stress_management": 1
  }
}
```

**响应示例** (成功):

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "child_id": "child-uuid-1234",
    "assessment_date": "2026-05-03",
    "module_scores": {
      "nutrition": {
        "score": 6.0,
        "max_score": 15.0,
        "weight": 0.30,
        "weighted_score": 1.8,
        "percentage": 0.4,
        "priority": "最高"
      },
      "sleep": {
        "score": 6.5,
        "max_score": 12.5,
        "weight": 0.25,
        "weighted_score": 1.625,
        "percentage": 0.52,
        "priority": "高"
      },
      "exercise": {
        "score": 8.0,
        "max_score": 12.5,
        "weight": 0.25,
        "weighted_score": 2.0,
        "percentage": 0.64,
        "priority": "中"
      },
      "health": {
        "score": 4.0,
        "max_score": 5.0,
        "weight": 0.10,
        "weighted_score": 0.4,
        "percentage": 0.8,
        "priority": "低"
      },
      "mental": {
        "score": 3.5,
        "max_score": 5.0,
        "weight": 0.10,
        "weighted_score": 0.35,
        "percentage": 0.7,
        "priority": "低"
      }
    },
    "total_score": 28.0,
    "max_possible_score": 50.0,
    "intervention_zone": "medium",
    "zone_label": "良好",
    "interpretation": "得分28.0分(中分区): 针对性改善薄弱环节，每3-6月复评",
    "genetic_target_height": 172.5,
    "environment_increment": 5.6,
    "predicted_height": 178.1,
    "prediction_method": "Khamis-Roche + 环境增量",
    "error_range": 8.0,
    "khamis_roche": {
      "target_height": 172.5,
      "lower_bound": 164.5,
      "upper_bound": 180.5,
      "father_height": 175.0,
      "mother_height": 162.0,
      "gender": "male",
      "method": "Khamis-Roche",
      "error_range": 8.0
    },
    "age_weights": {
      "age_group": "学龄期",
      "min_age": 6,
      "max_age": 12,
      "genetic_weight": 0.75,
      "environment_weight": 0.25,
      "core_intervention": "运动、睡眠、学业压力"
    },
    "action_plan": {
      "top_priorities": [
        {
          "priority": 1,
          "module": "nutrition",
          "title": "每天早餐加一杯牛奶+一个鸡蛋",
          "description": "确保蛋白质和钙质摄入，这是最容易执行的起点",
          "why": "蛋白质是生长激素合成的原料，钙是骨骼矿化的基础",
          "how_to_start": "明天早餐开始，固定200ml牛奶+1个鸡蛋",
          "difficulty": "easy"
        },
        {
          "priority": 2,
          "module": "sleep",
          "title": "今晚21:30前上床熄灯",
          "description": "22:00-02:00是生长激素分泌黄金窗口",
          "why": "错过这个窗口无法弥补，深睡眠时分泌量占全天50-70%",
          "how_to_start": "21:00关电视/手机，21:30准时熄灯",
          "difficulty": "medium"
        },
        {
          "priority": 3,
          "module": "nutrition",
          "title": "每天保证5种颜色的食物",
          "description": "白(米/奶)、绿(蔬菜)、红(肉/番茄)、黄(蛋/玉米)、紫/深色",
          "why": "食物多样性确保微量营养素全面覆盖",
          "how_to_start": "记录今天吃了几种颜色，明天补缺少的颜色",
          "difficulty": "easy"
        }
      ],
      "nutrition_plan": [
        "每天早餐固定一杯200ml牛奶",
        "每天保证2种蔬菜+1种水果",
        "每周吃2次鱼或虾",
        "减少炸鸡、薯条、奶茶等高糖高脂零食"
      ],
      "sleep_plan": [
        "21:30前上床，22:00前入睡",
        "睡前1小时不用电子设备",
        "卧室保持黑暗、安静、18-22℃",
        "午睡不超过30分钟"
      ],
      "exercise_plan": [
        "每天跳绳10分钟或摸高50次",
        "每周2次游泳或篮球",
        "每周设1-2天完全休息",
        "运动后30分钟内补充牛奶+香蕉"
      ],
      "health_plan": [
        "每年至少1次体检(身高体重+血常规)",
        "每月固定日期测量身高",
        "如有慢性病，定期复查控制情况"
      ],
      "mental_plan": [
        "每天15分钟专注陪伴",
        "学业压力大时优先保证睡眠",
        "家庭冲突避免当着孩子面发生",
        "情绪持续低落超过2周寻求专业帮助"
      ],
      "track_reminder": "建议每月复评一次，持续追踪改善效果"
    },
    "clinical_interpretation": "得分28.0分(中分区): 针对性改善薄弱环节，每3-6月复评"
  }
}
```

**错误响应**:

```json
{
  "code": 400,
  "msg": "参数错误: Key: 'CreateEnvironmentAssessmentRequest.Nutrition.DietDiversity' Error:Field validation for 'DietDiversity' failed on the 'min' tag",
  "data": null
}
```

---

### 2. 获取最新环境评估结果

```
GET /children/{child_id}/environment-assessment/latest
```

**响应**: 同 "提交环境问卷评估" 的成功响应

**错误响应** (暂无记录):

```json
{
  "code": 404,
  "msg": "暂无评估记录",
  "data": null
}
```

---

### 3. 获取环境评估历史列表

```
GET /children/{child_id}/environment-assessment/history?page=1&page_size=20
```

**请求参数** (Query):

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| page_size | int | 否 | 每页条数，默认20 |

**响应示例**:

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "items": [
      {
        "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
        "assessment_date": "2026-05-03",
        "total_score": 28.0,
        "intervention_zone": "medium",
        "predicted_height": 178.1,
        "environment_increment": 5.6
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 20
  }
}
```

---

## 二、靶身高与生长速度

### 4. 靶身高综合分析

获取孩子的遗传靶身高、Khamis-Roche预测、环境预测、当前百分位、生长速度等综合分析。

```
GET /children/{child_id}/target-height-comparison
```

**响应示例**:

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "genetic_target_height": 172.5,
    "quantitative_genetics": {
      "target_height": 171.2,
      "heritability": 0.75,
      "father_selection": 3.0,
      "mother_selection": 2.0,
      "child_deviation": 3.75,
      "explanation": "父亲身高高于均值3.0cm，母亲身高高于均值2.0cm"
    },
    "khamis_roche": {
      "target_height": 172.5,
      "lower_bound": 164.5,
      "upper_bound": 180.5,
      "father_height": 175.0,
      "mother_height": 162.0,
      "gender": "male",
      "method": "Khamis-Roche",
      "error_range": 8.0
    },
    "environment_prediction": {
      "genetic_target_height": 172.5,
      "environment_score": {
        "total_score": 28.0,
        "max_possible_score": 50.0,
        "module_scores": {},
        "interpretation": "得分28.0分(中分区): 针对性改善薄弱环节，每3-6月复评",
        "intervention_zone": "medium"
      },
      "environment_increment": 5.6,
      "predicted_height": 178.1,
      "prediction_method": "Khamis-Roche + 环境增量",
      "error_range": 8.0,
      "age_weights": {
        "age_group": "学龄期",
        "min_age": 6,
        "max_age": 12,
        "genetic_weight": 0.75,
        "environment_weight": 0.25,
        "core_intervention": "运动、睡眠、学业压力"
      },
      "clinical_interpretation": "综合预测身高: 178.1cm(±8.0cm)\n遗传靶身高: 172.5cm\n环境得分: 28.0/50分，环境增量: +5.6cm\n干预分区: 得分28.0分(中分区): 针对性改善薄弱环节，每3-6月复评\n年龄分层: 遗传权重75%，环境权重25%"
    },
    "current_height": 135.5,
    "current_percentile": 45,
    "target_percentile": 50,
    "potential_status": "遗传潜力正常发挥",
    "growth_velocity": {
      "velocity": 6.2,
      "months_back": 12,
      "latest_height": 135.5,
      "previous_height": 129.3,
      "status": "生长速率正常",
      "alert_level": "",
      "action": "继续保持"
    }
  }
}
```

**错误响应**:

```json
{
  "code": 500,
  "msg": "暂无生长记录",
  "data": null
}
```

---

### 5. 生长速度监测

计算孩子最近N个月的年生长速度，并评估是否正常。

```
GET /children/{child_id}/growth-velocity?months_back=12
```

**请求参数** (Query):

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| months_back | int | 否 | 回溯月数，默认12个月 |

**响应示例**:

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "velocity": 6.2,
    "months_back": 12,
    "latest_height": 135.5,
    "previous_height": 129.3,
    "expected_min": 5.0,
    "status": "生长速率正常",
    "alert_level": "",
    "action": "继续保持",
    "deviation": -1.2
  }
}
```

**预警状态说明**:

| status | alert_level | 含义 |
|--------|-------------|------|
| 生长速率正常 | "" | 年增速达到期望值 |
| 生长速率偏慢 | yellow | 偏差5-8cm，3个月内复诊 |
| 生长速率偏慢 | orange | 偏差8-12cm，1个月内专科评估 |
| 生长速率偏慢 | red | 年增长<4cm，立即转诊 |
| 生长速率偏慢 | info | 建议关注营养、睡眠、运动 |

---

## 三、预警系统

### 6. 获取宝宝预警列表

```
GET /children/{child_id}/alerts?page=1&page_size=20&level=
```

**请求参数** (Query):

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| page_size | int | 否 | 每页条数，默认20 |
| level | string | 否 | 筛选等级: info/warning/danger |

**响应示例**:

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "items": [
      {
        "id": "alert-uuid-001",
        "child_id": "child-uuid-1234",
        "user_id": "user-uuid-5678",
        "alert_type": "target_gap_low",
        "alert_level": "warning",
        "title": "当前身高低于靶身高下限",
        "description": "当前身高135.5cm，靶身高下限164.5cm，差距较大",
        "dimension": "身高差距",
        "metric_value": 29.0,
        "threshold": 5.0,
        "is_read": false,
        "is_dismissed": false,
        "created_at_ago": "2天前"
      }
    ],
    "total": 3,
    "unread_count": 2,
    "active_count": 3
  }
}
```

---

### 7. 标记预警已读

```
POST /alerts/{alert_id}/read
```

**响应**:

```json
{
  "code": 0,
  "msg": "标记成功",
  "data": null
}
```

---

### 8. 忽略预警

```
POST /alerts/{alert_id}/dismiss
```

**请求参数** (Body):

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| alert_id | string | 是 | 预警ID |
| reason | string | 否 | 忽略原因 |

**响应**:

```json
{
  "code": 0,
  "msg": "已忽略",
  "data": null
}
```

---

### 9. 获取预警摘要

```
GET /alerts/summary
```

**响应示例**:

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "has_active_alert": true,
    "highest_level": "warning",
    "top_alerts": [
      {
        "id": "alert-uuid-001",
        "child_id": "child-uuid-1234",
        "alert_type": "target_gap_low",
        "alert_level": "warning",
        "title": "当前身高低于靶身高下限",
        "description": "当前身高135.5cm，靶身高下限164.5cm，差距较大",
        "is_read": false,
        "created_at_ago": "2天前"
      }
    ],
    "total_active": 3
  }
}
```

---

### 10. 设置生长阶段

```
POST /children/{child_id}/growth-stage
```

**请求参数** (Body):

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| growth_stage | string | 是 | pre_puberty(青春期前)/puberty(青春期)/post_puberty(青春期后) |
| source | string | 是 | self_assessment(自评)/doctor_visit(医生诊断) |

**响应**:

```json
{
  "code": 0,
  "msg": "设置成功",
  "data": null
}
```

---

## 四、业务流程图

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           靶身高 App 主业务流程                               │
└─────────────────────────────────────────────────────────────────────────────┘

  ┌──────────────┐
  │   用户登录    │
  └──────┬───────┘
         │
         ▼
  ┌──────────────┐     否      ┌──────────────┐
  │ 有无宝宝资料? │───────────▶│  创建宝宝     │
  └──────┬───────┘            └──────────────┘
         │ 是
         ▼
  ┌──────────────┐
  │  录入当前身高 │◀─────────────────────────────────────┐
  └──────┬───────┘                                      │
         │                                               │
         ▼                                               │
  ┌──────────────────────────┐                          │
  │  环境问卷评估 (26题)      │                          │
  │  ├── 基础信息 (Q1-Q6)     │                          │
  │  ├── 营养状况 (Q7-Q12)    │                          │
  │  ├── 睡眠质量 (Q13-Q17)   │                          │
  │  ├── 运动状况 (Q18-Q21)   │                          │
  │  ├── 健康状况 (Q22-Q23)   │                          │
  │  └── 心理状况 (Q24-Q26)   │                          │
  └──────┬───────────────────┘                          │
         │                                               │
         ▼                                               │
  ┌──────────────────────────┐                          │
  │      后端计算引擎         │                          │
  │  ├─ 遗传靶身高(MPH)      │                          │
  │  ├─ Khamis-Roche预测     │                          │
  │  ├─ 环境得分(0-50)       │                          │
  │  ├─ 预测身高             │                          │
  │  └─ 个性化行动计划       │                          │
  └──────┬───────────────────┘                          │
         │                                               │
         ▼                                               │
  ┌──────────────────────────┐                          │
  │      结果展示页面         │                          │
  │  ├─ 综合得分与等级        │                          │
  │  ├─ 各模块雷达图          │                          │
  │  ├─ 靶身高对比图          │                          │
  │  ├─ 预测身高区间          │                          │
  │  └─ 本周3个优先行动       │                          │
  └──────┬───────────────────┘                          │
         │                                               │
         ▼                                               │
  ┌──────────────────────────┐     是                   │
  │   是否需要追踪复评?       │──────────────────────────┘
  └──────┬───────────────────┘    (每月/每3月复评)
         │ 否
         ▼
  ┌──────────────┐
  │   结束       │
  └──────────────┘
```

---

## 五、时序图

### 5.1 提交环境问卷评估

```
  前端小程序          后端API             预警引擎             数据库
      │                 │                  │                  │
      │  POST /children/{id}/environment-assessment           │
      │────────────────▶│                  │                  │
      │                 │                  │                  │
      │                 │── 1. 校验宝宝归属权 ─────────────────▶│
      │                 │◀─ 宝宝信息 ──────│                  │
      │                 │                  │                  │
      │                 │── 2. 计算年龄 ────│                  │
      │                 │◀─ 年龄结果 ──────│                  │
      │                 │                  │                  │
      │                 │── 3. 计算环境得分 ──────────────────▶│
      │                 │◀─ 各模块得分+总分 ──────────────────│
      │                 │                  │                  │
      │                 │── 4. 计算Khamis-Roche ─────────────▶│
      │                 │◀─ 预测身高 ──────│                  │
      │                 │                  │                  │
      │                 │── 5. 生成行动计划 ──────────────────▶│
      │                 │◀─ 行动计划 ──────│                  │
      │                 │                  │                  │
      │                 │── 6. 保存评估记录 ──────────────────▶│
      │                 │◀─ 保存成功 ──────│                  │
      │                 │                  │                  │
      │◀─ 评估结果+行动计划 ───────────────│                  │
      │                 │                  │                  │
```

### 5.2 创建生长记录触发预警

```
  前端小程序          后端API             预警引擎             数据库
      │                 │                  │                  │
      │  POST /records  │                  │                  │
      │────────────────▶│                  │                  │
      │                 │                  │                  │
      │                 │── 1. 保存记录 ──────────────────────▶│
      │                 │◀─ 保存成功 ──────│                  │
      │                 │                  │                  │
      │                 │── 2. 调用预警评估 ──────────────────▶│
      │                 │   (evaluateAndSaveAlerts)           │
      │                 │                  │                  │
      │                 │── 3. 获取所有历史记录 ──────────────▶│
      │                 │◀─ 记录列表 ──────│                  │
      │                 │                  │                  │
      │                 │── 4. 计算百分位 ────────────────────▶│
      │                 │◀─ 当前/区域百分位 ─────────────────│
      │                 │                  │                  │
      │                 │── 5. 运行6维度预警检查 ─────────────▶│
      │                 │   ├─ 靶身高差距                     │
      │                 │   ├─ 区域偏差                       │
      │                 │   ├─ 骨龄超前/延迟                  │
      │                 │   ├─ 生长停滞                       │
      │                 │   ├─ 速度偏慢                       │
      │                 │   └─ 百分位下降                     │
      │                 │◀─ 预警列表 ──────│                  │
      │                 │                  │                  │
      │                 │── 6. 去重并保存预警 ────────────────▶│
      │                 │◀─ 保存成功 ──────│                  │
      │                 │                  │                  │
      │◀─ 记录创建成功 ──│                  │                  │
      │                 │                  │                  │
```

---

## 六、错误码汇总

| 错误码 | 含义 | 常见场景 |
|--------|------|----------|
| 0 | 成功 | - |
| 400 | 参数错误 | 必填字段缺失、字段类型错误、字段值超出范围 |
| 401 | 未授权 | Token缺失或过期 |
| 403 | 禁止访问 | 访问了不属于自己的宝宝数据 |
| 404 | 未找到 | 宝宝不存在、评估记录不存在、预警不存在 |
| 500 | 服务器错误 | 数据库错误、计算异常 |
| 1001 | 额度耗尽 | AI调用次数用完 |
| 1002 | 非会员 | 需要订阅会员才能使用 |
| 2001 | 邀请码无效 | 家庭邀请码错误 |
| 2002 | 家庭已满 | 家庭成员数量达到上限 |

---

## 附录：问卷模块与DTO字段对照表

| 问卷问题 | 所属模块 | DTO字段 | 分值范围 |
|----------|----------|---------|----------|
| Q7 饮食多样性 | nutrition | diet_diversity | 0-2 |
| Q8 蛋白质摄入 | nutrition | protein_adequacy | 0-2 |
| Q9 钙质摄入 | nutrition | calcium_intake | 0-2 |
| Q10 维生素D | nutrition | vitamin_d_status | 0-1 |
| Q11 不良饮食习惯 | nutrition | bad_eating_behavior | 0-2 |
| Q12 体重管理 | nutrition | weight_management | 0-1 |
| Q13 睡眠时长 | sleep | duration (float) + bedtime_regularity | 综合计算 |
| Q14 入睡时间 | sleep | bedtime_regularity | 0-2 |
| Q15 深睡眠覆盖 | sleep | deep_sleep_cover | 0-2 |
| Q16 睡眠连续性 | sleep | sleep_continuity | 0-2 |
| Q17 睡眠环境 | sleep | sleep_environment | 0-2 |
| Q18 运动频率 | exercise | frequency | 0-3 |
| Q19 运动类型 | exercise | type_suitability | 0-3 |
| Q20 单次时长 | exercise | duration | 0-3 |
| Q21 运动强度 | exercise | intensity | 0-3 |
| Q22 慢性病 | health | disease_control | 0-2 |
| Q23 体检 | health | checkup_compliance | 0-2 |
| - 用药安全 | health | medication_safety | 0-1 |
| Q24 情绪状态 | mental | emotion_regulation | 0-2 |
| Q25 家庭环境 | mental | family_support | 0-2 |
| Q26 学业压力 | mental | stress_management | 0-1 |

---

*文档生成时间: 2026-05-03*
*API版本: v1.0*
