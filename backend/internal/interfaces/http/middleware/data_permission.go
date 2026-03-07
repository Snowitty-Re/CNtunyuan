package middleware

import (
	"github.com/Snowitty-Re/CNtunyuan/internal/infrastructure/permission"
	"github.com/Snowitty-Re/CNtunyuan/pkg/response"
	"github.com/gin-gonic/gin"
)

// DataPermissionMiddleware 数据权限中间件
type DataPermissionMiddleware struct {
	provider permission.DataPermissionProvider
}

// NewDataPermissionMiddleware 创建数据权限中间件
func NewDataPermissionMiddleware(provider permission.DataPermissionProvider) *DataPermissionMiddleware {
	return &DataPermissionMiddleware{
		provider: provider,
	}
}

// InjectContext 注入数据权限上下文到 gin.Context
func (m *DataPermissionMiddleware) InjectContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 token 或 context 获取用户信息
		userID := GetUserID(c)
		orgID := GetOrgID(c)
		userRole := GetUserRole(c)
		
		// 设置到 context
		if userID != "" {
			c.Set("user_id", userID)
		}
		if orgID != "" {
			c.Set("org_id", orgID)
		}
		if userRole != "" {
			c.Set("user_role", string(userRole))
		}
		
		c.Next()
	}
}

// RequireOrgAccess 要求有特定组织访问权限
func (m *DataPermissionMiddleware) RequireOrgAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取目标组织ID（从参数或查询）
		targetOrgID := c.Param("org_id")
		if targetOrgID == "" {
			targetOrgID = c.Query("org_id")
		}
		if targetOrgID == "" {
			targetOrgID = c.PostForm("org_id")
		}
		
		// 如果没有指定组织ID，让后续处理
		if targetOrgID == "" {
			c.Next()
			return
		}
		
		// 检查权限
		if err := permission.CheckDataPermission(c.Request.Context(), m.provider, targetOrgID); err != nil {
			response.Forbidden(c, "无权访问该组织数据")
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// RequireResourceAccess 检查资源访问权限
func (m *DataPermissionMiddleware) RequireResourceAccess(getOrgID func(c *gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		targetOrgID := getOrgID(c)
		
		if targetOrgID == "" {
			c.Next()
			return
		}
		
		if err := permission.CheckDataPermission(c.Request.Context(), m.provider, targetOrgID); err != nil {
			response.Forbidden(c, "无权访问该资源")
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// GetDataPermissionContext 获取数据权限上下文
func GetDataPermissionContext(c *gin.Context, provider permission.DataPermissionProvider) (*permission.DataPermissionContext, error) {
	return provider.GetDataPermissionContext(c.Request.Context())
}

// HasOrgPermission 检查是否有组织权限
func HasOrgPermission(c *gin.Context, targetOrgID string, provider permission.DataPermissionProvider) bool {
	dpCtx, err := GetDataPermissionContext(c, provider)
	if err != nil {
		return false
	}
	return dpCtx.HasPermission(targetOrgID)
}
