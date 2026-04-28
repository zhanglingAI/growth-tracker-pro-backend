# Growth Tracker Pro 系统架构文档

**文档版本**: v2.0
**编制日期**: 2026-04-28
**最后更新**: 基于 PRD v2.0 补全所有功能模块

---

## 1. 系统概述

### 1.1 产品定位
**一句话定义**: 帮助家长科学追踪儿童生长发育，AI辅助解读检查报告，判断是否需要科学干预。

### 1.2 核心价值主张
不只记录身高，更帮助你了解孩子的发育状态，判断是否需要科学干预。

### 1.3 系统边界

```
┌─────────────────────────────────────────────────────────────────┐
│                         用户端 (微信小程序)                       │
│  登录/注册 │ 宝宝管理 │ 生长记录 │ 曲线图表 │ AI解析 │ 医院推荐   │
└────────────────────────────┬────────────────────────────────────┘
                             │ HTTPS + JWT
┌────────────────────────────┴────────────────────────────────────┐
│                      后端服务 (Go + Gin)                         │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌────────┐ │
│  │ 认证模块 │ │ 宝宝模块 │ │ 记录模块 │ │ AI Agent │ │ 会员模块│ │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └────────┘ │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐           │
│  │ 家庭组模块│ │ 医院推荐 │ │ 报告解析 │ │ 订阅消息 │           │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘           │
└────────────────────────────┬────────────────────────────────────┘
                             │
         ┌───────────────────┼───────────────────┐
         ▼                   ▼                   ▼
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│    MySQL     │    │    Redis     │    │  外部服务     │
│  (主数据库)   │    │   (缓存)     │    │ 微信支付/地图 │
└──────────────┘    └──────────────┘    └──────────────┘
```

---

## 2. 技术架构

### 2.1 技术栈

| 层级 | 技术选型 | 说明 |
|-----|---------|------|
| **后端框架** | Go 1.21+ / Gin | 高性能 RESTful API |
| **ORM** | GORM | MySQL 数据库操作 |
| **缓存** | Redis | 会话、热点数据缓存 |
| **认证** | JWT | 无状态身份认证 |
| **外部集成** | 微信 OpenAPI / 微信支付 | 登录、支付能力 |
| **AI服务** | DeepSeek-VL / DeepSeek API | 化验单解析、对话 |
| **地图服务** | 腾讯位置服务 / 高德地图 | LBS 医院推荐 |
| **部署** | Docker / Docker Compose | 容器化部署 |

### 2.2 项目结构

```
growth-tracker-pro-backend/
├── cmd/
│   └── server/
│       └── main.go              # 应用入口
├── internal/
│   ├── config/                  # 配置管理
│   │   └── config.go
│   ├── models/                  # 数据模型
│   │   ├── models.go           # 核心实体
│   │   ├── dto.go              # 数据传输对象
│   │   └── standards.go        # LMS标准数据
│   ├── repository/              # 数据访问层
│   │   ├── repository.go       # 仓储接口
│   │   ├── user_repo.go       # 用户仓储
│   │   ├── child_repo.go      # 宝宝仓储
│   │   ├── record_repo.go     # 记录仓储
│   │   ├── family_repo.go     # 家庭组仓储
│   │   ├── hospital_repo.go    # 医院仓储
│   │   ├── membership_repo.go  # 会员仓储
│   │   └── report_repo.go     # 报告仓储
│   ├── service/                 # 业务逻辑层
│   │   ├── service.go
│   │   ├── auth_service.go    # 认证服务
│   │   ├── child_service.go   # 宝宝服务
│   │   ├── record_service.go  # 记录服务
│   │   ├── family_service.go  # 家庭服务
│   │   ├── hospital_service.go # 医院服务
│   │   ├── membership_service.go # 会员服务
│   │   ├── report_service.go  # 报告服务
│   │   └── subscription_service.go # 订阅服务
│   ├── handler/                 # HTTP 处理层
│   │   ├── handler.go
│   │   ├── auth_handler.go    # 认证接口
│   │   ├── child_handler.go   # 宝宝接口
│   │   ├── record_handler.go  # 记录接口
│   │   ├── family_handler.go  # 家庭接口
│   │   ├── hospital_handler.go # 医院接口
│   │   ├── membership_handler.go # 会员接口
│   │   └── report_handler.go  # 报告接口
│   ├── agent/                   # AI 智能体模块
│   │   ├── ai_agent.go        # 核心智能体
│   │   ├── child_profile.go   # 宝宝画像
│   │   ├── medical_guard.go   # 医疗合规守护
│   │   └── recommendation_engine.go # 推荐引擎
│   ├── middleware/              # 中间件
│   │   ├── auth.go            # JWT 认证中间件
│   │   ├── ratelimit.go       # 限流中间件
│   │   └── cors.go            # 跨域中间件
│   ├── utils/                  # 工具函数
│   │   ├── lms.go             # LMS百分位计算
│   │   ├── target_height.go   # 靶身高计算
│   │   ├── response.go        # 统一响应
│   │   └── wechat.go          # 微信工具
│   └── constant/               # 常量定义
│       └── constants.go
├── migrations/                  # 数据库迁移
│   └── init.sql
├── docs/                        # API 文档
│   ├── API.md
│   └── ai_agent_api.md
├── scripts/                     # 脚本
│   └── deploy.sh
├── config.yaml                  # 配置文件
├── Dockerfile
├── docker-compose.yml
└── README.md
```

---

## 3. 功能模块

### 3.1 功能模块列表

| 模块 | 功能 | 优先级 | 状态 |
|-----|------|-------|------|
| **P0 - 核心功能** | | | |
| 认证模块 | 微信登录/注册、JWT认证 | P0 | 🔄 开发中 |
| 宝宝管理 | 添加/编辑/删除宝宝信息 | P0 | ✅ 已完成 |
| 生长记录 | 身高体重记录、编辑、删除 | P0 | ✅ 已完成 |
| 同龄对比 | 中国标准/WHO标准百分位计算 | P0 | ✅ 已完成 |
| 异常预警 | 生长异常检测与就医指引 | P0 | ✅ 已完成 |
| **P1 - MVP功能** | | | |
| 医院推荐 | LBS附近儿童医院推荐 | P1 | 🔄 开发中 |
| 家庭组 | 家庭成员邀请与共享 | P1 | 🔄 开发中 |
| AI报告解析 | 化验单图片识别与解读 | P1 | 🔄 开发中 |
| 会员体系 | 月卡/季卡/年卡购买 | P1 | 🔄 开发中 |
| 订阅消息 | 身高录入提醒 | P2 | ⏳ 待开发 |
| 数据导出 | PDF报告生成 | P2 | ⏳ 待开发 |

### 3.2 认证模块 (auth)

**功能说明**: 微信静默登录，获取用户 openid，完成用户注册。

**业务流程**:
```
用户首次打开小程序
    ↓
wx.login() → 获取 code
    ↓
后端通过 code 换取 openid
    ↓
查询 users 表是否存在该 openid
    ├── 不存在 → 引导填写家长信息 → 创建用户
    └── 存在 → 直接进入首页
```

**API 接口**:
| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | /api/v1/auth/login | 微信登录 |
| GET | /api/v1/auth/profile | 获取用户信息 |
| PUT | /api/v1/auth/profile | 更新用户信息 |

**数据模型 - User**:
```go
type User struct {
    ID        uint      `gorm:"primaryKey"`
    OpenID    string    `gorm:"uniqueIndex;size:64"`  // 微信 openid
    Nickname  string    `gorm:"size:50"`              // 昵称
    Avatar    string    `gorm:"size:500"`             // 头像URL
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### 3.3 宝宝管理模块 (child)

**功能说明**: 管理宝宝的档案信息，支持添加、编辑、删除多个宝宝。

**API 接口**:
| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/v1/children | 获取宝宝列表 |
| POST | /api/v1/children | 添加宝宝 |
| GET | /api/v1/children/:id | 获取宝宝详情 |
| PUT | /api/v1/children/:id | 更新宝宝信息 |
| DELETE | /api/v1/children/:id | 删除宝宝 |
| GET | /api/v1/children/:id/profile | AI宝宝画像 |

**数据模型 - Child**:
```go
type Child struct {
    ID           uint      `gorm:"primaryKey"`
    UserID       uint      `gorm:"index"`              // 所属用户
    FamilyID     *uint     `gorm:"index"`              // 所属家庭组
    Nickname     string    `gorm:"size:50"`            // 宝宝昵称
    Gender       string    `gorm:"size:10"`           // 男/女
    Birthday     time.Time                             // 出生日期
    InitialHeight float32                               // 初始身高(cm)
    InitialWeight *float32                            // 初始体重(kg)
    StandardType  string    `gorm:"size:10;default:'cn'"` // cn/who
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

### 3.4 生长记录模块 (record)

**功能说明**: 每周录入宝宝的身高体重数据，系统自动计算同龄对比结果。

**API 接口**:
| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/v1/children/:id/records | 获取记录列表 |
| POST | /api/v1/children/:id/records | 添加记录 |
| GET | /api/v1/records/:id | 获取记录详情 |
| PUT | /api/v1/records/:id | 更新记录 |
| DELETE | /api/v1/records/:id | 删除记录 |
| GET | /api/v1/children/:id/growth-curve | 获取生长曲线数据 |
| GET | /api/v1/children/:id/comparison | 获取同龄对比结果 |
| GET | /api/v1/children/:id/target-height | 计算靶身高 |

**数据模型 - GrowthRecord**:
```go
type GrowthRecord struct {
    ID           uint      `gorm:"primaryKey"`
    ChildID      uint      `gorm:"index"`
    MeasureDate  time.Time                            // 测量日期
    Height       float32                              // 身高(cm)
    Weight       *float32                            // 体重(kg)
    HeightPercentile  *float32                        // 身高百分位
    WeightPercentile  *float32                       // 体重百分位
    HeightZScore      *float32                        // 身高Z分数
    WeightZScore      *float32                        // 体重Z分数
    HeightStatus      string    `gorm:"size:20"`     // normal/low/high/very_low/very_high
    WeightStatus      string    `gorm:"size:20"`
    Remarks           string    `gorm:"size:500"`    // 备注
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

### 3.5 家庭组模块 (family)

**功能说明**: 邀请家人共同查看和管理宝宝的生长数据。

**家庭角色**:
| 角色 | 权限 |
|-----|------|
| 创建者 | 全部权限 |
| 成员 | 查看/编辑 |
| 访客 | 仅查看 |

**API 接口**:
| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | /api/v1/families | 创建家庭组 |
| GET | /api/v1/families/:id | 获取家庭组信息 |
| GET | /api/v1/families | 获取我的家庭组列表 |
| POST | /api/v1/families/:id/invite | 生成邀请码 |
| POST | /api/v1/families/join | 加入家庭组 |
| DELETE | /api/v1/families/:id/members/:userId | 移除成员 |
| GET | /api/v1/families/:id/children | 获取家庭内所有宝宝 |

**数据模型 - Family**:
```go
type Family struct {
    ID        uint      `gorm:"primaryKey"`
    Name      string    `gorm:"size:50"`              // 家庭名称
    InviteCode string   `gorm:"uniqueIndex;size:6"`  // 6位邀请码
    CreatedAt time.Time
    UpdatedAt time.Time
}

type FamilyMember struct {
    ID        uint      `gorm:"primaryKey"`
    FamilyID  uint      `gorm:"index"`
    UserID    uint      `gorm:"index"`
    Role      string    `gorm:"size:20"`              // creator/member/guest
    CreatedAt time.Time
}
```

### 3.6 医院推荐模块 (hospital)

**功能说明**: 基于用户位置，推荐附近提供儿童内分泌科的医院。

**功能特点**:
- 使用地图组件展示
- 支持按距离排序
- 显示医院名称、地址、电话、科室信息
- 支持导航和拨打电话
- 显示预估检查费用

**API 接口**:
| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/v1/hospitals/nearby | 附近医院列表 |
| GET | /api/v1/hospitals/:id | 医院详情 |
| GET | /api/v1/hospitals/:id/departments | 医院科室列表 |

**数据模型 - Hospital**:
```go
type Hospital struct {
    ID           uint      `gorm:"primaryKey"`
    Name         string    `gorm:"size:200"`          // 医院名称
    Level        string    `gorm:"size:20"`           // 三级甲等/二级甲等
    Address      string    `gorm:"size:500"`          // 地址
    Latitude     float64                              // 纬度
    Longitude    float64                              // 经度
    Phone        string    `gorm:"size:20"`           // 电话
    Logo         string    `gorm:"size:500"`          // 医院Logo
    PediatricEndo bool     `gorm:"default:true"`     // 有儿童内分泌科
    EstimatedFee string    `gorm:"size:50"`           // 预估费用
    City         string    `gorm:"size:50"`           // 城市
    District     string    `gorm:"size:50"`           // 区县
}

type HospitalDepartment struct {
    ID          uint      `gorm:"primaryKey"`
    HospitalID  uint      `gorm:"index"`
    Name        string    `gorm:"size:50"`            // 科室名称
    Description string    `gorm:"size:500"`          // 科室描述
}
```

### 3.7 会员模块 (membership)

**功能说明**: 会员购买与权益管理。

**会员方案**:
| 方案 | 价格 | AI解析额度 |
|-----|------|-----------|
| 月卡 | ¥29.9/月 | 20次/月 |
| 季卡 | ¥69.9/季 | 20次/月 |
| 年卡 | ¥199.9/年 | 20次/月 |

**API 接口**:
| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/v1/membership/status | 获取会员状态 |
| POST | /api/v1/membership/purchase | 购买会员 |
| GET | /api/v1/membership/quota | 获取AI解析额度 |
| POST | /api/v1/membership/webhook | 微信支付回调 |

**数据模型 - Membership**:
```go
type Membership struct {
    ID           uint      `gorm:"primaryKey"`
    UserID       uint      `gorm:"uniqueIndex"`
    PlanType     string    `gorm:"size:20"`           // monthly/quarterly/yearly
    StartDate    time.Time
    EndDate      time.Time
    Status       string    `gorm:"size:20"`          // active/expired/cancelled
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

type UsageQuota struct {
    ID            uint      `gorm:"primaryKey"`
    UserID        uint      `gorm:"index"`
    Year          int
    Month         int
    UsedCount     int      `gorm:"default:0"`        // 已使用次数
    FreeQuota     int      `gorm:"default:3"`        // 免费额度
    PaidQuota     int      `gorm:"default:20"`       // 付费额度
}
```

### 3.8 AI报告解析模块 (report)

**功能说明**: 用户上传化验单图片，AI自动识别关键指标并给出解读。

**免费额度**: 每月3张免费AI解析

**支持报告类型**:
- 血常规
- 肝肾功能
- 甲状腺功能
- 生长激素检测
- 性激素六项
- 骨龄片报告

**API 接口**:
| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | /api/v1/reports/upload | 上传化验单图片 |
| POST | /api/v1/reports/analyze | AI解析化验单 |
| GET | /api/v1/reports | 获取报告列表 |
| GET | /api/v1/reports/:id | 获取报告详情 |
| DELETE | /api/v1/reports/:id | 删除报告 |

**数据模型 - Report**:
```go
type Report struct {
    ID           uint      `gorm:"primaryKey"`
    UserID       uint      `gorm:"index"`
    ChildID      uint      `gorm:"index"`
    ReportType   string    `gorm:"size:50"`          // 报告类型
    ImageURL     string    `gorm:"size:500"`         // 图片URL
    Hospital     string    `gorm:"size:100"`         // 医院名称
    ReportDate   *time.Time                           // 报告日期
    AnalyzeResult string    `gorm:"type:text"`       // 解析结果JSON
    AIResponse   string    `gorm:"type:text"`        // AI解读文本
    CreatedAt    time.Time
}

type ReportIndicator struct {
    Name        string    `gorm:"size:100"`          // 指标名称
    Value       string    `gorm:"size:50"`           // 指标值
    Unit        string    `gorm:"size:20"`          // 单位
    ReferenceMin float32                              // 参考值最小
    ReferenceMax float32                              // 参考值最大
    Status      string    `gorm:"size:10"`           // normal/high/low
}
```

### 3.9 订阅消息模块 (subscription)

**功能说明**: 身高录入提醒功能。

**API 接口**:
| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | /api/v1/subscriptions/reminder | 设置提醒 |
| GET | /api/v1/subscriptions/reminder | 获取提醒设置 |
| DELETE | /api/v1/subscriptions/reminder | 取消提醒 |

---

## 4. 核心算法

### 4.1 LMS百分位计算

采用 WHO/CDC 推荐的 LMS (Lambda-Mu-Sigma) 方法计算百分位。

**计算公式**:
```
百分位值 = M × (1 + L × S × Z)^(1/L)
Z分数 = ((实测值/M)^L - 1) / (L × S)
```

**状态判定**:
| Z分数范围 | 状态 | 颜色 | 建议 |
|----------|------|------|------|
| Z ≥ 2 | very_high | 橙色 | 关注性早熟可能 |
| 1 ≤ Z < 2 | high | 绿色 | 正常偏高 |
| -1 < Z < 1 | normal | 绿色 | 正常 |
| -2 ≤ Z ≤ -1 | low | 橙色 | 略低，继续观察 |
| Z < -2 | very_low | 红色 | 显著偏低，建议就医 |

### 4.2 靶身高计算 (Khamis-Roche)

根据父母身高预测孩子靶身高：
```
男孩靶身高 = 45.99 + 0.78 × (父亲身高 + 母亲身高) / 2
女孩靶身高 = 37.85 + 0.75 × (父亲身高 + 母亲身高) / 2
```

---

## 5. 数据库设计

### 5.1 ER 图

```
┌─────────┐       ┌─────────────┐       ┌──────────────┐
│  User   │──────│   Family    │──────│ FamilyMember │
└────┬────┘       └─────────────┘       └──────────────┘
     │                                          │
     │ 1:N                              1:N     │
     │                                    │
     ▼                                    ▼
┌─────────┐       ┌──────────────────────────────┐
│  Child  │──────│      GrowthRecord             │
└────┬────┘       └──────────────────────────────┘
     │
     │ 1:N
     ▼
┌─────────┐       ┌──────────────┐
│ Report  │       │   Hospital   │
└─────────┘       └──────────────┘

┌─────────────┐
│ Membership  │
└─────────────┘
```

### 5.2 数据表

| 表名 | 说明 |
|-----|------|
| users | 用户表 |
| children | 宝宝表 |
| growth_records | 生长记录表 |
| families | 家庭组表 |
| family_members | 家庭成员表 |
| hospitals | 医院表 |
| hospital_departments | 医院科室表 |
| memberships | 会员表 |
| usage_quotas | 额度使用表 |
| reports | 化验单报告表 |

---

## 6. API 规范

### 6.1 认证方式
- 所有接口（除登录外）需要 JWT Token
- Token 在 Header 中传递: `Authorization: Bearer <token>`

### 6.2 统一响应格式

**成功响应**:
```json
{
    "code": 0,
    "message": "success",
    "data": {}
}
```

**错误响应**:
```json
{
    "code": 40001,
    "message": "参数错误",
    "data": null
}
```

### 6.3 错误码规范

| 错误码 | 说明 |
|-------|------|
| 0 | 成功 |
| 40001 | 参数错误 |
| 40002 | 认证失败 |
| 40003 | 权限不足 |
| 40004 | 资源不存在 |
| 40005 | 配额不足 |
| 50001 | 服务器错误 |

---

## 7. 安全设计

### 7.1 合规红线

| 可以做 | 绝对不能做 |
|--------|-----------|
| 数据提取与结构化展示 | 给出任何诊断建议 |
| 显示参考范围对比 | "您可能患有XX疾病" |
| 标注偏高/偏低状态 | "建议服用XX药物" |
| 提供标准百分位计算 | 预测疾病风险或预后 |
| 就医方向性建议 | 推荐具体医生/药品/治疗方案 |

### 7.2 免责声明

每页底部固定显示：
```
⚠️ 免责声明：
本工具提供的所有数据均来源于用户上传的图片通过AI自动提取，
仅供参考和学习使用，不构成任何形式的医疗诊断、治疗建议或专业意见。
如有健康问题请务必及时咨询专业医疗机构及医师。
```

---

## 8. 部署架构

### 8.1 Docker 部署

```
┌────────────────────────────────────────────┐
│              Docker Compose                 │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐ │
│  │ Backend  │  │  MySQL   │  │  Redis   │ │
│  │  (Go)    │  │          │  │          │ │
│  └────┬─────┘  └──────────┘  └──────────┘ │
│       │                                      │
└───────┼──────────────────────────────────────┘
        │
        ▼
   Nginx (可选)
        │
        ▼
   微信小程序 / Web
```

### 8.2 环境变量

| 变量 | 说明 | 示例 |
|-----|------|------|
| PORT | 服务端口 | 8080 |
| JWT_SECRET | JWT密钥 | your-secret-key |
| WECHAT_APPID | 微信AppID | wx1234567890 |
| WECHAT_SECRET | 微信AppSecret | abcdef123456 |
| MYSQL_DSN | MySQL连接串 | user:pass@tcp(host)/db |
| REDIS_URL | Redis连接串 | redis://host:6379 |
| DEEPSEEK_API_KEY | DeepSeek API Key | sk-xxxx |

---

**文档结束**
