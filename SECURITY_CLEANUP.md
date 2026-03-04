# 敏感信息清理指南

## ⚠️ 重要警告

1. **清理历史会改变所有提交的哈希值**
2. **强制推送会影响其他协作者** - 他们需要重新克隆仓库
3. **已泄露的凭证必须轮换** - 即使清理历史，敏感信息可能已被缓存

## 已泄露的敏感信息

根据 config.yaml 历史记录，以下信息已泄露：

- **微信 AppID**: `wx651b36f80c9aad00`
- **微信 AppSecret**: `0b2d8bbd3d57f01c340a683fa25b888d`
- **数据库密码**: `300133ys`

## 立即采取的措施

### 1. 轮换已泄露的凭证（必须在清理历史之前完成！）

#### 微信小程序凭证
1. 登录 [微信小程序后台](https://mp.weixin.qq.com)
2. 进入「开发」->「开发管理」->「开发设置」
3. 点击 AppSecret 的「重置」按钮
4. 更新你的本地 `backend/config/config.yaml` 文件

#### 数据库密码
1. 修改 PostgreSQL/MySQL 的密码
2. 更新你的本地 `backend/config/config.yaml` 文件

### 2. 清理 Git 历史记录

在 PowerShell 中执行以下命令：

```powershell
# 进入项目目录
cd C:\Users\SBEA\Desktop\CNtuanyuan

# 创建备份分支
git branch backup-main

# 使用 git filter-branch 重写历史，移除 config.yaml 中的敏感信息
git filter-branch --force --index-filter `
  "git checkout -- :/backend/config/config.yaml || true" `
  --prune-empty --tag-name-filter cat -- --all

# 清理 reflog
git reflog expire --expire=now --all
git gc --prune=now --aggressive
```

### 3. 强制推送到远程

```powershell
# 强制推送到远程（会覆盖远程历史）
git push origin --force --all
git push origin --force --tags
```

### 4. 清理本地备份文件

```powershell
# 删除本地备份（敏感信息）
Remove-Item backend/config/config.yaml.bak
```

## 其他协作者需要做的

如果其他人已经克隆了这个仓库，他们需要：

```bash
# 1. 备份他们的本地更改
git branch backup-my-changes

# 2. 重新克隆仓库
cd ..
Remove-Item -Recurse -Force CNtunyuan
git clone <repository-url>

# 3. 恢复他们的更改（如果有）
# 手动将 backup-my-changes 分支的更改应用到新克隆的仓库
```

## 验证清理结果

```powershell
# 检查历史中是否还有敏感信息
git log --all --full-history --oneline -- backend/config/config.yaml

# 搜索历史中的敏感字符串
git log --all -S "wx651b36f80c9aad00" --oneline
git log --all -S "0b2d8bbd3d57f01c340a683fa25b888d" --oneline
```

## 预防措施

1. **始终使用 .gitignore 忽略配置文件**
   - 已添加 `backend/config/config.yaml` 到 .gitignore

2. **使用示例配置文件**
   - `backend/config/config.yaml.example` 是安全的模板
   - 复制为 `config.yaml` 并填入真实值

3. **提交前检查**
   ```bash
   git diff --cached
   ```

4. **使用 git-secrets 工具**
   ```bash
   # 安装 git-secrets
   brew install git-secrets  # macOS
   
   # 配置
   git secrets --install
   git secrets --register-aws
   ```

## 如果无法清理历史

如果由于某些原因无法清理历史，至少要做：

1. ✅ 轮换所有已泄露的凭证（必须做！）
2. ✅ 将配置文件添加到 .gitignore（已完成）
3. ✅ 使用示例配置文件（已完成）
4. ⚠️ 考虑将仓库设为私有（如果是公开的）
