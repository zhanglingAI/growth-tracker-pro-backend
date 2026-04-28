# Growth Tracker Pro 后端服务

基于 Go 语言开发的儿童生长追踪管理小程序后端服务。

## 功能特性

- 宝宝管理：创建、编辑、删除、切换宝宝档案
- 生长记录：身高体重记录、查看历史、趋势分析
- 靶身高预测：基于父母身高的遗传潜力预测（Khamis-Roche算法）
- 干预窗口计算：计算最佳干预时间窗口
- AI 对话助手：基于 DeepSeek API 的智能问答
- 化验单解析：OCR + AI 智能解析
- 家庭共享：多成员管理、邀请码机制
- 会员订阅：微信支付集成

## 技术栈

- **语言**: Go 1.21+
- **框架**: Gin Web Framework
- **ORM**: GORM
- **数据库**: MySQL 8.0+
- **缓存**: Redis
- **认证**: JWT

## 项目结构

```
growth-tracker-pro-backend/
├── cmd/server/          # 应用入口
├── internal/
│   ├── config/          # 配置管理
│   ├── models/          # 数据模型
│   ├── repository/      # 数据访问层
│   ├── service/         # 业务逻辑层
│   └── handler/         # HTTP处理层
├── pkg/
│   ├── utils/           # 工具函数
│   └── response/        # 统一响应
├── migrations/          # 数据库迁移脚本
├── docs/                # API文档
└── config.yaml          # 配置文件
```

## 快速开始

### 1. 环境要求

- Go 1.21+
- MySQL 8.0+
- Redis 6.0+

### 2. 配置

```bash
cp config.yaml.example config.yaml
# 编辑 config.yaml 填入数据库和Redis配置
```

### 3. 初始化数据库

```bash
mysql -u root -p < migrations/init.sql
```

### 4. 安装依赖

```bash
go mod download
```

### 5. 运行服务

```bash
go run cmd/server/main.go
```

服务将在 `http://localhost:8080` 启动。

## 配置说明

### config.yaml

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "debug"

database:
  host: "localhost"
  port: 3306
  user: "root"
  password: "password"
  database: "growth_tracker"
  charset: "utf8mb4"

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

jwt:
  secret: "your-secret-key"
  expire_time: 604800  # 7天

ai:
  provider: "deepseek"
  api_key: "your-api-key"
  model: "deepseek-chat"
  base_url: "https://api.deepseek.com"
  max_tokens: 2000
  temperature: 0.7

wechat:
  app_id: "your-app-id"
  app_secret: "your-app-secret"
```

## 数据库表结构

### users 用户表

| 字段 | 类型 | 说明 |
|-----|------|------|
| id | VARCHAR(36) | 主键 |
| open_id | VARCHAR(128) | 微信OpenID |
| nick_name | VARCHAR(64) | 昵称 |
| avatar_url | VARCHAR(512) | 头像 |
| phone | VARCHAR(20) | 手机号 |
| family_id | VARCHAR(36) | 家庭ID |

### children 宝宝表

| 字段 | 类型 | 说明 |
|-----|------|------|
| id | VARCHAR(36) | 主键 |
| user_id | VARCHAR(36) | 所属用户 |
| name | VARCHAR(64) | 宝宝姓名 |
| gender | VARCHAR(10) | 性别 |
| birthday | DATE | 出生日期 |
| father_height | DECIMAL(5,1) | 父亲身高 |
| mother_height | DECIMAL(5,1) | 母亲身高 |

### growth_records 生长记录表

| 字段 | 类型 | 说明 |
|-----|------|------|
| id | VARCHAR(36) | 主键 |
| child_id | VARCHAR(36) | 宝宝ID |
| height | DECIMAL(5,1) | 身高(cm) |
| weight | DECIMAL(5,1) | 体重(kg) |
| date | DATE | 测量日期 |
| age_str | VARCHAR(20) | 年龄字符串 |
| age_in_days | INT | 年龄(天) |

### subscriptions 订阅表

| 字段 | 类型 | 说明 |
|-----|------|------|
| id | VARCHAR(36) | 主键 |
| user_id | VARCHAR(36) | 用户ID |
| plan | VARCHAR(20) | 订阅方案 |
| start_date | DATE | 开始日期 |
| end_date | DATE | 到期日期 |
| ai_quota | INT | AI额度 |
| ai_used | INT | 已使用 |

### families 家庭表

| 字段 | 类型 | 说明 |
|-----|------|------|
| family_id | VARCHAR(36) | 家庭ID |
| name | VARCHAR(64) | 家庭名称 |
| invite_code | VARCHAR(20) | 邀请码 |
| max_members | INT | 最大成员数 |

## 算法说明

### 靶身高计算 (Khamis-Roche 简化版)

```
男孩: (父亲身高 + 母亲身高 + 13) / 2
女孩: (父亲身高 + 母亲身高 - 13) / 2
误差范围: ±8cm
```

### 干预窗口

```
男孩: 10-15岁
女孩: 8-13岁
```

### 百分位计算

基于 WHO 儿童生长标准，简化计算模型。

## License

MIT
