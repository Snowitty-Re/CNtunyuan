# 团圆寻亲系统 - 单元测试文档

## 测试概览

本项目采用 Go 标准测试框架，使用 testify 断言库和 SQLite 内存数据库进行单元测试。

## 测试结构

```
backend/
├── internal/
│   ├── domain/entity/              # 领域实体测试 (69.0% 覆盖率)
│   │   ├── user_test.go            # 用户实体测试
│   │   ├── organization_test.go    # 组织实体测试
│   │   ├── missing_person_test.go  # 走失人员实体测试
│   │   ├── task_test.go            # 任务实体测试
│   │   └── dialect_test.go         # 方言实体测试
│   │
│   ├── domain/service/             # 领域服务测试 (74.6% 覆盖率)
│   │   └── auth_service_test.go    # 认证服务测试
│   │
│   ├── application/service/        # 应用服务测试 (64.3% 覆盖率)
│   │   ├── user_service_test.go           # 用户服务测试
│   │   ├── organization_service_test.go   # 组织服务测试
│   │   ├── missing_person_service_test.go # 走失人员服务测试
│   │   ├── task_service_test.go           # 任务服务测试
│   │   ├── dialect_service_test.go        # 方言服务测试
│   │   └── file_service_test.go           # 文件服务测试
│   │
│   ├── interfaces/http/middleware/ # 中间件测试 (64.3% 覆盖率)
│   │   └── middleware_test.go      # HTTP 中间件测试
│   │
│   └── testutil/                   # 测试工具包
│       ├── testutil.go             # 数据库和辅助函数
│       └── handler_testutil.go     # HTTP 测试工具
```

## 运行测试

### 运行所有测试
```bash
cd backend
go test ./internal/... -count=1
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
```

### 带覆盖率报告
```bash
# 查看覆盖率摘要
go test ./internal/... -count=1 -cover

# 生成覆盖率报告
go test ./internal/... -count=1 -coverprofile=coverage.out

# 查看 HTML 覆盖率报告
go tool cover -html=coverage.out
```

### 详细输出
```bash
go test ./internal/... -v
```

## 测试覆盖率

| 模块 | 包路径 | 覆盖率 | 测试文件数 |
|------|--------|--------|-----------|
| 领域实体 | `internal/domain/entity` | 69.0% | 5 |
| 领域服务 | `internal/domain/service` | 74.6% | 1 |
| 应用服务 | `internal/application/service` | 64.3% | 6 |
| HTTP中间件 | `internal/interfaces/http/middleware` | 64.3% | 1 |

**总计**: 14 个测试文件，110+ 测试函数，300+ 测试用例

## 测试说明

### 领域实体测试 (entity)
测试核心业务规则和验证逻辑：
- **User**: 创建、密码哈希、角色权限、验证
- **Organization**: 创建、层级关系、状态管理
- **MissingPerson**: 创建、状态流转、轨迹管理
- **Task**: 状态机、分配逻辑、进度跟踪
- **Dialect**: 创建、点赞统计、播放计数

### 领域服务测试 (service)
- **AuthService**: 登录/登出、Token 刷新、微信登录、手机号绑定

### 应用服务测试 (service)
测试业务逻辑和用例：
- **UserAppService**: 用户 CRUD、状态/角色更新、密码修改、权限检查
- **OrganizationAppService**: 组织 CRUD、树形结构、移动组织
- **MissingPersonAppService**: 走失人员 CRUD、状态更新、轨迹管理
- **TaskAppService**: 任务 CRUD、分配、状态流转（草稿→待分配→已分配→进行中→已完成）
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
// 创建测试数据库
tdb := testutil.SetupTestDB(t)
defer tdb.Close()

// 创建测试记录
testutil.MustCreate(t, db, entity)

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

## 持续集成

建议在 CI/CD 中运行：
```bash
# 运行测试
go test ./internal/... -count=1

# 运行测试（带竞态检测）
go test ./internal/... -race -count=1

# 生成覆盖率报告
go test ./internal/... -count=1 -coverprofile=coverage.out
```

## 添加新测试

### 1. 创建测试文件
在对应包目录下创建 `*_test.go` 文件：
```go
package service

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestNewFeature(t *testing.T) {
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
    
    // 执行测试
    // ...
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
