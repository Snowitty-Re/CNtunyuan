# 团圆寻亲志愿者系统 (CNtunyuan)

一个生产级别的寻亲志愿者管理系统，包含微信小程序端、Web管理后台和后端API服务。

## 功能特性

- **志愿者管理** - 微信一键登录，组织架构（省/市/区/街道），RBAC权限控制
- **走失人员数据库** - 登记、地图定位、轨迹跟踪、关联方言
- **方言语音库** - 15-20秒语音上传、区域标记、播放统计
- **任务管理** - 创建分配、进度追踪、逾期检测、操作日志
- **文件存储** - 本地/OSS/COS，安全检查，批量上传
- **审计日志** - 全量API请求记录，敏感操作监控
- **数据统计** - 仪表盘、趋势分析、实时统计

## 技术栈

| 模块 | 技术 |
|------|------|
| 后端 | Go 1.24+, Gin, GORM, PostgreSQL/MySQL, Redis |
| Web | React 18, TypeScript, Ant Design 5, Vite |
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
cd web-new
pnpm install
pnpm dev
```

访问 http://localhost:3000，默认账号：13800138000 / admin123

### 4. 微信小程序

使用微信开发者工具打开 `mini-program` 目录。

## 项目结构

```
CNtunyuan/
├── backend/          # Go 后端 (Clean Architecture)
│   ├── cmd/app/      # 主应用入口
│   ├── internal/     # 业务代码
│   ├── pkg/          # 公共包
│   └── migrations/   # 数据库迁移
├── web-new/          # Web 管理后台
├── mini-program/     # 微信小程序
└── docker/           # Docker 配置
```

## 文档

| 文档 | 说明 |
|------|------|
| [DEPLOY.md](DEPLOY.md) | 生产环境部署指南 |
| [CHANGELOG.md](CHANGELOG.md) | 版本更新记录 |
| [backend/README.md](backend/README.md) | 后端开发指南 |
| [backend/REFACTORING.md](backend/REFACTORING.md) | 架构设计文档 |
| [backend/TESTING.md](backend/TESTING.md) | 测试文档 |
| [backend/SWAGGER.md](backend/SWAGGER.md) | Swagger API 文档使用 |
| [backend/migrations/README.md](backend/migrations/README.md) | 数据库迁移指南 |
| [web-new/README.md](web-new/README.md) | Web 前端说明 |
| [mini-program/README.md](mini-program/README.md) | 小程序说明 |
| [AGENTS.md](AGENTS.md) | AI 助手开发指南 |

## 开发规范

### 后端
- Clean Architecture 分层架构
- 领域驱动设计 (DDD)
- 使用 `gofmt` 格式化代码
- 单元测试覆盖核心逻辑

### Web 前端
- 简洁办公 OA 风格
- 不使用 Tailwind，使用 Ant Design + 内联样式
- 主题色：#e67e22（温暖橙色）

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

详细部署说明参见 [DEPLOY.md](DEPLOY.md)。

## 许可证

MIT License

## 联系方式

- 项目地址: https://github.com/Snowitty-Re/CNtunyuan
