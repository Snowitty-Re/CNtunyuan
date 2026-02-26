package api

import (
	"github.com/Snowitty-Re/CNtunyuan/internal/service"
	"github.com/Snowitty-Re/CNtunyuan/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetUserRequest 获取用户请求
type GetUserRequest struct {
	ID string `uri:"id" binding:"required"`
}

// GetUser 获取用户详情
// @Summary 获取用户详情
// @Description 根据ID获取用户详细信息
// @Tags 用户管理
// @Produce json
// @Param id path string true "用户ID"
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response{data=model.User}
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	var req GetUserRequest
	if err := c.ShouldBindUri(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	id, err := uuid.Parse(req.ID)
	if err != nil {
		utils.BadRequest(c, "用户ID格式错误")
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), id)
	if err != nil {
		utils.NotFound(c, "用户不存在")
		return
	}

	utils.Success(c, user)
}

// ListUsersRequest 获取用户列表请求
type ListUsersRequest struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
	Role     string `form:"role"`
	Status   string `form:"status"`
	OrgID    string `form:"org_id"`
	Keyword  string `form:"keyword"`
}

// ListUsers 获取用户列表
// @Summary 获取用户列表
// @Description 分页获取用户列表
// @Tags 用户管理
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param role query string false "角色"
// @Param status query string false "状态"
// @Param org_id query string false "组织ID"
// @Param keyword query string false "关键词"
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response{data=utils.PageData}
// @Router /users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	var req ListUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	filters := map[string]interface{}{}
	if req.Role != "" {
		filters["role"] = req.Role
	}
	if req.Status != "" {
		filters["status"] = req.Status
	}
	if req.OrgID != "" {
		if orgID, err := uuid.Parse(req.OrgID); err == nil {
			filters["org_id"] = orgID
		}
	}

	users, total, err := h.userService.List(c.Request.Context(), req.Page, req.PageSize, filters)
	if err != nil {
		utils.ServerError(c, "获取用户列表失败")
		return
	}

	utils.PageSuccess(c, users, total, req.Page, req.PageSize)
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	ID string `uri:"id" binding:"required"`
}

// UpdateUser 更新用户
// @Summary 更新用户
// @Description 更新用户信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Param body body service.UpdateUserRequest true "用户信息"
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response{data=model.User}
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	var reqURI UpdateUserRequest
	if err := c.ShouldBindUri(&reqURI); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	var reqBody service.UpdateUserRequest
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	id, err := uuid.Parse(reqURI.ID)
	if err != nil {
		utils.BadRequest(c, "用户ID格式错误")
		return
	}

	user, err := h.userService.Update(c.Request.Context(), id, &reqBody)
	if err != nil {
		utils.ServerError(c, "更新用户失败: "+err.Error())
		return
	}

	utils.Success(c, user)
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Description 删除指定用户
// @Tags 用户管理
// @Produce json
// @Param id path string true "用户ID"
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	var req GetUserRequest
	if err := c.ShouldBindUri(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	id, err := uuid.Parse(req.ID)
	if err != nil {
		utils.BadRequest(c, "用户ID格式错误")
		return
	}

	if err := h.userService.Delete(c.Request.Context(), id); err != nil {
		utils.ServerError(c, "删除用户失败: "+err.Error())
		return
	}

	utils.Success(c, nil)
}

// GetUserStatistics 获取用户统计
// @Summary 获取用户统计
// @Description 获取用户相关统计数据
// @Tags 用户管理
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response
// @Router /users/statistics [get]
func (h *UserHandler) GetUserStatistics(c *gin.Context) {
	stats, err := h.userService.GetStatistics(c.Request.Context())
	if err != nil {
		utils.ServerError(c, "获取统计失败")
		return
	}

	utils.Success(c, stats)
}

// AssignToOrgRequest 分配组织请求
type AssignToOrgRequest struct {
	UserID string `json:"user_id" binding:"required"`
	OrgID  string `json:"org_id" binding:"required"`
}

// AssignToOrg 分配用户到组织
// @Summary 分配用户到组织
// @Description 将用户分配到指定组织
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param body body AssignToOrgRequest true "分配信息"
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response
// @Router /users/assign-to-org [post]
func (h *UserHandler) AssignToOrg(c *gin.Context) {
	var req AssignToOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		utils.BadRequest(c, "用户ID格式错误")
		return
	}

	orgID, err := uuid.Parse(req.OrgID)
	if err != nil {
		utils.BadRequest(c, "组织ID格式错误")
		return
	}

	if err := h.userService.AssignToOrg(c.Request.Context(), userID, orgID); err != nil {
		utils.ServerError(c, "分配失败: "+err.Error())
		return
	}

	utils.Success(c, nil)
}
