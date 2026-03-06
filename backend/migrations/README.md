# 团圆寻亲系统 - 数据库迁移指南

本文档说明如何手动执行数据库初始化和迁移。

## 目录结构

```
migrations/
├── postgres/           # PostgreSQL 迁移文件
│   ├── 01_schema.sql  # 表结构
│   └── 02_seed.sql    # 种子数据
├── mysql/             # MySQL 迁移文件
│   ├── 01_schema.sql  # 表结构
│   └── 02_seed.sql    # 种子数据
└── README.md          # 本文档
```

## 快速开始

### 1. 创建数据库

**PostgreSQL:**
```bash
# 连接到 PostgreSQL
psql -U postgres

# 创建数据库
CREATE DATABASE cntuanyuan WITH ENCODING = 'UTF8';

# 或使用命令行
createdb -U postgres -E UTF8 cntuanyuan
```

**MySQL:**
```bash
# 连接到 MySQL
mysql -u root -p

# 创建数据库
CREATE DATABASE cntuanyuan CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

# 或使用命令行
mysql -u root -p -e "CREATE DATABASE cntuanyuan CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
```

### 2. 执行迁移脚本

**PostgreSQL:**
```bash
# 执行表结构
psql -U postgres -d cntuanyuan -f migrations/postgres/01_schema.sql

# 插入种子数据
psql -U postgres -d cntuanyuan -f migrations/postgres/02_seed.sql
```

**MySQL:**
```bash
# 执行表结构
mysql -u root -p cntuanyuan < migrations/mysql/01_schema.sql

# 插入种子数据
mysql -u root -p cntuanyuan < migrations/mysql/02_seed.sql
```

### 3. 验证安装

```bash
# 检查数据库表结构
cd backend
go run cmd/app/main.go -check-db
```

## 表结构说明

### 核心表

| 表名 | 说明 | 主要字段 |
|------|------|----------|
| ty_organizations | 组织表 | id, name, code, type, parent_id |
| ty_users | 用户表 | id, nickname, phone, email, password, role, org_id |
| ty_permissions | 权限表 | id, name, code, resource, action |
| ty_user_permissions | 用户权限关联表 | user_id, permission_id |

### 业务表

| 表名 | 说明 | 主要字段 |
|------|------|----------|
| ty_missing_persons | 走失人员表 | id, name, gender, missing_time, status, urgency |
| ty_missing_person_tracks | 走失人员轨迹表 | id, missing_person_id, location, time, description |
| ty_tasks | 任务表 | id, title, type, status, priority, creator_id, assignee_id |
| ty_task_attachments | 任务附件表 | id, task_id, file_name, file_url |
| ty_task_logs | 任务日志表 | id, task_id, action, old_status, new_status |
| ty_task_comments | 任务评论表 | id, task_id, user_id, content, parent_id |

### 方言模块

| 表名 | 说明 | 主要字段 |
|------|------|----------|
| ty_dialects | 方言表 | id, title, region, audio_url, duration, uploader_id |
| ty_dialect_comments | 方言评论表 | id, dialect_id, user_id, content |
| ty_dialect_likes | 方言点赞表 | id, dialect_id, user_id |
| ty_dialect_play_logs | 方言播放记录表 | id, dialect_id, user_id, duration |

### 文件模块

| 表名 | 说明 | 主要字段 |
|------|------|----------|
| ty_files | 文件表 | id, file_name, file_type, path, url, storage_type |
| ty_org_stats | 组织统计表 | id, org_id, total_volunteers, total_cases, total_tasks |

## 默认账号

执行种子数据后，系统会创建以下默认账号：

- **超级管理员**
  - 手机号: `13800138000`
  - 密码: `admin123`

## 开发环境快速迁移

如果你正在开发环境，可以使用 GORM 的 AutoMigrate 功能（不推荐用于生产环境）：

```bash
# 自动创建表结构
go run cmd/app/main.go -migrate

# 插入种子数据
go run cmd/app/main.go -seed
```

## 数据种子工具

使用 Go 编写的数据种子工具可以生成测试数据：

```bash
# 生成所有类型的数据（各50条）
go run cmd/seed/main.go -all

# 生成指定数量的数据
go run cmd/seed/main.go -all -count 100

# 只生成特定类型的数据
go run cmd/seed/main.go -orgs      # 组织
go run cmd/seed/main.go -users     # 用户
go run cmd/seed/main.go -cases     # 走失人员
go run cmd/seed/main.go -dialects  # 方言
go run cmd/seed/main.go -tasks     # 任务

# 清理现有数据后重新生成
go run cmd/seed/main.go -all -clean
```

## 数据库配置

编辑 `backend/config/config.yaml`：

```yaml
database:
  type: postgres        # 数据库类型: postgres 或 mysql
  host: localhost
  port: 5432            # PostgreSQL 默认 5432, MySQL 默认 3306
  user: postgres        # PostgreSQL 默认 postgres, MySQL 默认 root
  password: "your-password"
  database: cntuanyuan
  ssl_mode: disable     # PostgreSQL 专用
  charset: UTF8         # MySQL 使用 utf8mb4
  max_idle_conns: 10
  max_open_conns: 100
  conn_max_lifetime: 3600
```

## 注意事项

1. **生产环境**: 请务必使用 SQL 文件手动执行迁移，不要使用 GORM AutoMigrate
2. **字符集**: PostgreSQL 使用 UTF8，MySQL 使用 utf8mb4 以支持 Emoji
3. **外键约束**: 所有外键都设置了适当的级联操作（CASCADE/SET NULL/RESTRICT）
4. **软删除**: 所有表都支持软删除（deleted_at 字段）
5. **索引**: 已为常用查询字段创建索引，优化查询性能

## 故障排除

### 连接失败
- 检查数据库服务是否启动
- 检查配置文件中的连接信息
- 检查防火墙设置

### 表已存在
- 如果是重新安装，先删除现有表或使用 `-clean` 参数
- 生产环境请谨慎操作，建议先备份数据

### 外键约束错误
- 确保按正确顺序执行 SQL 文件（先执行 01_schema.sql，再执行 02_seed.sql）
- 种子数据依赖表结构，必须在表结构创建后执行
