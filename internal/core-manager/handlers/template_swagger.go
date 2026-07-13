package handlers

// NOTE: This file contains swagger annotations for the template handlers.
// All types are defined in template.go

// HandleCreateTemplate godoc
// @Summary      Create template
// @Description  Creates a new persona template (admin only)
// @Tags         Templates
// @Accept       json
// @Produce      json
// @Param        request body handlers.CreateTemplateRequest true "Template creation details"
// @Success      201 {object} models.Persona "Template created successfully"
// @Failure      400 {object} map[string]interface{} "Invalid request body"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      403 {object} map[string]interface{} "Forbidden - admin access required"
// @Failure      500 {object} map[string]interface{} "Failed to create template"
// @Router       /api/v1/templates [post]
// @Security     Bearer
// @ID           createTemplate

// HandleListTemplates godoc
// @Summary      List templates
// @Description  Returns all available persona templates (admin only)
// @Tags         Templates
// @Accept       json
// @Produce      json
// @Success      200 {object} handlers.TemplateListResponse "List of templates retrieved successfully"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      403 {object} map[string]interface{} "Forbidden - admin access required"
// @Failure      500 {object} map[string]interface{} "Failed to retrieve templates"
// @Router       /api/v1/templates [get]
// @Security     Bearer
// @ID           listTemplates

// HandleGetTemplate godoc
// @Summary      Get template
// @Description  Returns a specific persona template by ID (admin only)
// @Tags         Templates
// @Accept       json
// @Produce      json
// @Param        id path string true "Template ID"
// @Success      200 {object} models.Persona "Template details retrieved successfully"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      403 {object} map[string]interface{} "Forbidden - admin access required"
// @Failure      404 {object} map[string]interface{} "Template not found"
// @Router       /api/v1/templates/{id} [get]
// @Security     Bearer
// @ID           getTemplate

// HandleUpdateTemplate godoc
// @Summary      Update template
// @Description  Updates an existing persona template (admin only)
// @Tags         Templates
// @Accept       json
// @Produce      json
// @Param        id path string true "Template ID"
// @Param        request body handlers.CreateTemplateRequest true "Template update details"
// @Success      200 {object} models.Persona "Template updated successfully"
// @Failure      400 {object} map[string]interface{} "Invalid request body"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      403 {object} map[string]interface{} "Forbidden - admin access required"
// @Failure      500 {object} map[string]interface{} "Failed to update template"
// @Router       /api/v1/templates/{id} [put]
// @Security     Bearer
// @ID           updateTemplate

// HandleDeleteTemplate godoc
// @Summary      Delete template
// @Description  Deletes a persona template (admin only)
// @Tags         Templates
// @Accept       json
// @Produce      json
// @Param        id path string true "Template ID"
// @Success      204 {string} string "Template deleted successfully"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      403 {object} map[string]interface{} "Forbidden - admin access required"
// @Failure      500 {object} map[string]interface{} "Failed to delete template"
// @Router       /api/v1/templates/{id} [delete]
// @Security     Bearer
// @ID           deleteTemplate
