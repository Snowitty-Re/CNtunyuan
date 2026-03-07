# Nginx 生产部署指南

## 部署架构

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   客户端     │────▶│   Nginx     │────▶│  Go 后端    │
│  (浏览器/小程序)│     │  (反向代理)  │     │  (:8080)   │
└─────────────┘     └──────┬──────┘     └─────────────┘
                           │
                           ▼
                    ┌─────────────┐
                    │  前端静态资源 │
                    │ (web-new/dist)│
                    └─────────────┘
```

## 快速部署步骤

### 1. 准备服务器

```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装 Nginx
sudo apt install nginx -y

# 安装 Certbot（用于 SSL 证书）
sudo apt install certbot python3-certbot-nginx -y
```

### 2. 部署后端服务

```bash
# 1. 上传后端代码到服务器
# 2. 编译后端
cd /path/to/backend
go build -o cntuanyuan cmd/app/main.go

# 3. 配置环境变量或修改 config/config.yaml
# 4. 启动后端服务（可以使用 systemd 或 supervisor）
./cntuanyuan

# 或使用 systemd 服务（推荐）
sudo systemctl start cntuanyuan
```

### 3. 构建前端

```bash
# 1. 在本地或服务器构建前端
cd web-new

# 2. 修改 .env.production 中的 API 地址
# VITE_API_BASE_URL=/api/v1

# 3. 构建
pnpm build

# 4. 上传构建结果到服务器
rsync -avz dist/ user@vps:/var/www/cntuanyuan/web-new/
```

### 4. 配置 Nginx

```bash
# 1. 复制配置文件
sudo cp docker/nginx/cntuanyuan.conf /etc/nginx/sites-available/

# 2. 修改配置（替换 your-domain.com 为你的域名）
sudo nano /etc/nginx/sites-available/cntuanyuan.conf

# 3. 启用站点
sudo ln -s /etc/nginx/sites-available/cntuanyuan.conf /etc/nginx/sites-enabled/

# 4. 测试配置
sudo nginx -t

# 5. 重启 Nginx
sudo systemctl restart nginx
```

### 5. 配置 SSL 证书（Let's Encrypt）

```bash
# 自动获取并配置证书
sudo certbot --nginx -d your-domain.com -d www.your-domain.com

# 自动续期测试
sudo certbot renew --dry-run
```

## 配置文件说明

### 关键配置项

| 配置项 | 说明 | 示例 |
|--------|------|------|
| `server_name` | 你的域名 | `cntuanyuan.org` |
| `upstream backend_api` | 后端服务地址 | `127.0.0.1:8080` |
| `root` | 前端静态资源目录 | `/var/www/cntuanyuan/web-new/dist` |
| `ssl_certificate` | SSL 证书路径 | `/etc/letsencrypt/live/...` |

### 跨域配置（已处理）

Nginx 反向代理已经处理了跨域问题，前端请求 `/api/xxx` 会被代理到后端，不需要额外的 CORS 配置。

如果小程序需要独立域名访问 API，请配置微信小程序域名白名单。

## 后端服务配置

### config/config.yaml

```yaml
server:
  port: "8080"
  mode: "release"  # 生产环境使用 release 模式
  domain: "https://your-domain.com"  # 修改为实际域名

database:
  type: "postgres"  # 或 mysql
  host: "localhost"
  port: 5432
  user: "cntuanyuan"
  password: "your-strong-password"
  database: "cntuanyuan"
  ssl_mode: "disable"

# 微信小程序配置（必须正确配置）
wechat:
  app_id: "your-wechat-app-id"
  app_secret: "your-wechat-app-secret"
  enable_login: true

# JWT 配置（生产环境必须修改密钥）
jwt:
  secret: "your-random-secret-key-at-least-32-characters"
  expire_time: 604800
```

## 小程序配置

### 1. 配置 request 合法域名

登录 [微信小程序后台](https://mp.weixin.qq.com) → 开发 → 开发设置 → 服务器域名：

- **request 合法域名**: `https://your-domain.com`
- **uploadFile 合法域名**: `https://your-domain.com`
- **downloadFile 合法域名**: `https://your-domain.com`

### 2. 修改小程序配置

```javascript
// mini-program/app.js
App({
  globalData: {
    // 生产环境 API 地址
    apiBaseUrl: 'https://your-domain.com/api/v1'
  }
})
```

## 常见问题

### 1. 前端刷新 404

确保 Nginx 配置中有：
```nginx
location / {
    try_files $uri $uri/ /index.html;
}
```

### 2. 文件上传失败

检查 Nginx 上传大小限制：
```nginx
client_max_body_size 100M;
```

### 3. WebSocket 连接失败

确保 Nginx 配置了 WebSocket 支持：
```nginx
proxy_set_header Upgrade $http_upgrade;
proxy_set_header Connection "upgrade";
```

### 4. 跨域错误

如果使用独立域名访问 API，需要配置 CORS：
```nginx
location /api/ {
    add_header 'Access-Control-Allow-Origin' 'https://your-domain.com';
    add_header 'Access-Control-Allow-Methods' 'GET, POST, PUT, DELETE, OPTIONS';
    add_header 'Access-Control-Allow-Headers' 'DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Authorization';
    
    if ($request_method = 'OPTIONS') {
        return 204;
    }
    
    proxy_pass http://backend_api/;
}
```

## 监控和日志

### 查看 Nginx 日志

```bash
# 访问日志
sudo tail -f /var/log/nginx/cntuanyuan_access.log

# 错误日志
sudo tail -f /var/log/nginx/cntuanyuan_error.log
```

### 查看后端日志

```bash
# 如果使用 systemd
sudo journalctl -u cntuanyuan -f
```

## 性能优化

### 启用 Gzip 压缩

在 `/etc/nginx/nginx.conf` 中添加：

```nginx
gzip on;
gzip_vary on;
gzip_proxied any;
gzip_comp_level 6;
gzip_types text/plain text/css text/xml application/json application/javascript application/rss+xml application/atom+xml image/svg+xml;
```

### 静态资源缓存

已在配置中启用 30 天缓存。

## 安全建议

1. **修改 JWT 密钥** - 生产环境必须使用强随机密钥
2. **修改数据库密码** - 使用强密码
3. **配置防火墙** - 只开放 80/443 端口
4. **定期更新证书** - Certbot 会自动续期
5. **启用 HTTPS** - 强制使用 HTTPS 访问
6. **设置日志轮转** - 避免日志文件过大

## 更新部署

```bash
# 1. 更新代码
cd /var/www/cntuanyuan
git pull

# 2. 重启后端
sudo systemctl restart cntuanyuan

# 3. 重新构建前端（如需要）
cd web-new && pnpm build

# 4. 重载 Nginx
sudo nginx -s reload
```
