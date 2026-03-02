# 团圆寻亲系统 - 后端

基于 Clean Architecture 的 Go 后端服务。

## 技术栈

- **Go 1.23+** - 编程语言
- **Gin** - Web 框架
- **GORM** - ORM 框架
- **PostgreSQL** - 数据库
- **JWT** - 认证
- **Zap** - 日志

## 项目结构

```
backend/
├── cmd/
│   ├── app/              # 主应用入口
│   └── seed/             # 种子数据导入工具
├── internal/
│   ├── config/           # 配置管理
│   ├── domain/           # 领域层
│   │   ├── entity/       # 领域实体
│   │   ├── repository/   # 仓储接口
│   │   └── service/      # 领域服务
│   ├── application/      # 应用层
│   │   ├── dto/          # 数据传输对象
│   │   └── service/      # 应用服务
│   ├── infrastructure/   # 基础设施层
│   │   ├── auth/         # 认证实现
│   │   ├── cache/        # 缓存实现
│   │   ├── database/     # 数据库连接
│   │   ├── repository/   # 仓储实现
│   │   └── storage/      # 文件存储
│   ├── interfaces/       # 接口层
│   │   └── http/
│   │       ├── handler/  # HTTP 处理器
│   │       ├── middleware/ # 中间件
│   │       └── router/   # 路由
│   └── di/               # 依赖注入
├── pkg/                  # 公共包
│   ├── logger/           # 日志
│   ├── response/         # HTTP 响应
│   └── validator/        # 验证器
├── config/               # 配置文件
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

## API 端点

### 认证
- `POST /api/v1/auth/login` - 登录
- `POST /api/v1/auth/register` - 注册
- `POST /api/v1/auth/logout` - 登出
- `POST /api/v1/auth/refresh` - 刷新令牌

### 用户
- `GET /api/v1/users` - 用户列表
- `GET /api/v1/users/:id` - 用户详情
- `PUT /api/v1/users/:id` - 更新用户
- `DELETE /api/v1/users/:id` - 删除用户

### 组织
- `GET /api/v1/organizations` - 组织列表
- `POST /api/v1/organizations` - 创建组织
- `GET /api/v1/organizations/:id` - 组织详情
- `PUT /api/v1/organizations/:id` - 更新组织
- `DELETE /api/v1/organizations/:id` - 删除组织

### 走失人员
- `GET /api/v1/missing-persons` - 案件列表
- `POST /api/v1/missing-persons` - 创建案件
- `GET /api/v1/missing-persons/:id` - 案件详情
- `PUT /api/v1/missing-persons/:id` - 更新案件
- `DELETE /api/v1/missing-persons/:id` - 删除案件

### 方言
- `GET /api/v1/dialects` - 方言列表
- `POST /api/v1/dialects` - 上传方言
- `GET /api/v1/dialects/:id` - 方言详情
- `PUT /api/v1/dialects/:id` - 更新方言
- `DELETE /api/v1/dialects/:id` - 删除方言

### 任务
- `GET /api/v1/tasks` - 任务列表
- `POST /api/v1/tasks` - 创建任务
- `GET /api/v1/tasks/:id` - 任务详情
- `PUT /api/v1/tasks/:id` - 更新任务
- `DELETE /api/v1/tasks/:id` - 删除任务

### 文件上传
- `POST /api/v1/upload` - 单文件上传
- `POST /api/v1/upload/batch` - 批量上传
- `DELETE /api/v1/upload/:id` - 删除文件

### 仪表盘
- `GET /api/v1/dashboard/stats` - 统计数据
- `GET /api/v1/dashboard/overview` - 概览数据
- `GET /api/v1/dashboard/trend` - 趋势数据

## 开发指南

### 添加新模块

1. **创建领域实体** (`internal/domain/entity/`)
2. **定义仓储接口** (`internal/domain/repository/`)
3. **实现仓储** (`internal/infrastructure/repository/`)
4. **创建应用服务** (`internal/application/service/`)
5. **创建 HTTP 处理器** (`internal/interfaces/http/handler/`)
6. **注册路由** (`internal/interfaces/http/router/`)
7. **更新 DI 容器** (`internal/di/`)

### 代码规范

- 使用 `gofmt` 格式化代码
- 函数命名使用驼峰式
- 接口命名使用动词+名词
- 错误处理返回具体错误信息

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

### Docker

```bash
# 构建镜像
docker build -t cntuanyuan-backend .

# 运行容器
docker run -p 8080:8080 cntuanyuan-backend
```

### 生产环境

```bash
# 编译
CGO_ENABLED=0 GOOS=linux go build -o app cmd/app/main.go

# 运行
./app
```

## 许可证

MIT License
