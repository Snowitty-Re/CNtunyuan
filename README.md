# 团圆寻亲志愿者系统 (CNtunyuan)

一个生产级别的寻亲志愿者管理系统，包含微信小程序端、Web管理后台和后端API服务。

## 功能特性

- **志愿者管理** - 微信一键登录，组织架构（省/市/区/街道），RBAC权限控制
- **走失人员数据库** - 登记、照片上传、轨迹跟踪、关联方言、状态流转
- **方言语音库** - 音频文件上传、区域标记、播放统计、精选管理
- **任务管理** - 创建分配、人员选择器、进度追踪、逾期检测、操作日志
- **文件存储** - 本地/OSS/COS，安全检查，图片和音频上传
- **审计日志** - 全量API请求记录，敏感操作监控，日志清理
- **数据统计** - 仪表盘、趋势分析、我的任务、实时统计
- **组织树管理** - 层级选择器、树形结构、上下级关联

## 技术栈

| 模块 | 技术 |
|------|------|
| 后端 | Go 1.23+, Gin, GORM, PostgreSQL, Redis, JWT |
| Web 前端 | React 18, TypeScript 5, Vite 5, Ant Design 5, Zustand |
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

服务启动后访问：http://localhost:8080/swagger/index.html

### 3. 启动 Web 后台

```bash
cd web
npm install
npm run dev
```

访问 http://localhost:3000（自动代理 API 到 :8080），默认账号：13800138000 / admin123

### 4. 微信小程序

使用微信开发者工具打开 `mini-program` 目录。

## 项目结构

```
CNtunyuan/
├── backend/          # Go 后端 (Clean Architecture / DDD)
│   ├── cmd/app/      # 主应用入口
│   ├── internal/     # 业务代码 (domain/application/infrastructure/interfaces)
│   ├── pkg/          # 公共包 (logger/response/middleware)
│   └── migrations/   # 数据库迁移
├── web/              # Web 管理后台 (React 18 + TypeScript + Ant Design 5)
│   ├── src/api/      # API 层 (Axios + 类型安全 + Token 刷新)
│   ├── src/pages/    # 页面组件 (仪表盘/案件/任务/用户/组织/方言/审计/个人中心)
│   ├── src/store/    # 状态管理 (Zustand)
│   ├── src/constants/# 集中常量 (状态映射/验证规则)
│   ├── src/components/# 公共组件 (ErrorBoundary/RouteGuard)
│   └── src/types/    # TypeScript 类型 (与后端 DTO 严格对齐)
├── mini-program/     # 微信小程序
├── docker/           # Docker 部署配置
└── scripts/          # 运维脚本
```

## 功能模块

| 模块 | 前端页面 | 后端 API |
|------|---------|---------|
| 仪表盘 | 统计卡片、趋势数据、我的任务、快捷操作 | `/dashboard/stats`, `/dashboard/trend` |
| 案件管理 | 列表筛选、详情(照片/轨迹)、表单(照片上传)、状态流转 | `/missing-persons/*` |
| 任务管理 | 全部/我的/待分配、详情、人员分配器、进度条、日志 | `/tasks/*` |
| 用户管理 | 列表、表单(组织树选择器)、角色/状态管理 | `/users/*` |
| 组织管理 | 列表(类型筛选)、表单(树形父级选择)、层级结构 | `/organizations/*` |
| 方言管理 | 列表、表单(音频上传/播放)、精选/审核 | `/dialects/*` |
| 文件管理 | 上传、下载、实体绑定 | `/upload/*` |
| 审计日志 | 日志列表、多维筛选、展开详情、日志清理 | `/audit/*` |
| 个人中心 | 头像展示、资料编辑(含身份证)、密码修改、账号信息 | `/profile/*` |

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
- Clean Architecture 分层架构 (Domain → Application → Infrastructure → Interfaces)
- 领域驱动设计 (DDD)，充血实体模型
- 单元测试覆盖核心逻辑 (110+ 测试函数)

### Web 前端
- 简洁办公 OA 风格，主题色 #e67e22（温暖橙色）
- Ant Design 5 组件库 + 内联样式
- 所有 TypeScript 类型与后端 DTO 严格一一对应
- 路由懒加载 + 代码分割 + React Error Boundary
- 集中常量管理 (`src/constants/`)，消除重复定义
- 表单增强：TreeSelect 组织选择器、用户搜索下拉、文件上传组件

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
