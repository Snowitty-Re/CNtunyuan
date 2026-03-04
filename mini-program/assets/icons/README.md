# 图标资源说明

## ⚠️ 重要提示

**当前目录中的图标都是占位符（Placeholder）**，你需要替换为真实的图标文件才能正常显示！

## TabBar 图标

需要准备以下图标（PNG格式，建议尺寸 81x81px）：

### 首页
- `tab_home.png` - 首页默认图标 (灰色)
- `tab_home_active.png` - 首页选中图标 (橙色 #FF8C42)

### 案件
- `tab_case.png` - 案件默认图标 (灰色)
- `tab_case_active.png` - 案件选中图标 (橙色 #FF8C42)

### 工作台
- `tab_work.png` - 工作台默认图标 (灰色)
- `tab_work_active.png` - 工作台选中图标 (橙色 #FF8C42)

### 我的
- `tab_profile.png` - 我的默认图标 (灰色)
- `tab_profile_active.png` - 我的选中图标 (橙色 #FF8C42)

## 其他图标

### Logo
- `logo.png` - 应用Logo（建议尺寸 200x200px）

### 地图标记
- `marker_red.png` - 红色标记（失踪中）
- `marker_orange.png` - 橙色标记（寻找中）
- `marker_green.png` - 绿色标记（已找到）
- `marker_blue.png` - 蓝色标记（已团圆）

### 默认头像
- `default-avatar.png` - 默认用户头像（建议尺寸 200x200px）

## 快速获取图标

### 方法一：阿里巴巴矢量图标库 (推荐)
1. 访问 https://www.iconfont.cn
2. 搜索关键词：home、file、work、user、map marker
3. 选择风格统一的图标
4. 下载 PNG 格式

### 方法二：使用设计工具自制
- 推荐工具：Figma、Sketch、Photoshop
- 画布尺寸：81x81px (TabBar) 或 40x40px (地图标记)
- 导出格式：PNG，保留透明背景

### 方法三：使用开源图标库
- [Heroicons](https://heroicons.com/)
- [Feather Icons](https://feathericons.com/)
- [Tabler Icons](https://tabler-icons.io/)

## 图标设计规范

### 颜色规范
- **默认状态**：灰色系 (#999999)
- **选中状态**：主题橙色 (#FF8C42)
- **背景**：透明

### 尺寸规范
| 用途 | 建议尺寸 | 文件大小 |
|------|----------|----------|
| TabBar 图标 | 81x81px | < 20KB |
| 地图标记 | 40x40px | < 10KB |
| Logo | 200x200px | < 50KB |
| 默认头像 | 200x200px | < 30KB |

### 格式要求
- PNG 格式
- 支持透明背景 (Alpha通道)
- 文件大小尽量控制在 20KB 以内以保证加载速度

## 替换步骤

1. 准备好真实图标文件
2. 重命名为对应的文件名
3. 替换到 `mini-program/assets/icons/` 目录
4. 在微信开发者工具中预览效果

## 图标字体

项目中使用了一些图标字体编码（如 `&#xe6a0;`），这些是占位符。

如需使用图标字体，建议：
1. 在 iconfont.cn 创建项目
2. 下载字体文件
3. 在 app.wxss 中引入字体
4. 替换所有 `&#xe6xx;` 编码为实际图标类名

或者，直接将所有图标字体替换为图片图标。
