package handlers

// NOTE: This file contains swagger annotations for the instance handlers.
// All types are defined in instance.go

// HandleCreateInstance godoc
// @Summary      Create instance
// @Description  Creates a new persona instance from a template
// @Tags         Instances
// @Accept       json
// @Produce      json
// @Param        request body handlers.CreateInstanceRequest true "Instance creation details"
// @Success      201 {object} models.Persona "Instance created successfully"
// @Failure      400 {object} map[string]interface{} "Invalid request"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      500 {object} map[string]interface{} "Could not create instance"
// @Router       /api/v1/personas/instances [post]
// @Security     Bearer
// @ID           createInstance

// HandleListInstances godoc
// @Summary      List instances
// @Description  Returns all persona instances for the authenticated user
// @Tags         Instances
// @Accept       json
// @Produce      json
// @Success      200 {object} handlers.InstanceListResponse "List of instances retrieved successfully"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      500 {object} map[string]interface{} "Could not retrieve instances"
// @Router       /api/v1/personas/instances [get]
// @Security     Bearer
// @ID           listInstances

// HandleGetInstance godoc
// @Summary      Get instance
// @Description  Returns a specific persona instance by ID
// @Tags         Instances
// @Accept       json
// @Produce      json
// @Param        id path string true "Instance ID"
// @Success      200 {object} models.Persona "Instance details retrieved successfully"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      404 {object} map[string]interface{} "Instance not found"
// @Router       /api/v1/personas/instances/{id} [get]
// @Security     Bearer
// @ID           getInstance

// HandleUpdateInstance godoc
// @Summary      Update instance
// @Description  Updates an existing persona instance
// @Tags         Instances
// @Accept       json
// @Produce      json
// @Param        id path string true "Instance ID"
// @Param        request body handlers.UpdateInstanceRequest true "Instance update details"
// @Success      200 {object} models.Persona "Instance updated successfully"
// @Failure      400 {object} map[string]interface{} "Invalid request body"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      500 {object} map[string]interface{} "Failed to update instance"
// @Router       /api/v1/personas/instances/{id} [patch]
// @Security     Bearer
// @ID           updateInstance

// HandleDeleteInstance godoc
// @Summary      Delete instance
// @Description  Deletes a persona instance
// @Tags         Instances
// @Accept       json
// @Produce      json
// @Param        id path string true "Instance ID"
// @Success      204 {string} string "Instance deleted successfully"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      500 {object} map[string]interface{} "Failed to delete instance"
// @Router       /api/v1/personas/instances/{id} [delete]
// @Security     Bearer
// @ID           deleteInstance
