# 后端架构重构说明

## 重构目标

将后端从传统的 MVC 架构重构为符合 **Clean Architecture**（整洁架构）的生产级架构，实现：

1. **关注点分离** - 核心业务逻辑独立于框架、UI和数据库
2. **可测试性** - 业务逻辑可以独立测试
3. **可维护性** - 清晰的依赖关系和模块边界
4. **可扩展性** - 易于添加新功能和适配新技术

---

## 新架构层次

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

---

## 目录结构

```
backend/
├── cmd/                      # 应用程序入口
│   ├── server/              # HTTP 服务器
│   │   └── main.go
│   ├── seed/                # 数据填充工具
│   │   └── main.go
│   └── migrate/             # 数据库迁移工具
│
├── internal/                # 私有应用代码
│   ├── domain/              # 领域层
│   │   ├── entity/          # 领域实体
│   │   │   ├── entity.go    # 基础实体
│   │   │   ├── user.go      # 用户实体
│   │   │   ├── organization.go
│   │   │   ├── missing_person.go
│   │   │   ├── dialect.go
│   │   │   └── task.go
│   │   ├── valueobject/     # 值对象
│   │   │   └── user.go
│   │   ├── repository/      # 仓储接口
│   │   │   ├── repository.go
│   │   │   ├── user_repository.go
│   │   │   ├── organization_repository.go
│   │   │   ├── missing_person_repository.go
│   │   │   ├── dialect_repository.go
│   │   │   └── task_repository.go
│   │   ├── service/         # 领域服务
│   │   │   └── auth_service.go
│   │   └── event/           # 领域事件
│   │
│   ├── application/         # 应用层
│   │   ├── dto/             # 数据传输对象
│   │   │   └── user_dto.go
│   │   ├── service/         # 应用服务
│   │   │   └── user_service.go
│   │   └── mapper/          # 对象映射器
│   │
│   ├── infrastructure/      # 基础设施层
│   │   ├── database/        # 数据库
│   │   │   └── db.go
│   │   ├── cache/           # 缓存
│   │   │   └── redis.go
│   │   ├── repository/      # 仓储实现
│   │   │   ├── base_repository.go
│   │   │   └── user_repository.go
│   │   ├── auth/            # 认证
│   │   │   └── jwt_service.go
│   │   └── storage/         # 文件存储
│   │
│   ├── interfaces/          # 接口适配层
│   │   └── http/
│   │       ├── handler/     # HTTP 处理器
│   │       │   ├── auth_handler.go
│   │       │   └── user_handler.go
│   │       ├── middleware/  # HTTP 中间件
│   │       │   └── auth.go
│   │       └── router/      # 路由
│   │           └── router.go
│   │
│   ├── di/                  # 依赖注入
│   │   ├── wire.go          # Wire 配置
│   │   └── wire_gen.go      # 生成的代码
│   │
│   └── config/              # 配置
│       └── config.go
│
├── pkg/                     # 公共库
│   ├── logger/              # 日志
│   ├── response/            # HTTP 响应
│   └── utils/               # 工具函数
│
└── go.mod
```

---

## 核心设计原则

### 1. 依赖方向

依赖关系必须**向内指向领域层**：

```
Interfaces → Application → Domain
     ↑          ↑
Infrastructure ─┴──────────┘
```

- **Domain** 不依赖任何其他层
- **Application** 只依赖 Domain
- **Infrastructure** 依赖 Domain 和 Application
- **Interfaces** 依赖所有内层

### 2. 领域实体

- 包含业务逻辑和行为方法
- 使用充血模型而非贫血模型
- 自验证（Validate 方法）
- 不依赖外部框架

```go
// 示例：用户实体的业务方法
func (u *User) CanModify(operator *User) bool {
    if operator.IsSuperAdmin() {
        return true
    }
    if u.IsSuperAdmin() {
        return false
    }
    return GetRoleLevel(operator.Role) > GetRoleLevel(u.Role)
}
```

### 3. 仓储模式

- **接口定义在 Domain 层** - 领域定义需要什么样的数据访问
- **实现在 Infrastructure 层** - 使用 GORM 或其他技术实现
- **便于测试** - 可以用内存实现替换真实数据库

```go
// Domain 层 - 接口
type UserRepository interface {
    FindByID(ctx context.Context, id string) (*User, error)
    Save(ctx context.Context, user *User) error
}

// Infrastructure 层 - 实现
type UserRepositoryImpl struct {
    db *gorm.DB
}
```

### 4. DTO 和 Mapper

- **DTO** 用于层间数据传输
- **Mapper** 负责 Entity ↔ DTO 转换
- 防止内部模型泄露到外部

---

## 关键技术决策

### 1. GORM 配置优化

```go
// 启用外键约束、连接池、日志
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
    DisableForeignKeyConstraintWhenMigrating: false,
    Logger: logger.New(
        log.New(os.Stdout, "\r\n", log.LstdFlags),
        logger.Config{
            SlowThreshold:             200 * time.Millisecond,
            LogLevel:                  logger.Warn,
            IgnoreRecordNotFoundError: true,
        },
    ),
})

// 连接池配置
sqlDB, err := db.DB()
sqlDB.SetMaxIdleConns(10)
sqlDB.SetMaxOpenConns(100)
sqlDB.SetConnMaxLifetime(time.Hour)
```

### 2. JWT 实现

- 支持 access token 和 refresh token
- Token 黑名单机制（使用 Redis）
- 安全的密钥管理

### 3. 缓存策略

- Redis 作为分布式缓存
- 可选的内存缓存回退
- 缓存 key 统一命名规范

### 4. 错误处理

```go
// 领域错误
var (
    ErrUserNotFound      = errors.New("用户不存在")
    ErrInvalidCredentials = errors.New("用户名或密码错误")
)

// HTTP 响应统一格式
{
    "code": 401,
    "message": "请先登录",
    "data": null
}
```

---

## 迁移指南

### 步骤 1: 保持现有代码运行

现有代码保留在 `internal/model`, `internal/service`, `internal/api` 等目录。

### 步骤 2: 逐步迁移模块

按以下顺序迁移：

1. **Domain 层** - 创建实体和仓储接口
2. **Infrastructure 层** - 实现仓储
3. **Application 层** - 创建应用服务
4. **Interfaces 层** - 创建新 handler

### 步骤 3: 双轨运行

新旧系统可以并行运行，逐步切换：

```go
// 旧 handler
router.GET("/users", oldHandler.GetUsers)

// 新 handler  
router.GET("/v2/users", newHandler.List)
```

### 步骤 4: 完全切换

所有模块迁移完成后，删除旧代码。

---

## 测试策略

### 单元测试

```go
// 测试领域逻辑
func TestUser_CanModify(t *testing.T) {
    admin := &User{Role: RoleAdmin}
    volunteer := &User{Role: RoleVolunteer}
    
    assert.True(t, admin.CanModify(volunteer))
    assert.False(t, volunteer.CanModify(admin))
}
```

### 集成测试

```go
// 使用内存数据库测试仓储
func TestUserRepository(t *testing.T) {
    db := setupTestDB()
    repo := NewUserRepository(db)
    
    user, _ := NewUser("test", "13800138000", orgID, RoleVolunteer)
    err := repo.Save(context.Background(), user)
    
    assert.NoError(t, err)
}
```

---

## 性能优化

1. **连接池** - 配置合理的数据库连接池
2. **缓存** - Redis 缓存热点数据
3. **N+1 查询** - 使用 Preload 避免
4. **索引** - 为常用查询字段添加索引
5. **分页** - 所有列表接口默认分页

---

## 安全考虑

1. **密码加密** - 使用 bcrypt 存储密码
2. **JWT 安全** - 使用强密钥，设置过期时间
3. **SQL 注入** - 使用参数化查询
4. **权限控制** - 基于角色的访问控制（RBAC）
5. **输入验证** - 所有输入都经过验证

---

## 待完成任务

- [ ] 完成所有仓储实现
- [ ] 创建所有应用服务
- [ ] 创建所有 HTTP Handler
- [ ] 添加 Swagger 文档
- [ ] 完善单元测试
- [ ] 添加集成测试
- [ ] 配置 CI/CD
- [ ] 性能基准测试

---

## 参考资源

- [The Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html) - Robert C. Martin
- [Domain-Driven Design](https://domainlanguage.com/ddd/reference/) - Eric Evans
- [Go Clean Architecture](https://github.com/bxcodec/go-clean-arch) - 示例项目
