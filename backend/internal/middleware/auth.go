package middleware

import (
	"strings"

	"github.com/Snowitty-Re/CNtunyuan/internal/utils"
	"github.com/Snowitty-Re/CNtunyuan/pkg/auth"
	"github.com/gin-gonic/gin"
)

const (
	// ContextKeyUserID 上下文用户ID键
	ContextKeyUserID = "user_id"
	// ContextKeyOpenID 上下文OpenID键
	ContextKeyOpenID = "open_id"
	// ContextKeyUnionID 上下文UnionID键
	ContextKeyUnionID = "union_id"
	// ContextKeyRole 上下文角色键
	ContextKeyRole = "role"
	// ContextKeyOrgID 上下文组织ID键
	ContextKeyOrgID = "org_id"
	// ContextKeyClaims 上下文Claims键
	ContextKeyClaims = "claims"
)

// JWTAuth JWT认证中间件
func JWTAuth(jwtAuth *auth.JWTAuth) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Header获取令牌
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Unauthorized(c, "缺少认证信息")
			c.Abort()
			return
		}

		// 解析Bearer令牌
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			utils.Unauthorized(c, "认证格式错误")
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := jwtAuth.ValidateToken(tokenString)
		if err != nil {
			utils.Unauthorized(c, err.Error())
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyOpenID, claims.OpenID)
		c.Set(ContextKeyUnionID, claims.UnionID)
		c.Set(ContextKeyRole, claims.Role)
		c.Set(ContextKeyOrgID, claims.OrgID)
		c.Set(ContextKeyClaims, claims)

		c.Next()
	}
}

// GetUserID 从上下文获取用户ID
func GetUserID(c *gin.Context) string {
	userID, _ := c.Get(ContextKeyUserID)
	if id, ok := userID.(string); ok {
		return id
	}
	return ""
}

// GetOpenID 从上下文获取OpenID
func GetOpenID(c *gin.Context) string {
	openID, _ := c.Get(ContextKeyOpenID)
	if id, ok := openID.(string); ok {
		return id
	}
	return ""
}

// GetRole 从上下文获取角色
func GetRole(c *gin.Context) string {
	role, _ := c.Get(ContextKeyRole)
	if r, ok := role.(string); ok {
		return r
	}
	return ""
}

// GetOrgID 从上下文获取组织ID
func GetOrgID(c *gin.Context) string {
	orgID, _ := c.Get(ContextKeyOrgID)
	if id, ok := orgID.(string); ok {
		return id
	}
	return ""
}

// GetClaims 从上下文获取Claims
func GetClaims(c *gin.Context) *auth.CustomClaims {
	claims, _ := c.Get(ContextKeyClaims)
	if c, ok := claims.(*auth.CustomClaims); ok {
		return c
	}
	return nil
}

// IsAdmin 检查是否为管理员
func IsAdmin(c *gin.Context) bool {
	role := GetRole(c)
	return role == "super_admin" || role == "admin"
}

// IsManager 检查是否为管理者
func IsManager(c *gin.Context) bool {
	role := GetRole(c)
	return role == "super_admin" || role == "admin" || role == "manager"
}
