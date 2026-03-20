# 团圆寻亲后端 — 综合技术评估报告

**评估日期**：2026-03-20
**代码版本**：commit `0250e7e`（含本轮接口对齐修复）
**评估范围**：全部 handler / service / entity / dto / repository 层
**评估人**：Claude Code 自动化审查

---

## 一、项目概览

| 指标 | 数据 |
|------|------|
| 语言 / 框架 | Go 1.21 / Gin + GORM + PostgreSQL |
| 架构模式 | DDD（domain → application → infrastructure → interfaces） |
| Handler 数量 | 9 个（auth / user / org / missing_person / task / dialect / upload / dashboard / notification） |
| Service 数量 | 8 个 |
| Entity 数量 | 14 个 |
| DTO 数量 | 9 个文件 |
| 测试文件 | 6 个（service / middleware / entity 层均覆盖） |
| **综合评分** | **8.4 / 10** |
| **可上线状态** | ⚠️ 需修复 P1 问题后方可上线 |

---

## 二、历次修复汇总（本轮之前）

| 轮次 | 类别 | 修复内容 |
|------|------|---------|
| Session 1 | 安全 | Shell 注入、SMS 轰炸、JWT 密钥校验、metrics 鉴权、文件上传 panic |
| Session 1 | 业务 | GetStats 计数 Bug、Task 所有权检查、Organization 自移动 |
| Session 1 | 验证 | 枚举值校验、UpdateStatus 先验证后查库、哨兵错误替换字符串比较 |
| Session 1 | 路由 | Gin 静态路由与参数路由冲突 |
| Session 2 | 接口对齐 | 走失坐标字段、is_liked、CompleteTask 附件、方言关联案件 |

---

## 三、本次评估发现的问题

### P1 — 功能错误（3 个）

#### #1 `UpdateProgress` 无所有权校验
**文件**：`internal/application/service/task_service.go:319`

任何已登录用户均可修改他人任务的进度，`userID` 参数传入后完全未使用：

```go
func (s *TaskAppService) UpdateProgress(ctx context.Context, id string, progress int, userID string) error {
    task, err := s.taskRepo.FindByID(ctx, id)
    // userID 未作任何校验，直接更新
    if err := task.UpdateProgress(progress); err != nil {
```

**影响**：志愿者可篡改其他人任务进度，审计日志失去意义。

**修复**：
```go
if task.AssigneeID == nil || *task.AssigneeID != userID {
    return ErrTaskNotAssignedToUser
}
```

---

#### #2 `GetStats` 活跃案件计数不准
**文件**：`internal/application/service/user_service.go:356`

`CompletedCases` 只统计 `reunited` 状态，未包含 `found` 状态，导致 `ActiveCases = TotalCases - CompletedCases` 偏高：

```go
completedCases, err := s.mpRepo.CountByReporterAndStatus(ctx, id, entity.MissingStatusReunited)
// found 状态被遗漏
stats.ActiveCases = stats.TotalCases - stats.CompletedCases // 结果偏大
```

**影响**：个人中心「活跃案件数」虚高，用户数据可信度下降。

**修复**：
```go
found, _ := s.mpRepo.CountByReporterAndStatus(ctx, id, entity.MissingStatusFound)
reunited, _ := s.mpRepo.CountByReporterAndStatus(ctx, id, entity.MissingStatusReunited)
stats.CompletedCases = found + reunited
```

---

#### #3 `MissingPerson` 状态流转无约束
**文件**：`internal/application/service/missing_person_service.go:260`

`UpdateStatus` 只校验枚举值合法性，不校验流转方向，允许如下非法状态跳转：
- `reunited → missing`（已团聚重新走失？）
- `closed → found`（已关闭案件被重新标记找到）
- `found → missing`（语义矛盾）

```go
mp.Status = newStatus  // 直接赋值，无流转校验
```

**影响**：案件状态流转混乱，统计报表失真。

**修复**：在 `entity.MissingPerson` 添加 `CanTransitionTo` 方法：
```go
func (m *MissingPerson) CanTransitionTo(s MissingStatus) bool {
    if m.Status == MissingStatusClosed  { return false }
    if m.Status == MissingStatusReunited && s != MissingStatusClosed { return false }
    return true
}
```
并在服务层调用前校验。

---

### P2 — 边界缺失（4 个）

#### #4 组织树移动允许环形引用
**文件**：`internal/application/service/organization_service.go:230`

当前只阻止「移动到自身」，不阻止「移动到自己的子孙节点」（间接循环）。
例如树 A→B→C，将 A 移到 C 之下会产生环 C→A→B→C。

**修复**：移动前沿目标节点向上遍历祖先链，若遇到 `id` 则拒绝。

---

#### #5 Task 可在截止日期过后开始
**文件**：`internal/domain/entity/task.go:122`

```go
func (t *Task) CanStart() bool {
    return t.Status == TaskStatusAssigned  // 无截止日期检查
}
```

过期任务仍可被执行人标记为「进行中」，完成时间戳失去时效性。

**修复**：
```go
if t.Deadline != nil && time.Now().After(*t.Deadline) {
    return false
}
```

---

#### #6 微信登录临时用户写入空手机号
**文件**：`internal/domain/service/auth_service.go:231`

微信登录创建临时用户时 `Phone: ""`，但 `entity.User.Validate()` 的 `ValidatePhone()` 要求手机号必须匹配 `^1[3-9]\d{9}$`。临时用户通过直接赋值绕过 `NewUser()` 的校验，若 `BindPhone` 流程未完成，DB 中会留存无效手机号记录。

**修复**：`Validate()` 中当 `Phone == ""` 且 `WxOpenID != ""` 时跳过手机号校验（标记为待绑定状态）。

---

#### #7 手机号格式校验仅在 Handler 层
**文件**：`internal/interfaces/http/handler/auth_handler.go:390`

`BindPhone` 的手机号格式校验在 Handler 做，Service 层无校验。若 Service 被直接调用（内部流程、测试、未来其他入口），则无保障。

**修复**：在 `auth_service.BindPhone()` 开头加格式校验，Handler 层校验可保留（防御纵深）。

---

### P3 — 质量问题（2 个）

#### #8 Like/Unlike 非原子操作
**文件**：`internal/application/service/dialect_service.go:205–230`

`HasLiked` → `AddLike` → `IncrementLikeCount` 三步非事务执行：
- 高并发下两个请求可同时通过 `HasLiked` 检查（`DialectLike` 唯一索引会拦截第二次写，但 `LikeCount` 已被递增一次多）
- `IncrementLikeCount` 失败时错误被 `logger.Error` 记录但不回滚，点赞记录存在但计数未同步

**说明**：`DialectLike` 已有 `idx_dialect_user unique` 约束，重复点赞的 DB 写入会失败，不会产生重复数据。但计数与实际记录数可能长期偏差。

**修复**：将三步操作包在同一事务中。

---

#### #9 轨迹时间无合理性校验
**文件**：`internal/application/service/missing_person_service.go:317`

`AddTrack` 接受任意时间，允许提交「走失前的目击轨迹」或「未来时间的目击记录」：

```go
track.Time = req.Time  // 无范围限制
```

**修复**：
```go
if req.Time.After(time.Now()) {
    return nil, fmt.Errorf("目击时间不能是未来时间")
}
```

---

## 四、已修复问题确认（本次评估中验证）

以下问题在上一轮修复中已正确实现，本次评估确认有效：

| # | 问题 | 修复状态 |
|---|------|---------|
| 走失人员坐标字段 | entity/DTO/service 全链路新增 `missing_latitude/missing_longitude` | ✅ |
| 方言 `is_liked` | `GetByID` 接受 `userID`，查询 `HasLiked()` 并写入响应 | ✅ |
| `CompleteTaskRequest` 附件 | `Feedback`/`Attachments` 字段新增，写入 Task 实体 | ✅ |
| 方言关联案件 | `MissingPersonID` 贯穿 entity/DTO/service | ✅ |
| JWT 密钥长度校验 | `NewJWTService` 返回 error，`wire_gen.go` 处理 | ✅ |
| SMS 轰炸防护 | Redis 60s 限速 | ✅ |
| Task 所有权校验 | `Start()`/`Complete()` 均检查 `AssigneeID == userID` | ✅ |
| 枚举值校验 | `IsValidMissingStatus`/`IsValidUrgencyLevel` 在服务层校验 | ✅ |
| Gin 路由冲突 | 静态路由注册在参数路由前 | ✅ |

---

## 五、架构质量评估

### 分层清晰度
- **Domain 层**：实体逻辑（Validate / MarkFound / Complete 等业务方法）完整，无基础设施依赖
- **Application 层**：DTO 转换集中，无跨服务直接调用
- **Infrastructure 层**：GORM 实现与 Repository 接口解耦，测试可替换
- **Interface 层**：Handler 职责单一，仅处理 HTTP 绑定和参数提取

### 安全性
- JWT 密钥强度有校验，Token 刷新逻辑正确
- 敏感字段（密码、Token）未出现在日志中
- 文件上传有 MIME 类型 + 扩展名双重校验

### 可测试性
- Service 层测试覆盖主要业务路径
- `testutil` 提供 SQLite 内存库，快速运行
- Handler 层暂无测试（需补充）

---

## 六、最终评分

| 维度 | 评分 | 说明 |
|------|------|------|
| API 接口对齐 | 9.0 / 10 | 本轮修复后坐标、is_liked、附件字段已对齐 |
| 业务逻辑正确性 | 7.5 / 10 | 状态流转无约束、进度更新无鉴权、统计计数偏差 |
| 安全性 | 9.0 / 10 | 主要漏洞已修复，BindPhone 校验层级需补强 |
| 代码质量 | 8.5 / 10 | DDD 架构整洁，P3 问题为非原子操作和缺少时间验证 |
| 测试覆盖率 | 7.0 / 10 | Service 层有测试，Handler 层空白 |
| **综合评分** | **8.4 / 10** | |

**可上线状态**：⚠️ **修复 P1 #1（UpdateProgress 鉴权）和 P1 #2（GetStats 计数）后方可上线**

---

## 七、修复优先级

| 优先级 | 问题 | 预估影响 |
|--------|------|---------|
| 🔴 立即修复 | #1 UpdateProgress 无鉴权 | 任意用户可篡改他人任务 |
| 🔴 立即修复 | #2 GetStats 活跃案件计数偏高 | 用户数据错误 |
| 🟡 上线前修复 | #3 状态流转无约束 | 数据状态不可信 |
| 🟡 上线前修复 | #4 组织树环形引用 | 潜在无限循环 |
| 🟢 迭代修复 | #5–#9 | 边界校验和质量问题 |

---

*本报告由 Claude Code 自动化审查生成，覆盖 backend commit `0250e7e` 全部源码。*
