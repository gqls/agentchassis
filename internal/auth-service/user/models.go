package user

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// User represents a user in the system
type User struct {
	ID               string     `json:"id" db:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Email            string     `json:"email" db:"email" example:"john.doe@example.com"`
	PasswordHash     string     `json:"-" db:"password_hash"`
	Role             string     `json:"role" db:"role" example:"user"`
	ClientID         string     `json:"client_id" db:"client_id" example:"client-123"`
	SubscriptionTier string     `json:"subscription_tier" db:"subscription_tier" example:"premium"`
	IsActive         bool       `json:"is_active" db:"is_active" example:"true"`
	EmailVerified    bool       `json:"email_verified" db:"email_verified" example:"true"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at" example:"2024-07-17T14:45:00Z"`
	LastLoginAt      *time.Time `json:"last_login_at,omitempty" db:"last_login_at" example:"2024-07-17T10:00:00Z"`

	// Additional fields
	Profile     *UserProfile `json:"profile,omitempty"`
	Permissions []string     `json:"permissions,omitempty" example:"read:agents,write:agents,read:workflows"`
}

// UserProfile contains additional user information
type UserProfile struct {
	UserID      string `json:"user_id" db:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	FirstName   string `json:"first_name" db:"first_name" example:"John"`
	LastName    string `json:"last_name" db:"last_name" example:"Doe"`
	Company     string `json:"company,omitempty" db:"company" example:"Acme Corp"`
	Phone       string `json:"phone,omitempty" db:"phone" example:"+1-555-123-4567"`
	AvatarURL   string `json:"avatar_url,omitempty" db:"avatar_url" example:"https://example.com/avatar.jpg"`
	Preferences JSONB  `json:"preferences,omitempty" db:"preferences"`
}

// JSONB handles JSON data in database
type JSONB map[string]interface{}

// Value implements driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

// Scan implements sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan JSONB")
	}
	return json.Unmarshal(bytes, j)
}

// CreateUserRequest for user registration
type CreateUserRequest struct {
	Email     string `json:"email" binding:"required,email" example:"john.doe@example.com"`
	Password  string `json:"password" binding:"required,min=8" example:"SecurePassword123!"`
	ClientID  string `json:"client_id" binding:"required" example:"client-123"`
	FirstName string `json:"first_name" example:"John"`
	LastName  string `json:"last_name" example:"Doe"`
	Company   string `json:"company" example:"Acme Corp"`
}

// UpdateUserRequest for user updates
type UpdateUserRequest struct {
	FirstName   *string `json:"first_name" example:"Jane"`
	LastName    *string `json:"last_name" example:"Smith"`
	Company     *string `json:"company" example:"Tech Corp"`
	Phone       *string `json:"phone" example:"+1-555-987-6543"`
	Preferences *JSONB  `json:"preferences"`
}

// ChangePasswordRequest for password changes
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required" example:"OldPassword123!"`
	NewPassword     string `json:"new_password" binding:"required,min=8" example:"NewSecurePassword456!"`
}
