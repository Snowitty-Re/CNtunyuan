# 团圆寻亲 - 微信小程序

## 项目简介

团圆寻亲志愿者系统微信小程序，是一个帮助寻找走失人员的公益平台。通过整合志愿者网络、方言语音数据库和任务系统，提高寻人效率。

## 功能模块

### 1. 首页
- 渐变问候区（用户头像 + 动态问候语）
- 统计数据展示（走失人员、已找到、志愿者、方言录音）
- 快捷入口（发布案件、录制方言、查看地图、我的任务）
- 最新案件横向滑动卡片
- 精选方言列表（带播放按钮）

### 2. 案件管理
- 走失人员列表（圆角搜索框、状态标签横向滚动、左图右信息卡片布局）
- 案件详情（顶部大图 + 状态胶囊、信息分组卡片、照片画廊、轨迹时间线、底部操作栏）
- 创建/编辑案件（支持 `?id=xxx` 编辑模式，表单填写、照片上传、地图选址）
- 状态管理（失踪中、寻找中、已找到、已团圆、已结案）

### 3. 方言录音
- 方言列表（地区筛选、音频卡片带波形装饰和播放按钮）
- 方言详情（大播放器 + 进度条、评论区、点赞/分享）
- 录制方言（圆形录音按钮 + 波形动画、试听、重录、表单填写、上传）

### 4. 任务系统
- 任务列表（统计卡片、状态筛选、优先级色带标记）
- 我的任务（标签页分组、进度条、快速操作）
- 任务详情（状态流程条、信息卡片、操作日志时间线）
- 任务创建（管理者权限、案件关联、指派志愿者）
- 任务反馈（文字反馈、图片上传）

### 5. 工作台
- 顶部日期 + 用户信息
- 今日统计卡片（待处理、进行中、已帮助）
- 快捷入口（4宫格圆形图标按钮）
- 最近任务列表卡片

### 6. 个人中心
- 渐变头部（大头像 + 角色徽章 + 积分）
- 统计横条（任务/案件/方言/积分）
- 功能菜单（emoji图标 + 标题 + 箭头 + 徽章分组）
- 用户ID复制、退出登录

### 7. 地图功能
- 全屏地图 + 毛玻璃搜索栏
- 状态筛选浮动标签
- 地图标记（彩色圆点区分状态）
- 底部滑动面板（案件列表 + 导航按钮）

### 8. 消息通知
- 本地消息中心（基于任务/案件变更动态生成）
- 已读/未读状态管理（localStorage）
- 全部标记已读、清空

### 9. 设置
- 分组菜单（emoji图标）
- 清除缓存（保留登录信息）
- 帮助中心（FAQ可展开）
- 关于、用户协议、隐私政策

## 技术架构

### 核心技术
- 微信小程序原生开发
- ES6+ 语法 / async/await
- CSS变量主题系统
- Emoji图标系统（无外部字体依赖）
- 组件化开发（empty / loading / status-tag）

### 设计系统
- 主题色：`#FF8C42`（暖橙色）
- 背景色：`#FDF8F3`（奶油色）
- 全局CSS变量：`--primary`, `--success`, `--danger`, `--warning`, `--bg`, `--card-radius`, `--spacing`
- 统一卡片样式：圆角 `16-24rpx`、暖色阴影
- 渐变头部装饰（decorative orbs）
- hover反馈、安全区域适配

### 项目结构
```
mini-program/
├── config/
│   └── index.js        # 环境配置（dev/prod 切换，统一 API_BASE）
├── app.js              # 应用入口（登录态管理、全局方法）
├── app.json            # 全局配置（23个页面、tabBar）
├── app.wxss            # 全局样式（CSS变量、基础样式）
├── services/           # API 服务层（8个模块）
│   ├── index.js        # 服务导出
│   ├── auth.js         # 认证服务
│   ├── user.js         # 用户服务
│   ├── missingPerson.js # 走失人员服务
│   ├── dialect.js      # 方言服务
│   ├── task.js         # 任务服务
│   ├── upload.js       # 上传/下载服务
│   ├── dashboard.js    # 仪表盘服务
│   └── organization.js # 组织服务
├── utils/              # 工具函数
│   ├── request.js      # 请求封装（Token刷新、错误处理）
│   └── util.js         # 通用工具（格式化、验证）
├── pages/              # 页面文件（23个页面）
│   ├── index/          # 首页
│   ├── login/          # 登录（微信登录 + 手机号登录）
│   ├── cases/          # 案件管理（list/detail/create）
│   ├── dialect/        # 方言录音（list/detail/create）
│   ├── tasks/          # 任务系统（list/my/detail/create/feedback）
│   ├── volunteer/      # 个人中心/工作台/编辑资料
│   ├── notification/   # 消息通知
│   ├── settings/       # 设置（index/about/help/agreement/privacy）
│   └── map/            # 地图
├── components/         # 公共组件
│   ├── empty/          # 空状态组件
│   ├── loading/        # 加载组件（含骨架屏）
│   └── status-tag/     # 状态标签组件
└── assets/             # 静态资源
    ├── icons/          # 图标资源
    ├── images/         # 图片资源（头像、标记、分享图）
    └── styles/         # 全局样式（emoji图标映射）
```

## API 接口

### 认证相关
- `POST /auth/wechat-login` - 微信登录
- `POST /auth/login` - 账号密码登录
- `POST /auth/refresh` - 刷新 Token
- `POST /auth/logout` - 退出登录
- `GET /auth/me` - 获取当前用户
- `POST /auth/bind-phone` - 绑定手机号
- `POST /auth/send-code` - 发送验证码

### 用户相关
- `GET /users` - 用户列表
- `GET /profile` - 个人资料
- `PUT /profile` - 更新资料
- `PUT /profile/password` - 修改密码

### 走失人员
- `GET /missing-persons` - 列表
- `POST /missing-persons` - 创建
- `GET /missing-persons/:id` - 详情
- `PUT /missing-persons/:id` - 更新
- `PUT /missing-persons/:id/status` - 更新状态
- `POST /missing-persons/:id/tracks` - 添加轨迹

### 方言录音
- `GET /dialects` - 列表
- `POST /dialects` - 创建
- `GET /dialects/:id` - 详情
- `POST /dialects/:id/play` - 播放记录
- `POST /dialects/:id/like` - 点赞
- `POST /dialects/:id/comments` - 评论

### 任务系统
- `GET /tasks` - 任务列表
- `GET /tasks/my` - 我的任务
- `POST /tasks` - 创建任务
- `GET /tasks/:id` - 详情
- `POST /tasks/:id/assign` - 分配任务
- `POST /tasks/:id/start` - 开始任务
- `POST /tasks/:id/complete` - 完成任务
- `POST /tasks/:id/cancel` - 取消任务

### 文件上传
- `POST /upload` - 单文件上传
- `GET /upload/:id` - 获取文件信息
- `GET /upload/:id/download` - 下载文件
- `DELETE /upload/:id` - 删除文件

### 组织管理
- `GET /organizations` - 组织列表
- `GET /organizations/:id` - 组织详情
- `GET /organizations/tree` - 组织树
- `GET /organizations/:id/children` - 子组织
- `GET /organizations/:id/path` - 组织路径

### 仪表盘
- `GET /dashboard/stats` - 统计数据
- `GET /dashboard/overview` - 概览数据
- `GET /dashboard/trend` - 趋势数据

## 开发规范

### 命名规范
- 页面文件：小写，单词间用 `-` 连接
- 组件文件：小写，单词间用 `-` 连接
- JS 变量：驼峰命名
- CSS 类名：小写，单词间用 `-` 连接

### 代码规范
- 使用 ES6+ 语法
- Promise 处理异步
- async/await 优先
- 统一错误处理（try/catch + showError）

### 样式规范
- 使用 rpx 作为单位
- 主题色：`#FF8C42`（暖橙色）
- 背景色：`#FDF8F3`（奶油色）
- 使用全局CSS变量（`var(--primary)` 等）
- 卡片圆角：`16-24rpx`
- 间距：`24rpx` 基准

### 图标规范
- 使用 Emoji 图标系统（`assets/styles/icons.wxss`）
- 用法：`<text class="iconfont icon-xxx"></text>`
- 支持尺寸类：`icon-xs` / `icon-sm` / `icon-md` / `icon-lg` / `icon-xl`

## 开发环境

### 环境要求
- 微信开发者工具 1.06.2307260+
- 基础库版本 2.32.0+
- Node.js 18+

### 开发配置
1. 克隆项目
2. 使用微信开发者工具打开 `mini-program` 目录
3. 确认 `config/index.js` 中 `env = 'dev'`（默认开发环境）
4. 开启"不校验合法域名"进行开发

### 生产环境
1. 将 `config/index.js` 中的 `env` 改为 `'prod'`，所有请求（含文件下载）自动指向生产域名
2. 配置服务器域名（request、upload、download）
3. 配置业务域名（webview）
4. 关闭开发调试选项
5. 上传代码并提交审核

## 注意事项

### 权限申请
- `scope.userLocation` - 位置信息（地图功能）
- `scope.record` - 录音功能（方言录制）
- `scope.camera` - 相机功能（照片拍摄）
- `scope.writePhotosAlbum` - 保存图片

### 安全规范
- 敏感操作需要二次确认
- Token 过期自动刷新（refresh_token 机制）
- 清除缓存时保留登录信息
- 退出登录清除所有本地数据

### 后端字段对照
- 走失位置字段：`missing_latitude` / `missing_longitude`（非 `last_seen_*`）
- 位置文本：拼接 `province` + `city` + `district` + `address`
- 任务状态：`draft` / `pending` / `assigned` / `processing` / `completed` / `cancelled`
- 案件状态：`missing` / `searching` / `found` / `reunited` / `closed`

## 更新日志

### v1.6.0 (2026-03)
**第五轮修复 — 模板渲染、登录交互、数据语义、上传容错**

- **录音时长提示修复**：`dialect/create.js` 将模块级常量 `MIN_DURATION`/`MAX_DURATION` 写入 `data.minDuration`/`data.maxDuration`；`create.wxml` 引用 `{{minDuration}}`/`{{maxDuration}}`，修复录音时长提示显示为空的问题
- **录音波形动画索引**：`dialect/create.wxml` 内层波形 `wx:for` 添加 `wx:for-index="barIdx"` 别名，与外层 `index` 隔离
- **登录链接修复**：`login/index.js` 添加 `goToRegister`/`goToForgot` 方法，点击不再触发 JS 错误
- **markFound 语义修复**：`cases/detail.js` 不再将走失地点作为找到地点传给后端
- **onShow 错误捕获**：`cases/detail.js` onShow 中 `Promise.all` 补充 `.catch()`
- **上传结果容错**：`tasks/feedback.js` 图片 URL 提取加 `r.data?.url` 回退并过滤空值
- **缓存显示优化**：`settings/index.js` 小于 1MB 显示 KB，大于等于 1MB 显示 MB

### v1.5.0 (2026-03)
**第四轮修复 — 空指针崩溃、字段类型、UI 文案、性能**

- **P0 地图崩溃修复**：`map/index.wxml` 标记弹窗中 `item.data.photos[0].url` 在照片列表为空时崩溃，改为带守卫的三目表达式
- **P0 方言标签修复**：`dialect/list.js` 和 `dialect/detail.js` — 后端返回逗号分隔字符串，WXML 的 `wx:for` 会逐字迭代；JS 层统一转数组（`split(',').filter(Boolean)`）
- **P1 轨迹时间回退**：`cases/detail.js` `loadTracks` 中 `displayTime: formatDate(t.time)` — 若 `time` 字段为空轨迹时间显示为空；改为 `t.time || t.created_at` 回退
- **P1 空状态文案修复**：`tasks/my.wxml` 条件字符串拼接导致"您还没有的任务"（全部标签时中间词为空）；改为 `wx:if/wx:elif` 独立文本节点
- **P2 波形动画索引冲突**：`dialect/list.wxml` 内层 `wx:for="{{12}}"` 使用外层 `{{index}}` 变量导致动画延迟全部相同；添加 `wx:for-index="barIndex"` 别名
- **P2 org 类型守卫**：`volunteer/profile.js` `userInfo.org?.name` 在 `org` 为字符串时行为未定义；改为先检查类型再访问 `.name`
- **P3 onTimeUpdate 节流**：`dialect/detail.js` 音频进度回调每秒触发多次 `setData`；添加 100ms 节流减少渲染压力
- **P3 删除 console.warn**：`index/index.js` `safeString()` 中残留的调试日志已移除

### v1.4.0 (2026-03)
**地图功能完整实现、性能优化、技术债清偿**

- **地图完整实现**：`map/index.js` 新增 `latitude/longitude/scale/markers` data 字段；实现 `locateCurrentPosition()`（`wx.getLocation` 定位）、`onMarkerTap()`（标记点选中）、`onRegionChange()`（区域变更）、`navigateToLocation()`（`wx.openLocation` 导航）；案件数据按 `missing_latitude/missing_longitude` 字段转换为地图标记数组
- **onShow 节流**：`tasks/list.js` 和 `index/index.js` 添加 30s 节流，避免每次返回页面触发全量刷新
- **并发加载**：`cases/detail.js` onShow 改为 `Promise.all([loadCaseDetail(), loadTracks()])`，减少串行等待
- **拨打电话确认**：`cases/detail.js`、`map/index.js` makePhoneCall 前弹出确认对话框
- **统一常量**：新建 `utils/constants.js`，提取 `TASK_STATUS_MAP`/`TASK_PRIORITY_MAP`/`TASK_PRIORITY_COLOR`/`CASE_STATUS_MAP`/`GENDER_MAP`/`ROLE_MAP` 等；`tasks/list`、`tasks/my`、`tasks/detail`、`volunteer/workbench` 改为引用共享常量
- **统一地址拼接**：`utils/util.js` 新增 `joinLocation(item)` 帮助函数；`cases/list`、`cases/detail`、`map/index` 三处重复的 `.filter(Boolean).join(' ')` 改为调用此函数
- **deadline 过去日期校验**：`tasks/create.js` validateForm 添加截止日期不能早于今天的校验
- **is_liked 布尔修复**：`dialect/detail.js` `is_liked || false` → `!!dialect.is_liked`，避免非布尔值判断错误
- **移除占位通知**：`notification/list.js` 删除硬编码的 `system_welcome` 系统欢迎通知
- **missingTime try-catch**：`cases/create.js` `new Date(form.missingTime)` 包裹 try-catch

### v1.3.0 (2026-03)
**全面代码审查 — P0 崩溃、WXML 渲染错误、内存泄漏修复**

- **P0 崩溃**：`tasks/detail.js` 补充 `showLoading`/`hideLoading` 导入（点击"分配任务"触发 ReferenceError）
- **P0 数据不显示**：`cases/list.wxml` 修正字段名 `missing_location` → `missingLocation`（地点列永远空白）
- **WXML 渲染错误**：`dialect/list`、`dialect/detail`、`dialect/create` 的 WXML 模板中直接调用 Page 方法（`{{formatTime()}}`、`{{formatPlayCount()}}`、`{{formatTimeAgo()}}`）— 小程序不支持此语法，时长/播放数/时间进度全部不渲染。改为在 JS 的 `setData` 处预计算文本字段写入 data
- **跳转路径错误**：`dialect/detail.js` 关联案件跳转 `/pages/missing/detail` 不存在，修正为 `/pages/cases/detail`
- **状态显示空白**：`tasks/my.js` statusMap 缺少 `assigned`/`draft` 状态，对应任务标签空白
- **音频实例泄漏**：`dialect/detail.js` `initAudioContext()` 无守卫，每次进入页面创建新实例不释放旧实例
- **定时器泄漏**：`login/index.js` 倒计时 `setInterval` 未在 `onUnload` 清理，页面销毁后持续运行
- **验证码功能失效**：`edit-profile.js` 调用不存在的 `userService.sendVerifyCode?()`（可选链静默失败）→ 改为 `authService.sendVerifyCode()`
- **搜索崩溃**：`map/index.js` 搜索时 `item.name.includes()` 在 name 为 null 时抛出异常，添加 null 守卫

### v1.2.0 (2026-03)
- 新增 `config/index.js` 统一环境配置，切换 dev/prod 只需改一行
- 修复 Token 刷新队列缺陷：并发 401 时入队前浅拷贝 options，防止重试时 header 污染
- 删除登录页测试绕过逻辑（`quickBindPhone` / `doBindPhoneWithoutCode`），绑定手机号强制走短信验证
- `user.getStats()` 返回值规范化，消除 profile/workbench 页面的重复 fallback 映射
- 统一权限检查：`tasks/list`、`tasks/detail` 改用 `app.isManager()`，与 `app.js` 保持单一来源
- 删除各列表页 `result.list || result` 响应兼容层，统一依赖后端标准结构

### v1.1.0 (2026-03)
- 修复11个P0级接口错误（字段名、不存在的API端点、页面跳转）
- 补充缺失图片资源（9个PNG文件）
- 统一状态枚举（添加 closed/draft 状态）
- 案件创建支持编辑模式（`?id=xxx` 参数）
- 通知重写为本地消息中心
- 上传服务添加下载方法，修复批量上传
- 组织服务修正（删除3个无效端点，添加3个有效端点）
- 全部23个页面UI重新设计（暖橙色主题 + 奶油色背景）
- 新增 status-tag 通用组件
- loading 组件支持骨架屏模式
- 添加全局CSS变量主题系统
- Emoji图标系统扩展（新增 arrow-down/heart/copy/list/navigation）

### v1.0.0 (2024-03)
- 初始版本发布
- 完整功能模块实现
- 对接后端 API

## 联系方式

- 项目地址：https://github.com/Snowitty-Re/CNtunyuan
- 问题反馈：issues

## 开源协议

MIT License
