# 数据生命周期策略（后端）

本文档定义后端各模块的数据删除策略与约束策略，目标是避免“策略不统一 + 约束不配套”导致的运行时问题。

## 1. 总体原则

1. 业务主数据默认使用 `deleted_at` 软删除。
2. 清理类数据（如审计日志）按保留期做物理删除。
3. 可空唯一字段统一使用 `NULL` 表示“无值”，禁止空字符串 `""` 占用唯一约束。
4. 唯一约束必须与删除策略配套：
   - PostgreSQL：使用 `WHERE deleted_at IS NULL` 的部分唯一索引。
   - MySQL：使用生成列 + 唯一索引，仅约束未删除记录。

## 2. 表级策略

- `ty_users`：软删除（`deleted_at`）；手机号/邮箱/微信 OpenID 对未删除记录唯一。
- `ty_organizations`：当前实现为硬删除（释放编码），并由外键约束拦截被引用删除。
- `ty_missing_persons`：软删除。
- `ty_missing_person_tracks`：软删除。
- `ty_tasks`：软删除。
- `ty_dialects`：软删除。
- `ty_files`：软删除（仅使用 `deleted_at`）。
- `ty_audit_logs`：按保留期物理删除（定时任务）。

## 3. 代码约束

1. 仓储查询默认不返回软删除数据（依赖 GORM `deleted_at` 默认作用域）。
2. 删除接口需显式说明：
   - 软删除：标记 `deleted_at`
   - 硬删除：`Unscoped().Delete(...)`
3. 文件删除执行顺序：
   - 先软删数据库记录
   - 再尝试删除物理文件（失败仅告警）

## 4. 迁移要求

执行 `04_data_lifecycle_alignment.sql` 与 `06_schema_consistency_and_performance.sql`（PostgreSQL / MySQL）后，需验证：

1. `ty_files` 仅使用 `deleted_at` 作为软删除标记（`is_deleted` 已移除）。
2. `ty_users` 唯一约束对软删除记录不再占用。
3. `ty_missing_person_tracks` / `ty_tasks` 的经纬度约束已生效。
4. `ty_audit_logs.user_id` 外键约束存在且删除用户时置空。
5. 高频查询复合索引已创建（users/missing_persons/tasks）。

## 5. 回归检查清单

1. 用户删除后重建同手机号：应成功。
2. 文件删除后按 ID 获取：应返回不存在。
3. 审计日志清理任务执行后：行数真实下降。
4. 组织删除时若仍被引用：应被外键拦截并返回业务错误。

## 6. 权限模型说明

当前运行时权限模型采用 **RBAC（角色层级）**：

- `super_admin > admin > manager > volunteer`
- 鉴权基于角色中间件（`RequireRole/RequireAdmin/RequireManager/RequireSuperAdmin`）
- `ty_permissions` / `ty_user_permissions` 当前作为权限字典与扩展保留，不作为运行时判定主路径
