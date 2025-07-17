package admin

import (
	u "github.com/gqls/agentchassis/internal/auth-service/user"
	"time"
)

// NOTE: This file contains model definitions for Swagger documentation.

// UserListRequest for querying users with filters
// @Description Query parameters for listing users
type UserListRequest struct {
	// Page number for pagination
	// @example 1
	Page int `form:"page,default=1" example:"1"`

	// Number of items per page
	// @example 20
	PageSize int `form:"page_size,default=20" example:"20"`

	// Filter by email (partial match)
	// @example john.doe@example.com
	Email string `form:"email" example:"john.doe@example.com"`

	// Filter by client ID
	// @example client-123
	ClientID string `form:"client_id" example:"client-123"`

	// Filter by role
	// @example admin
	Role string `form:"role" example:"admin" enums:"user,admin,moderator"`

	// Filter by subscription tier
	// @example premium
	Tier string `form:"tier" example:"premium" enums:"free,basic,premium,enterprise"`

	// Filter by active status
	// @example true
	IsActive *bool `form:"is_active" example:"true"`

	// Sort field
	// @example created_at
	SortBy string `form:"sort_by,default=created_at" example:"created_at" enums:"created_at,updated_at,email,last_login_at"`

	// Sort order
	// @example desc
	SortOrder string `form:"sort_order,default=desc" example:"desc" enums:"asc,desc"`
}

// UserListResponse for paginated user lists
// @Description Paginated list of users with metadata
type UserListResponse struct {
	// List of users
	Users []u.User `json:"users"`

	// Total number of users matching filters
	// @example 250
	TotalCount int `json:"total_count" example:"250"`

	// Current page number
	// @example 1
	Page int `json:"page" example:"1"`

	// Items per page
	// @example 20
	PageSize int `json:"page_size" example:"20"`

	// Total number of pages
	// @example 13
	TotalPages int `json:"total_pages" example:"13"`
}

// UpdateUserRequest for admin user updates
// @Description Admin request to update user details
type UpdateUserRequest struct {
	// User role (optional)
	// @example admin
	Role *string `json:"role,omitempty" example:"admin" enums:"user,admin,moderator"`

	// Subscription tier (optional)
	// @example enterprise
	SubscriptionTier *string `json:"subscription_tier,omitempty" example:"enterprise" enums:"free,basic,premium,enterprise"`

	// Active status (optional)
	// @example false
	IsActive *bool `json:"is_active,omitempty" example:"false"`

	// Email verification status (optional)
	// @example true
	EmailVerified *bool `json:"email_verified,omitempty" example:"true"`
}

// GrantPermissionRequest for granting permissions
// @Description Request to grant a permission to a user
type GrantPermissionRequest struct {
	// Permission name to grant
	// @example manage_users
	PermissionName string `json:"permission_name" binding:"required" example:"manage_users"`
}

// BulkUserOperation for bulk operations
// @Description Bulk operation configuration for multiple users
type BulkUserOperation struct {
	// List of user IDs to operate on
	// @example ["user-123", "user-456", "user-789"]
	UserIDs []string `json:"user_ids" binding:"required" example:"user-123,user-456,user-789"`

	// Operation to perform
	// @example deactivate
	Operation string `json:"operation" binding:"required,oneof=activate deactivate delete upgrade_tier" example:"deactivate" enums:"activate,deactivate,delete,upgrade_tier"`

	// Additional parameters for the operation
	Params map[string]interface{} `json:"params,omitempty"`

	// Reason for the operation (for audit trail)
	// @example Policy violation - multiple account abuse
	Reason string `json:"reason" example:"Policy violation - multiple account abuse"`
}

// UserExportRequest for exporting user data
// @Description Configuration for exporting user data
type UserExportRequest struct {
	// Export format
	// @example csv
	Format string `json:"format" binding:"required,oneof=csv json" example:"csv" enums:"csv,json"`

	// Filters to apply
	Filters UserFilters `json:"filters"`

	// Fields to include in export
	// @example ["id", "email", "role", "created_at"]
	Fields []string `json:"fields,omitempty" example:"id,email,role,created_at"`
}

// UserFilters for filtering users
// @Description Filter criteria for user queries
type UserFilters struct {
	// Filter by client ID
	// @example client-123
	ClientID string `json:"client_id,omitempty" example:"client-123"`

	// Filter by subscription tier
	// @example premium
	SubscriptionTier string `json:"subscription_tier,omitempty" example:"premium"`

	// Filter by role
	// @example admin
	Role string `json:"role,omitempty" example:"admin"`

	// Filter by active status
	// @example true
	IsActive *bool `json:"is_active,omitempty" example:"true"`

	// Filter users created after this date
	// @example 2024-01-01T00:00:00Z
	CreatedAfter *time.Time `json:"created_after,omitempty" example:"2024-01-01T00:00:00Z"`

	// Filter users created before this date
	// @example 2024-12-31T23:59:59Z
	CreatedBefore *time.Time `json:"created_before,omitempty" example:"2024-12-31T23:59:59Z"`
}

// UserImportResult for bulk import results
// @Description Results of a bulk user import operation
type UserImportResult struct {
	// Total number of rows processed
	// @example 100
	TotalProcessed int `json:"total_processed" example:"100"`

	// Number of successfully created users
	// @example 95
	Successful int `json:"successful" example:"95"`

	// Number of failed imports
	// @example 5
	Failed int `json:"failed" example:"5"`

	// List of error messages for failed imports
	// @example ["Row 23: Invalid email format", "Row 45: Email already exists"]
	Errors []string `json:"errors,omitempty" example:"Row 23: Invalid email format,Row 45: Email already exists"`

	// IDs of successfully created users
	// @example ["user-123", "user-456", "user-789"]
	UserIDs []string `json:"created_user_ids" example:"user-123,user-456,user-789"`
}

// TerminateSessionsRequest for session termination
// @Description Request to terminate user sessions
type TerminateSessionsRequest struct {
	// Reason for terminating sessions
	// @example Security breach detected
	Reason string `json:"reason" example:"Security breach detected"`
}

// ResetPasswordRequest for admin password reset
// @Description Admin request to reset a user's password
type ResetPasswordRequest struct {
	// New password (minimum 8 characters)
	// @example NewSecurePassword123!
	NewPassword string `json:"new_password" binding:"required,min=8" example:"NewSecurePassword123!"`

	// Require user to change password on next login
	// @example true
	RequireChange bool `json:"require_change" example:"true"`

	// Send notification to user
	// @example true
	NotifyUser bool `json:"notify_user" example:"true"`

	// Note to include in notification
	// @example Your password has been reset for security reasons. Please change it upon login.
	NotificationNote string `json:"notification_note,omitempty" example:"Your password has been reset for security reasons. Please change it upon login."`
}

// UserSession represents an active user session
// @Description Active user session information
type UserSession struct {
	// Session ID
	// @example sess_123e4567-e89b-12d3-a456-426614174000
	ID string `json:"id" example:"sess_123e4567-e89b-12d3-a456-426614174000"`

	// Session expiry time
	// @example 2024-07-18T15:30:00Z
	ExpiresAt time.Time `json:"expires_at" example:"2024-07-18T15:30:00Z"`

	// Session creation time
	// @example 2024-07-17T15:30:00Z
	CreatedAt time.Time `json:"created_at" example:"2024-07-17T15:30:00Z"`

	// Is session currently active
	// @example true
	IsActive bool `json:"is_active" example:"true"`

	// IP address of session
	// @example 192.168.1.100
	IPAddress string `json:"ip_address,omitempty" example:"192.168.1.100"`

	// User agent of session
	// @example Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36
	UserAgent string `json:"user_agent,omitempty" example:"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"`
}

// AuditLogEntry represents an audit log entry
// @Description Audit log entry for user actions
type AuditLogEntry struct {
	// Log entry ID
	// @example log_123e4567-e89b-12d3-a456-426614174000
	ID string `json:"id" example:"log_123e4567-e89b-12d3-a456-426614174000"`

	// Action performed
	// @example password_changed
	Action string `json:"action" example:"password_changed"`

	// Action details (JSON)
	Details map[string]interface{} `json:"details"`

	// IP address of the action
	// @example 192.168.1.100
	IPAddress string `json:"ip_address" example:"192.168.1.100"`

	// User agent string
	// @example Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36
	UserAgent string `json:"user_agent" example:"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"`

	// Timestamp of the action
	// @example 2024-07-17T14:30:00Z
	CreatedAt time.Time `json:"created_at" example:"2024-07-17T14:30:00Z"`
}

// BulkOperationResult for bulk operation outcomes
// @Description Results of a bulk operation
type BulkOperationResult struct {
	// Operation performed
	// @example deactivate
	Operation string `json:"operation" example:"deactivate"`

	// Total number of users processed
	// @example 10
	Total int `json:"total" example:"10"`

	// Number of successful operations
	// @example 8
	Succeeded int `json:"succeeded" example:"8"`

	// Number of failed operations
	// @example 2
	Failed int `json:"failed" example:"2"`

	// Error details for failed operations
	// @example ["User user-123: User not found", "User user-456: Database error"]
	Errors []string `json:"errors,omitempty" example:"User user-123: User not found,User user-456: Database error"`
}
