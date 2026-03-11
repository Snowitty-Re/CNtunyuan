# 团圆寻亲志愿者系统 (CNtunyuan)

一个生产级别的寻亲志愿者管理系统，包含微信小程序端和后端API服务。

## 功能特性

- **志愿者管理** - 微信一键登录，组织架构（省/市/区/街道），RBAC权限控制
- **走失人员数据库** - 登记、照片上传、轨迹跟踪、关联方言、状态流转
- **方言语音库** - 音频文件上传、区域标记、播放统计、精选管理
- **任务管理** - 创建分配、人员选择器、进度追踪、逾期检测、操作日志
- **文件存储** - 本地/OSS/COS，安全检查，图片和音频上传
- **审计日志** - 全量API请求记录，敏感操作监控，日志清理
- **数据统计** - 仪表盘、趋势分析、我的任务、实时统计
- **组织树管理** - 层级选择器、树形结构、上下级关联
- **Swagger API 文档** - 自动生成，完整覆盖所有接口

## 技术栈

| 模块 | 技术 |
|------|------|
| 后端 | Go 1.23+, Gin, GORM, PostgreSQL, Redis, JWT |
| API 文档 | Swagger UI (swaggo/swag 自动生成) |
| 小程序 | 微信小程序原生开发 |

## 快速开始

### 1. 克隆项目

```bash
git clone https://github.com/Snowitty-Re/CNtunyuan.git
cd CNtunyuan
```

### 2. 启动后端

```bash
cd backend

# 配置数据库 (config/config.yaml)
cp config/config.yaml.example config/config.yaml
# 编辑 config.yaml 填写数据库信息

# 数据库迁移
go run cmd/app/main.go -migrate

# 启动服务
go run cmd/app/main.go
```

服务启动后访问：
- API 文档：http://localhost:8080/swagger/index.html
- 健康检查：http://localhost:8080/api/v1/health

### 3. 微信小程序

使用微信开发者工具打开 `mini-program` 目录。

## 项目结构

```
CNtunyuan/
├── backend/          # Go 后端 (Clean Architecture / DDD)
│   ├── cmd/app/      # 主应用入口
│   ├── internal/     # 业务代码 (domain/application/infrastructure/interfaces)
│   ├── pkg/          # 公共包 (logger/response/middleware)
│   ├── docs/         # Swagger 自动生成文档
│   └── migrations/   # 数据库迁移
├── mini-program/     # 微信小程序
├── docker/           # Docker 部署配置
└── scripts/          # 运维脚本
```

## API 模块

| 模块 | API 路径 | 说明 |
|------|---------|------|
| 认证管理 | `/auth/*` | 登录、注册、微信登录、Token 刷新 |
| 用户管理 | `/users/*` | 用户 CRUD、角色/状态管理 |
| 个人中心 | `/profile/*` | 资料编辑、密码修改 |
| 组织管理 | `/organizations/*` | 组织架构、树形结构 |
| 案件管理 | `/missing-persons/*` | 走失人员登记、照片/轨迹、状态流转 |
| 任务管理 | `/tasks/*` | 任务创建分配、进度追踪、操作日志 |
| 方言管理 | `/dialects/*` | 方言音频上传、评论、精选 |
| 文件管理 | `/upload/*` | 文件上传下载、实体绑定 |
| 仪表盘 | `/dashboard/*` | 数据统计、趋势分析 |
| 审计日志 | `/audit/*` | 操作日志、统计、清理 |

所有接口详情请访问 Swagger UI：http://localhost:8080/swagger/index.html

## 角色权限 (RBAC)

| 角色 | 权重 | 权限范围 |
|------|------|---------|
| super_admin | 100 | 全部权限，日志清理 |
| admin | 80 | 用户/组织管理，数据管理 |
| manager | 60 | 案件/任务管理，审核 |
| volunteer | 40 | 基本查看和操作 |

## 文档

| 文档 | 说明 |
|------|------|
| [backend/README.md](backend/README.md) | 后端开发指南 |
| [backend/REFACTORING.md](backend/REFACTORING.md) | Clean Architecture 设计文档 |
| [backend/TESTING.md](backend/TESTING.md) | 单元测试文档 |
| [backend/SWAGGER.md](backend/SWAGGER.md) | Swagger API 文档使用 |
| [backend/migrations/README.md](backend/migrations/README.md) | 数据库迁移指南 |

## 开发规范

### 后端
- Clean Architecture 分层架构 (Domain -> Application -> Infrastructure -> Interfaces)
- 领域驱动设计 (DDD)，充血实体模型
- 单元测试覆盖核心逻辑 (110+ 测试函数)
- Swagger 注解自动生成 API 文档，运行 `swag init -g cmd/app/main.go -o docs --parseDependency --parseInternal`

## 默认账号

```
手机号: 13800138000
密码: admin123
```

## Docker 部署

```bash
cd docker
cp .env.example .env
# 编辑 .env 配置
docker-compose -f docker-compose.prod.yml up -d
```

## 许可证

MIT License
