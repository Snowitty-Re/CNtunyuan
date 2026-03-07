package handler

import (
	"github.com/Snowitty-Re/CNtunyuan/internal/infrastructure/websocket"
	"github.com/Snowitty-Re/CNtunyuan/internal/interfaces/http/middleware"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"github.com/Snowitty-Re/CNtunyuan/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// WebSocketHandler WebSocket处理器
type WebSocketHandler struct {
	wsManager *websocket.Manager
}

// NewWebSocketHandler 创建WebSocket处理器
func NewWebSocketHandler(wsManager *websocket.Manager) *WebSocketHandler {
	return &WebSocketHandler{
		wsManager: wsManager,
	}
}

// RegisterRoutes 注册路由
func (h *WebSocketHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	// WebSocket 连接端点
	r.GET("/ws", authMiddleware.Required(), h.HandleWebSocket)
	
	// 公开的健康检查端点（用于检测WebSocket服务状态）
	r.GET("/ws/health", h.HealthCheck)
}

// HandleWebSocket 处理WebSocket连接
// @Summary WebSocket连接
// @Description 建立WebSocket实时连接，用于接收实时通知
// @Tags websocket
// @Success 101 {string} string "Switching Protocols"
// @Router /api/v1/ws [get]
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.Unauthorized(c, "请先登录")
		return
	}
	
	orgID, _ := c.Get("orgID")
	orgIDStr, _ := orgID.(string)
	
	logger.Info("WebSocket connection request", zap.String("user_id", userID), zap.String("org_id", orgIDStr))
	
	// 升级HTTP连接为WebSocket
	h.wsManager.HandleWebSocket(c.Writer, c.Request, userID, orgIDStr)
}

// HealthCheck WebSocket健康检查
// @Summary WebSocket服务健康检查
// @Description 检查WebSocket服务是否正常运行
// @Tags websocket
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/v1/ws/health [get]
func (h *WebSocketHandler) HealthCheck(c *gin.Context) {
	stats := h.wsManager.GetStats()
	
	response.Success(c, gin.H{
		"status":             "UP",
		"active_connections": stats.ActiveConnections,
		"online_users":       h.wsManager.GetOnlineCount(),
	})
}
