package handler

import (
	"github.com/Snowitty-Re/CNtunyuan/internal/application/dto"
	"github.com/Snowitty-Re/CNtunyuan/internal/application/service"
	"github.com/Snowitty-Re/CNtunyuan/internal/interfaces/http/middleware"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"github.com/Snowitty-Re/CNtunyuan/pkg/response"
	"github.com/gin-gonic/gin"
)

// OrganizationHandler 组织处理器
type OrganizationHandler struct {
	orgService *service.OrganizationAppService
}

// NewOrganizationHandler 创建组织处理器
func NewOrganizationHandler(orgService *service.OrganizationAppService) *OrganizationHandler {
	return &OrganizationHandler{orgService: orgService}
}

// RegisterRoutes 注册路由
func (h *OrganizationHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	orgs := router.Group("/organizations")
	orgs.Use(authMiddleware.Required())
	{
		orgs.GET("", h.List)
		orgs.GET("/tree", h.GetTree)
		orgs.GET("/:id", h.GetByID)
		orgs.GET("/:id/children", h.GetChildren)
		orgs.GET("/:id/path", h.GetPath)
		orgs.POST("", middleware.RequireAdmin(), h.Create)
		orgs.PUT("/:id", middleware.RequireAdmin(), h.Update)
		orgs.DELETE("/:id", middleware.RequireAdmin(), h.Delete)
		orgs.PUT("/:id/move", middleware.RequireAdmin(), h.Move)
	}
}

// Create 创建组织
func (h *OrganizationHandler) Create(c *gin.Context) {
	var req dto.CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	org, err := h.orgService.Create(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case service.ErrOrganizationExists:
			response.Conflict(c, "organization code already exists")
		case service.ErrInvalidOrgType:
			response.BadRequest(c, "invalid organization type")
		default:
			logger.Error("Failed to create organization", logger.Err(err))
			response.InternalServerError(c, "failed to create organization")
		}
		return
	}

	response.Created(c, org)
}

// GetByID 获取组织详情
func (h *OrganizationHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "organization id is required")
		return
	}

	org, err := h.orgService.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrOrganizationNotFound {
			response.NotFound(c, "organization not found")
			return
		}
		response.InternalServerError(c, "failed to get organization")
		return
	}

	response.Success(c, org)
}

// List 组织列表
func (h *OrganizationHandler) List(c *gin.Context) {
	var req dto.OrganizationListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	orgs, err := h.orgService.List(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "failed to get organization list")
		return
	}

	response.Success(c, orgs)
}

// Update 更新组织
func (h *OrganizationHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "organization id is required")
		return
	}

	var req dto.UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	org, err := h.orgService.Update(c.Request.Context(), id, &req)
	if err != nil {
		switch err {
		case service.ErrOrganizationNotFound:
			response.NotFound(c, "organization not found")
		case service.ErrOrganizationExists:
			response.Conflict(c, "organization code already exists")
		default:
			logger.Error("Failed to update organization", logger.Err(err))
			response.InternalServerError(c, "failed to update organization")
		}
		return
	}

	response.Success(c, org)
}

// Delete 删除组织
func (h *OrganizationHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "organization id is required")
		return
	}

	if err := h.orgService.Delete(c.Request.Context(), id); err != nil {
		switch err {
		case service.ErrOrganizationNotFound:
			response.NotFound(c, "organization not found")
		case service.ErrCannotDeleteOrg:
			response.BadRequest(c, "cannot delete organization with children")
		default:
			logger.Error("Failed to delete organization", logger.Err(err))
			response.InternalServerError(c, "failed to delete organization")
		}
		return
	}

	response.NoContent(c)
}

// GetTree 获取组织树
func (h *OrganizationHandler) GetTree(c *gin.Context) {
	rootID := c.Query("root_id")

	tree, err := h.orgService.GetTree(c.Request.Context(), rootID)
	if err != nil {
		response.InternalServerError(c, "failed to get organization tree")
		return
	}

	response.Success(c, tree)
}

// GetChildren 获取子组织
func (h *OrganizationHandler) GetChildren(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "organization id is required")
		return
	}

	children, err := h.orgService.GetChildren(c.Request.Context(), id)
	if err != nil {
		response.InternalServerError(c, "failed to get children")
		return
	}

	response.Success(c, children)
}

// GetPath 获取组织路径
func (h *OrganizationHandler) GetPath(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "organization id is required")
		return
	}

	path, err := h.orgService.GetPath(c.Request.Context(), id)
	if err != nil {
		response.InternalServerError(c, "failed to get path")
		return
	}

	response.Success(c, path)
}

// Move 移动组织
func (h *OrganizationHandler) Move(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "organization id is required")
		return
	}

	var req dto.MoveOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.orgService.Move(c.Request.Context(), id, req.NewParentID); err != nil {
		switch err {
		case service.ErrOrganizationNotFound:
			response.NotFound(c, "organization not found")
		default:
			response.InternalServerError(c, "failed to move organization")
		}
		return
	}

	response.Success(c, nil)
}
