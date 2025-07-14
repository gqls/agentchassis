// FILE: internal/core-manager/database/personas.go

package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type PersonaRepository struct {
	templatesDB *pgxpool.Pool
	clientsDB   *pgxpool.Pool
	logger      *zap.Logger
}

// NewPersonaRepository creates a new repository instance
func NewPersonaRepository(templatesDB, clientsDB *pgxpool.Pool, logger *zap.Logger) *PersonaRepository {
	return &PersonaRepository{
		templatesDB: templatesDB,
		clientsDB:   clientsDB,
		logger:      logger,
	}
}

// ClientsDB returns the clients database pool
func (r *PersonaRepository) ClientsDB() *pgxpool.Pool {
	return r.clientsDB
}

// TemplatesDB returns the templates database pool
func (r *PersonaRepository) TemplatesDB() *pgxpool.Pool {
	return r.templatesDB
}

// Template Methods

func (r *PersonaRepository) CreateTemplate(ctx context.Context, template *models.Persona) (*models.Persona, error) {
	r.logger.Info("Creating new persona template", zap.String("name", template.Name))

	configJSON, _ := json.Marshal(template.Config)

	query := `
        INSERT INTO persona_templates (id, name, description, category, config, is_active, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id
    `

	err := r.templatesDB.QueryRow(ctx, query,
		template.ID,
		template.Name,
		template.Description,
		template.Category,
		configJSON,
		true,
		template.CreatedAt,
		template.UpdatedAt,
	).Scan(&template.ID)

	if err != nil {
		r.logger.Error("Failed to create template", zap.Error(err))
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	return template, nil
}

func (r *PersonaRepository) GetTemplateByID(ctx context.Context, id string) (*models.Persona, error) {
	r.logger.Info("Getting template by ID", zap.String("id", id))

	var template models.Persona
	var configJSON []byte

	query := `
        SELECT id, name, description, category, config, is_active, created_at, updated_at
        FROM persona_templates
        WHERE id = $1 AND is_active = true
    `

	err := r.templatesDB.QueryRow(ctx, query, id).Scan(
		&template.ID,
		&template.Name,
		&template.Description,
		&template.Category,
		&configJSON,
		&template.IsActive,
		&template.CreatedAt,
		&template.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("template not found")
		}
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	json.Unmarshal(configJSON, &template.Config)
	template.IsTemplate = true

	return &template, nil
}

func (r *PersonaRepository) ListTemplates(ctx context.Context) ([]models.Persona, error) {
	r.logger.Info("Listing all templates")

	query := `
        SELECT id, name, description, category, config, is_active, created_at, updated_at
        FROM persona_templates
        WHERE is_active = true
        ORDER BY category, name
    `

	rows, err := r.templatesDB.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}
	defer rows.Close()

	var templates []models.Persona
	for rows.Next() {
		var template models.Persona
		var configJSON []byte

		err := rows.Scan(
			&template.ID,
			&template.Name,
			&template.Description,
			&template.Category,
			&configJSON,
			&template.IsActive,
			&template.CreatedAt,
			&template.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan template row", zap.Error(err))
			continue
		}

		json.Unmarshal(configJSON, &template.Config)
		template.IsTemplate = true
		templates = append(templates, template)
	}

	return templates, nil
}

func (r *PersonaRepository) UpdateTemplate(ctx context.Context, template *models.Persona) (*models.Persona, error) {
	r.logger.Info("Updating template", zap.String("id", template.ID.String()))

	configJSON, _ := json.Marshal(template.Config)

	query := `
        UPDATE persona_templates
        SET name = $2, description = $3, category = $4, config = $5, updated_at = $6
        WHERE id = $1
        RETURNING updated_at
    `

	err := r.templatesDB.QueryRow(ctx, query,
		template.ID,
		template.Name,
		template.Description,
		template.Category,
		configJSON,
		time.Now(),
	).Scan(&template.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to update template: %w", err)
	}

	return template, nil
}

func (r *PersonaRepository) DeleteTemplate(ctx context.Context, id string) error {
	r.logger.Info("Deleting template", zap.String("id", id))

	// Soft delete
	query := `UPDATE persona_templates SET is_active = false, updated_at = $2 WHERE id = $1`

	_, err := r.templatesDB.Exec(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	return nil
}

// Instance Methods

func (r *PersonaRepository) CreateInstanceFromTemplate(ctx context.Context, templateID string, userID string, instanceName string) (*models.Persona, error) {
	r.logger.Info("Creating instance from template",
		zap.String("templateID", templateID),
		zap.String("userID", userID))

	// Get client ID from context
	clientID, ok := ctx.Value("client_id").(string)
	if !ok {
		return nil, fmt.Errorf("client_id not found in context")
	}

	// First, fetch the template
	template, err := r.GetTemplateByID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch template: %w", err)
	}

	// Create the instance
	instance := &models.Persona{
		ID:          uuid.New(),
		Name:        instanceName,
		Description: template.Description,
		Category:    template.Category,
		Config:      template.Config,
		IsTemplate:  false,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	configJSON, _ := json.Marshal(instance.Config)

	// Use client-specific schema
	query := fmt.Sprintf(`
        INSERT INTO client_%s.agent_instances 
        (id, template_id, owner_user_id, name, config, is_active, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `, clientID)

	_, err = r.clientsDB.Exec(ctx, query,
		instance.ID,
		templateID,
		userID,
		instance.Name,
		configJSON,
		true,
		instance.CreatedAt,
		instance.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create instance: %w", err)
	}

	return instance, nil
}

func (r *PersonaRepository) GetInstanceByID(ctx context.Context, id string) (*models.Persona, error) {
	clientID, _ := ctx.Value("client_id").(string)

	var instance models.Persona
	var configJSON []byte

	query := fmt.Sprintf(`
        SELECT id, name, config, is_active, created_at, updated_at
        FROM client_%s.agent_instances
        WHERE id = $1 AND is_active = true
    `, clientID)

	err := r.clientsDB.QueryRow(ctx, query, id).Scan(
		&instance.ID,
		&instance.Name,
		&configJSON,
		&instance.IsActive,
		&instance.CreatedAt,
		&instance.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}

	json.Unmarshal(configJSON, &instance.Config)
	return &instance, nil
}

func (r *PersonaRepository) ListInstances(ctx context.Context, userID string) ([]models.Persona, error) {
	clientID, _ := ctx.Value("client_id").(string)

	query := fmt.Sprintf(`
        SELECT id, name, config, is_active, created_at, updated_at
        FROM client_%s.agent_instances
        WHERE owner_user_id = $1 AND is_active = true
        ORDER BY created_at DESC
    `, clientID)

	rows, err := r.clientsDB.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list instances: %w", err)
	}
	defer rows.Close()

	var instances []models.Persona
	for rows.Next() {
		var instance models.Persona
		var configJSON []byte

		err := rows.Scan(
			&instance.ID,
			&instance.Name,
			&configJSON,
			&instance.IsActive,
			&instance.CreatedAt,
			&instance.UpdatedAt,
		)
		if err != nil {
			continue
		}

		json.Unmarshal(configJSON, &instance.Config)
		instances = append(instances, instance)
	}

	return instances, nil
}

func (r *PersonaRepository) UpdateInstance(ctx context.Context, id string, name *string, config map[string]interface{}) (*models.Persona, error) {
	clientID, _ := ctx.Value("client_id").(string)

	// First get the current instance
	instance, err := r.GetInstanceByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields
	if name != nil {
		instance.Name = *name
	}
	if config != nil {
		// Merge configs
		for k, v := range config {
			instance.Config[k] = v
		}
	}

	configJSON, _ := json.Marshal(instance.Config)

	query := fmt.Sprintf(`
        UPDATE client_%s.agent_instances
        SET name = $2, config = $3, updated_at = $4
        WHERE id = $1
    `, clientID)

	_, err = r.clientsDB.Exec(ctx, query, id, instance.Name, configJSON, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to update instance: %w", err)
	}

	return instance, nil
}

func (r *PersonaRepository) DeleteInstance(ctx context.Context, id string) error {
	clientID, _ := ctx.Value("client_id").(string)

	query := fmt.Sprintf(`
        UPDATE client_%s.agent_instances
        SET is_active = false, updated_at = $2
        WHERE id = $1
    `, clientID)

	_, err := r.clientsDB.Exec(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete instance: %w", err)
	}

	return nil
}

// AdminUpdateInstanceConfig allows an admin to update an instance's config
func (r *PersonaRepository) AdminUpdateInstanceConfig(ctx context.Context, clientID, instanceID string, config map[string]interface{}) error {
	r.logger.Info("Admin updating instance config", zap.String("instance_id", instanceID), zap.String("client_id", clientID))

	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	query := fmt.Sprintf(`
        UPDATE client_%s.agent_instances
        SET config = $2, updated_at = NOW()
        WHERE id = $1
    `, clientID)

	res, err := r.clientsDB.Exec(ctx, query, instanceID, configJSON)
	if err != nil {
		return fmt.Errorf("failed to execute update: %w", err)
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("instance not found")
	}

	return nil
}
