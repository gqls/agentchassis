// FILE: platform/orchestration/actions/record_probe_streak_action.go
//
// record_probe_streak: maintains the per-instance consecutive-unreachable probe
// counter on public.thunder_instances (migration 106) for the
// thunder-training-monitor, and routes the workflow accordingly via a next_step
// override. This is the "own step" that honors "unreachable for N consecutive
// ticks -> lost -> decommission" without polluting the pure classifier.
//
// Two modes (set in step config):
//   mode=reset : a reachable probe (ALIVE) — zero the counter, then route to
//                ok_step (leave the box running, end this tick).
//   mode=bump  : an unreachable / no-status probe — increment the counter and
//                read the new value; if it has reached unreachable_threshold,
//                route to lost_step (the shared mark_failed -> decommission
//                path); otherwise route to ok_step (leave it; next tick re-probes).
//
// Placement in the monitor sub-agent workflow:
//   classify --alive--------------> reset_streak (mode=reset, ok_step=done) -> done
//   classify --unreachable/no_status-> bump_streak (mode=bump,
//                                        unreachable_threshold=N,
//                                        lost_step=mark_failed, ok_step=done)
//                                        -> {mark_failed -> decommission -> done | done}
// (terminal verdicts done_ok/done_fail/gone_unknown are routed straight to the
//  mark steps by the classifier and never reach here.)
//
// Config:
//   - mode                 : "bump" | "reset"   (required)
//   - provisioning_id      : thunder_instances.id (uuid) — from input_data via
//                            input_mapping (required)
//   - unreachable_threshold: int, bump only (default 3) — a JSON number in config,
//                            read via datahelpers.GetIntField
//   - lost_step            : step to route to when the threshold is reached
//                            (required for the lost transition)
//   - ok_step              : step to route to otherwise / after reset (optional;
//                            empty leaves the step's configured NextStep in effect)
//
// trg_thunder_instances_updated_at maintains updated_at on UPDATE, so we set
// only last_probe_at and the counter here.

package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

const defaultUnreachableThreshold = 3

var RecordProbeStreakInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"provisioning_id"},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("record_probe_streak", RecordProbeStreakInputSpec)
}

// RecordProbeStreakAction bumps or resets the consecutive-unreachable counter and
// returns a next_step override.
func RecordProbeStreakAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "record_probe_streak"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("record_probe_streak: database connection required")
	}

	mode := strings.ToLower(strings.TrimSpace(datahelpers.GetStringField(params.StepConfig.Config, "mode", "")))
	if mode != "bump" && mode != "reset" {
		return nil, fmt.Errorf("record_probe_streak: mode must be \"bump\" or \"reset\", got %q", mode)
	}

	// provisioning_id is an input_data value (set by the orchestrator's call_agent
	// at spawn) -> resolve via ExtractActionInputs, not a direct config read.
	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		RecordProbeStreakInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("record_probe_streak: input extraction failed: %w", err)
	}
	instanceID, err := uuid.Parse(inputs.Get("provisioning_id"))
	if err != nil {
		return nil, fmt.Errorf("record_probe_streak: invalid/missing provisioning_id: %w", err)
	}

	okStep := datahelpers.GetStringField(params.StepConfig.Config, "ok_step", "")

	if mode == "reset" {
		if _, err := params.DB.ExecContext(ctx, `
			UPDATE public.thunder_instances
			   SET consecutive_unreachable_probes = 0,
			       last_probe_at = NOW()
			 WHERE id = $1
		`, instanceID); err != nil {
			return nil, fmt.Errorf("record_probe_streak: reset update: %w", err)
		}
		logger.Info("Reset unreachable streak",
			zap.String("provisioning_id", instanceID.String()),
			zap.String("routed_next_step", okStep),
		)
		result := map[string]interface{}{"mode": "reset", "streak": 0}
		if okStep != "" {
			result["next_step"] = okStep
		}
		return result, nil
	}

	// mode == "bump": increment and read the new value atomically.
	var streak int
	if err := params.DB.QueryRowContext(ctx, `
		UPDATE public.thunder_instances
		   SET consecutive_unreachable_probes = consecutive_unreachable_probes + 1,
		       last_probe_at = NOW()
		 WHERE id = $1
		RETURNING consecutive_unreachable_probes
	`, instanceID).Scan(&streak); err != nil {
		return nil, fmt.Errorf("record_probe_streak: bump update (instance may not exist): %w", err)
	}

	threshold := datahelpers.GetIntField(params.StepConfig.Config, "unreachable_threshold", defaultUnreachableThreshold)
	if threshold < 1 {
		threshold = defaultUnreachableThreshold
	}
	lost := streak >= threshold

	nextStep := okStep
	if lost {
		nextStep = datahelpers.GetStringField(params.StepConfig.Config, "lost_step", "")
		if nextStep == "" {
			return nil, fmt.Errorf("record_probe_streak: streak %d reached threshold %d but lost_step is not configured", streak, threshold)
		}
	}

	logger.Info("Bumped unreachable streak",
		zap.String("provisioning_id", instanceID.String()),
		zap.Int("streak", streak),
		zap.Int("threshold", threshold),
		zap.Bool("lost", lost),
		zap.String("routed_next_step", nextStep),
	)

	result := map[string]interface{}{
		"mode":      "bump",
		"streak":    streak,
		"threshold": threshold,
		"lost":      lost,
	}
	if nextStep != "" {
		result["next_step"] = nextStep
	}
	return result, nil
}
