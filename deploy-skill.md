# Growth Tracker Pro Docker 部署技能

## 本次部署遇到的问题与解决方案

### 1. 镜像拉取超时/失败
**问题**: Docker Hub访问慢或超时
**解决方案**:
- 使用阿里云镜像加速器：`https://4pppnuwu.mirror.aliyuncs.com`
- 添加多个备用镜像源：网易、中科大
- 使用国内镜像前缀：`m.daocloud.io/docker.io/library/`

**最佳实践**:
```json
// /etc/docker/daemon.json
{
  "registry-mirrors": [
    "https://4pppnuwu.mirror.aliyuncs.com",
    "https://hub-mirror.c.163.com",
    "https://docker.mirrors.ustc.edu.cn"
  ]
}
```

### 2. 端口冲突
**问题**: 本地MySQL(3306)和Redis(6379)已在运行
**解决方案**:
- MySQL映射端口：`3307:3306`
- Redis映射端口：`6380:6379`
- 容器内部仍使用默认端口通信

### 3. 配置文件未更新
**问题**: 容器内config.yaml还是旧的IP配置
**解决方案**:
- 使用Docker服务名而非IP：`mysql`, `redis`
- 修改config.yaml后必须重新构建镜像

### 4. docker-compose版本兼容问题
**问题**: `KeyError: 'ContainerConfig'`
**解决方案**:
- 删除旧容器再重新创建：`docker-compose rm -f backend && docker-compose up -d`

---

## 标准部署流程

### 步骤1: 配置Docker加速（仅首次）
```bash
# 检查现有配置
cat /etc/docker/daemon.json

# 如果没有配置，执行以下命令
cat > /etc/docker/daemon.json << 'EOF'
{
  "registry-mirrors": [
    "https://4pppnuwu.mirror.aliyuncs.com",
    "https://hub-mirror.c.163.com",
    "https://docker.mirrors.ustc.edu.cn"
  ]
}
EOF

# 重启Docker
systemctl restart docker
```

### 步骤2: 检查本地镜像
```bash
# 查看已有镜像
docker images | grep -E "mysql|redis"

# 如果本地有DaoCloud镜像，直接使用
# m.daocloud.io/docker.io/library/mysql:8.0
# m.daocloud.io/docker.io/library/redis:7-alpine
```

### 步骤3: 验证配置文件
**config.yaml**:
```yaml
database:
  host: "mysql"      # 使用Docker服务名
  port: 3306         # 容器内部端口

redis:
  host: "redis"      # 使用Docker服务名
  port: 6379         # 容器内部端口
```

**docker-compose.yml**:
```yaml
services:
  mysql:
    image: m.daocloud.io/docker.io/library/mysql:8.0
    ports:
      - "3307:3306"  # 外部:内部

  redis:
    image: m.daocloud.io/docker.io/library/redis:7-alpine
    ports:
      - "6380:6379"  # 外部:内部
```

### 步骤4: 构建与启动
```bash
# 1. 清理旧容器（如有问题）
docker-compose rm -f

# 2. 构建镜像（--no-cache确保配置文件更新）
docker-compose build --no-cache

# 3. 启动服务
docker-compose up -d

# 4. 等待并检查状态
sleep 10
docker-compose ps

# 5. 测试健康检查
curl http://localhost:8080/health

# 6. 查看日志排错
docker logs growth-tracker-backend --tail 50
```

### 步骤5: 验证
```bash
# 检查所有服务状态
docker-compose ps

# 预期输出：
# Name          State        Ports
# growth-mysql  Up           0.0.0.0:3307->3306/tcp
# growth-redis  Up           0.0.0.0:6380->6379/tcp
# backend       Up           0.0.0.0:8080->8080/tcp

# 测试API
curl http://localhost:8080/health
# 期望返回: {"code":0,"msg":"success","data":{"status":"healthy","version":"1.0.0"}}
```

---

## 常见问题排查

### 问题: 后端不断重启
检查:
```bash
# 查看后端日志
docker logs growth-tracker-backend --tail 30

# 常见原因:
# 1. 数据库连接失败 - 检查config.yaml的host是否为"mysql"
# 2. 数据库未就绪 - MySQL启动需要时间，会自动重试
# 3. 配置文件未更新 - 需要重新build镜像
```

### 问题: 镜像拉取失败
解决:
```bash
# 更换镜像源
# DaoCloud: m.daocloud.io/docker.io/library/
# 阿里云: registry.cn-hangzhou.aliyuncs.com/acs/
# 或直接使用官方镜像，依赖daemon.json加速器
```

### 问题: 端口占用
解决:
```bash
# 查找占用端口的进程
netstat -tlnp | grep -E "3306|6379|8080"

# 修改docker-compose.yml映射端口
# MySQL: 3307, 3308...
# Redis: 6380, 6381...
```

### 问题: 配置修改后不生效
解决:
```bash
# 必须重新构建+删除旧容器
docker-compose build --no-cache backend
docker-compose rm -f backend
docker-compose up -d backend
```

---

## Dockerfile优化点

```dockerfile
# 使用国内Go代理
ENV GOPROXY=https://goproxy.cn,direct

# 分阶段构建 - builder阶段编译, runtime阶段运行
# 确保COPY . . 在go mod download之后，利用缓存
```

---

## 下次部署检查清单
- [ ] Docker镜像加速器已配置
- [ ] config.yaml使用服务名(mysql/redis)而非127.0.0.1
- [ ] docker-compose.yml端口映射避免冲突
- [ ] 使用 `--no-cache` 重新构建确保配置更新
- [ ] 先启动mysql/redis，再启动backend（depends_on自动处理）
- [ ] 部署后验证健康检查接口
- [ ] 查看日志确认无错误
