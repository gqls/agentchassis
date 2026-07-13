package gateway

// NOTE: This file contains swagger annotations for the gateway handlers.
// The gateway proxies requests to the core-manager service.
// Run `swag init` to generate the swagger documentation.

// HandleTemplateRoutes godoc
// @Summary      Template management (proxy)
// @Description  Proxies template management requests to core-manager service (admin only)
// @Tags         Templates (Gateway)
// @Accept       json
// @Produce      json
// @Param        path path string false "Additional path segments"
// @Success      200 {object} map[string]interface{} "Request proxied successfully"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      403 {object} map[string]interface{} "Forbidden - admin access required"
// @Failure      502 {object} map[string]interface{} "Bad gateway - core-manager unavailable"
// @Router       /templates [get]
// @Router       /templates [post]
// @Router       /templates/{path} [get]
// @Router       /templates/{path} [put]
// @Router       /templates/{path} [delete]
// @Security     Bearer
// @ID           gatewayTemplateRoutes

// HandleInstanceRoutes godoc
// @Summary      Instance management (proxy)
// @Description  Proxies persona instance requests to core-manager service
// @Tags         Instances (Gateway)
// @Accept       json
// @Produce      json
// @Param        path path string false "Additional path segments"
// @Success      200 {object} map[string]interface{} "Request proxied successfully"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      502 {object} map[string]interface{} "Bad gateway - core-manager unavailable"
// @Router       /personas/instances [get]
// @Router       /personas/instances [post]
// @Router       /personas/instances/{path} [get]
// @Router       /personas/instances/{path} [put]
// @Router       /personas/instances/{path} [delete]
// @Security     Bearer
// @ID           gatewayInstanceRoutes

// HandleAdminRoutes godoc
// @Summary      Admin routes (proxy)
// @Description  Proxies various admin routes to core-manager service (admin only)
// @Tags         Admin (Gateway)
// @Accept       json
// @Produce      json
// @Param        path path string false "Additional path segments"
// @Success      200 {object} map[string]interface{} "Request proxied successfully"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      403 {object} map[string]interface{} "Forbidden - admin access required"
// @Failure      502 {object} map[string]interface{} "Bad gateway - core-manager unavailable"
// @Router       /admin/clients [any]
// @Router       /admin/clients/{path} [any]
// @Router       /admin/system/{path} [any]
// @Router       /admin/workflows/{path} [any]
// @Router       /admin/agent-definitions/{path} [any]
// @Security     Bearer
// @ID           gatewayAdminRoutes

// Gateway Proxy Information
// @Description The gateway service acts as a reverse proxy, forwarding requests to the core-manager service.
// @Description All requests are enriched with user context headers before forwarding:
// @Description - X-User-ID: The authenticated user's ID
// @Description - X-Client-ID: The client/tenant ID
// @Description - X-User-Role: The user's role (user, admin, moderator)
// @Description - X-User-Tier: The user's subscription tier
// @Description - X-User-Email: The user's email address
// @Description - X-User-Permissions: Comma-separated list of user permissions

// Core-Manager Endpoints (Proxied)
// @Description The following endpoints are proxied to the core-manager service:

// Templates (Admin Only)
// @Description Template management for persona definitions
// GET    /api/v1/templates              - List all templates
// POST   /api/v1/templates              - Create new template
// GET    /api/v1/templates/:id          - Get template details
// PUT    /api/v1/templates/:id          - Update template
// DELETE /api/v1/templates/:id          - Delete template
// POST   /api/v1/templates/:id/clone    - Clone template
// POST   /api/v1/templates/:id/validate - Validate template

// Persona Instances
// @Description User-created persona instances
// GET    /api/v1/personas/instances                 - List user's instances
// POST   /api/v1/personas/instances                 - Create new instance
// GET    /api/v1/personas/instances/:id             - Get instance details
// PUT    /api/v1/personas/instances/:id             - Update instance
// DELETE /api/v1/personas/instances/:id             - Delete instance
// POST   /api/v1/personas/instances/:id/execute     - Execute instance
// GET    /api/v1/personas/instances/:id/history     - Get execution history
// POST   /api/v1/personas/instances/:id/stop        - Stop running instance
// GET    /api/v1/personas/instances/:id/logs        - Get instance logs

// Admin - Client Management
// @Description Multi-tenant client management (admin only)
// GET    /api/v1/admin/clients             - List all clients
// POST   /api/v1/admin/clients             - Create new client
// GET    /api/v1/admin/clients/:id         - Get client details
// PUT    /api/v1/admin/clients/:id         - Update client
// DELETE /api/v1/admin/clients/:id         - Delete client
// GET    /api/v1/admin/clients/:id/stats   - Get client statistics

// Admin - System Management
// @Description System configuration and monitoring (admin only)
// GET    /api/v1/admin/system/health       - System health check
// GET    /api/v1/admin/system/metrics      - System metrics
// GET    /api/v1/admin/system/config       - Get system configuration
// PUT    /api/v1/admin/system/config       - Update configuration
// POST   /api/v1/admin/system/maintenance  - Toggle maintenance mode
// GET    /api/v1/admin/system/logs         - Get system logs

// Admin - Workflow Management
// @Description Workflow definitions and management (admin only)
// GET    /api/v1/admin/workflows           - List all workflows
// POST   /api/v1/admin/workflows           - Create workflow
// GET    /api/v1/admin/workflows/:id       - Get workflow details
// PUT    /api/v1/admin/workflows/:id       - Update workflow
// DELETE /api/v1/admin/workflows/:id       - Delete workflow
// GET    /api/v1/admin/workflows/:id/runs  - Get workflow runs

// Admin - Agent Definitions
// @Description Agent type definitions (admin only)
// GET    /api/v1/admin/agent-definitions           - List agent definitions
// POST   /api/v1/admin/agent-definitions           - Create agent definition
// GET    /api/v1/admin/agent-definitions/:id       - Get agent definition
// PUT    /api/v1/admin/agent-definitions/:id       - Update agent definition
// DELETE /api/v1/admin/agent-definitions/:id       - Delete agent definition
