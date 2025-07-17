package auth

// NOTE: This file contains model definitions for Swagger documentation.
// These are used by swag to understand the structure of requests and responses.

// TokenResponse represents the authentication response
// @Description Authentication response containing JWT tokens and user information
type TokenResponse struct {
	// JWT access token for API authentication
	// @example eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTIzIiwiZW1haWwiOiJ1c2VyQGV4YW1wbGUuY29tIiwiZXhwIjoxNjQyMDgwMDAwfQ.abc123
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`

	// JWT refresh token for obtaining new access tokens
	// @example eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTIzIiwidG9rZW5fdHlwZSI6InJlZnJlc2giLCJleHAiOjE2NDQ2NzIwMDB9.def456
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`

	// Token type, always "Bearer"
	// @example Bearer
	TokenType string `json:"token_type" example:"Bearer"`

	// Access token expiry time in seconds
	// @example 3600
	ExpiresIn int `json:"expires_in" example:"3600"`

	// Authenticated user information
	User UserInfo `json:"user"`
}

// UserInfo represents user information in token responses
// @Description Basic user information included in authentication responses
type UserInfo struct {
	// Unique user identifier
	// @example 123e4567-e89b-12d3-a456-426614174000
	ID string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`

	// User's email address
	// @example john.doe@example.com
	Email string `json:"email" example:"john.doe@example.com"`

	// User's full name
	// @example John Doe
	Name string `json:"name" example:"John Doe"`

	// User's role in the system
	// @example user
	// @enum user
	// @enum admin
	// @enum developer
	Role string `json:"role" example:"user" enums:"user,admin,developer"`

	// User's subscription tier
	// @example premium
	// @enum free
	// @enum starter
	// @enum premium
	// @enum enterprise
	Tier string `json:"tier" example:"premium" enums:"free,starter,premium,enterprise"`

	// Client ID for multi-tenant support
	// @example client-123
	ClientID string `json:"client_id" example:"client-123"`

	// User's permissions
	// @example ["read:agents", "write:agents", "read:workflows"]
	Permissions []string `json:"permissions" example:"read:agents,write:agents,read:workflows"`

	// Account creation timestamp
	// @example 2024-01-15T10:30:00Z
	CreatedAt string `json:"created_at,omitempty" example:"2024-01-15T10:30:00Z"`
}

// Additional model definitions for error responses
// @Description Standard error response
type ErrorResponse struct {
	// Error code or type
	// @example VALIDATION_ERROR
	Error string `json:"error" example:"VALIDATION_ERROR"`

	// Human-readable error message
	// @example Invalid email format
	Message string `json:"message,omitempty" example:"Invalid email format"`

	// Additional error details
	Details map[string]interface{} `json:"details,omitempty"`
}

// MessageResponse for simple message responses
// @Description Simple message response
type MessageResponse struct {
	// Response message
	// @example Operation completed successfully
	Message string `json:"message" example:"Operation completed successfully"`
}
