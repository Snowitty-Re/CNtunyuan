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
type UserHandler struct {
	userService *service.UserAppService
}

// NewUserHandler create user handler
func NewUserHandler(userService *service.UserAppService) *UserHandler {
	return &UserHandler{userService: userService}
}

// RegisterRoutes register routes
func (h *UserHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	users := router.Group("/users")
	users.Use(authMiddleware.Required())
	{
		users.GET("", h.List)
		users.POST("", middleware.RequireAdmin(), h.Create)
		users.GET("/:id", h.GetByID)
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
func (h *UserHandler) Create(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.userService.Create(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case service.ErrPhoneExists:
			response.Conflict(c, "phone already exists")
		case service.ErrEmailExists:
			response.Conflict(c, "email already exists")
		case service.ErrInvalidRole:
			response.BadRequest(c, "invalid role")
		default:
			logger.Error("Create user failed", logger.Err(err))
			response.InternalServerError(c, "create user failed")
		}
		return
	}

	response.Created(c, user)
}

// GetByID get user by ID
func (h *UserHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "user id is required")
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrUserNotFound {
			response.NotFound(c, "user not found")
			return
		}
		response.InternalServerError(c, "get user failed")
		return
	}

	response.Success(c, user)
}

// List user list
func (h *UserHandler) List(c *gin.Context) {
	var req dto.UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	users, err := h.userService.List(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "get user list failed")
		return
	}

	response.Success(c, users)
}

// Update update user
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

	operatorID := middleware.GetUserID(c)
	operator, err := h.userService.GetByID(c.Request.Context(), operatorID)
	if err != nil {
		response.Unauthorized(c, "invalid user")
		return
	}

	user, err := h.userService.Update(c.Request.Context(), id, &req, &entity.User{
		BaseEntity: entity.BaseEntity{ID: operator.ID},
		Role:       operator.Role,
	})
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			response.NotFound(c, "user not found")
		case service.ErrCannotModify:
			response.Forbidden(c, "permission denied")
		case service.ErrInvalidRole:
			response.BadRequest(c, "invalid role")
		default:
			logger.Error("Update user failed", logger.Err(err))
			response.InternalServerError(c, "update user failed")
		}
		return
	}

	response.Success(c, user)
}

// Delete delete user
func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "user id is required")
		return
	}

	operatorID := middleware.GetUserID(c)
	operator, err := h.userService.GetByID(c.Request.Context(), operatorID)
	if err != nil {
		response.Unauthorized(c, "invalid user")
		return
	}

	if err := h.userService.Delete(c.Request.Context(), id, &entity.User{
		BaseEntity: entity.BaseEntity{ID: operator.ID},
		Role:       operator.Role,
	}); err != nil {
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

	operatorID := middleware.GetUserID(c)
	operator, err := h.userService.GetByID(c.Request.Context(), operatorID)
	if err != nil {
		response.Unauthorized(c, "invalid user")
		return
	}

	if err := h.userService.UpdateStatus(c.Request.Context(), id, req.Status, &entity.User{
		BaseEntity: entity.BaseEntity{ID: operator.ID},
		Role:       operator.Role,
	}); err != nil {
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
func (h *UserHandler) UpdateRole(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "user id is required")
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	operatorID := middleware.GetUserID(c)
	operator, err := h.userService.GetByID(c.Request.Context(), operatorID)
	if err != nil {
		response.Unauthorized(c, "invalid user")
		return
	}

	if err := h.userService.UpdateRole(c.Request.Context(), id, req.Role, &entity.User{
		BaseEntity: entity.BaseEntity{ID: operator.ID},
		Role:       operator.Role,
	}); err != nil {
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
		if err.Error() == "old password is wrong" {
			response.BadRequest(c, "old password is wrong")
		} else {
			response.InternalServerError(c, "change password failed")
		}
		return
	}

	response.Success(c, nil)
}

// GetStats get stats
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
