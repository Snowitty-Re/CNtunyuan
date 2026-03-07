# 生产环境关键问题修复指南

## 修复概述

本次修复解决了以下生产环境关键问题：

1. **登录掉出问题**：修改密码、删除人员等操作后用户被踢出登录
2. **Web 数据展示问题**：数据项不完整、显示不正确
3. **生产环境稳定性**：确保核心功能可用

---

## 1. 登录掉出问题修复

### 问题原因
- 前端请求拦截器未实现 Token 自动刷新机制
- 修改密码接口路径错误（调用了不存在的 `/auth/change-password`）

### 修复内容

#### 1.1 Web 前端 - 添加 Token 自动刷新机制
**文件**: `web-new/src/utils/request.ts`

- 添加 `getRefreshTokenFromStorage()` 函数从 localStorage 解析 refresh token
- 添加 `updateTokens()` 函数更新存储中的 token
- 添加 `isRefreshing` 和 `refreshSubscribers` 管理并发刷新
- 在 401 响应时自动调用 `/auth/refresh` 接口刷新 token
- 刷新成功后自动重试失败的请求

#### 1.2 小程序 - 添加 Token 自动刷新机制
**文件**: `mini-program/utils/request.js`

- 添加完整的 token 刷新逻辑
- 添加请求队列管理（刷新期间暂存请求）
- 刷新成功后重试队列中的请求
- 刷新失败时自动跳转登录页

#### 1.3 修改密码接口路径修复
**文件**: `web-new/src/pages/profile/index.tsx`

- 从 `/auth/change-password` 改为 `/profile/password`
- 方法从 `POST` 改为 `PUT`

#### 1.4 登录页面 API 地址修复
**文件**: `web-new/src/pages/login/index.tsx`

- 移除 localhost 回退，使用 `/api/v1`

---

## 2. 数据展示问题修复

### 问题原因
- 后端 MissingPerson 实体缺少 `case_no` 字段
- 后端返回 `photo_url` 字符串，前端期望 `photos` 数组
- 仓储层未预加载 Photos 关联

### 修复内容

#### 2.1 后端实体 - 添加 case_no 和 Photos
**文件**: `backend/internal/domain/entity/missing_person.go`

- 添加 `CaseNo` 字段（唯一索引）
- 添加 `Photos []MissingPhoto` 关联
- 添加 `MissingPhoto` 实体定义
- 添加 `generateCaseNo()` 函数生成案件编号（格式：CASE-YYYYMMDD-XXXX）
- 更新 `NewMissingPerson` 自动生成 case_no

#### 2.2 后端 DTO - 更新响应结构
**文件**: `backend/internal/application/dto/missing_person_dto.go`

- 添加 `MissingPersonPhoto` 响应类型
- `MissingPersonResponse` 添加 `CaseNo` 和 `Photos` 字段
- 更新 `ToMissingPersonResponse` 函数，添加 case_no 和 photos 转换逻辑

#### 2.3 后端仓储 - 预加载 Photos
**文件**: `backend/internal/infrastructure/repository/missing_person_repository.go`

- `List` 方法添加 `.Preload("Photos")`
- `FindByID` 方法添加 `.Preload("Photos")`

#### 2.4 数据库迁移 - 添加新表和字段
**文件**: 
- `backend/migrations/postgres/01_schema.sql`
- `backend/migrations/mysql/01_schema.sql`

**PostgreSQL 修改**:
- ty_missing_persons 表添加 `case_no VARCHAR(50) UNIQUE`
- 新增 ty_missing_photos 表
- 添加相关索引和触发器

**MySQL 修改**:
- ty_missing_persons 表添加 `case_no VARCHAR(50) UNIQUE`
- 新增 ty_missing_photos 表
- 添加相关索引

---

## 3. 部署步骤

### 3.1 更新数据库（PostgreSQL 示例）

```bash
# 连接到数据库
psql -U postgres -d cntuanyuan

# 执行以下 SQL 添加新字段和表

-- 添加 case_no 字段
ALTER TABLE ty_missing_persons ADD COLUMN case_no VARCHAR(50) UNIQUE;

-- 创建 ty_missing_photos 表
CREATE TABLE IF NOT EXISTS ty_missing_photos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    missing_person_id UUID NOT NULL,
    url VARCHAR(500) NOT NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'normal' CHECK (type IN ('normal', 'simulated', 'feature')),
    description TEXT,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    
    CONSTRAINT fk_mp_photos_missing_person FOREIGN KEY (missing_person_id) REFERENCES ty_missing_persons(id) ON DELETE CASCADE ON UPDATE CASCADE
);

-- 添加索引
CREATE INDEX idx_photos_missing_person ON ty_missing_photos(missing_person_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_photos_primary ON ty_missing_photos(missing_person_id, is_primary) WHERE is_primary = TRUE AND deleted_at IS NULL;

-- 创建触发器
CREATE TRIGGER update_missing_photos_updated_at BEFORE UPDATE ON ty_missing_photos
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 为现有数据生成 case_no
UPDATE ty_missing_persons 
SET case_no = 'CASE-' || TO_CHAR(created_at, 'YYYYMMDD') || '-' || SUBSTRING(id::text, 1, 4)
WHERE case_no IS NULL;
```

### 3.2 部署后端

```bash
cd backend

# 格式化代码
go fmt ./...

# 编译
go build -o cntuanyuan-api ./cmd/app/main.go

# 重启服务（根据你的部署方式）
systemctl restart cntuanyuan-api
# 或
# ./cntuanyuan-api &
```

### 3.3 部署 Web 前端

```bash
cd web-new

# 安装依赖
pnpm install

# 构建
pnpm build

# 部署 dist 目录到 Nginx
rsync -avz dist/ /var/www/cntuanyuan/web-new/
```

### 3.4 部署小程序

使用微信开发者工具上传代码。

---

## 4. 验证修复

### 4.1 验证 Token 刷新
1. 登录 Web 后台
2. 等待 Token 过期（或手动修改 token 为过期）
3. 执行任意操作（如刷新列表）
4. 观察是否自动刷新 token，而不是跳转到登录页

### 4.2 验证修改密码
1. 进入个人中心 -> 修改密码
2. 修改密码
3. 确认不会掉登录

### 4.3 验证案件编号显示
1. 进入案件列表
2. 确认案件编号显示正常（如 CASE-20260307-A1B2）
3. 确认头像可以显示（如果有上传照片）

---

## 5. 注意事项

1. **数据库迁移**: 务必先备份数据库再执行迁移
2. **Token 过期时间**: 检查 `config.yaml` 中的 JWT 配置
   - `expire_time`: 访问 token 过期时间（默认 604800 秒 = 7 天）
   - 刷新 token 过期时间为 2 倍 access token 过期时间
3. **Redis**: 确保 Redis 配置正确（用于存储 token 黑名单）

---

## 6. 回滚方案

如果出现问题，可以回滚：

```bash
# 后端回滚（使用上一个版本）
cd backend
git checkout HEAD~1
go build -o cntuanyuan-api ./cmd/app/main.go

# 数据库回滚（需要手动执行）
# 删除 case_no 字段和 ty_missing_photos 表
```

---

## 修复文件清单

### 后端
- `backend/internal/domain/entity/missing_person.go`
- `backend/internal/application/dto/missing_person_dto.go`
- `backend/internal/infrastructure/repository/missing_person_repository.go`
- `backend/migrations/postgres/01_schema.sql`
- `backend/migrations/mysql/01_schema.sql`

### Web 前端
- `web-new/src/utils/request.ts`
- `web-new/src/pages/profile/index.tsx`
- `web-new/src/pages/login/index.tsx`

### 小程序
- `mini-program/utils/request.js`
