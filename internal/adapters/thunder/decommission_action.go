// FILE: internal/adapters/thunder/decommission_action.go
//
// decommission_instance action handler.
//
// Lifecycle:
//   1. Resolve the thunder_instances row (by provisioning_id OR thunder_identifier).
//   2. MarkDecommissioning — atomic state-transition + idempotency anchor.
//      If the row was already terminal (decommissioned/failed/reaped),
//      return success without further work.
//   3. Call Thunder API DeleteInstance (idempotent on 404).
//   4. Delete the k8s SSH Secret (idempotent on 404).
//   5. Compute cost from running_since to NOW().
//   6. MarkDecommissioned with computed cost.
//
// Idempotency: every step is safe to retry. A crashed mid-flight call
// can be re-driven (row already in 'decommissioning' state — steps 3-6
// run again, all of them no-op on already-clean resources).

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

	"github.com/gqls/agentchassis/internal/adapters/thunder/store"
)

// ─────────────────────────────────────────────────────────────────────────
// Request / Response shapes
// ─────────────────────────────────────────────────────────────────────────

// DecommissionInstanceRequest. Caller supplies ONE of the two identifiers.
// provisioning_id is preferred (we own that UUID and it's the natural
// handle); thunder_identifier is the fallback for reaper-driven cleanup
// where we discovered an instance via the Thunder API without a DB lookup.
type DecommissionInstanceRequest struct {
	ProvisioningID    string `json:"provisioning_id,omitempty"`    // DB row UUID
	ThunderIdentifier string `json:"thunder_identifier,omitempty"` // numeric ID as string
	Reason            string `json:"reason,omitempty"`             // optional audit
}

// DecommissionInstanceResult.
type DecommissionInstanceResult struct {
	ProvisioningID    string    `json:"provisioning_id"`
	ThunderIdentifier string    `json:"thunder_identifier"`
	CostUSD           float64   `json:"cost_usd"`
	DecommissionedAt  time.Time `json:"decommissioned_at"`
	WasAlreadyDone    bool      `json:"was_already_done"` // true if no API/cleanup work happened
}

// ─────────────────────────────────────────────────────────────────────────
// DecommissionAction
// ─────────────────────────────────────────────────────────────────────────

// DecommissionAction holds dependencies. Reuses the same thunderAPI and
// secretManager interfaces defined in provision_action.go — both actions
// need the same operations on the same things.
type DecommissionAction struct {
	thunderAPI thunderAPI
	secretMgr  secretManager
	db         *sql.DB
	logger     *zap.Logger
}

// NewDecommissionAction builds a DecommissionAction.
func NewDecommissionAction(api thunderAPI, mgr secretManager, db *sql.DB, logger *zap.Logger) *DecommissionAction {
	return &DecommissionAction{
		thunderAPI: api,
		secretMgr:  mgr,
		db:         db,
		logger:     logger.Named("decommission_instance"),
	}
}

// Execute runs the decommission flow. Idempotent end-to-end.
func (d *DecommissionAction) Execute(
	ctx context.Context,
	req DecommissionInstanceRequest,
) (*DecommissionInstanceResult, error) {
	// ── 1. Resolve the row ──
	inst, err := d.resolveInstance(ctx, req)
	if err != nil {
		return nil, err
	}

	// ── 2. State transition + idempotency check ──
	alreadyTerminal, err := store.MarkDecommissioning(ctx, d.db, inst.ID)
	if err != nil {
		return nil, fmt.Errorf("mark decommissioning: %w", err)
	}
	if alreadyTerminal {
		// Already done — return existing state without touching APIs.
		d.logger.Info("Instance already decommissioned, returning cached result",
			zap.String("provisioning_id", inst.ID.String()),
			zap.String("thunder_identifier", inst.ThunderInstanceID),
			zap.String("status", inst.Status))
		decommAt := time.Now().UTC()
		if inst.DecommissionedAt.Valid {
			decommAt = inst.DecommissionedAt.Time
		}
		var cost float64
		if inst.CostUSD.Valid {
			cost = inst.CostUSD.Float64
		}
		return &DecommissionInstanceResult{
			ProvisioningID:    inst.ID.String(),
			ThunderIdentifier: inst.ThunderInstanceID,
			CostUSD:           cost,
			DecommissionedAt:  decommAt,
			WasAlreadyDone:    true,
		}, nil
	}

	// ── 3. Delete the Thunder instance ──
	// thunder_instance_id is TEXT in our schema — parse back to int for the API.
	identifierInt, parseErr := strconv.Atoi(inst.ThunderInstanceID)
	if parseErr != nil {
		// Schema integrity violation — the DB has something that isn't
		// a parseable integer. Surface this; manual cleanup needed.
		return nil, fmt.Errorf("thunder_instance_id %q is not parseable as int — manual cleanup required: %w",
			inst.ThunderInstanceID, parseErr)
	}
	if err := d.thunderAPI.DeleteInstance(ctx, identifierInt); err != nil {
		// Note: api.Client.DeleteInstance already treats 404 as success.
		// Any error reaching here is a real failure (auth, 5xx, network).
		return nil, fmt.Errorf("thunder api delete (instance %d): %w", identifierInt, err)
	}

	// ── 4. Delete the SSH Secret ──
	// ssh.SecretManager.DeleteKeypairSecret is idempotent on 404 — safe.
	if inst.SSHKeySecretName != "" {
		if err := d.secretMgr.DeleteKeypairSecret(ctx, inst.SSHKeySecretName); err != nil {
			// Don't fail the whole decommission on a Secret delete error —
			// the instance is gone from Thunder (billable resource cleaned),
			// the orphan Secret is non-billable, and the reaper can clean
			// it up later. Log and continue.
			d.logger.Warn("SSH Secret delete failed — proceeding with status update; reaper will clean up",
				zap.String("secret_name", inst.SSHKeySecretName),
				zap.Error(err))
		}
	}

	// ── 5. Compute accrued cost ──
	decommAt := time.Now().UTC()
	cost := inst.ComputeCost(decommAt)

	// ── 6. Finalize the row ──
	if err := store.MarkDecommissioned(ctx, d.db, inst.ID, cost, decommAt); err != nil {
		// The instance and Secret are gone — finalization is the only
		// remaining work. If this errors, the row stays in 'decommissioning'
		// status; a later retry of this action will land on the alreadyTerminal=false
		// branch again and call delete-API/delete-Secret idempotently before
		// reaching this UPDATE. Acceptable.
		return nil, fmt.Errorf("mark decommissioned: %w", err)
	}

	d.logger.Info("Decommission complete",
		zap.String("provisioning_id", inst.ID.String()),
		zap.String("thunder_identifier", inst.ThunderInstanceID),
		zap.Float64("cost_usd", cost),
		zap.Duration("uptime", decommAt.Sub(zeroIfInvalid(inst.RunningSince))),
		zap.String("reason", req.Reason),
	)

	return &DecommissionInstanceResult{
		ProvisioningID:    inst.ID.String(),
		ThunderIdentifier: inst.ThunderInstanceID,
		CostUSD:           cost,
		DecommissionedAt:  decommAt,
		WasAlreadyDone:    false,
	}, nil
}

// ─────────────────────────────────────────────────────────────────────────
// helpers
// ─────────────────────────────────────────────────────────────────────────

// resolveInstance picks the right lookup based on which identifier the
// caller provided. provisioning_id wins if both are set.
func (d *DecommissionAction) resolveInstance(
	ctx context.Context,
	req DecommissionInstanceRequest,
) (*store.Instance, error) {
	if req.ProvisioningID != "" {
		id, err := uuid.Parse(req.ProvisioningID)
		if err != nil {
			return nil, fmt.Errorf("provisioning_id %q is not a valid UUID: %w",
				req.ProvisioningID, err)
		}
		inst, err := store.LookupByID(ctx, d.db, id)
		if err != nil {
			if errors.Is(err, store.ErrInstanceNotFound) {
				return nil, fmt.Errorf("no thunder_instances row with id %s", req.ProvisioningID)
			}
			return nil, err
		}
		return inst, nil
	}
	if req.ThunderIdentifier != "" {
		inst, err := store.LookupByThunderIdentifier(ctx, d.db, req.ThunderIdentifier)
		if err != nil {
			if errors.Is(err, store.ErrInstanceNotFound) {
				return nil, fmt.Errorf("no thunder_instances row with thunder_instance_id %s", req.ThunderIdentifier)
			}
			return nil, err
		}
		return inst, nil
	}
	return nil, fmt.Errorf("decommission_instance: must supply provisioning_id or thunder_identifier")
}

// zeroIfInvalid is a tiny helper for the structured log line.
func zeroIfInvalid(t sql.NullTime) time.Time {
	if t.Valid {
		return t.Time
	}
	return time.Time{}
}
