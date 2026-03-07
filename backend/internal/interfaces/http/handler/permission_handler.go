package handler

import (
	"github.com/Snowitty-Re/CNtunyuan/internal/application/dto"
	"github.com/Snowitty-Re/CNtunyuan/internal/application/service"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/interfaces/http/middleware"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"github.com/Snowitty-Re/CNtunyuan/pkg/response"
	"github.com/gin-gonic/gin"
)

// PermissionHandler 权限处理器
type PermissionHandler struct {
	permService *service.PermissionAppService
}

// NewPermissionHandler 创建权限处理器
func NewPermissionHandler(permService *service.PermissionAppService) *PermissionHandler {
	return &PermissionHandler{
		permService: permService,
	}
}

// RegisterRoutes 注册路由
func (h *PermissionHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	// 权限管理（系统管理员）
	perms := router.Group("/permissions")
	perms.Use(authMiddleware.Required())
	{
		perms.GET("", h.ListPermissions)
		perms.POST("", middleware.RequireSuperAdmin(), h.CreatePermission)
		perms.GET("/:id", h.GetPermission)
		perms.PUT("/:id", middleware.RequireSuperAdmin(), h.UpdatePermission)
		perms.DELETE("/:id", middleware.RequireSuperAdmin(), h.DeletePermission)
	}
	
	// 角色管理
	roles := router.Group("/roles")
	roles.Use(authMiddleware.Required())
	{
		roles.GET("", h.ListRoles)
		roles.POST("", middleware.RequireManager(), h.CreateRole)
		roles.GET("/:id", h.GetRole)
		roles.PUT("/:id", middleware.RequireManager(), h.UpdateRole)
		roles.DELETE("/:id", middleware.RequireManager(), h.DeleteRole)
		
		// 角色权限
		roles.GET("/:id/permissions", h.GetRolePermissions)
		roles.POST("/:id/permissions", middleware.RequireManager(), h.GrantPermissions)
		roles.DELETE("/:id/permissions", middleware.RequireManager(), h.RevokePermissions)
	}
	
	// 用户角色
	users := router.Group("/users")
	users.Use(authMiddleware.Required())
	{
		users.GET("/:id/roles", h.GetUserRoles)
		users.POST("/:id/roles", middleware.RequireManager(), h.AssignRole)
		users.DELETE("/:id/roles/:role_id", middleware.RequireManager(), h.RemoveRole)
	}
	
	// 字段权限
	fieldPerms := router.Group("/field-permissions")
	fieldPerms.Use(authMiddleware.Required(), middleware.RequireManager())
	{
		fieldPerms.GET("", h.GetFieldPermissions)
		fieldPerms.POST("", h.SetFieldPermission)
	}
	
	// 权限检查
	router.GET("/check-permission", authMiddleware.Required(), h.CheckPermission)
	
	// 初始化
	router.POST("/init-permissions", authMiddleware.Required(), middleware.RequireSuperAdmin(), h.InitPermissions)
}

// CreatePermission 创建权限
func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	var req dto.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	
	resp, err := h.permService.CreatePermission(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to create permission", logger.Err(err))
		response.InternalServerError(c, err.Error())
		return
	}
	
	response.Created(c, resp)
}

// UpdatePermission 更新权限
func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "id is required")
		return
	}
	
	var req dto.UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	
	resp, err := h.permService.UpdatePermission(c.Request.Context(), id, &req)
	if err != nil {
		logger.Error("Failed to update permission", logger.Err(err))
		if err == service.ErrPermissionNotFound {
			response.NotFound(c, "permission not found")
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}
	
	response.Success(c, resp)
}

// DeletePermission 删除权限
func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "id is required")
		return
	}
	
	if err := h.permService.DeletePermission(c.Request.Context(), id); err != nil {
		logger.Error("Failed to delete permission", logger.Err(err))
		if err == service.ErrPermissionNotFound {
			response.NotFound(c, "permission not found")
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}
	
	response.NoContent(c)
}

// GetPermission 获取权限
func (h *PermissionHandler) GetPermission(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "id is required")
		return
	}
	
	resp, err := h.permService.GetPermission(c.Request.Context(), id)
	if err != nil {
		logger.Error("Failed to get permission", logger.Err(err))
		if err == service.ErrPermissionNotFound {
			response.NotFound(c, "permission not found")
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}
	
	response.Success(c, resp)
}

// ListPermissions 列表查询权限
func (h *PermissionHandler) ListPermissions(c *gin.Context) {
	var req dto.ListPermissionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	
	resp, err := h.permService.ListPermissions(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to list permissions", logger.Err(err))
		response.InternalServerError(c, err.Error())
		return
	}
	
	response.Success(c, resp)
}

// CreateRole 创建角色
func (h *PermissionHandler) CreateRole(c *gin.Context) {
	var req dto.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	
	req.OrgID = middleware.GetOrgID(c)
	req.CreatedBy = middleware.GetUserID(c)
	
	resp, err := h.permService.CreateRole(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to create role", logger.Err(err))
		if err == service.ErrRoleExists {
			response.Conflict(c, "role already exists")
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}
	
	response.Created(c, resp)
}

// UpdateRole 更新角色
func (h *PermissionHandler) UpdateRole(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "id is required")
		return
	}
	
	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	
	resp, err := h.permService.UpdateRole(c.Request.Context(), id, &req)
	if err != nil {
		logger.Error("Failed to update role", logger.Err(err))
		if err == service.ErrRoleNotFound {
			response.NotFound(c, "role not found")
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}
	
	response.Success(c, resp)
}

// DeleteRole 删除角色
func (h *PermissionHandler) DeleteRole(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "id is required")
		return
	}
	
	if err := h.permService.DeleteRole(c.Request.Context(), id); err != nil {
		logger.Error("Failed to delete role", logger.Err(err))
		if err == service.ErrRoleNotFound {
			response.NotFound(c, "role not found")
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}
	
	response.NoContent(c)
}

// GetRole 获取角色
func (h *PermissionHandler) GetRole(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "id is required")
		return
	}
	
	resp, err := h.permService.GetRole(c.Request.Context(), id)
	if err != nil {
		logger.Error("Failed to get role", logger.Err(err))
		if err == service.ErrRoleNotFound {
			response.NotFound(c, "role not found")
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}
	
	response.Success(c, resp)
}

// ListRoles 列表查询角色
func (h *PermissionHandler) ListRoles(c *gin.Context) {
	var req dto.ListRolesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	
	req.OrgID = middleware.GetOrgID(c)
	
	resp, err := h.permService.ListRoles(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to list roles", logger.Err(err))
		response.InternalServerError(c, err.Error())
		return
	}
	
	response.Success(c, resp)
}

// GrantPermissions 授予权限
func (h *PermissionHandler) GrantPermissions(c *gin.Context) {
	roleID := c.Param("id")
	if roleID == "" {
		response.BadRequest(c, "id is required")
		return
	}
	
	var req dto.GrantPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	
	req.GrantedBy = middleware.GetUserID(c)
	
	if err := h.permService.GrantPermissions(c.Request.Context(), roleID, &req); err != nil {
		logger.Error("Failed to grant permissions", logger.Err(err))
		response.InternalServerError(c, err.Error())
		return
	}
	
	response.Success(c, nil)
}

// RevokePermissions 撤销权限
func (h *PermissionHandler) RevokePermissions(c *gin.Context) {
	roleID := c.Param("id")
	if roleID == "" {
		response.BadRequest(c, "id is required")
		return
	}
	
	var req dto.RevokePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	
	if err := h.permService.RevokePermissions(c.Request.Context(), roleID, &req); err != nil {
		logger.Error("Failed to revoke permissions", logger.Err(err))
		response.InternalServerError(c, err.Error())
		return
	}
	
	response.Success(c, nil)
}

// GetRolePermissions 获取角色权限
func (h *PermissionHandler) GetRolePermissions(c *gin.Context) {
	roleID := c.Param("id")
	if roleID == "" {
		response.BadRequest(c, "id is required")
		return
	}
	
	resp, err := h.permService.GetRolePermissions(c.Request.Context(), roleID)
	if err != nil {
		logger.Error("Failed to get role permissions", logger.Err(err))
		response.InternalServerError(c, err.Error())
		return
	}
	
	response.Success(c, resp)
}

// AssignRole 分配角色给用户
func (h *PermissionHandler) AssignRole(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "id is required")
		return
	}
	
	var req dto.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	
	req.OrgID = middleware.GetOrgID(c)
	req.AssignedBy = middleware.GetUserID(c)
	
	if err := h.permService.AssignRole(c.Request.Context(), userID, &req); err != nil {
		logger.Error("Failed to assign role", logger.Err(err))
		response.InternalServerError(c, err.Error())
		return
	}
	
	response.Success(c, nil)
}

// RemoveRole 移除用户角色
func (h *PermissionHandler) RemoveRole(c *gin.Context) {
	userID := c.Param("id")
	roleID := c.Param("role_id")
	if userID == "" || roleID == "" {
		response.BadRequest(c, "id and role_id are required")
		return
	}
	
	if err := h.permService.RemoveRole(c.Request.Context(), userID, roleID); err != nil {
		logger.Error("Failed to remove role", logger.Err(err))
		response.InternalServerError(c, err.Error())
		return
	}
	
	response.Success(c, nil)
}

// GetUserRoles 获取用户角色
func (h *PermissionHandler) GetUserRoles(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "id is required")
		return
	}
	
	resp, err := h.permService.GetUserRoles(c.Request.Context(), userID)
	if err != nil {
		logger.Error("Failed to get user roles", logger.Err(err))
		response.InternalServerError(c, err.Error())
		return
	}
	
	response.Success(c, resp)
}

// SetFieldPermission 设置字段权限
func (h *PermissionHandler) SetFieldPermission(c *gin.Context) {
	var req dto.SetFieldPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	
	if err := h.permService.SetFieldPermission(c.Request.Context(), &req); err != nil {
		logger.Error("Failed to set field permission", logger.Err(err))
		response.InternalServerError(c, err.Error())
		return
	}
	
	response.Success(c, nil)
}

// GetFieldPermissions 获取字段权限
func (h *PermissionHandler) GetFieldPermissions(c *gin.Context) {
	resource := c.Query("resource")
	roleID := c.Query("role_id")
	
	resp, err := h.permService.GetFieldPermissions(c.Request.Context(), resource, roleID)
	if err != nil {
		logger.Error("Failed to get field permissions", logger.Err(err))
		response.InternalServerError(c, err.Error())
		return
	}
	
	response.Success(c, resp)
}

// CheckPermission 检查权限
func (h *PermissionHandler) CheckPermission(c *gin.Context) {
	userID := middleware.GetUserID(c)
	resource := c.Query("resource")
	action := c.Query("action")
	
	if resource == "" || action == "" {
		response.BadRequest(c, "resource and action are required")
		return
	}
	
	result, err := h.permService.CheckPermission(c.Request.Context(), userID, entity.PermissionResource(resource), entity.PermissionAction(action))
	if err != nil {
		logger.Error("Failed to check permission", logger.Err(err))
		response.InternalServerError(c, err.Error())
		return
	}
	
	resp := dto.PermissionCheckResponse{
		Allowed:   result.Allowed,
		Reason:    result.Reason,
		DataScope: string(result.DataScope),
	}
	
	if !result.Allowed {
		resp.MissingPerm = result.MissingPerm
	}
	
	response.Success(c, resp)
}

// InitPermissions 初始化系统权限
func (h *PermissionHandler) InitPermissions(c *gin.Context) {
	if err := h.permService.InitSystemPermissions(c.Request.Context()); err != nil {
		logger.Error("Failed to init permissions", logger.Err(err))
		response.InternalServerError(c, err.Error())
		return
	}
	
	orgID := middleware.GetOrgID(c)
	if err := h.permService.InitSystemRoles(c.Request.Context(), orgID); err != nil {
		logger.Error("Failed to init roles", logger.Err(err))
		response.InternalServerError(c, err.Error())
		return
	}
	
	response.Success(c, gin.H{
		"message": "permissions and roles initialized successfully",
	})
}
