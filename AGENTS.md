# CNtunyuan - 开发指南

本文档为 AI 助手和开发者提供项目背景信息和开发规范。

## 项目背景

团圆寻亲志愿者系统是一个帮助寻找走失人员的公益项目，通过整合志愿者网络、方言语音数据库和工作流系统，提高寻人效率。

### 核心价值
- **志愿者协作**: 组织架构化的志愿者管理
- **方言辅助**: 通过方言语音帮助确认走失人员身份
- **任务驱动**: OA工作流确保寻人任务有序进行

## 技术架构

### 后端架构 (Clean Architecture)
```
┌─────────────────────────────────────────────┐
│            Interfaces Layer                  │
│   (HTTP Handlers, Middleware, Routes)       │
├─────────────────────────────────────────────┤
│           Application Layer                  │
│   (Use Cases, Application Services, DTO)    │
├─────────────────────────────────────────────┤
│            Domain Layer                      │
│   (Entities, Value Objects, Repository      │
│    Interfaces, Domain Services)             │
├─────────────────────────────────────────────┤
│         Infrastructure Layer                 │
│   (DB, Cache, External APIs, Repository     │
│    Implementations)                         │
└─────────────────────────────────────────────┘
```

依赖关系：**向内指向领域层**，Domain 层不依赖任何其他层。

## 开发规范

### Go 后端代码规范
- 使用 `gofmt` 格式化代码
- 遵循 Go 官方代码规范
- 函数命名使用驼峰式
- 接口命名使用动词+名词，如 `CreateUser`
- 错误处理必须返回具体错误信息

### 后端项目结构
```
backend/
├── cmd/                      # 应用程序入口
│   └── app/                 # HTTP 服务器（统一入口）
│
├── internal/                # 私有应用代码
│   ├── domain/              # 领域层
│   │   ├── entity/          # 领域实体
│   │   ├── valueobject/     # 值对象
│   │   ├── repository/      # 仓储接口
│   │   └── service/         # 领域服务
│   │
│   ├── application/         # 应用层
│   │   ├── dto/             # 数据传输对象
│   │   └── service/         # 应用服务
│   │
│   ├── infrastructure/      # 基础设施层
│   │   ├── database/        # 数据库
│   │   ├── cache/           # 缓存 (Redis)
│   │   ├── repository/      # 仓储实现
│   │   ├── auth/            # JWT 认证
│   │   ├── storage/         # 文件存储 (本地/OSS/COS)
│   │   ├── sms/             # 短信服务 (阿里云/腾讯云)
│   │   └── wechat/          # 微信小程序客户端
│   │
│   ├── interfaces/          # 接口适配层
│   │   └── http/
│   │       ├── handler/     # HTTP 处理器
│   │       ├── middleware/   # HTTP 中间件
│   │       └── router/      # 路由
│   │
│   ├── di/                  # 依赖注入
│   ├── config/              # 配置
│   └── task/                # 定时任务调度器
│
├── pkg/                     # 公共库
│   ├── logger/              # 日志 (Zap)
│   ├── errors/              # 统一错误处理
│   ├── response/            # HTTP 响应
│   ├── middleware/           # 通用中间件
│   ├── validator/           # 请求验证
│   ├── metrics/             # Prometheus 监控
│   └── utils/               # 工具函数
│
├── migrations/              # 数据库迁移 (PostgreSQL/MySQL)
├── docs/                    # Swagger 文档
└── config/                  # 配置文件
```

### 数据库规范

#### 表命名
- 使用前缀 `ty_`
- 复数形式，如 `ty_users`, `ty_organizations`
- 关联表使用 `_` 连接，如 `ty_user_permissions`

#### 字段命名
- 使用下划线命名法
- 常用字段: `created_at`, `updated_at`, `deleted_at`
- 外键使用 `_id` 后缀
- 布尔值建议使用 `is_` 前缀（历史兼容除外）

### API 设计规范

#### RESTful API
```
GET    /api/v1/resources      # 列表
POST   /api/v1/resources      # 创建
GET    /api/v1/resources/:id   # 详情
PUT    /api/v1/resources/:id   # 更新
DELETE /api/v1/resources/:id   # 删除
```

#### 响应格式
```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

## 常用命令

### 后端命令

```bash
cd backend

# 启动服务
go run cmd/app/main.go

# 检查数据库
go run cmd/app/main.go -check-db

# 首次初始化数据库（PostgreSQL）
psql -U postgres -d cntuanyuan -f migrations/postgres/00_bootstrap.sql

# 运行测试
go test ./...

# 生成 Swagger 文档
swag init -g cmd/app/main.go -o docs --parseDependency --parseInternal
```

### 数据库迁移

**PostgreSQL:**
```bash
createdb -U postgres -E UTF8 cntuanyuan
psql -U postgres -d cntuanyuan -f backend/migrations/postgres/00_bootstrap.sql
```

**MySQL:**
```bash
mysql -u root -p -e "CREATE DATABASE cntuanyuan CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
mysql -u root -p cntuanyuan < backend/migrations/mysql/00_bootstrap.sql
```

历史库增量升级（若未执行最新结构对齐）：

```bash
# PostgreSQL
psql -U postgres -d cntuanyuan -f backend/migrations/postgres/06_schema_consistency_and_performance.sql

# MySQL
mysql -u root -p cntuanyuan < backend/migrations/mysql/06_schema_consistency_and_performance.sql
```

## 配置说明 (config/config.yaml)

```yaml
server:
  port: "8080"
  mode: "debug"  # debug/release

database:
  type: "postgres"   # postgres 或 mysql
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "postgres"
  database: "cntuanyuan"

redis:
  host: ""       # 空表示不使用 Redis（自动降级内存缓存）

jwt:
  secret: "your-secret-key"  # 生产环境必须修改
  expire_time: 604800        # 7天
  refresh_time: 2592000      # 30天

wechat:
  app_id: ""
  app_secret: ""
  enable_login: true

storage:
  type: local           # local/oss/cos
  local_path: ./uploads

sms:
  provider: aliyun      # aliyun/tencent
  dev_mode: true        # 开发模式不实际发送短信
  sign_name: "团圆寻亲"

backup:
  enabled: false        # 生产环境建议启用
  backup_dir: ./backups
  retention: 7          # 保留天数
```

## 权限控制

### 角色层级
| 角色 | 权重 | 权限范围 |
|------|------|---------|
| super_admin | 100 | 全部权限，日志清理 |
| admin | 80 | 用户/组织管理，数据管理 |
| manager | 60 | 案件/任务管理，审核 |
| volunteer | 40 | 基本查看和操作 |

### 权限中间件
- `RequireRole(minRole)`: 需要指定角色及以上
- `RequireAdmin()`: 需要管理员权限
- `RequireManager()`: 需要管理者权限
- `RequireSuperAdmin()`: 需要超级管理员权限

说明：当前运行时权限判定采用 RBAC（角色层级），`ty_permissions` / `ty_user_permissions` 作为扩展字典保留。

## 后端模块状态

| 模块 | 仓储 | 应用服务 | Handler | 状态 |
|------|------|----------|---------|------|
| 认证授权 | - | AuthService | AuthHandler | ✅ 完整 |
| 用户管理 | UserRepository | UserAppService | UserHandler | ✅ 完整 |
| 组织管理 | OrganizationRepository | OrganizationAppService | OrganizationHandler | ✅ 完整 |
| 走失人员 | MissingPersonRepository | MissingPersonAppService | MissingPersonHandler | ✅ 完整 |
| 方言管理 | DialectRepository | DialectAppService | DialectHandler | ✅ 完整 |
| 任务管理 | TaskRepository | TaskAppService | TaskHandler | ✅ 完整 |
| 文件管理 | FileRepository | FileAppService | UploadHandler | ✅ 完整 |
| 仪表盘 | - | DashboardService | DashboardHandler | ✅ 完整 |
| 审计日志 | AuditLogRepository | AuditLogService | AuditHandler | ✅ 完整 |
| 短信服务 | - | SMSService | - | ✅ 阿里云/腾讯云 |
| 微信客户端 | - | WechatClient | - | ✅ 登录/解密/手机号 |
| 定时任务 | - | Scheduler | - | ✅ 逾期检测/备份/清理 |
| 数据库备份 | - | Scheduler | - | ✅ pg_dump/mysqldump |
| 病毒扫描 | - | ClamAVScanner | - | ✅ ClamAV TCP协议 |

## 默认账号

```
手机号: 13800138000
密码: admin123
```
