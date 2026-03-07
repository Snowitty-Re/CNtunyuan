# 团圆寻亲志愿者系统 - 后端 API

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-blue.svg)](https://postgresql.org)
[![Redis](https://img.shields.io/badge/Redis-7+-red.svg)](https://redis.io)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

> 帮助寻找走失人员的公益平台后端 API 服务

## 技术栈

- **语言**: Go 1.21+
- **Web框架**: Gin
- **ORM**: GORM
- **数据库**: PostgreSQL 15+
- **缓存**: Redis 7+
- **架构**: Clean Architecture / DDD

## 快速开始

### 1. 环境要求

- Go 1.21 或更高版本
- PostgreSQL 15 或更高版本
- Redis 7 或更高版本
- Make (可选)

### 2. 安装依赖

```bash
cd backend
go mod download
```

### 3. 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 文件，配置数据库和 Redis 连接信息
```

### 4. 数据库迁移

```bash
# 使用 PostgreSQL
psql -U postgres -d cntuanyuan -f migrations/postgres/001_init.sql
psql -U postgres -d cntuanyuan -f migrations/postgres/002_workflow.sql
psql -U postgres -d cntuanyuan -f migrations/postgres/003_audit_log.sql
psql -U postgres -d cntuanyuan -f migrations/postgres/004_permissions.sql
psql -U postgres -d cntuanyuan -f migrations/postgres/005_permissions.sql
psql -U postgres -d cntuanyuan -f migrations/postgres/006_notifications.sql
```

### 5. 启动服务

```bash
# 开发模式
go run cmd/app/main.go

# 或构建后运行
go build -o app cmd/app/main.go
./app
```

服务默认在 `:8080` 端口启动。

## 项目结构

```
backend/
├── cmd/                    # 应用程序入口
│   ├── app/               # HTTP 服务器
│   └── seed/              # 数据填充工具
├── internal/              # 私有应用代码
│   ├── domain/            # 领域层
│   │   ├── entity/        # 领域实体
│   │   ├── repository/    # 仓储接口
│   │   └── service/       # 领域服务
│   ├── application/       # 应用层
│   │   ├── dto/           # 数据传输对象
│   │   └── service/       # 应用服务
│   ├── infrastructure/    # 基础设施层
│   │   ├── database/      # 数据库
│   │   ├── cache/         # 缓存
│   │   ├── repository/    # 仓储实现
│   │   └── websocket/     # WebSocket
│   └── interfaces/        # 接口适配层
│       └── http/
│           ├── handler/   # HTTP 处理器
│           ├── middleware/# 中间件
│           └── router/    # 路由
├── pkg/                   # 公共包
│   ├── errors/            # 错误处理
│   ├── logger/            # 日志
│   └── utils/             # 工具函数
├── migrations/            # 数据库迁移
└── test/                  # 测试
```

## API 文档

启动服务后，访问 Swagger 文档：

```
http://localhost:8080/api/v1/swagger/index.html
```

## 主要功能

### Phase 1: 基础设施强化 ✅
- [x] 统一错误处理
- [x] 审计日志系统
- [x] 数据权限框架
- [x] 缓存抽象层

### Phase 2: OA 工作流引擎 ✅
- [x] 流程定义管理
- [x] 工作流状态机
- [x] 审批节点实现
- [x] 任务委托/催办

### Phase 3: 权限系统升级 ✅
- [x] RBAC 权限矩阵
- [x] 数据权限规则
- [x] 字段级权限
- [x] 权限缓存优化

### Phase 4: 通知系统 ✅
- [x] WebSocket 实时服务
- [x] 消息模板引擎
- [x] 多渠道推送
- [x] 消息中心

### Phase 5: 测试与文档 ✅
- [x] 单元测试覆盖
- [x] 集成测试
- [x] API 文档 (Swagger)
- [x] 部署文档

## 测试

### 运行单元测试

```bash
# 运行所有测试
go test ./...

# 运行指定包测试
go test ./internal/domain/entity/...
go test ./internal/application/service/...

# 生成测试覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 运行集成测试

```bash
go test ./test/integration/...
```

## 部署

### Docker 部署

```bash
docker-compose up -d
```

详见 [Docker 部署指南](../docs/deployment/docker.md)

### Kubernetes 部署

```bash
kubectl apply -f k8s/
```

详见 [Kubernetes 部署指南](../docs/deployment/kubernetes.md)

## 贡献

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 许可证

[MIT](LICENSE) © 团圆寻亲团队
