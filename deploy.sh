#!/bin/bash
# Growth Tracker Pro - 一键部署脚本
# 适用于腾讯云服务器 (CentOS 7+ / Ubuntu 18+)

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置变量
PROJECT_NAME="growth-tracker-pro"
PROJECT_DIR="/opt/${PROJECT_NAME}"
BACKUP_DIR="/opt/backups"
LOG_FILE="/var/log/${PROJECT_NAME}-deploy.log"

# 输出带颜色的日志
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] [INFO] $1" >> $LOG_FILE
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] [SUCCESS] $1" >> $LOG_FILE
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] [ERROR] $1" >> $LOG_FILE
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] [WARNING] $1" >> $LOG_FILE
}

# 检查root权限
check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_error "请使用sudo或root权限运行此脚本"
        exit 1
    fi
}

# 检查系统环境
check_system() {
    log_info "检查系统环境..."

    # 检查Docker
    if command -v docker &> /dev/null; then
        DOCKER_VERSION=$(docker --version | grep -oP '\d+\.\d+\.\d+')
        log_success "Docker已安装: $DOCKER_VERSION"
    else
        log_error "Docker未安装，正在安装..."
        install_docker
    fi

    # 检查Docker Compose
    if command -v docker-compose &> /dev/null || docker compose version &> /dev/null; then
        log_success "Docker Compose已安装"
    else
        log_error "Docker Compose未安装"
        exit 1
    fi

    # 检查端口占用
    if lsof -i:8080 &> /dev/null; then
        log_warning "端口8080已被占用"
    fi
    if lsof -i:3306 &> /dev/null; then
        log_warning "端口3306已被占用(MySQL)"
    fi
    if lsof -i:6379 &> /dev/null; then
        log_warning "端口6379已被占用(Redis)"
    fi
}

# 安装Docker
install_docker() {
    log_info "安装Docker..."

    if [ -f /etc/os-release ]; then
        . /etc/os-release
        case $ID in
            ubuntu|debian)
                apt-get update
                apt-get install -y apt-transport-https ca-certificates curl gnupg lsb-release
                curl -fsSL https://download.docker.com/linux/$ID/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
                echo "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/$ID $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
                apt-get update
                apt-get install -y docker-ce docker-ce-cli containerd.io
                ;;
            centos|rhel)
                yum install -y yum-utils
                yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
                yum install -y docker-ce docker-ce-cli containerd.io
                ;;
        esac
    fi

    systemctl enable docker
    systemctl start docker
    log_success "Docker安装完成"
}

# 创建项目目录
create_dirs() {
    log_info "创建项目目录..."

    mkdir -p $PROJECT_DIR
    mkdir -p $BACKUP_DIR
    mkdir -p /var/log/${PROJECT_NAME}
    mkdir -p $PROJECT_DIR/logs

    log_success "目录创建完成"
}

# 备份函数
backup() {
    log_info "备份现有数据..."

    TIMESTAMP=$(date +%Y%m%d_%H%M%S)

    if [ -d "$PROJECT_DIR" ]; then
        mkdir -p $BACKUP_DIR
        tar -czf $BACKUP_DIR/${PROJECT_NAME}_backup_$TIMESTAMP.tar.gz -C /opt ${PROJECT_NAME} 2>/dev/null || true
        log_success "备份已保存: $BACKUP_DIR/${PROJECT_NAME}_backup_$TIMESTAMP.tar.gz"
    fi
}

# 拉取最新代码
pull_code() {
    log_info "拉取最新代码..."

    cd $PROJECT_DIR

    # 如果是git仓库
    if [ -d ".git" ]; then
        git pull origin master
        log_success "代码已更新"
    else
        log_warning "不是git仓库，跳过拉取"
    fi
}

# 构建并启动
deploy() {
    log_info "开始构建和部署..."

    cd $PROJECT_DIR

    # 停止旧容器
    log_info "停止旧容器..."
    docker-compose down 2>/dev/null || true

    # 重新构建并启动
    log_info "构建Docker镜像..."
    docker-compose build --no-cache

    log_info "启动服务..."
    docker-compose up -d

    # 等待服务启动
    log_info "等待服务启动..."
    sleep 10

    # 检查健康状态
    if curl -f http://localhost:8080/health &> /dev/null; then
        log_success "服务启动成功!"
    else
        log_error "服务启动失败，请检查日志"
        docker-compose logs
        exit 1
    fi
}

# 配置防火墙
configure_firewall() {
    log_info "配置防火墙..."

    # CentOS/RHEL
    if command -v firewall-cmd &> /dev/null; then
        firewall-cmd --permanent --add-port=8080/tcp
        firewall-cmd --permanent --add-port=80/tcp
        firewall-cmd --reload
        log_success "防火墙规则已添加"
    fi

    # Ubuntu/Debian (ufw)
    if command -v ufw &> /dev/null; then
        ufw allow 8080/tcp
        ufw allow 80/tcp
        ufw reload
        log_success "防火墙规则已添加"
    fi
}

# 初始化数据库
init_database() {
    log_info "检查数据库初始化..."

    # 等待MySQL就绪
    for i in {1..30}; do
        if docker exec growth-tracker-mysql mysqladmin ping -h localhost --silent 2>/dev/null; then
            log_success "MySQL已就绪"
            break
        fi
        log_info "等待MySQL启动... ($i/30)"
        sleep 2
    done
}

# 显示服务状态
show_status() {
    echo ""
    echo "========================================="
    echo "     Growth Tracker Pro 部署完成"
    echo "========================================="
    echo ""
    echo "服务状态:"
    docker-compose ps
    echo ""
    echo "访问地址:"
    echo "  API: http://你的服务器IP:8080"
    echo "  健康检查: http://你的服务器IP:8080/health"
    echo ""
    echo "常用命令:"
    echo "  查看日志: docker-compose logs -f"
    echo "  重启服务: docker-compose restart"
    echo "  停止服务: docker-compose down"
    echo ""
    echo "配置文件位置: $PROJECT_DIR/.env"
    echo "日志文件位置: /var/log/${PROJECT_NAME}-deploy.log"
    echo ""
}

# 主函数
main() {
    echo ""
    echo "========================================="
    echo "   Growth Tracker Pro 一键部署脚本"
    echo "========================================="
    echo ""

    check_root
    check_system
    create_dirs

    # 询问是否备份
    read -p "是否备份现有数据? (y/n): " backup_choice
    if [ "$backup_choice" = "y" ] || [ "$backup_choice" = "Y" ]; then
        backup
    fi

    # 拉取代码或使用当前目录代码
    read -p "是否从Git拉取最新代码? (y/n): " pull_choice
    if [ "$pull_choice" = "y" ] || [ "$pull_choice" = "Y" ]; then
        pull_code
    fi

    deploy
    configure_firewall
    init_database
    show_status
}

# 运行
main "$@"