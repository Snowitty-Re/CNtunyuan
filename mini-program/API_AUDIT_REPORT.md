# 小程序与后端接口审查报告

> 审查日期: 2026-03-14
> 审查方法: 逐行对比小程序 `services/*.js` + `pages/*.js` 的请求字段/响应字段 vs 后端 `dto/*.go` 的 JSON tag 定义。

---

## 一、严重错误（功能直接报错/崩溃）

### E1. cases/create.js 提交字段名大面积不匹配
**文件**: `pages/cases/create.js:374-392`

| 小程序提交字段 | 后端期望字段 | 说明 |
|---|---|---|
| `missing_location` | **不存在** | 后端无此字段，应分别提交 `province`, `city`, `district`, `address` |
| `missing_latitude` | **不存在** | 后端 `CreateMissingPersonRequest` 无经纬度字段 |
| `missing_longitude` | **不存在** | 同上 |
| `missing_detail` | `description` | 后端字段名为 `description` |
| `case_type` | **不存在** | 后端 `CreateMissingPersonRequest` 无此字段 |
| `special_features` | `features` | 后端字段名为 `features` |
| `contact_relation` | `contact_rel` | 后端字段名为 `contact_rel` |
| `photos` (数组) | `photo_url` (字符串) | 后端只接受单个 `photo_url` 字符串，不接受照片数组 |

**影响**: 创建案件必定失败或数据丢失。

### E2. cases/create.js 编辑模式读取字段名不匹配
**文件**: `pages/cases/create.js:86-103` (loadCaseData)

| 小程序读取字段 | 后端实际返回字段 |
|---|---|
| `data.case_type` | **不存在** (后端无此字段) |
| `data.missing_location` | **不存在** (应从 `province+city+district+address` 拼接) |
| `data.missing_latitude` | **不存在** (响应无经纬度) |
| `data.missing_longitude` | **不存在** |
| `data.missing_detail` | `data.description` |
| `data.appearance` | **不存在** (后端无此字段) |
| `data.clothing` | `data.clothes` |
| `data.special_features` | `data.features` |
| `data.contact_relation` | `data.contact_rel` |

**影响**: 编辑模式加载的数据全部为空。

### E3. cases/detail.js 读取不存在的字段
**文件**: `pages/cases/detail.js:89-95`

| 小程序读取字段 | 后端实际字段 |
|---|---|
| `data.missing_latitude` | **不存在** |
| `data.missing_longitude` | **不存在** |
| `data.missing_location` | **不存在** |
| `data.possible_location` | **不存在** |
| `data.appearance` | **不存在** |
| `data.clothing` | `data.clothes` |
| `data.special_features` | `data.features` |

**影响**: 地图标记不显示、走失地点显示空白、外貌特征显示空白。

### E4. cases/detail.js addTrack 提交字段错误
**文件**: `pages/cases/detail.js:192-194`

小程序提交:
```js
{ description: '...', track_time: '...' }
```
后端期望 `CreateMissingPersonTrackRequest`:
```go
{ location, province, city, district, address, time, description, is_key_point, lat, lng }
```

- `track_time` -> 应为 `time`
- 缺少必要的 `location` 字段

**影响**: 添加线索接口调用可能失败。

### E5. cases/detail.js markFound 字段错误
**文件**: `pages/cases/detail.js:251-254`

小程序提交:
```js
{ found_location, found_time, description }
```
后端 `MarkFoundRequest` 期望:
```go
{ location, note }
```

- `found_location` -> `location`
- `found_time` -> **不存在**
- `description` -> `note`

**影响**: 标记已找到功能失败。

### E6. map/index.js 读取不存在的经纬度字段
**文件**: `pages/map/index.js:93-97`

小程序读取 `item.missing_latitude` / `item.missing_longitude`，但后端 `MissingPersonResponse` 中**没有任何经纬度字段**。

后端只有 `province`, `city`, `district`, `address` 文本地址。

**影响**: 地图页面所有标记点无法显示（filter 过滤掉全部数据）。

### E7. workbench.js loadTodayStats 字段名全部不匹配
**文件**: `pages/volunteer/workbench.js:116-118`

小程序读取:
```js
stats.pending_count, stats.processing_count, stats.helped_count, stats.completed_count
```
后端 `TaskStatsResponse` 实际返回:
```go
total, draft, pending, assigned, processing, completed, cancelled, overdue, my_tasks, my_pending, my_completed
```

**影响**: 工作台统计数据全部显示为 0。

### E8. volunteer/profile.js loadStats 字段名不匹配
**文件**: `pages/volunteer/profile.js:115-118`

小程序读取:
```js
userStats.task_count, userStats.case_count, userStats.dialect_count, userStats.points
```
后端 `UserStatsResponse` 实际返回:
```go
total_cases, active_cases, completed_cases, total_tasks, pending_tasks
```

- 无 `task_count`，应为 `total_tasks`
- 无 `case_count`，应为 `total_cases`
- **无 `dialect_count`**（后端用户统计不含方言数）
- **无 `points`**（后端无积分系统）

**影响**: 个人中心统计数据全部显示为 0。

### E9. tasks/list.js statusMap 缺少 assigned 和 draft
**文件**: `pages/tasks/list.js:22-26`

小程序 statusMap:
```js
{ pending: '待分配', processing: '进行中', completed: '已完成', cancelled: '已取消' }
```

后端任务状态包含: `draft`, `pending`, `assigned`, `processing`, `completed`, `cancelled`

- 缺少 `draft`（草稿）
- 缺少 `assigned`（已分配）
- `pending` 显示为"待分配"，但 `assigned` 才是"已分配"；`pending` 应该是"待处理"

**影响**: 状态为 `assigned` 或 `draft` 的任务显示为 undefined。

### E10. tasks/detail.js statusMap 同样缺少 assigned 和 draft
**文件**: `pages/tasks/detail.js:17-21`

同 E9 的问题。

**影响**: 任务详情页状态显示不正确。

---

## 二、中等错误（数据显示异常但不崩溃）

### E11. cases/list.js 列表显示缺 missing_location
**文件**: `pages/cases/list.js` WXML 引用 `item.missing_location`

后端返回的是 `province`, `city`, `district`, `address` 分开的字段，JS 层未做拼接，WXML 直接读 `missing_location` 会显示空白。

### E12. index/index.js dashboard 统计字段猜测可能不匹配
**文件**: `pages/index/index.js:124-141`

代码尝试读取 `dashboardStats.missing_persons.total` 等嵌套结构，但后端 dashboard/stats 的返回结构未确认是否嵌套。如果是扁平结构则取不到数据（已有 fallback 但 fallback 字段名也可能不对）。

### E13. cases/detail.wxml 引用不存在字段
**文件**: `pages/cases/detail.wxml`

WXML 中引用了 `caseData.missing_location`、`caseData.appearance`、`caseData.clothing` 等，但后端返回的字段是 `province+city+district+address`（分散）、无 `appearance`、`clothes`（非 `clothing`）。

### E14. dialect/create.js 创建方言缺少必填字段
**文件**: `pages/dialect/create.js:437-445`

小程序提交:
```js
{ title, description, audio_url, duration, region, tags, missing_person_id }
```
后端 `CreateDialectRequest` 期望:
```go
{ title, content, region, province, city, dialect_type, audio_url, duration, file_size, format, tags, description }
```

- `content` 未提交（binding 不要求 required 但建议传）
- `province`, `city` 未提交（只传了 `region`）
- `dialect_type` 未提交
- `file_size`, `format` 未提交
- `missing_person_id` 不在后端 DTO 中（可能被忽略）

**影响**: 方言记录可能创建成功但缺少关键分类信息。

### E15. upload.js batch 接口仍可能有问题
**文件**: `services/upload.js:26-30`

虽然已改为逐个调用 `/upload`，但 `request.js:314` 的 `uploadFiles` 函数仍引用 `/upload/batch` 作为默认值，如果有其他地方调用 `uploadFiles` 会走错端点。

### E16. upload.js download 方法 API_BASE_URL 硬编码
**文件**: `services/upload.js:41`

download 方法中硬编码了 `const API_BASE_URL = 'https://cntuanyuan.com/api/v1'`，但 `request.js:8` 当前配置为 `http://localhost:8080/api/v1`，两者不一致。

---

## 三、轻度问题（不会报错但逻辑有瑕疵）

### E17. tasks/detail.js 跳转不存在的 assign 页面
**文件**: `pages/tasks/detail.js:105`
```js
wx.navigateTo({ url: `/pages/tasks/assign?id=${this.data.taskId}` })
```
`pages/tasks/assign` 页面不存在于 `app.json` 中。

### E18. tasks/detail.js 位置字段名不一致
**文件**: `pages/tasks/detail.js:248`

读取 `task.latitude` / `task.longitude`，但后端 `TaskResponse` 字段为 `lat` / `lng`。

### E19. cases/detail.js 分享文本引用 missing_location
**文件**: `pages/cases/detail.js:328`
```js
title: `寻亲：${caseData.name}，${caseData.age}岁，${caseData.missing_location}`
```
`missing_location` 不存在，分享标题会显示 `undefined`。

### E20. cases/detail.js openLocation 引用 missing_location
**文件**: `pages/cases/detail.js:318`

`address: caseData.missing_location` -> 不存在的字段。

### E21. 后端无积分(points)系统
`volunteer/profile.js` 和 `volunteer/workbench.js` 多处引用 `points`，但后端 `UserResponse` 和 `UserStatsResponse` 均无 `points` 字段，始终显示 0。

### E22. 后端 UserResponse 无 org 嵌套对象
`volunteer/profile.js:94` 读取 `userInfo.org.name`，但后端 `UserResponse` 只有 `org_id` 和 `org_name`，无嵌套 `org` 对象。

### E23. request.js 每个请求都触发 showLoading
**文件**: `utils/request.js:129-133`

默认 `options.loading !== false` 时每次请求都会显示全局 Loading 遮罩层，导致并行请求出现 Loading 闪烁。页面 JS 中也有自己的 `showLoading` 调用，会出现双重 Loading。

---

## 四、新增问题（2026-03-14 更新）

### N1. 【严重】dialect/create.js tags 类型不匹配
**文件**: `pages/dialect/create.js:443`

小程序提交:
```js
tags: this.data.form.tags  // 数组: ['粤语', '广东话']
```
后端 `CreateDialectRequest` 期望:
```go
Tags string `json:"tags"`  // 字符串: "粤语,广东话"
```

**影响**: 方言创建时标签数据格式错误，后端可能无法正确解析或存储。

### N2. 【严重】cases/detail.js 轨迹时间字段不匹配
**文件**: `pages/cases/detail.js:132`, `pages/cases/detail.wxml:212`

小程序读取:
```js
track_time: formatDate(t.track_time)
```
后端 `MissingPersonTrackResponse` 返回:
```go
Time time.Time `json:"time"`  // 字段名为 time
```

**影响**: 轨迹列表时间显示为空白或 undefined。

### N3. 【中等】index/index.js 仪表盘字段结构问题
**文件**: `pages/index/index.js:124-141`

小程序尝试读取嵌套结构:
```js
dashboardStats.missing_persons.total
dashboardStats.tasks.pending
```
但后端 `/dashboard/stats` 返回的是多个独立的统计对象（扁平结构），不是嵌套结构。

**影响**: 首页仪表盘数据无法正常显示。

### N4. 【中等】cases/detail.wxml 更多字段引用错误
**文件**: `pages/cases/detail.wxml`

补充 E13 未列全的 WXML 字段引用错误:

| 行号 | WXML引用 | 后端实际字段 | 状态 |
|------|----------|-------------|------|
| 26 | `caseData.case_type` | ❌ 不存在 | 新增 |
| 98 | `caseData.missing_location` | ❌ 不存在 | E13已提及 |
| 102-104 | `caseData.possible_location` | ❌ 不存在 | 新增 |
| 107-109 | `caseData.missing_detail` | `description` | E13已提及 |
| 114-147 | `caseData.appearance` | ❌ 不存在 | E13已提及 |
| 129 | `caseData.clothing` | `clothes` | E13已提及 |
| 159 | `caseData.contact_relation` | `contact_rel` | 新增 |
| 169 | `caseData.contact_address` | ❌ 不存在 | 新增 |
| 212 | `item.track_time` | `time` | N2 |

### N5. 【中等】tasks/create.js assignee_id 字段问题
**文件**: `pages/tasks/create.js`

后端 `CreateTaskRequest` 无 `assignee_id` 字段，任务创建后需要单独调用 `/tasks/{id}/assign` 进行分配。如果小程序在创建时提交 `assignee_id`，该字段会被后端忽略。

**影响**: 创建任务时直接分配执行人的功能不生效。

### N6. 【轻度】app.js 用户数据处理问题
**文件**: `app.js`

登录后获取的用户信息中，`org` 对象和 `points` 字段后端不存在，但小程序代码中多处依赖这些字段，需要确保有适当的默认值和错误处理。

### N7. 【轻度】多个页面缺少错误边界处理
多个页面（如 `cases/create.js` 的 `loadCaseData`）在加载失败时只是打印错误，没有提供回退UI或重试机制，用户体验不佳。

---

## 五、汇总表

| 级别 | 编号 | 页面/文件 | 问题描述 |
|------|------|-----------|----------|
| **严重** | E1 | cases/create.js | 提交字段 8 处不匹配 |
| **严重** | E2 | cases/create.js | 编辑模式读取字段 9 处不匹配 |
| **严重** | E3 | cases/detail.js | 详情字段 7 处不匹配 |
| **严重** | E4 | cases/detail.js | addTrack 字段错误 |
| **严重** | E5 | cases/detail.js | markFound 字段错误 |
| **严重** | E6 | map/index.js | 后端无经纬度字段，地图标记全部为空 |
| **严重** | E7 | workbench.js | todayStats 字段名全错 |
| **严重** | E8 | profile.js | 用户统计字段名全错 |
| **严重** | E9 | tasks/list.js | 缺少 assigned/draft 状态 |
| **严重** | E10 | tasks/detail.js | 缺少 assigned/draft 状态 |
| **严重** | **N1** | **dialect/create.js** | **tags 数组类型应为字符串** |
| **严重** | **N2** | **cases/detail.js** | **轨迹时间字段名应为 time** |
| 中等 | E11 | cases/list.js | 列表无 missing_location 拼接 |
| 中等 | E12 | index/index.js | dashboard 统计字段结构不确定 |
| 中等 | E13 | cases/detail.wxml | WXML 引用不存在字段 |
| 中等 | E14 | dialect/create.js | 创建方言缺少分类字段 |
| 中等 | E15 | upload (request.js) | uploadFiles 默认走 /upload/batch |
| 中等 | E16 | upload.js | download 方法 URL 硬编码不一致 |
| 中等 | **N3** | **index/index.js** | **仪表盘字段结构不匹配** |
| 中等 | **N4** | **cases/detail.wxml** | **更多 WXML 字段引用错误** |
| 中等 | **N5** | **tasks/create.js** | **assignee_id 字段被后端忽略** |
| 轻度 | E17 | tasks/detail.js | 跳转不存在的 assign 页面 |
| 轻度 | E18 | tasks/detail.js | 位置用 latitude/longitude 而非 lat/lng |
| 轻度 | E19 | cases/detail.js | 分享标题引用 undefined 字段 |
| 轻度 | E20 | cases/detail.js | 导航 address 引用 undefined 字段 |
| 轻度 | E21 | profile/workbench | 后端无积分系统 |
| 轻度 | E22 | profile.js | 后端无 org 嵌套对象 |
| 轻度 | E23 | request.js | 全局 showLoading 与页面 Loading 冲突 |
| 轻度 | **N6** | **app.js** | **用户数据字段处理不完善** |
| 轻度 | **N7** | **多个页面** | **缺少错误边界处理** |

**共计 30 个问题：严重 12 个，中等 11 个，轻度 7 个。**

---

## 六、核心根因

后端 `MissingPersonResponse` **没有**以下字段，但小程序大量使用：

| 小程序使用的字段 | 后端实际对应 | 影响范围 |
|-----------------|-------------|---------|
| `missing_location` | 需从 `province + city + district + address` 拼接 | 4个页面 |
| `missing_latitude` | **不存在**（后端无经纬度） | 地图功能瘫痪 |
| `missing_longitude` | **不存在**（后端无经纬度） | 地图功能瘫痪 |
| `missing_detail` | `description` | 详情描述 |
| `appearance` | **不存在** | 外貌特征卡片 |
| `clothing` | `clothes` | 衣着描述 |
| `special_features` | `features` | 特殊特征 |
| `contact_relation` | `contact_rel` | 联系人关系 |
| `case_type` | **不存在** | 案件类型显示 |
| `possible_location` | **不存在** | 可能去向 |
| `contact_address` | **不存在** | 联系地址 |
| `photos` (数组提交) | `photo_url` (字符串) | 照片上传 |
| `track_time` | `time` | 轨迹时间 |

后端 `TaskStatsResponse` 字段为 `pending` / `processing` / `completed`（无 `_count` 后缀）。

后端 `UserStatsResponse` 字段为 `total_cases` / `total_tasks`（无 `task_count` / `case_count` / `dialect_count` / `points`）。

后端 `CreateDialectRequest` 的 `tags` 为字符串类型，小程序传的是数组。

---

## 七、修复建议优先级

### P0（立即修复）- 功能完全不可用
1. **E1, E2, E3, E6**: 案件创建/编辑/详情字段问题
2. **E4, E5**: 添加轨迹、标记找到功能
3. **N1**: 方言 tags 类型转换（数组转字符串）
4. **N2**: 轨迹时间字段（`track_time` -> `time`）

### P1（高优先级）- 数据显示异常
5. **E7, E8**: 工作台和个人中心统计字段
6. **E9, E10**: 任务状态映射（添加 `assigned`/`draft`）
7. **N3**: 首页仪表盘字段结构
8. **N4**: WXML 字段引用修正
9. **N5**: 任务创建后分配逻辑调整

### P2（中优先级）- 体验问题
10. **E11-E23, N6-N7**: 其他字段映射和UI优化问题

---

## 八、修复方案建议

### 方案一：后端修改（推荐）
在 `MissingPersonResponse` 中添加以下字段：
```go
MissingLocation  string  `json:"missing_location"`  // province+city+district+address 拼接
MissingLat       float64 `json:"missing_latitude"`   // 补充经纬度
MissingLng       float64 `json:"missing_longitude"`
CaseType         string  `json:"case_type"`          // 案件类型
Appearance       string  `json:"appearance"`         // 外貌特征
PossibleLocation string  `json:"possible_location"`  // 可能去向
```

### 方案二：小程序修改
在小程序端添加字段转换层，将所有后端返回的字段转换为小程序需要的格式。

### 方案三：双向适配
后端添加缺失的字段，同时小程序也做好字段映射的容错处理。
