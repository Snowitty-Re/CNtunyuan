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

// UserHandler user handler
// @Description User management endpoints
// @Tags 用户管理
// @BasePath /api/v1
type UserHandler struct {
	userService *service.UserAppService
}

// NewUserHandler create user handler
func NewUserHandler(userService *service.UserAppService) *UserHandler {
	return &UserHandler{userService: userService}
}

func userOperator(c *gin.Context) *entity.User {
	return &entity.User{
		BaseEntity: entity.BaseEntity{ID: middleware.GetUserID(c)},
		Role:       middleware.GetUserRole(c),
		OrgID:      middleware.GetOrgID(c),
	}
}

// RegisterRoutes register routes
func (h *UserHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	users := router.Group("/users")
	users.Use(authMiddleware.Required())
	{
		users.GET("", middleware.RequireManager(), h.List)
		users.POST("", middleware.RequireAdmin(), h.Create)
		users.GET("/:id", middleware.RequireManager(), h.GetByID)
		users.PUT("/:id", middleware.RequireAdmin(), h.Update)
		users.DELETE("/:id", middleware.RequireAdmin(), h.Delete)
		users.PUT("/:id/status", middleware.RequireManager(), h.UpdateStatus)
		users.PUT("/:id/role", middleware.RequireAdmin(), h.UpdateRole)
	}

	profile := router.Group("/profile")
	profile.Use(authMiddleware.Required())
	{
		profile.GET("", h.GetProfile)
		profile.PUT("", h.UpdateProfile)
		profile.PUT("/password", h.ChangePassword)
		profile.GET("/stats", h.GetStats)
	}
}

// Create create user
// @Summary Create user
// @Description Create a new user (admin only)
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.CreateUserRequest true "User creation request"
// @Success 201 {object} response.Response{data=dto.UserResponse} "User created successfully"
// @Failure 400 {object} response.Response "Invalid request parameters"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Forbidden - admin only"
// @Failure 409 {object} response.Response "Phone or email already exists"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.userService.Create(c.Request.Context(), &req, userOperator(c))
	if err != nil {
		switch err {
		case service.ErrPhoneExists:
			response.Conflict(c, "phone already exists")
		case service.ErrEmailExists:
			response.Conflict(c, "email already exists")
		case service.ErrUserAlreadyExists:
			response.Conflict(c, "user already exists")
		case service.ErrInvalidRole:
			response.BadRequest(c, "invalid role")
		case service.ErrInvalidOrgID:
			response.BadRequest(c, "invalid organization id")
		case service.ErrCannotModify:
			response.Forbidden(c, "permission denied")
		default:
			logger.Error("Create user failed", logger.Err(err))
			response.InternalServerError(c, "create user failed")
		}
		return
	}

	response.Created(c, user)
}

// GetByID get user by ID
// @Summary Get user by ID
// @Description Get user details by ID
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "User ID"
// @Success 200 {object} response.Response{data=dto.UserResponse} "User details retrieved successfully"
// @Failure 400 {object} response.Response "Invalid user ID"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 404 {object} response.Response "User not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /users/{id} [get]
func (h *UserHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "user id is required")
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), id, userOperator(c))
	if err != nil {
		if err == service.ErrUserNotFound {
			response.NotFound(c, "user not found")
			return
		}
		if err == service.ErrCannotModify {
			response.Forbidden(c, "permission denied")
			return
		}
		response.InternalServerError(c, "get user failed")
		return
	}

	response.Success(c, user)
}

// List user list
// @Summary List users
// @Description Get paginated list of users with optional filters
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "Page number (default: 1)" default(1) minimum(1)
// @Param page_size query int false "Page size (default: 10, max: 100)" default(10) minimum(1) maximum(100)
// @Param keyword query string false "Search keyword for nickname, phone, or email"
// @Param role query string false "Filter by role: super_admin, admin, manager, volunteer"
// @Param status query string false "Filter by status: active, inactive, banned"
// @Param org_id query string false "Filter by organization ID"
// @Success 200 {object} response.Response{data=dto.UserListResponse} "User list retrieved successfully"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /users [get]
func (h *UserHandler) List(c *gin.Context) {
	var req dto.UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	users, err := h.userService.List(c.Request.Context(), &req, userOperator(c))
	if err != nil {
		response.InternalServerError(c, "get user list failed")
		return
	}

	response.Success(c, users)
}

// Update update user
// @Summary Update user
// @Description Update user details (admin only)
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "User ID"
// @Param request body dto.UpdateUserRequest true "User update request"
// @Success 200 {object} response.Response{data=dto.UserResponse} "User updated successfully"
// @Failure 400 {object} response.Response "Invalid request parameters"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Forbidden - admin only or cannot modify this user"
// @Failure 404 {object} response.Response "User not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "user id is required")
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.userService.Update(c.Request.Context(), id, &req, userOperator(c))
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			response.NotFound(c, "user not found")
		case service.ErrCannotModify:
			response.Forbidden(c, "permission denied")
		case service.ErrInvalidRole:
			response.BadRequest(c, "invalid role")
		case service.ErrPhoneExists:
			response.Conflict(c, "phone already exists")
		case service.ErrEmailExists:
			response.Conflict(c, "email already exists")
		case service.ErrUserAlreadyExists:
			response.Conflict(c, "user already exists")
		case service.ErrInvalidOrgID:
			response.BadRequest(c, "invalid organization id")
		default:
			logger.Error("Update user failed", logger.Err(err))
			response.InternalServerError(c, "update user failed")
		}
		return
	}

	response.Success(c, user)
}

// Delete delete user
// @Summary Delete user
// @Description Delete a user (admin only)
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "User ID"
// @Success 204 "User deleted successfully"
// @Failure 400 {object} response.Response "Invalid user ID"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Forbidden - admin only or cannot modify this user"
// @Failure 404 {object} response.Response "User not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "user id is required")
		return
	}

	if err := h.userService.Delete(c.Request.Context(), id, userOperator(c)); err != nil {
		switch err {
		case service.ErrUserNotFound:
			response.NotFound(c, "user not found")
		case service.ErrCannotModify:
			response.Forbidden(c, "permission denied")
		default:
			logger.Error("Delete user failed", logger.Err(err))
			response.InternalServerError(c, "delete user failed")
		}
		return
	}

	response.NoContent(c)
}

// UpdateStatus update user status
// @Summary Update user status
// @Description Update user status (manager and above)
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "User ID"
// @Param request body object{status=entity.UserStatus} true "Status update request (active, inactive, banned)"
// @Success 200 {object} response.Response "User status updated successfully"
// @Failure 400 {object} response.Response "Invalid request parameters"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Forbidden - manager and above only"
// @Failure 404 {object} response.Response "User not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /users/{id}/status [put]
func (h *UserHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "user id is required")
		return
	}

	var req struct {
		Status entity.UserStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.userService.UpdateStatus(c.Request.Context(), id, req.Status, userOperator(c)); err != nil {
		switch err {
		case service.ErrUserNotFound:
			response.NotFound(c, "user not found")
		case service.ErrCannotModify:
			response.Forbidden(c, "permission denied")
		default:
			response.InternalServerError(c, "update status failed")
		}
		return
	}

	response.Success(c, nil)
}

// UpdateRole update user role
// @Summary Update user role
// @Description Update user role (admin only)
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "User ID"
// @Param request body object{role=entity.Role} true "Role update request (super_admin, admin, manager, volunteer)"
// @Success 200 {object} response.Response "User role updated successfully"
// @Failure 400 {object} response.Response "Invalid request parameters"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Forbidden - admin only or cannot modify this user"
// @Failure 404 {object} response.Response "User not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /users/{id}/role [put]
func (h *UserHandler) UpdateRole(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "user id is required")
		return
	}

	var req struct {
		Role entity.Role `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.userService.UpdateRole(c.Request.Context(), id, req.Role, userOperator(c)); err != nil {
		switch err {
		case service.ErrUserNotFound:
			response.NotFound(c, "user not found")
		case service.ErrCannotModify:
			response.Forbidden(c, "permission denied")
		case service.ErrInvalidRole:
			response.BadRequest(c, "invalid role")
		default:
			response.InternalServerError(c, "update role failed")
		}
		return
	}

	response.Success(c, nil)
}

// GetProfile get profile
// @Summary Get current user profile
// @Description Get the profile information of the currently authenticated user
// @Tags 个人中心
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=dto.UserProfileResponse} "Profile retrieved successfully"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.Unauthorized(c, "please login first")
		return
	}

	profile, err := h.userService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		response.InternalServerError(c, "get profile failed")
		return
	}

	response.Success(c, profile)
}

// UpdateProfile update profile
// @Summary Update current user profile
// @Description Update the profile information of the currently authenticated user
// @Tags 个人中心
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.UpdateProfileRequest true "Profile update request"
// @Success 200 {object} response.Response{data=dto.UserProfileResponse} "Profile updated successfully"
// @Failure 400 {object} response.Response "Invalid request parameters"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.Unauthorized(c, "please login first")
		return
	}

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	profile, err := h.userService.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		response.InternalServerError(c, "update profile failed")
		return
	}

	response.Success(c, profile)
}

// ChangePassword change password
// @Summary Change password
// @Description Change the password of the currently authenticated user
// @Tags 个人中心
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.ChangePasswordRequest true "Change password request"
// @Success 200 {object} response.Response "Password changed successfully"
// @Failure 400 {object} response.Response "Invalid request or old password is wrong"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /profile/password [put]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.Unauthorized(c, "please login first")
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.userService.ChangePassword(c.Request.Context(), userID, &req); err != nil {
		if err == service.ErrOldPasswordWrong {
			response.BadRequest(c, "old password is wrong")
		} else {
			response.InternalServerError(c, "change password failed")
		}
		return
	}

	response.Success(c, nil)
}

// GetStats get stats
// @Summary Get user statistics
// @Description Get statistics for the currently authenticated user
// @Tags 个人中心
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=dto.UserStatsResponse} "Statistics retrieved successfully"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /profile/stats [get]
func (h *UserHandler) GetStats(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.Unauthorized(c, "please login first")
		return
	}

	stats, err := h.userService.GetStats(c.Request.Context(), userID)
	if err != nil {
		response.InternalServerError(c, "get stats failed")
		return
	}

	response.Success(c, stats)
}
