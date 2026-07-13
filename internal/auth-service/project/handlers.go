// FILE: internal/auth-service/project/handlers.go
package project

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// HTTPHandler handles project-related HTTP requests
type HTTPHandler struct {
	repo   *Repository
	logger *zap.Logger
}

// NewHTTPHandler creates a new project HTTP handler
func NewHTTPHandler(repo *Repository, logger *zap.Logger) *HTTPHandler {
	return &HTTPHandler{
		repo:   repo,
		logger: logger,
	}
}

// CreateProjectRequest for creating a new project
type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required" example:"My AI Assistant Project"`
	Description string `json:"description,omitempty" example:"A project for developing custom AI assistants"`
}

// UpdateProjectRequest for updating a project
type UpdateProjectRequest struct {
	Name        *string `json:"name,omitempty" example:"Updated Project Name"`
	Description *string `json:"description,omitempty" example:"Updated project description"`
}

// ProjectListResponse represents a list of projects
type ProjectListResponse struct {
	Projects []Project `json:"projects"`
	Count    int       `json:"count" example:"5"`
}

// CreateProject handles project creation
func (h *HTTPHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	clientID := r.Context().Value("client_id").(string)

	var req CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	project := &Project{
		ID:          uuid.New().String(),
		ClientID:    clientID,
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     userID,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.repo.Create(r.Context(), project); err != nil {
		h.logger.Error("Failed to create project", zap.Error(err))
		http.Error(w, "Failed to create project", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(project)
}

// ListProjects returns all projects for a user
func (h *HTTPHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	clientID := r.Context().Value("client_id").(string)

	projects, err := h.repo.ListByUser(r.Context(), clientID, userID)
	if err != nil {
		h.logger.Error("Failed to list projects", zap.Error(err))
		http.Error(w, "Failed to retrieve projects", http.StatusInternalServerError)
		return
	}

	response := ProjectListResponse{
		Projects: projects,
		Count:    len(projects),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetProject returns a specific project
func (h *HTTPHandler) GetProject(w http.ResponseWriter, r *http.Request, projectID string) {
	userID := r.Context().Value("user_id").(string)
	clientID := r.Context().Value("client_id").(string)

	project, err := h.repo.GetByID(r.Context(), projectID)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	// Verify ownership
	if project.ClientID != clientID || project.OwnerID != userID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

// UpdateProject updates a project
func (h *HTTPHandler) UpdateProject(w http.ResponseWriter, r *http.Request, projectID string) {
	userID := r.Context().Value("user_id").(string)
	clientID := r.Context().Value("client_id").(string)

	// Verify ownership first
	project, err := h.repo.GetByID(r.Context(), projectID)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	if project.ClientID != clientID || project.OwnerID != userID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Parse update request
	var req UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Apply updates
	if req.Name != nil {
		project.Name = *req.Name
	}
	if req.Description != nil {
		project.Description = *req.Description
	}
	project.UpdatedAt = time.Now()

	if err := h.repo.Update(r.Context(), project); err != nil {
		h.logger.Error("Failed to update project", zap.Error(err))
		http.Error(w, "Failed to update project", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

// DeleteProject deletes a project
func (h *HTTPHandler) DeleteProject(w http.ResponseWriter, r *http.Request, projectID string) {
	userID := r.Context().Value("user_id").(string)
	clientID := r.Context().Value("client_id").(string)

	// Verify ownership
	project, err := h.repo.GetByID(r.Context(), projectID)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	if project.ClientID != clientID || project.OwnerID != userID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	if err := h.repo.Delete(r.Context(), projectID); err != nil {
		h.logger.Error("Failed to delete project", zap.Error(err))
		http.Error(w, "Failed to delete project", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
