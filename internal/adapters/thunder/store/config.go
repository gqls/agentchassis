// FILE: internal/adapters/thunder/store/config.go
//
// Database access for thunder_config singleton and thunder_provision_check
// view. Pure transport over *sql.DB — no business logic.
//
// Used by:
//   - provision_instance action handler (defaults + pre-check)
//   - decommission_instance action handler (none of these — see instances.go)
//   - thunder-reaper scheduled task (config for hard_uptime_hours)

package store

import (
	"context"
	"database/sql"
	"fmt"
)

// Config mirrors the thunder_config singleton row.
// All fields are populated from the DB; no defaulting in Go.
type Config struct {
	DailyCapUSD            float64
	MaxConcurrentInstances int
	DefaultHardUptimeHours int
	DefaultHourlyRateUSD   float64
	EstimatedNewRunCostUSD float64
	IsPaused               bool
	PauseReason            sql.NullString
}

// ProvisionCheck mirrors thunder_provision_check view shape.
type ProvisionCheck struct {
	CanProvision           bool
	DenialReason           sql.NullString
	DailyCapUSD            float64
	MaxConcurrentInstances int
	Total24hSpend          float64
	ActiveCount            int
}

// LoadConfig returns the (singleton) row from thunder_config.
// Errors if the row is missing — that indicates migration 025 wasn't applied.
func LoadConfig(ctx context.Context, db *sql.DB) (*Config, error) {
	const q = `
		SELECT daily_cap_usd, max_concurrent_instances,
		       default_hard_uptime_hours, default_hourly_rate_usd,
		       estimated_new_run_cost_usd, is_paused, pause_reason
		FROM thunder_config
		LIMIT 1
	`
	var c Config
	err := db.QueryRowContext(ctx, q).Scan(
		&c.DailyCapUSD,
		&c.MaxConcurrentInstances,
		&c.DefaultHardUptimeHours,
		&c.DefaultHourlyRateUSD,
		&c.EstimatedNewRunCostUSD,
		&c.IsPaused,
		&c.PauseReason,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("thunder_config row missing — migration 025 not applied?")
	}
	if err != nil {
		return nil, fmt.Errorf("load thunder_config: %w", err)
	}
	return &c, nil
}

// CheckCanProvision queries thunder_provision_check view. The view encodes
// the cap logic (daily spend + concurrent count); we trust it rather than
// reproducing the rules in Go.
//
// Returns (true, nil) if provisioning is allowed.
// Returns (false, "<reason>") if denied.
// Returns (_, err) only on database errors.
func CheckCanProvision(ctx context.Context, db *sql.DB) (allowed bool, reason string, _ error) {
	const q = `
		SELECT can_provision, denial_reason,
		       daily_cap_usd, max_concurrent_instances,
		       total_24h_spend, active_count
		FROM thunder_provision_check
	`
	var pc ProvisionCheck
	err := db.QueryRowContext(ctx, q).Scan(
		&pc.CanProvision,
		&pc.DenialReason,
		&pc.DailyCapUSD,
		&pc.MaxConcurrentInstances,
		&pc.Total24hSpend,
		&pc.ActiveCount,
	)
	if err == sql.ErrNoRows {
		return false, "thunder_provision_check view returned no rows", nil
	}
	if err != nil {
		return false, "", fmt.Errorf("provision pre-check: %w", err)
	}
	if !pc.CanProvision {
		r := "denied"
		if pc.DenialReason.Valid {
			r = pc.DenialReason.String
		}
		return false, fmt.Sprintf("%s (spend=$%.2f/cap=$%.2f, active=%d/max=%d)",
			r, pc.Total24hSpend, pc.DailyCapUSD, pc.ActiveCount, pc.MaxConcurrentInstances), nil
	}
	return true, "", nil
}
