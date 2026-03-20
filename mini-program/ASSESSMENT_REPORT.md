# 团圆寻亲小程序 — 技术评估报告

**评估日期**：2026-03-20 | **基准版本**：v1.3.0 | **评估人**：Claude Code 自动化审查

---

## 一、项目概览

| 指标 | 数据 |
|------|------|
| 页面数量 | 23 个页面 |
| 服务模块 | 9 个 service（auth / user / missingPerson / dialect / task / upload / dashboard / organization） |
| JS 代码量 | ~6,500 行 |
| WXML 代码量 | ~3,200 行 |
| **整体评分** | **6.2 / 10** |
| **可上线状态** | ❌ 不建议上线（修复后预计 8.0/10） |

---

## 二、功能完整性（7 / 10）

### ✅ 已完整实现

- 微信登录 + 手机号密码登录 + Token 自动刷新
- 案件（走失人员）增删改查、状态流转、轨迹记录
- 方言录音上传、播放、点赞、评论、地区筛选
- 任务全生命周期（创建→分配→开始→更新进度→完成→取消）
- 个人资料编辑、密码修改、统计看板

### ❌ 未完整实现

| 功能 | 文件 | 问题 |
|------|------|------|
| **地图标记** | `pages/map/index.js` | WXML 引用 latitude/longitude/scale/markers/locateCurrentPosition 等字段和方法，JS 中均未实现 |
| **消息通知** | `pages/notification/list.js` | 无专属通知 API，从任务/案件聚合模拟，残留"系统欢迎"占位消息 |
| **志愿者证书** | `pages/volunteer/profile.js` | 点击仅弹 toast"功能开发中" |
| **通知设置开关** | `pages/settings/index.js:56` | switch 控件无业务逻辑 |

---

## 三、代码 Bug 清单

### 🔴 P0 — 崩溃 / 数据不显示

| # | 文件 | 行号 | 问题 | 修复状态 |
|---|------|------|------|---------|
| 1 | `tasks/detail.js` | 2 | `showLoading` 未导入，assignTask 崩溃 | ✅ v1.3.0 |
| 2 | `cases/list.wxml` | 71 | 字段名 `missing_location` vs JS 的 `missingLocation` | ✅ v1.3.0 |
| 3 | `map/index.js` | — | `latitude/longitude/scale/markers` 未在 data 中定义，地图无法显示标记点；`locateCurrentPosition`/`onMarkerTap`/`onRegionChange`/`navigateToLocation` 方法缺失 | ✅ v1.4.0 |

### 🟡 P1 — 功能异常

| # | 文件 | 行号 | 问题 | 修复状态 |
|---|------|------|------|---------|
| 4 | `dialect/list.wxml` | 89,109,113,124 | WXML 直接调用 Page 方法（不支持） | ✅ v1.3.0 |
| 5 | `dialect/detail.wxml` | 62,63,97,102 | 同上 | ✅ v1.3.0 |
| 6 | `dialect/create.wxml` | 30,34,54,55 | 同上 | ✅ v1.3.0 |
| 7 | `dialect/detail.js` | 84 | `initAudioContext` 无守卫，多次进入页面泄漏音频实例 | ✅ v1.3.0 |
| 8 | `dialect/detail.js` | 300 | 跳转路径 `/pages/missing/detail` 不存在 | ✅ v1.3.0 |
| 9 | `tasks/my.js` | 14 | statusMap 缺少 `assigned`/`draft`，显示空白 | ✅ v1.3.0 |
| 10 | `edit-profile.js` | 145 | `userService.sendVerifyCode?.()` 不存在，可选链静默失败 | ✅ v1.3.0 |
| 11 | `cases/detail.js` | 52 | `onShow` 中两个 async 串行，应 Promise.all | ✅ v1.4.0 |
| 12 | `cases/create.js` | 427 | setTimeout 内 prevPage.loadCases() 无 await | ✅ v1.4.0 |
| 13 | `tasks/create.js` | 148 | deadline 无过去日期校验 | ✅ v1.4.0 |
| 14 | `cases/create.js` | 397 | `new Date(form.missingTime)` 无 try-catch | ✅ v1.4.0 |
| 15 | `dialect/detail.js` | 149 | `is_liked \|\| false` 布尔判断错误 | ✅ v1.4.0 |

### 🟠 P2 — 体验 / 性能

| # | 文件 | 问题 | 修复状态 |
|---|------|------|---------|
| 16 | `tasks/list.js` `index/index.js` | onShow 全量刷新无节流 | ✅ v1.4.0 |
| 17 | `cases/detail.js` | 手机拨打无确认弹窗 | ✅ v1.4.0 |
| 18 | 多处 | 状态映射各文件重复定义 | ✅ v1.4.0 |
| 19 | 多处 | 地址拼接逻辑重复 3 处 | ✅ v1.4.0 |
| 20 | `map/index.js` | makePhoneCall 无确认 | ✅ v1.4.0 |

---

## 四、代码质量（6 / 10）

### 重复代码问题

- **状态映射**：`TASK_STATUS_MAP`/`TASK_PRIORITY_MAP` 在 `tasks/list.js`、`tasks/my.js`、`tasks/detail.js`、`volunteer/workbench.js` 中逐字重复（4份）。已提取至 `utils/constants.js`
- **地址拼接**：`[province, city, district, address].filter(Boolean).join(' ')` 重复 3 处。已提取至 `util.js` 的 `joinLocation()`

### API 利用率

```
已定义 service 方法：52 个
实际被调用：28 个（54%）
从未调用：24 个（46%）
```

**完全未使用的模块**：`organization.js`（8 个方法），表明组织管理功能在小程序端未实现

---

## 五、安全性（6.5 / 10）

| 风险项 | 位置 | 说明 |
|--------|------|------|
| Token 明文存储 | `utils/request.js:24` | `wx.getStorageSync` 无加密，越狱设备可读 |
| 手机拨打无确认 | `cases/detail.js`、`map/index.js` | 直接调用 `wx.makePhoneCall` |
| region 参数无白名单 | `dialect/list.js:106` | 任意字符串传给后端，依赖后端验证 |
| 分享含敏感信息 | `cases/detail.js:326` | 分享卡片包含走失者姓名、年龄 |

---

## 六、性能（6.5 / 10）

| 问题 | 位置 | 影响 |
|------|------|------|
| `onShow` 全量刷新 | `tasks/list`、`index/index` | 每次返回页面触发 2–6 个请求 |
| 备用请求风暴 | `index/index.js:147` | 统计接口失败后最多触发 4 个备用请求 |
| 音频 `onTimeUpdate` 高频 setData | `dialect/detail.js:92` | 每秒多次触发页面更新 |
| 搜索无请求取消 | `cases/list.js:137` | 快速搜索结果可能乱序 |

---

## 七、综合评分

| 维度 | 评分 |
|------|------|
| 功能完整性 | 7 / 10 |
| 代码质量 | 6 / 10 |
| 安全性 | 6.5 / 10 |
| 性能 | 6.5 / 10 |
| 错误处理 | 5.5 / 10 |
| API 利用率 | 5 / 10 |
| **综合** | **6.2 / 10** |

修复所有标注问题后预计：**8.0 / 10**

---

## 八、修复进度跟踪

| 版本 | 修复内容 | 状态 |
|------|---------|------|
| v1.1.0 | 11 个接口错误、UI 重构、组件补全 | ✅ 完成 |
| v1.2.0 | 统一配置、Token 队列、权限检查、响应规范化 | ✅ 完成 |
| v1.3.0 | P0 崩溃、WXML 方法调用、内存泄漏、测试代码 | ✅ 完成 |
| v1.4.0 | 地图完整实现（标记/定位/导航/拨号确认）、onShow 节流（tasks/index）、utils/constants.js 统一常量、joinLocation 统一地址拼接、Promise.all 并发加载、deadline 过去日期校验、is_liked 布尔修复、系统欢迎占位清除 | ✅ 完成 |

---

## 九、待办事项（不影响上线的迭代项）

- [ ] 对接后端专用通知 API（WebSocket 或轮询）
- [ ] 志愿者证书功能实现
- [ ] 通知设置开关实现
- [ ] 字段命名蛇形/驼峰统一（需后端配合）
- [ ] 图片上传前压缩
- [ ] 离线缓存机制
- [ ] Token 加密存储
