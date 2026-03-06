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

### 前端架构
```
┌─────────────┐
│   页面层    │  (pages/)
├─────────────┤
│  服务层     │  (services/)
├─────────────┤
│  状态层     │  (stores/)
├─────────────┤
│  组件层     │  (components/)
└─────────────┘
```

## 开发规范

### 代码规范

#### Go 后端
- 使用 `gofmt` 格式化代码
- 遵循 Go 官方代码规范
- 函数命名使用驼峰式
- 接口命名使用动词+名词，如 `CreateUser`
- 错误处理必须返回具体错误信息

#### React 前端 (web-new)
- 使用 TypeScript 严格模式
- 组件命名使用大驼峰式
- Props 必须定义类型
- 使用 hooks 进行状态管理
- **样式规范**: 不使用 Tailwind className，使用 Ant Design 组件默认样式 + 内联 style
- **颜色规范**: 使用温馨橙色主题 (`#e67e22`)，背景 `#f5f7fa`，主文字 `#1f2329`
- **设计原则**: 简洁办公OA风格，去除多余装饰，注重信息层级

### 后端项目结构
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
└── pkg/                     # 公共库
    ├── logger/              # 日志
    ├── response/            # HTTP 响应
    └── utils/               # 工具函数
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

#### 索引规范
- 外键自动创建索引
- 常用查询字段添加索引
- 唯一索引使用 `uniqueIndex` 标签
- 复合索引遵循最左前缀原则

#### 外键约束
- 启用外键约束（生产环境）
- 删除策略：
  - `CASCADE` - 级联删除（子表数据）
  - `SET NULL` - 设为 NULL（可选关联）
  - `RESTRICT` - 限制删除（有关联数据时禁止删除）
- 更新策略：`ON UPDATE CASCADE`

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

#### 错误码
- 0: 成功
- 400: 参数错误
- 401: 未授权
- 403: 禁止访问
- 404: 资源不存在
- 500: 服务器错误

## 常用命令

### 后端命令

```bash
cd backend

# 开发模式启动（统一入口：cmd/app/main.go）
go run cmd/app/main.go

# 检查数据库表结构
go run cmd/app/main.go -check-db

# 数据填充（生成测试数据）
go run cmd/seed/main.go -all

# 重置密码
go run cmd/resetpassword/main.go -phone=13800138000 -password=newpassword

# 格式化代码
go fmt ./...

# 运行测试
go test ./...
```

### 数据库迁移命令

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
mysql -u root -p -e "CREATE DATABASE cntuanyuan CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 执行表结构
mysql -u root -p cntuanyuan < backend/migrations/mysql/01_schema.sql

# 插入种子数据
mysql -u root -p cntuanyuan < backend/migrations/mysql/02_seed.sql
```

### 前端命令

```bash
cd web-new

# 安装依赖
pnpm install

# 开发模式
pnpm dev

# 构建
pnpm build

# 预览
pnpm preview
```

## 数据初始化

### 种子数据导入
```bash
# 导入所有种子数据
cd backend && go run cmd/seed/main.go -all

# 只导入特定类型数据
cd backend && go run cmd/seed/main.go -orgs     # 只导入组织
cd backend && go run cmd/seed/main.go -users    # 只导入用户
```

### 完整初始化流程（新环境）

#### 命令行初始化
```bash
# 1. 确保数据库已创建（或配置好数据库连接）
# 2. 执行数据库迁移
cd backend && go run cmd/app/main.go -migrate

# 3. 启动服务器
cd backend && go run cmd/app/main.go
```

## 配置说明

### 后端配置 (config/config.yaml)

```yaml
server:
  port: "8080"
  mode: "debug"  # debug/release
  domain: "http://localhost:8080"  # 后端域名，用于生成文件URL等

database:
  type: "postgres"   # 数据库类型: postgres 或 mysql
  host: "localhost"
  port: 5432         # MySQL 默认 3306
  user: "postgres"   # MySQL 默认 root
  password: "postgres"
  database: "cntuanyuan"
  ssl_mode: "disable"  # PostgreSQL 专用
  charset: "UTF8"      # MySQL 默认 utf8mb4
  max_idle_conns: 10
  max_open_conns: 100
  conn_max_lifetime: 3600

redis:
  host: ""       # 空表示不使用Redis
  port: 6379
  password: ""
  db: 0

jwt:
  secret: "your-secret-key"  # 生产环境必须修改
  expire_time: 604800  # 7天
  refresh_time: 2592000  # 30天

wechat:
  app_id: "your-app-id"          # 微信小程序 AppID
  app_secret: "your-app-secret"  # 微信小程序 AppSecret
  enable_login: true             # 是否启用微信登录
  mch_id: ""      # 商户号(支付用，可选)
  api_key: ""     # API密钥(支付用，可选)
  notify_url: ""  # 支付回调地址(可选)

storage:
  type: local           # 存储类型: local/oss/cos
  local_path: ./uploads # 本地存储路径
  base_url: http://localhost:8080/uploads
  max_file_size: 52428800  # 50MB
  allowed_types: "jpg,png,gif,mp4,mp3,wav"
  
  # 阿里云OSS配置（可选）
  oss_access_key_id: ""
  oss_access_key_secret: ""
  oss_endpoint: ""
  oss_bucket: ""
  oss_region: ""
  
  # 腾讯云COS配置（可选）
  cos_secret_id: ""
  cos_secret_key: ""
  cos_bucket: ""
  cos_region: ""

sms:
  provider: aliyun  # aliyun/tencent
  sign_name: "团圆寻亲"
  # 阿里云短信
  aliyun_access_key_id: ""
  aliyun_access_key_secret: ""
  # 腾讯云短信
  tencent_secret_id: ""
  tencent_secret_key: ""
  tencent_app_id: ""

system:
  default_org_name: "团圆寻亲志愿者协会"
  enable_register: true      # 是否开放注册
  enable_wechat_login: true  # 是否启用微信登录
  enable_sms_login: false    # 是否启用短信登录
  rate_limit: 100            # 每分钟请求限制
```

### 前端配置 (.env)

```env
VITE_API_BASE_URL=/api/v1
```

## 注意事项

1. **Redis 可选**: 如果 Redis 未配置，系统会自动使用内存缓存
2. **数据库表前缀**: 所有表使用 `ty_` 前缀
3. **数据库类型**: 支持 PostgreSQL 16+ 和 MySQL 8.0+，通过 `database.type` 配置切换
4. **JWT 密钥**: 生产环境必须修改默认密钥
5. **微信小程序**: 需要配置正确的 appid 和密钥
6. **文件存储**: 
   - 本地存储需要确保 `./uploads` 目录存在且有写入权限
   - 生产环境建议使用 OSS 或 COS
   - 文件上传大小限制默认为 50MB
7. **初始化**: 首次启动前请确保已配置数据库并执行数据迁移

## 安全指南

### 生产环境部署检查清单

#### 敏感信息保护
- [ ] 所有敏感信息使用环境变量或配置文件（config.yaml 已添加到 .gitignore）
- [ ] JWT 密钥已修改（长度 >= 32，不要使用默认密钥）
- [ ] 数据库密码已修改为强密码
- [ ] 微信小程序凭证已配置且安全存储

#### 网络安全
- [ ] HTTPS 已启用（生产环境必须）
- [ ] 安全响应头已配置（X-Content-Type-Options, X-Frame-Options, CSP 等）
- [ ] CORS 配置正确，只允许需要的域名
- [ ] 限流已启用（IP 限流：每秒 100 请求，突发 200）

#### 输入安全
- [ ] 所有 Handler 入口进行参数验证
- [ ] 使用参数化查询防止 SQL 注入（GORM 默认防护）
- [ ] XSS 防护：字符串清理、输出编码
- [ ] 文件上传限制类型和大小

### 凭证轮换流程

如果怀疑敏感信息泄露，立即执行：

1. **微信小程序凭证**
   - 登录 [微信小程序后台](https://mp.weixin.qq.com)
   - 进入「开发」->「开发管理」->「开发设置」
   - 点击 AppSecret 的「重置」按钮

2. **数据库密码**
   - 修改 PostgreSQL/MySQL 的密码
   - 更新 `backend/config/config.yaml` 文件

3. **JWT 密钥**
   - 生成新的随机密钥（长度 >= 32）
   - 更新配置文件
   - 所有用户需要重新登录（密钥更改后现有 token 失效）

### 安全最佳实践

1. **提交前检查**
   ```bash
   git diff --cached
   # 确保没有敏感信息被提交
   ```

2. **使用 git-secrets 工具**
   ```bash
   # 安装
   brew install git-secrets  # macOS
   
   # 配置
   git secrets --install
   git secrets --register-aws
   ```

3. **配置文件管理**
   - 始终使用 `config.yaml.example` 作为模板
   - 复制为 `config.yaml` 并填入真实值
   - `config.yaml` 已在 .gitignore 中，不会被提交

## 常见问题

### 1. 数据库连接失败
- 检查数据库服务是否启动（PostgreSQL/MySQL）
- 检查 config.yaml 中的数据库配置（类型、主机、端口、用户名、密码）
- 确认数据库 `cntuanyuan` 已手动创建
- 检查防火墙设置是否允许连接

### 2. 首次启动如何初始化系统
1. **手动创建数据库**
   - PostgreSQL: `createdb -U postgres -E UTF8 cntuanyuan`
   - MySQL: `CREATE DATABASE cntuanyuan CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;`

2. **执行 SQL 迁移文件**
   - PostgreSQL: `psql -U postgres -d cntuanyuan -f backend/migrations/postgres/01_schema.sql`
   - MySQL: `mysql -u root -p cntuanyuan < backend/migrations/mysql/01_schema.sql`

3. **插入种子数据**
   - PostgreSQL: `psql -U postgres -d cntuanyuan -f backend/migrations/postgres/02_seed.sql`
   - MySQL: `mysql -u root -p cntuanyuan < backend/migrations/mysql/02_seed.sql`

4. **启动后端服务**: `cd backend && go run cmd/app/main.go`

5. **验证安装**: `cd backend && go run cmd/app/main.go -check-db`

6. 访问 `http://localhost:3000`，使用种子数据的默认账号登录
   - 超级管理员: 13800138000 / admin123

### 3. 开发环境快速初始化（不推荐用于生产）
```bash
cd backend
# 自动迁移表结构
go run cmd/app/main.go -migrate
# 插入种子数据
go run cmd/app/main.go -seed
```

### 3. 登录失败
- 确认系统已完成初始化（访问 `/setup` 查看状态）
- 检查密码是否正确
- 查看后端日志确认错误信息

### 4. 微信小程序登录配置
- 在 [微信小程序后台](https://mp.weixin.qq.com) 获取 AppID 和 AppSecret
- 在 `config/config.yaml` 中配置 `wechat.app_id` 和 `wechat.app_secret`
- 设置 `wechat.enable_login: true` 启用微信登录
- 确保小程序的 `request` 合法域名已配置后端地址
- 重新启动后端服务使配置生效

### 5. 微信登录后跳回登录页
- 检查后端日志确认是否成功获取 openid
- 确认 `wechat.app_id` 和 `wechat.app_secret` 配置正确
- 检查后端是否生成了有效的 JWT token（不是 mock token）
- 确认小程序存储 token 成功（检查 `wx.getStorageSync('token')`）

### 4. 前端代理问题
- 检查 vite.config.ts 中的代理配置
- 确认后端服务已启动
- 检查端口号是否正确

### 5. MySQL 连接问题
- 确保 MySQL 8.0+ 版本
- 检查字符集设置为 `utf8mb4`
- 如果使用 `caching_sha2_password` 认证，确保驱动版本兼容

## 更新日志

### 2026-03-06
- **生产级重构完成**:
  - 统一错误处理 (`pkg/errors`): 完整的错误码体系、AppError 结构体、HTTP 状态码映射
  - 统一响应格式 (`pkg/response`): 标准化 API 响应、分页数据支持
  - 请求验证 (`pkg/validator`): 基于 go-playground/validator、自定义验证规则、XSS 防护
  - 中间件 (`pkg/middleware`): TraceID、日志、恢复、安全头、限流、CORS、请求大小限制
  - Prometheus 监控 (`pkg/metrics`): HTTP/DB 指标、业务操作计数、缓存命中率
  - 数据库连接池优化: 连接池监控、慢查询检测(>100ms)、连接等待告警
  - 前端 Error Boundary: React 错误边界处理
  - 请求工具增强: 自动重试(最多3次)、防抖请求、统一错误处理

### 2026-03-05
- **移除系统初始化向导**:
  - 删除 `/setup` 页面和相关后端接口
  - 回归传统手动配置方式
  - 简化部署流程，减少潜在问题

### 2026-03-04
- **MySQL 8.0 支持**:
  - 新增 MySQL 驱动支持 (`gorm.io/driver/mysql`)
  - 数据库配置新增 `type` 字段，支持 `postgres` 和 `mysql`
  - 自动检测数据库类型并创建相应连接
  - MySQL 使用 `utf8mb4` 编码支持 Emoji
  - 自动创建数据库（如果不存在）
- **系统初始化向导** (已移除):
  - ~~新增 `/setup` 页面用于首次初始化~~
  - ~~支持在浏览器中配置数据库连接（PostgreSQL/MySQL）~~
- **Bug 修复**:
  - 修复登录接口 token 字段名不匹配问题
  - 修复响应状态码处理（支持 code 0 和 200）

### 2026-03-03
- **数据库编码优化**:
  - 统一使用 UTF-8 编码，支持中文和 Emoji
  - 数据库连接字符串添加 `client_encoding=UTF8`
  - 配置文件中添加 `charset` 选项
  - 创建完整的数据库初始化 SQL 脚本
  - 创建表结构 SQL 脚本（支持 GORM AutoMigrate）
- **完善后端功能**:
  - 文件存储服务（本地/OSS/COS）
  - 仪表盘统计服务
  - 应用入口和种子数据工具

### 2026-03-02
- **后端架构重构为 Clean Architecture**:
  - 领域层：实体、值对象、仓储接口、领域服务
  - 应用层：DTO、应用服务
  - 基础设施层：数据库、缓存、JWT、仓储实现
  - 接口层：HTTP处理器、中间件、路由
  - 依赖注入容器
- **Web 平台全面重构**:
  - 删除旧版 `web-admin`，使用新版 `web-new`
  - 全新简洁办公OA风格设计
  - 温馨橙色主题 (`#e67e22`)
  - 去除 Tailwind CSS，使用 Ant Design 5 + 内联样式

### 2024-03-02
- **完善权限控制系统**：
  - 后端RBAC权限中间件
  - API路由精细化权限控制
  - 前端usePermission Hooks和PermissionGuard组件
- **操作日志审计**：
  - 自动记录所有API请求
  - 支持按用户、模块、操作、时间筛选
- **数据迁移与初始化**：
  - 完善AutoMigrate，支持所有模型
  - GORM自动外键约束
  - 新增种子数据导入工具

## 权限控制

### 后端权限控制

#### 角色层级
- `super_admin`: 超级管理员 - 拥有所有权限
- `admin`: 管理员 - 管理用户、组织
- `manager`: 管理者 - 分配任务
- `volunteer`: 志愿者 - 基本操作

#### 权限中间件
- `RequireRole(minRole)`: 需要指定角色及以上
- `RequireAdmin()`: 需要管理员权限
- `RequireManager()`: 需要管理者权限
- `RequireSuperAdmin()`: 需要超级管理员权限

### 前端权限控制

#### React Hooks
- `usePermission()`: 获取当前用户权限信息
- `PermissionGuard`: 权限包装组件

#### 路由权限
- `/users`: Admin+
- `/settings`: Admin+
- 其他页面: 所有登录用户

## 参考文档

- [README.md](README.md) - 项目主文档（项目介绍、快速开始、功能特性）
- [backend/README.md](backend/README.md) - 后端开发指南（API 文档、开发规范）
- [backend/ARCHITECTURE.md](backend/ARCHITECTURE.md) - 后端架构详细说明（Clean Architecture 设计）
- [backend/REFACTORING.md](backend/REFACTORING.md) - 架构重构指南
- [backend/migrations/README.md](backend/migrations/README.md) - 数据库迁移指南
