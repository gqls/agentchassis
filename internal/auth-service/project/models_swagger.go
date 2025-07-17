package project

// NOTE: This file contains model definitions for Swagger documentation.

// CreateProjectRequest for creating a new project
// @Description Project creation request payload
type CreateProjectRequest struct {
	// Project name (must be unique per user)
	// @example My AI Assistant Project
	Name string `json:"name" binding:"required" example:"My AI Assistant Project"`

	// Project description
	// @example A project for developing custom AI assistants
	Description string `json:"description,omitempty" example:"A project for developing custom AI assistants"`

	// Project type
	// @example ai-assistant
	Type string `json:"type" binding:"required" example:"ai-assistant" enums:"ai-assistant,workflow,integration,custom"`

	// Project settings
	Settings ProjectSettings `json:"settings,omitempty"`

	// Project metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateProjectRequest for updating a project
// @Description Project update request payload
type UpdateProjectRequest struct {
	// Project name
	// @example Updated Project Name
	Name string `json:"name,omitempty" example:"Updated Project Name"`

	// Project description
	// @example Updated project description
	Description string `json:"description,omitempty" example:"Updated project description"`

	// Project status
	// @example active
	Status string `json:"status,omitempty" example:"active" enums:"active,paused,archived"`

	// Project settings
	Settings ProjectSettings `json:"settings,omitempty"`

	// Project metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ProjectResponse represents a project
// @Description Project response with complete project information
type ProjectResponse struct {
	// Project details
	Project *Project `json:"project"`

	// Optional response message
	// @example Project retrieved successfully
	Message string `json:"message,omitempty" example:"Project retrieved successfully"`
}

// ProjectListResponse represents a list of projects
// @Description Paginated list of projects
type ProjectListResponse struct {
	// List of projects
	Projects []Project `json:"projects"`

	// Pagination information
	Pagination PaginationInfo `json:"pagination"`
}

// Project represents complete project information
// @Description Complete project information
type Project struct {
	// Unique project identifier
	// @example proj_123e4567-e89b-12d3-a456-426614174000
	ID string `json:"id" example:"proj_123e4567-e89b-12d3-a456-426614174000"`

	// User ID who owns the project
	// @example 123e4567-e89b-12d3-a456-426614174000
	UserID string `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`

	// Client ID for multi-tenant support
	// @example client-123
	ClientID string `json:"client_id" example:"client-123"`

	// Project name
	// @example My AI Assistant Project
	Name string `json:"name" example:"My AI Assistant Project"`

	// Project description
	// @example A project for developing custom AI assistants
	Description string `json:"description" example:"A project for developing custom AI assistants"`

	// Project type
	// @example ai-assistant
	Type string `json:"type" example:"ai-assistant" enums:"ai-assistant,workflow,integration,custom"`

	// Project status
	// @example active
	Status string `json:"status" example:"active" enums:"active,paused,archived,deleted"`

	// Project settings
	Settings ProjectSettings `json:"settings"`

	// Project statistics
	Stats ProjectStats `json:"stats"`

	// Project metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Creation timestamp
	// @example 2024-01-15T10:30:00Z
	CreatedAt string `json:"created_at" example:"2024-01-15T10:30:00Z"`

	// Last update timestamp
	// @example 2024-07-17T14:45:00Z
	UpdatedAt string `json:"updated_at" example:"2024-07-17T14:45:00Z"`
}

// ProjectSettings represents project configuration
// @Description Project-specific settings and configuration
type ProjectSettings struct {
	// API key for project (auto-generated)
	// @example proj_key_abc123xyz789
	APIKey string `json:"api_key,omitempty" example:"proj_key_abc123xyz789"`

	// Webhook URL for notifications
	// @example https://example.com/webhooks/project
	WebhookURL string `json:"webhook_url,omitempty" example:"https://example.com/webhooks/project"`

	// Allowed domains for CORS
	// @example ["https://app.example.com", "https://api.example.com"]
	AllowedDomains []string `json:"allowed_domains,omitempty" example:"https://app.example.com,https://api.example.com"`

	// Rate limiting settings
	RateLimits RateLimitSettings `json:"rate_limits,omitempty"`

	// Feature flags
	Features map[string]bool `json:"features,omitempty"`
}

// RateLimitSettings for project API usage
// @Description Rate limiting configuration
type RateLimitSettings struct {
	// Requests per minute
	// @example 60
	RequestsPerMinute int `json:"requests_per_minute" example:"60"`

	// Requests per hour
	// @example 1000
	RequestsPerHour int `json:"requests_per_hour" example:"1000"`

	// Requests per day
	// @example 10000
	RequestsPerDay int `json:"requests_per_day" example:"10000"`
}

// ProjectStats represents project usage statistics
// @Description Project usage and activity statistics
type ProjectStats struct {
	// Total API calls
	// @example 15420
	TotalAPICalls int64 `json:"total_api_calls" example:"15420"`

	// API calls this month
	// @example 3250
	APICallsThisMonth int64 `json:"api_calls_this_month" example:"3250"`

	// Number of agents
	// @example 5
	AgentCount int `json:"agent_count" example:"5"`

	// Number of workflows
	// @example 12
	WorkflowCount int `json:"workflow_count" example:"12"`

	// Storage used in MB
	// @example 256.5
	StorageUsedMB float64 `json:"storage_used_mb" example:"256.5"`

	// Last activity timestamp
	// @example 2024-07-17T13:30:00Z
	LastActivityAt string `json:"last_activity_at,omitempty" example:"2024-07-17T13:30:00Z"`
}

// PaginationInfo represents pagination metadata
// @Description Pagination information for list responses
type PaginationInfo struct {
	// Current page number
	// @example 1
	Page int `json:"page" example:"1"`

	// Items per page
	// @example 20
	Limit int `json:"limit" example:"20"`

	// Total number of items
	// @example 45
	Total int `json:"total" example:"45"`

	// Total number of pages
	// @example 3
	TotalPages int `json:"total_pages" example:"3"`
}
