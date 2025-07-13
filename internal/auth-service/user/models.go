package user

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// User represents a user in the system
type User struct {
	ID               string     `json:"id" db:"id"`
	Email            string     `json:"email" db:"email"`
	PasswordHash     string     `json:"-" db:"password_hash"`
	Role             string     `json:"role" db:"role"`
	ClientID         string     `json:"client_id" db:"client_id"`
	SubscriptionTier string     `json:"subscription_tier" db:"subscription_tier"`
	IsActive         bool       `json:"is_active" db:"is_active"`
	EmailVerified    bool       `json:"email_verified" db:"email_verified"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
	LastLoginAt      *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`

	// Additional fields
	Profile     *UserProfile `json:"profile,omitempty"`
	Permissions []string     `json:"permissions,omitempty"`
}

// UserProfile contains additional user information
type UserProfile struct {
	UserID      string `json:"user_id" db:"user_id"`
	FirstName   string `json:"first_name" db:"first_name"`
	LastName    string `json:"last_name" db:"last_name"`
	Company     string `json:"company,omitempty" db:"company"`
	Phone       string `json:"phone,omitempty" db:"phone"`
	AvatarURL   string `json:"avatar_url,omitempty" db:"avatar_url"`
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
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	ClientID  string `json:"client_id" binding:"required"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Company   string `json:"company"`
}

// UpdateUserRequest for user updates
type UpdateUserRequest struct {
	FirstName   *string `json:"first_name"`
	LastName    *string `json:"last_name"`
	Company     *string `json:"company"`
	Phone       *string `json:"phone"`
	Preferences *JSONB  `json:"preferences"`
}

// ChangePasswordRequest for password changes
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}
