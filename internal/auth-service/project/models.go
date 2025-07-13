// FILE: internal/auth-service/project/models.go
package project

import "time"

// Project represents a project in the system
type Project struct {
	ID          string    `json:"id" db:"id"`
	ClientID    string    `json:"client_id" db:"client_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	OwnerID     string    `json:"owner_id" db:"owner_id"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
