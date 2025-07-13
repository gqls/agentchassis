// FILE: core-manager/pkg/models/interfaces.go
// We update the repository interface to include the new methods.
package models

import (
	"context"
	"github.com/google/uuid"
)

type PersonaRepository interface {
	// ... (existing methods: CreateProject, GetProjectDetails, etc.)

	// --- NEW Template Methods ---
	CreateTemplate(ctx context.Context, template *Persona) (*Persona, error)
	GetTemplateByID(ctx context.Context, id string) (*Persona, error)
	ListTemplates(ctx context.Context) ([]Persona, error)
	UpdateTemplate(ctx context.Context, template *Persona) (*Persona, error)
	DeleteTemplate(ctx context.Context, id string) error

	// --- NEW Instance Methods ---
	// Note: CreateInstanceFromTemplate was already defined, but we refine it here.
	CreateInstanceFromTemplate(ctx context.Context, templateID string, userID string, instanceName string) (*Persona, error)
	GetInstanceByID(ctx context.Context, id string) (*Persona, error)
	ListInstances(ctx context.Context, userID string) ([]Persona, error)
	UpdateInstance(ctx context.Context, id string, name *string, config map[string]interface{}) (*Persona, error)
	DeleteInstance(ctx context.Context, id string) error
}
