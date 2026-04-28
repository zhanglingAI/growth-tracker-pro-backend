# Growth Tracker Pro - 腾讯云部署指南

## 环境要求

- **操作系统**: CentOS 7+ / Ubuntu 18+ / TencentOS Server
- **配置**: 2核4G内存起步
- **网络**: 公网IP，带宽 ≥ 5Mbps

---

## 一、服务器准备

### 1.1 登录腾讯云控制台

1. 进入 **云服务器 CVM** 控制台
2. 购买或选择一台服务器
3. 选择 **Ubuntu 20.04 LTS** 或 **CentOS 7** 镜像
4. 配置安全组：开放端口 `22`, `80`, `443`, `8080`

### 1.2 连接服务器

```bash
# 使用SSH连接
ssh root@你的服务器IP

# 如果使用密钥登录
ssh -i ~/your-key.pem root@你的服务器IP
```

---

## 二、Docker环境安装

### 2.1 Ubuntu 系统

```bash
# 更新系统
apt update && apt upgrade -y

# 安装依赖
apt install -y apt-transport-https ca-certificates curl gnupg lsb-release

# 添加Docker GPG密钥
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

# 添加Docker源
echo "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null

# 安装Docker
apt update
apt install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# 启动Docker
systemctl enable docker
systemctl start docker

# 验证安装
docker --version
docker compose version
```

### 2.2 CentOS 系统

```bash
# 安装依赖
yum install -y yum-utils device-mapper-persistent-data lvm2

# 添加Docker源
yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo

# 安装Docker
yum install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# 启动Docker
systemctl enable docker
systemctl start docker

# 验证安装
docker --version
```

---

## 三、部署应用

### 3.1 创建项目目录

```bash
# 创建项目目录
mkdir -p /opt/growth-tracker-pro
cd /opt/growth-tracker-pro

# 创建日志目录
mkdir -p logs
```

### 3.2 下载代码

**方式1: 从GitHub拉取**
```bash
# 克隆后端代码
git clone https://github.com/zhanglingAI/growth-tracker-pro-backend.git .

# 克隆前端代码(如有需要)
cd ..
git clone https://github.com/zhanglingAI/growth-tracker-pro-miniprogram.git miniprogram
```

**方式2: 使用SCP上传**
```bash
# 在本地执行，将项目上传到服务器
scp -r ./growth-tracker-pro-backend root@你的服务器IP:/opt/
```

### 3.3 配置环境变量

```bash
cd /opt/growth-tracker-pro

# 创建.env文件
cat > .env << 'EOF'
# 数据库配置
DB_HOST=mysql
DB_PORT=3306
DB_USER=growth_tracker
DB_PASSWORD=GrowthTracker2024!
DB_NAME=growth_tracker

# Redis配置
REDIS_HOST=redis
REDIS_PORT=6379

# JWT配置 (请修改为安全的随机字符串)
JWT_SECRET=your-super-secret-jwt-key-change-this-2024

# AI配置 (DeepSeek)
AI_PROVIDER=deepseek
AI_API_KEY=your_deepseek_api_key

# 微信小程序配置
WECHAT_APP_ID=your_wechat_appid
WECHAT_APP_SECRET=your_wechat_appsecret

# MySQL
MYSQL_ROOT_PASSWORD=GrowthTracker2024!
MYSQL_DATABASE=growth_tracker
EOF
```

### 3.4 启动服务

```bash
# 进入项目目录
cd /opt/growth-tracker-pro

# 使用Docker Compose启动所有服务
docker compose up -d

# 查看服务状态
docker compose ps

# 查看日志
docker compose logs -f
```

### 3.5 验证部署

```bash
# 健康检查
curl http://localhost:8080/health

# 预期输出
{"code":0,"msg":"success","data":{"status":"healthy"}}
```

---

## 四、配置Nginx反向代理 (可选)

### 4.1 安装Nginx

```bash
# Ubuntu
apt install -y nginx

# CentOS
yum install -y nginx
```

### 4.2 配置反向代理

```bash
# 复制Nginx配置
cp deploy/nginx.conf /etc/nginx/conf.d/growth-api.conf

# 测试配置
nginx -t

# 重启Nginx
systemctl enable nginx
systemctl restart nginx
```

### 4.3 配置SSL (可选)

```bash
# 安装Certbot
apt install -y certbot python3-certbot-nginx

# 获取SSL证书
certbot --nginx -d api.growthtracker.cn
```

---

## 五、配置微信小程序

### 5.1 小程序后台配置

1. 登录 [微信公众平台](https://mp.weixin.qq.com/)
2. 进入 **开发管理** → **开发设置**
3. 配置 **服务器域名**:
   - request合法域名: `https://你的域名.com`
   - uploadFile合法域名: `https://你的域名.com`

### 5.2 修改前端API地址

修改小程序 `app.js` 中的 `apiBaseUrl`:
```javascript
apiBaseUrl: 'https://api.growthtracker.cn/api/v1'
```

---

## 六、常用运维命令

### 服务管理

```bash
# 停止服务
docker compose down

# 重启服务
docker compose restart

# 重新构建
docker compose down && docker compose up -d --build

# 查看日志
docker compose logs -f --tail=100

# 进入容器
docker exec -it growth-tracker-api /bin/sh
```

### 备份数据库

```bash
# 备份
docker exec growth-tracker-mysql mysqldump -u root -pGrowthTracker2024! growth_tracker > backup_$(date +%Y%m%d).sql

# 恢复
docker exec -i growth-tracker-mysql mysql -u root -pGrowthTracker2024! growth_tracker < backup_file.sql
```

### 更新代码

```bash
cd /opt/growth-tracker-pro
git pull origin master
docker compose up -d --build
```

---

## 七、防火墙配置

### 腾讯云安全组

在腾讯云控制台安全组中添加规则:

| 协议 | 端口 | 来源 |
|------|------|------|
| TCP | 22 | 0.0.0.0/0 |
| TCP | 80 | 0.0.0.0/0 |
| TCP | 443 | 0.0.0.0/0 |
| TCP | 8080 | 0.0.0.0/0 |

### 服务器防火墙

```bash
# Ubuntu
ufw allow 22/tcp
ufw allow 80/tcp
ufw allow 443/tcp
ufw allow 8080/tcp
ufw enable

# CentOS
firewall-cmd --permanent --add-port=22/tcp
firewall-cmd --permanent --add-port=80/tcp
firewall-cmd --permanent --add-port=443/tcp
firewall-cmd --permanent --add-port=8080/tcp
firewall-cmd --reload
```

---

## 八、故障排查

### 常见问题

**1. 容器启动失败**
```bash
# 查看详细日志
docker compose logs api

# 检查端口占用
netstat -tlnp | grep 8080
```

**2. 数据库连接失败**
```bash
# 检查MySQL容器
docker exec -it growth-tracker-mysql mysql -u root -p

# 检查连接
docker exec growth-tracker-api ping mysql
```

**3. 前端无法访问API**
```bash
# 检查Nginx状态
systemctl status nginx

# 检查API是否正常
curl http://localhost:8080/health
```

---

## 九、监控与日志

### 查看实时日志

```bash
# 后端API日志
docker compose logs -f api

# MySQL日志
docker compose logs -f mysql

# 所有服务日志
docker compose logs -f
```

### 日志目录

- 后端日志: `/opt/growth-tracker-pro/logs/`
- Nginx日志: `/var/log/nginx/`
- Docker日志: `docker compose logs`

---

## 十、快速部署命令汇总

```bash
# 一键部署命令
# 1. 安装Docker
curl -fsSL https://get.docker.com | sh && systemctl enable docker

# 2. 下载代码
cd /opt
git clone https://github.com/zhanglingAI/growth-tracker-pro-backend.git
cd growth-tracker-pro-backend

# 3. 配置环境
cat > .env << 'EOF'
DB_HOST=mysql
DB_PORT=3306
DB_USER=growth_tracker
DB_PASSWORD=GrowthTracker2024!
DB_NAME=growth_tracker
REDIS_HOST=redis
REDIS_PORT=6379
JWT_SECRET=change-this-secret-key
AI_PROVIDER=deepseek
AI_API_KEY=your_api_key
MYSQL_ROOT_PASSWORD=GrowthTracker2024!
MYSQL_DATABASE=growth_tracker
EOF

# 4. 启动服务
docker compose up -d

# 5. 验证
curl http://localhost:8080/health
```

---

## 联系与支持

如有问题，请检查:
- Docker日志: `docker compose logs`
- GitHub Issues: https://github.com/zhanglingAI/growth-tracker-pro-backend/issues