# 团圆寻亲系统 - 后端

基于 Clean Architecture 的 Go 后端服务。

## 技术栈

- **Go 1.23+** - 编程语言（推荐 1.24+）
- **Gin** - Web 框架
- **GORM** - ORM 框架
- **PostgreSQL / MySQL** - 数据库（双数据库支持）
- **Redis** - 缓存（可选，自动降级内存缓存）
- **JWT** - 认证
- **Zap** - 日志
- **Swagger** - API 文档自动生成

## 项目结构

```
backend/
├── cmd/
│   ├── app/                # 主应用入口（HTTP 服务器）
│   ├── seed/               # 种子数据导入工具
│   └── resetpassword/      # 密码重置工具
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

### 2. 配置数据库

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
```

### 3. 数据库迁移

```bash
cd backend
go run cmd/app/main.go -migrate
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
- `POST /api/v1/auth/send-code` - 发送验证码
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
- `PUT    /api/v1/organizations/:id/move` - 移动组织（管理员）

### 走失人员 `/missing-persons`
- `GET    /api/v1/missing-persons` - 案件列表
- `GET    /api/v1/missing-persons/search` - 搜索
- `GET    /api/v1/missing-persons/stats` - 统计
- `POST   /api/v1/missing-persons` - 创建案件
- `GET    /api/v1/missing-persons/:id` - 案件详情
- `PUT    /api/v1/missing-persons/:id` - 更新案件
- `DELETE /api/v1/missing-persons/:id` - 删除案件（管理者）
- `PUT    /api/v1/missing-persons/:id/status` - 更新状态（管理者）
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
- `POST   /api/v1/tasks/:id/start` - 开始任务
- `POST   /api/v1/tasks/:id/complete` - 完成任务
- `POST   /api/v1/tasks/:id/cancel` - 取消任务（管理者）
- `PUT    /api/v1/tasks/:id/progress` - 更新进度

### 文件上传 `/upload`
- `POST   /api/v1/upload` - 单文件上传
- `POST   /api/v1/upload/batch` - 批量上传
- `GET    /api/v1/upload/:id` - 文件信息
- `GET    /api/v1/upload/:id/download` - 下载文件
- `DELETE /api/v1/upload/:id` - 删除文件
- `GET    /api/v1/upload/entity/:type/:id` - 实体关联文件
- `PUT    /api/v1/upload/:id/bind` - 绑定到实体
- `GET    /api/v1/upload/stats` - 文件统计（管理员）

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
- `GET /api/v1/metrics` - Prometheus 指标

## 开发指南

### 添加新模块

1. **创建领域实体** (`internal/domain/entity/`)
2. **定义仓储接口** (`internal/domain/repository/`)
3. **实现仓储** (`internal/infrastructure/repository/`)
4. **创建应用服务** (`internal/application/service/`)
5. **创建 HTTP 处理器** (`internal/interfaces/http/handler/`)
6. **注册路由** (`internal/interfaces/http/router/`)
7. **更新 DI 容器** (`internal/di/`)

### 常用命令

```bash
# 启动服务
go run cmd/app/main.go

# 数据库迁移
go run cmd/app/main.go -migrate

# 检查数据库
go run cmd/app/main.go -check-db

# 生成测试数据
go run cmd/seed/main.go -all

# 重置密码
go run cmd/resetpassword/main.go -phone=13800138000 -password=newpassword

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
```

## 部署

### 编译

```bash
CGO_ENABLED=0 GOOS=linux go build -o app cmd/app/main.go
```

### Docker

```bash
docker build -t cntuanyuan-backend .
docker run -p 8080:8080 cntuanyuan-backend
```

## 许可证

MIT License
