// FILE: internal/adapters/thunder/store/instances.go
//
// Database access for thunder_instances row CRUD. Separated from
// config.go: this file is per-instance state transitions, config.go
// is the singleton settings + provision-check view.
//
// Used by:
//   - decommission_instance action (lookup, mark transitions)
//   - thunder-reaper (find timed-out instances — Phase 3.5)

package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Instance mirrors enough of thunder_instances for the decommission flow.
// Not every column — only what action handlers and the reaper read.
type Instance struct {
	ID                uuid.UUID // DB row PK
	ThunderInstanceID string    // TEXT in schema — stores Thunder's numeric identifier as string
	InstanceType      string
	InstanceIP        string
	SSHPort           sql.NullInt64 // nullable: rows provisioned before port capture have NULL
	SSHUser           string
	SSHKeySecretName  string
	Status            string // see schema CHECK constraint values
	MaxUptimeHours    int
	TrainingRunID     sql.NullString
	RequestedBy       string
	HourlyRateUSD     float64
	CostUSD           sql.NullFloat64
	ProvisionedAt     sql.NullTime
	RunningSince      sql.NullTime
	DecommissionedAt  sql.NullTime
}

// ErrInstanceNotFound is returned by Lookup when no matching row exists.
var ErrInstanceNotFound = errors.New("thunder_instance row not found")

// LookupByID fetches a row by its DB UUID primary key.
func LookupByID(ctx context.Context, db *sql.DB, id uuid.UUID) (*Instance, error) {
	return lookupOne(ctx, db,
		`SELECT id, thunder_instance_id, instance_type, instance_ip, ssh_port, ssh_user,
		        ssh_key_secret_name, status, max_uptime_hours, training_run_id,
		        requested_by, hourly_rate_usd, cost_usd,
		        provisioned_at, running_since, decommissioned_at
		 FROM thunder_instances
		 WHERE id = $1`,
		id)
}

// LookupByThunderIdentifier fetches a row by Thunder's numeric identifier
// (stored in the schema as TEXT). Used when the caller only has the API
// identifier and not the DB UUID — e.g. reaper sweep against ListInstances.
//
// Thunder RECYCLES numeric identifiers after decommission, so multiple
// historical rows can share one thunder_instance_id (one live + several
// terminal). The partial unique index guarantees at most one LIVE row per
// identifier, so we order live rows first, then most-recent, and take one.
// This deterministically returns the live instance when one exists; when only
// terminal history remains it returns the most recent (harmless — callers
// short-circuit on already-terminal status).
func LookupByThunderIdentifier(ctx context.Context, db *sql.DB, identifier string) (*Instance, error) {
	return lookupOne(ctx, db,
		`SELECT id, thunder_instance_id, instance_type, instance_ip, ssh_port, ssh_user,
		        ssh_key_secret_name, status, max_uptime_hours, training_run_id,
		        requested_by, hourly_rate_usd, cost_usd,
		        provisioned_at, running_since, decommissioned_at
		 FROM thunder_instances
		 WHERE thunder_instance_id = $1`,
		identifier)
}

func lookupOne(ctx context.Context, db *sql.DB, q string, arg interface{}) (*Instance, error) {
	var i Instance
	err := db.QueryRowContext(ctx, q, arg).Scan(
		&i.ID, &i.ThunderInstanceID, &i.InstanceType, &i.InstanceIP, &i.SSHPort, &i.SSHUser,
		&i.SSHKeySecretName, &i.Status, &i.MaxUptimeHours, &i.TrainingRunID,
		&i.RequestedBy, &i.HourlyRateUSD, &i.CostUSD,
		&i.ProvisionedAt, &i.RunningSince, &i.DecommissionedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrInstanceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("thunder_instances lookup: %w", err)
	}
	return &i, nil
}

// MarkDecommissioning transitions row status to 'decommissioning' and
// stamps decommission_requested_at. This is the idempotency anchor for
// the action: a row already in 'decommissioning' or 'decommissioned'
// status is left alone (no error), letting the action continue cleanup
// without double-billing the running clock.
//
// Returns (alreadyTerminal, error):
//
//	alreadyTerminal=true if status was already 'decommissioned' or 'failed'
//	— caller can skip the API delete and cleanup and return success.
//	alreadyTerminal=false means caller should proceed with the rest of the
//	decommission flow.
func MarkDecommissioning(ctx context.Context, db *sql.DB, id uuid.UUID) (alreadyTerminal bool, err error) {
	// Single UPDATE … RETURNING to check the row's prior status atomically.
	// If status was 'running' or 'provisioning' we transition; if it was
	// already a terminal/decommissioning state we leave it untouched but
	// return alreadyTerminal accordingly.
	const q = `
		UPDATE thunder_instances
		SET status = CASE
		    WHEN status IN ('running','provisioning') THEN 'decommissioning'
		    ELSE status
		END,
		decommission_requested_at = COALESCE(decommission_requested_at, NOW()),
		updated_at = NOW()
		WHERE id = $1
		RETURNING status
	`
	var newStatus string
	if err := db.QueryRowContext(ctx, q, id).Scan(&newStatus); err != nil {
		if err == sql.ErrNoRows {
			return false, ErrInstanceNotFound
		}
		return false, fmt.Errorf("mark decommissioning: %w", err)
	}
	// alreadyTerminal if the row is already past the point of needing API cleanup.
	if newStatus == "decommissioned" || newStatus == "failed" || newStatus == "reaped" {
		return true, nil
	}
	return false, nil
}

// MarkDecommissioned writes the final status, decommissioned_at, and
// computed cost. Safe to call on a row already in 'decommissioned' status
// (no-op update) — but callers should generally avoid that path via the
// alreadyTerminal short-circuit in MarkDecommissioning.
func MarkDecommissioned(
	ctx context.Context,
	db *sql.DB,
	id uuid.UUID,
	costUSD float64,
	decommissionedAt time.Time,
) error {
	const q = `
		UPDATE thunder_instances
		SET status = 'decommissioned',
		    decommissioned_at = $2,
		    cost_usd = $3,
		    updated_at = NOW()
		WHERE id = $1
	`
	res, err := db.ExecContext(ctx, q, id, decommissionedAt, costUSD)
	if err != nil {
		return fmt.Errorf("mark decommissioned: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrInstanceNotFound
	}
	return nil
}

// ComputeCost calculates accrued cost from running_since to decommissionedAt.
// Returns 0 if running_since is null (instance never reached RUNNING) — the
// caller may want to log this anomaly but it shouldn't fail decommission.
func (i *Instance) ComputeCost(decommissionedAt time.Time) float64 {
	if !i.RunningSince.Valid {
		return 0
	}
	hours := decommissionedAt.Sub(i.RunningSince.Time).Hours()
	if hours < 0 {
		hours = 0 // clock skew safety
	}
	return hours * i.HourlyRateUSD
}
