# Growth Tracker Pro AI Agent API 文档

## 概述

AI Agent 是生长追踪系统的核心智能组件，基于宝宝的全方位数据（基础信息、生长记录、化验单等）提供个性化的生长发育分析和建议。

## 核心特性

- **宝宝立体信息表**：动态构建宝宝全景画像，涵盖基础信息、发育评估、营养状况、生活方式、健康风险、生长趋势等维度
- **医疗合规守护**：严格规避医疗红线，禁止诊断、处方等医疗行为，确保回复安全合规
- **个性化推荐引擎**：基于优先级评分系统，生成针对性的营养、运动、睡眠等建议
- **多模式对话**：支持自动分析、档案查询、建议生成、报告总结等多种对话模式

---

## API 接口

### 1. AI 智能对话

**端点**: `POST /api/v1/ai/chat`

**描述**: 与 AI Agent 进行智能对话，支持多种模式（自动分析、档案查询、建议生成、报告总结）

**请求头**:
```
Authorization: Bearer <token>
Content-Type: application/json
```

**请求体**:
```json
{
  "child_id": "string",           // 宝宝ID（必填）
  "message": "string",             // 用户消息（必填）
  "mode": "auto|profile|recommendation|report",  // 对话模式（可选，默认auto）
  "context": [                    // 历史对话上下文（可选）
    {
      "role": "user|assistant",
      "content": "string"
    }
  ]
}
```

**Mode 说明**:
- `auto`: 自动模式，根据消息内容自动判断回复类型
- `profile`: 返回宝宝立体信息表
- `recommendation`: 返回个性化建议列表
- `report`: 返回综合评估报告

**响应示例**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "content": "根据宝宝目前的生长发育数据分析...",
    "profile": {
      "basic_info": {
        "name": "小明",
        "gender": "male",
        "gender_label": "男孩",
        "age_str": "5岁6个月",
        "father_height": 175,
        "mother_height": 160
      },
      "growth_assessment": {
        "current_percentile": 50,
        "percentile_status": "normal",
        "target_height": {
          "target_height": 170.5,
          "min_height": 162.5,
          "max_height": 178.5
        }
      },
      "nutrition_status": {
        "score": 75,
        "level": "good",
        "recommended_foods": ["牛奶", "鸡蛋", "瘦肉"],
        "foods_to_limit": ["甜饮料", "油炸食品"]
      },
      "lifestyle_factors": {
        "exercise_status": {"score": 60, "level": "average"},
        "sleep_status": {"score": 70, "recommended_hours": 10}
      },
      "health_risks": [],
      "growth_trend": {
        "velocity": 5.5,
        "velocity_status": "normal"
      },
      "priority_scores": {
        "nutrition": 60,
        "exercise": 70,
        "sleep": 55,
        "medical": 30
      }
    },
    "recommendations": [
      {
        "category": "exercise",
        "priority": 2,
        "title": "加强运动锻炼",
        "description": "适当的运动可以刺激生长激素分泌...",
        "actions": ["跳绳", "篮球", "游泳"],
        "timeframe": "每天坚持"
      }
    ],
    "medical_alert": null,
    "tokens": 125,
    "session_id": "session_xxx"
  }
}
```

---

### 2. 获取宝宝立体信息表

**端点**: `GET /api/v1/ai/profile/:child_id`

**描述**: 获取指定宝宝的立体信息表

**请求头**:
```
Authorization: Bearer <token>
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "basic_info": {
      "name": "小明",
      "gender": "male",
      "birthday": "2019-10-15",
      "age_str": "5岁6个月",
      "father_height": 175,
      "mother_height": 160
    },
    "growth_assessment": {
      "target_height": {"target_height": 170.5, "min_height": 162.5, "max_height": 178.5},
      "current_percentile": 50,
      "percentile_status": "normal",
      "growth_status": "normal",
      "records_count": 12,
      "measurement_frequency": "测量频率良好"
    },
    "nutrition_status": {
      "score": 75,
      "level": "good",
      "strengths": ["体重在正常范围内"],
      "concerns": [],
      "recommended_foods": ["牛奶", "鸡蛋", "瘦肉"],
      "foods_to_limit": ["甜饮料", "油炸食品"]
    },
    "lifestyle_factors": {
      "exercise_status": {"score": 60, "level": "average", "recommended_types": ["跳绳", "篮球"]},
      "sleep_status": {"score": 70, "recommended_hours": 10},
      "sunlight_status": {"score": 70, "duration": "建议每天2小时户外活动"}
    },
    "health_risks": [],
    "growth_trend": {
      "velocity": 5.5,
      "velocity_status": "normal",
      "trend_direction": "stable"
    },
    "priority_scores": {
      "nutrition": 60,
      "exercise": 70,
      "sleep": 55,
      "lifestyle": 62,
      "medical": 30
    }
  }
}
```

---

### 3. 获取个性化建议

**端点**: `GET /api/v1/ai/recommendations/:child_id`

**描述**: 获取针对宝宝的个性化建议列表

**请求头**:
```
Authorization: Bearer <token>
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "recommendations": [
      {
        "category": "nutrition",
        "priority": 1,
        "title": "优化营养摄入",
        "description": "这个年龄段的宝宝营养需求旺盛...",
        "actions": [
          "每天300-500ml奶量",
          "辅食多样化",
          "避免添加糖和盐"
        ],
        "reason": "当前营养评分为65分，有提升空间",
        "timeframe": "立即开始，长期坚持"
      },
      {
        "category": "exercise",
        "priority": 2,
        "title": "加强运动锻炼",
        "description": "适当的运动可以刺激生长激素分泌...",
        "actions": ["跳绳", "游泳", "篮球"],
        "reason": "当前运动评分为50分",
        "timeframe": "每天坚持"
      }
    ]
  }
}
```

---

### 4. 获取每日计划

**端点**: `GET /api/v1/ai/daily-plan/:child_id`

**描述**: 根据宝宝年龄和发育状况生成每日作息计划

**请求头**:
```
Authorization: Bearer <token>
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "morning": [
      "06:30-07:30 起床，早餐",
      "07:30-08:00 晨跑或跳绳500个"
    ],
    "afternoon": [
      "12:00-13:00 午餐",
      "13:00-14:00 午休",
      "16:00-18:30 运动时间（篮球、游泳）"
    ],
    "evening": [
      "18:30-19:30 晚餐",
      "20:00-21:00 拉伸运动",
      "21:30-22:00 洗漱，准备睡觉",
      "22:00 前入睡"
    ]
  }
}
```

---

### 5. 获取每周计划

**端点**: `GET /api/v1/ai/weekly-plan/:child_id`

**描述**: 获取每周运动、营养计划和检查清单

**请求头**:
```
Authorization: Bearer <token>
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "exercise_plan": [
      "周一、周三、周五：跳绳（每天1000-2000个）",
      "周二、周四：篮球或游泳",
      "周六：户外活动（爬山、骑自行车）",
      "周日：休息或轻松活动"
    ],
    "nutrition_plan": [
      "每天：300-500ml牛奶",
      "每天：1-2个鸡蛋",
      "每天：50-100g肉类或鱼类",
      "避免：甜饮料、油炸食品、零食"
    ],
    "checklist": [
      "每周测量身高体重并记录",
      "观察孩子的食欲和精神状态",
      "确保充足的户外活动时间"
    ]
  }
}
```

---

### 6. 获取综合评估报告

**端点**: `GET /api/v1/ai/report/:child_id`

**描述**: 生成宝宝生长发育综合评估报告

**请求头**:
```
Authorization: Bearer <token>
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "report": "📊 小明 生长发育综合评估报告\n\n═══════════════════════════════════════\n\n【发育状态】\n✅ 当前身高处于正常百分位范围\n- 当前身高百分位: P50\n- 靶身高范围: 162.5-178.5 cm\n- 生长速度: 5.5 cm/年 (正常)\n\n【营养状态】\n- 营养评分: 75分 (良好)\n✓ 做得好的: 体重在正常范围内\n\n【生活方式】\n- 运动状态: 60分 (一般)\n- 睡眠状态: 70分 (良好)\n- 推荐睡眠时长: 10小时\n\n【下一步行动建议】\n1. 加强运动锻炼 (exercise)\n2. 优化营养摄入 (nutrition)\n\n═══════════════════════════════════════\n⚠️ 本报告仅供参考，具体情况请咨询专业医生。\n"
  }
}
```

---

### 7. 解析化验单

**端点**: `POST /api/v1/ai/parse-report`

**描述**: 上传化验单图片，OCR识别后 AI 智能解析

**请求头**:
```
Authorization: Bearer <token>
Content-Type: application/json
```

**请求体**:
```json
{
  "child_id": "string",           // 宝宝ID（必填）
  "image_url": "string",          // 图片URL（必填）
  "report_type": "string"         // 报告类型（必填）
}
```

**report_type 可选值**:
- `bone_age`: 骨龄片
- `hormone`: 生长激素/甲状腺激素
- `blood_routine`: 血常规
- `igf1`: IGF-1 胰岛素样生长因子
- `thyroid`: 甲状腺功能
- `sex_hormone`: 性激素六项
- `other`: 其他

**响应示例**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "ocr_text": "骨龄: 8岁\n骨骺: 未闭合\n评估: 骨龄与实际年龄相符",
    "ai_result": {
      "key_indicators": [
        {"name": "骨龄", "value": "8岁", "status": "normal"},
        {"name": "骨骺状态", "value": "未闭合", "status": "normal"}
      ],
      "normal_ranges": {
        "骨龄": "与实际年龄相差±1岁为正常"
      },
      "analysis": "根据化验单分析，结果显示骨骼发育正常。",
      "suggestions": [
        "继续保持均衡营养",
        "每天保证充足睡眠",
        "每3-6个月复查骨龄"
      ]
    }
  }
}
```

---

## 错误码

| 错误码 | 描述 |
|--------|------|
| 10001 | 参数错误 |
| 10002 | 宝宝不存在 |
| 10003 | 无权限访问该宝宝 |
| 20001 | AI额度不足 |
| 20002 | AI服务暂时不可用 |
| 20003 | 化验单解析失败 |

---

## 医疗合规说明

### 禁止行为
系统严格禁止 AI 进行以下操作：
- 疾病诊断
- 开药/处方建议
- 生长激素使用建议
- 任何形式的医疗处方

### 触发词检测
当检测到以下关键词时，系统会自动：
1. 替换敏感词汇为安全表述
2. 追加医生咨询提示
3. 记录医疗告警日志

### 强制提示场景
以下情况会强制提示"请咨询专业医生"：
- 用户询问是否需要用药/打针
- 检测到严重健康风险关键词
- 身高百分位低于 P3

---

## 使用限制

### 免费用户
- 每月 3 次 AI 对话额度
- 只能使用自动模式
- 无法查看完整信息表

### VIP 会员
- 月度会员: 30 次/月
- 季度会员: 100 次/季度
- 年度会员: 400 次/年
- 全部功能开放

---

## 示例代码

### 智能对话示例

```python
import requests

url = "https://api.growth-tracker.com/api/v1/ai/chat"
headers = {
    "Authorization": "Bearer YOUR_TOKEN",
    "Content-Type": "application/json"
}
payload = {
    "child_id": "child_xxx",
    "message": "宝宝发育正常吗？",
    "mode": "auto"
}

response = requests.post(url, json=payload, headers=headers)
print(response.json())
```

### 获取宝宝信息表示例

```python
import requests

url = "https://api.growth-tracker.com/api/v1/ai/profile/child_xxx"
headers = {
    "Authorization": "Bearer YOUR_TOKEN"
}

response = requests.get(url, headers=headers)
print(response.json())
```

### 上传承化验单示例

```python
import requests

url = "https://api.growth-tracker.com/api/v1/ai/parse-report"
headers = {
    "Authorization": "Bearer YOUR_TOKEN",
    "Content-Type": "application/json"
}
payload = {
    "child_id": "child_xxx",
    "image_url": "https://storage.example.com/lab_report.jpg",
    "report_type": "bone_age"
}

response = requests.post(url, json=payload, headers=headers)
print(response.json())
```
