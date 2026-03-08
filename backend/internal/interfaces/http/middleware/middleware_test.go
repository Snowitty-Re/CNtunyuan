// Package middleware_test HTTP中间件测试
package middleware_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/service"
	ifmiddleware "github.com/Snowitty-Re/CNtunyuan/internal/interfaces/http/middleware"
	"github.com/Snowitty-Re/CNtunyuan/internal/testutil"
	pkgerrors "github.com/Snowitty-Re/CNtunyuan/pkg/errors"
	pkgmiddleware "github.com/Snowitty-Re/CNtunyuan/pkg/middleware"
	"github.com/Snowitty-Re/CNtunyuan/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// setupTestRouter 创建测试路由
func setupTestRouter() *gin.Engine {
	return gin.New()
}

// performRequest 执行HTTP请求测试
func performRequest(router *gin.Engine, method, path string, headers map[string]string, body string) *httptest.ResponseRecorder {
	var bodyReader *bytes.Reader
	if body != "" {
		bodyReader = bytes.NewReader([]byte(body))
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// parseResponse 解析标准响应格式 (code, message, data)
func parseResponse(w *httptest.ResponseRecorder) map[string]interface{} {
	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)
	return result
}

// getDataFromResponse 从响应中获取data字段
func getDataFromResponse(resp map[string]interface{}) map[string]interface{} {
	if data, ok := resp["data"].(map[string]interface{}); ok {
		return data
	}
	return nil
}

// ==================== RBACMiddleware Tests ====================

func TestRequireRole_WithSufficientRole(t *testing.T) {
	tests := []struct {
		name       string
		userRole   entity.Role
		required   entity.Role
		shouldPass bool
	}{
		{"SuperAdmin can access Admin resource", entity.RoleSuperAdmin, entity.RoleAdmin, true},
		{"Admin can access Admin resource", entity.RoleAdmin, entity.RoleAdmin, true},
		{"Manager cannot access Admin resource", entity.RoleManager, entity.RoleAdmin, false},
		{"Volunteer cannot access Admin resource", entity.RoleVolunteer, entity.RoleAdmin, false},
		{"SuperAdmin can access SuperAdmin resource", entity.RoleSuperAdmin, entity.RoleSuperAdmin, true},
		{"Admin cannot access SuperAdmin resource", entity.RoleAdmin, entity.RoleSuperAdmin, false},
		{"Manager can access Manager resource", entity.RoleManager, entity.RoleManager, true},
		{"Volunteer cannot access Manager resource", entity.RoleVolunteer, entity.RoleManager, false},
		{"Any role can access Volunteer resource", entity.RoleVolunteer, entity.RoleVolunteer, true},
		{"Admin can access Volunteer resource", entity.RoleAdmin, entity.RoleVolunteer, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter()
			router.Use(func(c *gin.Context) {
				c.Set("userRole", tt.userRole)
				c.Next()
			})
			router.Use(ifmiddleware.RequireRole(tt.required))
			router.GET("/admin", func(c *gin.Context) {
				c.String(http.StatusOK, "success")
			})

			w := performRequest(router, "GET", "/admin", nil, "")

			if tt.shouldPass {
				assert.Equal(t, http.StatusOK, w.Code, tt.name)
			} else {
				assert.Equal(t, http.StatusForbidden, w.Code, tt.name)
			}
		})
	}
}

func TestRequireRole_WithoutRoleInfo(t *testing.T) {
	router := setupTestRouter()
	router.Use(ifmiddleware.RequireRole(entity.RoleAdmin))
	router.GET("/admin", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	w := performRequest(router, "GET", "/admin", nil, "")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	resp := parseResponse(w)
	assert.Equal(t, float64(401), resp["code"])
}

func TestRequireRole_WithInvalidRoleType(t *testing.T) {
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("userRole", "invalid-role-type")
		c.Next()
	})
	router.Use(ifmiddleware.RequireRole(entity.RoleAdmin))
	router.GET("/admin", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	w := performRequest(router, "GET", "/admin", nil, "")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	resp := parseResponse(w)
	assert.Equal(t, float64(401), resp["code"])
}

func TestRequireAdmin(t *testing.T) {
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("userRole", entity.RoleAdmin)
		c.Next()
	})
	router.Use(ifmiddleware.RequireAdmin())
	router.GET("/admin", func(c *gin.Context) {
		c.String(http.StatusOK, "admin access")
	})

	w := performRequest(router, "GET", "/admin", nil, "")
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireManager(t *testing.T) {
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("userRole", entity.RoleManager)
		c.Next()
	})
	router.Use(ifmiddleware.RequireManager())
	router.GET("/manager", func(c *gin.Context) {
		c.String(http.StatusOK, "manager access")
	})

	w := performRequest(router, "GET", "/manager", nil, "")
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireSuperAdmin(t *testing.T) {
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("userRole", entity.RoleSuperAdmin)
		c.Next()
	})
	router.Use(ifmiddleware.RequireSuperAdmin())
	router.GET("/super-admin", func(c *gin.Context) {
		c.String(http.StatusOK, "super admin access")
	})

	w := performRequest(router, "GET", "/super-admin", nil, "")
	assert.Equal(t, http.StatusOK, w.Code)
}

// ==================== Helper Functions Tests ====================

func TestGetUserID(t *testing.T) {
	router := setupTestRouter()
	router.GET("/test", func(c *gin.Context) {
		c.Set("userID", "test-user-id")
		userID := ifmiddleware.GetUserID(c)
		response.Success(c, gin.H{"user_id": userID})
	})

	w := performRequest(router, "GET", "/test", nil, "")

	resp := parseResponse(w)
	data := getDataFromResponse(resp)
	assert.NotNil(t, data)
	assert.Equal(t, "test-user-id", data["user_id"])
}

func TestGetUserRole(t *testing.T) {
	router := setupTestRouter()
	router.GET("/test", func(c *gin.Context) {
		c.Set("userRole", entity.RoleAdmin)
		role := ifmiddleware.GetUserRole(c)
		response.Success(c, gin.H{"role": string(role)})
	})

	w := performRequest(router, "GET", "/test", nil, "")

	resp := parseResponse(w)
	data := getDataFromResponse(resp)
	assert.NotNil(t, data)
	assert.Equal(t, "admin", data["role"])
}

func TestGetOrgID(t *testing.T) {
	router := setupTestRouter()
	router.GET("/test", func(c *gin.Context) {
		c.Set("orgID", "test-org-id")
		orgID := ifmiddleware.GetOrgID(c)
		response.Success(c, gin.H{"org_id": orgID})
	})

	w := performRequest(router, "GET", "/test", nil, "")

	resp := parseResponse(w)
	data := getDataFromResponse(resp)
	assert.NotNil(t, data)
	assert.Equal(t, "test-org-id", data["org_id"])
}

func TestGetClaims(t *testing.T) {
	claims := &service.TokenClaims{
		UserID:   "user-123",
		Nickname: "Test",
		Role:     "volunteer",
		OrgID:    "org-123",
	}

	router := setupTestRouter()
	router.GET("/test", func(c *gin.Context) {
		c.Set("claims", claims)
		retrievedClaims := ifmiddleware.GetClaims(c)
		response.Success(c, gin.H{
			"user_id": retrievedClaims.UserID,
			"role":    retrievedClaims.Role,
		})
	})

	w := performRequest(router, "GET", "/test", nil, "")

	resp := parseResponse(w)
	data := getDataFromResponse(resp)
	assert.NotNil(t, data)
	assert.Equal(t, "user-123", data["user_id"])
	assert.Equal(t, "volunteer", data["role"])
}

func TestIsAuthenticated(t *testing.T) {
	router := setupTestRouter()
	router.GET("/auth", func(c *gin.Context) {
		c.Set("userID", "user-123")
		isAuth := ifmiddleware.IsAuthenticated(c)
		response.Success(c, gin.H{"authenticated": isAuth})
	})

	w := performRequest(router, "GET", "/auth", nil, "")

	resp := parseResponse(w)
	data := getDataFromResponse(resp)
	assert.NotNil(t, data)
	assert.True(t, data["authenticated"].(bool))
}

func TestIsAuthenticated_NotAuthenticated(t *testing.T) {
	router := setupTestRouter()
	router.GET("/no-auth", func(c *gin.Context) {
		isAuth := ifmiddleware.IsAuthenticated(c)
		response.Success(c, gin.H{"authenticated": isAuth})
	})

	w := performRequest(router, "GET", "/no-auth", nil, "")

	resp := parseResponse(w)
	data := getDataFromResponse(resp)
	assert.NotNil(t, data)
	assert.False(t, data["authenticated"].(bool))
}

func TestIsAdmin(t *testing.T) {
	tests := []struct {
		role     entity.Role
		expected bool
	}{
		{entity.RoleSuperAdmin, true},
		{entity.RoleAdmin, true},
		{entity.RoleManager, false},
		{entity.RoleVolunteer, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			router := setupTestRouter()
			router.GET("/test", func(c *gin.Context) {
				c.Set("userRole", tt.role)
				isAdmin := ifmiddleware.IsAdmin(c)
				response.Success(c, gin.H{"is_admin": isAdmin})
			})

			w := performRequest(router, "GET", "/test", nil, "")

			resp := parseResponse(w)
			data := getDataFromResponse(resp)
			assert.NotNil(t, data)
			assert.Equal(t, tt.expected, data["is_admin"].(bool))
		})
	}
}

func TestIsSuperAdmin(t *testing.T) {
	tests := []struct {
		role     entity.Role
		expected bool
	}{
		{entity.RoleSuperAdmin, true},
		{entity.RoleAdmin, false},
		{entity.RoleManager, false},
		{entity.RoleVolunteer, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			router := setupTestRouter()
			router.GET("/test", func(c *gin.Context) {
				c.Set("userRole", tt.role)
				isSuperAdmin := ifmiddleware.IsSuperAdmin(c)
				response.Success(c, gin.H{"is_super_admin": isSuperAdmin})
			})

			w := performRequest(router, "GET", "/test", nil, "")

			resp := parseResponse(w)
			data := getDataFromResponse(resp)
			assert.NotNil(t, data)
			assert.Equal(t, tt.expected, data["is_super_admin"].(bool))
		})
	}
}

// ==================== CORS Middleware Tests ====================

func TestCORSMiddleware_AllowedOrigin(t *testing.T) {
	router := setupTestRouter()
	router.Use(pkgmiddleware.CORSMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	w := performRequest(router, "GET", "/test", map[string]string{
		"Origin": "http://localhost:3000",
	}, "")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
}

func TestCORSMiddleware_AnotherAllowedOrigin(t *testing.T) {
	router := setupTestRouter()
	router.Use(pkgmiddleware.CORSMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	w := performRequest(router, "GET", "/test", map[string]string{
		"Origin": "http://localhost:5173",
	}, "")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "http://localhost:5173", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSMiddleware_DisallowedOrigin(t *testing.T) {
	router := setupTestRouter()
	router.Use(pkgmiddleware.CORSMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	w := performRequest(router, "GET", "/test", map[string]string{
		"Origin": "http://evil-site.com",
	}, "")

	assert.Equal(t, http.StatusOK, w.Code)
	// Disallowed origin should not be reflected
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSMiddleware_PreflightRequest(t *testing.T) {
	router := setupTestRouter()
	router.Use(pkgmiddleware.CORSMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	w := performRequest(router, "OPTIONS", "/test", map[string]string{
		"Origin": "http://localhost:3000",
		"Access-Control-Request-Method": "POST",
	}, "")

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
}

// ==================== Recovery Middleware Tests ====================

func TestRecoveryMiddleware_RecoversFromPanic(t *testing.T) {
	router := setupTestRouter()
	// RecoveryMiddleware in pkg/middleware requires trace_id to be set first
	router.Use(pkgmiddleware.TraceIDMiddleware())
	router.Use(pkgmiddleware.RecoveryMiddleware())
	router.GET("/panic", func(c *gin.Context) {
		panic("something went wrong")
	})

	w := performRequest(router, "GET", "/panic", nil, "")

	// RecoveryMiddleware returns 500 status code
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRecoveryMiddleware_NormalRequest(t *testing.T) {
	router := setupTestRouter()
	router.Use(pkgmiddleware.TraceIDMiddleware())
	router.Use(pkgmiddleware.RecoveryMiddleware())
	router.GET("/normal", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	w := performRequest(router, "GET", "/normal", nil, "")

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRecoveryMiddleware_PanicWithDifferentTypes(t *testing.T) {
	tests := []struct {
		name  string
		panic interface{}
	}{
		{"panic with string", "string panic"},
		{"panic with error", errors.New("error panic")},
		{"panic with int", 42},
		{"panic with struct", struct{ msg string }{msg: "struct panic"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter()
			router.Use(pkgmiddleware.TraceIDMiddleware())
			router.Use(pkgmiddleware.RecoveryMiddleware())
			router.GET("/panic", func(c *gin.Context) {
				panic(tt.panic)
			})

			w := performRequest(router, "GET", "/panic", nil, "")

			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})
	}
}

// ==================== RequestID Middleware Tests ====================

func TestTraceIDMiddleware_GeneratesNewID(t *testing.T) {
	router := setupTestRouter()
	router.Use(pkgmiddleware.TraceIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		traceID, exists := c.Get("trace_id")
		response.Success(c, gin.H{
			"trace_id_exists": exists,
			"trace_id":        traceID,
		})
	})

	w := performRequest(router, "GET", "/test", nil, "")

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseResponse(w)
	data := getDataFromResponse(resp)
	assert.NotNil(t, data)
	traceID := data["trace_id"].(string)
	assert.NotEmpty(t, traceID)
	assert.Len(t, traceID, 36) // UUID length

	// Check response header
	responseTraceID := w.Header().Get("X-Request-ID")
	assert.Equal(t, traceID, responseTraceID)
}

func TestTraceIDMiddleware_UsesExistingID(t *testing.T) {
	router := setupTestRouter()
	router.Use(pkgmiddleware.TraceIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		traceID, _ := c.Get("trace_id")
		response.Success(c, gin.H{"trace_id": traceID})
	})

	existingTraceID := "custom-trace-id-12345"
	w := performRequest(router, "GET", "/test", map[string]string{
		"X-Request-ID": existingTraceID,
	}, "")

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseResponse(w)
	data := getDataFromResponse(resp)
	assert.NotNil(t, data)
	assert.Equal(t, existingTraceID, data["trace_id"])
	assert.Equal(t, existingTraceID, w.Header().Get("X-Request-ID"))
}

func TestTraceIDMiddleware_ContextValue(t *testing.T) {
	router := setupTestRouter()
	router.Use(pkgmiddleware.TraceIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		// Get trace ID from request context
		traceID := pkgmiddleware.GetTraceID(c.Request.Context())
		response.Success(c, gin.H{"trace_id": traceID})
	})

	w := performRequest(router, "GET", "/test", nil, "")

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseResponse(w)
	data := getDataFromResponse(resp)
	assert.NotNil(t, data)
	assert.NotEmpty(t, data["trace_id"])
}

// ==================== Logging Middleware Tests ====================

func TestRequestLoggerMiddleware(t *testing.T) {
	router := setupTestRouter()
	router.Use(ifmiddleware.RequestLoggerMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	w := performRequest(router, "GET", "/test", map[string]string{
		"X-Forwarded-For": "192.168.1.1",
	}, "")

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequestLoggerMiddleware_WithQueryParams(t *testing.T) {
	router := setupTestRouter()
	router.Use(ifmiddleware.RequestLoggerMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	w := performRequest(router, "GET", "/test?foo=bar&baz=qux", nil, "")

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequestLoggerMiddleware_ErrorStatus(t *testing.T) {
	router := setupTestRouter()
	router.Use(ifmiddleware.RequestLoggerMiddleware())
	router.GET("/error", func(c *gin.Context) {
		c.String(http.StatusInternalServerError, "error")
	})

	w := performRequest(router, "GET", "/error", nil, "")

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestLoggingMiddleware(t *testing.T) {
	router := setupTestRouter()
	// LoggingMiddleware requires trace_id to be set first
	router.Use(pkgmiddleware.TraceIDMiddleware())
	router.Use(pkgmiddleware.LoggingMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	w := performRequest(router, "GET", "/test", nil, "")

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLoggingMiddleware_WithRequestBody(t *testing.T) {
	router := setupTestRouter()
	router.Use(pkgmiddleware.TraceIDMiddleware())
	router.Use(pkgmiddleware.LoggingMiddleware())
	router.POST("/test", func(c *gin.Context) {
		c.String(http.StatusCreated, "created")
	})

	body := `{"name":"test","value":123}`
	w := performRequest(router, "POST", "/test", nil, body)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestLoggingMiddleware_SensitivePath(t *testing.T) {
	router := setupTestRouter()
	router.Use(pkgmiddleware.TraceIDMiddleware())
	router.Use(pkgmiddleware.LoggingMiddleware())
	router.POST("/api/v1/auth/login", func(c *gin.Context) {
		c.String(http.StatusOK, "token")
	})

	body := `{"username":"admin","password":"secret123"}`
	w := performRequest(router, "POST", "/api/v1/auth/login", nil, body)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ==================== Security Headers Middleware Tests ====================

func TestSecurityHeadersMiddleware(t *testing.T) {
	router := setupTestRouter()
	router.Use(pkgmiddleware.SecurityHeadersMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	w := performRequest(router, "GET", "/test", nil, "")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Contains(t, w.Header().Get("Content-Security-Policy"), "default-src")
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
}

// ==================== Rate Limit Middleware Tests ====================

func TestRateLimitMiddleware(t *testing.T) {
	router := setupTestRouter()
	router.Use(pkgmiddleware.RateLimitMiddleware(10, 2))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	// First request should succeed
	w := performRequest(router, "GET", "/test", nil, "")
	assert.Equal(t, http.StatusOK, w.Code)

	// Second request should also succeed (within burst)
	w = performRequest(router, "GET", "/test", nil, "")
	assert.Equal(t, http.StatusOK, w.Code)
}

// ==================== Error Handler Middleware Tests ====================

func TestErrorHandlerMiddleware(t *testing.T) {
	router := setupTestRouter()
	router.Use(pkgmiddleware.ErrorHandlerMiddleware())
	router.GET("/error", func(c *gin.Context) {
		// Use an AppError to get proper status code handling
		c.Error(pkgerrors.New(pkgerrors.CodeInvalidParam, "test error"))
	})

	w := performRequest(router, "GET", "/error", nil, "")

	// ErrorHandlerMiddleware returns the appropriate status code
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ==================== Request Size Middleware Tests ====================

func TestRequestSizeMiddleware_AllowedSize(t *testing.T) {
	router := setupTestRouter()
	router.Use(pkgmiddleware.RequestSizeMiddleware(1024)) // 1KB limit
	router.POST("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	body := `{"data":"small payload"}`
	w := performRequest(router, "POST", "/test", nil, body)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequestSizeMiddleware_ExceededSize(t *testing.T) {
	router := setupTestRouter()
	router.Use(pkgmiddleware.RequestSizeMiddleware(10)) // 10 bytes limit
	router.POST("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	body := `{"data":"this is a large payload"}`
	w := performRequest(router, "POST", "/test", map[string]string{
		"Content-Length": "100",
	}, body)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ==================== User Rate Limit Middleware Tests ====================

func TestUserRateLimitMiddleware_WithUserID(t *testing.T) {
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Next()
	})
	router.Use(pkgmiddleware.UserRateLimitMiddleware(10, 2))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	w := performRequest(router, "GET", "/test", nil, "")
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserRateLimitMiddleware_WithoutUserID(t *testing.T) {
	router := setupTestRouter()
	router.Use(pkgmiddleware.UserRateLimitMiddleware(10, 2))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	w := performRequest(router, "GET", "/test", nil, "")
	assert.Equal(t, http.StatusOK, w.Code)
}

// ==================== Audit Middleware Tests ====================

func TestAuditMiddleware_Audit(t *testing.T) {
	mockAuditRepo := testutil.NewMockAuditLogRepository()
	auditMiddleware := ifmiddleware.NewAuditMiddleware(mockAuditRepo)

	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("userID", "user-123")
		c.Set("userRole", entity.RoleAdmin)
		c.Set("trace_id", "trace-123")
		c.Next()
	})
	router.Use(auditMiddleware.Audit(ifmiddleware.AuditOptions{
		Module: "test",
		Action: "test_action",
		Type:   entity.AuditLogTypeQuery,
	}))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	w := performRequest(router, "GET", "/test", nil, "")

	assert.Equal(t, http.StatusOK, w.Code)
	// Wait for async log saving
	time.Sleep(100 * time.Millisecond)
}

func TestAuditMiddleware_AuditWithBody(t *testing.T) {
	mockAuditRepo := testutil.NewMockAuditLogRepository()
	auditMiddleware := ifmiddleware.NewAuditMiddleware(mockAuditRepo)

	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("userID", "user-123")
		c.Set("userRole", entity.RoleAdmin)
		c.Next()
	})
	router.Use(auditMiddleware.Audit(ifmiddleware.AuditOptions{
		Module:     "test",
		Action:     "create",
		Type:       entity.AuditLogTypeCreate,
		IgnoreBody: false,
	}))
	router.POST("/test", func(c *gin.Context) {
		c.String(http.StatusCreated, "created")
	})

	body := `{"name":"test","password":"secret123"}`
	w := performRequest(router, "POST", "/test", nil, body)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestAuditMiddleware_AutoAudit(t *testing.T) {
	mockAuditRepo := testutil.NewMockAuditLogRepository()
	auditMiddleware := ifmiddleware.NewAuditMiddleware(mockAuditRepo)

	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("userID", "user-123")
		c.Set("userRole", entity.RoleAdmin)
		c.Set("trace_id", "trace-123")
		c.Next()
	})
	router.Use(auditMiddleware.AutoAudit())
	router.GET("/api/v1/users", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	w := performRequest(router, "GET", "/api/v1/users", nil, "")

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuditMiddleware_SkipAuditPaths(t *testing.T) {
	mockAuditRepo := testutil.NewMockAuditLogRepository()
	auditMiddleware := ifmiddleware.NewAuditMiddleware(mockAuditRepo)

	router := setupTestRouter()
	router.Use(auditMiddleware.AutoAudit())
	router.GET("/api/v1/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := performRequest(router, "GET", "/api/v1/health", nil, "")

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuditMiddleware_FailedRequest(t *testing.T) {
	mockAuditRepo := testutil.NewMockAuditLogRepository()
	auditMiddleware := ifmiddleware.NewAuditMiddleware(mockAuditRepo)

	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("userID", "user-123")
		c.Set("userRole", entity.RoleAdmin)
		c.Next()
	})
	router.Use(auditMiddleware.Audit(ifmiddleware.AuditOptions{
		Module: "test",
		Action: "test_action",
		Type:   entity.AuditLogTypeQuery,
	}))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusBadRequest, "bad request")
	})

	w := performRequest(router, "GET", "/test", nil, "")

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ==================== Integration Tests ====================

func TestMiddlewareChain(t *testing.T) {
	router := setupTestRouter()
	router.Use(pkgmiddleware.TraceIDMiddleware())
	router.Use(pkgmiddleware.RecoveryMiddleware())
	router.Use(ifmiddleware.RequestLoggerMiddleware())
	router.Use(func(c *gin.Context) {
		c.Set("userID", "user-123")
		c.Set("userRole", entity.RoleVolunteer)
		c.Set("orgID", "org-123")
		c.Next()
	})
	router.Use(ifmiddleware.RequireRole(entity.RoleVolunteer))

	router.GET("/protected", func(c *gin.Context) {
		response.Success(c, gin.H{
			"message":   "success",
			"user_id":   ifmiddleware.GetUserID(c),
			"user_role": string(ifmiddleware.GetUserRole(c)),
		})
	})

	w := performRequest(router, "GET", "/protected", nil, "")

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseResponse(w)
	data := getDataFromResponse(resp)
	assert.NotNil(t, data)
	assert.Equal(t, "user-123", data["user_id"])
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestMiddlewareChain_WithPanic(t *testing.T) {
	router := setupTestRouter()
	router.Use(pkgmiddleware.TraceIDMiddleware())
	router.Use(pkgmiddleware.RecoveryMiddleware())
	router.Use(func(c *gin.Context) {
		c.Set("userID", "user-123")
		c.Next()
	})

	router.GET("/panic", func(c *gin.Context) {
		panic("unexpected error")
	})

	w := performRequest(router, "GET", "/panic", nil, "")

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	// Should have trace ID in response
	resp := parseResponse(w)
	assert.Contains(t, resp["message"], "trace_id")
}

func TestMiddleware_MultipleRequests(t *testing.T) {
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("userID", "user-123")
		c.Set("userRole", entity.RoleVolunteer)
		c.Next()
	})
	router.Use(ifmiddleware.RequireRole(entity.RoleVolunteer))

	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, ifmiddleware.GetUserID(c))
	})

	// Multiple requests should work independently
	for i := 0; i < 5; i++ {
		w := performRequest(router, "GET", "/test", nil, "")
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestExtractToken_DifferentFormats(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func(*gin.Context)
		expectedUserID string
	}{
		{
			name: "with userID in context",
			setupContext: func(c *gin.Context) {
				c.Set("userID", "test-user-123")
			},
			expectedUserID: "test-user-123",
		},
		{
			name: "without userID in context",
			setupContext: func(c *gin.Context) {
				// Don't set anything
			},
			expectedUserID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter()
			router.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			router.GET("/test", func(c *gin.Context) {
				userID := ifmiddleware.GetUserID(c)
				response.Success(c, gin.H{"user_id": userID})
			})

			w := performRequest(router, "GET", "/test", nil, "")

			resp := parseResponse(w)
			data := getDataFromResponse(resp)
			assert.NotNil(t, data)
			assert.Equal(t, tt.expectedUserID, data["user_id"])
		})
	}
}

func TestAuthMiddleware_ContextValues(t *testing.T) {
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		// Verify all context values are set
		c.Set("userID", "admin-123")
		c.Set("userRole", entity.RoleAdmin)
		c.Set("orgID", "org-123")
		c.Set("claims", &service.TokenClaims{
			UserID: "admin-123",
			Role:   "admin",
			OrgID:  "org-123",
		})
		c.Next()
	})
	router.GET("/test", func(c *gin.Context) {
		response.Success(c, gin.H{
			"has_user_id":  ifmiddleware.GetUserID(c) != "",
			"has_user_role": ifmiddleware.GetUserRole(c) != "",
			"has_org_id":   ifmiddleware.GetOrgID(c) != "",
			"has_claims":   ifmiddleware.GetClaims(c) != nil,
			"user_id":      ifmiddleware.GetUserID(c),
			"user_role":    string(ifmiddleware.GetUserRole(c)),
			"org_id":       ifmiddleware.GetOrgID(c),
			"claims_role":  ifmiddleware.GetClaims(c).Role,
		})
	})

	w := performRequest(router, "GET", "/test", nil, "")

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseResponse(w)
	data := getDataFromResponse(resp)
	assert.NotNil(t, data)
	assert.True(t, data["has_user_id"].(bool))
	assert.True(t, data["has_user_role"].(bool))
	assert.True(t, data["has_org_id"].(bool))
	assert.True(t, data["has_claims"].(bool))
	assert.Equal(t, "admin-123", data["user_id"])
	assert.Equal(t, "admin", data["user_role"])
	assert.Equal(t, "org-123", data["org_id"])
	assert.Equal(t, "admin", data["claims_role"])
}

// Benchmarks

func BenchmarkCORSMiddleware(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(pkgmiddleware.CORSMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkRecoveryMiddleware(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(pkgmiddleware.RecoveryMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkRBACMiddleware(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userRole", entity.RoleAdmin)
		c.Next()
	})
	router.Use(ifmiddleware.RequireRole(entity.RoleVolunteer))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// Ensure response package is correctly imported and used
func init() {
	// Force import of response package
	_ = response.Response{}
}
