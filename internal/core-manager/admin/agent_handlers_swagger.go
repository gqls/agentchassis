// FILE: internal/core-manager/admin/agent_handlers_swagger.go
package admin

// NOTE: This file contains swagger annotations for the agent handlers.
// All types are defined in agent_handlers.go

// HandleCreateAgentDefinition godoc
// @Summary      Create agent definition
// @Description  Creates a new agent type definition and triggers Kafka topic creation
// @Tags         Agent Definitions
// @Accept       json
// @Produce      json
// @Param        request body admin.AgentDefinitionRequest true "Agent definition details"
// @Success      201 {object} map[string]interface{} "Agent definition created successfully"
// @Failure      400 {object} map[string]interface{} "Invalid request body"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      403 {object} map[string]interface{} "Forbidden - admin access required"
// @Failure      409 {object} map[string]interface{} "Agent type already exists"
// @Failure      500 {object} map[string]interface{} "Failed to create agent definition"
// @Router       /admin/agent-definitions [post]
// @Security     Bearer
// @ID           createAgentDefinition

// HandleVerifyAgentTopics godoc
// @Summary      Verify agent topics
// @Description  Checks if all required Kafka topics exist for an agent type
// @Tags         Agent Definitions
// @Accept       json
// @Produce      json
// @Param        type path string true "Agent type"
// @Success      200 {object} map[string]interface{} "Topic verification result"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      403 {object} map[string]interface{} "Forbidden - admin access required"
// @Failure      500 {object} map[string]interface{} "Failed to verify topics"
// @Router       /admin/agent-definitions/{type}/topics/verify [get]
// @Security     Bearer
// @ID           verifyAgentTopics

// HandleRecreateAgentTopics godoc
// @Summary      Recreate agent topics
// @Description  Manually triggers Kafka topic creation for an agent type
// @Tags         Agent Definitions
// @Accept       json
// @Produce      json
// @Param        type path string true "Agent type"
// @Success      202 {object} map[string]interface{} "Topic creation initiated"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      403 {object} map[string]interface{} "Forbidden - admin access required"
// @Failure      404 {object} map[string]interface{} "Agent type not found"
// @Router       /admin/agent-definitions/{type}/topics/recreate [post]
// @Security     Bearer
// @ID           recreateAgentTopics

// HandleListAgentInstances godoc
// @Summary      List agent instances
// @Description  Returns all agent instances across all clients with optional filtering
// @Tags         Agent Instances
// @Accept       json
// @Produce      json
// @Param        client_id query string false "Filter by client ID"
// @Param        agent_type query string false "Filter by agent type"
// @Param        is_active query string false "Filter by active status (true/false)"
// @Success      200 {object} map[string]interface{} "List of agent instances"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      403 {object} map[string]interface{} "Forbidden - admin access required"
// @Failure      500 {object} map[string]interface{} "Failed to list instances"
// @Router       /admin/agent-instances [get]
// @Security     Bearer
// @ID           listAgentInstances

// HandleGetAgentInstance godoc
// @Summary      Get agent instance
// @Description  Returns detailed information about a specific agent instance
// @Tags         Agent Instances
// @Accept       json
// @Produce      json
// @Param        agent_id path string true "Agent instance ID"
// @Success      200 {object} map[string]interface{} "Agent instance details"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      403 {object} map[string]interface{} "Forbidden - admin access required"
// @Failure      404 {object} map[string]interface{} "Agent not found"
// @Router       /admin/agent-instances/{agent_id} [get]
// @Security     Bearer
// @ID           getAgentInstance

// HandleToggleAgentStatus godoc
// @Summary      Toggle agent status
// @Description  Enables or disables an agent instance
// @Tags         Agent Instances
// @Accept       json
// @Produce      json
// @Param        agent_id path string true "Agent instance ID"
// @Param        request body map[string]interface{} true "Status update request"
// @Success      200 {object} map[string]interface{} "Agent status updated"
// @Failure      400 {object} map[string]interface{} "Invalid request body"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      403 {object} map[string]interface{} "Forbidden - admin access required"
// @Failure      404 {object} map[string]interface{} "Agent not found"
// @Failure      500 {object} map[string]interface{} "Failed to update status"
// @Router       /admin/agent-instances/{agent_id}/status [put]
// @Security     Bearer
// @ID           toggleAgentStatus

// HandleRestartAgent godoc
// @Summary      Restart agent
// @Description  Sends a restart command to an agent instance
// @Tags         Agent Instances
// @Accept       json
// @Produce      json
// @Param        agent_id path string true "Agent instance ID"
// @Success      200 {object} map[string]interface{} "Restart command sent"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      403 {object} map[string]interface{} "Forbidden - admin access required"
// @Failure      500 {object} map[string]interface{} "Failed to send restart command"
// @Router       /admin/agent-instances/{agent_id}/restart [post]
// @Security     Bearer
// @ID           restartAgent
