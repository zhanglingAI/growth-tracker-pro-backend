#!/bin/bash
# Growth Tracker Pro - 直接部署脚本 (无需Docker)
# 从GitHub拉取代码，直接在服务器上运行

set -e

# ============================================
# 配置
# ============================================
PROJECT_NAME="growth-tracker-pro"
PROJECT_DIR="/opt/${PROJECT_NAME}"
LOG_DIR="/var/log/${PROJECT_NAME}"
CONFIG_FILE="${PROJECT_DIR}/config.yaml"
SERVICE_FILE="/etc/systemd/system/${PROJECT_NAME}.service"

# 颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_ok() { echo -e "${GREEN}[OK]${NC} $1"; }
log_err() { echo -e "${RED}[ERROR]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }

# ============================================
# 检查root
# ============================================
if [ "$EUID" -ne 0 ]; then
    log_err "请使用sudo或root权限运行"
    exit 1
fi

# ============================================
# 1. 安装依赖
# ============================================
install_deps() {
    log "安装系统依赖..."

    if command -v apt-get &> /dev/null; then
        apt update
        apt install -y golang-go mysql-server redis-server nginx certbot python3-certbot-nginx git wget curl
    elif command -v yum &> /dev/null; then
        yum install -y golang mysql-server redis nginx git wget curl
    fi

    log_ok "依赖安装完成"
}

# ============================================
# 2. 拉取代码
# ============================================
pull_code() {
    log "从GitHub拉取代码..."

    mkdir -p $PROJECT_DIR
    cd $PROJECT_DIR

    # 初始化git (如果是新目录)
    if [ ! -d ".git" ]; then
        git init
        git remote add origin https://github.com/zhanglingAI/growth-tracker-pro-backend.git
    fi

    git pull origin master --force
    log_ok "代码拉取完成"
}

# ============================================
# 3. 编译Go程序
# ============================================
build_app() {
    log "编译Go应用..."

    cd $PROJECT_DIR

    # 设置Go模块代理
    export GOPROXY=https://goproxy.cn,direct

    # 编译
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

    # 创建日志目录
    mkdir -p $LOG_DIR

    log_ok "编译完成"
}

# ============================================
# 4. 配置MySQL
# ============================================
setup_mysql() {
    log "配置MySQL..."

    # 启动MySQL
    if command -v systemctl &> /dev/null; then
        systemctl enable mysql
        systemctl start mysql
    elif command -v service &> /dev/null; then
        service mysql start
    fi

    # 等待MySQL启动
    sleep 5

    # 创建数据库和用户
    mysql -e "
    CREATE DATABASE IF NOT EXISTS growth_tracker CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
    CREATE USER IF NOT EXISTS 'growth'@'localhost' IDENTIFIED BY 'GrowthTracker2024!';
    GRANT ALL PRIVILEGES ON growth_tracker.* TO 'growth'@'localhost';
    FLUSH PRIVILEGES;
    "

    # 执行初始化SQL
    if [ -f "${PROJECT_DIR}/migrations/init.sql" ]; then
        mysql growth_tracker < ${PROJECT_DIR}/migrations/init.sql
    fi

    log_ok "MySQL配置完成"
}

# ============================================
# 5. 配置Redis
# ============================================
setup_redis() {
    log "配置Redis..."

    if command -v systemctl &> /dev/null; then
        systemctl enable redis
        systemctl start redis
    elif command -v service &> /dev/null; then
        service redis start
    fi

    log_ok "Redis配置完成"
}

# ============================================
# 6. 配置应用
# ============================================
setup_app_config() {
    log "配置应用..."

    # 创建配置文件
    cat > $CONFIG_FILE << 'EOF'
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
  secret: "growth-tracker-jwt-secret-change-in-production-2024"
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

    log_ok "应用配置完成"
}

# ============================================
# 7. 配置Systemd服务
# ============================================
setup_service() {
    log "配置Systemd服务..."

    cat > $SERVICE_FILE << EOF
[Unit]
Description=Growth Tracker Pro API Service
After=network.target mysql.service redis.service

[Service]
Type=simple
User=root
WorkingDirectory=${PROJECT_DIR}
ExecStart=${PROJECT_DIR}/server
Restart=always
RestartSec=5
StandardOutput=append:${LOG_DIR}/app.log
StandardError=append:${LOG_DIR}/error.log

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable ${PROJECT_NAME}

    log_ok "服务配置完成"
}

# ============================================
# 8. 配置Nginx
# ============================================
setup_nginx() {
    log "配置Nginx反向代理..."

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
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css application/json application/javascript;

    client_max_body_size 10M;

    location / {
        proxy_pass http://growth_backend;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    location /health {
        proxy_pass http://growth_backend/health;
        access_log off;
    }
}
EOF

    nginx -t && systemctl reload nginx
    systemctl enable nginx

    log_ok "Nginx配置完成"
}

# ============================================
# 9. 启动服务
# ============================================
start_service() {
    log "启动服务..."

    systemctl start ${PROJECT_NAME}

    # 等待启动
    sleep 3

    # 检查状态
    if systemctl is-active --quiet ${PROJECT_NAME}; then
        log_ok "服务启动成功!"
    else
        log_err "服务启动失败"
        journalctl -u ${PROJECT_NAME} --no-pager -n 20
        exit 1
    fi
}

# ============================================
# 10. 验证
# ============================================
verify() {
    log "验证部署..."

    # 检查API
    if curl -sf http://localhost:8080/health > /dev/null; then
        log_ok "API健康检查通过"
    else
        log_warn "API健康检查失败，请检查日志"
    fi

    # 检查Nginx
    if systemctl is-active --quiet nginx; then
        log_ok "Nginx运行正常"
    fi

    # 检查MySQL
    if systemctl is-active --quiet mysql; then
        log_ok "MySQL运行正常"
    fi

    # 检查Redis
    if systemctl is-active --quiet redis; then
        log_ok "Redis运行正常"
    fi
}

# ============================================
# 显示结果
# ============================================
show_result() {
    echo ""
    echo "=========================================="
    echo "     Growth Tracker Pro 部署完成!"
    echo "=========================================="
    echo ""
    echo "📍 访问地址:"
    echo "   API: http://服务器IP:8080"
    echo "   API: http://服务器IP/api/v1 (通过Nginx)"
    echo "   健康检查: http://服务器IP/health"
    echo ""
    echo "📁 目录:"
    echo "   项目: $PROJECT_DIR"
    echo "   日志: $LOG_DIR"
    echo "   配置: $CONFIG_FILE"
    echo ""
    echo "🔧 常用命令:"
    echo "   启动:   systemctl start ${PROJECT_NAME}"
    echo "   停止:   systemctl stop ${PROJECT_NAME}"
    echo "   重启:   systemctl restart ${PROJECT_NAME}"
    echo "   状态:   systemctl status ${PROJECT_NAME}"
    echo "   日志:   journalctl -u ${PROJECT_NAME} -f"
    echo ""
    echo "🔄 更新代码:"
    echo "   cd $PROJECT_DIR && git pull origin master"
    echo "   cd $PROJECT_DIR && go build -o server ./cmd/server"
    echo "   systemctl restart ${PROJECT_NAME}"
    echo ""
}

# ============================================
# 主流程
# ============================================
main() {
    echo ""
    echo "=========================================="
    echo "   Growth Tracker Pro 直接部署脚本"
    echo "=========================================="
    echo ""

    install_deps
    pull_code
    build_app
    setup_mysql
    setup_redis
    setup_app_config
    setup_service
    setup_nginx
    start_service
    verify
    show_result
}

main "$@"