package gateway

// NOTE: This file contains swagger annotations and WebSocket-related types for documentation.
// The gateway proxies WebSocket connections to the core-manager service.

// WebSocketMessage for WebSocket communication
type WebSocketMessage struct {
	Type      string      `json:"type" example:"event"`
	Event     string      `json:"event,omitempty" example:"instance.status.changed"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty" example:"Invalid command"`
	ID        string      `json:"id,omitempty" example:"msg_123e4567"`
	Timestamp string      `json:"timestamp" example:"2024-07-17T14:30:00Z"`
}

// WebSocketEventTypes contains all available WebSocket event types
type WebSocketEventTypes struct {
	InstanceEvents   []string `json:"instance_events" example:"instance.status.changed,instance.created,instance.deleted"`
	ExecutionEvents  []string `json:"execution_events" example:"execution.started,execution.completed,execution.failed"`
	SystemEvents     []string `json:"system_events" example:"system.notification,system.maintenance,system.alert"`
	ConnectionEvents []string `json:"connection_events" example:"connection.established,connection.closed,connection.error"`
}

// WebSocketCommand represents a command sent via WebSocket
type WebSocketCommand struct {
	Type  string      `json:"type" example:"command"`
	Event string      `json:"event" example:"subscribe"`
	Data  interface{} `json:"data"`
	ID    string      `json:"id" example:"cmd_123e4567"`
}

// WebSocketSubscribeData for subscription commands
type WebSocketSubscribeData struct {
	Events []string `json:"events" example:"instance.status.*,execution.complete"`
}

// WebSocketEventMessage represents an event from the server
type WebSocketEventMessage struct {
	Type      string      `json:"type" example:"event"`
	Event     string      `json:"event" example:"instance.status.changed"`
	Data      interface{} `json:"data"`
	Timestamp string      `json:"timestamp" example:"2024-07-17T14:30:00Z"`
	EntityID  string      `json:"entity_id,omitempty" example:"inst_123e4567"`
}

// InstanceStatusChangedEvent data for instance status changes
type InstanceStatusChangedEvent struct {
	InstanceID     string `json:"instance_id" example:"inst_123e4567"`
	PreviousStatus string `json:"previous_status" example:"running"`
	NewStatus      string `json:"new_status" example:"completed"`
	Reason         string `json:"reason,omitempty" example:"Execution completed successfully"`
	TriggeredBy    string `json:"triggered_by,omitempty" example:"user_123"`
}

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
// @Description All messages use JSON format with the gateway.WebSocketMessage structure
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
