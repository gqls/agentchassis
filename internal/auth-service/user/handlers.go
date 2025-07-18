package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handlers wraps the user service for HTTP handling
type Handlers struct {
	service *Service
}

// NewHandlers creates new user handlers
func NewHandlers(service *Service) *Handlers {
	return &Handlers{service: service}
}

// DeleteAccountRequest for account deletion
type DeleteAccountRequest struct {
	Password     string `json:"password" binding:"required" example:"SecurePassword123!"`
	Confirmation string `json:"confirmation" binding:"required" example:"DELETE MY ACCOUNT"`
	Reason       string `json:"reason,omitempty" example:"No longer need the service"`
}

// HandleGetCurrentUser returns the current user's details
func (h *Handlers) HandleGetCurrentUser(c *gin.Context) {
	userID := c.GetString("user_id")

	user, err := h.service.GetUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// HandleUpdateCurrentUser updates the current user's details
func (h *Handlers) HandleUpdateCurrentUser(c *gin.Context) {
	userID := c.GetString("user_id")

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateUser(c.Request.Context(), userID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	// Return updated user
	user, _ := h.service.GetUser(c.Request.Context(), userID)
	c.JSON(http.StatusOK, user)
}

// HandleChangePassword changes the user's password
func (h *Handlers) HandleChangePassword(c *gin.Context) {
	userID := c.GetString("user_id")

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ChangePassword(c.Request.Context(), userID, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// HandleDeleteAccount deletes the user's account
func (h *Handlers) HandleDeleteAccount(c *gin.Context) {
	userID := c.GetString("user_id")

	var req DeleteAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate confirmation
	if req.Confirmation != "DELETE MY ACCOUNT" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid confirmation text"})
		return
	}

	// Verify password
	user, err := h.service.GetUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify account"})
		return
	}

	// Validate password through login service
	_, err = h.service.Login(c.Request.Context(), user.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect password"})
		return
	}

	if err := h.service.DeleteUser(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
}
