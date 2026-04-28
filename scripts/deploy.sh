#!/bin/bash

# =====================================================
# Growth Tracker Pro Backend Deployment Script
# =====================================================

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo_step() {
    echo -e "${GREEN}[STEP]${NC} $1"
}

echo_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

echo_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

echo_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    echo_step "检查环境要求..."

    if ! command -v docker &> /dev/null; then
        echo_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi

    if ! command -v docker-compose &> /dev/null; then
        echo_warning "docker-compose 未安装，部分功能可能不可用"
    fi

    echo_success "环境检查完成"
}

# Build Docker image
build_image() {
    echo_step "构建 Docker 镜像..."

    docker build -t growth-tracker-backend:latest \
        --build-arg BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ') \
        --build-arg VCS_REF=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown") \
        .

    echo_success "Docker 镜像构建完成"
}

# Run database migrations
run_migrations() {
    echo_step "运行数据库迁移..."

    # Create a temporary container to run migrations
    docker run --rm \
        --network growth-network \
        -e DB_HOST=db \
        -e DB_PORT=3306 \
        -e DB_USER=${DB_USER} \
        -e DB_PASSWORD=${DB_PASSWORD} \
        -e DB_NAME=${DB_NAME} \
        growth-tracker-backend:latest ./scripts/migrate

    echo_success "数据库迁移完成"
}

# Start services
start_services() {
    echo_step "启动服务..."

    docker-compose up -d

    # Wait for services to be ready
    echo_warning "等待服务启动..."
    sleep 10

    # Check service health
    if curl -f http://localhost:8080/health &> /dev/null; then
        echo_success "服务启动成功"
    else
        echo_error "服务启动失败，请检查日志"
        docker-compose logs backend
        exit 1
    fi
}

# Stop services
stop_services() {
    echo_step "停止服务..."
    docker-compose down
    echo_success "服务已停止"
}

# Show logs
show_logs() {
    docker-compose logs -f backend
}

# Show status
show_status() {
    docker-compose ps
    echo ""
    echo "服务健康状态："
    curl -s http://localhost:8080/health | jq . || echo "无法获取健康状态"
}

# Main menu
main() {
    case "${1:-deploy}" in
        deploy)
            check_prerequisites
            build_image
            start_services
            ;;
        start)
            docker-compose start
            ;;
        stop)
            stop_services
            ;;
        restart)
            stop_services
            start_services
            ;;
        logs)
            show_logs
            ;;
        status)
            show_status
            ;;
        build)
            build_image
            ;;
        migrate)
            run_migrations
            ;;
        clean)
            echo_step "清理 Docker 资源..."
            docker-compose down -v --rmi local
            echo_success "清理完成"
            ;;
        *)
            echo "用法: $0 {deploy|start|stop|restart|logs|status|build|migrate|clean}"
            echo ""
            echo "命令说明："
            echo "  deploy   - 部署并启动服务（默认）"
            echo "  start    - 启动已存在的服务"
            echo "  stop     - 停止服务"
            echo "  restart  - 重启服务"
            echo "  logs     - 查看服务日志"
            echo "  status    - 查看服务状态"
            echo "  build     - 构建 Docker 镜像"
            echo "  migrate   - 运行数据库迁移"
            echo "  clean     - 清理 Docker 资源"
            exit 1
            ;;
    esac
}

main "$@"
