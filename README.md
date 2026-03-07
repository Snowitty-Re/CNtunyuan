# 团圆寻亲志愿者系统 API (CNtunyuan)

一个生产级别的寻亲志愿者管理后端 API 服务。

## 项目概述

本项目旨在帮助寻找走失人员，通过志愿者网络、方言语音数据库和OA工作流系统，提高寻人效率。

**此仓库仅包含后端 API 服务，前端部分已独立维护。**

## 功能特性

### 核心功能
- **志愿者管理**：组织架构、角色权限、微信登录集成
- **走失人员数据库**：登记、轨迹跟踪、状态管理
- **方言语音数据库**：语音上传、区域标记、分类管理
- **任务管理**：创建、分配、进度追踪
- **文件存储**：本地存储、阿里云 OSS、腾讯云 COS 支持
- **工作流管理**：审批流程、步骤管理

### 系统特性
- JWT 认证授权
- RBAC 权限控制
- 完整的审计日志
- 数据统计分析
- RESTful API 设计

## 技术栈

- **Go 1.24+**
- **Gin** 框架
- **GORM** ORM
- **PostgreSQL 16** / **MySQL 8.0**
- **Redis 7** (可选，用于缓存和 Token 黑名单)
- **Clean Architecture** 架构

## 项目结构

```
CNtunyuan/
├── backend/              # Go 后端服务 (Clean Architecture)
│   ├── cmd/              # 应用程序入口
│   │   ├── app/          # HTTP 服务器（统一入口）
│   │   ├── seed/         # 数据填充工具
│   │   └── resetpassword/# 密码重置工具
│   ├── internal/         # 内部包
│   │   ├── domain/       # 领域层 (实体、仓储接口、领域服务)
│   │   ├── application/  # 应用层 (DTO、应用服务)
│   │   ├── infrastructure/# 基础设施层 (DB、缓存、仓储实现)
│   │   ├── interfaces/   # 接口层 (HTTP处理器、中间件)
│   │   ├── di/           # 依赖注入
│   │   └── config/       # 配置
│   ├── pkg/              # 公共包
│   ├── migrations/       # 数据库迁移脚本
│   ├── uploads/          # 本地文件存储目录
│   └── Dockerfile
├── docker/               # Docker 配置
│   ├── docker-compose.yml
│   └── nginx/
└── scripts/              # 部署脚本
```

## 快速开始

### 环境要求
- Go 1.24+
- PostgreSQL 16 或 MySQL 8.0
- Redis 7 (可选)

### 配置说明

复制示例配置文件并修改：

```bash
cd backend
cp config/config.example.yaml config/config.yaml
```

主要配置项：

```yaml
server:
  port: 8080
  mode: debug  # debug/release
  domain: "https://api.your-domain.com"

database:
  type: "postgres"  # 或 mysql
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "your-password"
  database: "cntuanyuan"

wechat:
  app_id: "your-wechat-app-id"
  app_secret: "your-wechat-app-secret"
  enable_login: true

jwt:
  secret: "your-jwt-secret-key"  # 生产环境必须修改
  expire_time: 604800  # 7天

storage:
  type: local
  local_path: ./uploads
  base_url: https://api.your-domain.com/uploads
```

### 数据库初始化

```bash
# PostgreSQL
psql -U postgres -d cntuanyuan -f backend/migrations/postgres/01_schema.sql

# MySQL
mysql -u root -p cntuanyuan < backend/migrations/mysql/01_schema.sql
```

### 启动服务

```bash
cd backend

# 安装依赖
go mod download

# 运行
go run cmd/app/main.go

# 或使用 air 热重载
air
```

服务启动后访问：`http://localhost:8080/health`

## Docker 部署

```bash
cd docker
docker-compose up -d
```

## API 文档

启动服务后访问 Swagger UI：`http://localhost:8080/swagger/index.html`

## 默认账号

种子数据包含超级管理员账号：

- **手机号**: 13800138000
- **密码**: admin123

导入种子数据：
```bash
cd backend
go run cmd/seed/main.go -all
```

重置密码：
```bash
cd backend
go run cmd/resetpassword/main.go -phone=13800138000 -password=newpassword
```

## 开发指南

```bash
cd backend

# 格式化代码
go fmt ./...

# 运行测试
go test ./...

# 构建
go build -o cntuanyuan-api ./cmd/app/main.go
```

## 架构文档

- [backend/ARCHITECTURE.md](backend/ARCHITECTURE.md) - Clean Architecture 架构说明
- [AGENTS.md](AGENTS.md) - 开发者指南

## 许可证

MIT License
