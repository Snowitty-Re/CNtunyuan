# 团圆寻亲系统后端 - 问题修复记录

## 修复的问题列表

### 1. 数据库迁移脚本缺失审计日志表 ✅
**问题**：新增审计日志功能后，数据库迁移脚本中没有创建 `ty_audit_logs` 表的语句。

**修复**：
- PostgreSQL: `migrations/postgres/01_schema.sql` - 添加了审计日志表定义和索引
- MySQL: `migrations/mysql/01_schema.sql` - 添加了审计日志表定义和索引

### 2. AutoMigrate 缺少审计日志实体 ✅
**问题**：`database.AutoMigrate()` 函数没有包含 `entity.AuditLog`。

**修复**：`internal/infrastructure/database/db.go`
```go
err := db.AutoMigrate(
    &entity.User{},
    &entity.Organization{},
    &entity.MissingPerson{},
    &entity.Dialect{},
    &entity.Task{},
    &entity.File{},
    &entity.AuditLog{},  // 添加此行
)
```

### 3. 数据库表检查缺少审计日志表 ✅
**问题**：`check-db` 命令检查表结构时未包含审计日志表。

**修复**：`cmd/app/main.go`
```go
tables := []string{
    // ...
    "ty_files",
    "ty_audit_logs",  // 添加此行
}
```

### 4. Seed 清理数据缺少审计日志表 ✅
**问题**：`seed -clean` 命令清理数据时未清理审计日志表。

**修复**：`cmd/seed/main.go`
```go
tables := []string{
    // ...
    "ty_files",
    "ty_audit_logs",  // 添加此行
}
```

### 5. 审计中间件使用 nil Context ✅
**问题**：`saveAuditLog` 函数传入 `nil` 作为 context。

**修复**：`internal/interfaces/http/middleware/audit.go`
```go
func (m *AuditMiddleware) saveAuditLog(log *entity.AuditLog) {
    if m.auditRepo == nil {
        return
    }
    ctx := context.Background()  // 创建背景 context
    if err := m.auditRepo.Create(ctx, log); err != nil {
        logger.Error("Failed to save audit log", logger.Err(err))
    }
}
```

### 6. 文件上传使用错误的文件句柄 ✅
**问题**：`UploadFile` 在安全检查后重新打开文件，但上传时仍使用旧的已关闭的文件句柄。

**修复**：`internal/application/service/file_service.go`
```go
// 上传文件到存储（使用重新打开的文件句柄）
uploadedFile, err := s.storageService.Upload(ctx, uploadFile, ...)  // 使用 uploadFile 而非 file
```

### 7. FindOverdueTasks 缺少 deleted_at 检查 ✅
**问题**：查询逾期任务时未检查 `deleted_at IS NULL`，可能包含已删除的任务。

**修复**：`internal/infrastructure/repository/task_repository.go`
```go
err := r.db.WithContext(ctx).
    Where("deadline < ? AND status NOT IN (?, ?, ?) AND deleted_at IS NULL", ...)  // 添加 AND deleted_at IS NULL
    Find(&tasks).Error
```

### 8. 审计中间件缺少 context 导入 ✅
**问题**：修复 nil context 后未导入 `context` 包。

**修复**：`internal/interfaces/http/middleware/audit.go`
```go
import (
    "bytes"
    "context"  // 添加此行
    // ...
)
```

## 验证方法

### 编译检查
```bash
cd backend
go build ./...
```

### 静态检查
```bash
cd backend
go vet ./...
```

### 数据库检查
```bash
cd backend
go run cmd/app/main.go -check-db
```

## 部署前检查清单

- [ ] 执行数据库迁移（创建 ty_audit_logs 表）
- [ ] 运行 `go run cmd/app/main.go -check-db` 验证表结构
- [ ] 确认所有配置正确
- [ ] 测试文件上传功能
- [ ] 测试审计日志记录（查看 ty_audit_logs 表是否有数据）

## 数据库迁移命令

### PostgreSQL
```bash
psql -U postgres -d cntuanyuan -f backend/migrations/postgres/01_schema.sql
```

### MySQL
```bash
mysql -u root -p cntuanyuan < backend/migrations/mysql/01_schema.sql
```

## 注意事项

1. **审计日志表**：新部署时必须创建此表，否则审计功能无法工作
2. **定时任务**：默认未启动，需要在 `wire_gen.go` 中取消注释相关代码
3. **云存储**：OSS/COS 需要单独安装依赖并添加构建标签
