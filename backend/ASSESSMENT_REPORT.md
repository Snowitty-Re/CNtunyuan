# 团圆寻亲后端 — 综合技术评估报告（第二轮）

**评估日期**：2026-03-20
**代码版本**：commit `8b9d1e6`（含第一轮 9 项修复）
**评估轮次**：第二轮
**评估人**：Claude Code 自动化审查

---

## 一、项目概览

| 指标 | 数据 |
|------|------|
| **综合评分** | **8.8 / 10** |
| **可上线状态** | ⚠️ 修复 P1 问题后可上线 |

---

## 二、历次修复汇总

### 第一轮（Session 1）— 安全与核心 Bug
- Shell 注入、SMS 轰炸防护、JWT 密钥长度校验、metrics 鉴权
- 文件上传 panic、GetStats 计数 Bug、Task 所有权检查、Organization 自移动
- 枚举值校验、UpdateStatus 先验证后查库、哨兵错误

### 第二轮（Session 2）— 接口对齐
- 走失坐标字段（map 页无标记点）
- `is_liked` 字段（点赞状态恒 false）
- `CompleteTaskRequest` 附件字段
- 方言关联案件 `missing_person_id`

### 第三轮（Session 3）— 业务逻辑完整性
- `UpdateProgress` 无所有权校验（P1）
- `GetStats` 活跃案件计数偏高（P1）
- 状态流转无约束（P1）
- 组织树环形引用（P2）
- Task 截止日期检查（P2）
- 微信临时用户空手机号（P2）
- BindPhone 校验层级（P2）
- Like/Unlike 计数原子性（P3）
- 轨迹时间合理性校验（P3）

---

## 三、本轮发现的问题

### P1 — 权限与所有权缺失（3 个）

#### #1 Dialect `Update`/`Delete` 无所有权校验
**文件**：`internal/interfaces/http/handler/dialect_handler.go:41-42`
`internal/application/service/dialect_service.go:116, 158`

路由仅要求登录（`authMiddleware.Required()`），任何志愿者均可修改或删除他人上传的方言录音：

```go
// dialect_handler.go
dialects.PUT("/:id", authMiddleware.Required(), h.Update)    // 无所有权限制
dialects.DELETE("/:id", authMiddleware.Required(), h.Delete) // 无所有权限制

// dialect_service.go — Update() 中无 userID 参数
func (s *DialectAppService) Update(ctx context.Context, id string, req *dto.UpdateDialectRequest) ...
```

**影响**：任意用户可覆盖他人方言内容，或删除他人上传数据。

**修复**：`Update`/`Delete` 方法增加 `uploaderID` 参数；服务层检查 `d.UploaderID == userID`（管理员跳过）。

---

#### #2 `MissingPerson.Update` 无所有权/角色校验
**文件**：`internal/interfaces/http/handler/missing_person_handler.go:35`
`internal/application/service/missing_person_service.go:158`

`PUT /missing-persons/:id` 路由无角色守卫，任意已登录用户可修改任意案件：

```go
mps.PUT("/:id", h.Update)    // 无 RequireManager() 或所有权检查
```

对比：`DELETE`、`UpdateStatus`、`MarkFound` 均有 `middleware.RequireManager()`。

**影响**：志愿者可篡改其他人或管理员创建的走失案件信息（联系电话、照片等）。

**修复**：服务层 `Update()` 增加 `userID` 参数，校验 `mp.ReporterID == userID || isManager`；或在路由层加 `middleware.RequireManager()`。

---

#### #3 `DialectService.UpdateStatus` 无状态枚举校验
**文件**：`internal/application/service/dialect_service.go:167`

`UpdateStatus` 直接将任意字符串转为 `entity.DialectStatus`，无合法性校验，而其他服务（user/missing_person）均有：

```go
d.Status = entity.DialectStatus(status)  // 接受任意字符串，如 "hacked"
```

对比：
- `missing_person_service.go:258`：`if !entity.IsValidMissingStatus(newStatus)`
- `user_service.go:204`：`if !entity.IsValidUserStatus(status)`

**影响**：数据库写入非法状态值（DB 约束会报 500，而非返回 400 给客户端）。

**修复**：在 `entity/dialect.go` 新增 `IsValidDialectStatus()`，服务层调用校验。

---

### P2 — 功能缺失与边界问题（5 个）

#### #4 `Dialect.List` 和 `GetComments` 缺少分页校验
**文件**：`internal/application/service/dialect_service.go:88, 304`

`List()` 和 `GetComments()` 均未调用 `validator.SanitizePagination()`，客户端可传入 `page_size=999999`：

```go
query.Page = req.Page       // 未校验
query.PageSize = req.PageSize  // 未校验
```

对比：`missing_person_service.go:129`、`task_service.go:91` 均有 `SanitizePagination()`。

**影响**：超大分页查询可拖垮数据库。

**修复**：在 `List()` 开头加：`req.Page, req.PageSize = validator.SanitizePagination(req.Page, req.PageSize)`；在 `GetComments()` 开头加相同逻辑。

---

#### #5 `Unlike` 回滚使用新 UUID，违反唯一约束
**文件**：`internal/application/service/dialect_service.go:265`

`DecrementLikeCount` 失败后，回滚创建新 `DialectLike` 时生成了新的 UUID，但 `(dialect_id, user_id)` 已有唯一索引：

```go
rollback := &entity.DialectLike{
    ID:        uuid.New().String(),  // 新 UUID — 正确
    DialectID: dialectID,
    UserID:    userID,               // 同一 user+dialect，会被唯一索引拦截
}
```

实际上 `(dialect_id, user_id)` 的唯一约束已被 `RemoveLike` 删除，再插入新记录是允许的，UUID 不同不是问题。此 Bug 属于误判，实际逻辑正确。

---

#### #6 `AddComment` 未验证方言是否存在
**文件**：`internal/application/service/dialect_service.go:285`

`AddComment` 不检查 `dialectID` 对应的方言是否存在，直接写入评论；若 FK 约束触发，返回 500 而非有意义的 404：

```go
// 无 FindByID 检查
if err := s.dialectRepo.AddComment(ctx, comment); err != nil {
    return nil, err  // 返回原始 DB 错误，客户端无法区分 400/404/500
}
```

对比：`Like()` 在操作前有 `FindByID` 检查。

**影响**：客户端收到 500 而非 404，无法正确处理「方言不存在」的情况。

**修复**：
```go
if _, err := s.dialectRepo.FindByID(ctx, dialectID); err != nil {
    return nil, ErrDialectNotFound
}
```

---

#### #7 `Task.CanStart()` 未检查 `AssigneeID` 非空
**文件**：`internal/domain/entity/task.go:122`

任务在 `assigned` 状态下理论上必有执行人，但 `CanStart()` 未做防御性检查：

```go
func (t *Task) CanStart() bool {
    if t.Status != TaskStatusAssigned { return false }
    if t.Deadline != nil && time.Now().After(*t.Deadline) { return false }
    return true  // 若 AssigneeID 意外为 nil 仍返回 true
}
```

service 层的 `Start()` 会二次校验，但 entity 层方法语义不完整。

**修复**：
```go
if t.AssigneeID == nil { return false }
```

---

#### #8 `GetComments` handler 分页参数转换错误无处理
**文件**：`internal/interfaces/http/handler/dialect_handler.go`（GetComments handler）

```go
page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
```

`strconv.Atoi` 失败时 `page`/`pageSize` 为 0，未做零值防护。传入字符串如 `page=abc` 时 `page=0`，后续分页计算可能出现除零或越界。

**修复**：转换失败后使用默认值：
```go
if page <= 0 { page = 1 }
if pageSize <= 0 { pageSize = 10 }
```

---

## 四、已确认正确项（本轮评估澄清）

| 项目 | 结论 |
|------|------|
| `Unlike` 回滚 UUID 问题 | 新 UUID + 相同 (dialectID, userID) 在删除后可正常插入，逻辑正确 |
| `DialectService.UpdateStatus` 角色守卫 | 路由层已有 `middleware.RequireManager()`，管理员限制有效 |
| `MissingPerson.Delete/UpdateStatus` 权限 | 路由层有 `RequireManager()`，正确 |

---

## 五、最终评分

| 维度 | 评分 | 趋势 |
|------|------|------|
| API 接口对齐 | 9.5 / 10 | ↑（坐标、is_liked、附件均已修复） |
| 业务逻辑正确性 | 8.5 / 10 | ↑（状态流转、进度权限已修复） |
| 安全性 | 8.0 / 10 | →（Dialect/MissingPerson 写操作缺少所有权校验） |
| 代码质量 | 9.0 / 10 | ↑ |
| 测试覆盖率 | 7.5 / 10 | ↑（AddTrack/UpdateProgress 测试已修正） |
| **综合评分** | **8.8 / 10** | ↑（上轮 8.4） |

**可上线状态**：⚠️ **修复 #1（Dialect 所有权）和 #2（MissingPerson 写权限）后方可上线**

---

## 六、修复优先级

| 优先级 | # | 问题 |
|--------|---|------|
| 🔴 上线前必修 | #1 | Dialect Update/Delete 无所有权校验 |
| 🔴 上线前必修 | #2 | MissingPerson Update 任意用户可改 |
| 🟡 上线前建议修 | #3 | DialectStatus 无枚举校验 |
| 🟡 上线前建议修 | #4 | List/GetComments 分页无上限 |
| 🟢 迭代修复 | #6 | AddComment 无 dialect 存在检查 |
| 🟢 迭代修复 | #7 | CanStart() 防御性 nil 检查 |
| 🟢 迭代修复 | #8 | GetComments handler 零值防护 |

---

*本报告由 Claude Code 自动化审查生成，覆盖 backend commit `8b9d1e6` 全部源码。*
