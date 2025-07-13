package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gqls/ai-persona-system/internal/auth-service/user"
)

// Handlers wraps the auth service for HTTP handling
type Handlers struct {
	service *Service
}

// NewHandlers creates new auth handlers
func NewHandlers(service *Service) *Handlers {
	return &Handlers{service: service}
}

// RegisterRequest represents registration data
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	ClientID  string `json:"client_id" binding:"required"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Company   string `json:"company"`
}

// LoginRequest represents login data
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RefreshRequest represents token refresh data
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// HandleRegister handles user registration
func (h *Handlers) HandleRegister(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert to user create request
	userReq := &user.CreateUserRequest{
		Email:     req.Email,
		Password:  req.Password,
		ClientID:  req.ClientID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Company:   req.Company,
	}

	response, err := h.service.Register(c.Request.Context(), userReq)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// HandleLogin handles user login
func (h *Handlers) HandleLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// HandleRefresh handles token refresh
func (h *Handlers) HandleRefresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// HandleLogout handles user logout
func (h *Handlers) HandleLogout(c *gin.Context) {
	// Get token ID from claims
	tokenID := c.GetString("token_id")

	if err := h.service.Logout(c.Request.Context(), tokenID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

// HandleValidate validates the current token
func (h *Handlers) HandleValidate(c *gin.Context) {
	// Token is already validated by middleware if we reach here
	claims := c.MustGet("claims").(*jwt.Claims)

	c.JSON(http.StatusOK, gin.H{
		"valid": true,
		"user": gin.H{
			"id":        claims.UserID,
			"email":     claims.Email,
			"role":      claims.Role,
			"tier":      claims.Tier,
			"client_id": claims.ClientID,
		},
	})
}
