#!/bin/bash

# 团圆寻亲系统 - 生产部署脚本
# 使用方式: ./deploy.sh [dev|prod]

set -e

ENV=${1:-prod}
APP_NAME="cntuanyuan"
APP_DIR="/var/www/cntuanyuan"
NGINX_CONF="/etc/nginx/sites-available/cntuanyuan.conf"

echo "🚀 开始部署 - 环境: $ENV"

# 检查 root 权限
if [ "$EUID" -ne 0 ]; then 
    echo "❌ 请使用 sudo 运行"
    exit 1
fi

# 创建应用目录
echo "📁 创建应用目录..."
mkdir -p $APP_DIR
mkdir -p /var/log/nginx
mkdir -p /var/www/certbot

# 安装依赖
echo "📦 安装系统依赖..."
if ! command -v nginx &> /dev/null; then
    apt update
    apt install -y nginx certbot python3-certbot-nginx
fi

# 部署后端
echo "🔧 部署后端服务..."
if [ -f "backend/cntuanyuan" ]; then
    cp backend/cntuanyuan $APP_DIR/
    chmod +x $APP_DIR/cntuanyuan
else
    echo "⚠️ 后端可执行文件不存在，请先编译: cd backend && go build -o cntuanyuan cmd/app/main.go"
    exit 1
fi

# 复制配置文件
if [ -f "backend/config/config.yaml" ]; then
    cp backend/config/config.yaml $APP_DIR/
    echo "⚠️ 请修改 $APP_DIR/config.yaml 中的配置（数据库密码、JWT密钥等）"
fi

# 部署前端
echo "🎨 部署前端..."
if [ -d "web-new/dist" ]; then
    mkdir -p $APP_DIR/web-new
    cp -r web-new/dist $APP_DIR/web-new/
else
    echo "⚠️ 前端构建文件不存在，请先构建: cd web-new && pnpm build"
    exit 1
fi

# 配置 Nginx
echo "🌐 配置 Nginx..."
if [ -f "docker/nginx/cntuanyuan.conf" ]; then
    cp docker/nginx/cntuanyuan.conf $NGINX_CONF
    
    # 启用站点
    if [ ! -L "/etc/nginx/sites-enabled/cntuanyuan.conf" ]; then
        ln -s $NGINX_CONF /etc/nginx/sites-enabled/
    fi
    
    # 测试配置
    nginx -t
else
    echo "❌ Nginx 配置文件不存在"
    exit 1
fi

# 配置 systemd 服务
echo "⚙️ 配置系统服务..."
if [ -f "docker/nginx/cntuanyuan.service" ]; then
    cp docker/nginx/cntuanyuan.service /etc/systemd/system/
    systemctl daemon-reload
    systemctl enable cntuanyuan
fi

# 设置权限
echo "🔒 设置文件权限..."
chown -R www-data:www-data $APP_DIR
chmod -R 755 $APP_DIR

# 启动服务
echo "🏃 启动服务..."
systemctl restart cntuanyuan
systemctl restart nginx

# 检查状态
echo ""
echo "📊 服务状态:"
systemctl status cntuanyuan --no-pager || true
echo ""
systemctl status nginx --no-pager || true

echo ""
echo "✅ 部署完成！"
echo ""
echo "📋 后续步骤:"
echo "1. 修改配置文件: $APP_DIR/config.yaml"
echo "2. 配置 SSL 证书: certbot --nginx -d your-domain.com"
echo "3. 修改 Nginx 配置: $NGINX_CONF"
echo "4. 访问网站: http://your-domain.com"
echo ""
echo "📚 查看日志:"
echo "  后端: journalctl -u cntuanyuan -f"
echo "  Nginx: tail -f /var/log/nginx/cntuanyuan_error.log"
