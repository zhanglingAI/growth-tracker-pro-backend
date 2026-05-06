# Growth Tracker Pro 项目状态快照

> **重要**: 每次窗口重新启动后，请先读取此文件了解项目当前状态。
> 此文件由 Claude 在 2026-05-06 生成，记录截至该时间点的项目全貌。

---

## 1. 项目基本信息

- **项目名称**: Growth Tracker Pro Backend（儿童生长追踪管理小程序后端）
- **技术栈**: Go 1.21+ / Gin / GORM / MySQL 8.0 / Redis
- **工作目录**: `/root/growth-tracker-pro-backend`
- **当前分支**: `main`
- **最近提交**: `695f776` — "update: 按卫健委2025年3月最新标准更新生长数据"

---

## 2. Git 状态

### 未提交修改（7个文件）
这些文件有修改但尚未 commit，包含靶身高与环境评估模块的全部新增代码：

| 文件 | 状态 | 说明 |
|------|------|------|
| `CLAUDE.md` | M | 新增"只在Docker中编译"的项目准则 |
| `cmd/server/main.go` | M | `EnvironmentAssessment` 加入 AutoMigrate |
| `internal/handler/handler.go` | M | 新增5个API端点的路由和Handler方法 |
| `internal/models/dto.go` | M | 新增环境问卷/靶身高/生长速度/预警的DTO |
| `internal/models/models.go` | M | 新增 `HeightAlert`、`EnvironmentAssessment` 模型，扩展 `Child`/`GrowthRecord` |
| `internal/models/standards.go` | M | 新增计算函数（Khamis-Roche、环境得分、综合预测、生长速度等） |
| `internal/service/service.go` | M | 新增5个核心Service方法 + 辅助函数 |

### 新增未跟踪文件
- `internal/alert/engine.go` — 预警引擎核心逻辑
- `internal/alert/engine_test.go` — 预警引擎测试
- `API文档-靶身高与环境评估.md` — 前端接口文档（26KB）
- `TASKS.md` — 开发任务清单
- `.claude/` — Claude Code 配置目录
- `儿童靶身高精准评估：问卷设计、关键因子权重与综合算法框架.pdf` — 产品需求文档

---

## 3. 最近完成的工作

### 靶身高与环境评估模块（已完成）

**新增API端点（5个）**：
1. `POST /api/v1/children/:id/environment-assessment` — 提交环境问卷，返回得分+预测身高+行动计划
2. `GET /api/v1/children/:id/environment-assessment/latest` — 获取最新评估结果
3. `GET /api/v1/children/:id/environment-assessment/history` — 分页获取评估历史
4. `GET /api/v1/children/:id/target-height-comparison` — 靶身高综合分析（MPH/KR/环境预测）
5. `GET /api/v1/children/:id/growth-velocity` — 生长速度监测（支持 `?months_back=`）

**预警系统API（5个，此前已完成）**：
6. `POST /api/v1/children/:id/growth-stage` — 设置生长阶段
7. `GET /api/v1/children/:id/alerts` — 获取宝宝预警列表
8. `POST /api/v1/alerts/:alertId/read` — 标记预警已读
9. `POST /api/v1/alerts/:alertId/dismiss` — 忽略预警
10. `GET /api/v1/alerts/summary` — 获取预警摘要

**核心计算逻辑**（复用 `standards.go` 中的函数）：
- `CalculateEnvironmentScore` — 5模块计分（营养30%+睡眠25%+运动25%+健康10%+心理10%）
- `CalculateComprehensivePrediction` — Khamis-Roche + 环境增量（上限10cm）
- `CalculateKhamisRoche` / `CalculateQuantitativeGeneticsTargetHeight`
- `CalculateAnnualGrowthVelocity` / `EvaluateGrowthVelocityWithAlert`
- `Evaluate`（预警引擎）— 6维度预警检查

---

## 4. Docker 服务状态

| 服务 | 容器名 | 端口 | 状态 |
|------|--------|------|------|
| Backend API | growth-tracker-backend | 8080 | Up 2 days (healthy) |
| MySQL | growth-mysql | 3307→3306 | Up 5 days |
| Redis | growth-redis | 6380→6379 | Up 5 days |

**本地访问地址**: `http://localhost:8080`
**健康检查**: `curl http://localhost:8080/health` → `{"code":0,"msg":"success","data":{"status":"healthy","version":"1.0.0"}}`

---

## 5. 项目准则（来自 CLAUDE.md）

- **只在Docker中编译**：WSL本地缺少可用的Go工具链，以 `docker-compose build backend` 为编译成功的唯一标准。
- **所有API响应统一格式**: `BaseResponse { code, msg, data }`
- **Token格式**: `Authorization: Bearer jwt_token_{user_id}_{timestamp}`
- **日期格式**: "YYYY-MM-DD"

---

## 6. 关键文件位置

| 文件 | 路径 | 用途 |
|------|------|------|
| 接口文档 | `/root/growth-tracker-pro-backend/API文档-靶身高与环境评估.md` | 前端对接参考 |
| 任务清单 | `/root/growth-tracker-pro-backend/TASKS.md` | 开发任务跟踪 |
| 项目准则 | `/root/growth-tracker-pro-backend/CLAUDE.md` | Claude Code 工作指南 |
| 入口文件 | `/root/growth-tracker-pro-backend/cmd/server/main.go` | 服务启动、路由注册、AutoMigrate |
| Handler | `/root/growth-tracker-pro-backend/internal/handler/handler.go` | 所有HTTP端点 |
| Service | `/root/growth-tracker-pro-backend/internal/service/service.go` | 业务逻辑 |
| 模型 | `/root/growth-tracker-pro-backend/internal/models/models.go` | 数据库模型 |
| DTO | `/root/growth-tracker-pro-backend/internal/models/dto.go` | 请求/响应结构 |
| 计算标准 | `/root/growth-tracker-pro-backend/internal/models/standards.go` | 生长标准、百分位、预测算法 |
| 预警引擎 | `/root/growth-tracker-pro-backend/internal/alert/engine.go` | 预警评估逻辑 |
| Docker配置 | `/root/growth-tracker-pro-backend/docker-compose.yml` | 本地开发环境 |
| 原生部署脚本 | `/root/growth-tracker-pro-backend/deploy-native.sh` | 服务器原生部署 |
| Nginx配置 | `/root/growth-tracker-pro-backend/deploy/nginx.conf` | 生产环境反向代理 |
| 应用配置 | `/root/growth-tracker-pro-backend/config.yaml` | 数据库/Redis/JWT/AI配置 |

---

## 7. 已知问题与待办

### 已完成 ✅
- Docker编译通过
- 服务正常运行
- 全部API接口实现完毕
- 前端接口文档已生成

### 待处理 ⏳
1. **代码未提交**：当前有7个修改文件和4个新增文件尚未 `git commit`
2. **部署到生产服务器**：域名审核已通过，用户要求部署到服务器（需要服务器SSH信息）
3. **数据库迁移**：`EnvironmentAssessment` 和 `HeightAlert` 表已通过 AutoMigrate 创建，但旧数据不受影响
4. **AI配置**：`config.yaml` 中 `ai.api_key` 和 `wechat.app_id/app_secret` 为空（当前是mock实现）

---

## 8. 常用命令速查

```bash
# Docker 编译（唯一标准）
docker-compose build backend

# 重启服务
docker-compose up -d backend

# 查看日志
docker-compose logs -f backend

# 健康检查
curl http://localhost:8080/health

# 提交代码（当用户要求时）
git add -A
git commit -m "feat: 靶身高与环境评估模块"

# 查看Git状态
git status
git diff --stat
```

---

*最后更新: 2026-05-06*
*更新者: Claude*
