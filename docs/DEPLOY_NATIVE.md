# Growth Tracker Pro - 腾讯云直接部署指南

## 环境要求

- **操作系统**: Ubuntu 20.04+ / CentOS 7+
- **配置**: 2核2G内存起步
- **软件**: MySQL 8.0, Redis 6+, Go 1.21+

---

## 一、服务器准备

### 1.1 登录服务器

```bash
ssh root@你的服务器IP
```

### 1.2 安全组配置

在腾讯云控制台开放端口:
| 端口 | 用途 |
|------|------|
| 22 | SSH |
| 80 | HTTP |
| 443 | HTTPS |
| 3306 | MySQL (仅内网) |
| 6379 | Redis (仅内网) |
| 8080 | API (可选) |

---

## 二、快速部署 (一行命令)

```bash
# 下载部署脚本并执行
curl -fsSL https://raw.githubusercontent.com/zhanglingAI/growth-tracker-pro-backend/master/deploy-native.sh | bash
```

或手动执行:

```bash
# 1. SSH登录后，下载部署脚本
cd /opt
git clone https://github.com/zhanglingAI/growth-tracker-pro-backend.git
cd growth-tracker-pro-backend

# 2. 给脚本执行权限
chmod +x deploy-native.sh

# 3. 执行部署
./deploy-native.sh
```

---

## 三、详细部署步骤

### Step 1: 安装系统依赖

```bash
# Ubuntu
apt update && apt upgrade -y
apt install -y golang-go mysql-server redis-server nginx git wget curl

# CentOS
yum update -y
yum install -y golang mysql-server redis nginx git wget curl
```

### Step 2: 拉取代码

```bash
mkdir -p /opt/growth-tracker-pro
cd /opt/growth-tracker-pro
git init
git remote add origin https://github.com/zhanglingAI/growth-tracker-pro-backend.git
git pull origin master
```

### Step 3: 编译应用

```bash
cd /opt/growth-tracker-pro

# 设置Go代理
export GOPROXY=https://goproxy.cn,direct

# 编译
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server
```

### Step 4: 配置MySQL

```bash
# 启动MySQL
systemctl enable mysql
systemctl start mysql

# 创建数据库
mysql << 'EOF'
CREATE DATABASE IF NOT EXISTS growth_tracker CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'growth'@'localhost' IDENTIFIED BY 'GrowthTracker2024!';
GRANT ALL PRIVILEGES ON growth_tracker.* TO 'growth'@'localhost';
FLUSH PRIVILEGES;
EOF

# 导入初始数据
mysql growth_tracker < /opt/growth-tracker-pro/migrations/init.sql
```

### Step 5: 配置Redis

```bash
systemctl enable redis
systemctl start redis
```

### Step 6: 创建配置文件

```bash
cat > /opt/growth-tracker-pro/config.yaml << 'EOF'
server:
  host: "0.0.0.0"
  port: 8080
  mode: "release"

database:
  host: "localhost"
  port: 3306
  user: "growth"
  password: "GrowthTracker2024!"
  database: "growth_tracker"
  charset: "utf8mb4"

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

jwt:
  secret: "change-this-secret-key-in-production"
  expire_time: 604800

ai:
  provider: "deepseek"
  api_key: ""
  model: "deepseek-chat"
  base_url: "https://api.deepseek.com"
  max_tokens: 2000
  temperature: 0.7

wechat:
  app_id: ""
  app_secret: ""
EOF
```

### Step 7: 配置Systemd服务

```bash
cat > /etc/systemd/system/growth-tracker-pro.service << 'EOF'
[Unit]
Description=Growth Tracker Pro API
After=network.target mysql.service redis.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/growth-tracker-pro
ExecStart=/opt/growth-tracker-pro/server
Restart=always
RestartSec=5
StandardOutput=append:/var/log/growth-tracker-pro/app.log
StandardError=append:/var/log/growth-tracker-pro/error.log

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable growth-tracker-pro
```

### Step 8: 配置Nginx反向代理

```bash
cat > /etc/nginx/conf.d/growth-api.conf << 'EOF'
upstream growth_backend {
    server 127.0.0.1:8080;
    keepalive 64;
}

server {
    listen 80;
    server_name _;

    access_log /var/log/nginx/growth-api.access.log;
    error_log /var/log/nginx/growth-api.error.log;

    gzip on;
    gzip_types text/plain text/css application/json application/javascript;

    location / {
        proxy_pass http://growth_backend;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
EOF

nginx -t
systemctl enable nginx
systemctl reload nginx
```

### Step 9: 启动服务

```bash
# 创建日志目录
mkdir -p /var/log/growth-tracker-pro

# 启动服务
systemctl start growth-tracker-pro

# 检查状态
systemctl status growth-tracker-pro
```

---

## 四、验证部署

```bash
# 健康检查
curl http://localhost/health

# API测试
curl http://localhost/api/v1/home
```

预期输出:
```json
{"code":0,"msg":"success","data":{"status":"healthy"}}
```

---

## 五、配置SSL (可选)

```bash
# 安装Certbot
apt install -y certbot python3-certbot-nginx

# 获取证书
certbot --nginx -d your-domain.com

# 自动续期测试
certbot renew --dry-run
```

---

## 六、常用运维命令

| 操作 | 命令 |
|------|------|
| 启动服务 | `systemctl start growth-tracker-pro` |
| 停止服务 | `systemctl stop growth-tracker-pro` |
| 重启服务 | `systemctl restart growth-tracker-pro` |
| 查看状态 | `systemctl status growth-tracker-pro` |
| 查看日志 | `journalctl -u growth-tracker-pro -f` |
| 更新代码 | `cd /opt/growth-tracker-pro && git pull && go build -o server ./cmd/server && systemctl restart growth-tracker-pro` |

---

## 七、目录结构

```
/opt/growth-tracker-pro/
├── server              # 编译后的二进制文件
├── config.yaml         # 配置文件
├── migrations/         # 数据库脚本
├── internal/           # 源代码
└── docs/               # 文档

/var/log/growth-tracker-pro/
├── app.log             # 应用日志
└── error.log           # 错误日志
```

---

## 八、故障排查

### 服务启动失败

```bash
# 查看详细日志
journalctl -u growth-tracker-pro -n 50

# 手动运行测试
cd /opt/growth-tracker-pro
./server
```

### 数据库连接失败

```bash
# 检查MySQL状态
systemctl status mysql

# 测试连接
mysql -u growth -pGrowthTracker2024! growth_tracker
```

### 端口被占用

```bash
# 查看端口占用
netstat -tlnp | grep 8080

# 杀死占用进程
kill -9 <PID>
```