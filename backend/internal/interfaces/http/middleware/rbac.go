package middleware

import (
	"net/http"
	"strings"

	"github.com/Snowitty-Re/CNtunyuan/internal/application/service"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"github.com/Snowitty-Re/CNtunyuan/pkg/response"
	"github.com/gin-gonic/gin"
)

// RBACMiddleware RBAC 权限中间件
type RBACMiddleware struct {
	permService *service.PermissionAppService
}

// NewRBACMiddleware 创建 RBAC 中间件
func NewRBACMiddleware(permService *service.PermissionAppService) *RBACMiddleware {
	return &RBACMiddleware{
		permService: permService,
	}
}

// RequirePermission 要求特定权限
func (m *RBACMiddleware) RequirePermission(resource entity.PermissionResource, action entity.PermissionAction) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == "" {
			response.Unauthorized(c, "user not authenticated")
			c.Abort()
			return
		}
		
		result, err := m.permService.CheckPermission(c.Request.Context(), userID, resource, action)
		if err != nil {
			logger.Error("Failed to check permission", logger.Err(err))
			response.InternalServerError(c, "permission check failed")
			c.Abort()
			return
		}
		
		if !result.Allowed {
			logger.Warn("Permission denied",
				logger.String("user_id", userID),
				logger.String("resource", string(resource)),
				logger.String("action", string(action)),
			)
			response.Forbidden(c, "permission denied: "+result.MissingPerm)
			c.Abort()
			return
		}
		
		// 设置数据范围到上下文
		c.Set("data_scope", result.DataScope)
		
		c.Next()
	}
}

// RequireAnyPermission 要求任意一个权限
func (m *RBACMiddleware) RequireAnyPermission(perms []struct {
	Resource entity.PermissionResource
	Action   entity.PermissionAction
}) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == "" {
			response.Unauthorized(c, "user not authenticated")
			c.Abort()
			return
		}
		
		for _, perm := range perms {
			result, err := m.permService.CheckPermission(c.Request.Context(), userID, perm.Resource, perm.Action)
			if err != nil {
				continue
			}
			if result.Allowed {
				c.Set("data_scope", result.DataScope)
				c.Next()
				return
			}
		}
		
		response.Forbidden(c, "permission denied")
		c.Abort()
	}
}

// RequireAllPermissions 要求所有权限
func (m *RBACMiddleware) RequireAllPermissions(perms []struct {
	Resource entity.PermissionResource
	Action   entity.PermissionAction
}) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == "" {
			response.Unauthorized(c, "user not authenticated")
			c.Abort()
			return
		}
		
		var missingPerms []string
		for _, perm := range perms {
			result, err := m.permService.CheckPermission(c.Request.Context(), userID, perm.Resource, perm.Action)
			if err != nil || !result.Allowed {
				missingPerms = append(missingPerms, entity.GetCode(perm.Resource, perm.Action))
			}
		}
		
		if len(missingPerms) > 0 {
			response.Forbidden(c, "permission denied: "+strings.Join(missingPerms, ", "))
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// FieldPermissionFilter 字段权限过滤器
func (m *RBACMiddleware) FieldPermissionFilter(resource string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == "" {
			c.Next()
			return
		}
		
		// 在响应处理后过滤字段
		c.Next()
		
		// TODO: 实现响应字段过滤
		// 这里需要捕获响应数据进行过滤
	}
}

// PermissionChecker 权限检查器
func PermissionChecker(permService *service.PermissionAppService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 在上下文中添加权限检查函数
		c.Set("check_permission", func(resource entity.PermissionResource, action entity.PermissionAction) bool {
			userID := GetUserID(c)
			if userID == "" {
				return false
			}
			
			result, err := permService.CheckPermission(c.Request.Context(), userID, resource, action)
			if err != nil {
				return false
			}
			
			if result.Allowed {
				c.Set("data_scope", result.DataScope)
			}
			
			return result.Allowed
		})
		
		c.Next()
	}
}

// CheckPermission 在 handler 中检查权限
func CheckPermission(c *gin.Context, resource entity.PermissionResource, action entity.PermissionAction) bool {
	checker, exists := c.Get("check_permission")
	if !exists {
		return false
	}
	
	if fn, ok := checker.(func(entity.PermissionResource, entity.PermissionAction) bool); ok {
		return fn(resource, action)
	}
	
	return false
}

// ResourceFromPath 从路径解析资源
func ResourceFromPath(path string) entity.PermissionResource {
	// 移除 /api/v1/ 前缀
	if len(path) > 8 && path[:8] == "/api/v1/" {
		path = path[8:]
	}
	
	// 获取第一个路径段
	parts := strings.SplitN(path, "/", 2)
	resource := parts[0]
	
	switch resource {
	case "users":
		return entity.ResourceUser
	case "organizations":
		return entity.ResourceOrganization
	case "tasks":
		return entity.ResourceTask
	case "missing-persons":
		return entity.ResourceMissingPerson
	case "dialects":
		return entity.ResourceDialect
	case "files":
		return entity.ResourceFile
	case "workflows", "workflow-definitions", "workflow-instances", "workflow-tasks":
		return entity.ResourceWorkflow
	case "audit-logs":
		return entity.ResourceAuditLog
	case "dashboard":
		return entity.ResourceDashboard
	default:
		return entity.ResourceSystem
	}
}

// ActionFromMethod 从 HTTP 方法解析操作
func ActionFromMethod(method string) entity.PermissionAction {
	switch method {
	case http.MethodGet:
		return entity.ActionRead
	case http.MethodPost:
		return entity.ActionCreate
	case http.MethodPut, http.MethodPatch:
		return entity.ActionUpdate
	case http.MethodDelete:
		return entity.ActionDelete
	default:
		return entity.ActionAll
	}
}

// AutoPermissionCheck 自动权限检查中间件
func AutoPermissionCheck(permService *service.PermissionAppService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过公开路由
		if isPublicPath(c.Request.URL.Path) {
			c.Next()
			return
		}
		
		userID := GetUserID(c)
		if userID == "" {
			c.Next()
			return
		}
		
		resource := ResourceFromPath(c.Request.URL.Path)
		action := ActionFromMethod(c.Request.Method)
		
		result, err := permService.CheckPermission(c.Request.Context(), userID, resource, action)
		if err != nil {
			logger.Error("Failed to check permission", logger.Err(err))
			c.Next()
			return
		}
		
		if !result.Allowed {
			logger.Warn("Permission denied",
				logger.String("user_id", userID),
				logger.String("path", c.Request.URL.Path),
				logger.String("resource", string(resource)),
				logger.String("action", string(action)),
			)
			response.Forbidden(c, "permission denied")
			c.Abort()
			return
		}
		
		c.Set("data_scope", result.DataScope)
		c.Next()
	}
}

// isPublicPath 检查是否是公开路径
func isPublicPath(path string) bool {
	publicPaths := []string{
		"/api/v1/health",
		"/api/v1/metrics",
		"/api/v1/auth",
		"/uploads/",
	}
	
	for _, pp := range publicPaths {
		if strings.HasPrefix(path, pp) {
			return true
		}
	}
	
	return false
}

// RequireResourceAccess 要求资源访问权限（增强版）
func RequireResourceAccess(permService *service.PermissionAppService, resource entity.PermissionResource, actions ...entity.PermissionAction) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == "" {
			response.Unauthorized(c, "user not authenticated")
			c.Abort()
			return
		}
		
		// 如果没有指定操作，根据 HTTP 方法自动判断
		var action entity.PermissionAction
		if len(actions) > 0 {
			action = actions[0]
		} else {
			action = ActionFromMethod(c.Request.Method)
		}
		
		result, err := permService.CheckPermission(c.Request.Context(), userID, resource, action)
		if err != nil {
			logger.Error("Failed to check permission", logger.Err(err))
			response.InternalServerError(c, "permission check failed")
			c.Abort()
			return
		}
		
		if !result.Allowed {
			response.Forbidden(c, "permission denied for "+string(resource)+":"+string(action))
			c.Abort()
			return
		}
		
		// 设置权限信息到上下文
		c.Set("data_scope", result.DataScope)
		c.Set("permission_check", result)
		
		c.Next()
	}
}
