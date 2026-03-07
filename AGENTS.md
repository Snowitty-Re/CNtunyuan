# CNtunyuan - 开发指南

本文档为 AI 助手和开发者提供后端 API 项目的背景信息和开发规范。

## 项目背景

团圆寻亲志愿者系统是一个帮助寻找走失人员的公益项目后端 API 服务。

### 核心价值
- **志愿者协作**: 组织架构化的志愿者管理
- **方言辅助**: 通过方言语音帮助确认走失人员身份
- **任务驱动**: OA工作流确保寻人任务有序进行

## 技术架构

### Clean Architecture
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

### Go 后端
- 使用 `gofmt` 格式化代码
- 遵循 Go 官方代码规范
- 函数命名使用驼峰式
- 接口命名使用动词+名词，如 `CreateUser`
- 错误处理必须返回具体错误信息

### 项目结构
```
backend/
├── cmd/                      # 应用程序入口
│   ├── app/                 # HTTP 服务器（统一入口）
│   ├── seed/                # 数据填充工具
│   └── resetpassword/       # 密码重置工具
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
│   │   ├── cache/           # 缓存
│   │   ├── repository/      # 仓储实现
│   │   └── auth/            # 认证
│   │
│   ├── interfaces/          # 接口适配层
│   │   └── http/
│   │       ├── handler/     # HTTP 处理器
│   │       ├── middleware/  # HTTP 中间件
│   │       └── router/      # 路由
│   │
│   ├── di/                  # 依赖注入
│   └── config/              # 配置
│
└── pkg/                     # 公共包
```

### 数据库规范

#### 表命名
- 使用前缀 `ty_`
- 复数形式，如 `users`, `organizations`
- 关联表使用 `_` 连接，如 `user_roles`

#### 字段命名
- 使用下划线命名法
- 常用字段: `created_at`, `updated_at`, `deleted_at`
- 外键使用 `_id` 后缀
- 布尔值使用 `is_` 前缀

#### API 设计规范

##### RESTful API
```
GET    /api/v1/resources      # 列表
POST   /api/v1/resources      # 创建
GET    /api/v1/resources/:id   # 详情
PUT    /api/v1/resources/:id   # 更新
DELETE /api/v1/resources/:id   # 删除
```

##### 响应格式
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

# 开发模式启动
go run cmd/app/main.go

# 数据填充
go run cmd/seed/main.go -all

# 重置密码
go run cmd/resetpassword/main.go -phone=13800138000 -password=newpassword

# 格式化代码
go fmt ./...

# 运行测试
go test ./...
```

### 数据库迁移

**PostgreSQL:**
```bash
# 创建数据库
createdb -U postgres -E UTF8 cntuanyuan

# 执行表结构
psql -U postgres -d cntuanyuan -f backend/migrations/postgres/01_schema.sql

# 插入种子数据
psql -U postgres -d cntuanyuan -f backend/migrations/postgres/02_seed.sql
```

**MySQL:**
```bash
# 创建数据库
mysql -u root -p -e "CREATE DATABASE cntuanyuan CHARACTER SET utf8mb4;"

# 执行表结构
mysql -u root -p cntuanyuan < backend/migrations/mysql/01_schema.sql

# 插入种子数据
mysql -u root -p cntuanyuan < backend/migrations/mysql/02_seed.sql
```

## 参考文档

- [backend/README.md](backend/README.md) - 后端详细说明
- [backend/ARCHITECTURE.md](backend/ARCHITECTURE.md) - 架构详细说明
