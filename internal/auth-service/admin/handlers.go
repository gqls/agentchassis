package admin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	u "github.com/gqls/agentchassis/internal/auth-service/user"
	"go.uber.org/zap"
)

// Handlers provides admin endpoints for user management
type Handlers struct {
	userRepo *u.Repository
	logger   *zap.Logger
}

// NewHandlers creates new admin handlers
func NewHandlers(userRepo *u.Repository, logger *zap.Logger) *Handlers {
	return &Handlers{
		userRepo: userRepo,
		logger:   logger,
	}
}

// UserListRequest represents query parameters for listing users
type UserListRequest struct {
	Page      int    `form:"page,default=1" json:"page" example:"1"`
	PageSize  int    `form:"page_size,default=20" json:"page_size" example:"20"`
	Email     string `form:"email" json:"email,omitempty" example:"john.doe@example.com"`
	ClientID  string `form:"client_id" json:"client_id,omitempty" example:"client-123"`
	Role      string `form:"role" json:"role,omitempty" example:"admin"`
	Tier      string `form:"tier" json:"tier,omitempty" example:"premium"`
	IsActive  *bool  `form:"is_active" json:"is_active,omitempty" example:"true"`
	SortBy    string `form:"sort_by,default=created_at" json:"sort_by" example:"created_at"`
	SortOrder string `form:"sort_order,default=desc" json:"sort_order" example:"desc"`
}

// UserListResponse represents paginated user list
type UserListResponse struct {
	Users      []u.User `json:"users"`
	TotalCount int      `json:"total_count" example:"250"`
	Page       int      `json:"page" example:"1"`
	PageSize   int      `json:"page_size" example:"20"`
	TotalPages int      `json:"total_pages" example:"13"`
}

// UpdateUserRequest represents admin user update
type UpdateUserRequest struct {
	Role             *string `json:"role,omitempty" example:"admin"`
	SubscriptionTier *string `json:"subscription_tier,omitempty" example:"enterprise"`
	IsActive         *bool   `json:"is_active,omitempty" example:"false"`
	EmailVerified    *bool   `json:"email_verified,omitempty" example:"true"`
}

// HandleListUsers returns a paginated list of users with filtering
func (h *Handlers) HandleListUsers(c *gin.Context) {
	var req UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate pagination
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	// Get users from repository
	users, totalCount, err := h.userRepo.ListUsers(c.Request.Context(), u.ListUsersParams{
		Offset:    (req.Page - 1) * req.PageSize,
		Limit:     req.PageSize,
		Email:     req.Email,
		ClientID:  req.ClientID,
		Role:      req.Role,
		Tier:      req.Tier,
		IsActive:  req.IsActive,
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
	})

	if err != nil {
		h.logger.Error("Failed to list users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}

	totalPages := (totalCount + req.PageSize - 1) / req.PageSize

	c.JSON(http.StatusOK, UserListResponse{
		Users:      users,
		TotalCount: totalCount,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	})
}

// HandleGetUser returns detailed information about a specific user
func (h *Handlers) HandleGetUser(c *gin.Context) {
	userID := c.Param("user_id")

	user, err := h.userRepo.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get additional statistics
	stats, err := h.userRepo.GetUserStats(c.Request.Context(), userID)
	if err != nil {
		h.logger.Warn("Failed to get user stats", zap.Error(err))
		stats = &u.UserStats{}
	}

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"stats": stats,
	})
}

// HandleUpdateUser allows admins to update user details
func (h *Handlers) HandleUpdateUser(c *gin.Context) {
	userID := c.Param("user_id")

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate role if provided
	if req.Role != nil {
		validRoles := map[string]bool{"user": true, "admin": true, "moderator": true}
		if !validRoles[*req.Role] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
			return
		}
	}

	// Validate tier if provided
	if req.SubscriptionTier != nil {
		validTiers := map[string]bool{"free": true, "basic": true, "premium": true, "enterprise": true}
		if !validTiers[*req.SubscriptionTier] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription tier"})
			return
		}
	}

	// Update user
	err := h.userRepo.AdminUpdateUser(c.Request.Context(), userID, &u.AdminUpdateRequest{
		Role:             req.Role,
		SubscriptionTier: req.SubscriptionTier,
		IsActive:         req.IsActive,
		EmailVerified:    req.EmailVerified,
	})

	if err != nil {
		h.logger.Error("Failed to update user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	// Return updated user
	updatedUser, _ := h.userRepo.GetUserByID(c.Request.Context(), userID)
	c.JSON(http.StatusOK, updatedUser)
}

// HandleDeleteUser soft deletes a user account
func (h *Handlers) HandleDeleteUser(c *gin.Context) {
	userID := c.Param("user_id")

	// Prevent self-deletion
	currentUserID := c.GetString("user_id")
	if currentUserID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete your own account"})
		return
	}

	err := h.userRepo.DeleteUser(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to delete user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// HandleGetUserActivity returns user activity logs
func (h *Handlers) HandleGetUserActivity(c *gin.Context) {
	userID := c.Param("user_id")

	// Parse query parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	activities, err := h.userRepo.GetUserActivityLog(c.Request.Context(), userID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get user activity", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve activity log"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":    userID,
		"activities": activities,
		"count":      len(activities),
	})
}

// GrantPermissionRequest for granting permissions
type GrantPermissionRequest struct {
	PermissionName string `json:"permission_name" binding:"required" example:"manage_users"`
}

// HandleGrantPermission grants a permission to a user
func (h *Handlers) HandleGrantPermission(c *gin.Context) {
	userID := c.Param("user_id")

	var req GrantPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.userRepo.GrantPermission(c.Request.Context(), userID, req.PermissionName)
	if err != nil {
		h.logger.Error("Failed to grant permission", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to grant permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Permission granted successfully",
		"user_id":    userID,
		"permission": req.PermissionName,
	})
}

// HandleRevokePermission revokes a permission from a user
func (h *Handlers) HandleRevokePermission(c *gin.Context) {
	userID := c.Param("user_id")
	permissionName := c.Param("permission_name")

	err := h.userRepo.RevokePermission(c.Request.Context(), userID, permissionName)
	if err != nil {
		h.logger.Error("Failed to revoke permission", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Permission revoked successfully",
		"user_id":    userID,
		"permission": permissionName,
	})
}
