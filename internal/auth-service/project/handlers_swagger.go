package project

// NOTE: This file contains swagger annotations for the project handlers.
// Run `swag init` to generate the swagger documentation.
// All types are defined in their respective files.

// ListProjects godoc
// @Summary      List projects
// @Description  Get a list of all projects for the authenticated user
// @Tags         Projects
// @Accept       json
// @Produce      json
// @Success      200 {object} project.ProjectListResponse "List of projects retrieved successfully"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      500 {object} map[string]interface{} "Internal server error"
// @Router       /projects [get]
// @Security     Bearer
// @ID           listProjects

// CreateProject godoc
// @Summary      Create project
// @Description  Create a new project for the authenticated user
// @Tags         Projects
// @Accept       json
// @Produce      json
// @Param        request body project.CreateProjectRequest true "Project details"
// @Success      201 {object} project.Project "Project created successfully"
// @Failure      400 {object} map[string]interface{} "Invalid request body"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      409 {object} map[string]interface{} "Project with this name already exists"
// @Failure      500 {object} map[string]interface{} "Internal server error"
// @Router       /projects [post]
// @Security     Bearer
// @ID           createProject

// GetProject godoc
// @Summary      Get project
// @Description  Get detailed information about a specific project
// @Tags         Projects
// @Accept       json
// @Produce      json
// @Param        id path string true "Project ID"
// @Success      200 {object} project.Project "Project details retrieved successfully"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      403 {object} map[string]interface{} "Forbidden - no access to this project"
// @Failure      404 {object} map[string]interface{} "Project not found"
// @Failure      500 {object} map[string]interface{} "Internal server error"
// @Router       /projects/{id} [get]
// @Security     Bearer
// @ID           getProject

// UpdateProject godoc
// @Summary      Update project
// @Description  Update an existing project
// @Tags         Projects
// @Accept       json
// @Produce      json
// @Param        id path string true "Project ID"
// @Param        request body project.UpdateProjectRequest true "Project update details"
// @Success      200 {object} project.Project "Project updated successfully"
// @Failure      400 {object} map[string]interface{} "Invalid request body"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      403 {object} map[string]interface{} "Forbidden - no access to this project"
// @Failure      404 {object} map[string]interface{} "Project not found"
// @Failure      409 {object} map[string]interface{} "Project name already in use"
// @Failure      500 {object} map[string]interface{} "Internal server error"
// @Router       /projects/{id} [put]
// @Security     Bearer
// @ID           updateProject

// DeleteProject godoc
// @Summary      Delete project
// @Description  Delete a project and all associated resources
// @Tags         Projects
// @Accept       json
// @Produce      json
// @Param        id path string true "Project ID"
// @Success      204 {string} string "Project deleted successfully"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      403 {object} map[string]interface{} "Forbidden - no access to this project"
// @Failure      404 {object} map[string]interface{} "Project not found"
// @Failure      500 {object} map[string]interface{} "Internal server error"
// @Router       /projects/{id} [delete]
// @Security     Bearer
// @ID           deleteProject
