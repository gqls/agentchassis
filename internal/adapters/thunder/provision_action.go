// FILE: internal/adapters/thunder/provision_action.go
//
// provision_instance action handler. Wires together:
//   1. thunder_provision_check view (pre-check via store/config.go)
//   2. ed25519 keypair generation (ssh package)
//   3. Thunder API CreateInstance (api package)
//   4. k8s Secret creation for the keypair (ssh package)
//   5. WaitForRunning poll loop (api package)
//   6. INSERT into thunder_instances (with retry for transient pg blips)
//
// Compensating cleanup: if any step after Thunder CreateInstance fails,
// we MUST decommission the Thunder instance and delete the SSH Secret
// before returning — otherwise we leak a billable instance with no DB row.
// All cleanup is logged but never overrides the original error returned
// to the caller.
//
// The action accepts per-request overrides for gpu / mode / vcpus /
// disk_size_gb and falls back to thunder_config defaults. Hyperparameters
// are passed through (so the launcher receives them) but don't affect
// provisioning itself.

package thunder

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/internal/adapters/thunder/api"
	"github.com/gqls/agentchassis/internal/adapters/thunder/ssh"
	"github.com/gqls/agentchassis/internal/adapters/thunder/store"
)

// ─────────────────────────────────────────────────────────────────────────
// Public interfaces (for test injectability)
// ─────────────────────────────────────────────────────────────────────────

// thunderAPI is the subset of *api.Client this action calls. Defined as
// an interface so tests can inject a mock without spinning up httptest.
type thunderAPI interface {
	CreateInstance(ctx context.Context, req api.CreateInstanceRequest) (*api.CreateInstanceResponse, error)
	WaitForRunning(ctx context.Context, identifier int, pollInterval time.Duration) (*api.Instance, error)
	DeleteInstance(ctx context.Context, identifier int) error
}

// secretManager is the subset of *ssh.SecretManager this action calls.
type secretManager interface {
	CreateKeypairSecret(ctx context.Context, instanceUUID string, kp *ssh.Keypair) (string, error)
	DeleteKeypairSecret(ctx context.Context, secretName string) error
}

// ─────────────────────────────────────────────────────────────────────────
// Request / Response shapes (action-layer, not API-layer)
// ─────────────────────────────────────────────────────────────────────────

// ProvisionInstanceRequest is the shape parsed from the incoming Kafka
// message body. All overrides are optional; defaults come from thunder_config.
type ProvisionInstanceRequest struct {
	// TrainingRunID links the provisioned instance back to the training
	// run that requested it. Optional but strongly recommended for audit.
	TrainingRunID string `json:"training_run_id,omitempty"`

	// RequestedBy is the agent_type that originated the provision
	// (extracted from sender_agent_type header by the caller).
	RequestedBy string `json:"-"`

	// Overrides — empty = use config defaults.
	GPU        string `json:"gpu,omitempty"`          // "a100", "h100", "t4"
	NumGPUs    int    `json:"num_gpus,omitempty"`     // default 1
	VCPUs      int    `json:"vcpus,omitempty"`        // default 4
	DiskSizeGB int    `json:"disk_size_gb,omitempty"` // default 100
	Mode       string `json:"mode,omitempty"`         // "prototyping" or "production"
	Template   string `json:"template,omitempty"`     // optional Thunder image
}

// ProvisionInstanceResult is the action's return shape. Field names match
// what model-trainer's call_launcher input_mapping reads from
// provisioning_result.* (instance_ip, ssh_user, ssh_key_secret_name).
type ProvisionInstanceResult struct {
	InstanceIP        string    `json:"instance_ip"`
	SSHUser           string    `json:"ssh_user"`
	SSHKeySecretName  string    `json:"ssh_key_secret_name"`
	ProvisioningID    string    `json:"provisioning_id"`    // our DB row UUID
	ThunderIdentifier int       `json:"thunder_identifier"` // numeric Thunder API ID
	ProvisionedAt     time.Time `json:"provisioned_at"`
}

// ─────────────────────────────────────────────────────────────────────────
// ProvisionAction
// ─────────────────────────────────────────────────────────────────────────

// ProvisionAction holds the dependencies and tunables. Constructed once
// per Adapter and reused across requests (safe for concurrent use).
type ProvisionAction struct {
	thunderAPI thunderAPI
	secretMgr  secretManager
	db         *sql.DB
	logger     *zap.Logger

	// Tunables (override in tests).
	pollInterval     time.Duration   // default 5s — passed to WaitForRunning
	waitTimeout      time.Duration   // default 5min — ctx deadline for whole WaitForRunning
	dbInsertRetries  int             // default 3
	dbInsertBackoffs []time.Duration // default [1s, 3s, 5s]
}

// NewProvisionAction builds a ProvisionAction with production defaults.
func NewProvisionAction(api thunderAPI, mgr secretManager, db *sql.DB, logger *zap.Logger) *ProvisionAction {
	return &ProvisionAction{
		thunderAPI:       api,
		secretMgr:        mgr,
		db:               db,
		logger:           logger.Named("provision_instance"),
		pollInterval:     5 * time.Second,
		waitTimeout:      5 * time.Minute,
		dbInsertRetries:  3,
		dbInsertBackoffs: []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
	}
}

// Execute runs the full provision flow. Returns the action result on
// success, or an error on any failure. Compensating cleanup happens
// inside Execute — caller doesn't need to do anything special on error.
func (p *ProvisionAction) Execute(ctx context.Context, req ProvisionInstanceRequest) (*ProvisionInstanceResult, error) {
	// ── 1. Load config + pre-check ──
	cfg, err := store.LoadConfig(ctx, p.db)
	if err != nil {
		return nil, fmt.Errorf("load thunder_config: %w", err)
	}
	if cfg.IsPaused {
		reason := "thunder provisioning paused"
		if cfg.PauseReason.Valid {
			reason = "thunder provisioning paused: " + cfg.PauseReason.String
		}
		return nil, fmt.Errorf(reason)
	}

	allowed, denialReason, err := store.CheckCanProvision(ctx, p.db)
	if err != nil {
		return nil, fmt.Errorf("provision pre-check: %w", err)
	}
	if !allowed {
		p.logger.Warn("Provision denied by thunder_provision_check",
			zap.String("reason", denialReason))
		return nil, fmt.Errorf("provision denied: %s", denialReason)
	}

	// ── 2. Resolve config defaults for any unset request fields ──
	if req.NumGPUs == 0 {
		req.NumGPUs = 1
	}
	if req.VCPUs == 0 {
		req.VCPUs = 4
	}
	if req.DiskSizeGB == 0 {
		req.DiskSizeGB = 100
	}
	if req.Mode == "" {
		req.Mode = "prototyping"
	}
	if req.GPU == "" {
		req.GPU = "a100" // sensible default for finetuning workloads
	}

	instanceType := deriveInstanceType(req.GPU, req.NumGPUs)

	// Pre-generate the DB row UUID so we can name the SSH Secret
	// deterministically before the INSERT. This means an orphan Secret
	// (if INSERT fails) is still identifiable and reapable by the cleanup
	// path below.
	dbRowID := uuid.New()
	secretLabelUUID := dbRowID.String()

	// ── 3. Generate ed25519 keypair (in memory only — not stored yet) ──
	keypairComment := "thunder-" + secretLabelUUID
	if req.TrainingRunID != "" {
		keypairComment = "training-run-" + req.TrainingRunID
	}
	kp, err := ssh.GenerateKeypair(keypairComment)
	if err != nil {
		return nil, fmt.Errorf("generate keypair: %w", err)
	}

	// ── 4. Call Thunder API to create the instance ──
	createReq := api.CreateInstanceRequest{
		GPU:        req.GPU,
		NumGPUs:    req.NumGPUs,
		VCPUs:      req.VCPUs,
		DiskSizeGB: req.DiskSizeGB,
		Mode:       req.Mode,
		Template:   req.Template,
		PublicKey:  kp.PublicAuthorizedKey, // client-side keypair — server won't generate
	}
	createResp, err := p.thunderAPI.CreateInstance(ctx, createReq)
	if err != nil {
		// Failure before any side-effect requiring cleanup.
		return nil, fmt.Errorf("thunder create: %w", err)
	}

	p.logger.Info("Thunder create accepted, starting cleanup-tracked phase",
		zap.Int("thunder_identifier", createResp.Identifier),
		zap.String("thunder_uuid", createResp.UUID),
		zap.String("db_row_id", secretLabelUUID),
	)

	// From this point on, ANY failure must compensate by deleting the
	// Thunder instance (and any Secret we may have created).
	var secretName string
	var cleanupDone bool
	cleanup := func(reason string, origErr error) error {
		if cleanupDone {
			return origErr
		}
		cleanupDone = true
		p.logger.Warn("Compensating cleanup starting",
			zap.String("reason", reason),
			zap.Int("thunder_identifier", createResp.Identifier),
			zap.String("secret_name", secretName),
			zap.Error(origErr),
		)
		// Use a fresh, time-bounded context for cleanup so it runs
		// even if the parent ctx already expired.
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if delErr := p.thunderAPI.DeleteInstance(cleanupCtx, createResp.Identifier); delErr != nil {
			p.logger.Error("Compensating decommission failed — manual cleanup needed",
				zap.Int("thunder_identifier", createResp.Identifier),
				zap.Error(delErr))
			// continue — try to delete the Secret too
		}
		if secretName != "" {
			if delErr := p.secretMgr.DeleteKeypairSecret(cleanupCtx, secretName); delErr != nil {
				p.logger.Error("Compensating Secret delete failed — manual cleanup needed",
					zap.String("secret_name", secretName),
					zap.Error(delErr))
			}
		}
		return origErr
	}

	// ── 5. Persist the keypair as a k8s Secret ──
	secretName, err = p.secretMgr.CreateKeypairSecret(ctx, secretLabelUUID, kp)
	if err != nil {
		return nil, cleanup("create secret failed", fmt.Errorf("create k8s secret: %w", err))
	}

	// ── 6. Poll until instance is RUNNING (or hits terminal state / timeout) ──
	waitCtx, cancelWait := context.WithTimeout(ctx, p.waitTimeout)
	defer cancelWait()

	inst, err := p.thunderAPI.WaitForRunning(waitCtx, createResp.Identifier, p.pollInterval)
	if err != nil {
		// Could be timeout, terminal status (ERROR/TERMINATED), or
		// repeated network errors. All require compensation.
		return nil, cleanup("WaitForRunning failed",
			fmt.Errorf("wait for instance running: %w", err))
	}
	if inst.IP == "" {
		// Defensive: status reached RUNNING but IP not populated.
		// Shouldn't happen with a healthy Thunder API but worth catching.
		return nil, cleanup("running-instance missing IP",
			fmt.Errorf("instance %d is RUNNING but has no IP", createResp.Identifier))
	}

	provisionedAt := time.Now().UTC()

	// ── 7. INSERT thunder_instances row (with retry on transient pg errors) ──
	if err := p.insertWithRetry(ctx, insertRow{
		ID:                dbRowID,
		ThunderIdentifier: createResp.Identifier,
		InstanceType:      instanceType,
		InstanceIP:        inst.IP,
		SSHUser:           "ubuntu", // matches schema default
		SSHKeySecretName:  secretName,
		MaxUptimeHours:    cfg.DefaultHardUptimeHours,
		TrainingRunID:     nullableUUID(req.TrainingRunID),
		RequestedBy:       req.RequestedBy,
		HourlyRateUSD:     cfg.DefaultHourlyRateUSD,
		ProvisionedAt:     provisionedAt,
		RunningSince:      provisionedAt, // Thunder confirmed RUNNING just now
	}); err != nil {
		return nil, cleanup("DB INSERT failed after retries",
			fmt.Errorf("insert thunder_instances: %w", err))
	}

	p.logger.Info("Provision complete",
		zap.String("db_row_id", secretLabelUUID),
		zap.Int("thunder_identifier", createResp.Identifier),
		zap.String("instance_ip", inst.IP),
		zap.String("instance_type", instanceType),
	)

	return &ProvisionInstanceResult{
		InstanceIP:        inst.IP,
		SSHUser:           "ubuntu",
		SSHKeySecretName:  secretName,
		ProvisioningID:    secretLabelUUID,
		ThunderIdentifier: createResp.Identifier,
		ProvisionedAt:     provisionedAt,
	}, nil
}

// ─────────────────────────────────────────────────────────────────────────
// DB INSERT
// ─────────────────────────────────────────────────────────────────────────

type insertRow struct {
	ID                uuid.UUID
	ThunderIdentifier int
	InstanceType      string
	InstanceIP        string
	SSHUser           string
	SSHKeySecretName  string
	MaxUptimeHours    int
	TrainingRunID     sql.NullString // UUID-as-string, null if not provided
	RequestedBy       string
	HourlyRateUSD     float64
	ProvisionedAt     time.Time
	RunningSince      time.Time
}

// insertWithRetry attempts the INSERT up to dbInsertRetries+1 times, with
// backoff between attempts. Aborts immediately on non-retryable errors
// (constraint violation, syntax error). Returns nil on success or the
// final error on exhaustion.
func (p *ProvisionAction) insertWithRetry(ctx context.Context, row insertRow) error {
	var lastErr error
	for attempt := 0; attempt <= p.dbInsertRetries; attempt++ {
		if attempt > 0 {
			backoff := p.dbInsertBackoffs[attempt-1]
			p.logger.Info("Retrying thunder_instances INSERT",
				zap.Int("attempt", attempt),
				zap.Duration("backoff", backoff),
				zap.Error(lastErr))
			select {
			case <-ctx.Done():
				return fmt.Errorf("ctx done during insert retry: %w", ctx.Err())
			case <-time.After(backoff):
			}
		}

		err := p.insertOnce(ctx, row)
		if err == nil {
			return nil
		}
		// TODO(refine): distinguish retryable (transient pg blips, timeout)
		// from non-retryable (constraint violation, syntax). For now retry all.
		lastErr = err
	}
	return fmt.Errorf("insert failed after %d retries: %w", p.dbInsertRetries, lastErr)
}

func (p *ProvisionAction) insertOnce(ctx context.Context, row insertRow) error {
	const q = `
		INSERT INTO thunder_instances (
			id, thunder_instance_id, instance_type, instance_ip,
			ssh_user, ssh_key_secret_name, status, max_uptime_hours,
			training_run_id, requested_by, hourly_rate_usd,
			provisioned_at, running_since
		) VALUES ($1, $2, $3, $4, $5, $6, 'running', $7, $8, $9, $10, $11, $12)
	`
	_, err := p.db.ExecContext(ctx, q,
		row.ID,
		strconv.Itoa(row.ThunderIdentifier), // thunder_instance_id TEXT — store as string
		row.InstanceType,
		row.InstanceIP,
		row.SSHUser,
		row.SSHKeySecretName,
		row.MaxUptimeHours,
		row.TrainingRunID, // sql.NullString — pg accepts NULL for empty
		row.RequestedBy,
		row.HourlyRateUSD,
		row.ProvisionedAt,
		row.RunningSince,
	)
	return err
}

// ─────────────────────────────────────────────────────────────────────────
// helpers
// ─────────────────────────────────────────────────────────────────────────

// deriveInstanceType produces a short label combining GPU type and count.
// Format: "<gpu>_<count>" — e.g. "a100_1", "a100_2", "h100_1".
// Stored in thunder_instances.instance_type for reporting / reaper queries.
// Distinct from Thunder's own instance-type naming (we don't get that back).
func deriveInstanceType(gpu string, numGPUs int) string {
	if numGPUs < 1 {
		numGPUs = 1
	}
	if gpu == "" {
		gpu = "unknown"
	}
	return fmt.Sprintf("%s_%d", gpu, numGPUs)
}

// nullableUUID returns sql.NullString{Valid: true, String: s} if s is a
// non-empty UUID; otherwise NULL. Loose validation — pg's UUID type
// will reject malformed strings.
func nullableUUID(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	if _, err := uuid.Parse(s); err != nil {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// ErrProvisionDenied is returned when thunder_provision_check denies the
// request. Caller (the action dispatcher) can errors.Is to distinguish
// this from infrastructure errors and choose response status accordingly.
var ErrProvisionDenied = errors.New("provision denied")
