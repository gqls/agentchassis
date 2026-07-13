package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gqls/agentchassis/internal/auth-service/user"
)

// Handlers wraps the auth service for HTTP handling
type Handlers struct {
	service ServiceInterface
}

// NewHandlers creates new auth handlers
func NewHandlers(service *Service) *Handlers {
	return &Handlers{service: service}
}

// RegisterRequest represents registration data
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email" example:"john.doe@example.com"`
	Password  string `json:"password" binding:"required,min=8" example:"SecurePassword123!"`
	ClientID  string `json:"client_id" binding:"required" example:"client-123"`
	FirstName string `json:"first_name" example:"John"`
	LastName  string `json:"last_name" example:"Doe"`
	Company   string `json:"company" example:"Acme Corp"`
}

// LoginRequest represents login data
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"john.doe@example.com"`
	Password string `json:"password" binding:"required" example:"SecurePassword123!"`
}

// RefreshRequest represents token refresh data
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
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
	// Get the token from the Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
		return
	}

	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

	// Validate the token using the service
	claims, err := h.service.ValidateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token", "details": err.Error()})
		return
	}

	// Return the validation result
	c.JSON(http.StatusOK, gin.H{
		"valid": true,
		"user": gin.H{
			"id":          claims.UserID,
			"email":       claims.Email,
			"role":        claims.Role,
			"tier":        claims.Tier,
			"client_id":   claims.ClientID,
			"permissions": claims.Permissions,
		},
	})
}
