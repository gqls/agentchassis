// FILE: internal/auth-service/project/models.go
package project

import "time"

// Project represents a project in the system
type Project struct {
	ID          string    `json:"id" db:"id" example:"proj_123e4567-e89b-12d3-a456-426614174000"`
	ClientID    string    `json:"client_id" db:"client_id" example:"client-123"`
	Name        string    `json:"name" db:"name" example:"My AI Assistant Project"`
	Description string    `json:"description" db:"description" example:"A project for developing custom AI assistants"`
	OwnerID     string    `json:"owner_id" db:"owner_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	IsActive    bool      `json:"is_active" db:"is_active" example:"true"`
	CreatedAt   time.Time `json:"created_at" db:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at" example:"2024-07-17T14:45:00Z"`
}
