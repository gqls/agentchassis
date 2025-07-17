package gateway

// HandleWebSocket godoc
// @Summary      WebSocket connection
// @Description  Establishes a WebSocket connection for real-time communication with core-manager
// @Tags         WebSocket (Gateway)
// @Accept       json
// @Produce      json
// @Success      101 {string} string "Switching Protocols"
// @Failure      400 {object} map[string]interface{} "Bad request - invalid WebSocket upgrade"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      426 {object} map[string]interface{} "Upgrade required"
// @Failure      502 {object} map[string]interface{} "Bad gateway - core-manager WebSocket unavailable"
// @Router       /ws [get]
// @Security     Bearer
// @ID           gatewayWebSocket

// WebSocket Protocol Documentation
// @Description The WebSocket endpoint provides real-time bidirectional communication with the core-manager service.
// @Description
// @Description ## Connection Process:
// @Description 1. Client sends WebSocket upgrade request with Bearer token
// @Description 2. Gateway validates authentication
// @Description 3. Gateway establishes upstream connection to core-manager
// @Description 4. Messages are proxied bidirectionally
// @Description
// @Description ## Message Format:
// @Description All messages use JSON format with the following structure:
// @Description ```json
// @Description {
// @Description   "type": "command|event|response|error",
// @Description   "event": "event.name.here",
// @Description   "data": { ... },
// @Description   "id": "unique-message-id",
// @Description   "timestamp": "2024-07-17T14:30:00Z"
// @Description }
// @Description ```
// @Description
// @Description ## Available Commands:
// @Description - `subscribe`: Subscribe to events
// @Description   - Data: `{"events": ["instance.status.*", "execution.complete"]}`
// @Description - `unsubscribe`: Unsubscribe from events
// @Description   - Data: `{"events": ["instance.status.*"]}`
// @Description - `ping`: Keep-alive ping
// @Description   - Response: `{"type": "response", "event": "pong"}`
// @Description
// @Description ## Event Types:
// @Description - `instance.status.changed`: Instance status update
// @Description - `instance.execution.started`: Execution started
// @Description - `instance.execution.completed`: Execution completed
// @Description - `instance.execution.failed`: Execution failed
// @Description - `instance.log`: Instance log entry
// @Description - `system.notification`: System notification
// @Description - `connection.established`: Connection established
// @Description - `connection.closed`: Connection closed
// @Description
// @Description ## Error Handling:
// @Description Errors are sent as messages with type "error":
// @Description ```json
// @Description {
// @Description   "type": "error",
// @Description   "error": "Invalid command",
// @Description   "data": { "details": "Command 'foo' not recognized" }
// @Description }
// @Description ```
// @Description
// @Description ## Authentication:
// @Description The WebSocket connection inherits authentication from the initial HTTP upgrade request.
// @Description User context is automatically forwarded to core-manager.

// WebSocketEventTypes contains all available WebSocket event types
// @Description Available WebSocket event types for subscription
type WebSocketEventTypes struct {
	// Instance-related events
	InstanceEvents []string `json:"instance_events" example:"instance.status.changed,instance.created,instance.deleted"`

	// Execution-related events
	ExecutionEvents []string `json:"execution_events" example:"execution.started,execution.completed,execution.failed"`

	// System events
	SystemEvents []string `json:"system_events" example:"system.notification,system.maintenance,system.alert"`

	// Connection events
	ConnectionEvents []string `json:"connection_events" example:"connection.established,connection.closed,connection.error"`
}

// WebSocketCommand represents a command sent via WebSocket
// @Description Command message format for WebSocket
type WebSocketCommand struct {
	// Command type (always "command")
	// @example command
	Type string `json:"type" example:"command"`

	// Command name
	// @example subscribe
	Event string `json:"event" example:"subscribe" enums:"subscribe,unsubscribe,ping,get_status"`

	// Command data
	Data interface{} `json:"data"`

	// Unique message ID for correlation
	// @example cmd_123e4567
	ID string `json:"id" example:"cmd_123e4567"`
}

// WebSocketSubscribeData for subscription commands
// @Description Data payload for subscribe/unsubscribe commands
type WebSocketSubscribeData struct {
	// Event patterns to subscribe to (supports wildcards)
	// @example ["instance.status.*", "execution.complete"]
	Events []string `json:"events" example:"instance.status.*,execution.complete"`
}

// WebSocketEventMessage represents an event from the server
// @Description Event message format from server
type WebSocketEventMessage struct {
	// Message type (always "event")
	// @example event
	Type string `json:"type" example:"event"`

	// Event name
	// @example instance.status.changed
	Event string `json:"event" example:"instance.status.changed"`

	// Event data (varies by event type)
	Data interface{} `json:"data"`

	// Event timestamp
	// @example 2024-07-17T14:30:00Z
	Timestamp string `json:"timestamp" example:"2024-07-17T14:30:00Z"`

	// Related entity ID
	// @example inst_123e4567
	EntityID string `json:"entity_id,omitempty" example:"inst_123e4567"`
}

// InstanceStatusChangedEvent data for instance status changes
// @Description Event data for instance.status.changed events
type InstanceStatusChangedEvent struct {
	// Instance ID
	// @example inst_123e4567
	InstanceID string `json:"instance_id" example:"inst_123e4567"`

	// Previous status
	// @example running
	PreviousStatus string `json:"previous_status" example:"running"`

	// New status
	// @example completed
	NewStatus string `json:"new_status" example:"completed" enums:"idle,running,completed,failed,stopped"`

	// Status change reason
	// @example Execution completed successfully
	Reason string `json:"reason,omitempty" example:"Execution completed successfully"`

	// User who triggered the change
	// @example user_123
	TriggeredBy string `json:"triggered_by,omitempty" example:"user_123"`
}
