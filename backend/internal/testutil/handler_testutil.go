// Package testutil 提供测试工具
package testutil

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// CreateTestRequest 创建测试请求
func CreateTestRequest(t *testing.T, method, path string, body interface{}) (*httptest.ResponseRecorder, *http.Request) {
	gin.SetMode(gin.TestMode)
	
	var reqBody []byte
	if body != nil {
		var err error
		reqBody, err = json.Marshal(body)
		require.NoError(t, err)
	}
	
	req := httptest.NewRequest(method, path, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	return w, req
}

// ExecuteRequest 执行请求并返回响应
func ExecuteRequest(router *gin.Engine, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}
