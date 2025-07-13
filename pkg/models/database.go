// FILE: pkg/models/database.go
package models

import (
	"context"
	"github.com/google/uuid"
	"time"
)

// Persona represents both templates and instances
type Persona struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Config      map[string]interface{} `json:"config"`
	IsTemplate  bool                   `json:"is_template"`
	IsActive    bool                   `json:"is_active"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// PersonaRepository defines the interface for persona data access
type PersonaRepository interface {
	// Template methods
	CreateTemplate(ctx context.Context, template *Persona) (*Persona, error)
	GetTemplateByID(ctx context.Context, id string) (*Persona, error)
	ListTemplates(ctx context.Context) ([]Persona, error)
	UpdateTemplate(ctx context.Context, template *Persona) (*Persona, error)
	DeleteTemplate(ctx context.Context, id string) error

	// Instance methods
	CreateInstanceFromTemplate(ctx context.Context, templateID string, userID string, instanceName string) (*Persona, error)
	GetInstanceByID(ctx context.Context, id string) (*Persona, error)
	ListInstances(ctx context.Context, userID string) ([]Persona, error)
	UpdateInstance(ctx context.Context, id string, name *string, config map[string]interface{}) (*Persona, error)
	DeleteInstance(ctx context.Context, id string) error
}
