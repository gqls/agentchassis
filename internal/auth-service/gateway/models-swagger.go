package gateway

// NOTE: This file contains model definitions for Swagger documentation.
// The gateway primarily acts as a proxy, so most models are defined in core-manager.

// GatewayErrorResponse for gateway-specific errors
// @Description Error response from the gateway service
type GatewayErrorResponse struct {
	// Error code
	// @example BAD_GATEWAY
	Error string `json:"error" example:"BAD_GATEWAY"`

	// Human-readable error message
	// @example Service temporarily unavailable
	Message string `json:"message" example:"Service temporarily unavailable"`

	// HTTP status code
	// @example 502
	StatusCode int `json:"status_code,omitempty" example:"502"`

	// Target service that failed
	// @example core-manager
	Service string `json:"service,omitempty" example:"core-manager"`
}

// ProxyHeaders represents headers added by the gateway
// @Description Headers automatically added to proxied requests
type ProxyHeaders struct {
	// User ID from authentication token
	// @example 123e4567-e89b-12d3-a456-426614174000
	XUserID string `json:"X-User-ID" example:"123e4567-e89b-12d3-a456-426614174000"`

	// Client/Tenant ID
	// @example client-123
	XClientID string `json:"X-Client-ID" example:"client-123"`

	// User's role
	// @example admin
	XUserRole string `json:"X-User-Role" example:"admin"`

	// User's subscription tier
	// @example premium
	XUserTier string `json:"X-User-Tier" example:"premium"`

	// User's email address
	// @example user@example.com
	XUserEmail string `json:"X-User-Email" example:"user@example.com"`

	// Comma-separated permissions
	// @example read:templates,write:templates,manage:instances
	XUserPermissions string `json:"X-User-Permissions" example:"read:templates,write:templates,manage:instances"`
}

// WebSocketMessage for WebSocket communication
// @Description WebSocket message format
type WebSocketMessage struct {
	// Message type
	// @example event
	Type string `json:"type" example:"event" enums:"event,command,response,error"`

	// Event or command name
	// @example instance.status.changed
	Event string `json:"event,omitempty" example:"instance.status.changed"`

	// Message payload
	Data interface{} `json:"data,omitempty"`

	// Error details (for error type)
	Error string `json:"error,omitempty" example:"Invalid command"`

	// Message ID for request/response correlation
	// @example msg_123e4567
	ID string `json:"id,omitempty" example:"msg_123e4567"`

	// Timestamp
	// @example 2024-07-17T14:30:00Z
	Timestamp string `json:"timestamp" example:"2024-07-17T14:30:00Z"`
}

// ServiceHealth represents core-manager service health
// @Description Health status of the core-manager service
type ServiceHealth struct {
	// Service name
	// @example core-manager
	Service string `json:"service" example:"core-manager"`

	// Health status
	// @example healthy
	Status string `json:"status" example:"healthy" enums:"healthy,degraded,unhealthy"`

	// Service version
	// @example 1.2.3
	Version string `json:"version" example:"1.2.3"`

	// Uptime in seconds
	// @example 86400
	Uptime int64 `json:"uptime" example:"86400"`

	// Last check timestamp
	// @example 2024-07-17T14:30:00Z
	LastCheck string `json:"last_check" example:"2024-07-17T14:30:00Z"`
}

// RateLimitInfo for rate limiting information
// @Description Rate limiting information added to responses
type RateLimitInfo struct {
	// Requests limit per window
	// @example 1000
	Limit int `json:"limit" example:"1000"`

	// Remaining requests in current window
	// @example 950
	Remaining int `json:"remaining" example:"950"`

	// Window reset timestamp (Unix)
	// @example 1689696000
	Reset int64 `json:"reset" example:"1689696000"`

	// Retry after (seconds, only when rate limited)
	// @example 60
	RetryAfter int `json:"retry_after,omitempty" example:"60"`
}

// ProxyConfig represents gateway proxy configuration
// @Description Gateway proxy configuration
type ProxyConfig struct {
	// Target service URL
	// @example http://core-manager:8080
	TargetURL string `json:"target_url" example:"http://core-manager:8080"`

	// Request timeout in seconds
	// @example 30
	Timeout int `json:"timeout" example:"30"`

	// Enable request logging
	// @example true
	EnableLogging bool `json:"enable_logging" example:"true"`

	// Buffer size for WebSocket connections
	// @example 1024
	WebSocketBufferSize int `json:"websocket_buffer_size" example:"1024"`
}
