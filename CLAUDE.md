# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Growth Tracker Pro Backend** - 儿童生长追踪管理小程序后端服务

This is a Go-based backend service for a WeChat mini-program that helps parents track children's growth development, manage medical records, and get AI-powered growth recommendations.

## Tech Stack

- **Language**: Go 1.21+
- **Web Framework**: Gin
- **ORM**: GORM
- **Database**: MySQL 8.0+
- **Cache**: Redis
- **Authentication**: JWT

## Project Structure

```
growth-tracker-pro-backend/
├── cmd/server/          # Application entry point
│   └── main.go         # Server startup, DB initialization, routes
├── internal/            # Private application code
│   ├── config/         # Configuration management (config loading)
│   ├── models/         # Data models, DTOs, and standards
│   │   ├── models.go      # Core models: Child, GrowthRecord, Subscription
│   │   ├── user.go        # User model
│   │   ├── dto.go         # Request/Response DTOs
│   │   └── standards.go   # Growth standards and percentiles
│   ├── repository/     # Data access layer (DB operations)
│   ├── service/        # Business logic layer
│   │   └── service.go     # Implements Service interface with all business methods
│   ├── handler/        # HTTP handlers (Gin handlers)
│   │   └── handler.go     # All API endpoints and middleware
│   └── agent/          # AI Agent module (intelligent features)
│       ├── ai_agent.go        # Main GrowthAgent orchestrator
│       ├── child_profile.go   # Child profile builder (立体信息表)
│       ├── medical_guard.go   # Medical compliance checking
│       └── recommendation_engine.go  # Personalized recommendations
├── migrations/         # Database migration scripts
├── config.yaml         # Application configuration
└── go.mod             # Go modules
```

## Architecture Layers

**Flow**: `HTTP Request → handler → service → repository → database`

1. **Handler Layer** (`internal/handler/`): HTTP handlers, request binding/validation, middleware (CORS, auth), response formatting
2. **Service Layer** (`internal/service/`): Business logic, orchestrates operations between repository and agent
3. **Repository Layer** (`internal/repository/`): Database CRUD operations, query building
4. **AI Agent Layer** (`internal/agent/`): Intelligent features
   - **GrowthAgent**: Main orchestrator that coordinates all AI capabilities
   - **ProfileBuilder**: Builds comprehensive child growth profiles
   - **MedicalGuard**: Medical safety and compliance checks
   - **RecommendationEngine**: Generates personalized plans

## Key Features & API Endpoints

All endpoints are under `/api/v1/` prefix.

### Public Endpoints
- `GET /health` - Health check
- `POST /auth/login` - User login (WeChat code)
- `POST /pay/callback` - WeChat payment callback

### Protected Endpoints (requires JWT token)

**User Management**
- `GET /user/info` - Get user profile
- `PUT /user/info` - Update user profile

**Children Management**
- `GET /children` - List children
- `POST /children` - Create child profile
- `GET /children/:id` - Get child detail
- `PUT /children/:id` - Update child
- `DELETE /children/:id` - Delete child
- `POST /children/switch` - Switch active child

**Growth Records**
- `GET /records` - List records (with pagination)
- `POST /records` - Create record
- `PUT /records/:id` - Update record
- `DELETE /records/:id` - Delete record

**Family Sharing**
- `GET /family` - Get family info
- `POST /family` - Create family
- `POST /family/join` - Join family via invite code
- `DELETE /family/leave` - Leave family
- `PUT /family/members/:id/role` - Update member role
- `POST /family/inviteCode` - Generate invite code

**Subscription/Membership**
- `GET /subscription` - Get subscription status
- `POST /subscription/createOrder` - Create payment order

**AI Features**
- `POST /ai/chat` - AI chat assistant (growth questions)
- `POST /ai/parseReport` - Parse lab report (OCR + AI analysis)

**Home**
- `GET /home` - Get home dashboard data

## Common Commands

### Install Dependencies
```bash
go mod download
```

### Run Server
```bash
go run cmd/server/main.go
```

Server runs on `http://localhost:8080`

### Run Tests
```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/handler/
go test ./internal/agent/

# Run with verbose output
go test -v ./...
```

### Build Binary
```bash
go build -o growth-server cmd/server/main.go
```

## Configuration

Copy `config.yaml` and configure:
- Database connection (MySQL)
- Redis connection
- JWT secret
- AI provider API keys
- WeChat app credentials

## Important Conventions

### Response Format
All API responses use unified format:
```go
type BaseResponse struct {
    Code int         // 0 = success, non-zero = error
    Msg  string      // Error message or "success"
    Data interface{} // Response payload
}
```

Common error codes:
- `0` = Success
- `1` = Server error
- `2` = Invalid parameter
- `3` = Unauthorized
- `4` = Not found
- `5` = Quota exhausted
- `6` = Invalid invite code
- `7` = Family full

### Authentication Middleware
Token format: `Authorization: Bearer jwt_token_{user_id}_{timestamp}`

### Database Models
All models embed `BaseModel` with UUID primary key and timestamps. Use GORM auto-migrate on startup.

### Growth Calculations
- **Target Height**: Uses Khamis-Roche algorithm
  - Boys: `(father_height + mother_height + 13) / 2`
  - Girls: `(father_height + mother_height - 13) / 2`
  - Error range: ±8cm
- **Growth Percentiles**: Based on WHO child growth standards
- **Intervention Windows**: Boys 10-15 years, Girls 8-13 years

## AI Agent Architecture

The GrowthAgent is the core intelligent component:

1. **Input**: User question + child ID
2. **Profile Building**: Constructs comprehensive child profile from:
   - Basic info (age, gender, parent heights)
   - Growth records history
   - Growth percentiles and velocity
   - Nutrition and lifestyle factors
3. **Mode Detection**: Auto-detects user intent (profile analysis / recommendation / report)
4. **Response Generation**: Context-aware, personalized responses
5. **Medical Guard**: Sanitizes output for medical safety compliance
6. **Persistence**: Saves conversation history

## Testing Patterns

Tests use standard Go testing package with:
- `httptest` for HTTP handler tests
- Testify assertions (if used)
- Table-driven tests for edge cases

## Deployment

- Docker deployment available (see Dockerfile and docker-compose.yml)
- Native deployment scripts in deploy/ directory
- deploy-native.sh for direct server deployment

## Important Notes

1. **Simplified Implementation**: The current code has simplified implementations for:
   - JWT token parsing (no actual signature verification)
   - WeChat login and payment (mocked)
   - AI API calls (mock responses)
   These are placeholders - production implementations needed.

2. **Family Sharing Permissions**: Family members can access shared children's data. Check `buildChildProfile()` in ai_agent.go for permission logic.

3. **Medical Disclaimer**: The MedicalGuard component ensures AI responses don't provide medical diagnoses - only general guidance.

4. **ID Generation**: Uses UUID for all primary keys via `BeforeCreate` hook in BaseModel.

5. **Date Formatting**: Uses "YYYY-MM-DD" format for all date inputs.

---

## 当前状态 (2024-04-29)

### ✅ 已完成工作

**Docker部署配置：**
- Dockerfile 使用国内 Go proxy 加速构建
- docker-compose.yml 使用阿里云镜像加速器配置
- 端口映射调整：MySQL 3307:3306, Redis 6380:6379, Backend 8080:8080
- config.yaml 已配置使用 Docker服务名连接 (mysql, redis)

**Bug修复：**
1. **JWT Token解析bug** - 修复 `internal/handler/handler.go:796`
   - token格式: `jwt_token_{user_id}_{timestamp}`
   - 按 `_` 分割后 user_id 在 `parts[2]`，不是 `parts[0]`

2. **数据库迁移缺失表** - 修复 `cmd/server/main.go:121`
   - Subscription 表已加入 AutoMigrate

**文档：**
- 完整API接口文档 `API文档.md` 已生成（完全对照代码实际返回）
- Docker部署经验文档 `deploy-skill.md` 已创建
- 前端对接注意事项已明确标注

### 🚀 服务运行状态

| 服务 | 端口 | 状态 |
|------|------|------|
| MySQL | 3307 | 运行中 |
| Redis | 6380 | 运行中 |
| Backend API | 8080 | 运行中 |

健康检查接口正常：
```bash
curl http://localhost:8080/health
# 返回: {"code":0,"msg":"success","data":{"status":"healthy","version":"1.0.0"}}
```

### ⚠️ 已知前端对接问题

1. **HTTP状态码判断错误**
   - 创建类接口 (POST) 返回 HTTP 201 Created，不是 200
   - 前端必须判断 `code === 0` 作为成功依据，不是 HTTP statusCode

2. **/records 接口必须传 child_id**
   - 缺少参数返回 400 + code=2
   - 正确调用方式：`/api/v1/records?child_id=xxx`

3. **字段命名规范**
   - 用户相关字段是 `nick_name` / `avatar_url` (下划线)
   - 不是驼峰命名

### 📝 Git状态

当前分支：main
- 修改文件：Dockerfile, cmd/server/main.go, config.yaml, docker-compose.yml, internal/models/models.go, internal/service/service.go
- 新增文件：API文档.md, deploy-skill.md, CLAUDE.md
- 最近commit: feat: 添加Docker部署配置和技能文档

## Agent skills

### Issue tracker

Issues are tracked in GitHub Issues. See `docs/agents/issue-tracker.md`.

### Triage labels

Using default label vocabulary. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context repo. See `docs/agents/domain.md`.
