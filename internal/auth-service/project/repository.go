// FILE: internal/auth-service/project/repository.go
package project

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// Repository handles project data access
type Repository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewRepository creates a new project repository
func NewRepository(db *sql.DB, logger *zap.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new project
func (r *Repository) Create(ctx context.Context, project *Project) error {
	query := `
        INSERT INTO projects (id, client_id, name, description, owner_id, is_active, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `

	_, err := r.db.ExecContext(ctx, query,
		project.ID, project.ClientID, project.Name, project.Description,
		project.OwnerID, project.IsActive, project.CreatedAt, project.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	return nil
}

// GetByID retrieves a project by ID
func (r *Repository) GetByID(ctx context.Context, id string) (*Project, error) {
	var p Project
	query := `
        SELECT id, client_id, name, description, owner_id, is_active, created_at, updated_at
        FROM projects
        WHERE id = ? AND is_active = true
    `

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.ClientID, &p.Name, &p.Description,
		&p.OwnerID, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project not found")
		}
		return nil, err
	}

	return &p, nil
}

// ListByUser returns all projects for a user
func (r *Repository) ListByUser(ctx context.Context, clientID, userID string) ([]Project, error) {
	query := `
        SELECT id, client_id, name, description, owner_id, is_active, created_at, updated_at
        FROM projects
        WHERE client_id = ? AND owner_id = ? AND is_active = true
        ORDER BY created_at DESC
    `

	rows, err := r.db.QueryContext(ctx, query, clientID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(
			&p.ID, &p.ClientID, &p.Name, &p.Description,
			&p.OwnerID, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			r.logger.Error("Failed to scan project", zap.Error(err))
			continue
		}
		projects = append(projects, p)
	}

	return projects, nil
}

// Update updates a project
func (r *Repository) Update(ctx context.Context, project *Project) error {
	query := `
        UPDATE projects
        SET name = ?, description = ?, updated_at = ?
        WHERE id = ?
    `

	_, err := r.db.ExecContext(ctx, query,
		project.Name, project.Description, project.UpdatedAt, project.ID,
	)

	return err
}

// Delete soft deletes a project
func (r *Repository) Delete(ctx context.Context, id string) error {
	query := `
        UPDATE projects
        SET is_active = false, updated_at = ?
        WHERE id = ?
    `

	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}
