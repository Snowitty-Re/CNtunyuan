# 图标资源说明

## 图标系统

本项目使用 **Emoji + 开源图标** 的组合方案，无需外部字体文件！

## 图标方案

### 1. Emoji 图标（推荐）

通过 CSS `::before` 伪元素使用 Unicode Emoji 作为图标，定义在 `mini-program/assets/styles/icons.wxss`。

**使用方式：**
```html
<text class="iconfont icon-add"></text>
<text class="iconfont icon-location"></text>
```

**可用图标：**
| 类名 | Emoji | 说明 |
|------|-------|------|
| icon-add | ➕ | 添加 |
| icon-mic | 🎙️ | 麦克风 |
| icon-location | 📍 | 位置 |
| icon-task | 📋 | 任务 |
| icon-people | 👥 | 人员 |
| icon-success | ✅ | 成功/已找到 |
| icon-volunteer | 🤝 | 志愿者 |
| icon-audio | 🔊 | 音频/方言 |
| icon-arrow-right | → | 箭头 |
| icon-empty | 📭 | 空状态 |
| icon-play | ▶ | 播放 |
| icon-pause | ⏸ | 暂停 |
| icon-like | ❤️ | 点赞 |
| icon-search | 🔍 | 搜索 |
| icon-user | 👤 | 用户 |
| icon-time | ⏰ | 时间 |
| icon-info | ℹ️ | 信息 |
| icon-edit | ✏️ | 编辑 |
| icon-settings | ⚙️ | 设置 |
| icon-notification | 🔔 | 通知 |
| icon-camera | 📷 | 相机 |
| icon-close | ✕ | 关闭 |
| icon-delete | 🗑️ | 删除 |
| icon-phone | 📞 | 电话 |
| icon-email | 📧 | 邮件 |
| icon-home | 🏠 | 首页 |
| icon-menu | ☰ | 菜单 |
| icon-back | ← | 返回 |
| icon-more | ⋮ | 更多 |
| icon-refresh | 🔄 | 刷新 |
| icon-share | 📤 | 分享 |
| icon-upload | ⬆️ | 上传 |
| icon-download | ⬇️ | 下载 |
| icon-calendar | 📅 | 日历 |
| icon-check | ✓ | 勾选 |
| icon-warning | ⚠️ | 警告 |
| icon-error | ❗ | 错误 |
| icon-help | ❓ | 帮助 |
| icon-case | 📋 | 案件 |
| icon-assign | 📤 | 分配 |
| icon-certificate | 📜 | 证书 |

### 2. 直接 Emoji

在 WXML 中直接使用 Emoji：
```html
<text>🔍</text>
<text>📍</text>
<text>📅</text>
```

### 3. 图片图标

以下场景使用 PNG 图片图标：

| 图标 | 文件名 | 用途 |
|------|--------|------|
| 首页 | tab_home.png / tab_home_active.png | TabBar |
| 案件 | tab_case.png / tab_case_active.png | TabBar |
| 工作台 | tab_work.png / tab_work_active.png | TabBar |
| 我的 | tab_profile.png / tab_profile_active.png | TabBar |
| Logo | logo.png | 登录页/启动页 |
| 默认头像 | default-avatar.png | 用户头像 |

## 图标尺寸

```css
.icon-xs { font-size: 20rpx; }
.icon-sm { font-size: 28rpx; }
.icon-md { font-size: 36rpx; }
.icon-lg { font-size: 48rpx; }
.icon-xl { font-size: 64rpx; }
.icon-xxl { font-size: 96rpx; }
```

## 如何添加新图标

在 `mini-program/assets/styles/icons.wxss` 中添加：

```css
.icon-newicon::before,
.iconfont.icon-newicon::before { content: "🎯"; }
```

然后在页面中使用：
```html
<text class="iconfont icon-newicon"></text>
```

## Emoji 资源

查找更多 Emoji：
- [Emoji 大全](https://emojipedia.org/)
- [常用 Emoji](https://www.emojiall.com/)

## 优势

1. **无需网络** - Emoji 是系统自带
2. **跨平台** - iOS、Android、小程序都支持
3. **免配置** - 无需下载字体文件
4. **即时生效** - 添加新图标不需要重新编译
5. **体积小** - 纯 CSS 方案，不增加包大小

## 注意事项

- 不同平台 Emoji 显示效果略有差异
- 建议使用 Unicode 6.0 以上的常用 Emoji
- 避免使用过于生僻或新发布的 Emoji
