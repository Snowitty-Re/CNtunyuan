package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/service"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"github.com/Snowitty-Re/CNtunyuan/pkg/response"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware auth middleware
type AuthMiddleware struct {
	authService *service.AuthService
}

// NewAuthMiddleware create auth middleware
func NewAuthMiddleware(authService *service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

// Required require auth
func (m *AuthMiddleware) Required() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractToken(c)
		if token == "" {
			response.Unauthorized(c, "please login first")
			c.Abort()
			return
		}

		claims, err := m.authService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			logger.Warn("Token validation failed", logger.Err(err))
			response.Unauthorized(c, "token expired, please login again")
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("userRole", claims.Role)
		c.Set("orgID", claims.OrgID)
		c.Set("claims", claims)

		c.Next()
	}
}

// Optional optional auth
func (m *AuthMiddleware) Optional() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractToken(c)
		if token != "" {
			claims, err := m.authService.ValidateToken(c.Request.Context(), token)
			if err == nil {
				c.Set("userID", claims.UserID)
				c.Set("userRole", claims.Role)
				c.Set("orgID", claims.OrgID)
				c.Set("claims", claims)
			}
		}
		c.Next()
	}
}

// extractToken extract token from request
func (m *AuthMiddleware) extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return parts[1]
		}
	}

	return c.Query("token")
}

// RequireRole require role
func RequireRole(minRole entity.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("userRole")
		if !exists {
			response.Unauthorized(c, "please login first")
			c.Abort()
			return
		}

		role, ok := userRole.(entity.Role)
		if !ok {
			response.Unauthorized(c, "invalid user info")
			c.Abort()
			return
		}

		if !entity.HasRole(role, minRole) {
			response.Forbidden(c, "permission denied")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin require admin
func RequireAdmin() gin.HandlerFunc {
	return RequireRole(entity.RoleAdmin)
}

// RequireManager require manager
func RequireManager() gin.HandlerFunc {
	return RequireRole(entity.RoleManager)
}

// RequireSuperAdmin require super admin
func RequireSuperAdmin() gin.HandlerFunc {
	return RequireRole(entity.RoleSuperAdmin)
}

// GetUserID get user ID from context
func GetUserID(c *gin.Context) string {
	userID, _ := c.Get("userID")
	if id, ok := userID.(string); ok {
		return id
	}
	return ""
}

// GetUserRole get user role from context
func GetUserRole(c *gin.Context) entity.Role {
	userRole, _ := c.Get("userRole")
	if role, ok := userRole.(entity.Role); ok {
		return role
	}
	return ""
}

// GetOrgID get org ID from context
func GetOrgID(c *gin.Context) string {
	orgID, _ := c.Get("orgID")
	if id, ok := orgID.(string); ok {
		return id
	}
	return ""
}

// GetClaims get claims from context
func GetClaims(c *gin.Context) *service.TokenClaims {
	claims, _ := c.Get("claims")
	if c, ok := claims.(*service.TokenClaims); ok {
		return c
	}
	return nil
}

// IsAuthenticated check if authenticated
func IsAuthenticated(c *gin.Context) bool {
	_, exists := c.Get("userID")
	return exists
}

// IsAdmin check if admin
func IsAdmin(c *gin.Context) bool {
	role := GetUserRole(c)
	return role == entity.RoleAdmin || role == entity.RoleSuperAdmin
}

// IsSuperAdmin check if super admin
func IsSuperAdmin(c *gin.Context) bool {
	role := GetUserRole(c)
	return role == entity.RoleSuperAdmin
}

// CORSMiddleware CORS middleware
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RequestLoggerMiddleware request logger middleware
func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		logger.Info("HTTP Request",
			logger.String("client_ip", clientIP),
			logger.String("method", method),
			logger.String("path", path),
			logger.Int("status", statusCode),
			logger.Duration("latency", latency),
		)
	}
}

// RecoveryMiddleware recovery middleware
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered", logger.Any("error", err))
				response.InternalServerError(c, "internal server error")
				c.Abort()
			}
		}()
		c.Next()
	}
}
