// FILE: platform/orchestration/actions/diagnose_dormant_agents_action.go
//
// DORMANT-AGENTS — the capability-inventory sweep (bugs_open/044). The
// producer-side mirror of the silent-check checker: silent-check finds work
// that no worker ever touches; this finds a WORKER that no work ever routes to.
// Both are invisible because nothing fails — an active agent that never runs
// errors nowhere, so a capability can quietly go unused for months until a
// thread declares it missing (bugs_closed/002 D: section-editor, shipped, 3
// production runs, declared nonexistent through a handoff and two sign-offs).
// The platform had no inventory of its own capabilities and no detector for
// unused ones; this is that detector.
//
// DELIBERATELY DETERMINISTIC: no LLM. It measures; it never diagnoses and never
// fixes. Per sweep it:
//   - computes, by the STEP-FINGERPRINT method (a workflow step key belonging
//     to exactly ONE agent), which active non-snapshot agents have NEVER been
//     observed running in any orchestration_states.workflow_plan. owner_agent_
//     type is NOT usable (95k+ rows carry 'generic', the dispatch path not the
//     agent — the trap that produced the wrong "110", WRONG_CALLS 2026-07-20);
//   - applies an AGE FLOOR so a freshly-seeded agent that simply has not had a
//     chance to run yet is not flagged (evidence-researcher, seeded the day the
//     bug was filed, is new — not dormant);
//   - leaves the MIRRORED-AGENT BLIND SPOT unflagged by construction: an agent
//     with no step key unique to it (e.g. council-gate, whose 099 roster mirror
//     copies fix-proposer's steps verbatim) is unmeasurable this way, so it can
//     never enter the never-observed set. Those agents are listed in the report,
//     never emitted. (orchestration_name was evaluated as a second signal and
//     REJECTED: it is 'generic-orchestrate-<ts>', it does not name the agent.)
//   - emits ONE inert dormant_agent item per past-floor never-run agent
//     (status='dormant', pipeline='maintenance', unclaimable — nothing
//     dispatches it; a human triages wire/retire/paused), capped, deduped by
//     item_key via ON CONFLICT DO NOTHING;
//   - closes its own items ('complete') once the agent is observed to have run;
//   - writes one doc_note per sweep (categories dormant-agents+fixloop) — the
//     inventory artifact, which is the thing that makes a capability discoverable
//     before it is declared missing.
//
// KNOWN LIMITATION, stated honestly in the report: "never observed" means "no
// unique top-level workflow step seen in the RETAINED orchestration history"
// (orchestration_states is not kept forever). An agent that ran only before the
// window, or that runs only via a council/subtree path whose steps never surface
// as top-level plan keys (feature-designer is one — its own workflow has never
// run as an orchestration; its council approval went through other machinery),
// reads as never-observed. That is why this ships as a REPORT for human triage,
// not an auto-fix, and why the raw count is grouped, never asserted as "N bugs".
//
// Manual-trigger, ships dry_run=true (owner: more awareness before autonomy).
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// dormantSource labels everything this checker writes. Anchors to the
// system.internal pseudo-site (triageSystemSiteID) — a dormant agent is not a
// per-site fact, so it lives on the platform's own internal site, exactly as
// triage's needs_diagnosis items do.
const dormantSource = "diagnosis-dormant-agents"

// dormantAgent is one never-observed active agent, with the evidence that
// makes the finding checkable.
type dormantAgent struct {
	Type         string    // agent_definitions.type
	SampleStep   string    // one of its unique fingerprint steps (evidence)
	UniqueSteps  int        // how many step keys are unique to this agent
	ActiveRows   int        // active non-snapshot rows for this type (>1 is a hygiene flag)
	AgeDays      float64    // now() - min(created_at) across its active rows
	FirstCreated time.Time  // min(created_at) — when the capability first existed
}

var DiagnoseDormantAgentsInputSpec = datahelpers.ActionInputSpec{
	Required: []string{},
	Optional: []string{
		"age_floor_days", "max_emit", "dry_run",
	},
	Defaults: map[string]interface{}{
		"age_floor_days": 14,    // a fortnight's grace to have run at least once
		"max_emit":       10,    // cap per sweep while confidence builds (mirrors triage/silent-check)
		"dry_run":        false, // when true: report only — no items written, none closed
	},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("diagnose_dormant_agents", DiagnoseDormantAgentsInputSpec)
}

// DiagnoseDormantAgentsAction runs one capability-inventory sweep.
func DiagnoseDormantAgentsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	logger := params.Logger.With(zap.String("action", "diagnose_dormant_agents"))
	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("diagnose_dormant_agents: no DB handle")
	}

	ageFloor := datahelpers.GetIntField(config, "age_floor_days", 14)
	if ageFloor < 0 {
		ageFloor = 14
	}
	maxEmit := datahelpers.GetIntField(config, "max_emit", 10)
	dryRun := false
	if d, ok := config["dry_run"].(bool); ok {
		dryRun = d
	}

	// Gather: the never-observed set + the summary stats + the blind-spot list +
	// the retention window (for the report's honesty note).
	never, err := dormantGatherNeverObserved(ctx, params.DB)
	if err != nil {
		return nil, fmt.Errorf("gather never-observed: %w", err)
	}
	stats, err := dormantGatherStats(ctx, params.DB)
	if err != nil {
		return nil, fmt.Errorf("gather stats: %w", err)
	}
	blindSpot, err := dormantGatherBlindSpot(ctx, params.DB)
	if err != nil {
		return nil, fmt.Errorf("gather blind spot: %w", err)
	}
	oldestObserved, err := dormantOldestObservedRun(ctx, params.DB)
	if err != nil {
		return nil, fmt.Errorf("gather retention window: %w", err)
	}

	// Split by the age floor: only past-floor agents are eligible to EMIT; the
	// report shows both (nothing is hidden — an under-floor agent is a candidate
	// that simply has not aged in yet).
	past, under := dormantPartition(never, float64(ageFloor))

	// Emit (live only): one item per past-floor agent, newest-first so the
	// freshest genuinely-unused capabilities surface before the legacy backlog,
	// capped, deduped by ON CONFLICT (status='dormant' is inside idx_swi_dedup,
	// so ON CONFLICT dedups cleanly — unlike silent-check's 'failed' items).
	created, deduped, capped := 0, 0, 0
	var emittedKeys []string
	if !dryRun {
		for _, a := range dormantEmitOrder(past) {
			if created >= maxEmit {
				capped++
				continue
			}
			ins, err := dormantInsertItem(ctx, params.DB, a, ageFloor)
			if err != nil {
				logger.Warn("dormant-agents: emit failed", zap.String("agent_type", a.Type), zap.Error(err))
				continue
			}
			if ins {
				created++
				emittedKeys = append(emittedKeys, dormantItemKey(a.Type))
			} else {
				deduped++
			}
		}
	}

	// Close-out: any of our still-dormant items whose agent is no longer in the
	// never-observed set (it ran, or was deactivated/deleted) gets honestly
	// completed. Being non-'dormant', a completed row never blocks a future
	// re-emission if the agent somehow reads dormant again.
	stillDormant := map[string]bool{}
	for _, a := range never {
		stillDormant[a.Type] = true
	}
	closed := 0
	if !dryRun {
		closed, err = dormantCloseResolved(ctx, params.DB, stillDormant)
		if err != nil {
			logger.Warn("dormant-agents: close-resolved failed", zap.Error(err))
		}
	}

	now := time.Now().UTC()
	report := renderDormantAgents(now, ageFloor, stats, past, under, blindSpot, oldestObserved,
		created, deduped, capped, maxEmit, closed, dryRun)

	noteID, nerr := insertDocNote(ctx, params.DB, "pipeline", "diagnose", "",
		report, `["dormant-agents","fixloop"]`, "diagnose_dormant_agents", nullSafeAgentType(params), "", "diagnose_dormant_agents")
	if nerr != nil {
		logger.Warn("dormant-agents: doc_note persist failed (report still returned)", zap.Error(nerr))
	}

	logger.Info("diagnose_dormant_agents: swept",
		zap.Int("active_with_workflow", stats.ActiveWithWorkflow),
		zap.Int("measurable", stats.Measurable),
		zap.Int("blind_spot", stats.BlindSpot),
		zap.Int("never_observed", len(never)),
		zap.Int("past_floor", len(past)),
		zap.Int("under_floor", len(under)),
		zap.Int("emitted", created),
		zap.Int("deduped", deduped),
		zap.Int("capped", capped),
		zap.Int("closed_resolved", closed),
		zap.Bool("dry_run", dryRun),
		zap.String("note_id", noteID),
		zap.String("orchestration_id", orchIDForLog(params)))

	return map[string]interface{}{
		"report":               report,
		"active_with_workflow": stats.ActiveWithWorkflow,
		"measurable":           stats.Measurable,
		"blind_spot":           stats.BlindSpot,
		"never_observed":       len(never),
		"past_floor":           len(past),
		"under_floor":          len(under),
		"emitted":              created,
		"deduped":              deduped,
		"capped":               capped,
		"closed_resolved":      closed,
		"emitted_keys":         emittedKeys,
		"note_id":              noteID,
	}, nil
}

// dormantStats holds the sweep's headline counts.
type dormantStats struct {
	ActiveWithWorkflow int // active non-snapshot agents with a workflow
	Measurable         int // …with ≥1 unique fingerprint step
	BlindSpot          int // …with NO unique step (unmeasurable this way)
}

// dormantObservedCTE / dormantAgentStepsCTE / dormantFingerprintsCTE are the
// shared building blocks of the fingerprint method, kept as one source of truth
// so the gather and the stats queries cannot drift apart.
const dormantObservedCTE = `
	observed_steps AS (
		SELECT DISTINCT jsonb_object_keys(workflow_plan->'steps') AS step
		FROM orchestration_states
		WHERE workflow_plan ? 'steps'
		  AND jsonb_typeof(workflow_plan->'steps') = 'object'
	)`

const dormantAgentStepsCTE = `
	agent_steps AS (
		SELECT a.type, jsonb_object_keys(a.default_config#>'{workflow,steps}') AS step
		FROM agent_definitions a
		WHERE a.is_active AND a.deleted_at IS NULL AND COALESCE(a.is_snapshot,false)=false
		  AND jsonb_typeof(a.default_config#>'{workflow,steps}') = 'object'
	)`

// fingerprints: step keys unique to exactly ONE agent type.
const dormantFingerprintsCTE = `
	fingerprints AS (
		SELECT step, min(type) AS type FROM agent_steps
		GROUP BY step HAVING count(DISTINCT type) = 1
	)`

// dormantGatherNeverObserved returns the measurable agents whose unique
// fingerprint steps were NEVER observed as a top-level step in any orchestration
// plan, with the evidence and age needed to report and (past the floor) emit.
func dormantGatherNeverObserved(ctx context.Context, db *sql.DB) ([]dormantAgent, error) {
	query := `
	WITH` + dormantObservedCTE + `,` + dormantAgentStepsCTE + `,` + dormantFingerprintsCTE + `,
	agent_fp AS (
		SELECT f.type,
		       count(*) AS unique_steps,
		       count(*) FILTER (WHERE f.step IN (SELECT step FROM observed_steps)) AS observed_unique,
		       min(f.step) AS sample_step
		FROM fingerprints f
		GROUP BY f.type
	),
	never AS (
		SELECT type, unique_steps, sample_step
		FROM agent_fp
		WHERE observed_unique = 0
	)
	SELECT n.type, n.sample_step, n.unique_steps,
	       count(a.*) AS active_rows,
	       min(a.created_at) AS first_created,
	       round(extract(epoch FROM (now() - min(a.created_at))) / 86400.0, 1) AS age_days
	FROM never n
	JOIN agent_definitions a ON a.type = n.type
	  AND a.is_active AND a.deleted_at IS NULL AND COALESCE(a.is_snapshot,false)=false
	GROUP BY n.type, n.sample_step, n.unique_steps
	ORDER BY age_days DESC, n.type`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []dormantAgent
	for rows.Next() {
		var a dormantAgent
		if err := rows.Scan(&a.Type, &a.SampleStep, &a.UniqueSteps, &a.ActiveRows, &a.FirstCreated, &a.AgeDays); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// dormantGatherStats returns the three headline counts.
func dormantGatherStats(ctx context.Context, db *sql.DB) (dormantStats, error) {
	var s dormantStats
	query := `
	WITH` + dormantAgentStepsCTE + `,` + dormantFingerprintsCTE + `
	SELECT
		(SELECT count(DISTINCT type) FROM agent_steps) AS active_with_workflow,
		(SELECT count(DISTINCT type) FROM fingerprints) AS measurable,
		(SELECT count(DISTINCT type) FROM agent_steps
		  WHERE type NOT IN (SELECT type FROM fingerprints)) AS blind_spot`
	err := db.QueryRowContext(ctx, query).Scan(&s.ActiveWithWorkflow, &s.Measurable, &s.BlindSpot)
	return s, err
}

// dormantGatherBlindSpot lists the active agents with NO unique step key — the
// ones this method cannot measure and therefore never flags. Reported so the
// blind spot is visible, not silent.
func dormantGatherBlindSpot(ctx context.Context, db *sql.DB) ([]string, error) {
	query := `
	WITH` + dormantAgentStepsCTE + `,` + dormantFingerprintsCTE + `
	SELECT DISTINCT type FROM agent_steps
	WHERE type NOT IN (SELECT type FROM fingerprints)
	ORDER BY type`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// dormantOldestObservedRun returns the oldest orchestration in the retained
// history — the honest lower bound of "never observed since ...".
func dormantOldestObservedRun(ctx context.Context, db *sql.DB) (time.Time, error) {
	var t sql.NullTime
	err := db.QueryRowContext(ctx, `SELECT min(created_at) FROM orchestration_states`).Scan(&t)
	if err != nil {
		return time.Time{}, err
	}
	if !t.Valid {
		return time.Time{}, nil
	}
	return t.Time.UTC(), nil
}

// dormantPartition splits the never-observed set into agents past the age floor
// (eligible to emit) and under it (reported only). Pure — tested.
func dormantPartition(agents []dormantAgent, floorDays float64) (past, under []dormantAgent) {
	for _, a := range agents {
		if a.AgeDays >= floorDays {
			past = append(past, a)
		} else {
			under = append(under, a)
		}
	}
	return past, under
}

// dormantEmitOrder orders past-floor agents newest-first, so a capped sweep
// surfaces the freshest genuinely-unused capabilities before the legacy
// backlog. Pure — tested.
func dormantEmitOrder(past []dormantAgent) []dormantAgent {
	out := make([]dormantAgent, len(past))
	copy(out, past)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].AgeDays != out[j].AgeDays {
			return out[i].AgeDays < out[j].AgeDays // youngest first
		}
		return out[i].Type < out[j].Type
	})
	return out
}

// dormantItemKey is the STABLE per-agent dedup key.
func dormantItemKey(agentType string) string {
	return "dormant:" + agentType
}

func dormantSummary(a dormantAgent) string {
	return fmt.Sprintf("DORMANT: agent %q is active but its workflow has never been observed running (age %.0fd; %d unique step(s), e.g. %q). Human triage: wire it, retire it, or record it as paused.",
		a.Type, a.AgeDays, a.UniqueSteps, a.SampleStep)
}

func dormantSpecJSON(a dormantAgent, ageFloor int) string {
	b, _ := json.Marshal(map[string]interface{}{
		"agent_type":    a.Type,
		"sample_step":   a.SampleStep,
		"unique_steps":  a.UniqueSteps,
		"active_rows":   a.ActiveRows,
		"age_days":      a.AgeDays,
		"first_created": a.FirstCreated.UTC().Format(time.RFC3339),
		"age_floor":     ageFloor,
		"method":        "step-fingerprint: an active non-snapshot agent none of whose unique workflow step keys appears as a top-level orchestration_states.workflow_plan step; owner_agent_type is NOT used",
		"caveat":        "never-observed = never seen in RETAINED orchestration history; may miss council/subtree execution — triage before acting",
		"source":        dormantSource,
	})
	return string(b)
}

// dormantInsertItem writes one never-run agent as an INERT dormant_agent item
// anchored to system.internal. status='dormant' is inside idx_swi_dedup, so
// ON CONFLICT DO NOTHING dedups a re-sweep cleanly. pipeline='maintenance' and
// a non-triaged/approved status mean nothing claims or dispatches it — it is a
// human-triage signal, not queued work. Returns true only on a fresh insert.
func dormantInsertItem(ctx context.Context, db *sql.DB, a dormantAgent, ageFloor int) (bool, error) {
	res, err := db.ExecContext(ctx, `
		INSERT INTO site_work_items (
			site_id, source, pipeline, item_type, severity, summary,
			spec, priority, handler_agent, status, created_by, item_key, max_attempts
		) VALUES (
			$1, $2, 'maintenance', 'dormant_agent', 'low', $3,
			$4::jsonb, 30, '', 'dormant', $2, $5, 1
		)
		ON CONFLICT DO NOTHING`,
		triageSystemSiteID, dormantSource, dormantSummary(a), dormantSpecJSON(a, ageFloor), dormantItemKey(a.Type))
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// dormantCloseResolved completes any of our still-'dormant' items whose agent
// is no longer in the never-observed set (it ran, or was deactivated/deleted).
func dormantCloseResolved(ctx context.Context, db *sql.DB, stillDormant map[string]bool) (int, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id::text, COALESCE(spec->>'agent_type','') AS agent_type
		FROM site_work_items
		WHERE created_by = $1 AND status = 'dormant' AND item_type = 'dormant_agent'`, dormantSource)
	if err != nil {
		return 0, err
	}
	type ref struct{ id, agentType string }
	var stale []ref
	for rows.Next() {
		var r ref
		if err := rows.Scan(&r.id, &r.agentType); err != nil {
			rows.Close()
			return 0, err
		}
		if r.agentType != "" && !stillDormant[r.agentType] {
			stale = append(stale, r)
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, err
	}
	closed := 0
	for _, r := range stale {
		res, err := db.ExecContext(ctx, `
			UPDATE site_work_items
			SET status = 'complete', completed_at = now(), updated_at = now(),
			    error = COALESCE(error,'') || ' [dormant-agents: agent has since been observed running (or was deactivated); closed]'
			WHERE id = $1::uuid AND status = 'dormant' AND created_by = $2`, r.id, dormantSource)
		if err != nil {
			return closed, err
		}
		if n, _ := res.RowsAffected(); n > 0 {
			closed++
		}
	}
	return closed, nil
}

// dormantEra is a REPORT-ONLY, [HEURISTIC] label to make the inventory readable
// — it groups a never-run agent by a coarse guess at why. It is deliberately
// not used for any emit/close decision, because "paused by decision" vs
// "retired but still active" vs "current capability that never fired" is a
// judgement call the detector cannot make; the seed-date + name are only hints.
func dormantEra(a dormantAgent) string {
	name := a.Type
	switch {
	case strings.HasPrefix(name, "med-"), strings.HasPrefix(name, "ch-"),
		strings.HasPrefix(name, "vet-"), strings.HasPrefix(name, "thunder-"),
		strings.HasPrefix(name, "training-"), strings.HasPrefix(name, "model-"),
		strings.HasPrefix(name, "gpu-"), name == "business-intel", name == "area-sweep-orchestrator",
		name == "area-sweep-discoverer", name == "vet-pipeline-orchestrator":
		return "paused workstream (dormant by decision — vet / companies-house / training lanes)"
	case a.AgeDays >= 180:
		return "legacy / superseded (predates the current pipeline — likely is_active hygiene)"
	case a.AgeDays >= 60:
		return "older generation (verify: superseded, or a real capability that never routed)"
	default:
		return "current generation (the sharp set — a wired capability nothing routes to)"
	}
}

// renderDormantAgents is pure — tested. The owner-readable inventory artifact.
func renderDormantAgents(now time.Time, ageFloor int, stats dormantStats,
	past, under []dormantAgent, blind []string, oldestObserved time.Time,
	created, deduped, capped, maxEmit, closed int, dryRun bool) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Dormant-agents sweep (generated %s UTC; age floor %dd)\n\n", now.Format("2006-01-02 15:04"), ageFloor)
	if dryRun {
		b.WriteString("_DRY RUN — findings below are report-only this sweep: NO dormant_agent items were written, none were closed._\n\n")
	}

	fmt.Fprintf(&b, "**Capability inventory (fingerprint method).** Of **%d** active non-snapshot agents with a workflow, **%d** are measurable (have ≥1 step key unique to them) and **%d** are in the mirrored-agent blind spot (no unique step — unmeasurable this way, never flagged). Of the measurable, **%d** have never been observed running: **%d** past the %dd age floor (eligible to emit), **%d** too new to flag yet.\n\n",
		stats.ActiveWithWorkflow, stats.Measurable, stats.BlindSpot,
		len(past)+len(under), len(past), ageFloor, len(under))

	if !oldestObserved.IsZero() {
		fmt.Fprintf(&b, "> **What \"never observed\" means here:** no unique top-level workflow step seen in the retained orchestration history (since **%s**). An agent that ran only before that, or only via a council/subtree path whose steps never surface as top-level plan keys, reads as never-observed. Triage each finding before acting — this is an inventory for human review, not a fix.\n\n",
			oldestObserved.Format("2006-01-02"))
	}

	writeGroup := func(title string, agents []dormantAgent) {
		fmt.Fprintf(&b, "## %s — %d agent(s)\n\n", title, len(agents))
		if len(agents) == 0 {
			b.WriteString("None.\n\n")
			return
		}
		// Group by the heuristic era for readability.
		byEra := map[string][]dormantAgent{}
		var eras []string
		for _, a := range agents {
			e := dormantEra(a)
			if _, ok := byEra[e]; !ok {
				eras = append(eras, e)
			}
			byEra[e] = append(byEra[e], a)
		}
		sort.Strings(eras)
		for _, e := range eras {
			fmt.Fprintf(&b, "**%s** _[HEURISTIC label]_\n\n", e)
			for _, a := range byEra[e] {
				dup := ""
				if a.ActiveRows > 1 {
					dup = fmt.Sprintf(" ⚠ %d active rows (shadowed duplicates — is_active hygiene)", a.ActiveRows)
				}
				fmt.Fprintf(&b, "- `%s` — age %.0fd, %d unique step(s) (e.g. `%s`)%s\n", a.Type, a.AgeDays, a.UniqueSteps, a.SampleStep, dup)
			}
			b.WriteString("\n")
		}
	}

	writeGroup(fmt.Sprintf("Never observed, past the age floor (%dd)", ageFloor), past)
	writeGroup("Never observed, too new to flag (under the age floor)", under)

	// The mirrored-agent blind spot: listed, never flagged. A reader must see
	// that these are unmeasured (council-gate, whose 099 mirror copies
	// fix-proposer's steps), not silently assumed to be running.
	fmt.Fprintf(&b, "## Mirrored-agent blind spot — %d agent(s) unmeasurable by this method\n\n", len(blind))
	if len(blind) == 0 {
		b.WriteString("None.\n\n")
	} else {
		b.WriteString("_No step key is unique to these agents (their steps are shared/mirrored), so the fingerprint method cannot tell whether they run. They are NEVER flagged as dormant — a second signal would be needed to measure them (orchestration_name was evaluated and rejected: it does not name the agent)._\n\n")
		for _, t := range blind {
			fmt.Fprintf(&b, "- `%s`\n", t)
		}
		b.WriteString("\n")
	}

	if !dryRun {
		fmt.Fprintf(&b, "## Bookkeeping\n\nEmitted %d, deduped (already open) %d, capped %d (cap=%d); closed as resolved %d.\n",
			created, deduped, capped, maxEmit, closed)
		if capped > 0 {
			fmt.Fprintf(&b, "\n> %d finding(s) NOT emitted this sweep (cap=%d) — coverage was capped, not complete. Raise `max_emit` once the report has been reviewed.\n", capped, maxEmit)
		}
		b.WriteString("\n")
	}

	b.WriteString("---\n_Deterministic capability inventory: SQL-gathered, Go-rendered, no model. `owner_agent_type` is deliberately NOT used (95k+ rows carry 'generic'). Findings emit as INERT `dormant_agent` items (status='dormant', pipeline='maintenance', unclaimable) anchored to system.internal; a human decides wire / retire / paused. The mirrored-agent blind spot is never flagged. bugs_open/044._\n")
	return b.String()
}
