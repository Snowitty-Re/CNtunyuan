# 团圆寻亲 - Web 平台

基于 React 18 + TypeScript + Vite + Ant Design 5 构建的简洁办公风格 Web 平台。

## 设计理念

- **简洁干净** - 去除多余装饰，信息层级清晰
- **办公OA风格** - 类似飞书、钉钉的专业内部系统体验
- **温馨感** - 采用柔和暖色调，体现团圆寻亲的人文关怀

## 视觉规范

### 色彩系统

| 用途 | 色值 | 说明 |
|------|------|------|
| 主色 | `#e67e22` | 温暖的橙色，用于按钮、高亮 |
| 主色悬停 | `#d35400` | 深一点的橙色 |
| 背景色 | `#f5f7fa` | 浅灰白背景 |
| 主文字 | `#1f2329` | 深灰色标题文字 |
| 次要文字 | `#646a73` | 灰色说明文字 |
| 边框色 | `#e8e9eb` | 细边框分隔 |

### 字体规范

- 主字体：系统默认字体栈（-apple-system, PingFang SC, Microsoft YaHei）
- 基础字号：14px
- 标题：16-20px，font-weight: 600
- 辅助文字：12-13px

### 圆角规范

- 按钮：6px
- 卡片：8px
- 输入框：6px
- 标签：4px
- 头像：8px

## 技术栈

- **框架**: React 18
- **语言**: TypeScript 5
- **构建工具**: Vite 5
- **UI 组件**: Ant Design 5
- **状态管理**: Zustand
- **样式**: 内联样式 + Ant Design 主题覆盖

## 项目结构

```
src/
├── components/
│   └── layout/
│       ├── MainLayout.tsx    # 主布局
│       └── Sidebar.tsx       # 侧边栏
├── pages/
│   ├── login/                # 登录页
│   ├── dashboard/            # 工作台
│   ├── cases/                # 寻人案件
│   ├── tasks/                # 任务管理
│   ├── volunteers/           # 志愿者管理
│   ├── organizations/        # 组织架构
│   └── dialects/             # 方言管理
├── router/
├── stores/
├── services/
├── utils/
├── types/
└── index.css                 # 全局样式覆盖
```

## 开发规范

### 样式规范
- 不使用 Tailwind className
- 使用 Ant Design 组件默认样式
- 特殊样式使用内联 `style={{}}`
- 全局主题覆盖在 `index.css` 中

### 颜色使用
```tsx
// 主色按钮
<Button type="primary" style={{ backgroundColor: '#e67e22', borderColor: '#e67e22' }}>

// 次要文字
<span style={{ color: '#646a73' }}>

// 背景色
<div style={{ backgroundColor: '#f5f7fa' }}>
```

## 快速开始

```bash
# 安装依赖
pnpm install

# 开发模式
pnpm dev

# 构建生产版本
pnpm build

# 预览生产版本
pnpm preview
```

## 默认登录信息

- **手机号**: 13800138000
- **密码**: admin123

## 页面说明

### 工作台
- 欢迎区域：渐变色背景，快速操作按钮
- 统计卡片：4 个核心数据，简洁图标
- 最近任务/案件：列表展示，悬停效果

### 寻人案件
- 表格展示：头像 + 基本信息
- 状态标签：不同颜色区分
- 操作下拉：查看、编辑、删除

### 任务管理
- 进度条：橙色主色调
- 优先级标签：紧急/高/普通/低
- 状态流转：草稿→待分配→进行中→已完成

### 志愿者管理
- 头像展示：默认橙色背景
- 角色标签：超级管理员/管理员/组织者/志愿者
- 状态标识：正常/禁用

### 组织架构
- 左侧树形：层级展示
- 右侧详情：选中组织信息
- 统计卡片：志愿者/案件/任务数

### 方言管理
- 音频列表：播放、时长、地区
- 采集信息：采集人、播放次数
- 地区标签：紫色系区分

## 许可证

MIT
