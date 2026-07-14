// FILE: platform/orchestration/actions/fixloop_digest_action.go
//
// THE AWARENESS SURFACE (owner standing rule, 2026-07-12: "more awareness
// BEFORE wider autonomy"). The fix loop's activity is durable but pull-only —
// an owner who must go digging is not aware. This action composes a periodic
// DIGEST of what the loop did and decided, and persists it to doc_notes (the
// platform's human-notes surface).
//
// DELIBERATELY DETERMINISTIC: no LLM anywhere in this path. An awareness
// surface that could hallucinate what the system did would defeat its own
// purpose. Facts are gathered by SQL, rendered by plain Go, and every section
// names its window so absence of a line means absence of activity.
//
// Sections:
//  1. Loop runs        — orchestrations by fix-loop agent types: status + terminal step
//  2. Decisions        — per correlation: artifact counts, latest council decision & why
//  3. Write outcomes   — build-gate verdicts and any PRs opened (url/number)
//  4. Escalation channel (Phase 4, 2026-07-14) — the immune system on the same
//     page: sweep counts, the open diagnosis queue (new vs parked), open/closed
//     silent-failure findings, and standing capability gaps.
//  5. Config changes   — agent_definitions_backup snapshots in-window (type + reason):
//     every seeded/updated agent leaves one, so this is the
//     "what changed about the machine itself" ledger.
package actions

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var FixloopDigestInputSpec = datahelpers.ActionInputSpec{
	Required: []string{},
	Optional: []string{"window_hours", "subject_type", "subject_key", "persist"},
	Defaults: map[string]interface{}{
		"window_hours": 24,
		"subject_type": "pipeline",
		"subject_key":  "diagnose",
		"persist":      true,
	},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("fixloop_digest", FixloopDigestInputSpec)
}

// fix-loop agent types the digest reports on. A new loop agent must be added
// here to appear in the digest — named so the omission is discoverable.
var fixloopAgentTypes = []string{
	"diagnose-orchestrator", "diagnose-agent",
	"fix-proposer", "fix-implementer", "fix-implementer-orchestrator",
	"fixloop-digest",
	"diagnosis-triage", "diagnosis-silent-check",
}

type digestRun struct {
	AgentType string
	OrchID    string
	Status    string
	Terminal  string
	CreatedAt string
	GateNote  string // "gate PASS" / "gate FAIL" when present
	PRNote    string // "PR #n <url>" when present
}

type digestCorrelation struct {
	CorrelationID string
	Kinds         string // e.g. "fix_plan:2, council_report:2"
	Decision      string
	DecidedBy     string
}

type digestSnapshot struct {
	AgentType string
	Reason    string
	TakenAt   string
}

type digestEscalation struct {
	ItemKey   string
	Status    string
	Summary   string
	CreatedAt string
	New       bool // created inside this digest's window
}

type digestSilentFinding struct {
	ItemKey string
	Summary string
}

// digestImmune is the escalation channel's state for the digest — the immune
// system and the loop on one page (Phase 4 of the triage design).
type digestImmune struct {
	TriageSweeps int
	SilentSweeps int
	Escalations  []digestEscalation    // the OPEN diagnosis queue (new + parked)
	SilentOpen   []digestSilentFinding // silent-check findings still standing
	SilentClosed []digestSilentFinding // findings honestly closed in-window
	Gaps         []capabilityGap       // standing capability gaps (roadmap)
}

// FixloopDigestAction gathers and persists the digest.
func FixloopDigestAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	logger := params.Logger.With(zap.String("action", "fixloop_digest"))
	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("fixloop_digest: no DB handle")
	}

	hours := datahelpers.GetIntField(config, "window_hours", 24)
	if hours <= 0 {
		hours = 24
	}
	window := fmt.Sprintf("%d hours", hours)

	runs, err := digestGatherRuns(ctx, params.DB, window)
	if err != nil {
		return nil, fmt.Errorf("gather runs: %w", err)
	}
	corrs, err := digestGatherCorrelations(ctx, params.DB, window)
	if err != nil {
		return nil, fmt.Errorf("gather decisions: %w", err)
	}
	snaps, err := digestGatherSnapshots(ctx, params.DB, window)
	if err != nil {
		return nil, fmt.Errorf("gather config changes: %w", err)
	}
	immune, err := digestGatherImmune(ctx, params.DB, window)
	if err != nil {
		return nil, fmt.Errorf("gather escalation channel: %w", err)
	}

	body := renderDigest(hours, time.Now().UTC(), runs, corrs, immune, snaps)

	noteID := ""
	if persist, ok := config["persist"].(bool); !ok || persist {
		noteID, err = insertDocNote(ctx, params.DB,
			datahelpers.GetStringField(config, "subject_type", "pipeline"),
			datahelpers.GetStringField(config, "subject_key", "diagnose"),
			"", body, `["digest","fixloop"]`,
			"fixloop_digest", nullSafeAgentType(params), "", "fixloop_digest")
		if err != nil {
			return nil, fmt.Errorf("persist digest note: %w", err)
		}
	}

	logger.Info("fixloop_digest: composed",
		zap.Int("window_hours", hours),
		zap.Int("runs", len(runs)),
		zap.Int("correlations", len(corrs)),
		zap.Int("config_changes", len(snaps)),
		zap.String("note_id", noteID),
		zap.String("orchestration_id", orchIDForLog(params)))

	return map[string]interface{}{
		"digest":         body,
		"note_id":        noteID,
		"runs":           len(runs),
		"correlations":   len(corrs),
		"config_changes": len(snaps),
	}, nil
}

func nullSafeAgentType(params ActionParams) string {
	if params.AgentType == "" {
		return "fixloop-digest"
	}
	return params.AgentType
}

// digestGatherRuns lists in-window orchestrations for the loop's agent types,
// with gate/PR outcomes where the run recorded them.
func digestGatherRuns(ctx context.Context, db *sql.DB, window string) ([]digestRun, error) {
	// Agent type lives in TWO places depending on how the run started:
	// agent_group.type on directly-fired runs; __execution_context__.sender
	// .agent_type on SPAWNED children (found by the digest's own first run,
	// which missed the dedicated implementer pods — and with them the gate
	// verdicts and the PR url).
	rows, err := db.QueryContext(ctx, `
		SELECT COALESCE(collected_data->'agent_group'->>'type',
		                collected_data->'__execution_context__'->'sender'->>'agent_type', ''),
		       orchestration_id::text, status, current_step,
		       to_char(created_at, 'YYYY-MM-DD HH24:MI'),
		       COALESCE(collected_data->'gate'->>'passed',''),
		       COALESCE(collected_data->'pr_result'->'response'->'data'->>'pr_url',
		                collected_data->'pr_result'->'data'->>'pr_url', '')
		FROM orchestration_states
		WHERE created_at > now() - $1::interval
		  AND COALESCE(collected_data->'agent_group'->>'type',
		               collected_data->'__execution_context__'->'sender'->>'agent_type') = ANY($2)
		ORDER BY created_at`, window, pqStringArray(fixloopAgentTypes))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []digestRun
	for rows.Next() {
		var r digestRun
		var gatePassed, prURL string
		if err := rows.Scan(&r.AgentType, &r.OrchID, &r.Status, &r.Terminal, &r.CreatedAt, &gatePassed, &prURL); err != nil {
			return nil, err
		}
		switch gatePassed {
		case "true":
			r.GateNote = "gate PASS"
		case "false":
			r.GateNote = "gate FAIL"
		}
		if prURL != "" {
			r.PRNote = "PR " + prURL
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// digestGatherCorrelations summarises in-window diagnosis_artifacts per
// correlation: what kinds were written, and the latest council decision.
func digestGatherCorrelations(ctx context.Context, db *sql.DB, window string) ([]digestCorrelation, error) {
	rows, err := db.QueryContext(ctx, `
		WITH recent AS (
		  SELECT correlation_id, kind, body, metadata, created_at
		  FROM diagnosis_artifacts WHERE created_at > now() - $1::interval
		)
		SELECT r.correlation_id,
		       (SELECT string_agg(k.kind || ':' || k.n, ', ' ORDER BY k.kind)
		        FROM (SELECT kind, count(*) AS n FROM recent r2
		              WHERE r2.correlation_id = r.correlation_id GROUP BY kind) k),
		       COALESCE((SELECT cr.metadata->>'decision' FROM recent cr
		                 WHERE cr.correlation_id = r.correlation_id AND cr.kind='council_report'
		                 ORDER BY cr.created_at DESC LIMIT 1), ''),
		       COALESCE((SELECT cr.body::jsonb->>'decided_by' FROM recent cr
		                 WHERE cr.correlation_id = r.correlation_id AND cr.kind='council_report'
		                 ORDER BY cr.created_at DESC LIMIT 1), '')
		FROM recent r
		GROUP BY r.correlation_id
		ORDER BY r.correlation_id`, window)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []digestCorrelation
	for rows.Next() {
		var c digestCorrelation
		if err := rows.Scan(&c.CorrelationID, &c.Kinds, &c.Decision, &c.DecidedBy); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// digestGatherSnapshots lists in-window agent-definition snapshots — the
// "what changed about the machine itself" ledger.
func digestGatherSnapshots(ctx context.Context, db *sql.DB, window string) ([]digestSnapshot, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT type, COALESCE(snapshot_reason,''), to_char(snapshot_taken_at, 'YYYY-MM-DD HH24:MI')
		FROM agent_definitions_backup
		WHERE snapshot_taken_at > now() - $1::interval
		ORDER BY snapshot_taken_at`, window)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []digestSnapshot
	for rows.Next() {
		var s digestSnapshot
		if err := rows.Scan(&s.AgentType, &s.Reason, &s.TakenAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// digestGatherImmune assembles the escalation channel's state: sweep counts
// in-window, the whole OPEN diagnosis queue (parked items must stay visible
// every digest until resolved — they are decisions waiting on the owner, not
// activity), silent-check findings open + closed-in-window, and standing
// capability gaps.
func digestGatherImmune(ctx context.Context, db *sql.DB, window string) (digestImmune, error) {
	var im digestImmune

	if err := db.QueryRowContext(ctx, `
		SELECT count(*) FILTER (WHERE categories ? 'triage'),
		       count(*) FILTER (WHERE categories ? 'silent-check')
		FROM doc_notes WHERE created_at > now() - $1::interval`, window).
		Scan(&im.TriageSweeps, &im.SilentSweeps); err != nil {
		return im, err
	}

	rows, err := db.QueryContext(ctx, `
		SELECT item_key, status, left(COALESCE(summary,''),110),
		       to_char(created_at, 'YYYY-MM-DD HH24:MI'),
		       created_at > now() - $1::interval
		FROM site_work_items
		WHERE item_type = 'needs_diagnosis'
		  AND status IN ('awaiting_diagnosis','diagnosing')
		ORDER BY created_at DESC`, window)
	if err != nil {
		return im, err
	}
	for rows.Next() {
		var e digestEscalation
		if err := rows.Scan(&e.ItemKey, &e.Status, &e.Summary, &e.CreatedAt, &e.New); err != nil {
			rows.Close()
			return im, err
		}
		im.Escalations = append(im.Escalations, e)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return im, err
	}

	rows, err = db.QueryContext(ctx, `
		SELECT item_key, left(COALESCE(summary,''),110), status
		FROM site_work_items
		WHERE created_by = 'silent-check'
		  AND (status = 'failed'
		       OR (status = 'complete' AND updated_at > now() - $1::interval))
		ORDER BY created_at DESC`, window)
	if err != nil {
		return im, err
	}
	for rows.Next() {
		var f digestSilentFinding
		var status string
		if err := rows.Scan(&f.ItemKey, &f.Summary, &status); err != nil {
			rows.Close()
			return im, err
		}
		if status == "failed" {
			im.SilentOpen = append(im.SilentOpen, f)
		} else {
			im.SilentClosed = append(im.SilentClosed, f)
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return im, err
	}

	rows, err = db.QueryContext(ctx, `
		SELECT COALESCE(spec->>'builder_needed', COALESCE(handler_agent,'?')) AS builder,
		       count(*) AS n, count(DISTINCT site_id) AS sites
		FROM site_work_items
		WHERE (item_type = 'capability_gap' OR status = 'deferred')
		  AND status NOT IN ('complete','cancelled','wont_fix','rejected')
		GROUP BY 1 ORDER BY n DESC`)
	if err != nil {
		return im, err
	}
	for rows.Next() {
		var g capabilityGap
		if err := rows.Scan(&g.Builder, &g.Count, &g.Sites); err != nil {
			rows.Close()
			return im, err
		}
		im.Gaps = append(im.Gaps, g)
	}
	rows.Close()
	return im, rows.Err()
}

// renderDigest is pure — tested directly. Every section states its window so
// an empty section reads as "no activity", never as "not checked".
func renderDigest(hours int, now time.Time, runs []digestRun, corrs []digestCorrelation, immune digestImmune, snaps []digestSnapshot) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Fix-loop digest — last %dh (generated %s UTC)\n\n", hours, now.Format("2006-01-02 15:04"))

	fmt.Fprintf(&b, "## Runs (%d)\n\n", len(runs))
	if len(runs) == 0 {
		b.WriteString("No fix-loop runs in this window.\n\n")
	} else {
		for _, r := range runs {
			extra := ""
			if r.GateNote != "" {
				extra += " — " + r.GateNote
			}
			if r.PRNote != "" {
				extra += " — " + r.PRNote
			}
			fmt.Fprintf(&b, "- %s  `%s`  %s → %s (%s)%s\n", r.CreatedAt, r.AgentType, r.Status, r.Terminal, shortID(r.OrchID), extra)
		}
		b.WriteString("\n")
	}

	fmt.Fprintf(&b, "## Decisions by correlation (%d)\n\n", len(corrs))
	if len(corrs) == 0 {
		b.WriteString("No diagnosis artifacts written in this window.\n\n")
	} else {
		for _, c := range corrs {
			line := fmt.Sprintf("- `%s`: %s", shortID(c.CorrelationID), c.Kinds)
			if c.Decision != "" {
				line += fmt.Sprintf(" — latest council: **%s** (%s)", c.Decision, c.DecidedBy)
			}
			b.WriteString(line + "\n")
		}
		b.WriteString("\n")
	}

	// Escalation channel — the immune system on the same page. The diagnosis
	// queue always shows EVERY open item: a parked escalation is a decision
	// waiting on the owner and must not fade out of the digest with time.
	newCount := 0
	for _, e := range immune.Escalations {
		if e.New {
			newCount++
		}
	}
	fmt.Fprintf(&b, "## Escalation channel — immune system → loop (%d open, %d new)\n\n",
		len(immune.Escalations), newCount)
	fmt.Fprintf(&b, "Sweeps this window: %d triage, %d silent-check (full reports: doc_notes categories `triage` / `silent-check`).\n\n",
		immune.TriageSweeps, immune.SilentSweeps)

	if len(immune.Escalations) == 0 {
		b.WriteString("Diagnosis queue empty — nothing is waiting on a dispatch decision.\n\n")
	} else {
		b.WriteString("**Diagnosis queue** (parked inert; dispatching is the owner's call):\n\n")
		for _, e := range immune.Escalations {
			marker := ""
			if e.New {
				marker = "NEW "
			}
			fmt.Fprintf(&b, "- %s%s  `%s` [%s] — %s\n", marker, e.CreatedAt, e.ItemKey, e.Status, e.Summary)
		}
		b.WriteString("\n")
	}

	if len(immune.SilentOpen) == 0 && len(immune.SilentClosed) == 0 {
		b.WriteString("No silent-failure findings open or closed in this window.\n\n")
	} else {
		fmt.Fprintf(&b, "**Silent-failure findings** (%d open, %d closed this window):\n\n",
			len(immune.SilentOpen), len(immune.SilentClosed))
		for _, f := range immune.SilentOpen {
			fmt.Fprintf(&b, "- OPEN `%s` — %s\n", f.ItemKey, f.Summary)
		}
		for _, f := range immune.SilentClosed {
			fmt.Fprintf(&b, "- CLOSED `%s` — %s\n", f.ItemKey, f.Summary)
		}
		b.WriteString("\n")
	}

	if len(immune.Gaps) == 0 {
		b.WriteString("No standing capability gaps.\n\n")
	} else {
		b.WriteString("**Capability gaps** (roadmap decisions, never the loop):\n\n")
		for _, g := range immune.Gaps {
			fmt.Fprintf(&b, "- **%s** needed — %d item(s) across %d site(s).\n", g.Builder, g.Count, g.Sites)
		}
		b.WriteString("\n")
	}

	fmt.Fprintf(&b, "## Agent config changes (%d)\n\n", len(snaps))
	if len(snaps) == 0 {
		b.WriteString("No agent definitions were changed in this window.\n\n")
	} else {
		b.WriteString("Every seeded/updated agent leaves a snapshot with its reason — this is the ledger of changes to the machine itself.\n\n")
		for _, s := range snaps {
			fmt.Fprintf(&b, "- %s  `%s`: %s\n", s.TakenAt, s.AgentType, s.Reason)
		}
		b.WriteString("\n")
	}

	b.WriteString("---\n_Deterministic digest: facts gathered by SQL, rendered by Go — no model in this path. Full trails: diagnosis_artifacts / orchestration_states by the ids above._\n")
	return b.String()
}

func shortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

// pqStringArray renders a []string as a Postgres array literal for ANY($n).
// lib/pq's array support is avoided to keep this file dependency-free.
func pqStringArray(items []string) string {
	escaped := make([]string, len(items))
	for i, s := range items {
		escaped[i] = `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
	}
	return "{" + strings.Join(escaped, ",") + "}"
}
