# 团圆寻亲 - 微信小程序

## 项目简介

团圆寻亲志愿者系统微信小程序，是一个帮助寻找走失人员的公益平台。通过整合志愿者网络、方言语音数据库和任务系统，提高寻人效率。

## 功能模块

### 1. 首页
- 统计数据展示（走失人员、已找到、志愿者、方言录音）
- 快捷入口（发布案件、录制方言、查看地图、我的任务）
- 最新案件和方言展示

### 2. 案件管理
- 走失人员列表（支持搜索、筛选、分页）
- 案件详情（信息展示、轨迹记录、地图位置）
- 创建案件（表单填写、照片上传、地图选址）
- 状态管理（失踪中、寻找中、已找到、已团圆）

### 3. 方言录音
- 方言列表（地区筛选、播放录音）
- 方言详情（播放控制、点赞、评论）
- 录制方言（15-20秒录音、地区选择、标签添加）

### 4. 任务系统
- 任务列表（状态筛选、分页加载）
- 我的任务（按状态分组、快速操作）
- 任务详情（信息展示、操作日志）
- 任务创建（管理者权限、案件关联）
- 任务反馈（文字反馈、图片上传）

### 5. 工作台
- 今日统计（待处理、进行中、已帮助）
- 快捷入口（发布案件、录制方言、待分配任务）
- 最近任务列表

### 6. 个人中心
- 用户信息展示（头像、昵称、角色、积分）
- 数据统计（我的任务、我的案件、方言录音）
- 功能菜单（编辑资料、消息通知、志愿者证书、设置）

### 7. 地图功能
- 地图标记（走失人员位置分布）
- 筛选查看（按状态筛选）
- 导航功能（调起地图导航）

## 技术架构

### 核心技术
- 微信小程序原生开发
- ES6+ 语法
- Promise 异步处理
- 组件化开发

### 项目结构
```
mini-program/
├── app.js              # 应用入口
├── app.json            # 全局配置
├── app.wxss            # 全局样式
├── services/           # API 服务层
│   ├── index.js        # 服务导出
│   ├── auth.js         # 认证服务
│   ├── user.js         # 用户服务
│   ├── missingPerson.js # 走失人员服务
│   ├── dialect.js      # 方言服务
│   ├── task.js         # 任务服务
│   ├── upload.js       # 上传服务
│   ├── dashboard.js    # 仪表盘服务
│   └── organization.js # 组织服务
├── utils/              # 工具函数
│   ├── request.js      # 请求封装
│   └── util.js         # 通用工具
├── pages/              # 页面文件
│   ├── index/          # 首页
│   ├── login/          # 登录
│   ├── cases/          # 案件管理
│   ├── dialect/        # 方言录音
│   ├── tasks/          # 任务系统
│   ├── volunteer/      # 工作台和个人中心
│   ├── notification/   # 消息通知
│   ├── settings/       # 设置
│   └── map/            # 地图
├── components/         # 公共组件
└── assets/             # 静态资源
    └── icons/          # 图标资源
```

## API 接口

### 认证相关
- `POST /auth/wechat-login` - 微信登录
- `POST /auth/login` - 账号密码登录
- `POST /auth/refresh` - 刷新 Token
- `POST /auth/logout` - 退出登录
- `GET /auth/me` - 获取当前用户

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
- `POST /missing-persons/:id/tracks` - 添加轨迹

### 方言录音
- `GET /dialects` - 列表
- `POST /dialects` - 创建
- `GET /dialects/:id` - 详情
- `POST /dialects/:id/play` - 播放记录
- `POST /dialects/:id/like` - 点赞

### 任务系统
- `GET /tasks` - 任务列表
- `GET /tasks/my` - 我的任务
- `POST /tasks` - 创建任务
- `GET /tasks/:id` - 详情
- `POST /tasks/:id/start` - 开始任务
- `POST /tasks/:id/complete` - 完成任务

### 文件上传
- `POST /upload` - 单文件上传
- `POST /upload/batch` - 批量上传

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
- 统一错误处理

### 样式规范
- 使用 rpx 作为单位
- 主题色：`#FF8C42`
- 遵循设计稿规范

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
- `scope.userLocation` - 位置信息
- `scope.record` - 录音功能
- `scope.camera` - 相机功能
- `scope.writePhotosAlbum` - 保存图片

### 安全规范
- 敏感操作需要二次确认
- Token 过期自动刷新
- 关键数据加密存储

## 更新日志

### v1.0.0 (2024-03)
- 初始版本发布
- 完整功能模块实现
- 对接后端 API
- 移除微信云开发

## 联系方式

- 项目地址：https://github.com/Snowitty-Re/CNtunyuan
- 问题反馈：issues

## 开源协议

MIT License
