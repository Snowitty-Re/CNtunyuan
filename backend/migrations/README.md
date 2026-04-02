# 助力团圆系统 - 数据库迁移指南

本文档定义当前统一的数据库初始化与升级流程。

## 迁移策略

- `00_bootstrap.sql`：**全新数据库首次初始化**（一键完成 Schema + 对齐补丁 + 任务跟进工作流 + 默认种子数据）
- `01~06`：历史增量脚本，保留用于旧环境升级与审计追溯

## 目录结构

```
migrations/
├── postgres/
│   ├── 00_bootstrap.sql
│   ├── 01_schema.sql
│   ├── 02_seed.sql
│   ├── 03_missing_person_geo_case_type.sql
│   ├── 04_data_lifecycle_alignment.sql
│   ├── 05_task_follow_up_workflow.sql
│   └── 06_schema_consistency_and_performance.sql
├── mysql/
│   ├── 00_bootstrap.sql
│   ├── 01_schema.sql
│   ├── 02_seed.sql
│   ├── 03_missing_person_geo_case_type.sql
│   ├── 04_data_lifecycle_alignment.sql
│   ├── 05_task_follow_up_workflow.sql
│   └── 06_schema_consistency_and_performance.sql
└── README.md
```

## 首次初始化（推荐）

### PostgreSQL

```bash
createdb -U postgres -E UTF8 cntuanyuan
psql -U postgres -d cntuanyuan -f backend/migrations/postgres/00_bootstrap.sql
```

### MySQL

```bash
mysql -u root -p -e "CREATE DATABASE cntuanyuan CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
mysql -u root -p cntuanyuan < backend/migrations/mysql/00_bootstrap.sql
```

## 旧环境增量升级（仅历史库）

如果数据库是早期版本（曾执行过旧版 `01/02`），按顺序执行缺失脚本：

### PostgreSQL

```bash
psql -U postgres -d cntuanyuan -f backend/migrations/postgres/03_missing_person_geo_case_type.sql
psql -U postgres -d cntuanyuan -f backend/migrations/postgres/04_data_lifecycle_alignment.sql
psql -U postgres -d cntuanyuan -f backend/migrations/postgres/05_task_follow_up_workflow.sql
psql -U postgres -d cntuanyuan -f backend/migrations/postgres/06_schema_consistency_and_performance.sql
```

### MySQL

```bash
mysql -u root -p cntuanyuan < backend/migrations/mysql/03_missing_person_geo_case_type.sql
mysql -u root -p cntuanyuan < backend/migrations/mysql/04_data_lifecycle_alignment.sql
mysql -u root -p cntuanyuan < backend/migrations/mysql/05_task_follow_up_workflow.sql
mysql -u root -p cntuanyuan < backend/migrations/mysql/06_schema_consistency_and_performance.sql
```

## 验证

```bash
cd backend
go run cmd/app/main.go -check-db
```

## 默认账号

- 手机号：`13800138000`
- 密码：`admin123`

## 注意事项

1. 新环境只执行 `00_bootstrap.sql`，不要再重复执行 `01~06`
2. 升级环境只执行“未执行过”的增量脚本
3. 生产环境迁移前务必先做数据库备份
