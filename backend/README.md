# 助力团圆系统 - 后端

基于 Clean Architecture 的 Go 后端服务。

## 技术栈

- **Go 1.23+** - 编程语言（推荐 1.24+）
- **Gin** - Web 框架
- **GORM** - ORM 框架
- **PostgreSQL / MySQL** - 数据库（双数据库支持）
- **Redis** - 缓存（可选，自动降级内存缓存）
- **JWT** - 认证（access + refresh 双令牌，Redis 黑名单）
- **Zap** - 日志
- **Swagger** - API 文档自动生成
- **Prometheus** - 监控指标

## 项目结构

```
backend/
├── cmd/
│   └── app/                # 主应用入口（HTTP 服务器）
├── internal/
│   ├── config/             # 配置管理
│   ├── domain/             # 领域层
│   │   ├── entity/         # 领域实体
│   │   ├── valueobject/    # 值对象
│   │   ├── repository/     # 仓储接口
│   │   └── service/        # 领域服务
│   ├── application/        # 应用层
│   │   ├── dto/            # 数据传输对象
│   │   └── service/        # 应用服务
│   ├── infrastructure/     # 基础设施层
│   │   ├── auth/           # JWT 认证
│   │   ├── cache/          # Redis / 内存缓存
│   │   ├── database/       # 数据库连接
│   │   ├── repository/     # 仓储实现
│   │   ├── storage/        # 文件存储（本地/OSS/COS）+ ClamAV 病毒扫描
│   │   ├── sms/            # 短信服务（阿里云/腾讯云）
│   │   └── wechat/         # 微信小程序客户端
│   ├── interfaces/         # 接口层
│   │   └── http/
│   │       ├── handler/    # HTTP 处理器
│   │       ├── middleware/  # 中间件（认证/审计/CORS/限流）
│   │       └── router/     # 路由
│   ├── di/                 # 依赖注入
│   └── task/               # 定时任务调度器
├── pkg/                    # 公共包
│   ├── logger/             # 日志
│   ├── errors/             # 统一错误处理
│   ├── response/           # HTTP 响应
│   ├── middleware/          # 通用中间件
│   ├── validator/          # 验证器
│   ├── metrics/            # Prometheus 监控
│   └── utils/              # 工具函数
├── migrations/             # 数据库迁移（PostgreSQL/MySQL）
├── docs/                   # Swagger 文档
├── config/                 # 配置文件
└── go.mod
```

## 快速开始

### 1. 安装依赖

```bash
cd backend
go mod download
```

### 2. 配置文件

编辑 `config/config.yaml`:

```yaml
database:
  type: "postgres"       # postgres 或 mysql
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "yourpassword"
  database: "cntuanyuan"
  ssl_mode: "disable"

jwt:
  secret: "your-secret-key-must-be-at-least-32-chars"   # 必须 ≥ 32 字符
  expire_time: 7200      # access token 有效期（秒）
```

> **安全要求**：`jwt.secret` 必须至少 32 个字符，否则服务启动失败。建议使用随机生成的强密钥（64+ 字符）。

### 3. 数据库初始化（首次）

```bash
cd backend
# PostgreSQL
psql -U postgres -d cntuanyuan -f migrations/postgres/00_bootstrap.sql

# MySQL
mysql -u root -p cntuanyuan < migrations/mysql/00_bootstrap.sql
```

### 3.1 旧环境增量升级（历史库）

仅当历史库未使用 `00_bootstrap.sql` 初始化时执行：

```bash
# PostgreSQL
psql -U postgres -d cntuanyuan -f migrations/postgres/06_schema_consistency_and_performance.sql
psql -U postgres -d cntuanyuan -f migrations/postgres/07_dialect_schema_alignment.sql

# MySQL
mysql -u root -p cntuanyuan < migrations/mysql/06_schema_consistency_and_performance.sql
mysql -u root -p cntuanyuan < migrations/mysql/07_dialect_schema_alignment.sql
```

### 4. 启动服务

```bash
cd backend
go run cmd/app/main.go
```

服务将在 `http://localhost:8080` 启动。

- API 文档: http://localhost:8080/swagger/index.html
- 健康检查: http://localhost:8080/api/v1/health

## API 端点

### 认证 `/auth`
- `POST /api/v1/auth/login` - 手机号密码登录
- `POST /api/v1/auth/admin-login` - 管理员登录
- `POST /api/v1/auth/wechat-login` - 微信小程序登录
- `POST /api/v1/auth/refresh` - 刷新令牌
- `POST /api/v1/auth/logout` - 登出
- `POST /api/v1/auth/bind-phone` - 绑定手机号
- `POST /api/v1/auth/send-code` - 发送验证码（60 秒内每号码限 1 次）
- `POST /api/v1/auth/reset-password` - 重置密码
- `GET  /api/v1/auth/me` - 获取当前用户

### 用户 `/users`
- `GET    /api/v1/users` - 用户列表
- `POST   /api/v1/users` - 创建用户（管理员）
- `GET    /api/v1/users/:id` - 用户详情
- `PUT    /api/v1/users/:id` - 更新用户（管理员）
- `DELETE /api/v1/users/:id` - 删除用户（管理员）
- `PUT    /api/v1/users/:id/status` - 更新用户状态（管理者）
- `PUT    /api/v1/users/:id/role` - 更新用户角色（管理员）

### 个人中心 `/profile`
- `GET /api/v1/profile` - 获取个人资料
- `PUT /api/v1/profile` - 更新个人资料
- `PUT /api/v1/profile/password` - 修改密码
- `GET /api/v1/profile/stats` - 个人统计

### 组织 `/organizations`
- `GET    /api/v1/organizations` - 组织列表
- `GET    /api/v1/organizations/tree` - 组织树
- `POST   /api/v1/organizations` - 创建组织（管理员）
- `GET    /api/v1/organizations/:id` - 组织详情
- `PUT    /api/v1/organizations/:id` - 更新组织（管理员）
- `DELETE /api/v1/organizations/:id` - 删除组织（管理员）
- `GET    /api/v1/organizations/:id/children` - 子组织
- `GET    /api/v1/organizations/:id/path` - 组织路径
- `PUT    /api/v1/organizations/:id/move` - 移动组织（管理员，不能移动到自身）

### 走失人员 `/missing-persons`
- `GET    /api/v1/missing-persons` - 案件列表
- `GET    /api/v1/missing-persons/search` - 搜索
- `GET    /api/v1/missing-persons/stats` - 统计
- `POST   /api/v1/missing-persons` - 创建案件
- `GET    /api/v1/missing-persons/:id` - 案件详情
- `PUT    /api/v1/missing-persons/:id` - 更新案件
- `DELETE /api/v1/missing-persons/:id` - 删除案件（管理者）
- `PUT    /api/v1/missing-persons/:id/status` - 更新状态（管理者，合法值：missing/searching/found/reunited/closed）
- `POST   /api/v1/missing-persons/:id/found` - 标记已找到（管理者）
- `POST   /api/v1/missing-persons/:id/reunited` - 标记已团圆（管理者）
- `GET    /api/v1/missing-persons/:id/tracks` - 获取轨迹
- `POST   /api/v1/missing-persons/:id/tracks` - 添加轨迹

### 方言 `/dialects`
- `GET    /api/v1/dialects` - 方言列表
- `GET    /api/v1/dialects/featured` - 精选方言
- `GET    /api/v1/dialects/stats` - 方言统计
- `POST   /api/v1/dialects` - 上传方言
- `GET    /api/v1/dialects/:id` - 方言详情
- `PUT    /api/v1/dialects/:id` - 更新方言
- `DELETE /api/v1/dialects/:id` - 删除方言
- `POST   /api/v1/dialects/:id/play` - 播放记录
- `POST   /api/v1/dialects/:id/like` - 点赞
- `DELETE /api/v1/dialects/:id/like` - 取消点赞
- `GET    /api/v1/dialects/:id/comments` - 获取评论
- `POST   /api/v1/dialects/:id/comments` - 添加评论
- `PUT    /api/v1/dialects/:id/status` - 更新状态（管理者）
- `POST   /api/v1/dialects/:id/feature` - 设为精选（管理员）
- `DELETE /api/v1/dialects/:id/feature` - 取消精选（管理员）

### 任务 `/tasks`
- `GET    /api/v1/tasks` - 任务列表
- `GET    /api/v1/tasks/my` - 我的任务
- `GET    /api/v1/tasks/pending` - 待处理任务
- `GET    /api/v1/tasks/overdue` - 逾期任务（管理者）
- `GET    /api/v1/tasks/stats` - 任务统计
- `POST   /api/v1/tasks` - 创建任务
- `GET    /api/v1/tasks/:id` - 任务详情
- `PUT    /api/v1/tasks/:id` - 更新任务
- `DELETE /api/v1/tasks/:id` - 删除任务（管理者）
- `GET    /api/v1/tasks/:id/logs` - 操作日志
- `POST   /api/v1/tasks/:id/assign` - 分配任务（管理者）
- `POST   /api/v1/tasks/:id/start` - 开始任务（仅被分配者）
- `POST   /api/v1/tasks/:id/complete` - 完成任务（仅被分配者）
- `POST   /api/v1/tasks/:id/cancel` - 取消任务（管理者）
- `PUT    /api/v1/tasks/:id/progress` - 更新进度

### 文件上传 `/upload`
- `POST   /api/v1/upload` - 单文件上传（最大 50MB）
- `POST   /api/v1/upload/batch` - 批量上传
- `GET    /api/v1/upload/stats` - 文件统计（管理员）
- `GET    /api/v1/upload/entity/:type/:id` - 实体关联文件
- `GET    /api/v1/upload/:id` - 文件信息
- `GET    /api/v1/upload/:id/download` - 下载文件
- `DELETE /api/v1/upload/:id` - 删除文件
- `PUT    /api/v1/upload/:id/bind` - 绑定到实体

### 仪表盘 `/dashboard`
- `GET /api/v1/dashboard/stats` - 统计数据
- `GET /api/v1/dashboard/overview` - 概览数据
- `GET /api/v1/dashboard/trend` - 趋势数据

### 审计日志 `/audit`
- `GET  /api/v1/audit/logs` - 日志列表（管理员）
- `GET  /api/v1/audit/logs/:id` - 日志详情（管理员）
- `GET  /api/v1/audit/stats` - 审计统计（管理员）
- `GET  /api/v1/audit/user-activity/:userId` - 用户活动（管理员）
- `GET  /api/v1/audit/module-stats` - 模块统计（管理员）
- `POST /api/v1/audit/cleanup` - 清理日志（超级管理员）

### 系统
- `GET /api/v1/health` - 健康检查
- `GET /api/v1/health/detailed` - 详细健康检查
- `GET /api/v1/metrics` - Prometheus 指标（**需要管理员权限**）

## 枚举值参考

| 字段 | 合法值 |
|------|--------|
| 用户角色 | `super_admin` / `admin` / `manager` / `volunteer` |
| 用户状态 | `active` / `inactive` / `banned` |
| 走失人员状态 | `missing` / `searching` / `found` / `reunited` / `closed` |
| 紧急程度 | `critical` / `high` / `medium` / `low` |
| 任务类型 | `search` / `verify` / `assist` / `follow` / `interview` / `other` |
| 任务状态 | `draft` → `pending` → `assigned` → `processing` → `completed` / `cancelled` |

## HTTP 状态码

| 状态码 | 说明 |
|--------|------|
| 200 | 请求成功 |
| 201 | 创建成功 |
| 204 | 删除成功（无内容） |
| 400 | 请求参数错误 |
| 401 | 未授权（Token 无效或过期） |
| 403 | 禁止访问（权限不足，或任务未分配给当前用户） |
| 404 | 资源不存在 |
| 409 | 资源冲突（如重复创建） |
| 500 | 服务器内部错误 |

## 安全特性

- **JWT**：access + refresh 双令牌；密钥强度校验（最少 32 字符）；登出后 token 加入 Redis 黑名单
- **密码**：bcrypt 哈希存储；密码至少 8 位
- **短信**：每手机号 60 秒内限发 1 条验证码，防刷
- **登录**：连续失败自动锁定账号（Redis 计数，配置阈值）
- **文件**：magic number 双重校验（扩展名 + MIME 嗅探）；ClamAV 病毒扫描；文件必须有扩展名
- **备份**：数据库密码通过环境变量传递（`PGPASSWORD` / `MYSQL_PWD`），不嵌入命令行
- **监控**：Prometheus `/metrics` 端点需要管理员认证
- **RBAC**：四级角色权限（super_admin > admin > manager > volunteer）
- **权限判定**：运行时采用 RBAC 角色层级；`ty_permissions/ty_user_permissions` 作为扩展字典保留
- **任务**：Start/Complete 操作仅允许被分配者本人执行
- **数据一致性**：`ty_files` 软删除统一为 `deleted_at`，不再使用 `is_deleted`
- **约束补齐**：任务与轨迹经纬度范围约束、审计日志 `user_id` 外键约束
- **索引优化**：users/missing_persons/tasks 增加高频复合索引

## 开发指南

### 添加新模块

1. **创建领域实体** (`internal/domain/entity/`)
2. **定义仓储接口** (`internal/domain/repository/`)
3. **实现仓储** (`internal/infrastructure/repository/`)
4. **创建应用服务** (`internal/application/service/`)
5. **创建 HTTP 处理器** (`internal/interfaces/http/handler/`)
6. **注册路由** (`internal/interfaces/http/router/`)
7. **更新 DI 容器** (`internal/di/wire_gen.go`)

### 常用命令

```bash
# 启动服务
go run cmd/app/main.go

# 检查数据库
go run cmd/app/main.go -check-db

# 生成 Swagger 文档
swag init -g cmd/app/main.go -o docs --parseDependency --parseInternal
```

## 测试

```bash
# 运行所有测试
go test ./...

# 运行指定包测试
go test ./internal/domain/...

# 带覆盖率
go test -cover ./...

# 带竞态检测
go test -race ./...
```

## 部署

### 编译

```bash
CGO_ENABLED=0 GOOS=linux go build -o app cmd/app/main.go
```

### 生产配置检查清单

- [ ] `jwt.secret` 至少 32 字符（推荐 64+）
- [ ] 数据库使用独立账号，非 root
- [ ] Redis 设置密码
- [ ] 配置文件不纳入版本控制（使用环境变量注入）
- [ ] 文件存储使用 OSS/COS，非本地磁盘
- [ ] 开启 ClamAV 病毒扫描（`storage.scan_virus: true`）
- [ ] 配置数据库定时备份路径（`backup.path`）

## 许可证

MIT License
