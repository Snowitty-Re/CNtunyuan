# 助力团圆系统 - 单元测试文档

## 测试概览

本项目采用 Go 标准测试框架，使用 testify 断言库和 SQLite 内存数据库进行单元测试。

## 测试结构

```
backend/
├── internal/
│   ├── domain/entity/              # 领域实体测试
│   │   ├── user_test.go
│   │   ├── organization_test.go
│   │   ├── missing_person_test.go
│   │   ├── task_test.go
│   │   └── dialect_test.go
│   │
│   ├── domain/service/             # 领域服务测试
│   │   └── auth_service_test.go
│   │
│   ├── application/service/        # 应用服务测试
│   │   ├── user_service_test.go
│   │   ├── organization_service_test.go
│   │   ├── missing_person_service_test.go
│   │   ├── task_service_test.go
│   │   ├── dialect_service_test.go
│   │   └── file_service_test.go
│   │
│   ├── interfaces/http/middleware/ # 中间件测试
│   │   └── middleware_test.go
│   │
│   └── testutil/                   # 测试工具包
│       ├── testutil.go
│       └── handler_testutil.go
```

## 运行测试

### 运行所有测试
```bash
go test ./... -count=1
```

### 运行特定包测试
```bash
# 领域实体测试
go test ./internal/domain/entity/... -v

# 领域服务测试
go test ./internal/domain/service/... -v

# 应用服务测试
go test ./internal/application/service/... -v

# 中间件测试
go test ./internal/interfaces/http/middleware/... -v
```

### 运行特定测试
```bash
go test -run TestNewUser ./internal/domain/entity/...
go test -run TestUserAppService ./internal/application/service/...
go test -run TestTaskAppService_Start ./internal/application/service/...
```

### 带覆盖率报告
```bash
# 查看覆盖率摘要
go test ./... -count=1 -cover

# 生成覆盖率报告
go test ./... -count=1 -coverprofile=coverage.out

# 查看 HTML 覆盖率报告
go tool cover -html=coverage.out
```

### 竞态检测
```bash
go test -race ./...
```

## 测试覆盖率

| 模块 | 包路径 | 测试文件数 |
|------|--------|-----------|
| 领域实体 | `internal/domain/entity` | 5 |
| 领域服务 | `internal/domain/service` | 1 |
| 应用服务 | `internal/application/service` | 6 |
| HTTP中间件 | `internal/interfaces/http/middleware` | 1 |

**总计**: 13 个测试文件，110+ 测试函数

## 测试说明

### 领域实体测试 (entity)
测试核心业务规则和验证逻辑：
- **User**: 创建、密码哈希、角色权限、验证、枚举状态（`IsValidUserStatus`）
- **Organization**: 创建、层级关系、状态管理
- **MissingPerson**: 创建、状态流转（`IsValidMissingStatus` / `IsValidUrgencyLevel`）、轨迹管理
- **Task**: 状态机、分配逻辑、进度跟踪
- **Dialect**: 创建、点赞统计、播放计数

### 领域服务测试 (service)
- **AuthService**: 登录/登出、Token 刷新、微信登录、手机号绑定、登录锁定

### 应用服务测试 (service)
- **UserAppService**: 用户 CRUD、状态/角色更新、密码修改（`ErrOldPasswordWrong` Sentinel）、权限检查、`GetStats` 按 reporter 范围统计
- **OrganizationAppService**: 组织 CRUD、树形结构、移动组织（自移动防护 `ErrOrganizationInvalidMove`）
- **MissingPersonAppService**: 走失人员 CRUD、状态枚举校验、轨迹管理
- **TaskAppService**: 任务 CRUD、分配、状态流转、**任务所有权执行**（Start/Complete 仅允许被分配者）
- **DialectAppService**: 方言 CRUD、审核、点赞、精选
- **FileAppService**: 文件上传（含安全检查）、下载、删除

### 中间件测试 (middleware)
- **AuthMiddleware**: JWT 认证、角色权限检查
- **CORSMiddleware**: 跨域请求处理
- **LoggerMiddleware**: 请求日志记录
- **RecoveryMiddleware**: Panic 恢复
- **RequestIDMiddleware**: 请求追踪 ID

## 测试工具 (testutil)

### 数据库工具
```go
// 创建测试数据库（SQLite 内存）
tdb := testutil.SetupTestDB(t)
defer tdb.Close()

// 创建测试记录
testutil.MustCreate(t, tdb.DB, entity)

// 更新测试记录（使用 GORM 直接操作）
require.NoError(t, tdb.DB.Save(entity).Error)

// 查找记录
testutil.MustFind(t, db, &result, "id = ?", id)
```

### HTTP 测试工具
```go
// 创建测试请求
w, req := testutil.CreateTestRequest(t, "POST", "/api/v1/users", body)

// 执行请求
router.ServeHTTP(w, req)

// 断言响应
assert.Equal(t, http.StatusOK, w.Code)
```

## 测试技术

### SQLite 内存数据库
- 使用 `github.com/glebarez/sqlite` 纯 Go 实现
- 无需 CGO，支持 Windows/Linux/macOS
- 每个测试使用独立内存数据库，完全隔离

### 表驱动测试
```go
func TestXXX(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"case1", "input1", "output1", false},
        {"case2", "input2", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 测试逻辑
        })
    }
}
```

### Mock 对象
文件服务测试使用 Mock 实现：
- `MockStorageService`: 模拟文件存储
- `MockVirusScanner`: 模拟病毒扫描
- `MockFileRepository`: 模拟文件仓储

## 重要测试行为说明

### 任务所有权检查优先于状态检查

`task_service.go` 中 `Start()` 和 `Complete()` 会先检查任务是否分配给当前用户，再检查状态是否合法。因此编写无效状态测试时，**必须先将任务分配给测试用户**，否则会得到 `ErrTaskNotAssignedToUser` 而非预期的状态错误：

```go
// 正确：测试前先设置 AssigneeID
task := &entity.Task{
    ...,
    AssigneeID: &assignee.ID,  // 必须设置，否则 Start() 先返回 ErrTaskNotAssignedToUser
}
testutil.MustCreate(t, tdb.DB, task)
err := service.Start(ctx, task.ID, assignee.ID)
assert.Contains(t, err.Error(), "当前状态不能开始任务")
```

### Sentinel 错误与 HTTP 状态码映射

| Sentinel 错误 | HTTP 状态码 |
|--------------|-------------|
| `ErrUserNotFound` | 404 |
| `ErrTaskNotFound` | 404 |
| `ErrFileNotFound` | 404 |
| `ErrOrganizationNotFound` | 404 |
| `ErrTaskNotAssignedToUser` | 403 |
| `ErrCannotModify` | 403 |
| `ErrPhoneExists` / `ErrEmailExists` | 409 |
| `ErrFileTooLarge` | 400 |
| `ErrOldPasswordWrong` | 400 |
| `ErrOrganizationInvalidMove` | 400 |
| `ErrAlreadyLiked` / `ErrNotLiked` | 400 |

## 持续集成

建议在 CI/CD 中运行：
```bash
# 运行测试（带竞态检测）
go test -race ./... -count=1

# 生成覆盖率报告
go test ./... -count=1 -coverprofile=coverage.out

# 构建验证
go build ./...
```

## 添加新测试

### 1. 创建测试文件
在对应包目录下创建 `*_test.go` 文件：
```go
package service

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestNewFeature(t *testing.T) {
    service, tdb := setupXxxTest(t)
    defer tdb.Close()
    // 测试代码
}
```

### 2. 使用测试工具
```go
func TestWithDB(t *testing.T) {
    tdb := testutil.SetupTestDB(t)
    defer tdb.Close()

    // 创建测试数据
    user := &entity.User{...}
    testutil.MustCreate(t, tdb.DB, user)

    // 更新数据（用 GORM 直接操作）
    user.Role = entity.RoleAdmin
    require.NoError(t, tdb.DB.Save(user).Error)

    // 执行测试
}
```

### 3. 运行测试
```bash
go test -v -run TestNewFeature ./path/to/package/...
```

## 测试规范

1. **测试命名**: `TestXxx` 或 `TestXxx_Yyy`（子测试）
2. **测试隔离**: 每个测试独立，不依赖其他测试
3. **清理资源**: 使用 `defer tdb.Close()` 清理数据库
4. **唯一数据**: 使用 UUID 或时间戳确保测试数据唯一
5. **错误检查**: 同时测试成功和失败场景
6. **Sentinel 错误**: 使用 `assert.Equal(t, service.ErrXxx, err)` 而非 `assert.Contains(t, err.Error(), "...")`（字符串比较脆弱）
