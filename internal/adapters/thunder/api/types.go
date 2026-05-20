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

// Valid gpu_type values, lowercase per Thunder's CLI reference
// (the OpenAPI spec's "H100" example casing is misleading — the real
// enum is lowercase). t4=16GB, a100=40GB, a100xl=80GB, h100=80GB.
const (
	GPUTypeT4     = "t4"
	GPUTypeA100   = "a100"
	GPUTypeA100XL = "a100xl"
	GPUTypeH100   = "h100"
)

// Default values the adapter applies when caller doesn't specify.
const (
	DefaultCPUCores   = 4
	DefaultDiskSizeGB = 100
	// "base" is Thunder's plain GPU instance template (verified against the
	// docs — the OpenAPI spec's "ubuntu-22.04" example is NOT a real template
	// name and Thunder rejects it with 400 "invalid template"). The named
	// templates (ollama, comfy-ui, forge-neo, unsloth) are pre-built AI stacks
	// we don't want for fine-tuning. To enumerate live templates, GET
	// /thunder-templates.
	DefaultTemplate = "base"
)

// ─── CreateInstance ─────────────────────────────────────────────────────────

// CreateInstanceRequest matches Thunder's InstanceCreateRequest schema.
// All fields marked `required` in the OpenAPI spec MUST be sent — Thunder
// rejects the body with 400 "Invalid request body" otherwise. We removed
// omitempty from the required fields so a zero value still gets transmitted
// (caller code must populate these explicitly).
type CreateInstanceRequest struct {
	// Required by Thunder API:
	GpuType    string `json:"gpu_type"`     // e.g. "a100", "h100" — lowercase, case-sensitive
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
