// FILE: platform/orchestration/actions/diagnose_dormant_agents_action.go
//
// DORMANT-AGENTS — the capability-inventory sweep (bugs_open/044). The
// producer-side mirror of the silent-check checker: silent-check finds work
// that no worker ever touches; this finds a WORKER that no work ever routes to.
// Both are invisible because nothing fails — an active agent that never runs
// errors nowhere, so a capability can quietly go unused for months until a
// thread declares it missing (bugs_closed/002 D: section-editor, shipped, 3
// production runs, declared nonexistent through a handoff and two sign-offs).
//
// SIGNAL (rewired 2026-07-24, bugs_open/060). The durable per-agent run record
// `agent_run_stats` — upserted at every orchestration START on the RESOLVED real
// agent type, and NOT pruned. An active agent with NO `agent_run_stats` row has
// never run since tracking began. This replaces the earlier step-fingerprint
// method over `orchestration_states` (which is pruned hourly at 24h, so it
// over-flagged constantly — fix-proposer read as dormant because its runs were
// pruned). Because the record keys on the actual dispatched type, it also has NO
// mirrored-agent blind spot (council-gate is recorded as itself if it runs) and
// no council/subtree false-negatives.
//
// DELIBERATELY DETERMINISTIC: no LLM. It measures; it never diagnoses and never
// fixes. Per sweep it:
//   - finds active non-snapshot agents with a workflow that have NO row in
//     agent_run_stats (never run since tracking began);
//   - applies an AGE FLOOR so a freshly-seeded agent that has not had a chance to
//     run yet is reported but not emitted;
//   - applies a WINDOW GUARD: "dormant for >= floor days" cannot be substantiated
//     until we have been TRACKING for at least the floor. The tracking window is
//     `now - min(first_ran_at)` and it GROWS (agent_run_stats is not pruned), so
//     emission un-gates itself once the record is old enough — the honest handling
//     of the forward-only cold start;
//   - emits ONE inert dormant_agent item per past-floor never-run agent
//     (status='dormant', pipeline='maintenance', unclaimable — nothing dispatches
//     it; a human triages wire/retire/paused), capped, deduped by ON CONFLICT;
//   - closes its own items ('complete') once the agent is observed to have run
//     (an agent_run_stats row appears);
//   - writes one doc_note per sweep (categories dormant-agents+fixloop) — the
//     inventory artifact that makes a capability discoverable before it is
//     declared missing.
//
// COLD START (stated honestly in the report): agent_run_stats is forward-only —
// it cannot know pre-deploy history. Right after deploy it is nearly empty, so
// "no row" is weak until it has accumulated >= the age floor. The window guard
// handles exactly this: it emits nothing until the tracking window >= the floor.
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

// dormantAgent is one active agent that has never run since tracking began.
type dormantAgent struct {
	Type         string    // agent_definitions.type
	ActiveRows   int       // active non-snapshot rows for this type (>1 is a hygiene flag)
	AgeDays      float64   // now() - min(created_at) across its active rows
	FirstCreated time.Time // min(created_at) — when the capability first existed
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

	// Gather: never-run agents + the summary stats + the tracking window.
	never, err := dormantGatherNeverRun(ctx, params.DB)
	if err != nil {
		return nil, fmt.Errorf("gather never-run: %w", err)
	}
	stats, err := dormantGatherStats(ctx, params.DB)
	if err != nil {
		return nil, fmt.Errorf("gather stats: %w", err)
	}
	trackingSince, err := dormantTrackingSince(ctx, params.DB)
	if err != nil {
		return nil, fmt.Errorf("gather tracking window: %w", err)
	}

	// Split by the age floor: only past-floor agents are eligible to EMIT; the
	// report shows both (nothing is hidden — an under-floor agent is a candidate
	// that simply has not aged in yet).
	past, under := dormantPartition(never, float64(ageFloor))

	now := time.Now().UTC()

	// WINDOW GUARD. "never run" is only as strong as how long we have been
	// tracking. agent_run_stats is forward-only, so right after deploy the window
	// is ~0 and every agent looks never-run. Do NOT emit until the tracking window
	// (now - min(first_ran_at)) is at least the age floor — long enough that "has
	// not run in >= floor days" is substantiable. The window GROWS (the table is
	// not pruned), so emission un-gates itself over time. The report always writes.
	windowDays := 0.0
	if !trackingSince.IsZero() {
		windowDays = now.Sub(trackingSince).Hours() / 24.0
	}
	windowSufficient := !trackingSince.IsZero() && windowDays >= float64(ageFloor)

	// Emit (live only, AND only when the window supports the claim): one item per
	// past-floor agent, newest-first so the freshest genuinely-unused capabilities
	// surface before the legacy backlog, capped, deduped by ON CONFLICT
	// (status='dormant' is inside idx_swi_dedup, so ON CONFLICT dedups cleanly).
	created, deduped, capped := 0, 0, 0
	var emittedKeys []string
	if !dryRun && windowSufficient {
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
	// never-run set (it ran, or was deactivated/deleted) gets honestly completed.
	// Being non-'dormant', a completed row never blocks a future re-emission.
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

	report := renderDormantAgents(now, ageFloor, stats, past, under, trackingSince,
		windowDays, windowSufficient, created, deduped, capped, maxEmit, closed, dryRun)

	noteID, nerr := insertDocNote(ctx, params.DB, "pipeline", "diagnose", "",
		report, `["dormant-agents","fixloop"]`, "diagnose_dormant_agents", nullSafeAgentType(params), "", "diagnose_dormant_agents")
	if nerr != nil {
		logger.Warn("dormant-agents: doc_note persist failed (report still returned)", zap.Error(nerr))
	}

	logger.Info("diagnose_dormant_agents: swept",
		zap.Int("active_with_workflow", stats.ActiveWithWorkflow),
		zap.Int("ran", stats.Ran),
		zap.Int("never_run", len(never)),
		zap.Int("past_floor", len(past)),
		zap.Int("under_floor", len(under)),
		zap.Int("emitted", created),
		zap.Int("deduped", deduped),
		zap.Int("capped", capped),
		zap.Int("closed_resolved", closed),
		zap.Float64("window_days", windowDays),
		zap.Bool("window_sufficient", windowSufficient),
		zap.Bool("dry_run", dryRun),
		zap.String("note_id", noteID),
		zap.String("orchestration_id", orchIDForLog(params)))

	return map[string]interface{}{
		"report":               report,
		"active_with_workflow": stats.ActiveWithWorkflow,
		"ran":                  stats.Ran,
		"never_run":            len(never),
		"past_floor":           len(past),
		"under_floor":          len(under),
		"window_days":          windowDays,
		"window_sufficient":    windowSufficient,
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
	Ran                int // …that have an agent_run_stats row (ran at least once since tracking began)
}

// dormantActiveAgentsCTE — active non-snapshot agents with a workflow. Shared by
// the gather and stats queries so they cannot drift apart. Collapses the
// >1-active-row (is_active hygiene) duplicates by type.
const dormantActiveAgentsCTE = `
	active_agents AS (
		SELECT a.type,
		       count(*) AS active_rows,
		       min(a.created_at) AS first_created
		FROM agent_definitions a
		WHERE a.is_active AND a.deleted_at IS NULL AND COALESCE(a.is_snapshot,false)=false
		  AND jsonb_typeof(a.default_config#>'{workflow,steps}') = 'object'
		GROUP BY a.type
	)`

// dormantGatherNeverRun returns active agents with NO agent_run_stats row.
func dormantGatherNeverRun(ctx context.Context, db *sql.DB) ([]dormantAgent, error) {
	query := `
	WITH` + dormantActiveAgentsCTE + `
	SELECT aa.type, aa.active_rows, aa.first_created,
	       round(extract(epoch FROM (now() - aa.first_created)) / 86400.0, 1) AS age_days
	FROM active_agents aa
	WHERE NOT EXISTS (SELECT 1 FROM agent_run_stats s WHERE s.agent_type = aa.type)
	ORDER BY age_days DESC, aa.type`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []dormantAgent
	for rows.Next() {
		var a dormantAgent
		if err := rows.Scan(&a.Type, &a.ActiveRows, &a.FirstCreated, &a.AgeDays); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// dormantGatherStats returns the headline counts.
func dormantGatherStats(ctx context.Context, db *sql.DB) (dormantStats, error) {
	var s dormantStats
	query := `
	WITH` + dormantActiveAgentsCTE + `
	SELECT
		(SELECT count(*) FROM active_agents) AS active_with_workflow,
		(SELECT count(*) FROM active_agents aa
		  WHERE EXISTS (SELECT 1 FROM agent_run_stats s WHERE s.agent_type = aa.type)) AS ran`
	err := db.QueryRowContext(ctx, query).Scan(&s.ActiveWithWorkflow, &s.Ran)
	return s, err
}

// dormantTrackingSince returns the oldest first_ran_at — how long the durable
// record has been accumulating (the honest lower bound of "never run since ...").
// Zero when agent_run_stats is empty (nothing recorded yet).
func dormantTrackingSince(ctx context.Context, db *sql.DB) (time.Time, error) {
	var t sql.NullTime
	err := db.QueryRowContext(ctx, `SELECT min(first_ran_at) FROM agent_run_stats`).Scan(&t)
	if err != nil {
		return time.Time{}, err
	}
	if !t.Valid {
		return time.Time{}, nil
	}
	return t.Time.UTC(), nil
}

// dormantPartition splits the never-run set into agents past the age floor
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
	return fmt.Sprintf("DORMANT: agent %q is active but has never run since tracking began (age %.0fd, no agent_run_stats row). Human triage: wire it, retire it, or record it as paused.",
		a.Type, a.AgeDays)
}

func dormantSpecJSON(a dormantAgent, ageFloor int) string {
	b, _ := json.Marshal(map[string]interface{}{
		"agent_type":    a.Type,
		"active_rows":   a.ActiveRows,
		"age_days":      a.AgeDays,
		"first_created": a.FirstCreated.UTC().Format(time.RFC3339),
		"age_floor":     ageFloor,
		"method":        "durable run record: an active non-snapshot agent with NO row in agent_run_stats (never run since tracking began); owner_agent_type / orchestration_states are NOT used (the latter is pruned at 24h)",
		"caveat":        "never-run = never recorded since agent_run_stats began accumulating (forward-only); triage before acting",
		"source":        dormantSource,
	})
	return string(b)
}

// dormantInsertItem writes one never-run agent as an INERT dormant_agent item
// anchored to system.internal. status='dormant' is inside idx_swi_dedup, so
// ON CONFLICT DO NOTHING dedups a re-sweep cleanly. pipeline='maintenance' and a
// non-triaged/approved status mean nothing claims or dispatches it — it is a
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

// dormantCloseResolved completes any of our still-'dormant' items whose agent is
// no longer in the never-run set (it ran, or was deactivated/deleted).
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
			    error = COALESCE(error,'') || ' [dormant-agents: agent has since run (or was deactivated); closed]'
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
// — it groups a never-run agent by a coarse guess at why. It is deliberately not
// used for any emit/close decision, because "paused by decision" vs "retired but
// still active" vs "current capability that never fired" is a judgement call the
// detector cannot make; the seed-date + name are only hints.
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
	past, under []dormantAgent, trackingSince time.Time,
	windowDays float64, windowSufficient bool,
	created, deduped, capped, maxEmit, closed int, dryRun bool) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Dormant-agents sweep (generated %s UTC; age floor %dd)\n\n", now.Format("2006-01-02 15:04"), ageFloor)
	if dryRun {
		b.WriteString("_DRY RUN — findings below are report-only this sweep: NO dormant_agent items were written, none were closed._\n\n")
	} else if !windowSufficient {
		fmt.Fprintf(&b, "> ⚠ **WINDOW TOO SHORT — NO items emitted this sweep (report only).** The durable run record has only been accumulating for **%.1f days**, shorter than the %dd age floor, so \"dormant for ≥ %dd\" cannot be substantiated yet and emitting would flood false positives. `agent_run_stats` is forward-only (it cannot know pre-deploy history); the window GROWS, so emission un-gates itself once tracking ≥ the floor. See `bugs_open/060`.\n\n",
			windowDays, ageFloor, ageFloor)
	}

	neverRun := len(past) + len(under)
	fmt.Fprintf(&b, "**Capability inventory (durable run record).** Of **%d** active non-snapshot agents with a workflow, **%d** have run at least once since tracking began and **%d** have never run: **%d** past the %dd age floor, **%d** too new to flag yet.\n\n",
		stats.ActiveWithWorkflow, stats.Ran, neverRun, len(past), ageFloor, len(under))

	if !trackingSince.IsZero() {
		fmt.Fprintf(&b, "> **What \"never run\" means here — READ THIS before acting.** It means \"no row in `agent_run_stats`\", which has been accumulating since **%s** (**%.1f days**). The record is written at every orchestration start on the resolved real agent type, and it is NOT pruned — so unlike the old step-fingerprint method it has no mirrored-agent blind spot and no 24h-window false positives. But it is FORWARD-ONLY: it cannot see runs from before tracking began, so an agent that last ran pre-deploy reads as never-run until the window matures. Triage before acting.\n\n",
			trackingSince.Format("2006-01-02"), windowDays)
	} else {
		b.WriteString("> **No runs recorded yet** — `agent_run_stats` is empty (tracking has just begun, or no orchestration has started since deploy). Every active agent therefore reads as never-run; this is the cold start, not a fleet-wide outage. Emission is gated until the window matures.\n\n")
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
				fmt.Fprintf(&b, "- `%s` — age %.0fd%s\n", a.Type, a.AgeDays, dup)
			}
			b.WriteString("\n")
		}
	}

	writeGroup(fmt.Sprintf("Never run, past the age floor (%dd)", ageFloor), past)
	writeGroup("Never run, too new to flag (under the age floor)", under)

	if !dryRun {
		fmt.Fprintf(&b, "## Bookkeeping\n\nEmitted %d, deduped (already open) %d, capped %d (cap=%d); closed as resolved %d.\n",
			created, deduped, capped, maxEmit, closed)
		if capped > 0 {
			fmt.Fprintf(&b, "\n> %d finding(s) NOT emitted this sweep (cap=%d) — coverage was capped, not complete. Raise `max_emit` once the report has been reviewed.\n", capped, maxEmit)
		}
		b.WriteString("\n")
	}

	b.WriteString("---\n_Deterministic capability inventory: SQL-gathered, Go-rendered, no model. Signal = the durable `agent_run_stats` record (a row per agent type, upserted at orchestration start on the resolved real type, not pruned). Findings emit as INERT `dormant_agent` items (status='dormant', pipeline='maintenance', unclaimable) anchored to system.internal; a human decides wire / retire / paused. bugs_open/044 + 060._\n")
	return b.String()
}
