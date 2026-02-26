package api

import (
	"github.com/Snowitty-Re/CNtunyuan/internal/service"
	"github.com/Snowitty-Re/CNtunyuan/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// OrgHandler 组织处理器
type OrgHandler struct {
	orgService *service.OrganizationService
}

// NewOrgHandler 创建组织处理器
func NewOrgHandler(orgService *service.OrganizationService) *OrgHandler {
	return &OrgHandler{orgService: orgService}
}

// CreateOrg 创建组织
// @Summary 创建组织
// @Description 创建新的组织架构节点
// @Tags 组织管理
// @Accept json
// @Produce json
// @Param body body service.CreateOrgRequest true "组织信息"
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response
// @Router /organizations [post]
func (h *OrgHandler) CreateOrg(c *gin.Context) {
	var req service.CreateOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	org, err := h.orgService.Create(c.Request.Context(), &req)
	if err != nil {
		utils.ServerError(c, "创建组织失败: "+err.Error())
		return
	}

	utils.Created(c, org)
}

// GetOrg 获取组织详情
// @Summary 获取组织详情
// @Description 根据ID获取组织详细信息
// @Tags 组织管理
// @Produce json
// @Param id path string true "组织ID"
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response
// @Router /organizations/{id} [get]
func (h *OrgHandler) GetOrg(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "组织ID格式错误")
		return
	}

	org, err := h.orgService.GetByID(c.Request.Context(), id)
	if err != nil {
		utils.NotFound(c, "组织不存在")
		return
	}

	utils.Success(c, org)
}

// UpdateOrg 更新组织
// @Summary 更新组织
// @Description 更新组织信息
// @Tags 组织管理
// @Accept json
// @Produce json
// @Param id path string true "组织ID"
// @Param body body service.UpdateOrgRequest true "组织信息"
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response
// @Router /organizations/{id} [put]
func (h *OrgHandler) UpdateOrg(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "组织ID格式错误")
		return
	}

	var req service.UpdateOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	org, err := h.orgService.Update(c.Request.Context(), id, &req)
	if err != nil {
		utils.ServerError(c, "更新组织失败: "+err.Error())
		return
	}

	utils.Success(c, org)
}

// DeleteOrg 删除组织
// @Summary 删除组织
// @Description 删除指定组织
// @Tags 组织管理
// @Produce json
// @Param id path string true "组织ID"
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response
// @Router /organizations/{id} [delete]
func (h *OrgHandler) DeleteOrg(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "组织ID格式错误")
		return
	}

	if err := h.orgService.Delete(c.Request.Context(), id); err != nil {
		utils.ServerError(c, "删除组织失败: "+err.Error())
		return
	}

	utils.Success(c, nil)
}

// ListOrgs 获取组织列表
// @Summary 获取组织列表
// @Description 分页获取组织列表
// @Tags 组织管理
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param type query string false "类型"
// @Param parent_id query string false "父级ID"
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response
// @Router /organizations [get]
func (h *OrgHandler) ListOrgs(c *gin.Context) {
	page := 1
	pageSize := 20
	c.ShouldBindQuery(&struct {
		Page     int `form:"page,default=1"`
		PageSize int `form:"page_size,default=20"`
	}{})

	filters := map[string]interface{}{}
	if orgType := c.Query("type"); orgType != "" {
		filters["type"] = orgType
	}
	if parentID := c.Query("parent_id"); parentID != "" {
		if pid, err := uuid.Parse(parentID); err == nil {
			filters["parent_id"] = pid
		}
	}

	orgs, total, err := h.orgService.List(c.Request.Context(), page, pageSize, filters)
	if err != nil {
		utils.ServerError(c, "获取组织列表失败")
		return
	}

	utils.PageSuccess(c, orgs, total, page, pageSize)
}

// GetOrgTree 获取组织架构树
// @Summary 获取组织架构树
// @Description 获取组织架构树形结构
// @Tags 组织管理
// @Produce json
// @Param parent_id query string false "父级ID"
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response
// @Router /organizations/tree [get]
func (h *OrgHandler) GetOrgTree(c *gin.Context) {
	var parentID *uuid.UUID
	if pid := c.Query("parent_id"); pid != "" {
		if id, err := uuid.Parse(pid); err == nil {
			parentID = &id
		}
	}

	tree, err := h.orgService.GetTree(c.Request.Context(), parentID)
	if err != nil {
		utils.ServerError(c, "获取组织架构树失败")
		return
	}

	utils.Success(c, tree)
}
