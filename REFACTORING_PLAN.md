# 团圆寻亲系统 - 生产级重构计划

> 版本: 1.0.0  
> 日期: 2026-03-07  
> 目标: 将现有系统升级为生产级企业应用，完善 OA 工作流特性

---

## 一、现状分析

### 1.1 现有架构优势 ✅

| 方面 | 现状 | 评价 |
|------|------|------|
| 架构模式 | Clean Architecture | ✅ 分层清晰，依赖向内 |
| 技术栈 | Go + Gin + GORM + PostgreSQL | ✅ 现代化技术栈 |
| 认证授权 | JWT + RBAC 基础 | ✅ 基础权限控制 |
| 中间件 | 限流、日志、恢复、安全头 | ✅ 生产基础 |
| 数据库 | 支持 PostgreSQL/MySQL | ✅ 双数据库支持 |
| 实体设计 | 领域实体完整 | ✅ DDD 实践良好 |

### 1.2 缺失的生产级特性 ⚠️

| 优先级 | 特性 | 影响 |
|--------|------|------|
| P0 | 工作流引擎 | OA 核心功能缺失 |
| P0 | 数据权限控制 | 多组织数据隔离不完整 |
| P0 | 审计日志 | 合规性要求 |
| P1 | 通知系统 | 实时消息推送 |
| P1 | 高级 RBAC | 细粒度权限控制 |
| P1 | API 文档 | Swagger/OpenAPI |
| P1 | 测试覆盖 | 单元/集成/接口测试 |
| P2 | 缓存策略 | Redis 缓存优化 |
| P2 | 多租户 | SaaS 化准备 |
| P2 | 监控告警 | Prometheus/Grafana |

---

## 二、重构目标

### 2.1 核心目标

```
┌─────────────────────────────────────────────────────────────┐
│                      生产级应用目标                          │
├─────────────────────────────────────────────────────────────┤
│  1. 完善 OA 工作流引擎  - 支持可视化流程设计                 │
│  2. 企业级权限系统      - RBAC + ABAC + 数据权限             │
│  3. 全方位审计追踪      - 操作日志、数据变更历史             │
│  4. 实时通知体系        - WebSocket + 推送 + 站内信         │
│  5. 完整测试覆盖        - 单元测试 80%+                      │
│  6. DevOps 就绪        - Docker + K8s + CI/CD               │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 阶段规划

```
Phase 1 (4周): 基础设施强化
    ├── 统一错误处理
    ├── 审计日志系统
    ├── 数据权限框架
    └── 缓存抽象层

Phase 2 (6周): OA 工作流引擎
    ├── 流程定义模型
    ├── 状态机引擎
    ├── 审批节点实现
    └── 任务委托/催办

Phase 3 (4周): 权限系统升级
    ├── 角色权限矩阵
    ├── 数据权限规则
    ├── 字段级权限
    └── 权限缓存优化

Phase 4 (3周): 通知系统
    ├── WebSocket 服务
    ├── 消息模板引擎
    ├── 多渠道推送
    └── 消息中心

Phase 5 (3周): 测试与文档
    ├── 单元测试覆盖
    ├── 集成测试
    ├── API 文档
    └── 部署文档
```

---

## 三、详细重构方案

### Phase 1: 基础设施强化

#### 3.1.1 统一错误处理增强

**目标**: 建立完整的错误码体系和错误处理机制

```go
// 扩展现有错误码
const (
    // 工作流错误 (7000-7099)
    CodeWorkflowNotFound     ErrorCode = 7000
    CodeWorkflowInvalidState ErrorCode = 7001
    CodeWorkflowTransition   ErrorCode = 7002
    CodeWorkflowApproval     ErrorCode = 7003
    
    // 权限错误 (8000-8099)
    CodePermissionDenied     ErrorCode = 8000
    CodeDataPermissionDenied ErrorCode = 8001
    CodeFieldPermissionDenied ErrorCode = 8002
)

// 错误包装器
func WrapWithContext(err error, code ErrorCode, ctx map[string]interface{}) error
```

**实现清单**:
- [ ] 扩展错误码体系
- [ ] 添加错误上下文信息
- [ ] 错误日志自动记录
- [ ] 错误通知机制

#### 3.1.2 审计日志系统

**目标**: 记录所有关键操作，支持合规审计

```
表结构: ty_audit_logs
├── id: UUID PK
├── user_id: UUID FK
├── org_id: UUID FK
├── action: VARCHAR(50)      # CREATE/UPDATE/DELETE/LOGIN
├── resource: VARCHAR(50)    # user/task/missing_person
├── resource_id: UUID
├── old_values: JSONB
├── new_values: JSONB
├── ip_address: VARCHAR(50)
├── user_agent: VARCHAR(255)
├── trace_id: VARCHAR(50)
├── created_at: TIMESTAMP
└── INDEX: user_id, org_id, resource, created_at
```

**实现清单**:
- [ ] 创建审计日志表
- [ ] 开发 AuditLogService
- [ ] 实现自动审计切面
- [ ] 审计日志查询 API
- [ ] 敏感数据脱敏

**代码位置**: 
- `backend/internal/domain/entity/audit_log.go`
- `backend/internal/application/service/audit_service.go`
- `backend/internal/infrastructure/audit/audit_aspect.go`

#### 3.1.3 数据权限框架

**目标**: 实现多组织环境下的数据隔离和共享

```go
// 数据权限上下文
type DataPermissionContext struct {
    UserID       string
    OrgID        string
    Role         entity.Role
    OrgIDs       []string        // 可访问的组织列表
    IsSuperAdmin bool
}

// 数据权限策略接口
type DataPermissionStrategy interface {
    GetOrgFilter(ctx *DataPermissionContext) *gorm.DB
    CanAccess(ctx *DataPermissionContext, resourceOrgID string) bool
    GetAccessibleOrgs(ctx *DataPermissionContext) []string
}
```

**权限规则**:
| 角色 | 数据可见范围 |
|------|-------------|
| SuperAdmin | 全部 |
| Admin | 本组织及子组织 |
| Manager | 本组织及子组织 |
| Volunteer | 仅本组织 |

**实现清单**:
- [ ] DataPermissionContext 中间件
- [ ] 组织树查询优化
- [ ] 自动数据过滤切面
- [ ] 跨组织数据共享白名单

#### 3.1.4 缓存抽象层

**目标**: 统一缓存接口，支持多级缓存

```go
// 缓存管理器接口
type CacheManager interface {
    Get(ctx context.Context, key string, dest interface{}) error
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    DeleteByPattern(ctx context.Context, pattern string) error
    Exists(ctx context.Context, key string) bool
}

// 多级缓存实现
// Local (BigCache) -> Redis -> Database
```

**实现清单**:
- [ ] CacheManager 接口定义
- [ ] Redis 实现
- [ ] 本地缓存实现
- [ ] 缓存穿透/击穿/雪崩防护

---

### Phase 2: OA 工作流引擎

#### 3.2.1 工作流核心模型

**流程定义表**:
```
ty_workflow_definitions
├── id: UUID PK
├── name: VARCHAR(100)           # 流程名称
├── key: VARCHAR(50) UNIQUE     # 流程标识
├── version: INT                 # 版本号
├── description: TEXT
├── category: VARCHAR(50)        # 分类
├── status: VARCHAR(20)         # draft/active/archived
├── start_node_id: UUID         # 起始节点
├── org_id: UUID                 # 所属组织
├── config: JSONB               # 扩展配置
├── created_by: UUID
└── created_at/updated_at
```

**节点定义表**:
```
ty_workflow_nodes
├── id: UUID PK
├── workflow_id: UUID FK
├── name: VARCHAR(100)
├── type: VARCHAR(20)           # start/approval/task/branch/end
├── config: JSONB              # 节点配置
├── assignee_type: VARCHAR(20) # user/role/dept/expression
├── assignees: JSONB           # 处理人配置
├── position: JSONB            # 画布位置
└── order_index: INT
```

**流程实例表**:
```
ty_workflow_instances
├── id: UUID PK
├── definition_id: UUID FK
├── business_key: VARCHAR(100)  # 业务标识
├── business_id: UUID          # 业务数据ID
├── title: VARCHAR(200)
├── status: VARCHAR(20)        # running/completed/cancelled
├── current_node_id: UUID
├── started_by: UUID
├── started_at: TIMESTAMP
├── completed_at: TIMESTAMP
├── variables: JSONB          # 流程变量
└── result: VARCHAR(20)        # approve/reject
```

**任务表** (增强现有 ty_tasks):
```
# 扩展现有任务表，添加工作流相关字段
ALTER TABLE ty_tasks ADD COLUMN
├── workflow_instance_id UUID
├── workflow_node_id UUID
├── task_type VARCHAR(20)     # normal/workflow
├── due_time TIMESTAMP        # 截止时间
├── reminded_at TIMESTAMP     # 提醒时间
└── delegate_from UUID        # 委托来源
```

#### 3.2.2 工作流状态机

```go
// 状态机定义
type WorkflowState string

const (
    StateDraft      WorkflowState = "draft"
    StatePending    WorkflowState = "pending"      // 待提交
    StateProcessing WorkflowState = "processing"   // 审批中
    StateApproved   WorkflowState = "approved"     // 已通过
    StateRejected   WorkflowState = "rejected"     // 已拒绝
    StateCancelled  WorkflowState = "cancelled"    // 已取消
    StateReturned   WorkflowState = "returned"     // 已退回
)

// 状态转换规则
type StateTransition struct {
    From      WorkflowState
    To        WorkflowState
    Event     string           // submit/approve/reject/return/cancel
    Guards    []TransitionGuard // 转换条件
    Actions   []TransitionAction // 执行动作
}
```

**状态转换图**:
```
                    ┌──────────┐
                    │  DRAFT   │
                    └────┬─────┘
                         │ submit
                    ┌────▼─────┐
        ┌───────────┤ PENDING  ├───────────┐
        │           └────┬─────┘           │
   cancel│                │ approve/reject  │return
        ┌▼────────┐  ┌───▼────┐         ┌──▼───┐
        │CANCELLED│  │PROCESS-│         │RETURN│
        └─────────┘  │  ING   │         │ -ED  │
                     └───┬────┘         └──┬───┘
                    approve│               │resubmit
                   ┌───────▼──────┐   ┌────▼───┐
                   │   APPROVED   │   │PENDING │
                   └──────────────┘   └────────┘
                   
                   ┌──────────────┐
                   │   REJECTED   │
                   └──────────────┘
```

#### 3.2.3 审批节点实现

```go
// 审批节点接口
type ApprovalNode interface {
    Execute(ctx context.Context, inst *WorkflowInstance) error
    GetAssignees(ctx context.Context, inst *WorkflowInstance) ([]string, error)
    Validate(ctx context.Context, action ApprovalAction) error
}

// 审批动作
type ApprovalAction struct {
    Action   string   // approve/reject/transfer/delegate
    Comment  string
    NextNode string   // 跳转节点（用于退回）
    Delegates []string // 转办人
}

// 节点类型实现
// 1. 单人审批
// 2. 会签（全部通过）
// 3. 或签（一人通过）
// 4. 条件分支
// 5. 并行审批
```

**审批配置示例**:
```json
{
  "node_type": "approval",
  "assignee_config": {
    "type": "role",
    "value": ["manager", "admin"],
    "strategy": "any"  // any/all
  },
  "approval_config": {
    "mode": "sequential",  // sequential/parallel
    "required": 1,
    "allow_transfer": true,
    "allow_delegate": true,
    "auto_pass": false
  },
  "timeout_config": {
    "duration": 24,
    "unit": "hour",
    "action": "remind",
    "escalation": "admin"
  },
  "conditions": [
    {
      "field": "amount",
      "operator": ">",
      "value": 10000,
      "next_node": "high_level_approval"
    }
  ]
}
```

#### 3.2.4 任务委托与催办

```go
// 委托记录
type TaskDelegation struct {
    ID          string
    TaskID      string
    FromUserID  string
    ToUserID    string
    StartTime   time.Time
    EndTime     *time.Time
    Reason      string
    Status      string  // active/cancelled/expired
    CreatedBy   string
}

// 催办记录
type TaskReminder struct {
    ID          string
    TaskID      string
    ReminderType string // system/manual
    RemindCount int
    LastRemindAt *time.Time
    NextRemindAt *time.Time
    Status      string
}

// 催办服务
type ReminderService interface {
    // 创建自动催办计划
    ScheduleReminder(ctx context.Context, taskID string, config ReminderConfig) error
    // 手动催办
    SendReminder(ctx context.Context, taskID string, userID string) error
    // 批量催办
    BatchRemind(ctx context.Context, taskIDs []string) error
}
```

#### 3.2.5 工作流服务接口

```go
// 工作流引擎服务
type WorkflowEngine interface {
    // 流程定义管理
    CreateDefinition(ctx context.Context, def *WorkflowDefinition) error
    PublishDefinition(ctx context.Context, id string) error
    GetDefinition(ctx context.Context, key string, version int) (*WorkflowDefinition, error)
    
    // 流程实例管理
    StartInstance(ctx context.Context, req StartInstanceRequest) (*WorkflowInstance, error)
    CancelInstance(ctx context.Context, id string, reason string) error
    GetInstance(ctx context.Context, id string) (*WorkflowInstance, error)
    
    // 任务处理
    CompleteTask(ctx context.Context, taskID string, action ApprovalAction) error
    TransferTask(ctx context.Context, taskID string, toUserID string) error
    DelegateTask(ctx context.Context, taskID string, toUserID string) error
    ReturnTask(ctx context.Context, taskID string, toNodeID string) error
    
    // 查询
    GetTodoTasks(ctx context.Context, userID string, page PageParam) ([]Task, int64, error)
    GetDoneTasks(ctx context.Context, userID string, page PageParam) ([]Task, int64, error)
    GetMySubmitted(ctx context.Context, userID string, page PageParam) ([]WorkflowInstance, int64, error)
}
```

---

### Phase 3: 权限系统升级

#### 3.3.1 RBAC 权限矩阵

```
权限模型: 用户-角色-权限-资源

ty_roles (扩展)
├── id: UUID PK
├── name: VARCHAR(50)
├── code: VARCHAR(50) UNIQUE
├── org_id: UUID
├── parent_id: UUID          # 角色继承
├── permissions: JSONB       # 权限列表
├── data_scope: VARCHAR(20)  # all/dept/custom/self
├── custom_scope: JSONB      # 自定义范围
└── is_system: BOOLEAN       # 系统角色

ty_permissions (优化)
├── id: UUID PK
├── name: VARCHAR(100)
├── code: VARCHAR(100) UNIQUE
├── resource: VARCHAR(50)    # 资源标识
├── action: VARCHAR(50)      # 操作类型
├── description: VARCHAR(255)
└── category: VARCHAR(50)    # 分类
```

**权限码规范**:
```
格式: {resource}:{action}

user:create    - 创建用户
user:read      - 查看用户
user:update    - 更新用户
user:delete    - 删除用户
user:export    - 导出用户

task:create    - 创建任务
task:assign    - 分配任务
task:complete  - 完成任务
```

#### 3.3.2 数据权限规则

```go
// 数据范围类型
type DataScope string

const (
    DataScopeAll       DataScope = "all"       // 全部数据
    DataScopeDept      DataScope = "dept"      // 本部门及子部门
    DataScopeDeptOnly  DataScope = "dept_only" // 仅本部门
    DataScopeSelf      DataScope = "self"      // 仅本人
    DataScopeCustom    DataScope = "custom"    // 自定义
)

// 数据权限过滤器
type DataPermissionFilter struct {
    Scope     DataScope
    OrgIDs    []string  // 组织白名单
    UserIDs   []string  // 用户白名单
    Exclude   []string  // 排除列表
}

// 动态权限规则
type DataPermissionRule struct {
    Resource    string           // 资源类型
    Field       string           // 字段名
    Operator    string           // eq/ne/in/like
    ValueSource string           // const/context/expression
    Value       interface{}
}
```

**实现方式**:
```go
// 在 Repository 层自动注入
func (r *Repository) WithDataPermission(ctx context.Context, db *gorm.DB) *gorm.DB {
    dpCtx := GetDataPermissionContext(ctx)
    
    if dpCtx.IsSuperAdmin {
        return db
    }
    
    switch dpCtx.DataScope {
    case DataScopeAll:
        return db
    case DataScopeDept:
        return db.Where("org_id IN ?", dpCtx.AccessibleOrgIDs)
    case DataScopeSelf:
        return db.Where("created_by = ?", dpCtx.UserID)
    default:
        return db.Where("1=0") // 无权限
    }
}
```

#### 3.3.3 字段级权限

```go
// 字段权限配置
type FieldPermission struct {
    Resource   string
    Field      string
    Role       string
    Permission string  // read/write/none
}

// 响应过滤器
func FilterResponseFields(ctx context.Context, data interface{}, resource string) interface{}

// 示例: 普通志愿者不能查看用户手机号
{
    "resource": "user",
    "field": "phone",
    "role": "volunteer",
    "permission": "none"
}
```

---

### Phase 4: 通知系统

#### 3.4.1 WebSocket 实时服务

```go
// WebSocket 管理器
type WebSocketManager struct {
    clients    map[string]*Client      // user_id -> Client
    broadcast  chan Message
    register   chan *Client
    unregister chan *Client
}

// 消息结构
type Message struct {
    ID          string
    Type        string    // notification/message/system
    Title       string
    Content     string
    Data        map[string]interface{}
    ToUserID    string
    FromUserID  string
    CreatedAt   time.Time
}

// WebSocket 服务启动
func (m *WebSocketManager) Start()
func (m *WebSocketManager) SendToUser(userID string, msg Message) error
func (m *WebSocketManager) SendToUsers(userIDs []string, msg Message) error
func (m *WebSocketManager) Broadcast(msg Message) error
```

#### 3.4.2 消息模板引擎

```
ty_message_templates
├── id: UUID PK
├── code: VARCHAR(50) UNIQUE
├── name: VARCHAR(100)
├── channel: VARCHAR(20)  # sms/email/push/websocket
├── subject: VARCHAR(200)
├── content: TEXT         # 模板内容
├── variables: JSONB      # 变量定义
└── status: VARCHAR(20)
```

**模板示例**:
```
模板: task_assigned
渠道: websocket
标题: 新任务分配
内容: 您有一个新的{{.TaskType}}任务「{{.TaskTitle}}」需要处理，截止日期：{{.Deadline}}

模板: workflow_approved
渠道: push
标题: 审批通过
内容: 您的「{{.WorkflowName}}」申请已通过{{.ApprovedBy}}的审批
```

#### 3.4.3 多渠道推送

```go
// 推送服务接口
type NotificationService interface {
    Send(ctx context.Context, notification Notification) error
    SendBatch(ctx context.Context, notifications []Notification) error
    GetUserNotifications(ctx context.Context, userID string, page PageParam) ([]Notification, error)
    MarkAsRead(ctx context.Context, notificationID string) error
    MarkAllAsRead(ctx context.Context, userID string) error
    GetUnreadCount(ctx context.Context, userID string) (int, error)
}

// 渠道实现
// 1. WebSocket (站内实时)
// 2. 个推/极光 (App Push)
// 3. 阿里云短信
// 4. 邮件
```

---

### Phase 5: 测试与文档

#### 3.5.1 测试策略

```
测试金字塔:
                    ┌─────────┐
                    │   E2E   │  10%  (Playwright/Cypress)
                   ├───────────┤
                   │ Integration│ 30%  (API 测试)
                  ├─────────────┤
                  │    Unit     │ 60%  (Go test)
                 └───────────────┘
```

**单元测试规范**:
```go
// 测试文件命名: xxx_test.go
// 测试函数命名: Test{FunctionName}_{Scenario}

func TestTaskService_Create_Success(t *testing.T)
func TestTaskService_Create_InvalidTitle(t *testing.T)
func TestTaskService_Create_PermissionDenied(t *testing.T)
func TestTaskService_Assign_AlreadyAssigned(t *testing.T)
```

**Mock 规范**:
```go
// 使用 mockery 生成 Repository Mock
//go:generate mockery --name=TaskRepository --dir=. --output=./mocks

// 使用 testify mock
type MockTaskRepository struct {
    mock.Mock
}
```

#### 3.5.2 API 文档 (Swagger)

```go
// 使用 swaggo 注解
// @Summary      创建任务
// @Description  创建一个新的寻人任务
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        request  body      dto.CreateTaskRequest  true  "任务信息"
// @Success      201      {object}  dto.TaskResponse
// @Failure      400      {object}  response.ErrorResponse
// @Failure      401      {object}  response.ErrorResponse
// @Failure      403      {object}  response.ErrorResponse
// @Router       /api/v1/tasks [post]
func (h *TaskHandler) Create(c *gin.Context) { ... }
```

#### 3.5.3 部署文档

```
docs/
├── deployment/
│   ├── docker.md           # Docker 部署
│   ├── kubernetes.md       # K8s 部署
│   ├── database.md         # 数据库配置
│   └── monitoring.md       # 监控配置
├── development/
│   ├── setup.md            # 开发环境
│   ├── architecture.md     # 架构说明
│   └── workflow.md         # 开发流程
└── api/
    └── swagger.yaml        # API 文档
```

---

## 四、技术实现要点

### 4.1 目录结构调整

```
backend/
├── cmd/
│   ├── app/                    # 主应用
│   ├── worker/                 # 后台任务队列
│   └── migrate/                # 数据库迁移工具
├── internal/
│   ├── domain/
│   │   ├── entity/
│   │   │   ├── task.go
│   │   │   ├── workflow.go     # NEW
│   │   │   ├── audit_log.go    # NEW
│   │   │   └── ...
│   │   ├── repository/
│   │   ├── service/
│   │   └── valueobject/
│   ├── application/
│   │   ├── dto/
│   │   ├── service/
│   │   ├── event/              # NEW: 领域事件
│   │   └── mapper/
│   ├── infrastructure/
│   │   ├── database/
│   │   ├── cache/
│   │   ├── messaging/          # NEW: 消息队列
│   │   ├── websocket/          # NEW: WebSocket
│   │   ├── audit/              # NEW: 审计
│   │   ├── permission/         # NEW: 权限
│   │   └── workflow/           # NEW: 工作流引擎
│   ├── interfaces/
│   │   └── http/
│   │       ├── handler/
│   │       ├── middleware/
│   │       │   ├── auth.go
│   │       │   ├── audit.go    # NEW: 审计中间件
│   │       │   ├── permission.go # NEW: 权限中间件
│   │       │   └── data_scope.go # NEW: 数据范围
│   │       └── router/
│   └── config/
├── pkg/
│   ├── errors/
│   ├── logger/
│   ├── validator/
│   ├── security/               # NEW: 安全工具
│   └── utils/
└── migrations/
    ├── 001_init.sql
    ├── 002_workflow.sql        # NEW
    ├── 003_audit_log.sql       # NEW
    └── 004_permissions.sql     # NEW
```

### 4.2 关键依赖

```go
// go.mod 新增依赖
require (
    // 工作流引擎
    github.com/looplab/fsm v1.0.2           # 状态机
    
    // 消息队列
    github.com/hibiken/asynq v0.24.0       # Redis 任务队列
    
    // WebSocket
    github.com/gorilla/websocket v1.5.1    # WebSocket
    
    // 缓存
    github.com/allegro/bigcache/v3 v3.1.0  # 本地缓存
    
    // 测试
    github.com/stretchr/testify v1.9.0     # 测试框架
    github.com/vektra/mockery/v2 v2.42.0   # Mock 生成
    
    // API 文档
    github.com/swaggo/gin-swagger v1.6.0   # Swagger
    github.com/swaggo/swag v1.16.3
)
```

### 4.3 配置扩展

```yaml
# config.yaml 扩展
workflow:
  enabled: true
  async: true
  default_timeout: 24h
  reminder_intervals: [30m, 1h, 4h]

notification:
  websocket_enabled: true
  push_enabled: true
  sms_enabled: true
  channels:
    - type: websocket
      priority: 1
    - type: push
      priority: 2
    - type: sms
      priority: 3

audit:
  enabled: true
  log_level: info
  retention_days: 365
  sensitive_fields: [password, id_card, phone]

permission:
  cache_ttl: 10m
  data_scope_enabled: true
```

---

## 五、风险与应对

| 风险 | 可能性 | 影响 | 应对措施 |
|------|--------|------|----------|
| 工作流复杂性超预期 | 中 | 高 | 分阶段实现，先简化版 |
| 数据迁移问题 | 低 | 高 | 完整备份，回滚方案 |
| 性能下降 | 中 | 中 | 压测，缓存优化 |
| 开发周期延期 | 高 | 中 | 敏捷迭代，优先核心功能 |

---

## 六、成功指标

| 指标 | 当前 | 目标 |
|------|------|------|
| 单元测试覆盖率 | <10% | >80% |
| API 响应时间 P99 | - | <200ms |
| 并发用户数 | - | >1000 |
| 系统可用性 | - | 99.9% |
| 代码质量评级 | - | A |

---

## 七、附录

### 7.1 参考资源

- [Temporal - 工作流引擎参考](https://temporal.io/)
- [Casbin - 权限引擎](https://casbin.org/)
- [Go Clean Architecture](https://github.com/bxcodec/go-clean-arch)

### 7.2 相关文档

- [后端 README](backend/README.md)
- [架构文档](backend/ARCHITECTURE.md)
- [开发规范](AGENTS.md)

---

**计划制定者**: AI Assistant  
**审核状态**: 待审核  
**最后更新**: 2026-03-07
