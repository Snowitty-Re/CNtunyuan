# Swagger API 文档

## 概述

本项目使用 [Swaggo](https://github.com/swaggo/swag) 自动生成 Swagger API 文档。

## 访问文档

启动服务后，可以通过以下地址访问 API 文档：

- **Swagger UI**: http://localhost:8080/swagger/index.html
- **API Docs 入口**: http://localhost:8080/api/v1/docs

## 文档结构

```
backend/
├── docs/                       # Swagger 文档目录
│   ├── docs.go                # Go 代码格式的文档
│   ├── swagger.json           # JSON 格式文档
│   └── swagger.yaml           # YAML 格式文档
├── internal/interfaces/http/handler/  # Handler Swagger 注释
│   ├── auth_handler.go
│   ├── user_handler.go
│   ├── organization_handler.go
│   ├── missing_person_handler.go
│   ├── task_handler.go
│   ├── dialect_handler.go
│   ├── upload_handler.go
│   ├── dashboard_handler.go
│   └── audit_handler.go
└── cmd/app/main.go            # 主入口 Swagger 配置
```

## API 模块

| 模块 | Tag | 说明 |
|------|-----|------|
| 认证管理 | 认证管理 | 登录、注册、Token刷新、微信登录等 |
| 用户管理 | 用户管理 | 用户 CRUD、角色管理、状态管理等 |
| 个人中心 | 个人中心 | 当前用户资料、密码修改等 |
| 组织管理 | 组织管理 | 组织架构、层级管理、组织树等 |
| 走失人员 | 走失人员 | 走失人员登记、状态更新、轨迹记录等 |
| 任务管理 | 任务管理 | 寻人任务创建、分配、执行、完成等 |
| 方言管理 | 方言管理 | 方言语音库、点赞、评论、精选等 |
| 文件管理 | 文件管理 | 文件上传、下载、管理等 |
| 仪表盘 | 仪表盘 | 数据统计、概览、趋势等 |
| 审计日志 | 审计日志 | 系统操作日志查询、统计等 |
| 健康检查 | 健康检查 | 服务健康状态检查 |

## 认证方式

所有需要认证的 API 都需要在请求头中携带 JWT Token：

```
Authorization: Bearer {your-jwt-token}
```

在 Swagger UI 中，点击 "Authorize" 按钮，输入 `Bearer {token}` 即可。

## 重新生成文档

如果需要重新生成 Swagger 文档（在添加或修改 API 后）：

```bash
cd backend

# 安装 swag 工具
go install github.com/swaggo/swag/cmd/swag@latest

# 生成文档
swag init -g cmd/app/main.go

# 指定输出目录
swag init -g cmd/app/main.go -o docs
```

## Swagger 注释规范

### Handler 方法注释

```go
// MethodName 方法说明
// @Summary 简短描述
// @Description 详细描述
// @Tags 模块名称
// @Accept json
// @Produce json
// @Param id path string true "ID参数"
// @Param body body dto.Request true "请求体"
// @Success 200 {object} response.Response{data=dto.Response}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Security Bearer
// @Router /api/v1/path/{id} [put]
func (h *Handler) MethodName(c *gin.Context) {
    // 实现代码
}
```

### 常用注解说明

| 注解 | 说明 | 示例 |
|------|------|------|
| @Summary | 简短描述 | 用户登录 |
| @Description | 详细描述 | 使用用户名和密码登录系统 |
| @Tags | API 分组 | 认证管理 |
| @Accept | 接收格式 | json, xml, multipart/form-data |
| @Produce | 返回格式 | json, xml |
| @Param | 参数定义 | @Param id path string true "用户ID" |
| @Success | 成功响应 | @Success 200 {object} dto.UserResponse |
| @Failure | 失败响应 | @Failure 400 {object} response.Response |
| @Security | 认证方式 | @Security Bearer |
| @Router | 路由定义 | @Router /users/{id} [get] |

### 参数位置 (in)

- `path` - URL 路径参数 (/users/{id})
- `query` - URL 查询参数 (?page=1)
- `header` - 请求头参数
- `body` - 请求体参数 (JSON)
- `formData` - 表单数据 (文件上传)

### 参数类型

- `string` - 字符串
- `integer` - 整数
- `boolean` - 布尔值
- `number` - 浮点数
- `file` - 文件
- `object` - 对象
- `array` - 数组

## 响应格式

所有 API 返回统一的响应格式：

```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

错误响应：

```json
{
  "code": 400,
  "message": "参数错误",
  "data": null
}
```

## 状态码说明

| 状态码 | 说明 |
|--------|------|
| 200 | 请求成功 |
| 201 | 创建成功 |
| 204 | 删除成功（无内容返回）|
| 400 | 请求参数错误 |
| 401 | 未授权（Token 无效或过期）|
| 403 | 禁止访问（权限不足）|
| 404 | 资源不存在 |
| 409 | 资源冲突（如重复创建）|
| 500 | 服务器内部错误 |

## 示例

### 登录接口

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "13800138000",
    "password": "password123"
  }'
```

### 使用 Token 访问受保护接口

```bash
curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

### 文件上传

```bash
curl -X POST http://localhost:8080/api/v1/upload \
  -H "Authorization: Bearer {token}" \
  -F "file=@/path/to/file.jpg"
```
