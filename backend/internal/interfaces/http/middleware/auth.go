package middleware

import (
	"context"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/service"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"github.com/Snowitty-Re/CNtunyuan/pkg/response"
	"github.com/gin-gonic/gin"
)

var mainlandPhoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)

// AuthMiddleware auth middleware
type AuthMiddleware struct {
	authService *service.AuthService
	userRepo    repository.UserRepository
	statusTTL   time.Duration
	statusMu    sync.Mutex
	statusCache map[string]cachedUserStatus
}

type cachedUserStatus struct {
	user      *entity.User
	expiresAt time.Time
}

// NewAuthMiddleware create auth middleware
func NewAuthMiddleware(authService *service.AuthService, userRepo ...repository.UserRepository) *AuthMiddleware {
	m := &AuthMiddleware{
		authService: authService,
		statusTTL:   30 * time.Second,
		statusCache: make(map[string]cachedUserStatus),
	}
	if len(userRepo) > 0 {
		m.userRepo = userRepo[0]
	}
	return m
}

// Required require auth; enforces active status + real phone except whitelist routes
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

		// Reject refresh tokens on normal API routes
		if claims != nil && strings.EqualFold(claims.TokenType, "refresh") {
			response.Unauthorized(c, "invalid access token")
			c.Abort()
			return
		}

		role := entity.Role(claims.Role)
		orgID := claims.OrgID

		if m.userRepo != nil {
			user, loadErr := m.loadUser(c.Request.Context(), claims.UserID)
			if loadErr != nil || user == nil {
				response.Unauthorized(c, "user not found, please login again")
				c.Abort()
				return
			}
			// Prefer live DB role/org over stale JWT claims
			role = user.Role
			orgID = user.OrgID

			path := c.Request.URL.Path
			whitelist := isAccountBootstrapPath(path)
			if !user.IsActive() && !whitelist {
				response.Forbidden(c, "账号未激活或已禁用，请等待审批或联系管理员")
				c.Abort()
				return
			}
			if !mainlandPhoneRegex.MatchString(strings.TrimSpace(user.Phone)) && !whitelist {
				response.Forbidden(c, "请先绑定真实手机号")
				c.Abort()
				return
			}
			c.Set("userPhone", user.Phone)
			c.Set("userStatus", string(user.Status))
		}

		c.Set("userID", claims.UserID)
		c.Set("userRole", role)
		c.Set("orgID", orgID)
		c.Set("claims", claims)

		c.Next()
	}
}

// Optional optional auth (no status/phone enforcement — used for public views)
func (m *AuthMiddleware) Optional() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractToken(c)
		if token != "" {
			claims, err := m.authService.ValidateToken(c.Request.Context(), token)
			if err == nil && claims != nil && !strings.EqualFold(claims.TokenType, "refresh") {
				// Only mark authenticated when user is active with real phone
				if m.userRepo != nil {
					user, loadErr := m.loadUser(c.Request.Context(), claims.UserID)
					if loadErr == nil && user != nil && user.IsActive() && mainlandPhoneRegex.MatchString(strings.TrimSpace(user.Phone)) {
						c.Set("userID", user.ID)
						c.Set("userRole", user.Role)
						c.Set("orgID", user.OrgID)
						c.Set("claims", claims)
					}
				} else {
					c.Set("userID", claims.UserID)
					c.Set("userRole", entity.Role(claims.Role))
					c.Set("orgID", claims.OrgID)
					c.Set("claims", claims)
				}
			}
		}
		c.Next()
	}
}

func isAccountBootstrapPath(path string) bool {
	// Allow incomplete accounts to finish onboarding
	whitelistSuffixes := []string{
		"/auth/bind-phone",
		"/auth/logout",
		"/auth/me",
		"/auth/refresh",
	}
	for _, s := range whitelistSuffixes {
		if strings.HasSuffix(path, s) {
			return true
		}
	}
	return false
}

func (m *AuthMiddleware) loadUser(ctx context.Context, userID string) (*entity.User, error) {
	if m.userRepo == nil {
		return nil, nil
	}
	now := time.Now()
	m.statusMu.Lock()
	if cached, ok := m.statusCache[userID]; ok && cached.expiresAt.After(now) {
		u := cached.user
		m.statusMu.Unlock()
		return u, nil
	}
	m.statusMu.Unlock()

	user, err := m.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	m.statusMu.Lock()
	m.statusCache[userID] = cachedUserStatus{user: user, expiresAt: now.Add(m.statusTTL)}
	// opportunistic prune
	if len(m.statusCache) > 5000 {
		for k, v := range m.statusCache {
			if v.expiresAt.Before(now) {
				delete(m.statusCache, k)
			}
		}
	}
	m.statusMu.Unlock()
	return user, nil
}

// extractToken extract token from request (header only, no query param for security)
func (m *AuthMiddleware) extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return parts[1]
		}
	}

	return ""
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
