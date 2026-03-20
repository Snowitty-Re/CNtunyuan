# 小程序 API 字段对齐修复报告

## 修复概览

根据 API 审计报告，已修复小程序中所有与后端 API 字段不匹配的问题，共计修复 **30 个问题**。

---

## 修复详情

### 🔴 关键修复（Critical - 12个）

| 文件 | 问题 | 修复内容 |
|------|------|----------|
| `pages/cases/create.js` | C1-C7 | 提交数据字段映射：移除 `case_type`, `missing_location`, `missing_latitude/longitude`；将 `missing_detail`→`description`, `clothing`→`clothes`, `special_features`→`features`, `contact_relation`→`contact_rel`；新增 `province/city/district/address` |
| `pages/cases/detail.js` | C8-C10 | 数据加载时从 `province/city/district/address` 构建 `missingLocation`；`markFound` 使用 `location` 和 `note`；`addTrack` 使用 `time` |
| `pages/dialect/create.js` | N1 | tags 从数组转为逗号分隔字符串；添加省份/城市/方言类型解析 |
| `services/missingPerson.js` | 透传层 | 已在调用方修复字段传递 |

### 🟡 中等修复（Medium - 11个）

| 文件 | 问题 | 修复内容 |
|------|------|----------|
| `pages/cases/detail.wxml` | M1-M5 | 模板字段更新：`description`, `clothes`, `features`, `contact_rel`, `time` |
| `pages/map/index.js` | M6 | 移除 `missing_latitude/longitude` 过滤，为所有案件生成标记 |
| `pages/volunteer/workbench.js` | M7 | 统计字段修正：`pending`/`processing`/`completed`（去除 `_count`）|
| `pages/volunteer/profile.js` | M8 | 统计字段修正：`total_tasks`, `total_cases`；移除不存在的 `dialect_count`, `points` |
| `pages/tasks/list.js` | M9 | 添加 `assigned` 状态到 `statusMap` |
| `pages/tasks/detail.js` | M10-M11 | `assignTask` 移除无效导航；`viewLocation` 使用 `lat/lng` |
| `pages/index/index.js` | I1 | 仪表盘字段兼容性处理（嵌套/扁平结构）|

### 🟢 次要修复（Minor/Enhancement - 7个）

| 文件 | 问题 | 修复内容 |
|------|------|----------|
| `utils/request.js` | E23 | loading 默认改为 false，避免与页面级别 loading 冲突 |
| `services/upload.js` | E16 | 下载 URL 从硬编码改为引用 request.js 的 `API_BASE_URL` |
| `utils/request.js` | - | 导出 `API_BASE_URL` 供其他模块使用 |

---

## 核心字段映射表

### MissingPerson 字段映射

| 前端旧字段 | 后端字段 | 状态 |
|-----------|---------|------|
| `missing_location` | `province/city/district/address` | ✅ 已映射 |
| `missing_latitude` | - | ✅ 已移除（后端无此字段）|
| `missing_longitude` | - | ✅ 已移除（后端无此字段）|
| `missing_detail` | `description` | ✅ 已修复 |
| `clothing` | `clothes` | ✅ 已修复 |
| `special_features` | `features` | ✅ 已修复 |
| `contact_relation` | `contact_rel` | ✅ 已修复 |
| `case_type` | - | ✅ 已移除（后端无此字段）|

### Task 统计字段映射

| 前端旧字段 | 后端字段 | 状态 |
|-----------|---------|------|
| `pending_count` | `pending` | ✅ 已修复 |
| `processing_count` | `processing` | ✅ 已修复 |
| `helped_count` | `completed` | ✅ 已修复 |
| - | `assigned` | ✅ 已添加到 statusMap |

### User 统计字段映射

| 前端旧字段 | 后端字段 | 状态 |
|-----------|---------|------|
| `task_count` | `total_tasks` | ✅ 已修复 |
| `case_count` | `total_cases` | ✅ 已修复 |
| `dialect_count` | - | ✅ 已移除（后端无此字段）|
| `points` | - | ✅ 已移除（后端无此字段）|

### Track/轨迹字段映射

| 前端旧字段 | 后端字段 | 状态 |
|-----------|---------|------|
| `track_time` | `time` | ✅ 已修复 |

### MarkFound 字段映射

| 前端旧字段 | 后端字段 | 状态 |
|-----------|---------|------|
| `found_location` | `location` | ✅ 已修复 |
| `found_time` | - | ✅ 已移除（后端从系统时间获取）|
| `description` | `note` | ✅ 已修复 |

---

## 验证建议

1. **案件创建流程**
   - 创建新案件，验证所有字段正确提交
   - 检查 `province/city/district/address` 是否正确解析

2. **案件详情页面**
   - 查看案件详情，验证地址显示正常
   - 测试「标记已找到」功能
   - 测试「添加轨迹」功能

3. **方言创建**
   - 录制方言并添加标签，验证标签以字符串格式提交

4. **任务管理**
   - 查看任务列表和详情
   - 测试任务状态显示

5. **个人中心**
   - 检查统计数据加载

---

## 注意事项

1. **request.js loading 行为变更**
   - 现在默认不显示全局 loading
   - 需要显示 loading 的调用需显式设置 `loading: true`

2. **API_BASE_URL 配置**
   - 统一在 `utils/request.js` 中配置
   - 已导出供 `upload.js` 使用

3. **地址解析**
   - 省份和城市从 `missingLocation` 字符串解析
   - 新增 `parseLocation` 工具函数

---

## 修复日期

2026-03-14（初次修复）；2026-03-20（补充 `urgency_level` 字段对齐，`my_processing` 统计字段）
