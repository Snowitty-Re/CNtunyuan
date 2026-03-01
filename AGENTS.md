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

#### React 前端 (web-new)
- 使用 TypeScript 严格模式
- 组件命名使用大驼峰式
- Props 必须定义类型
- 使用 hooks 进行状态管理
- **样式规范**: 不使用 Tailwind className，使用 Ant Design 组件默认样式 + 内联 style
- **颜色规范**: 使用温馨橙色主题 (`#e67e22`)，背景 `#f5f7fa`，主文字 `#1f2329`
- **设计原则**: 简洁办公OA风格，去除多余装饰，注重信息层级

### 前端项目结构 (web-new)
```
web-new/
├── src/
│   ├── components/
│   │   ├── layout/         # 布局组件
│   │   │   ├── MainLayout.tsx
│   │   │   └── Sidebar.tsx
│   │   ├── common/         # 通用组件
│   │   └── ui/             # UI 组件
│   ├── pages/              # 页面
│   │   ├── login/
│   │   ├── dashboard/      # 工作台
│   │   ├── cases/          # 寻人案件
│   │   ├── tasks/          # 任务管理
│   │   ├── volunteers/     # 志愿者管理
│   │   ├── organizations/  # 组织架构
│   │   └── dialects/       # 方言管理
│   ├── router/             # 路由配置
│   ├── services/           # API 服务
│   ├── stores/             # 状态管理 (Zustand)
│   ├── utils/              # 工具函数
│   └── types/              # TypeScript 类型
├── package.json
└── vite.config.ts
```

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
- 复合索引遵循最左前缀原则
- 全文搜索索引使用 PostgreSQL 的 GIN 索引

#### 外键约束
- 启用外键约束（生产环境）
- 删除策略：
  - `CASCADE` - 级联删除（子表数据）
  - `SET NULL` - 设为 NULL（可选关联）
  - `RESTRICT` - 限制删除（有关联数据时禁止删除）
- 更新策略：`ON UPDATE CASCADE`

#### 性能优化
- 常用查询字段添加索引
- 大表（>10万条）考虑分区
- 定期清理软删除数据
- 使用连接池（配置最大连接数）

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

### 7. 文件存储模块
- 文件位置: `internal/service/storage.go`, `internal/api/upload.go`
- 功能: 文件上传、存储管理、支持多种存储方式
- 存储类型: local（本地）、oss（阿里云）、cos（腾讯云）
- 支持文件类型: images（图片）、audio（音频）、video（视频）、document（文档）

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
cd web-new && pnpm install

# 开发模式
cd web-new && pnpm dev

# 构建
cd web-new && pnpm build

# 预览
cd web-new && pnpm preview
```

## 数据初始化

### 1. 数据库迁移
创建/更新表结构、外键约束和索引：
```bash
cd backend && go run cmd/main.go -migrate
```

此命令会：
- 创建所有表（如果不存在）
- 更新表结构（添加新字段）
- 创建外键约束
- 创建性能索引
- **注意**: 不会删除已有数据

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

### 4. 种子数据导入（开发测试用）
导入示例数据用于开发测试：
```bash
# 导入所有种子数据
cd backend && go run cmd/seed/main.go -all

# 只导入特定类型数据
cd backend && go run cmd/seed/main.go -orgs     # 只导入组织
cd backend && go run cmd/seed/main.go -users    # 只导入用户
cd backend && go run cmd/seed/main.go -cases    # 只导入走失人员
cd backend && go run cmd/seed/main.go -dialects # 只导入方言
cd backend && go run cmd/seed/main.go -tasks    # 只导入任务

# 清空数据后重新导入（危险！会删除所有业务数据）
cd backend && go run cmd/seed/main.go -clean -all
```

种子数据包含：
- **组织**: 北京、上海、广东、深圳志愿者协会
- **用户**: 1个管理员、1个管理者、2个志愿者（带初始密码）
- **走失人员**: 2个示例案件
- **方言**: 北京话、上海话、粤语示例
- **任务**: 1个示例任务

### 完整初始化流程（新环境）
```bash
# 1. 确保数据库已创建
# 2. 执行迁移
cd backend && go run cmd/main.go -migrate

# 3. 初始化根组织
cd backend && go run cmd/main.go -init

# 4. 创建超级管理员
cd backend && go run cmd/initdata/main.go -exec -phone=13800138000 -password=admin123

# 5. 导入种子数据（可选，开发环境推荐）
cd backend && go run cmd/seed/main.go -all
```

### 数据库结构说明

#### 核心表
| 表名 | 说明 | 主要关联 |
|------|------|---------|
| ty_users | 用户表 | ty_organizations |
| ty_user_profiles | 用户扩展信息 | ty_users |
| ty_organizations | 组织架构 | ty_organizations(自关联) |
| ty_org_stats | 组织统计 | ty_organizations |

#### 业务表
| 表名 | 说明 | 主要关联 |
|------|------|---------|
| ty_missing_persons | 走失人员 | ty_users, ty_organizations |
| ty_missing_photos | 走失人员照片 | ty_missing_persons |
| ty_missing_person_tracks | 轨迹记录 | ty_missing_persons, ty_users |
| ty_dialects | 方言语音 | ty_users, ty_organizations |
| ty_dialect_comments | 方言评论 | ty_dialects, ty_users |
| ty_dialect_likes | 方言点赞 | ty_dialects, ty_users |
| ty_dialect_play_logs | 播放记录 | ty_dialects, ty_users |
| ty_tasks | 任务 | ty_missing_persons, ty_users, ty_organizations |
| ty_task_attachments | 任务附件 | ty_tasks |
| ty_task_logs | 任务日志 | ty_tasks, ty_users |
| ty_task_comments | 任务评论 | ty_tasks, ty_users |

#### 工作流表
| 表名 | 说明 | 主要关联 |
|------|------|---------|
| ty_workflows | 工作流定义 | ty_users |
| ty_workflow_steps | 工作流步骤 | ty_workflows |
| ty_workflow_instances | 工作流实例 | ty_workflows, ty_workflow_steps, ty_users |
| ty_workflow_histories | 工作流历史 | ty_workflow_instances, ty_workflow_steps, ty_users |

#### 系统表
| 表名 | 说明 | 主要关联 |
|------|------|---------|
| ty_tags | 标签 | - |
| ty_notifications | 通知 | ty_users |
| ty_operation_logs | 操作日志 | ty_users |
| ty_configs | 系统配置 | - |
| ty_dashboard_stats | 仪表盘统计 | - |

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

storage:
  type: local           # 存储类型: local/oss/cos
  local_path: ./uploads # 本地存储路径
  base_url: http://localhost:8080/uploads  # 文件访问基础URL
  max_file_size: 52428800  # 最大文件大小(50MB)
  allowed_types: "jpg,png,gif,mp4,mp3,wav"  # 允许的文件类型
  # OSS配置(使用阿里云时填写)
  oss_access_key_id: ""
  oss_access_key_secret: ""
  oss_endpoint: ""
  oss_bucket: ""
  # COS配置(使用腾讯云时填写)
  cos_secret_id: ""
  cos_secret_key: ""
  cos_bucket: ""
  cos_region: ""
```

### 前端配置 (.env)

```env
VITE_API_BASE_URL=/api/v1
```

## 注意事项

1. **Redis 可选**: 如果 Redis 未配置，系统会自动使用内存缓存
2. **数据库表前缀**: 所有表使用 `ty_` 前缀
3. **外键约束**: 已启用外键约束（通过GORM constraint标签自动创建）
4. **JWT 密钥**: 生产环境必须修改默认密钥
5. **微信小程序**: 需要配置正确的 appid 和密钥
6. **文件存储**: 
   - 本地存储需要确保 `./uploads` 目录存在且有写入权限
   - 生产环境建议使用 OSS 或 COS
   - 文件上传大小限制默认为 50MB
7. **数据迁移**:
   - 使用 `-migrate` 参数可安全地更新表结构（不会删除数据）
   - 外键约束会自动创建
   - 性能索引会自动创建

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

### 2026-03-02
- **Web 平台全面重构**:
  - 删除旧版 `web-admin`，使用新版 `web-new`
  - 全新简洁办公OA风格设计
  - 温馨橙色主题 (`#e67e22`)
  - 去除 Tailwind CSS，使用 Ant Design 5 + 内联样式
  - 优化侧边栏、顶部导航、工作台布局
  - 统一所有列表页面的表格样式
  - 优化移动端适配
  - 添加 Framer Motion 页面切换动画
  - 完善权限控制的前端展示

### 2024-03-02
- **完善权限控制系统**：
  - 后端RBAC权限中间件（RequireRole/RequireAdmin/RequireManager/RequireSuperAdmin）
  - API路由精细化权限控制
  - 前端usePermission Hooks和PermissionGuard组件
  - 侧边栏菜单根据角色动态显示
  - 新增操作日志页面（仅超级管理员可见）
  - 新增系统设置页面（仅管理员可见）
- **操作日志审计**：
  - 自动记录所有API请求
  - 支持按用户、模块、操作、时间筛选
  - 统计报表和可视化
  - 旧日志自动清理功能
- **数据迁移与初始化**：
  - 完善AutoMigrate，支持所有模型（28个表）
  - GORM自动外键约束（constraint标签）
  - 新增种子数据导入工具（cmd/seed）
  - 支持按类型导入（组织/用户/案件/方言/任务）
  - 支持清空后重新导入
  - 完善初始化文档
- 完善小程序端任务管理功能（列表、详情、领取、完成）
- 添加文件存储服务，支持本地/OSS/COS三种存储方式
- 添加文件上传API（单文件、批量上传、删除）
- 添加日志服务
- 扩展配置文件，添加SMS、Email、Map、Notification等配置
- 完善小程序端工作台页面
- **数据库优化**：
  - 为所有模型添加完整的外键约束（constraint）
  - 添加复合索引优化常用查询
  - 启用外键约束（DisableForeignKeyConstraintWhenMigrating: false）
  - 添加全文搜索索引支持
  - 优化连接池配置（支持最大生命周期）
- **小程序完善**：
  - 添加消息通知页面（notification/list）
  - 添加设置页面（settings/index）
  - 添加个人资料编辑页面（volunteer/edit-profile）
  - 优化个人资料页面UI
  - 添加公共组件（loading、empty）
  - 完善用户信息和统计展示

### 2024-02-27
- 实现完整的任务管理功能（分配、转派、完成、取消）
- 实现工作流管理功能（定义、步骤、审批）
- 重构数据初始化方式，支持SQL导入
- 完善前端任务管理和工作流页面

## 权限控制

### 后端权限控制

#### 角色层级
- `super_admin`: 超级管理员 - 拥有所有权限
- `admin`: 管理员 - 管理用户、组织、工作流定义
- `manager`: 管理者 - 分配任务、审批工作流
- `volunteer`: 志愿者 - 基本操作

#### 权限中间件
- `RequireRole(minRole)`: 需要指定角色及以上
- `RequireAdmin()`: 需要管理员权限
- `RequireManager()`: 需要管理者权限
- `RequireSuperAdmin()`: 需要超级管理员权限

#### API权限配置
| 路由 | 所需权限 |
|------|---------|
| /api/v1/users/* | Admin |
| /api/v1/organizations (写操作) | Admin |
| /api/v1/missing-persons (创建/更新) | Manager |
| /api/v1/tasks/* | 根据操作不同 |
| /api/v1/workflows (写操作) | Admin |
| /api/v1/operation-logs/* | SuperAdmin |

### 前端权限控制

#### React Hooks
- `usePermission()`: 获取当前用户权限信息
- `PermissionGuard`: 权限包装组件
- `AdminOnly`: 仅管理员可见
- `ManagerOnly`: 仅管理者可见
- `SuperAdminOnly`: 仅超级管理员可见

#### 路由权限
- `/users`: Admin+
- `/settings`: Admin+
- `/logs`: SuperAdmin
- 其他页面: 所有登录用户

#### 侧边栏菜单
根据用户角色动态显示菜单项。

### 操作日志审计

#### 功能
- 自动记录所有API请求
- 支持按用户、模块、操作、时间筛选
- 统计报表（仅超级管理员可见）
- 自动清理旧日志

#### 字段
- 用户ID、角色
- 模块、操作类型
- 请求方法、路径、参数
- 响应状态、耗时
- IP地址、User-Agent

## 待办事项

- [x] 完善小程序端功能
- [x] 添加文件存储服务
- [x] 完善权限控制
- [x] 添加操作日志审计
- [ ] 实现数据大屏页面
- [ ] 添加更多测试数据
- [ ] 实现消息推送服务
- [ ] 添加短信服务
