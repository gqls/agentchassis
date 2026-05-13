// FILE: internal/adapters/thunder/api/types.go
//
// Request and response types for the Thunder Compute API.
// Source: https://www.thundercompute.com/docs/api-reference/
// CLI reference: https://www.thundercompute.com/docs/cli-reference
//
// VERIFICATION NOTE: A few field names in CreateInstanceRequest and Instance
// are educated guesses based on the CLI flag names (`tnr create --gpu ...`)
// and observable response shapes. They will need verification on the first
// real API call — the OpenAPI spec at api.thundercompute.com:8443/openapi.json
// is the source of truth.

package api

import (
	"time"
)

// ─────────────────────────────────────────────────────────────────────────
// Create instance — POST /v1/instances/create
// ─────────────────────────────────────────────────────────────────────────

// CreateInstanceRequest matches the body of POST /v1/instances/create.
// All fields are optional except where noted. When PublicKey is provided,
// Thunder seeds the instance's authorized_keys with it; when omitted,
// Thunder generates a keypair and returns the private key in the response
// (only once — store immediately). For this adapter we always generate
// client-side and pass PublicKey, so the response Key field will be empty.
type CreateInstanceRequest struct {
	// GPU type: "a100", "h100", "t4", etc. See thundercompute.com pricing
	// page for current options. Empty string accepts Thunder's default.
	GPU string `json:"gpu,omitempty"`

	// NumGPUs defaults to 1 if omitted. Multi-GPU support varies by tier.
	NumGPUs int `json:"num_gpus,omitempty"`

	// VCPUs defaults to 4 (which means 32GB RAM at 8GB/vCPU).
	// Prototyping mode allows custom; production has fixed values.
	VCPUs int `json:"vcpus,omitempty"`

	// DiskSizeGB defaults to 100. Can be increased post-creation but not decreased.
	DiskSizeGB int `json:"disk_size_gb,omitempty"`

	// Mode: "prototyping" (cheaper, recommended for ≤24h jobs) or "production"
	// (premium, for long-running or customer-facing). Defaults to prototyping.
	Mode string `json:"mode,omitempty"`

	// Template: pre-configured instance image — "pytorch", "webui-forge", etc.
	// Empty string means a base Ubuntu image.
	Template string `json:"template,omitempty"`

	// PublicKey: SSH public key (ssh-ed25519 or ssh-rsa) added to authorized_keys.
	// Strongly preferred: client generates keypair, keeps private key in
	// k8s Secret, passes only the public key here. If omitted, Thunder
	// generates a keypair and returns the private key in CreateInstanceResponse.Key.
	PublicKey string `json:"public_key,omitempty"`
}

// CreateInstanceResponse is the 201 body of POST /v1/instances/create.
// Confirmed shape from Thunder docs:
//
//	{
//	  "identifier": 123,
//	  "key": "<string>",
//	  "uuid": "<string>"
//	}
type CreateInstanceResponse struct {
	// Identifier is the numeric instance ID used by other endpoints
	// (e.g. POST /v1/instances/{identifier}/delete).
	Identifier int `json:"identifier"`

	// Key is the server-generated SSH private key, ONLY populated when
	// PublicKey was omitted in the request. Returned ONCE — Thunder
	// does not store it. For our client-side-keypair flow, this is "".
	Key string `json:"key,omitempty"`

	// UUID is a stable per-instance identifier used by audit endpoints.
	UUID string `json:"uuid"`
}

// ─────────────────────────────────────────────────────────────────────────
// List/Get instance — GET /v1/instances, GET /v1/instances/{id}
// ─────────────────────────────────────────────────────────────────────────

// Instance is the shape returned by GET /v1/instances entries and the
// status payload from `tnr status`. Includes everything needed for the
// thunder_instances database row.
//
// TODO(verify-on-first-call): the exact JSON field names below are
// educated guesses from the CLI status output. Adjust after first real
// API call confirms the wire shape.
type Instance struct {
	Identifier int    `json:"identifier"`
	UUID       string `json:"uuid"`
	Status     string `json:"status"` // see InstanceStatus* constants below

	// IP is the public IP for SSH. Empty until Status reaches "RUNNING".
	IP string `json:"ip"`

	// Hardware
	GPU        string `json:"gpu"`
	NumGPUs    int    `json:"num_gpus"`
	VCPUs      int    `json:"vcpus"`
	DiskSizeGB int    `json:"disk_size_gb"`
	Template   string `json:"template,omitempty"`
	Mode       string `json:"mode"`

	// Timing
	CreatedAt time.Time `json:"created_at"`
}

// Instance status constants. The Thunder API uses these (upper-case)
// values; verify exact spelling on first real call.
const (
	InstanceStatusPending     = "PENDING"
	InstanceStatusRunning     = "RUNNING"
	InstanceStatusStopped     = "STOPPED"
	InstanceStatusTerminating = "TERMINATING"
	InstanceStatusTerminated  = "TERMINATED"
	InstanceStatusError       = "ERROR"
)

// IsTerminalStatus returns true if the instance cannot transition further
// (no more polling needed). RUNNING is "terminal" for our provision flow
// (we wait until RUNNING then stop polling); TERMINATED and ERROR mean
// we should give up entirely.
func IsTerminalStatus(s string) bool {
	switch s {
	case InstanceStatusRunning, InstanceStatusTerminated, InstanceStatusError:
		return true
	}
	return false
}

// IsReadyStatus returns true if the instance is ready for SSH access.
func IsReadyStatus(s string) bool {
	return s == InstanceStatusRunning
}
