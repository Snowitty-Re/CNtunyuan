# 助力团圆系统部署指南（前后端 + Nginx + WS）

本文档用于当前项目的**线上测试环境部署**，覆盖：
- Go 后端（`backend`）
- Next.js Web 管理端（`web`）
- Nginx 反向代理（含 `/api`、`/uploads`、`/swagger`、`/ws`）

> 适用场景：单机测试环境（推荐 Ubuntu 22.04+）

---

## 1. 环境准备

- Linux: Ubuntu 22.04+
- Go: 1.23+
- Node.js: 18+（建议 20 LTS）
- PostgreSQL: 14+
- Redis: 可选（不配置会自动降级内存缓存）
- Nginx: 1.20+

建议目录：

```bash
/opt/cntuanyuan/
  backend/
  web/
  uploads/
  logs/
```

---

## 2. 后端部署

### 2.1 拉取与编译

```bash
cd /opt/cntuanyuan
# git clone <repo-url> .

cd backend
go mod download
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /opt/cntuanyuan/backend/app cmd/app/main.go
```

### 2.2 配置文件

编辑：`/opt/cntuanyuan/backend/config/config.yaml`

至少检查以下配置：
- `server.mode: release`
- `server.domain: https://你的域名`
- `server.cors_origins: https://你的域名`
- `database.*`
- `jwt.secret`（>=32位强随机）
- `storage.local_path: /opt/cntuanyuan/uploads`
- `storage.base_url: https://你的域名/uploads`

### 2.3 数据库初始化（首次）

PostgreSQL：

```bash
createdb -U postgres -E UTF8 cntuanyuan
psql -U postgres -d cntuanyuan -f /opt/cntuanyuan/backend/migrations/postgres/00_bootstrap.sql
```

MySQL：

```bash
mysql -u root -p -e "CREATE DATABASE cntuanyuan CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
mysql -u root -p cntuanyuan < /opt/cntuanyuan/backend/migrations/mysql/00_bootstrap.sql
```

### 2.4 启动后端（你当前使用方式）

```bash
cd /opt/cntuanyuan/backend
./app
# 或 go run cmd/app/main.go
```

后端健康检查：

```bash
curl http://127.0.0.1:8080/api/v1/health
```

---

## 3. Web 部署

### 3.1 安装与构建

```bash
cd /opt/cntuanyuan/web
cp .env.example .env.local
```

编辑 `.env.local`：

```bash
NEXT_PUBLIC_API_BASE=https://你的域名/api/v1
```

继续：

```bash
npm ci
npm run build
npm start
```

本机检查：

```bash
curl -I http://127.0.0.1:3000
```

---

## 4. Nginx 配置（含 WS 转发）

推荐站点配置：`/etc/nginx/conf.d/你的域名.conf`

```nginx
map $http_upgrade $connection_upgrade {
    default upgrade;
    ''      close;
}

upstream cnty_backend {
    server 127.0.0.1:8080;
    keepalive 32;
}

upstream cnty_web {
    server 127.0.0.1:3000;
    keepalive 32;
}

server {
    listen 80;
    listen [::]:80;
    server_name yourdomain.com www.yourdomain.com;
    return 301 https://yourdomain.com$request_uri;
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name yourdomain.com www.yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;
    include /etc/letsencrypt/options-ssl-nginx.conf;
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;

    if ($host = 'www.yourdomain.com') {
        return 301 https://yourdomain.com$request_uri;
    }

    client_max_body_size 60m;

    # API
    location /api/ {
        proxy_pass http://cnty_backend/api/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-Proto https;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }

    # WebSocket
    location /ws/ {
        proxy_pass http://cnty_backend/ws/;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-Proto https;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
    }

    # 上传文件
    location /uploads/ {
        proxy_pass http://cnty_backend/uploads/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-Proto https;
    }

    # Swagger
    location /swagger/ {
        proxy_pass http://cnty_backend/swagger/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-Proto https;
    }

    # Web 前端
    location / {
        proxy_pass http://cnty_web;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-Proto https;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

加载配置：

```bash
sudo nginx -t
sudo systemctl reload nginx
```

---

## 5. WebSocket 测试方法

> 说明：若后端尚未实现 `/ws` 路由，测试会返回 404，这是“服务未提供 WS 端点”而非 Nginx 转发错误。

### 方法 A：`wscat`（推荐）

安装：

```bash
npm i -g wscat
```

连接：

```bash
wscat -c wss://你的域名/ws
```

若连接成功会进入交互模式；若失败会显示握手错误码。

### 方法 B：`curl` 检查握手响应

```bash
curl -i -N \
  -H "Connection: Upgrade" \
  -H "Upgrade: websocket" \
  -H "Sec-WebSocket-Version: 13" \
  -H "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==" \
  https://你的域名/ws
```

- `101 Switching Protocols`：WS 握手成功
- `404`：路径不存在（通常后端没实现该 WS 路由）
- `502`：Nginx 找不到后端服务

### 方法 C：本机强制验证 Nginx HTTPS 命中

```bash
curl -kI --resolve 你的域名:443:127.0.0.1 https://你的域名/
```

用于排查是否命中当前机器的 Nginx 配置。

---

## 6. 常见排障

### 6.1 访问域名返回 `{"code":404,"message":"route not found"}`

说明请求命中了后端根路径，而不是 Web。

排查：

```bash
sudo nginx -T | grep -n "server_name 你的域名"
sudo nginx -T | grep -n "location / {"
```

确认同一 HTTPS `server` 块中：
- `/api`、`/ws`、`/uploads`、`/swagger` 指向 8080
- `/` 指向 3000

### 6.2 API 正常但 Web 不正常

检查 3000 服务是否在线：

```bash
curl -I http://127.0.0.1:3000
```

若不通，重启 `npm start`。

### 6.3 上传失败

检查：
- `storage.local_path` 是否存在且可写
- Nginx `client_max_body_size` 是否足够

---

## 7. 建议的进程托管（后续）

你当前是手工启动。线上长期测试建议改为：
- `systemd`（稳定、开机自启、日志统一）
- 或 `pm2`（前端 Node 进程）

---

## 8. 最小上线检查清单

- [ ] `https://你的域名/api/v1/health` 返回正常
- [ ] `https://你的域名/` 打开 Web 页面（非后端 404 JSON）
- [ ] 登录可用、案件/任务/方言可创建
- [ ] 上传图片后可通过 `https://你的域名/uploads/...` 访问
- [ ] （如启用）`wss://你的域名/ws` 握手成功

