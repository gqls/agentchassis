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

import (
	"strconv"
	"strings"
	"time"
)

// ─── Constants ──────────────────────────────────────────────────────────────

// InstanceMode values from Thunder's OpenAPI enum.
const (
	InstanceModePrototyping = "prototyping"
	InstanceModeProduction  = "production"
)

// Valid gpu_type values, VERIFIED 2026-05-20 against GET /v1/specs (a 201
// create with "a100xl" succeeded). These are the FIRST segment of the /specs
// composite keys (e.g. spec key "a100xl_x1_prototyping" → gpu_type "a100xl",
// where x1 is the GPU COUNT and prototyping is the mode); the composite key
// itself is NOT a valid gpu_type. Note Thunder offers NO plain "a100" — the
// 80GB A100 is "a100xl". The CLI's `--gpu a100` flag maps to a100xl internally,
// which is why every "a100" API attempt 400'd.
const (
	GPUTypeA100XL = "a100xl" // NVIDIA A100 80GB
	GPUTypeA6000  = "a6000"  // RTX A6000 48GB (prototyping only)
	GPUTypeH100   = "h100"   // NVIDIA H100 80GB
	GPUTypeL40    = "l40"    // NVIDIA L40 48GB
	GPUTypeL40S   = "l40s"   // NVIDIA L40S 48GB
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
	Key        string `json:"key"`        // server-generated SSH PRIVATE key (PEM). VERIFIED 2026-05-20: Thunder returns its own keypair's private key here. OPEN: does Thunder also honour the public_key we send, or only its own key? Affects which key the adapter must store for SSH.
	UUID       string `json:"uuid"`       // string ID (may be the alternative public ID)
}

// ─── GetInstance / ListInstances ────────────────────────────────────────────
//
// Instance shape VERIFIED 2026-05-20 from `tnr status --json` (the list/status
// endpoint). Still unverified: (a) the CreateInstanceResponse shape — that's a
// different endpoint (POST /instances/create) and may differ; the body-preview
// logging in client.go will reveal it on the next real provision. (b) Whether
// GET /instances/{id} returns a bare object or an array — GetInstance falls
// back to list-and-filter on 404, but a non-404 shape surprise would surface
// in the response body_preview log.
// ────────────────────────────────────────────────────────────────────────────

// Instance is the response payload for GetInstance and elements of
// ListInstances.
//
// VERIFIED 2026-05-20 against `tnr status --json` on a live instance.
// Three surprises confirmed by real data:
//  1. The RESPONSE uses camelCase (gpuType, cpuCores, numGpus, createdAt) —
//     even though the CREATE REQUEST uses snake_case (gpu_type, cpu_cores).
//     Thunder's API is asymmetric. Do NOT assume request and response share
//     field names.
//  2. Numeric fields come back as JSON STRINGS: "id":"0", "cpuCores":"4",
//     "numGpus":"1". Only "storage" and "port" are real ints. So id/cpuCores/
//     numGpus are typed as string here; use IdentifierInt() to get the int id.
//  3. Enum values are UPPERCASE in responses: "status":"RUNNING",
//     "gpuType":"A100XL" — even though we send them lowercase. Status
//     comparisons must be case-insensitive (see IsReadyStatus).
type Instance struct {
	// ID arrives as a JSON string (e.g. "0"). Use IdentifierInt() for the int.
	ID     string `json:"id"`
	UUID   string `json:"uuid"`
	Name   string `json:"name"`
	Status string `json:"status"` // UPPERCASE in responses, e.g. "RUNNING"

	// IP becomes populated once the instance reaches running state.
	IP   string `json:"ip"`
	Port int    `json:"port"`

	// Hardware spec — camelCase, numbers as strings.
	GpuType  string `json:"gpuType"`  // e.g. "A100XL" (uppercase in responses)
	NumGpus  string `json:"numGpus"`  // JSON string, e.g. "1"
	CPUCores string `json:"cpuCores"` // JSON string, e.g. "4"
	Memory   string `json:"memory"`   // JSON string, GB, e.g. "32"
	Storage  int    `json:"storage"`  // real int, GB, e.g. 100
	Template string `json:"template"`
	Mode     string `json:"mode"`

	CreatedAt time.Time `json:"createdAt"`
}

// IdentifierInt parses the string id into an int for endpoints that take a
// numeric identifier (delete/get). Returns 0 and false if unparseable.
func (i *Instance) IdentifierInt() (int, bool) {
	n, err := strconv.Atoi(i.ID)
	if err != nil {
		return 0, false
	}
	return n, true
}

// InstanceStatus* values for polling. Stored lowercase; compare case-
// insensitively because Thunder returns UPPERCASE in responses ("RUNNING").
const (
	InstanceStatusPending = "pending"
	InstanceStatusRunning = "running"
	InstanceStatusFailed  = "failed"
	InstanceStatusDeleted = "deleted"
)

// IsReadyStatus returns true if the instance is in a state where it has an IP
// and can accept SSH connections. Case-insensitive: Thunder returns "RUNNING"
// (uppercase) in API responses while our constants are lowercase, so we
// lowercase both sides — a casing change on either side won't break the loop.
func IsReadyStatus(status string) bool {
	return strings.ToLower(strings.TrimSpace(status)) == InstanceStatusRunning
}

// IsTerminalStatus returns true if the instance is in a state where polling
// should stop (success or terminal failure). Case-insensitive, same reason.
func IsTerminalStatus(status string) bool {
	s := strings.ToLower(strings.TrimSpace(status))
	return s == InstanceStatusRunning ||
		s == InstanceStatusFailed ||
		s == InstanceStatusDeleted
}
