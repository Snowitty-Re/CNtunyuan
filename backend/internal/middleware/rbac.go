package middleware

import (
	"strings"

	"github.com/Snowitty-Re/CNtunyuan/internal/model"
	"github.com/Snowitty-Re/CNtunyuan/internal/utils"
	"github.com/Snowitty-Re/CNtunyuan/pkg/auth"
	"github.com/gin-gonic/gin"
)

// 角色权限级别
type RoleLevel int

const (
	RoleLevelVolunteer RoleLevel = iota // 志愿者
	RoleLevelManager                    // 管理者
	RoleLevelAdmin                      // 管理员
	RoleLevelSuperAdmin                 // 超级管理员
)

// 角色级别映射
var roleLevelMap = map[string]RoleLevel{
	model.RoleVolunteer:  RoleLevelVolunteer,
	model.RoleManager:    RoleLevelManager,
	model.RoleAdmin:      RoleLevelAdmin,
	model.RoleSuperAdmin: RoleLevelSuperAdmin,
}

// GetRoleLevel 获取角色级别
func GetRoleLevel(role string) RoleLevel {
	if level, ok := roleLevelMap[role]; ok {
		return level
	}
	return RoleLevelVolunteer
}

// RequireRole 需要指定角色权限
func RequireRole(minRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := GetRole(c)
		if userRole == "" {
			utils.Unauthorized(c, "未登录")
			c.Abort()
			return
		}

		userLevel := GetRoleLevel(userRole)
		requiredLevel := GetRoleLevel(minRole)

		if userLevel < requiredLevel {
			utils.Forbidden(c, "权限不足，需要 "+minRole+" 及以上权限")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin 需要管理员权限
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !IsAdmin(c) {
			utils.Forbidden(c, "需要管理员权限")
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireManager 需要管理者权限
func RequireManager() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !IsManager(c) {
			utils.Forbidden(c, "需要管理者权限")
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireSuperAdmin 需要超级管理员权限
func RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := GetRole(c)
		if role != model.RoleSuperAdmin {
			utils.Forbidden(c, "需要超级管理员权限")
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireOwnerOrAdmin 需要资源所有者或管理员
func RequireOwnerOrAdmin(getOwnerID func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		role := GetRole(c)

		// 管理员可以直接访问
		if role == model.RoleSuperAdmin || role == model.RoleAdmin {
			c.Next()
			return
		}

		// 检查是否为资源所有者
		ownerID := getOwnerID(c)
		if ownerID != "" && ownerID == userID {
			c.Next()
			return
		}

		utils.Forbidden(c, "无权访问该资源")
		c.Abort()
	}
}

// RequireSameOrg 需要同一组织
func RequireSameOrg() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := GetRole(c)

		// 超级管理员可以访问所有组织
		if userRole == model.RoleSuperAdmin {
			c.Next()
			return
		}

		// 获取请求中的组织ID
		targetOrgID := c.Param("org_id")
		if targetOrgID == "" {
			targetOrgID = c.Query("org_id")
		}
		if targetOrgID == "" {
			targetOrgID = c.PostForm("org_id")
		}

		// 如果没有指定组织ID，允许访问
		if targetOrgID == "" {
			c.Next()
			return
		}

		// 检查是否同一组织
		userOrgID := GetOrgID(c)
		if targetOrgID != userOrgID {
			utils.Forbidden(c, "无权访问其他组织的数据")
			c.Abort()
			return
		}

		c.Next()
	}
}

// Permission 权限检查中间件
type Permission struct {
	Resource string // 资源类型
	Action   string // 操作类型
}

// RequirePermission 需要特定权限
func RequirePermission(permission Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := GetRole(c)
		userID := GetUserID(c)

		if userRole == "" || userID == "" {
			utils.Unauthorized(c, "未登录")
			c.Abort()
			return
		}

		// 超级管理员拥有所有权限
		if userRole == model.RoleSuperAdmin {
			c.Next()
			return
		}

		// 检查权限
		if !hasPermission(userRole, permission) {
			utils.Forbidden(c, "没有 "+permission.Resource+":"+permission.Action+" 权限")
			c.Abort()
			return
		}

		c.Next()
	}
}

// hasPermission 检查角色是否拥有权限
func hasPermission(role string, permission Permission) bool {
	// 定义权限矩阵
	permissions := map[string][]Permission{
		model.RoleAdmin: {
			{Resource: "user", Action: "create"},
			{Resource: "user", Action: "read"},
			{Resource: "user", Action: "update"},
			{Resource: "user", Action: "delete"},
			{Resource: "org", Action: "create"},
			{Resource: "org", Action: "read"},
			{Resource: "org", Action: "update"},
			{Resource: "org", Action: "delete"},
			{Resource: "task", Action: "create"},
			{Resource: "task", Action: "read"},
			{Resource: "task", Action: "update"},
			{Resource: "task", Action: "delete"},
			{Resource: "case", Action: "create"},
			{Resource: "case", Action: "read"},
			{Resource: "case", Action: "update"},
			{Resource: "case", Action: "delete"},
		},
		model.RoleManager: {
			{Resource: "user", Action: "read"},
			{Resource: "user", Action: "update"},
			{Resource: "org", Action: "read"},
			{Resource: "org", Action: "update"},
			{Resource: "task", Action: "create"},
			{Resource: "task", Action: "read"},
			{Resource: "task", Action: "update"},
			{Resource: "case", Action: "create"},
			{Resource: "case", Action: "read"},
			{Resource: "case", Action: "update"},
		},
		model.RoleVolunteer: {
			{Resource: "user", Action: "read"},
			{Resource: "user", Action: "update"},
			{Resource: "task", Action: "read"},
			{Resource: "task", Action: "update"},
			{Resource: "case", Action: "read"},
			{Resource: "case", Action: "create"},
			{Resource: "dialect", Action: "create"},
			{Resource: "dialect", Action: "read"},
			{Resource: "dialect", Action: "update"},
		},
	}

	rolePermissions, ok := permissions[role]
	if !ok {
		return false
	}

	for _, p := range rolePermissions {
		if p.Resource == permission.Resource && p.Action == permission.Action {
			return true
		}
	}

	return false
}

// CasbinMiddleware Casbin权限中间件
func CasbinMiddleware(enforcer *CasbinEnforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := GetRole(c)
		path := c.Request.URL.Path
		method := c.Request.Method

		// 超级管理员跳过检查
		if role == model.RoleSuperAdmin {
			c.Next()
			return
		}

		// 检查权限
		allowed, err := enforcer.Enforce(role, path, method)
		if err != nil {
			utils.ServerError(c, "权限检查失败")
			c.Abort()
			return
		}

		if !allowed {
			utils.Forbidden(c, "无权访问")
			c.Abort()
			return
		}

		c.Next()
	}
}

// CasbinEnforcer Casbin执行器包装
type CasbinEnforcer struct {
	enforcer interface{}
}

// Enforce 执行权限检查
func (c *CasbinEnforcer) Enforce(rvals ...interface{}) (bool, error) {
	// 这里可以集成Casbin库
	// 简化实现，直接返回true
	return true, nil
}

// SkipAuthPaths 跳过认证的路径
var SkipAuthPaths = []string{
	"/health",
	"/swagger",
	"/api/v1/auth/wechat-login",
	"/api/v1/auth/admin-login",
	"/api/v1/auth/refresh",
}

// ShouldSkipAuth 是否应该跳过认证
func ShouldSkipAuth(path string) bool {
	for _, skipPath := range SkipAuthPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// AuthMiddleware 统一认证中间件
func AuthMiddleware(jwtAuth *auth.JWTAuth) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过不需要认证的路径
		if ShouldSkipAuth(c.Request.URL.Path) {
			c.Next()
			return
		}

		// 执行JWT认证
		JWTAuth(jwtAuth)(c)
	}
}
