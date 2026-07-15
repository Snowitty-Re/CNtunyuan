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
// @Summary      创建组织
// @Description  创建新的组织，需要管理员权限
// @Tags         组织管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      dto.CreateOrganizationRequest  true  "创建组织请求参数"
// @Success      201      {object}  response.Response{data=dto.OrganizationResponse}  "创建成功"
// @Failure      400      {object}  response.Response  "参数错误"
// @Failure      401      {object}  response.Response  "未授权"
// @Failure      403      {object}  response.Response  "禁止访问"
// @Failure      409      {object}  response.Response  "组织代码已存在"
// @Failure      500      {object}  response.Response  "服务器内部错误"
// @Router       /organizations [post]
func (h *OrganizationHandler) Create(c *gin.Context) {
	var req dto.CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	org, err := h.orgService.Create(c.Request.Context(), &req, userOperator(c))
	if err != nil {
		switch err {
		case service.ErrOrganizationExists:
			response.Conflict(c, "organization code already exists")
		case service.ErrInvalidOrgType:
			response.BadRequest(c, "invalid organization type")
		case service.ErrOrganizationForbidden:
			response.Forbidden(c, "no permission to manage this organization")
		default:
			logger.Error("Failed to create organization", logger.Err(err))
			response.InternalServerError(c, "failed to create organization")
		}
		return
	}

	response.Created(c, org)
}

// GetByID 获取组织详情
// @Summary      获取组织详情
// @Description  根据ID获取组织详细信息
// @Tags         组织管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "组织ID"
// @Success      200  {object}  response.Response{data=dto.OrganizationResponse}  "获取成功"
// @Failure      400  {object}  response.Response  "参数错误"
// @Failure      401  {object}  response.Response  "未授权"
// @Failure      404  {object}  response.Response  "组织不存在"
// @Failure      500  {object}  response.Response  "服务器内部错误"
// @Router       /organizations/{id} [get]
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
// @Summary      获取组织列表
// @Description  分页获取组织列表，支持关键字搜索和筛选
// @Tags         组织管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page       query     int     false  "页码，默认1"
// @Param        page_size  query     int     false  "每页数量，默认10，最大100"
// @Param        keyword    query     string  false  "搜索关键字（名称或代码）"
// @Param        type       query     string  false  "组织类型"
// @Param        status     query     string  false  "组织状态"
// @Param        parent_id  query     string  false  "父组织ID"
// @Success      200        {object}  response.Response{data=dto.OrganizationListResponse}  "获取成功"
// @Failure      400        {object}  response.Response  "参数错误"
// @Failure      401        {object}  response.Response  "未授权"
// @Failure      500        {object}  response.Response  "服务器内部错误"
// @Router       /organizations [get]
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
// @Summary      更新组织
// @Description  更新组织信息，需要管理员权限
// @Tags         组织管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string                         true  "组织ID"
// @Param        request  body      dto.UpdateOrganizationRequest  true  "更新组织请求参数"
// @Success      200      {object}  response.Response{data=dto.OrganizationResponse}  "更新成功"
// @Failure      400      {object}  response.Response  "参数错误"
// @Failure      401      {object}  response.Response  "未授权"
// @Failure      403      {object}  response.Response  "禁止访问"
// @Failure      404      {object}  response.Response  "组织不存在"
// @Failure      409      {object}  response.Response  "组织代码已存在"
// @Failure      500      {object}  response.Response  "服务器内部错误"
// @Router       /organizations/{id} [put]
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

	org, err := h.orgService.Update(c.Request.Context(), id, &req, userOperator(c))
	if err != nil {
		switch err {
		case service.ErrOrganizationNotFound:
			response.NotFound(c, "organization not found")
		case service.ErrOrganizationExists:
			response.Conflict(c, "organization code already exists")
		case service.ErrOrganizationForbidden:
			response.Forbidden(c, "no permission to manage this organization")
		default:
			logger.Error("Failed to update organization", logger.Err(err))
			response.InternalServerError(c, "failed to update organization")
		}
		return
	}

	response.Success(c, org)
}

// Delete 删除组织
// @Summary      删除组织
// @Description  删除指定组织，需要管理员权限。如果组织下有子组织则无法删除
// @Tags         组织管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "组织ID"
// @Success      204  "删除成功"
// @Failure      400  {object}  response.Response  "参数错误或组织下有子组织"
// @Failure      401  {object}  response.Response  "未授权"
// @Failure      403  {object}  response.Response  "禁止访问"
// @Failure      404  {object}  response.Response  "组织不存在"
// @Failure      500  {object}  response.Response  "服务器内部错误"
// @Router       /organizations/{id} [delete]
func (h *OrganizationHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "organization id is required")
		return
	}

	if err := h.orgService.Delete(c.Request.Context(), id, userOperator(c)); err != nil {
		switch err {
		case service.ErrOrganizationNotFound:
			response.NotFound(c, "organization not found")
		case service.ErrCannotDeleteOrg:
			response.BadRequest(c, "cannot delete organization with children")
		case service.ErrOrganizationForbidden:
			response.Forbidden(c, "no permission to manage this organization")
		default:
			logger.Error("Failed to delete organization", logger.Err(err))
			response.InternalServerError(c, "failed to delete organization")
		}
		return
	}

	response.NoContent(c)
}

// GetTree 获取组织树
// @Summary      获取组织树
// @Description  获取组织树形结构，可指定根组织ID
// @Tags         组织管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        root_id  query     string  false  "根组织ID，为空则返回整棵树"
// @Success      200      {object}  response.Response{data=dto.OrganizationTreeResponse}  "获取成功"
// @Failure      401      {object}  response.Response  "未授权"
// @Failure      500      {object}  response.Response  "服务器内部错误"
// @Router       /organizations/tree [get]
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
// @Summary      获取子组织列表
// @Description  获取指定组织的直接子组织列表
// @Tags         组织管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "组织ID"
// @Success      200  {object}  response.Response{data=[]dto.OrganizationResponse}  "获取成功"
// @Failure      400  {object}  response.Response  "参数错误"
// @Failure      401  {object}  response.Response  "未授权"
// @Failure      500  {object}  response.Response  "服务器内部错误"
// @Router       /organizations/{id}/children [get]
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
// @Summary      获取组织路径
// @Description  获取从根组织到指定组织的完整路径
// @Tags         组织管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "组织ID"
// @Success      200  {object}  response.Response{data=[]dto.OrganizationResponse}  "获取成功，返回从根到当前组织的路径数组"
// @Failure      400  {object}  response.Response  "参数错误"
// @Failure      401  {object}  response.Response  "未授权"
// @Failure      500  {object}  response.Response  "服务器内部错误"
// @Router       /organizations/{id}/path [get]
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
// @Summary      移动组织
// @Description  将组织移动到新的父组织下，需要管理员权限
// @Tags         组织管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string                      true  "组织ID"
// @Param        request  body      dto.MoveOrganizationRequest true  "移动组织请求参数"
// @Success      200      {object}  response.Response  "移动成功"
// @Failure      400      {object}  response.Response  "参数错误"
// @Failure      401      {object}  response.Response  "未授权"
// @Failure      403      {object}  response.Response  "禁止访问"
// @Failure      404      {object}  response.Response  "组织不存在"
// @Failure      500      {object}  response.Response  "服务器内部错误"
// @Router       /organizations/{id}/move [put]
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

	if err := h.orgService.Move(c.Request.Context(), id, req.NewParentID, userOperator(c)); err != nil {
		switch err {
		case service.ErrOrganizationNotFound:
			response.NotFound(c, "organization not found")
		case service.ErrOrganizationInvalidMove:
			response.BadRequest(c, "cannot move organization to itself")
		case service.ErrOrganizationForbidden:
			response.Forbidden(c, "no permission to manage this organization")
		default:
			response.InternalServerError(c, "failed to move organization")
		}
		return
	}

	response.Success(c, nil)
}
