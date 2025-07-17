package user

// NOTE: This file contains model definitions for Swagger documentation.

// UpdateProfileRequest for updating user profile
// @Description Profile update request payload
type UpdateProfileRequest struct {
	// User's full name
	// @example Jane Doe
	Name string `json:"name,omitempty" example:"Jane Doe"`

	// User's email address
	// @example jane.doe@example.com
	Email string `json:"email,omitempty" example:"jane.doe@example.com"`

	// User's company/organization
	// @example Tech Corp
	Company string `json:"company,omitempty" example:"Tech Corp"`

	// User's phone number
	// @example +1-555-123-4567
	Phone string `json:"phone,omitempty" example:"+1-555-123-4567"`

	// Preferred language code
	// @example en-US
	Language string `json:"language,omitempty" example:"en-US"`

	// User's timezone
	// @example America/New_York
	Timezone string `json:"timezone,omitempty" example:"America/New_York"`
}

// ChangePasswordRequest for changing user password
// @Description Password change request payload
type ChangePasswordRequest struct {
	// Current password for verification
	// @example OldPassword123!
	CurrentPassword string `json:"current_password" binding:"required" example:"OldPassword123!"`

	// New password (minimum 8 characters)
	// @example NewSecurePassword456!
	NewPassword string `json:"new_password" binding:"required,min=8" example:"NewSecurePassword456!"`

	// Confirm new password (must match new_password)
	// @example NewSecurePassword456!
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword" example:"NewSecurePassword456!"`
}

// DeleteAccountRequest for account deletion
// @Description Account deletion confirmation request
type DeleteAccountRequest struct {
	// Current password for verification
	// @example SecurePassword123!
	Password string `json:"password" binding:"required" example:"SecurePassword123!"`

	// Confirmation text (must be exactly "DELETE MY ACCOUNT")
	// @example DELETE MY ACCOUNT
	Confirmation string `json:"confirmation" binding:"required" example:"DELETE MY ACCOUNT"`

	// Optional reason for account deletion
	// @example No longer need the service
	Reason string `json:"reason,omitempty" example:"No longer need the service"`
}

// UserProfileResponse represents user profile data
// @Description User profile response with complete user information
type UserProfileResponse struct {
	// User details
	User *User `json:"user"`

	// Optional response message
	// @example Profile retrieved successfully
	Message string `json:"message,omitempty" example:"Profile retrieved successfully"`
}

// User represents complete user information
// @Description Complete user profile information
type User struct {
	// Unique user identifier
	// @example 123e4567-e89b-12d3-a456-426614174000
	ID string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`

	// User's email address
	// @example john.doe@example.com
	Email string `json:"email" example:"john.doe@example.com"`

	// User's first name
	// @example John
	FirstName string `json:"first_name" example:"John"`

	// User's last name
	// @example Doe
	LastName string `json:"last_name" example:"Doe"`

	// User's full name
	// @example John Doe
	Name string `json:"name" example:"John Doe"`

	// User's role
	// @example user
	Role string `json:"role" example:"user" enums:"user,admin,developer"`

	// Subscription tier
	// @example premium
	Tier string `json:"tier" example:"premium" enums:"free,starter,premium,enterprise"`

	// Client ID for multi-tenant support
	// @example client-123
	ClientID string `json:"client_id" example:"client-123"`

	// User's company
	// @example Acme Corp
	Company string `json:"company,omitempty" example:"Acme Corp"`

	// Phone number
	// @example +1-555-123-4567
	Phone string `json:"phone,omitempty" example:"+1-555-123-4567"`

	// Preferred language
	// @example en-US
	Language string `json:"language" example:"en-US"`

	// Timezone
	// @example America/New_York
	Timezone string `json:"timezone" example:"America/New_York"`

	// Email verification status
	// @example true
	EmailVerified bool `json:"email_verified" example:"true"`

	// Two-factor authentication enabled
	// @example false
	TwoFactorEnabled bool `json:"two_factor_enabled" example:"false"`

	// Account status
	// @example active
	Status string `json:"status" example:"active" enums:"active,suspended,deleted"`

	// Account creation timestamp
	// @example 2024-01-15T10:30:00Z
	CreatedAt string `json:"created_at" example:"2024-01-15T10:30:00Z"`

	// Last update timestamp
	// @example 2024-07-17T14:45:00Z
	UpdatedAt string `json:"updated_at" example:"2024-07-17T14:45:00Z"`

	// Last login timestamp
	// @example 2024-07-17T10:00:00Z
	LastLoginAt string `json:"last_login_at,omitempty" example:"2024-07-17T10:00:00Z"`
}
