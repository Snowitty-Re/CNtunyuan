# 团圆寻亲系统 - 生产环境部署指南

## 系统要求

- Docker 20.10+
- Docker Compose 2.0+
- 至少 2GB 可用内存
- 至少 10GB 可用磁盘空间

## 快速开始

### 1. 克隆代码

```bash
git clone <your-repo-url>
cd CNtunyuan
```

### 2. 配置环境变量

```bash
cd docker
cp .env.example .env
# 编辑 .env 文件，填入实际配置
nano .env
```

**重要配置项：**

| 配置项 | 说明 | 必填 |
|--------|------|------|
| `DB_PASSWORD` | 数据库密码 | 是 |
| `JWT_SECRET` | JWT密钥（至少32位） | 是 |
| `WECHAT_APP_ID` | 微信小程序AppID | 微信登录时必填 |
| `WECHAT_APP_SECRET` | 微信小程序AppSecret | 微信登录时必填 |

### 3. 初始化数据库

```bash
# 启动数据库
docker-compose -f docker-compose.prod.yml up -d postgres

# 等待数据库启动完成（约10秒）
sleep 10

# 执行数据库迁移
docker-compose -f docker-compose.prod.yml exec postgres psql -U postgres -d cntuanyuan -f /docker-entrypoint-initdb.d/01_schema.sql
```

### 4. 启动服务

```bash
# 启动所有服务
docker-compose -f docker-compose.prod.yml up -d

# 查看服务状态
docker-compose -f docker-compose.prod.yml ps

# 查看日志
docker-compose -f docker-compose.prod.yml logs -f backend
```

### 5. 验证部署

```bash
# 健康检查
curl http://localhost:8080/api/v1/health

# 详细健康检查
curl http://localhost:8080/api/v1/health/detailed
```

## 数据库迁移

### 首次部署

```bash
# 进入后端目录
cd backend

# 检查数据库连接
go run cmd/app/main.go -check-db
```

### 数据备份

```bash
# 备份数据库
docker-compose -f docker-compose.prod.yml exec postgres pg_dump -U postgres cntuanyuan > backup_$(date +%Y%m%d_%H%M%S).sql

# 备份上传文件
docker-compose -f docker-compose.prod.yml exec backend tar czf /tmp/uploads_backup.tar.gz /app/uploads
docker cp cntuanyuan-backend:/tmp/uploads_backup.tar.gz ./
```

### 数据恢复

```bash
# 恢复数据库
docker-compose -f docker-compose.prod.yml exec -T postgres psql -U postgres -d cntuanyuan < backup_file.sql

# 恢复上传文件
docker cp uploads_backup.tar.gz cntuanyuan-backend:/tmp/
docker-compose -f docker-compose.prod.yml exec backend tar xzf /tmp/uploads_backup.tar.gz -C /
```

## 安全配置检查清单

在上线前，请确保完成以下检查：

- [ ] JWT 密钥已修改为随机生成的强密钥（至少32位）
- [ ] 数据库密码已修改为强密码
- [ ] 微信小程序 AppSecret 已配置正确
- [ ] 生产环境使用 release 模式运行
- [ ] 已配置 HTTPS（使用 Nginx 反向代理）
- [ ] 文件上传目录已设置正确的权限
- [ ] 已配置日志轮转防止磁盘占满
- [ ] 已配置数据库备份计划

## 常见问题

### 服务无法启动

```bash
# 检查日志
docker-compose -f docker-compose.prod.yml logs backend

# 检查配置验证
docker-compose -f docker-compose.prod.yml run --rm backend ./main -check-db
```

### 数据库连接失败

```bash
# 检查数据库服务状态
docker-compose -f docker-compose.prod.yml ps postgres

# 检查数据库日志
docker-compose -f docker-compose.prod.yml logs postgres
```

### 文件上传失败

```bash
# 检查上传目录权限
docker-compose -f docker-compose.prod.yml exec backend ls -la /app/uploads

# 检查磁盘空间
docker system df
```

## 更新部署

```bash
# 拉取最新代码
git pull

# 重新构建镜像
docker-compose -f docker-compose.prod.yml build

# 重启服务
docker-compose -f docker-compose.prod.yml up -d

# 清理旧镜像
docker image prune -f
```

## 监控和告警

系统内置 Prometheus 指标端点：

```
http://localhost:8080/api/v1/metrics
```

建议配置的告警项：

- 数据库连接池使用率 > 80%
- API 响应时间 > 500ms
- 错误率 > 1%
- 磁盘使用率 > 80%

## 常见问题修复

### Token 刷新与登录掉出问题

**症状**：修改密码、删除人员等操作后用户被踢出登录

**原因**：前端未实现 Token 自动刷新机制

**修复**：确保前端 `request.ts` 已实现：
- 401 响应时自动调用 `/auth/refresh` 接口
- 刷新成功后自动重试失败的请求
- 并发请求队列管理

### 数据展示问题

**症状**：案件编号、头像等数据不显示

**原因**：数据库缺少 `case_no` 字段和 `ty_missing_photos` 表

**修复**：执行数据库迁移补充字段和表

```sql
-- PostgreSQL 示例
ALTER TABLE ty_missing_persons ADD COLUMN case_no VARCHAR(50) UNIQUE;

CREATE TABLE IF NOT EXISTS ty_missing_photos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    missing_person_id UUID NOT NULL,
    url VARCHAR(500) NOT NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'normal',
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (missing_person_id) REFERENCES ty_missing_persons(id) ON DELETE CASCADE
);

-- 为现有数据生成 case_no
UPDATE ty_missing_persons 
SET case_no = 'CASE-' || TO_CHAR(created_at, 'YYYYMMDD') || '-' || SUBSTRING(id::text, 1, 4)
WHERE case_no IS NULL;
```

## 技术支持

如有问题，请检查：

1. 应用日志：`docker-compose logs backend`
2. 数据库日志：`docker-compose logs postgres`
3. 配置文件是否正确
4. 端口是否被占用

更多问题参考项目 Wiki 或提交 Issue
