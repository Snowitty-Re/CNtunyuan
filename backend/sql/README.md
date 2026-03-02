# 数据库脚本说明

## 文件列表

| 文件名 | 说明 |
|--------|------|
| `create_database.sql` | 创建数据库脚本（需在 postgres 数据库执行） |
| `init_database.sql` | 初始化扩展脚本（需在 cntunyuan 数据库执行） |
| `create_tables.sql` | 表结构创建脚本 |

## 使用说明

### 1. 创建数据库（首次安装）

**方式一：使用脚本创建**

```bash
# 创建数据库（在 postgres 数据库中执行）
psql -U postgres -f backend/sql/create_database.sql

# 初始化扩展（在 cntunyuan 数据库中执行）
psql -U postgres -d cntunyuan -f backend/sql/init_database.sql
```

**方式二：手动执行**

```bash
# 连接到 postgres 数据库
psql -U postgres

# 创建数据库
CREATE DATABASE cntunyuan WITH ENCODING = 'UTF8';

# 连接到新数据库
\c cntunyuan

# 启用扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
```

### 2. 创建表结构

**方式一：使用 GORM AutoMigrate（推荐）**

```bash
cd backend
go run cmd/app/main.go -migrate
```

**方式二：使用 SQL 脚本**

```bash
psql -U postgres -d cntunyuan -f backend/sql/create_tables.sql
```

### 3. 检查数据库编码

```sql
-- 检查数据库编码
SELECT datname, pg_encoding_to_char(encoding) 
FROM pg_database 
WHERE datname = 'cntunyuan';

-- 应该返回 UTF8

-- 检查客户端编码
SHOW client_encoding;

-- 应该返回 UTF8
```

## 字符集配置

### 后端配置 (config.yaml)

```yaml
database:
  host: localhost
  port: 5432
  user: postgres
  password: yourpassword
  database: cntunyuan
  ssl_mode: disable
  charset: UTF8  # 数据库字符集
```

### PostgreSQL 服务器配置

确保 PostgreSQL 服务器配置使用 UTF-8：

```bash
# 查看服务器编码
psql -U postgres -c "SHOW server_encoding;"

# 查看客户端编码
psql -U postgres -c "SHOW client_encoding;"
```

如果服务器编码不是 UTF8，需要在 `postgresql.conf` 中修改：

```conf
# 在 postgresql.conf 中
client_encoding = utf8
```

## 表结构说明

### 核心表

| 表名 | 说明 | 主要字段 |
|------|------|----------|
| `ty_users` | 用户表 | id, nickname, phone, email, role, status |
| `ty_organizations` | 组织表 | id, name, code, type, parent_id |
| `ty_missing_persons` | 走失人员表 | id, name, status, missing_time |
| `ty_dialects` | 方言表 | id, title, region, audio_url |
| `ty_tasks` | 任务表 | id, title, status, assignee_id |
| `ty_files` | 文件表 | id, file_name, file_type, url |

### 关联表

| 表名 | 说明 | 关联 |
|------|------|------|
| `ty_user_permissions` | 用户权限关联 | users <-> permissions |
| `ty_dialect_comments` | 方言评论 | dialects <- users |
| `ty_dialect_likes` | 方言点赞 | dialects <- users |
| `ty_task_attachments` | 任务附件 | tasks |
| `ty_task_logs` | 任务日志 | tasks <- users |
| `ty_missing_person_tracks` | 轨迹记录 | missing_persons <- users |

## 索引说明

所有表都包含以下标准索引：
- 主键索引（UUID）
- 外键索引
- 状态字段索引
- 创建时间索引

## 软删除

支持软删除的表包含 `deleted_at` 字段，使用 GORM 的软删除功能。

## 编码注意事项

1. **UTF-8 支持**：所有文本字段都支持完整的 UTF-8 编码，包括中文和 Emoji
2. **字符串长度**：VARCHAR 长度按字符计算，不是字节
3. **TEXT 类型**：大文本内容使用 TEXT 类型，无长度限制

## 常见问题

### Q: 数据库连接出现编码错误？
A: 确保 `config.yaml` 中的 `charset` 设置为 `UTF8`

### Q: 如何修改现有数据库的编码？
A: 需要导出数据，重新创建数据库，再导入数据：

```bash
# 导出数据
pg_dump -U postgres cntunyuan > backup.sql

# 删除并重新创建数据库
psql -U postgres -c "DROP DATABASE cntunyuan;"
psql -U postgres -c "CREATE DATABASE cntunyuan WITH ENCODING = 'UTF8';"

# 导入数据
psql -U postgres cntunyuan < backup.sql
```
