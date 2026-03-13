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
3. 修改 `app.js` 中的 `API_CONFIG` 配置
4. 开启"不校验合法域名"进行开发

### 生产环境
1. 配置服务器域名（request、upload、download）
2. 配置业务域名（webview）
3. 关闭开发调试选项
4. 上传代码并提交审核

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
