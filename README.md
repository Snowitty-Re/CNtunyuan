# 助力团圆志愿者系统 (CNtunyuan)

一个生产级别的志愿者协作系统，核心目标是帮助走失人员寻找亲属、助力团圆。当前包含微信小程序端、Web 管理端和后端 API 服务。

## 功能特性

- **志愿者管理** - 微信一键登录，组织架构（省/市/区/街道），RBAC 权限控制
- **走失人员数据库** - 登记、照片上传、轨迹跟踪、关联方言、状态流转
- **方言语音库** - 音频文件上传、区域标记、播放统计、点赞评论、精选管理
- **任务管理** - 创建分配、进度追踪、逾期自动检测、操作日志
- **文件存储** - 本地/OSS/COS 多后端，ClamAV 病毒扫描
- **短信服务** - 阿里云/腾讯云短信，验证码发送
- **审计日志** - 全量 API 请求记录，敏感操作监控，定时清理
- **数据库备份** - 定时自动备份（pg_dump/mysqldump），过期清理
- **数据统计** - 仪表盘、趋势分析、个人统计
- **Swagger API 文档** - 自动生成，67 个 API 端点完整覆盖

## 技术栈

| 模块 | 技术 |
|------|------|
| 后端 | Go 1.23+, Gin, GORM, PostgreSQL/MySQL, Redis, JWT |
| API 文档 | Swagger UI (swaggo/swag 自动生成) |
| 小程序 | 微信小程序原生开发 |
| Web | Next.js 14, React 18, TypeScript |

## 快速开始

### 1. 克隆项目

```bash
git clone https://github.com/Snowitty-Re/CNtunyuan.git
cd CNtunyuan
```

### 2. 初始化数据库并启动后端

```bash
cd backend

# 配置数据库 (config/config.yaml)
cp config/config.example.yaml config/config.yaml
# 编辑 config.yaml 填写数据库信息

# 首次初始化数据库（二选一）
# PostgreSQL:
psql -U postgres -d cntuanyuan -f migrations/postgres/00_bootstrap.sql
# MySQL:
mysql -u root -p cntuanyuan < migrations/mysql/00_bootstrap.sql

# 历史库升级（仅旧环境）
# PostgreSQL:
# psql -U postgres -d cntuanyuan -f migrations/postgres/06_schema_consistency_and_performance.sql
# psql -U postgres -d cntuanyuan -f migrations/postgres/07_dialect_schema_alignment.sql
# MySQL:
# mysql -u root -p cntuanyuan < migrations/mysql/06_schema_consistency_and_performance.sql
# mysql -u root -p cntuanyuan < migrations/mysql/07_dialect_schema_alignment.sql

# 启动服务
go run cmd/app/main.go
```

服务启动后访问：
- API 文档：http://localhost:8080/swagger/index.html
- 健康检查：http://localhost:8080/api/v1/health

### 3. 启动 Web 管理端（可选）

```bash
cd web
cp .env.example .env.local
npm install
npm run dev
```

访问：http://localhost:3000

### 4. 微信小程序

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
├── web/              # Web 管理端 (Next.js + TypeScript)
└── mini-program/     # 微信小程序
```

## API 模块

| 模块 | API 路径 | 说明 |
|------|---------|------|
| 认证管理 | `/auth/*` | 登录、注册、微信登录（小程序/Web 扫码）、Token 刷新、验证码、密码重置、微信绑定/解绑 |
| 用户管理 | `/users/*` | 用户 CRUD、角色/状态管理 |
| 个人中心 | `/profile/*` | 资料编辑、密码修改、个人统计 |
| 组织管理 | `/organizations/*` | 组织架构、树形结构、层级路径 |
| 案件管理 | `/missing-persons/*` | 走失人员登记、照片/轨迹、状态流转 |
| 任务管理 | `/tasks/*` | 任务创建分配、进度追踪、操作日志 |
| 方言管理 | `/dialects/*`、`/dialect-cards/*` | 方言批次录入、卡片模板管理、点赞评论、精选 |
| 文件管理 | `/upload/*` | 文件上传下载、实体绑定 |
| 仪表盘 | `/dashboard/*` | 数据统计、趋势分析 |
| 审计日志 | `/audit/*` | 操作日志、统计、清理 |

所有接口详情请访问 Swagger UI：http://localhost:8080/swagger/index.html

账号安全策略：新注册用户与微信新用户默认状态为待审批（`inactive`），需管理员启用后方可登录。

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
| [backend/README.md](backend/README.md) | 后端开发指南（含完整 API 端点列表） |
| [web/README.md](web/README.md) | Web 管理端说明 |
| [mini-program/README.md](mini-program/README.md) | 小程序开发说明 |
| [backend/REFACTORING.md](backend/REFACTORING.md) | Clean Architecture 设计文档 |
| [backend/TESTING.md](backend/TESTING.md) | 单元测试文档 |
| [backend/SWAGGER.md](backend/SWAGGER.md) | Swagger API 文档使用 |
| [backend/migrations/README.md](backend/migrations/README.md) | 数据库迁移指南 |
| [DEPLOY.md](DEPLOY.md) | 生产环境部署指南 |

## 默认账号

```
手机号: 13800138000
密码: admin123
```

## 许可证

MIT License
