package handler

import (
	"github.com/Snowitty-Re/CNtunyuan/internal/application/dto"
	"github.com/Snowitty-Re/CNtunyuan/internal/application/service"
	"github.com/Snowitty-Re/CNtunyuan/internal/interfaces/http/middleware"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"github.com/Snowitty-Re/CNtunyuan/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// NotificationHandler 通知处理器
type NotificationHandler struct {
	notifService *service.NotificationAppService
}

// NewNotificationHandler 创建通知处理器
func NewNotificationHandler(notifService *service.NotificationAppService) *NotificationHandler {
	return &NotificationHandler{
		notifService: notifService,
	}
}

// RegisterRoutes 注册路由
func (h *NotificationHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	notifications := r.Group("/notifications")
	notifications.Use(authMiddleware.Required())
	{
		// 通知管理
		notifications.GET("", h.List)
		notifications.GET("/unread", h.GetUnread)
		notifications.GET("/unread/count", h.GetUnreadCount)
		notifications.GET("/stats", h.GetStats)
		notifications.POST("/read", h.MarkAsRead)
		notifications.POST("/read/all", h.MarkAllAsRead)
		notifications.POST("/:id/archive", h.Archive)
		notifications.DELETE("/:id", h.Delete)
		
		// 设置管理
		notifications.GET("/settings", h.GetSettings)
		notifications.PUT("/settings", h.UpdateSettings)
		
		// 模板管理（管理员）
		notifications.GET("/templates", h.ListTemplates)
		notifications.POST("/templates", h.CreateTemplate)
		notifications.GET("/templates/:id", h.GetTemplate)
		notifications.PUT("/templates/:id", h.UpdateTemplate)
		notifications.DELETE("/templates/:id", h.DeleteTemplate)
		notifications.POST("/templates/:id/render", h.RenderTemplate)
		
		// WebSocket 统计（管理员）
		notifications.GET("/websocket/stats", h.GetWebSocketStats)
		notifications.GET("/websocket/online-users", h.GetOnlineUsers)
		
		// 广播（管理员）
		notifications.POST("/broadcast", h.Broadcast)
	}
}

// List 获取通知列表
// @Summary 获取通知列表
// @Description 获取当前用户的通知列表
// @Tags notifications
// @Accept json
// @Produce json
// @Param type query string false "通知类型"
// @Param status query string false "通知状态"
// @Param unread_only query bool false "仅未读"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=dto.NotificationListResponse}
// @Router /api/v1/notifications [get]
func (h *NotificationHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.Unauthorized(c, "请先登录")
		return
	}
	
	var req dto.NotificationListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	
	req.UserID = userID
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}
	
	result, err := h.notifService.GetNotificationList(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to get notifications", zap.Error(err), zap.String("user_id", userID))
		response.Error(c, err)
		return
	}
	
	response.Success(c, result)
}

// GetUnread 获取未读通知
func (h *NotificationHandler) GetUnread(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.Unauthorized(c, "请先登录")
		return
	}
	
	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed := parseInt(l); parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	
	notifications, err := h.notifService.GetUnreadNotifications(c.Request.Context(), userID, limit)
	if err != nil {
		logger.Error("Failed to get unread notifications", zap.Error(err), zap.String("user_id", userID))
		response.Error(c, err)
		return
	}
	
	response.Success(c, notifications)
}

// GetUnreadCount 获取未读数量
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.Unauthorized(c, "请先登录")
		return
	}
	
	count, err := h.notifService.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		logger.Error("Failed to get unread count", zap.Error(err), zap.String("user_id", userID))
		response.Error(c, err)
		return
	}
	
	response.Success(c, dto.UnreadCountResponse{Count: count})
}

// GetStats 获取通知统计
func (h *NotificationHandler) GetStats(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.Unauthorized(c, "请先登录")
		return
	}
	
	stats, err := h.notifService.GetNotificationStats(c.Request.Context(), userID)
	if err != nil {
		logger.Error("Failed to get notification stats", zap.Error(err), zap.String("user_id", userID))
		response.Error(c, err)
		return
	}
	
	response.Success(c, dto.NotificationStatsResponse{
		TotalCount:    stats.TotalCount,
		UnreadCount:   stats.UnreadCount,
		ReadCount:     stats.ReadCount,
		SystemCount:   stats.SystemCount,
		TaskCount:     stats.TaskCount,
		WorkflowCount: stats.WorkflowCount,
		AlertCount:    stats.AlertCount,
	})
}

// MarkAsRead 标记为已读
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.Unauthorized(c, "请先登录")
		return
	}
	
	var req dto.MarkReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 尝试从URL参数获取
		notifID := c.Param("id")
		if notifID == "" {
			response.BadRequest(c, "通知ID不能为空")
			return
		}
		req.IDs = []string{notifID}
	}
	
	markedCount := 0
	for _, id := range req.IDs {
		if err := h.notifService.MarkAsRead(c.Request.Context(), id); err == nil {
			markedCount++
		}
	}
	
	response.Success(c, dto.MarkReadResponse{MarkedCount: markedCount})
}

// MarkAllAsRead 标记所有为已读
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.Unauthorized(c, "请先登录")
		return
	}
	
	if err := h.notifService.MarkAllAsRead(c.Request.Context(), userID); err != nil {
		logger.Error("Failed to mark all as read", zap.Error(err), zap.String("user_id", userID))
		response.Error(c, err)
		return
	}
	
	response.Success(c, gin.H{"message": "全部标记为已读"})
}

// Archive 归档通知
func (h *NotificationHandler) Archive(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.Unauthorized(c, "请先登录")
		return
	}
	
	notifID := c.Param("id")
	if notifID == "" {
		response.BadRequest(c, "通知ID不能为空")
		return
	}
	
	if err := h.notifService.MarkAsArchived(c.Request.Context(), notifID); err != nil {
		logger.Error("Failed to archive notification", zap.Error(err), zap.String("notification_id", notifID))
		response.Error(c, err)
		return
	}
	
	response.Success(c, gin.H{"message": "归档成功"})
}

// Delete 删除通知
func (h *NotificationHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.Unauthorized(c, "请先登录")
		return
	}
	
	notifID := c.Param("id")
	if notifID == "" {
		response.BadRequest(c, "通知ID不能为空")
		return
	}
	
	if err := h.notifService.DeleteNotification(c.Request.Context(), notifID); err != nil {
		logger.Error("Failed to delete notification", zap.Error(err), zap.String("notification_id", notifID))
		response.Error(c, err)
		return
	}
	
	response.Success(c, gin.H{"message": "删除成功"})
}

// GetSettings 获取通知设置
func (h *NotificationHandler) GetSettings(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.Unauthorized(c, "请先登录")
		return
	}
	
	setting, err := h.notifService.GetUserSettings(c.Request.Context(), userID)
	if err != nil {
		logger.Error("Failed to get notification settings", zap.Error(err), zap.String("user_id", userID))
		response.Error(c, err)
		return
	}
	
	response.Success(c, dto.ToNotificationSettingResponse(setting))
}

// UpdateSettings 更新通知设置
func (h *NotificationHandler) UpdateSettings(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.Unauthorized(c, "请先登录")
		return
	}
	
	var req dto.UpdateNotificationSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	
	if err := h.notifService.UpdateUserSettings(c.Request.Context(), userID, &req); err != nil {
		logger.Error("Failed to update notification settings", zap.Error(err), zap.String("user_id", userID))
		response.Error(c, err)
		return
	}
	
	response.Success(c, gin.H{"message": "设置已更新"})
}

// ListTemplates 获取模板列表
func (h *NotificationHandler) ListTemplates(c *gin.Context) {
	var req dto.MessageTemplateListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}
	
	result, err := h.notifService.ListMessageTemplates(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to list templates", zap.Error(err))
		response.Error(c, err)
		return
	}
	
	response.Success(c, result)
}

// CreateTemplate 创建模板
func (h *NotificationHandler) CreateTemplate(c *gin.Context) {
	var req dto.CreateMessageTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	
	template, err := h.notifService.CreateMessageTemplate(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to create template", zap.Error(err))
		response.Error(c, err)
		return
	}
	
	response.Success(c, dto.ToMessageTemplateResponse(template))
}

// GetTemplate 获取模板详情
func (h *NotificationHandler) GetTemplate(c *gin.Context) {
	templateID := c.Param("id")
	if templateID == "" {
		response.BadRequest(c, "模板ID不能为空")
		return
	}
	
	template, err := h.notifService.GetMessageTemplate(c.Request.Context(), templateID)
	if err != nil {
		logger.Error("Failed to get template", zap.Error(err), zap.String("template_id", templateID))
		response.NotFound(c, "模板不存在")
		return
	}
	
	response.Success(c, dto.ToMessageTemplateResponse(template))
}

// UpdateTemplate 更新模板
func (h *NotificationHandler) UpdateTemplate(c *gin.Context) {
	templateID := c.Param("id")
	if templateID == "" {
		response.BadRequest(c, "模板ID不能为空")
		return
	}
	
	var req dto.UpdateMessageTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	
	template, err := h.notifService.UpdateMessageTemplate(c.Request.Context(), templateID, &req)
	if err != nil {
		logger.Error("Failed to update template", zap.Error(err), zap.String("template_id", templateID))
		response.Error(c, err)
		return
	}
	
	response.Success(c, dto.ToMessageTemplateResponse(template))
}

// DeleteTemplate 删除模板
func (h *NotificationHandler) DeleteTemplate(c *gin.Context) {
	templateID := c.Param("id")
	if templateID == "" {
		response.BadRequest(c, "模板ID不能为空")
		return
	}
	
	if err := h.notifService.DeleteMessageTemplate(c.Request.Context(), templateID); err != nil {
		logger.Error("Failed to delete template", zap.Error(err), zap.String("template_id", templateID))
		response.Error(c, err)
		return
	}
	
	response.Success(c, gin.H{"message": "删除成功"})
}

// RenderTemplate 渲染模板（测试）
func (h *NotificationHandler) RenderTemplate(c *gin.Context) {
	templateID := c.Param("id")
	if templateID == "" {
		response.BadRequest(c, "模板ID不能为空")
		return
	}
	
	var req dto.RenderTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	
	result, err := h.notifService.RenderTemplate(c.Request.Context(), templateID, req.Variables)
	if err != nil {
		logger.Error("Failed to render template", zap.Error(err), zap.String("template_id", templateID))
		response.Error(c, err)
		return
	}
	
	response.Success(c, dto.RenderTemplateResponse{Result: result})
}

// GetWebSocketStats 获取WebSocket统计
func (h *NotificationHandler) GetWebSocketStats(c *gin.Context) {
	stats := h.notifService.GetWebSocketStats()
	onlineUsers := h.notifService.GetOnlineUsers()
	
	response.Success(c, dto.WebSocketStatsResponse{
		TotalConnections:  stats.TotalConnections,
		ActiveConnections: stats.ActiveConnections,
		MessagesSent:      stats.MessagesSent,
		MessagesReceived:  stats.MessagesReceived,
		BroadcastsSent:    stats.BroadcastsSent,
		OnlineUsers:       onlineUsers,
		OnlineCount:       len(onlineUsers),
	})
}

// GetOnlineUsers 获取在线用户
func (h *NotificationHandler) GetOnlineUsers(c *gin.Context) {
	onlineUsers := h.notifService.GetOnlineUsers()
	
	response.Success(c, gin.H{
		"online_users": onlineUsers,
		"online_count": len(onlineUsers),
	})
}

// Broadcast 广播消息
func (h *NotificationHandler) Broadcast(c *gin.Context) {
	var req dto.BroadcastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	
	var err error
	if req.OrgID != "" {
		err = h.notifService.BroadcastToOrg(c.Request.Context(), req.OrgID, req.Title, req.Content, req.Data)
	} else {
		err = h.notifService.Broadcast(c.Request.Context(), req.Title, req.Content, req.Data)
	}
	
	if err != nil {
		logger.Error("Failed to broadcast", zap.Error(err))
		response.Error(c, err)
		return
	}
	
	response.Success(c, gin.H{"message": "广播已发送"})
}

// parseInt 解析整数
func parseInt(s string) int {
	var n int
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			break
		}
		n = n*10 + int(ch-'0')
	}
	return n
}
