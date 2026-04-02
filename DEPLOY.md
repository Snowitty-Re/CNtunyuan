# 助力团圆系统 - 生产环境部署指南（无 Docker）

本文档描述当前仓库的推荐部署方式：**系统服务 + 反向代理（可选）**。

## 系统要求

- Linux（推荐 Ubuntu 22.04+）
- Go 1.23+
- PostgreSQL 14+ 或 MySQL 8+
- 至少 2GB 可用内存
- 至少 10GB 可用磁盘空间

## 快速部署

### 1. 获取代码

```bash
git clone <your-repo-url>
cd CNtunyuan/backend
```

### 2. 配置后端

```bash
cp config/config.example.yaml config/config.yaml
# 编辑 config/config.yaml
```

重点检查：
- `database.*`
- `jwt.secret`（至少 32 位）
- `wechat.*`（如启用微信登录）
- `storage.*`

### 3. 初始化数据库

PostgreSQL：
```bash
createdb -U postgres -E UTF8 cntuanyuan
psql -U postgres -d cntuanyuan -f migrations/postgres/00_bootstrap.sql
```

MySQL：
```bash
mysql -u root -p -e "CREATE DATABASE cntuanyuan CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
mysql -u root -p cntuanyuan < migrations/mysql/00_bootstrap.sql
```

### 4. 启动服务

开发/验证：
```bash
go run cmd/app/main.go
```

生产建议：
```bash
CGO_ENABLED=0 GOOS=linux go build -o /opt/cntuanyuan/app cmd/app/main.go
/opt/cntuanyuan/app
```

### 5. 验证

```bash
curl http://localhost:8080/api/v1/health
curl http://localhost:8080/api/v1/health/detailed
```

## systemd（推荐）

示例服务文件 `/etc/systemd/system/cntuanyuan.service`：

```ini
[Unit]
Description=CNtunyuan Backend
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/cntuanyuan
ExecStart=/opt/cntuanyuan/app
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

启用：
```bash
sudo systemctl daemon-reload
sudo systemctl enable cntuanyuan
sudo systemctl start cntuanyuan
sudo systemctl status cntuanyuan
```

## 备份与恢复

数据库备份（PostgreSQL）：
```bash
pg_dump -U postgres cntuanyuan > backup_$(date +%Y%m%d_%H%M%S).sql
```

数据库恢复（PostgreSQL）：
```bash
psql -U postgres -d cntuanyuan < backup_file.sql
```

上传目录备份（默认本地存储）：
```bash
tar czf uploads_backup_$(date +%Y%m%d_%H%M%S).tar.gz /path/to/uploads
```

## 安全检查清单

- [ ] `jwt.secret` 已设置强随机值（>=32）
- [ ] 数据库账号非超级管理员，且密码强度达标
- [ ] 上传目录权限最小化
- [ ] 开启日志轮转
- [ ] 已配置定期备份

## 常见问题

1. 服务启动失败：检查 `config/config.yaml` 与数据库连接
2. 登录失败：确认是否已执行 `00_bootstrap.sql`（默认管理员是否存在）
3. 上传失败：检查 `storage.local_path` 路径和读写权限
