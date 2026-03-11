# 更新日志 (Changelog)

本项目所有重大变更都将记录在此文件中。

## [2.0.0] - 2026-03-11

### Web 管理后台 (全新构建)

基于 React 18 + TypeScript + Vite + Ant Design 5 全新搭建生产级 Web 管理后台，覆盖后端全部 67+ API 端点。

#### 技术栈
- React 18 + TypeScript 5 + Vite 5
- Ant Design 5 (中文国际化)
- Zustand 4 (状态管理)
- React Router 6 (路由 + 懒加载)
- Axios (请求拦截 + Token 自动刷新)
- framer-motion (页面切换动画)

#### 项目结构 (78 个源文件)
- **类型层** (11 文件) - 完整镜像后端 DTO，覆盖 9 大模块
- **API 层** (10 文件) - Axios 实例 + 拦截器 + 67+ 端点封装
- **状态管理** - authStore (JWT 认证流) + appStore (UI 状态)
- **共享组件** (11 个) - PageContainer, StatusTag, RoleTag, UrgencyTag, ConfirmButton, StatCard, AudioPlayer, OrgTreeSelect, UserSelect, FileUpload, EmptyState
- **布局** - AdminLayout (侧边栏 + 头部 + 内容区) + BlankLayout

#### 页面模块 (25 个页面)
- **登录** - 管理员登录 + 忘记密码 (手机验证码两步流程)
- **工作台** - 统计卡片 + 近7日趋势 + 最近活动
- **用户管理** - 列表 (按角色/状态/组织筛选) + 详情 + 创建/编辑弹窗
- **组织管理** - 列表 + 交互式树形视图 + 创建/编辑弹窗
- **走失人员** - 列表 (丰富筛选器) + 详情 (轨迹时间线 + 标记找到/团圆) + 登记/编辑表单
- **任务管理** - 列表 (标签页：全部/我的/待处理/已逾期) + 详情 (分配/开始/完成/取消/进度) + 表单
- **方言库** - 列表 + 详情 (音频播放器 + 评论 + 点赞 + 精选) + 上传表单
- **文件管理** - 文件浏览器 (拖拽上传) + 存储统计
- **审计日志** - 日志列表 (综合筛选) + 详情 + 统计
- **个人资料** - 查看/编辑 + 修改密码
- **系统监控** - 系统健康仪表盘
- **错误页** - 403/404/500

#### 认证与权限
- JWT 双 Token 认证 (access_token + refresh_token)
- 401 响应自动刷新 Token 并重试请求
- 菜单按角色动态显示 (super_admin > admin > manager > volunteer)
- 路由级 AuthGuard + RoleGuard 权限守卫
- 按钮级权限控制 (usePermission hook)

#### 错误处理
- 后端业务错误码 (1000-6099) 中文映射
- Axios 响应拦截器统一处理 code 字段
- HTTP 错误友好提示

#### 构建优化
- 代码分割: vendor (React) + antd + 45+ 懒加载页面 chunk
- TypeScript 严格模式 0 错误
- 生产构建 5s 完成

### 文档更新
- README.md 新增 Web 管理后台技术栈、启动说明、项目结构
- CHANGELOG.md 新增 v2.0.0 变更记录
- .gitignore 新增 tsbuildinfo/d.ts.map 排除

## [1.1.0] - 2026-03-10

### 安全加固
- **移除 JWT Token 查询参数提取** - Token 仅从 Authorization Header 获取，防止 Token 泄露到日志/代理
- **CORS 白名单强化** - 移除 `Access-Control-Allow-Origin: *`，改为域名白名单机制
- **密码策略升级** - 最低密码长度从 6 位提升至 8 位，涵盖后端验证、DTO binding、前端表单

### 前端交付级优化 (25 个文件)

#### 新增组件
- **ErrorBoundary** - React 错误边界，防止页面崩溃白屏
- **集中常量管理** (`src/constants/index.ts`) - 统一 12 种状态/类型映射，消除 8+ 文件重复定义

#### 表单交互升级 (手动 ID 输入 → 智能选择器)
- **任务分配** - 执行人 ID 输入框 → 用户搜索下拉选择器
- **任务表单** - 关联案件 ID 输入框 → 案件搜索下拉选择器
- **用户表单** - 组织 ID 输入框 → TreeSelect 组织树选择器
- **组织表单** - 上级组织 ID 输入框 → TreeSelect 层级选择器
- **方言表单** - 手动音频 URL 输入框 → 文件上传组件 + 音频播放器预览
- **案件表单** - 新增照片上传组件 + 预览，分卡片布局（基本信息/走失信息/联系信息）

#### 页面完善
- **仪表盘** - 新增个性化问候、日期显示、"我的任务"列表、7 日趋势数据、可点击统计卡片
- **案件详情** - 新增照片画廊 (Image.PreviewGroup)、紧急程度中文标签、"标记团圆"确认对话框、关键线索标记
- **任务详情** - 新增"开始处理"确认对话框、逾期指示器、状态流转日志显示、开始/完成时间
- **个人中心** - 新增身份证号字段、头像展示卡片、账号信息区、邮箱格式验证
- **组织列表** - 新增组织类型筛选下拉
- **登录页** - 新增手机号正则验证 (`1[3-9]\d{9}`)、默认账号提示

#### 请求层优化
- **403 权限拦截** - 新增 403 Forbidden 友好提示"权限不足"
- **路由修复** - 修复"系统设置"菜单导航到不存在的 `/settings` 路由

### 文档更新
- README.md 功能模块表格更新
- AGENTS.md 修复 `web-new` → `web` 引用
- CHANGELOG.md 新增本次变更记录

## [1.0.0] - 2026-03-09

### 新增功能

### 1. 关键 Bug 修复 ✅

#### 1.1 微信登录临时用户缺少 ID
- **问题**：创建临时用户时未设置 ID，导致数据库插入失败
- **修复**：在 `auth_service.go` 中添加 `BaseEntity{ID: uuid.New().String()}`

#### 1.2 任务创建 Deadline 空指针问题
- **问题**：直接将 `req.Deadline` 地址赋值给指针字段，当为零值时导致空指针
- **修复**：添加零值检查后再赋值

### 2. 操作日志审计系统 ✅

**文件位置**：
- `backend/internal/domain/entity/audit_log.go` - 审计日志实体
- `backend/internal/domain/repository/audit_log_repository.go` - 仓储接口
- `backend/internal/infrastructure/repository/audit_log_repository.go` - 仓储实现
- `backend/internal/interfaces/http/middleware/audit.go` - 审计中间件
- `backend/internal/interfaces/http/handler/audit_handler.go` - API处理器
- `backend/internal/application/service/audit_service.go` - 应用服务

**功能特性**：
- 自动记录所有 API 请求
- 敏感信息脱敏（密码、手机号、身份证号）
- 按模块、操作类型、时间查询
- 用户活动统计
- 模块访问统计
- 定期清理旧日志（保留90天）

**API 端点**：
```
GET    /api/v1/audit/logs           # 查询审计日志
GET    /api/v1/audit/logs/:id       # 获取日志详情
GET    /api/v1/audit/stats          # 获取统计信息
GET    /api/v1/audit/user-activity/:userId  # 用户活动统计
GET    /api/v1/audit/module-stats   # 模块统计
POST   /api/v1/audit/cleanup        # 清理旧日志（管理员）
```

### 3. 定时任务系统 ✅

**文件位置**：
- `backend/internal/task/scheduler.go` - 定时任务调度器

**定时任务**：
- **每分钟**：检查逾期任务并自动标记
- **每小时**：更新统计数据
- **每天凌晨3点**：清理90天前的审计日志
- **每天凌晨2点**：数据备份

**逾期任务处理**：
- 自动检测已逾期但未标记的任务
- 更新任务状态为 "overdue"
- 记录操作日志

### 4. 文件上传安全检查 ✅

**文件位置**：
- `backend/internal/infrastructure/storage/security.go` - 文件安全检查
- `backend/internal/infrastructure/storage/scanner.go` - 病毒扫描接口

**安全功能**：
- 文件扩展名检查
- MIME类型检测（通过文件头）
- 文件头签名验证（防止伪造扩展名）
  - JPEG: FF D8
  - PNG: 89 50 4E 47 0D 0A 1A 0A
  - GIF: GIF87a/GIF89a
  - PDF: %PDF-
  - MP3: ID3标签或同步字
- 文件名安全检查（移除危险字符）
- 敏感信息脱敏（密码、手机号、身份证）

**病毒扫描**：
- 支持接口扩展（预留 ClamAV 集成）
- 默认使用空扫描器（开发模式）

### 5. 云存储支持 ✅

#### 5.1 阿里云 OSS
**文件位置**：`backend/internal/infrastructure/storage/oss_storage.go`

**特性**：
- 标准存储和低频存储支持
- 临时 URL 生成
- 分片上传支持（预留）
- CORS 配置

**启用方式**：
```bash
# 安装依赖后构建
go get github.com/aliyun/aliyun-oss-go-sdk/oss
go build -tags oss ./cmd/app/main.go
```

#### 5.2 腾讯云 COS
**文件位置**：`backend/internal/infrastructure/storage/cos_storage.go`

**特性**：
- 预签名 URL 生成
- 分片上传支持
- CORS 配置

**启用方式**：
```bash
# 安装依赖后构建
go get github.com/tencentyun/cos-go-sdk-v5
go build -tags cos ./cmd/app/main.go
```

#### 5.3 存储降级
当 OSS/COS 配置错误或不可用时，自动降级到本地存储，确保服务可用性。

### 6. 短信验证码服务 ✅

**文件位置**：
- `backend/internal/infrastructure/sms/sms.go` - 短信服务接口
- `backend/internal/infrastructure/sms/aliyun.go` - 阿里云短信
- `backend/internal/infrastructure/sms/tencent.go` - 腾讯云短信

**功能**：
- 支持阿里云和腾讯云
- 开发模式模拟发送（验证码：123456）
- 验证码5分钟有效期
- 自动缓存验证

### 7. 健康检查完善 ✅

**文件位置**：`backend/internal/application/service/health_service.go`

**检查项**：
- 数据库连接状态
- 连接池使用率监控
- Redis 缓存状态
- 系统资源（Goroutine、内存、GC）

### 8. 配置验证系统 ✅

**文件位置**：`backend/internal/config/validator.go`

**验证项**：
- 服务器配置（端口、模式）
- 数据库配置
- JWT 密钥强度（检查弱密钥）
- 存储配置完整性
- 微信登录配置

### 9. 生产环境部署配置 ✅

**文件位置**：
- `backend/Dockerfile` - 多阶段构建镜像
- `docker/docker-compose.prod.yml` - 生产编排配置
- `docker/.env.example` - 环境变量示例
- `DEPLOY.md` - 部署指南

**Docker 镜像特性**：
- 基于 Scratch 的最小镜像
- 静态链接
- 内置健康检查
- 自动构建脚本

## API 路由总览

```
/api/v1/
├── /                    # 欢迎信息
├── /health              # 健康检查
├── /health/detailed     # 详细健康检查
├── /metrics             # Prometheus 指标
├── /auth
│   ├── /login           # 登录
│   ├── /admin-login     # 管理员登录
│   ├── /refresh         # 刷新 Token
│   ├── /logout          # 登出
│   ├── /wechat-login    # 微信登录
│   ├── /bind-phone      # 绑定手机号
│   ├── /send-code       # 发送验证码
│   └── /me              # 当前用户信息
├── /users               # 用户管理
├── /organizations       # 组织管理
├── /missing-persons     # 走失人员管理
├── /tasks               # 任务管理
├── /dialects            # 方言管理
├── /upload              # 文件上传
├── /dashboard           # 仪表盘统计
└── /audit               # 审计日志
```

## 安全特性总结

| 功能 | 状态 | 说明 |
|------|------|------|
| JWT 认证 | ✅ | HS256 签名，支持刷新 Token |
| 密码加密 | ✅ | bcrypt 哈希 |
| 请求限流 | ✅ | IP 级别限流（100r/s） |
| CORS 控制 | ✅ | 可配置允许的域名 |
| 安全响应头 | ✅ | CSP, X-Frame-Options 等 |
| 敏感数据脱敏 | ✅ | 日志中自动脱敏 |
| 文件上传检查 | ✅ | MIME, 文件头, 文件名检查 |
| 操作审计 | ✅ | 全量 API 请求记录 |
| SQL 注入防护 | ✅ | GORM 参数化查询 |
| XSS 防护 | ✅ | 输入验证和输出编码 |

## 性能优化

| 优化项 | 实现 |
|--------|------|
| 数据库连接池 | 连接池监控，自动调整 |
| 慢查询检测 | 超过 100ms 记录警告 |
| Prometheus 监控 | HTTP/DB 指标收集 |
| 异步审计日志 | 不阻塞主请求 |
| 缓存支持 | Redis 集成（可选） |

## 部署检查清单

在上线前，请确认：

- [ ] 执行 `go run cmd/app/main.go -check-db` 检查数据库
- [ ] JWT 密钥已修改为随机强密钥（≥32位）
- [ ] 数据库密码已修改
- [ ] 微信小程序 AppID/AppSecret 已配置（如使用微信登录）
- [ ] 短信服务已配置（如使用短信验证码）
- [ ] OSS/COS 密钥已配置（如使用云存储）
- [ ] 配置验证通过
- [ ] 健康检查端点正常工作
- [ ] 数据库迁移已完成
- [ ] 定时任务已启动
- [ ] 日志轮转已配置
- [ ] HTTPS 已启用

## 启动日志示例

```
✓ Configuration validation passed
Database connected successfully
Redis connected successfully
Using local storage
Task scheduler started
SMS service running in DEV mode
Server starting on :8080
```

## 后续建议

1. **监控告警**：集成 Prometheus + Grafana 监控
2. **日志收集**：使用 ELK 或 Loki 集中收集日志
3. **分布式追踪**：集成 Jaeger 进行链路追踪
4. **单元测试**：补充核心业务的单元测试
5. **接口文档**：完善 Swagger 文档
