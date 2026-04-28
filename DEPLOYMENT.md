# Growth Tracker Pro 后端部署指南

## 部署方式

### 方式一：Docker Compose 部署（推荐）

```bash
# 1. 复制环境变量文件
cp .env.example .env
# 编辑 .env 填入实际配置

# 2. 启动所有服务
docker-compose up -d

# 3. 查看服务状态
docker-compose ps

# 4. 查看日志
docker-compose logs -f backend
```

### 方式二：Docker 单独部署

```bash
# 1. 构建镜像
docker build -t growth-tracker-backend:latest .

# 2. 运行容器
docker run -d \
  --name growth-tracker-backend \
  -p 8080:8080 \
  -e DB_HOST=your_mysql_host \
  -e DB_PORT=3306 \
  -e DB_USER=your_user \
  -e DB_PASSWORD=your_password \
  -e DB_NAME=growth_tracker \
  -e REDIS_HOST=your_redis_host \
  -e REDIS_PORT=6379 \
  -e JWT_SECRET=your_jwt_secret \
  growth-tracker-backend:latest
```

### 方式三：直接运行二进制文件

```bash
# 1. 确保 MySQL 和 Redis 服务可用

# 2. 编辑 config.yaml 配置数据库连接

# 3. 运行
./growth-tracker-server
```

## 环境变量说明

| 变量 | 说明 | 示例 |
|------|------|------|
| DB_HOST | MySQL 主机地址 | localhost |
| DB_PORT | MySQL 端口 | 3306 |
| DB_USER | 数据库用户名 | growth_tracker |
| DB_PASSWORD | 数据库密码 | your_password |
| DB_NAME | 数据库名 | growth_tracker |
| REDIS_HOST | Redis 主机地址 | localhost |
| REDIS_PORT | Redis 端口 | 6379 |
| JWT_SECRET | JWT 密钥 | your_jwt_secret |
| AI_PROVIDER | AI 服务提供商 | openai |
| AI_API_KEY | AI API 密钥 | your_api_key |

## 服务验证

```bash
# 健康检查
curl http://localhost:8080/health

# API 文档
curl http://localhost:8080/swagger/index.html
```

## 常用命令

```bash
# 使用部署脚本
./scripts/deploy.sh deploy    # 部署
./scripts/deploy.sh stop       # 停止
./scripts/deploy.sh restart    # 重启
./scripts/deploy.sh logs       # 查看日志
./scripts/deploy.sh status     # 查看状态

# Docker Compose
docker-compose up -d           # 启动
docker-compose down            # 停止
docker-compose logs -f         # 日志
docker-compose restart         # 重启
```

## 目录结构

```
growth-tracker-pro-backend/
├── Dockerfile              # Docker 镜像构建文件
├── docker-compose.yml       # Docker Compose 配置
├── config.yaml            # 应用配置文件
├── .env.example           # 环境变量示例
├── scripts/
│   └── deploy.sh         # 部署脚本
└── growth-tracker-server  # 编译后的二进制文件
```

## 端口说明

| 端口 | 服务 |
|------|------|
| 8080 | 后端 API 服务 |
| 3306 | MySQL 数据库（可选） |
| 6379 | Redis 缓存（可选） |

## 故障排查

```bash
# 查看容器状态
docker ps -a

# 查看容器日志
docker logs growth-tracker-backend

# 进入容器调试
docker exec -it growth-tracker-backend sh

# 检查端口占用
netstat -tlnp | grep 8080
```
