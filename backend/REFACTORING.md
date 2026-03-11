# 后端架构设计说明

## 架构概览

后端采用 **Clean Architecture**（整洁架构）设计，实现关注点分离、可测试性和可维护性。

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

## 目录结构

```
backend/
├── cmd/                      # 应用程序入口
│   ├── app/                 # HTTP 服务器（统一入口）
│   │   └── main.go
│   ├── seed/                # 数据填充工具
│   │   └── main.go
│   └── resetpassword/       # 密码重置工具
│       └── main.go
│
├── internal/                # 私有应用代码
│   ├── domain/              # 领域层
│   │   ├── entity/          # 领域实体（充血模型）
│   │   ├── valueobject/     # 值对象
│   │   ├── repository/      # 仓储接口
│   │   └── service/         # 领域服务（认证、存储、病毒扫描接口）
│   │
│   ├── application/         # 应用层
│   │   ├── dto/             # 数据传输对象
│   │   └── service/         # 应用服务（用例编排）
│   │
│   ├── infrastructure/      # 基础设施层
│   │   ├── database/        # 数据库连接（PostgreSQL/MySQL）
│   │   ├── cache/           # 缓存（Redis / 内存回退）
│   │   ├── repository/      # 仓储实现（GORM）
│   │   ├── auth/            # JWT 认证（access + refresh token）
│   │   ├── storage/         # 文件存储（本地/OSS/COS）+ ClamAV
│   │   ├── sms/             # 短信服务（阿里云/腾讯云）
│   │   └── wechat/          # 微信小程序客户端
│   │
│   ├── interfaces/          # 接口适配层
│   │   └── http/
│   │       ├── handler/     # HTTP 处理器（9个模块）
│   │       ├── middleware/   # 中间件（认证/审计/CORS/限流）
│   │       └── router/      # 路由注册
│   │
│   ├── di/                  # 手动依赖注入容器
│   ├── config/              # 配置加载
│   └── task/                # 定时任务（逾期检测/备份/清理）
│
├── pkg/                     # 公共库
│   ├── logger/              # Zap 日志
│   ├── errors/              # 统一错误处理
│   ├── response/            # HTTP 响应
│   ├── middleware/           # 通用中间件
│   ├── validator/           # 请求验证
│   ├── metrics/             # Prometheus 监控
│   └── utils/               # 工具函数
│
├── migrations/              # 数据库迁移（PostgreSQL/MySQL）
├── docs/                    # Swagger 文档
└── config/                  # 配置文件
```

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
- **Infrastructure** 依赖 Domain（实现其接口）
- **Interfaces** 依赖 Application

### 2. 领域实体（充血模型）

实体包含业务逻辑和行为方法，自验证，不依赖外部框架。

```go
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

接口定义在 Domain 层，实现在 Infrastructure 层，便于测试。

### 4. DTO 隔离

DTO 用于层间数据传输，防止内部模型泄露到外部。

## 关键技术决策

- **GORM**: 连接池（10 idle / 100 max），慢查询日志（200ms），软删除
- **JWT**: access + refresh token 双令牌，Redis 黑名单
- **缓存**: Redis 优先，不可用时自动降级内存缓存
- **密码**: bcrypt 加密存储
- **权限**: RBAC 四级角色（super_admin > admin > manager > volunteer）
- **短信**: 阿里云 HMAC-SHA1 签名 / 腾讯云 TC3-HMAC-SHA256 签名
- **文件**: 本地存储 + 阿里云 OSS + 腾讯云 COS，ClamAV 病毒扫描
- **备份**: 定时 pg_dump/mysqldump + gzip 压缩 + 过期清理

## 添加新模块

按以下顺序实现：

1. **Domain 层** - 创建实体和仓储接口
2. **Infrastructure 层** - 实现仓储
3. **Application 层** - 创建应用服务和 DTO
4. **Interfaces 层** - 创建 Handler 和注册路由
5. **DI 容器** - 在 `wire_gen.go` 中注入依赖

## 完成状态

- [x] Clean Architecture 分层完成
- [x] 9 个业务模块全部实现
- [x] 单元测试 110+ 测试函数
- [x] Swagger API 文档 67 个端点
- [x] 短信服务（阿里云/腾讯云）生产级实现
- [x] 微信小程序登录/解密
- [x] ClamAV 病毒扫描
- [x] 定时任务（逾期检测/数据库备份/日志清理）
- [ ] 集成测试
- [ ] CI/CD 配置
- [ ] 性能基准测试
