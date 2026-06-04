// FILE: platform/orchestration/actions/classify_training_probe_action.go
//
// classify_training_probe: reads the result of a prior ssh_get_status probe
// step (dispatch_thunder_ssh_get_status) and decides what the training-monitor
// should do with the instance — routing the workflow via a next_step override.
//
// This is pure logic: it does NOT touch the DB or Thunder. The terminal DB write
// (mark_training_run_terminal) and the release (dispatch_thunder_decommission)
// are separate downstream steps that this action routes TO. Keeping it pure makes
// it unit-testable and keeps "complexity in Go, workflow simple": the workflow is
// probe -> classify -> {mark_complete|mark_failed|done} -> [decommission] -> done.
//
// How it reads the probe result:
//   The probe step's reply lands in CollectedData under that step's NAME, with the
//   adapter body under ".response" (the await-result wrapper). So with probe_step
//   "probe", the fields are at probe.response.{stdout,reachable,exit_code}.
//   ExtractNestedField auto-unwraps .response per segment, so "probe.stdout"
//   resolves to probe.response.stdout. NOTE: reachable is a bool — read it via
//   ExtractNestedField (interface{}) and type-assert; ExtractNestedFieldString
//   returns "" for a non-string and would silently lose it.
//
// The probe command (the status_command sent to ssh_get_status) echoes exactly
// one of: STATUS=ALIVE | STATUS=DONE_OK | STATUS=DONE_FAIL | STATUS=GONE_UNKNOWN.
// Those map to run.sh's real markers: a running train process -> ALIVE; RUN_SH_DONE
// + adapter_config.json present -> DONE_OK; RUN_SH_FATAL -> DONE_FAIL; process gone
// with neither marker (mid-train crash/OOM/abort) -> GONE_UNKNOWN. If the box is
// unreachable the command never runs and no STATUS= is emitted; reachable=false is
// the signal, surfaced here as verdict "unreachable".
//
// Config (all step targets are config so the action stays decoupled from specific
// workflow step names):
//   - probe_step:       name of the ssh_get_status step to read (default "probe")
//   - complete_step:    step to route to on DONE_OK            (required for DONE_OK)
//   - failed_step:      step to route to on DONE_FAIL/GONE_UNKNOWN (required for those)
//   - alive_step:       step to route to on ALIVE (default: the configured NextStep)
//   - unreachable_step: step to route to on unreachable / no-status
//                       (default: alive_step, i.e. leave it for the next tick)
//
// DEFERRED (needs cross-tick state, not in this action): "unreachable for N
// consecutive ticks -> lost -> decommission". A single probe can't count ticks and
// thunder_instances has no consecutive-unreachable counter column. For now an
// unreachable box is left for the next scheduler tick to re-probe; the time-reaper
// (per-instance max_uptime_hours) remains the backstop for a truly-lost box.

package actions

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// Probe verdicts. The first four mirror the STATUS= tokens emitted by the probe
// command; the last two are synthesized here from reachability / missing output.
const (
	verdictAlive       = "alive"
	verdictDoneOK      = "done_ok"
	verdictDoneFail    = "done_fail"
	verdictGoneUnknown = "gone_unknown"
	verdictUnreachable = "unreachable"
	verdictNoStatus    = "no_status" // reachable, but probe emitted no STATUS= token
)

// ClassifyTrainingProbeAction inspects a prior ssh_get_status result and returns
// a verdict plus a next_step override routing the monitor sub-agent's workflow.
func ClassifyTrainingProbeAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "classify_training_probe"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	probeStep := configOrInput(params, "probe_step")
	if probeStep == "" {
		probeStep = "probe"
	}

	// stdout is a string; reachable is a bool (read via ExtractNestedField, then
	// type-assert — ExtractNestedFieldString returns "" for non-strings).
	stdout := datahelpers.ExtractNestedFieldString(params.CollectedData, probeStep+".stdout")
	reachable, reachableKnown := datahelpers.ExtractNestedField(params.CollectedData, probeStep+".reachable").(bool)

	rawStatus := parseStatusToken(stdout)
	verdict := classifyVerdict(rawStatus, reachable, reachableKnown)

	// Resolve the routing target for this verdict from config.
	aliveStep := configOrInput(params, "alive_step") // default handled below
	unreachableStep := configOrInput(params, "unreachable_step")
	if unreachableStep == "" {
		unreachableStep = aliveStep // leave it; next tick re-probes
	}

	var nextStep string
	switch verdict {
	case verdictAlive:
		nextStep = aliveStep // empty -> fall through to the step's configured NextStep
	case verdictUnreachable, verdictNoStatus:
		nextStep = unreachableStep
	case verdictDoneOK:
		nextStep = configOrInput(params, "complete_step")
		if nextStep == "" {
			return nil, fmt.Errorf("classify_training_probe: verdict %q requires complete_step in config", verdict)
		}
	case verdictDoneFail, verdictGoneUnknown:
		nextStep = configOrInput(params, "failed_step")
		if nextStep == "" {
			return nil, fmt.Errorf("classify_training_probe: verdict %q requires failed_step in config", verdict)
		}
	}

	logger.Info("Classified training probe",
		zap.String("probe_step", probeStep),
		zap.String("raw_status", rawStatus),
		zap.Bool("reachable", reachable),
		zap.Bool("reachable_known", reachableKnown),
		zap.String("verdict", verdict),
		zap.String("routed_next_step", nextStep),
	)

	result := map[string]interface{}{
		"verdict":         verdict,
		"raw_status":      rawStatus,
		"ssh_reachable":   reachable && reachableKnown,
		"is_terminal":     verdict == verdictDoneOK || verdict == verdictDoneFail || verdict == verdictGoneUnknown,
		"should_complete": verdict == verdictDoneOK,
	}
	// Only emit a next_step override when we actually chose a target; an empty
	// string here would be ignored by getNextStepFromResult, leaving the step's
	// configured NextStep in effect (the intended "leave it" path for ALIVE).
	if nextStep != "" {
		result["next_step"] = nextStep
	}
	return result, nil
}

// classifyVerdict maps the probe's STATUS= token plus reachability into a verdict.
// An unreachable box (command never ran) wins over a missing token; a reachable
// box with a recognized token takes the token; reachable-but-no-token is a
// distinct "no_status" (treated as leave-and-retry, like unreachable).
func classifyVerdict(rawStatus string, reachable, reachableKnown bool) string {
	if reachableKnown && !reachable {
		return verdictUnreachable
	}
	switch rawStatus {
	case "ALIVE":
		return verdictAlive
	case "DONE_OK":
		return verdictDoneOK
	case "DONE_FAIL":
		return verdictDoneFail
	case "GONE_UNKNOWN":
		return verdictGoneUnknown
	case "":
		// Reachable (or reachability unknown) but the probe emitted no STATUS=.
		// Don't guess a terminal state off a blank read — leave it for re-probe.
		return verdictNoStatus
	default:
		// Unrecognized token — treat conservatively as no_status, don't decommission.
		return verdictNoStatus
	}
}

// parseStatusToken extracts the token after the first "STATUS=" in the probe
// stdout, up to the next whitespace. Returns "" if absent.
func parseStatusToken(stdout string) string {
	const marker = "STATUS="
	idx := strings.Index(stdout, marker)
	if idx < 0 {
		return ""
	}
	rest := stdout[idx+len(marker):]
	end := strings.IndexFunc(rest, func(r rune) bool {
		return r == ' ' || r == '\n' || r == '\t' || r == '\r'
	})
	if end < 0 {
		return strings.TrimSpace(rest)
	}
	return strings.TrimSpace(rest[:end])
}
