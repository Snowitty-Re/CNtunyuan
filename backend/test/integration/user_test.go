package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Snowitty-Re/CNtunyuan/internal/application/dto"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// setupRouter 设置测试路由
func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func TestUserAPI_CreateUser(t *testing.T) {
	router := setupRouter()

	tests := []struct {
		name       string
		request    dto.CreateUserRequest
		wantStatus int
	}{
		{
			name: "create user successfully",
			request: dto.CreateUserRequest{
				Nickname: "Test User",
				Phone:    "13800138000",
				Password: "password123",
				Role:     "volunteer",
				OrgID:    "org-001",
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "fail with invalid phone",
			request: dto.CreateUserRequest{
				Nickname: "Test User",
				Phone:    "invalid",
				Password: "password123",
				Role:     "volunteer",
				OrgID:    "org-001",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "fail with short password",
			request: dto.CreateUserRequest{
				Nickname: "Test User",
				Phone:    "13800138001",
				Password: "123",
				Role:     "volunteer",
				OrgID:    "org-001",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestUserAPI_Login(t *testing.T) {
	router := setupRouter()

	tests := []struct {
		name       string
		request    dto.LoginRequest
		wantStatus int
	}{
		{
			name: "login successfully",
			request: dto.LoginRequest{
				Phone:    "13800138000",
				Password: "password123",
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "fail with invalid phone",
			request: dto.LoginRequest{
				Phone:    "invalid",
				Password: "password123",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "fail with empty password",
			request: dto.LoginRequest{
				Phone:    "13800138000",
				Password: "",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
