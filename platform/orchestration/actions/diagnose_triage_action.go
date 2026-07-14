// FILE: platform/orchestration/actions/diagnose_triage_action.go
//
// TRIAGE — the router from the operational immune system into the fix loop
// (DESIGN_triage_and_escalation.md, owner decisions 2026-07-14). Phase 1: the
// two failure flavours that already exist in the data.
//
//	loud failure   (status='failed')            → escalate to the fix loop
//	no handler yet (item_type='capability_gap') → surface to the roadmap
//
// DELIBERATELY DETERMINISTIC: no LLM. Triage ROUTES; it never diagnoses and
// never fixes. It scans site_work_items, DEDUPES by pattern (fifty pages
// failing the same way = ONE escalation, never fifty — the guardrail that
// stops the loop being buried), caps the number raised per sweep, and writes:
//   - one 'needs_diagnosis' work item per failure PATTERN (the loop's existing
//     intake: system.internal anchor, pipeline='diagnose', status=
//     'awaiting_diagnosis' — INERT until a human/dispatch claims it, so writing
//     is safe), deduped by a deterministic item_key + ON CONFLICT DO NOTHING;
//   - one doc_note per run: what it escalated + the open capability gaps, so
//     the owner has one readable artifact (categories triage+fixloop).
//
// Manual-trigger for now (owner). It creates queue entries, not diagnoses —
// dispatch stays separate and human-gated.
package actions

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// system.internal pseudo-site — every needs_diagnosis anchors here; the real
// site under diagnosis travels in the spec (090 contract).
const triageSystemSiteID = "eac60db8-b032-432b-b36d-76f37632045d"

// defaultTransientSignatures — the LOOP-WORTHINESS FILTER's denylist (Phase
// 1.1, from the first live dry-run: half the "failures" were the claim/dispatch
// layer timing out, which has NO code fix). A failure whose error contains one
// of these is OPERATIONAL — route to re-queue, never to the loop. Deliberately
// precise to the dispatch/pod/infra layer, so a genuine handler-logic error
// that merely mentions time is NOT swallowed. Tunable via config.
var defaultTransientSignatures = []string{
	"claim timed out",
	"pod likely died",
	"handler pod",
	"consumer rebalance",
	"rebalance in progress",
}

// triageRoute classifies one failure pattern by its error signature — the
// loop-worthiness filter. Pure and tested.
//
//	""              → "hold"    (no signal: nothing to diagnose — a human looks)
//	transient match → "requeue" (operational/infra: not a code bug)
//	else            → "loop"    (a real handler error worth diagnosing)
func triageRoute(errSig string, transient []string) string {
	s := strings.ToLower(strings.TrimSpace(errSig))
	if s == "" {
		return "hold"
	}
	for _, t := range transient {
		if t = strings.ToLower(strings.TrimSpace(t)); t != "" && strings.Contains(s, t) {
			return "requeue"
		}
	}
	return "loop"
}

var DiagnoseTriageInputSpec = datahelpers.ActionInputSpec{
	Required: []string{},
	Optional: []string{
		"window_hours", "max_escalations", "diagnose_handler",
		"repo_owner", "repo_name", "ref", "dry_run", "transient_signatures",
	},
	Defaults: map[string]interface{}{
		"window_hours":     336, // 14 days — failed items linger; a wide look is fine, dedup bounds output
		"max_escalations":  3,   // hard cap per sweep while confidence builds (owner)
		"diagnose_handler": "diagnose-orchestrator",
		"repo_owner":       "gqls",
		"repo_name":        "agentchassis",
		"ref":              "main",
		"dry_run":          false, // when true: preview only — the report writes, but NO needs_diagnosis items
	},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("diagnose_triage", DiagnoseTriageInputSpec)
}

type failurePattern struct {
	ItemType  string
	Handler   string
	ErrSig    string
	Count     int
	Sites     int
	SampleKey string
}

type capabilityGap struct {
	Builder string
	Count   int
	Sites   int
}

// DiagnoseTriageAction runs one triage sweep.
func DiagnoseTriageAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	logger := params.Logger.With(zap.String("action", "diagnose_triage"))
	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("diagnose_triage: no DB handle")
	}

	hours := datahelpers.GetIntField(config, "window_hours", 336)
	if hours <= 0 {
		hours = 336
	}
	window := fmt.Sprintf("%d hours", hours)
	maxEsc := datahelpers.GetIntField(config, "max_escalations", 3)
	handler := datahelpers.GetStringField(config, "diagnose_handler", "diagnose-orchestrator")
	owner := datahelpers.GetStringField(config, "repo_owner", "gqls")
	repo := datahelpers.GetStringField(config, "repo_name", "agentchassis")
	ref := datahelpers.GetStringField(config, "ref", "main")
	dryRun := false
	if d, ok := config["dry_run"].(bool); ok {
		dryRun = d
	}

	transient := configStringSlice(config, "transient_signatures", defaultTransientSignatures)

	patterns, err := triageGatherFailures(ctx, params.DB, window)
	if err != nil {
		return nil, fmt.Errorf("gather loud failures: %w", err)
	}
	gaps, err := triageGatherCapabilityGaps(ctx, params.DB, window)
	if err != nil {
		return nil, fmt.Errorf("gather capability gaps: %w", err)
	}

	// LOOP-WORTHINESS FILTER: classify each pattern, and escalate ONLY the ones
	// that look like real code bugs. Transient/infra failures and no-signal
	// ones are surfaced in the report but never sent to the loop.
	var loopPatterns, requeuePatterns, holdPatterns []failurePattern
	for _, p := range patterns {
		switch triageRoute(p.ErrSig, transient) {
		case "loop":
			loopPatterns = append(loopPatterns, p)
		case "requeue":
			requeuePatterns = append(requeuePatterns, p)
		default:
			holdPatterns = append(holdPatterns, p)
		}
	}

	// Escalate up to the cap. ON CONFLICT DO NOTHING (dedup on the system-site
	// + item_key partial index) makes a re-sweep idempotent: a pattern whose
	// needs_diagnosis is still open is not re-created. dry_run previews without
	// writing any work item. Report created vs deduped vs capped — never a
	// silent drop.
	created, deduped := 0, 0
	var escalated []string
	if !dryRun {
		for _, p := range loopPatterns {
			if created >= maxEsc {
				break
			}
			itemKey := triageItemKey(p.ItemType, p.Handler, p.ErrSig)
			symptom := triageSymptom(p)
			specJSON := triageSpecJSON(symptom, owner, repo, ref, handler)
			ins, err := triageInsertNeedsDiagnosis(ctx, params.DB, symptom, specJSON, handler, itemKey)
			if err != nil {
				logger.Warn("triage: escalation insert failed", zap.String("item_key", itemKey), zap.Error(err))
				continue
			}
			if ins {
				created++
				escalated = append(escalated, itemKey)
			} else {
				deduped++
			}
		}
	}

	capped := len(loopPatterns) - created - deduped
	if capped < 0 || dryRun {
		// dry-run writes nothing at all — reporting the whole group as
		// "capped" would mislabel a preview as a coverage gap.
		capped = 0
	}
	report := renderTriage(hours, time.Now().UTC(), loopPatterns, requeuePatterns, holdPatterns, gaps, created, deduped, capped, maxEsc, dryRun)

	// The report note ALWAYS writes — it is the visibility artifact, harmless,
	// and the whole point in dry-run. Only the escalation writes are gated.
	noteID, nerr := insertDocNote(ctx, params.DB, "pipeline", "diagnose", "",
		report, `["triage","fixloop"]`, "diagnose_triage", nullSafeAgentType(params), "", "diagnose_triage")
	if nerr != nil {
		logger.Warn("triage: doc_note persist failed (report still returned)", zap.Error(nerr))
	}

	logger.Info("diagnose_triage: swept",
		zap.Int("failure_patterns", len(patterns)),
		zap.Int("loop_patterns", len(loopPatterns)),
		zap.Int("requeue_patterns", len(requeuePatterns)),
		zap.Int("hold_patterns", len(holdPatterns)),
		zap.Int("escalated", created),
		zap.Int("deduped", deduped),
		zap.Int("capped", capped),
		zap.Int("capability_gaps", len(gaps)),
		zap.String("note_id", noteID),
		zap.String("orchestration_id", orchIDForLog(params)))

	return map[string]interface{}{
		"report":           report,
		"escalated":        created,
		"deduped":          deduped,
		"capped":           capped,
		"requeue_patterns": len(requeuePatterns),
		"hold_patterns":    len(holdPatterns),
		"capability_gaps":  len(gaps),
		"escalated_keys":   escalated,
		"note_id":          noteID,
	}, nil
}

// triageGatherFailures groups loud failures by (item_type, handler, error
// signature) — one row per PATTERN. The fix loop and triage's own queue are
// excluded so triage never escalates itself.
func triageGatherFailures(ctx context.Context, db *sql.DB, window string) ([]failurePattern, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT item_type, COALESCE(handler_agent,''),
		       left(COALESCE(error,''), 140) AS errsig,
		       count(*) AS n, count(DISTINCT site_id) AS sites
		FROM site_work_items
		WHERE status = 'failed'
		  AND pipeline <> 'diagnose'
		  AND item_type NOT IN ('needs_diagnosis','capability_gap')
		  AND COALESCE(updated_at, created_at) > now() - $1::interval
		GROUP BY item_type, COALESCE(handler_agent,''), left(COALESCE(error,''),140)
		ORDER BY n DESC`, window)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []failurePattern
	for rows.Next() {
		var p failurePattern
		if err := rows.Scan(&p.ItemType, &p.Handler, &p.ErrSig, &p.Count, &p.Sites); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// triageGatherCapabilityGaps groups the no-handler-yet signal by the builder
// that's missing — a roadmap view, not an escalation.
func triageGatherCapabilityGaps(ctx context.Context, db *sql.DB, window string) ([]capabilityGap, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT COALESCE(spec->>'builder_needed', COALESCE(handler_agent,'?')) AS builder,
		       count(*) AS n, count(DISTINCT site_id) AS sites
		FROM site_work_items
		WHERE (item_type = 'capability_gap' OR status = 'deferred')
		  AND COALESCE(updated_at, created_at) > now() - $1::interval
		GROUP BY 1
		ORDER BY n DESC`, window)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []capabilityGap
	for rows.Next() {
		var g capabilityGap
		if err := rows.Scan(&g.Builder, &g.Count, &g.Sites); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

// triageInsertNeedsDiagnosis writes one intake item (090 contract). Returns
// true if a row was created, false if ON CONFLICT deduped it (pattern already
// queued). Anchors to system.internal; parks at awaiting_diagnosis (inert).
func triageInsertNeedsDiagnosis(ctx context.Context, db *sql.DB, symptom, specJSON, handler, itemKey string) (bool, error) {
	res, err := db.ExecContext(ctx, `
		INSERT INTO site_work_items (
			site_id, source, pipeline, item_type, severity, summary,
			spec, priority, handler_agent, status, created_by, item_key, max_attempts
		) VALUES (
			$1, 'diagnosis-triage', 'diagnose', 'needs_diagnosis',
			'medium', $2, $3::jsonb, 50, $4, 'awaiting_diagnosis', 'diagnosis-triage', $5, 1
		)
		ON CONFLICT DO NOTHING`,
		triageSystemSiteID, symptom, specJSON, handler, itemKey)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// triageItemKey is a STABLE dedup key per failure pattern — the same pattern
// always maps to the same key, so ON CONFLICT collapses re-sweeps.
func triageItemKey(itemType, handler, errSig string) string {
	h := sha1.Sum([]byte(itemType + "|" + handler + "|" + strings.TrimSpace(errSig)))
	return fmt.Sprintf("triage-diag:%s:%x", itemType, h[:6])
}

// triageSymptom renders the human-and-loop-readable symptom string.
// silent_failure items come from the silent-check verification checker, not a
// failing handler — attributing them to a handler would misdirect the
// diagnosis, so they get their own rendering (page-level detail lives in the
// items' spec, item_key prefix 'silent:').
func triageSymptom(p failurePattern) string {
	errSig := strings.TrimSpace(p.ErrSig)
	if errSig == "" {
		errSig = "(no error text recorded)"
	}
	if p.ItemType == "silent_failure" {
		return fmt.Sprintf("silent failure on %d site(s) (structural invariant, no covering work item; detail in site_work_items item_key 'silent:%%'): %s",
			p.Sites, errSig)
	}
	return fmt.Sprintf("handler %q fails item_type %q on %d work item(s) across %d site(s); error: %s",
		p.Handler, p.ItemType, p.Count, p.Sites, errSig)
}

// triageSpecJSON builds the needs_diagnosis spec (the diagnose input envelope).
func triageSpecJSON(symptom, owner, repo, ref, handler string) string {
	b, _ := json.Marshal(map[string]interface{}{
		"symptom":        symptom,
		"owner":          owner,
		"repo":           repo,
		"ref":            ref,
		"correlation_id": uuid.New().String(),
		"source":         "diagnosis-triage",
	})
	return string(b)
}

// renderTriage is pure — tested. Names what was escalated, what was deduped/
// capped (never silently dropped), and the open capability gaps.
func renderTriage(hours int, now time.Time, loop, requeue, hold []failurePattern, gaps []capabilityGap, created, deduped, capped, maxEsc int, dryRun bool) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Triage sweep — last %dh (generated %s UTC)\n\n", hours, now.Format("2006-01-02 15:04"))
	if dryRun {
		b.WriteString("_DRY RUN — the code-bug patterns below WOULD be escalated, but NO needs_diagnosis items were written._\n\n")
	}

	patLine := func(p failurePattern) {
		errSig := strings.TrimSpace(p.ErrSig)
		if errSig == "" {
			errSig = "(no error text)"
		}
		if p.Handler == "" {
			fmt.Fprintf(&b, "- `%s` — %d item(s), %d site(s): %s\n", p.ItemType, p.Count, p.Sites, errSig)
			return
		}
		fmt.Fprintf(&b, "- `%s` via `%s` — %d item(s), %d site(s): %s\n", p.ItemType, p.Handler, p.Count, p.Sites, errSig)
	}

	// → FIX LOOP: real code bugs (the only group that escalates).
	fmt.Fprintf(&b, "## Code bugs → fix loop (%d pattern(s); escalated %d, deduped %d, capped %d; cap=%d)\n\n",
		len(loop), created, deduped, capped, maxEsc)
	if len(loop) == 0 {
		b.WriteString("No code-bug failure patterns in this window.\n\n")
	} else {
		for _, p := range loop {
			patLine(p)
		}
		if capped > 0 {
			fmt.Fprintf(&b, "\n> %d pattern(s) NOT escalated this sweep (cap=%d) — coverage was capped, not complete.\n", capped, maxEsc)
		}
		b.WriteString("\n")
	}

	// → RE-QUEUE: transient/infra — surfaced, NOT escalated (no code fix).
	fmt.Fprintf(&b, "## Transient / infra → re-queue, NOT the loop (%d pattern(s))\n\n", len(requeue))
	if len(requeue) == 0 {
		b.WriteString("None.\n\n")
	} else {
		b.WriteString("Operational failures (claim timeouts, dead pods) — no code fix; surfaced for the operational layer, not escalated:\n\n")
		for _, p := range requeue {
			patLine(p)
		}
		b.WriteString("\n")
	}

	// → HOLD: no error signal — a human looks.
	if len(hold) > 0 {
		fmt.Fprintf(&b, "## No error signal → hold for a human (%d pattern(s))\n\n", len(hold))
		b.WriteString("Failed, but with no recorded error to diagnose — nothing for the loop to work from:\n\n")
		for _, p := range hold {
			patLine(p)
		}
		b.WriteString("\n")
	}

	// → ROADMAP: capability gaps (no handler yet).
	fmt.Fprintf(&b, "## Capability gaps → roadmap, NOT the loop (%d)\n\n", len(gaps))
	if len(gaps) == 0 {
		b.WriteString("No capability_gap / deferred items in this window.\n\n")
	} else {
		b.WriteString("A missing handler is a capability decision, not a bug — these await a human roadmap call:\n\n")
		for _, g := range gaps {
			fmt.Fprintf(&b, "- **%s** needed — %d page(s) across %d site(s) waiting.\n", g.Builder, g.Count, g.Sites)
		}
		b.WriteString("\n")
	}

	b.WriteString("---\n_Deterministic router: SQL-gathered, Go-rendered, no model. Only code-bug patterns escalate (loop-worthiness filter); they park at `awaiting_diagnosis` (inert until dispatched). Patterns deduped by (item_type, handler, error signature)._\n")
	return b.String()
}
