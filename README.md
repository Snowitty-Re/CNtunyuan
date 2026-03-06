# 团圆寻亲志愿者系统 (CNtunyuan)

一个生产级别的寻亲志愿者管理系统，包含微信小程序端、Web管理后台端和后端API服务。

## 项目概述

本项目旨在帮助寻找走失人员，通过志愿者网络、方言语音数据库和OA工作流系统，提高寻人效率。

## 功能特性

### 志愿者管理
- 微信一键登录
- 组织架构：团圆机构 -> 省 -> 市 -> 区 -> 街道
- 角色权限：超级管理员、管理员、管理者、志愿者
- 完整的权限控制系统

### 走失人员数据库
- 走失人员登记
- 地图定位标记
- 关联方言语音数据库
- 关联任务创建
- 轨迹记录跟踪

### 方言语音数据库
- 15-20秒语音上传
- 区域标记
- 标签分类
- 备注说明
- 关联走失人员

### 文件存储
- 支持本地存储、阿里云 OSS、腾讯云 COS
- 单文件/批量上传
- 文件类型检查（图片/音频/视频/文档）
- 自动按日期目录组织

### 任务管理
- 任务创建与分配
- 任务转派
- 进度追踪
- 批量分配
- 自动分配（基于负载均衡）
- 任务评论
- 操作日志

### 工作流管理
- 工作流定义
- 步骤管理（审批节点）
- 工作流实例管理
- 审批流程（通过/驳回/转派）
- 审批历史追踪

### 数据展示
- 寻亲数据统计
- 志愿者工作台
- 管理者快速分配任务
- 数据大屏展示

### Web后台管理
- React 18 + TypeScript + Ant Design 5
- 简洁办公OA风格设计
- 温馨的橙色主题
- 完整的CRUD操作

## 技术栈

### 后端
- Go 1.24+
- Gin 框架
- GORM
- PostgreSQL 16 / MySQL 8.0
- Redis 7 (可选)
- JWT 认证
- Clean Architecture

### Web管理端 (web-new)
- React 18
- TypeScript 5
- Ant Design 5
- Vite 5
- Zustand 状态管理
- Axios

### 微信小程序
- 原生微信小程序
- 微信云开发

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
│   ├── sql/              # SQL初始化脚本
│   ├── uploads/          # 本地文件存储目录
│   └── Dockerfile
├── web-new/              # Web 管理后台（新版）
│   ├── src/
│   │   ├── components/   # 组件
│   │   ├── pages/        # 页面
│   │   ├── router/       # 路由配置
│   │   ├── services/     # API 服务
│   │   ├── stores/       # 状态管理
│   │   ├── utils/        # 工具函数
│   │   └── types/        # TypeScript 类型
│   ├── package.json
│   └── vite.config.ts
├── mini-program/         # 微信小程序
│   ├── pages/            # 页面
│   ├── components/       # 组件
│   └── utils/            # 工具函数
├── docker/               # Docker 配置
└── docs/                 # 文档
```

## 快速开始

### 环境要求
- Go 1.24+
- Node.js 18+
- PostgreSQL 16 或 MySQL 8.0
- Redis 7 (可选)
- 微信小程序开发者工具

### 配置说明

#### 后端配置 (backend/config/config.yaml)
```yaml
# 服务器配置
server:
  port: 8080
  mode: debug  # debug/release
  domain: "http://localhost:8080"

# 数据库配置（支持 PostgreSQL/MySQL）
database:
  type: "postgres"  # 或 mysql
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "your-password"
  database: "cntuanyuan"

# 微信小程序配置（用于小程序登录）
wechat:
  app_id: "your-wechat-app-id"
  app_secret: "your-wechat-app-secret"
  enable_login: true

# JWT 配置（生产环境必须修改密钥）
jwt:
  secret: "your-jwt-secret-key"
  expire_time: 604800  # 7天
```

#### 前端配置 (web-new/.env.development)
```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_APP_DOMAIN=http://localhost:8080
```

### 系统初始化

```bash
cd backend

# 数据库迁移（自动创建表结构）
go run cmd/app/main.go -migrate

# 数据填充（可选，用于测试数据）
go run cmd/seed/main.go -all
```

更多初始化方式请参考 [backend/ARCHITECTURE.md](backend/ARCHITECTURE.md)

### 后端启动

```bash
cd backend

# 安装依赖
go mod download

# 配置（修改 config/config.yaml）
# - 数据库连接
# - JWT 密钥（生产环境必须修改）
# - 微信小程序 AppID 和 AppSecret（如使用微信登录）

# 运行（统一入口 cmd/app/main.go）
go run cmd/app/main.go

# 数据库迁移
go run cmd/app/main.go -migrate

# 或使用 air 热重载
air
```

### Web 后台启动（新版）

```bash
cd web-new

# 安装依赖
pnpm install

# 启动开发服务器
pnpm dev

# 访问 http://localhost:3000
```

### Docker 部署

```bash
cd docker
docker-compose up -d
```

### 微信小程序

1. 打开微信开发者工具
2. 导入 `mini-program` 目录
3. 配置 appid
4. 编译运行

#### 小程序图标系统

小程序使用 **Emoji 图标系统**，无需外部字体文件：

- 图标定义在 `mini-program/assets/styles/icons.wxss`
- 使用方式：`<text class="iconfont icon-add"></text>`
- 支持 40+ 常用图标映射

详见 [mini-program/assets/icons/README.md](mini-program/assets/icons/README.md)

## API 文档

启动后端服务后，访问 `http://localhost:8080/swagger/index.html` 查看 Swagger API 文档。

## 默认账号

种子数据会自动创建超级管理员账号：

- **手机号**: 13800138000
- **密码**: admin123

使用种子数据导入：
```bash
cd backend

# 导入所有种子数据（包含超级管理员、组织、测试数据）
go run cmd/seed/main.go -all

# 或单独导入特定数据
go run cmd/seed/main.go -orgs     # 只导入组织
go run cmd/seed/main.go -users    # 只导入用户
```

### 重置密码

```bash
cd backend
go run cmd/resetpassword/main.go -phone=13800138000 -password=newpassword
```

## 开发指南

### 后端开发

```bash
cd backend

# 格式化代码
go fmt ./...

# 运行测试
go test ./...
```

### 前端开发

```bash
cd web-new

# 代码检查
pnpm lint

# 构建生产版本
pnpm build

# 预览生产版本
pnpm preview
```

### Web 前端设计规范

Web 前端采用简洁办公OA风格，温馨配色：

| 用途 | 色值 |
|------|------|
| 主色 | `#e67e22` |
| 背景色 | `#f5f7fa` |
| 主文字 | `#1f2329` |
| 次要文字 | `#646a73` |
| 边框色 | `#e8e9eb` |

详见 [web-new/README.md](web-new/README.md)

## 架构文档

- [backend/ARCHITECTURE.md](backend/ARCHITECTURE.md) - 后端 Clean Architecture 架构说明
- [backend/REFACTORING.md](backend/REFACTORING.md) - 架构重构详细指南
- [AGENTS.md](AGENTS.md) - AI 助手和开发者指南

## 贡献

欢迎提交 Issue 和 Pull Request。

## 许可证

MIT License

## 联系方式

- 项目地址: https://github.com/Snowitty-Re/CNtunyuan
