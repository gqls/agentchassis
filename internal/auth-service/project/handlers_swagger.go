package project

// NOTE: This file contains swagger annotations for the project handlers.
// Run `swag init` to generate the swagger documentation.

// ListProjects godoc
// @Summary      List user projects
// @Description  Returns all projects owned by the authenticated user
// @Tags         Projects
// @Accept       json
// @Produce      json
// @Success      200 {object} map[string]interface{} "Projects retrieved successfully"
// @Failure      401 {object} gin.H "Unauthorized"
// @Router       /api/v1/projects [get]
// @Security     BearerAuth
// @ID           listProjects

// CreateProject godoc
// @Summary      Create project
// @Description  Creates a new project
// @Tags         Projects
// @Accept       json
// @Produce      json
// @Param        request body map[string]string true "Project details with name and description"
// @Success      201 {object} Project "Project created successfully"
// @Failure      400 {string} string "Invalid request body"
// @Failure      401 {object} gin.H "Unauthorized"
// @Router       /api/v1/projects [post]
// @Security     BearerAuth
// @ID           createProject

// GetProject godoc
// @Summary      Get project details
// @Description  Returns details of a specific project
// @Tags         Projects
// @Accept       json
// @Produce      json
// @Param        id path string true "Project ID" format(uuid)
// @Success      200 {object} Project "Project retrieved successfully"
// @Failure      401 {object} gin.H "Unauthorized"
// @Failure      403 {string} string "Access denied"
// @Failure      404 {string} string "Project not found"
// @Router       /api/v1/projects/{id} [get]
// @Security     BearerAuth
// @ID           getProject

// UpdateProject godoc
// @Summary      Update project
// @Description  Updates a project's information
// @Tags         Projects
// @Accept       json
// @Produce      json
// @Param        id path string true "Project ID" format(uuid)
// @Param        request body map[string]*string true "Project update fields"
// @Success      200 {object} Project "Project updated successfully"
// @Failure      400 {string} string "Invalid request body"
// @Failure      401 {object} gin.H "Unauthorized"
// @Failure      403 {string} string "Access denied"
// @Failure      404 {string} string "Project not found"
// @Router       /api/v1/projects/{id} [put]
// @Security     BearerAuth
// @ID           updateProject

// DeleteProject godoc
// @Summary      Delete project
// @Description  Deletes a project
// @Tags         Projects
// @Accept       json
// @Produce      json
// @Param        id path string true "Project ID" format(uuid)
// @Success      204 "Project deleted successfully"
// @Failure      401 {object} gin.H "Unauthorized"
// @Failure      403 {string} string "Access denied"
// @Failure      404 {string} string "Project not found"
// @Router       /api/v1/projects/{id} [delete]
// @Security     BearerAuth
// @ID           deleteProject
