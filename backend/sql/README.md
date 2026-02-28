# 数据库初始化指南

## 概述

团圆寻亲志愿者系统支持多种数据初始化方式：

1. **命令行工具初始化**（推荐）- 自动完成所有步骤
2. **SQL文件导入** - 适合DBA手动操作
3. **程序自动初始化** - 仅初始化基础结构

## 方式一：命令行工具初始化（推荐）

### 1. 数据库迁移（创建表结构）

```bash
cd backend
go run cmd/main.go -migrate
```

### 2. 初始化基础数据

**使用默认配置**（手机号: 13800138000, 密码: admin123）：

```bash
go run cmd/initdata/main.go -exec
```

**自定义超级管理员信息**：

```bash
go run cmd/initdata/main.go -exec \
  -phone="13912345678" \
  -email="admin@example.com" \
  -password="yourpassword"
```

### 3. 仅生成SQL文件（不执行）

```bash
# 生成SQL文件到 sql/init_generated.sql
go run cmd/initdata/main.go -gen

# 或指定输出路径
go run cmd/initdata/main.go -gen -o="/path/to/init.sql"

# 然后手动执行
psql -U postgres -d cntuanyuan -f sql/init_generated.sql
```

## 方式二：SQL文件手动导入

### 1. 数据库迁移

```bash
cd backend
go run cmd/main.go -migrate
```

### 2. 导入基础数据

```bash
# 使用默认SQL文件
psql -U postgres -d cntuanyuan -f sql/init.sql

# 或者使用生成的SQL（包含正确的密码哈希）
go run cmd/initdata/main.go -gen
psql -U postgres -d cntuanyuan -f sql/init_generated.sql
```

## 方式三：分步初始化

### 1. 数据库迁移

```bash
cd backend
go run cmd/main.go -migrate
```

### 2. 仅初始化根组织

```bash
go run cmd/main.go -init
```

### 3. 创建超级管理员

```bash
go run cmd/initdata/main.go -exec
```

## 文件说明

| 文件 | 说明 |
|------|------|
| `sql/init.sql` | SQL模板文件，包含变量占位符 |
| `sql/init_generated.sql` | 生成的SQL文件（包含实际的密码哈希） |
| `sql/schema.sql` | 数据库结构参考文档 |
| `cmd/initdata/main.go` | 数据初始化命令行工具 |

## 常见问题

### Q: 如何修改超级管理员密码？

```bash
# 方式1：使用初始化工具重新创建（会删除旧的）
go run cmd/initdata/main.go -exec -phone="13800138000" -password="newpassword"

# 方式2：通过后端API（登录后修改）
# 方式3：直接在数据库中更新（需要生成bcrypt哈希）
```

### Q: 如何添加其他管理员账号？

通过Web管理后台的"志愿者管理"功能添加，角色选择：
- `admin` - 普通管理员
- `manager` - 管理者

### Q: 初始化失败了怎么办？

1. 检查数据库连接配置（`config/config.yaml`）
2. 确保数据库已创建：`createdb -U postgres cntuanyuan`
3. 查看具体错误信息，常见原因：
   - 表已存在（跳过迁移，直接执行 `-init` 或 `initdata`）
   - 权限不足
   - 网络连接问题

## 生产环境建议

1. **修改默认密码**：首次部署后务必修改超级管理员默认密码
2. **使用环境变量**：生产环境建议通过环境变量传递敏感配置
3. **备份数据**：定期备份数据库
4. **审计日志**：启用操作日志记录

## 命令行参数参考

### initdata 工具参数

```
-config string
    配置文件路径 (默认 "config/config.yaml")
-email string
    超级管理员邮箱 (默认 "admin@cntunyuan.com")
-exec
    直接执行SQL初始化
-gen
    仅生成SQL文件，不执行
-o string
    生成SQL文件的输出路径 (默认 "sql/init_generated.sql")
-password string
    超级管理员初始密码 (默认 "admin123")
-phone string
    超级管理员手机号 (默认 "13800138000")
```

### 后端主程序参数

```
-init
    初始化基础数据（根组织）
-migrate
    执行数据库迁移（创建/更新表结构）
```
