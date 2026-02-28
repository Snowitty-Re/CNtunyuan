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
- React + TypeScript + Ant Design
- 大屏数据展示
- 完整的CRUD操作

## 技术栈

### 后端
- Go 1.23+
- Gin 框架
- GORM
- PostgreSQL 16
- Redis 7
- JWT 认证

### Web管理端
- React 18
- TypeScript
- Ant Design 5
- Vite
- Zustand 状态管理

### 微信小程序
- 原生微信小程序
- 微信云开发

## 项目结构

```
CNtunyuan/
├── backend/              # Go 后端服务
│   ├── cmd/              # 主程序入口
│   │   ├── main.go       # 主程序
│   │   └── initdata/     # 数据初始化工具
│   ├── internal/         # 内部包
│   │   ├── api/          # API 处理器
│   │   ├── config/       # 配置
│   │   ├── middleware/   # 中间件
│   │   ├── model/        # 数据模型
│   │   ├── repository/   # 数据访问层
│   │   ├── service/      # 业务逻辑层
│   │   └── utils/        # 工具函数
│   ├── pkg/              # 公共包
│   ├── sql/              # SQL初始化脚本
│   ├── config/           # 配置文件
│   └── Dockerfile
├── web-admin/            # Web 管理后台
│   ├── src/
│   │   ├── components/   # 组件
│   │   ├── pages/        # 页面
│   │   ├── services/     # API 服务
│   │   ├── stores/       # 状态管理
│   │   └── utils/        # 工具函数
│   └── package.json
├── mini-program/         # 微信小程序
│   ├── pages/            # 页面
│   ├── components/       # 组件
│   └── utils/            # 工具函数
├── docker/               # Docker 配置
└── docs/                 # 文档
```

## 快速开始

### 环境要求
- Go 1.23+
- Node.js 18+
- PostgreSQL 16
- Redis 7 (可选)
- 微信小程序开发者工具

### 数据库初始化

```bash
cd backend

# 1. 执行数据库迁移（创建表结构）
go run cmd/main.go -migrate

# 2. 初始化基础数据（根组织）
go run cmd/main.go -init

# 3. 创建超级管理员（可自定义参数）
go run cmd/initdata/main.go -exec

# 或使用自定义参数
go run cmd/initdata/main.go -exec -phone="13800138000" -password="admin123" -email="admin@cntunyuan.com"
```

更多初始化方式请参考 [backend/sql/README.md](backend/sql/README.md)

### 后端启动

```bash
cd backend

# 安装依赖
go mod download

# 配置数据库（修改 config/config.yaml）
# 或使用默认配置

# 运行
go run cmd/main.go

# 或使用 air 热重载
air
```

### Web 后台启动

```bash
cd web-admin

# 安装依赖
pnpm install

# 启动开发服务器
pnpm dev
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

## API 文档

启动后端服务后，访问 `http://localhost:8080/swagger/index.html` 查看 Swagger API 文档。

## 主要接口

### 认证
- POST /api/v1/auth/wechat-login - 微信登录
- POST /api/v1/auth/admin-login - 管理后台登录
- POST /api/v1/auth/refresh - 刷新 Token
- GET /api/v1/auth/me - 获取当前用户

### 用户管理
- GET /api/v1/users - 用户列表
- GET /api/v1/users/:id - 用户详情
- PUT /api/v1/users/:id - 更新用户
- DELETE /api/v1/users/:id - 删除用户

### 组织架构
- GET /api/v1/organizations - 组织列表
- GET /api/v1/organizations/tree - 组织树
- POST /api/v1/organizations - 创建组织

### 走失人员
- GET /api/v1/missing-persons - 案件列表
- POST /api/v1/missing-persons - 创建案件
- GET /api/v1/missing-persons/:id - 案件详情
- PUT /api/v1/missing-persons/:id/status - 更新状态

### 方言管理
- GET /api/v1/dialects - 方言列表
- POST /api/v1/dialects - 创建方言
- GET /api/v1/dialects/:id - 方言详情
- POST /api/v1/dialects/:id/play - 播放记录

### 任务管理
- GET /api/v1/tasks - 任务列表
- POST /api/v1/tasks - 创建任务
- GET /api/v1/tasks/:id - 任务详情
- POST /api/v1/tasks/:id/assign - 分配任务
- POST /api/v1/tasks/:id/complete - 完成任务
- POST /api/v1/tasks/:id/cancel - 取消任务
- POST /api/v1/tasks/batch-assign - 批量分配
- POST /api/v1/tasks/auto-assign - 自动分配

### 工作流管理
- GET /api/v1/workflows - 工作流列表
- POST /api/v1/workflows - 创建工作流
- GET /api/v1/workflows/:id - 工作流详情
- POST /api/v1/workflows/:id/steps - 创建步骤
- PUT /api/v1/workflows/:id/steps/:step_id - 更新步骤

### 工作流实例
- GET /api/v1/workflow-instances - 实例列表
- POST /api/v1/workflow-instances - 启动实例
- POST /api/v1/workflow-instances/:id/approve - 审批
- GET /api/v1/workflow-instances/:id/history - 审批历史

## 默认账号

初始化后默认超级管理员账号：
- 手机号: `13800138000`
- 密码: `admin123`
- 角色: super_admin

## 开发指南

### 后端开发

```bash
cd backend

# 生成 Swagger 文档
swag init -g cmd/main.go

# 运行测试
go test ./...

# 格式化代码
go fmt ./...
```

### 前端开发

```bash
cd web-admin

# 代码检查
pnpm lint

# 构建生产版本
pnpm build
```

## 贡献

欢迎提交 Issue 和 Pull Request。

## 许可证

MIT License

## 联系方式

- 项目地址: https://github.com/Snowitty-Re/CNtunyuan
