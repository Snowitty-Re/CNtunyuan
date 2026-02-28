# CNtunyuan - 开发指南

本文档为 AI 助手和开发者提供项目背景信息和开发规范。

## 项目背景

团圆寻亲志愿者系统是一个帮助寻找走失人员的公益项目，通过整合志愿者网络、方言语音数据库和工作流系统，提高寻人效率。

### 核心价值
- **志愿者协作**: 组织架构化的志愿者管理
- **方言辅助**: 通过方言语音帮助确认走失人员身份
- **任务驱动**: OA工作流确保寻人任务有序进行

## 技术架构

### 后端架构
```
┌─────────────┐
│   API 层    │  (api/)
├─────────────┤
│  Service 层 │  (service/)
├─────────────┤
│ Repository层│  (repository/)
├─────────────┤
│   Model 层  │  (model/)
└─────────────┘
```

### 前端架构
```
┌─────────────┐
│   页面层    │  (pages/)
├─────────────┤
│  服务层     │  (services/)
├─────────────┤
│  状态层     │  (stores/)
├─────────────┤
│  组件层     │  (components/)
└─────────────┘
```

## 开发规范

### 代码规范

#### Go 后端
- 使用 `gofmt` 格式化代码
- 遵循 Go 官方代码规范
- 函数命名使用驼峰式
- 接口命名使用动词+名词，如 `CreateUser`
- 错误处理必须返回具体错误信息

#### React 前端
- 使用 TypeScript 严格模式
- 组件命名使用大驼峰式
- Props 必须定义类型
- 使用 hooks 进行状态管理

### 数据库规范

#### 表命名
- 使用前缀 `ty_`
- 复数形式，如 `users`, `organizations`
- 关联表使用 `_` 连接，如 `user_roles`

#### 字段命名
- 使用下划线命名法
- 常用字段: `created_at`, `updated_at`, `deleted_at`
- 外键使用 `_id` 后缀
- 布尔值使用 `is_` 前缀

#### 索引规范
- 外键自动创建索引
- 常用查询字段添加索引
- 唯一索引使用 `uniqueIndex` 标签

### API 设计规范

#### RESTful API
```
GET    /api/v1/resources      # 列表
POST   /api/v1/resources      # 创建
GET    /api/v1/resources/:id   # 详情
PUT    /api/v1/resources/:id   # 更新
DELETE /api/v1/resources/:id   # 删除
```

#### 响应格式
```json
{
  "code": 200,
  "message": "success",
  "data": {}
}
```

#### 错误码
- 200: 成功
- 400: 参数错误
- 401: 未授权
- 403: 禁止访问
- 404: 资源不存在
- 500: 服务器错误

## 功能模块

### 1. 用户管理模块
- 文件位置: `internal/service/user.go`, `internal/api/user.go`
- 功能: 用户CRUD、角色管理、组织关联
- 角色: super_admin > admin > manager > volunteer

### 2. 组织管理模块
- 文件位置: `internal/service/organization.go`
- 功能: 组织架构树、层级管理
- 类型: root > province > city > district > street

### 3. 走失人员模块
- 文件位置: `internal/service/missing_person.go`
- 功能: 案件管理、状态流转、轨迹记录
- 状态: missing -> searching -> found -> reunited/closed

### 4. 方言管理模块
- 文件位置: `internal/service/dialect.go`
- 功能: 语音上传、区域标记、播放统计
- 要求: 15-20秒语音

### 5. 任务管理模块
- 文件位置: `internal/service/task.go`
- 功能: 任务分配、进度追踪、自动分配
- 状态: draft -> pending -> assigned -> processing -> completed/cancelled

### 6. 工作流模块
- 文件位置: `internal/service/workflow.go`
- 功能: 流程定义、审批节点、实例管理
- 状态: draft -> active/inactive

## 常用命令

### 后端命令

```bash
# 开发模式启动
cd backend && go run cmd/main.go

# 数据库迁移
cd backend && go run cmd/main.go -migrate

# 初始化数据
cd backend && go run cmd/initdata/main.go -exec

# 重置密码
cd backend && go run cmd/reset_password.go -phone=13800138000 -password=newpassword

# 生成 Swagger
cd backend && swag init -g cmd/main.go
```

### 前端命令

```bash
# 安装依赖
cd web-admin && pnpm install

# 开发模式
cd web-admin && pnpm dev

# 构建
cd web-admin && pnpm build

# 预览
cd web-admin && pnpm preview
```

## 数据初始化

### 1. 数据库迁移
创建表结构，不插入数据：
```bash
cd backend && go run cmd/main.go -migrate
```

### 2. 基础数据初始化
创建根组织：
```bash
cd backend && go run cmd/main.go -init
```

### 3. 超级管理员创建
使用命令行工具创建：
```bash
cd backend && go run cmd/initdata/main.go -exec
```

参数说明:
- `-phone`: 手机号 (默认: 13800138000)
- `-password`: 密码 (默认: admin123)
- `-email`: 邮箱 (默认: admin@cntunyuan.com)
- `-gen`: 仅生成SQL文件
- `-exec`: 直接执行初始化

更多详情参考 [backend/sql/README.md](backend/sql/README.md)

## 配置说明

### 后端配置 (config/config.yaml)

```yaml
server:
  port: "8080"
  mode: "debug"  # debug/release

database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "postgres"
  database: "cntuanyuan"
  ssl_mode: "disable"

redis:
  host: ""       # 空表示不使用Redis
  port: 6379
  password: ""
  db: 0

jwt:
  secret: "your-secret-key"
  expire_time: 604800  # 7天

wechat:
  app_id: "your-app-id"
  app_secret: "your-app-secret"
```

### 前端配置 (.env)

```env
VITE_API_BASE_URL=/api/v1
```

## 注意事项

1. **Redis 可选**: 如果 Redis 未配置，系统会自动使用内存缓存
2. **数据库表前缀**: 所有表使用 `ty_` 前缀
3. **外键约束**: 迁移时禁用外键约束，生产环境建议启用
4. **JWT 密钥**: 生产环境必须修改默认密钥
5. **微信小程序**: 需要配置正确的 appid 和密钥

## 常见问题

### 1. 数据库连接失败
- 检查 PostgreSQL 是否启动
- 检查 config.yaml 中的数据库配置
- 确认数据库 `cntuanyuan` 已创建

### 2. 登录失败
- 确认已执行初始化命令创建超级管理员
- 检查密码是否正确
- 查看后端日志确认错误信息

### 3. 前端代理问题
- 检查 vite.config.ts 中的代理配置
- 确认后端服务已启动
- 检查端口号是否正确

## 更新日志

### 2024-02-27
- 实现完整的任务管理功能（分配、转派、完成、取消）
- 实现工作流管理功能（定义、步骤、审批）
- 重构数据初始化方式，支持SQL导入
- 完善前端任务管理和工作流页面

## 待办事项

- [ ] 完善小程序端功能
- [ ] 实现数据大屏页面
- [ ] 添加更多测试数据
- [ ] 完善权限控制
- [ ] 添加操作日志审计
