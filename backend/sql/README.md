# 数据库初始化指南

## 概述

团圆寻亲志愿者系统使用 GORM AutoMigrate 自动管理数据库结构，使用种子工具初始化数据。

## 数据初始化

### 使用种子工具（推荐）

```bash
cd backend

# 导入所有种子数据（组织和用户）
go run cmd/seed/main.go -all

# 只导入组织
go run cmd/seed/main.go -orgs

# 只导入用户
go run cmd/seed/main.go -users
```

### 清空数据重新导入

```bash
# 警告：这会删除所有数据！
go run cmd/seed/main.go -clean -all
```

## 默认账号

种子数据导入后会创建默认超级管理员：
- 手机号: `13800138000`
- 密码: `admin123`
- 角色: super_admin

## 修改密码

```bash
# 使用密码重置工具
go run cmd/resetpassword/main.go -phone="13800138000" -password="newpassword"
```

## 常见问题

### Q: 数据库连接失败？

1. 检查数据库是否启动
2. 检查 `config/config.yaml` 中的数据库配置
3. 确保数据库 `cntuanyuan` 已创建

### Q: 如何添加其他管理员？

通过 Web 管理后台的"志愿者管理"功能添加，角色选择：
- `admin` - 普通管理员
- `manager` - 管理者

## 生产环境建议

1. **修改默认密码**：首次部署后务必修改超级管理员默认密码
2. **使用环境变量**：生产环境建议通过环境变量传递敏感配置
3. **定期备份**：定期备份数据库
