// internal/adapters/thunder/api/types.go
//
// Shared request/response types for the Thunder Compute REST API client.
// Schema source: https://api.thundercompute.com:8443/openapi.json
// Cross-checked against https://www.thundercompute.com/docs/api-reference/instances/create-instance
//
// IMPORTANT: Thunder's API uses snake_case field names that DIFFER from our
// internal naming (we historically called GPU type "gpu" and CPU count
// "vcpus", but Thunder calls them "gpu_type" and "cpu_cores").
// Required fields per the OpenAPI spec: cpu_cores, disk_size_gb, gpu_type,
// mode, num_gpus, template. These cannot be omitted.
package api

import "time"

// ─── Constants ──────────────────────────────────────────────────────────────

// InstanceMode values from Thunder's OpenAPI enum.
const (
	InstanceModePrototyping = "prototyping"
	InstanceModeProduction  = "production"
)

// Common gpu_type values. Thunder accepts capitalised forms ("A100", "H100").
// Adapter callers should pass UPPERCASE; provision_action normalises lowercase
// input from older callers before dispatching to the API.
const (
	GPUTypeA100 = "A100"
	GPUTypeH100 = "H100"
)

// Default values the adapter applies when caller doesn't specify.
const (
	DefaultCPUCores   = 4
	DefaultDiskSizeGB = 100
	DefaultTemplate   = "ubuntu-22.04"
)

// ─── CreateInstance ─────────────────────────────────────────────────────────

// CreateInstanceRequest matches Thunder's InstanceCreateRequest schema.
// All fields marked `required` in the OpenAPI spec MUST be sent — Thunder
// rejects the body with 400 "Invalid request body" otherwise. We removed
// omitempty from the required fields so a zero value still gets transmitted
// (caller code must populate these explicitly).
type CreateInstanceRequest struct {
	// Required by Thunder API:
	GpuType    string `json:"gpu_type"`     // e.g. "A100", "H100" — case-sensitive
	NumGPUs    int    `json:"num_gpus"`     // 1, 2, 4, 8
	CPUCores   int    `json:"cpu_cores"`    // 4, 8, 16, ...
	DiskSizeGB int    `json:"disk_size_gb"` // 100, 200, ...
	Mode       string `json:"mode"`         // "prototyping" | "production"
	Template   string `json:"template"`     // e.g. "ubuntu-22.04"

	// Optional:
	PublicKey string `json:"public_key,omitempty"` // OpenSSH public key
}

// CreateInstanceResponse is Thunder's InstanceCreateResponse.
// All three fields are required per the OpenAPI spec.
type CreateInstanceResponse struct {
	Identifier int    `json:"identifier"` // numeric ID for delete/get/modify endpoints
	Key        string `json:"key"`        // verify purpose on first real call
	UUID       string `json:"uuid"`       // string ID (may be the alternative public ID)
}

// ─── GetInstance / ListInstances ────────────────────────────────────────────
//
// TODO(thunder/api): the OpenAPI excerpt fetched on 2026-05-15 was only for
// create-instance. The schemas for /instances/{id} (get) and /instances/list
// were NOT verified. On the first successful provision, capture the actual
// JSON response shape and confirm these struct tags. Suspect fields if
// polling silently fails: status casing ("RUNNING" vs "running"), and
// whether ListInstances returns a bare array or wraps it in something like
// {"instances": [...]}.
// ────────────────────────────────────────────────────────────────────────────

// Instance is the response payload for GetInstance and elements of
// ListInstances. Field tags below are best guesses pending verification.
type Instance struct {
	Identifier int    `json:"identifier"`
	UUID       string `json:"uuid"`
	Status     string `json:"status"` // see InstanceStatus* constants

	// IP becomes populated once the instance reaches running state.
	IP string `json:"ip"`

	// Hardware spec — likely-mirrored from CreateInstanceRequest.
	GpuType    string `json:"gpu_type"`
	NumGPUs    int    `json:"num_gpus"`
	CPUCores   int    `json:"cpu_cores"`
	DiskSizeGB int    `json:"disk_size_gb"`
	Template   string `json:"template,omitempty"`
	Mode       string `json:"mode"`

	// Timestamps — may vary.
	CreatedAt time.Time `json:"created_at"`
}

// InstanceStatus* values for polling.
// VERIFY on first real provision: actual casing of these strings.
const (
	InstanceStatusPending = "pending"
	InstanceStatusRunning = "running"
	InstanceStatusFailed  = "failed"
	InstanceStatusDeleted = "deleted"
)

// IsReadyStatus returns true if the instance is in a state where it has an IP
// and can accept SSH connections.
func IsReadyStatus(status string) bool {
	// Use exact comparison; if Thunder returns "RUNNING" we'll need to switch
	// to strings.EqualFold or pre-lowercase the status here.
	return status == InstanceStatusRunning
}

// IsTerminalStatus returns true if the instance is in a state where polling
// should stop (success or terminal failure).
func IsTerminalStatus(status string) bool {
	return status == InstanceStatusRunning ||
		status == InstanceStatusFailed ||
		status == InstanceStatusDeleted
}
