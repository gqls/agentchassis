// FILE: platform/orchestration/actions/diagnose_silent_check_action.go
//
// SILENT-CHECK — Phase 2 of the triage design (DESIGN_triage_and_escalation.md
// §"the hard case"): the structural verification checker for defects that NO
// work-item completion ever touches. The completion gate
// (complete_work_item_verification.go, the empty-sections thread) already
// de-silences every item_type with a registered verifier, and insertWorkItem's
// two-strike rule already catches recurrence — this checker exists for the
// remaining class: the platform's own invariants violated in observable state
// while site_work_items stays silent (the darts guides-index class: a page in
// the site's navigation that was never built, with no work item anywhere).
//
// DELIBERATELY DETERMINISTIC: no LLM. It verifies invariants; it never
// diagnoses and never fixes. Per sweep it:
//   - gathers violations per check, EXCLUDING any page some work item already
//     references in a non-closed status (those are the immune system's turf —
//     this checker only reports what nothing else can see);
//   - aggregates per (check, site) and — live — writes ONE site_work_items row
//     per (check, site) at item_type='silent_failure', status='failed' (inert:
//     terminal, unclaimable), so the EXISTING triage sweep routes it under its
//     own cap. The error text leads with a fixed ≥140-char signature so
//     triage's left(error,140) grouping collapses all sites into ONE
//     platform-level pattern (fix the cause once, not per site);
//   - dedupes its own emissions explicitly (NOT EXISTS on site+item_key+
//     status='failed' — idx_swi_dedup EXCLUDES failed rows, so ON CONFLICT
//     cannot do this) and touches updated_at on persisting violations so they
//     stay inside triage's window for as long as they remain true;
//   - closes its own items ('complete') when the violation is observed gone —
//     the minimal honest slice of Phase 3's feedback close-out;
//   - writes one doc_note per sweep (categories silent-check+fixloop).
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

// silentCheckSource labels everything this checker writes.
const silentCheckSource = "silent-check"

// The check registry. EmitDefault=false checks are report-only until the owner
// flips them (check 2 stays report-only in v1: a deployed page with zero
// components can be a deliberate gutting — e.g. audited-out content — so it
// needs human eyes on the report before it may escalate).
type silentCheck struct {
	Name        string
	Description string // one line for the report
	EmitDefault bool
	Gather      func(ctx context.Context, db *sql.DB, grace string) ([]silentViolation, error)
}

type silentViolation struct {
	SiteID   string
	Domain   string
	PageID   string
	PageName string
	PageType string
	Build    string
}

// silentChecks — start with the two named in the design; check 1 is the darts
// signature and the only emitter in v1.
var silentChecks = []silentCheck{
	{
		Name:        "nav_linked_never_built",
		Description: "page linked in the site's navigation, never built (build_status='planned' beyond the grace period), and no work item anywhere references it",
		EmitDefault: true,
		Gather:      gatherNavLinkedNeverBuilt,
	},
	{
		Name:        "deployed_zero_components",
		Description: "page built/deployed but serving zero components, and no work item anywhere references it (REPORT-ONLY: may be a deliberate content removal)",
		EmitDefault: false,
		Gather:      gatherDeployedZeroComponents,
	},
}

var DiagnoseSilentCheckInputSpec = datahelpers.ActionInputSpec{
	Required: []string{},
	Optional: []string{
		"grace_hours", "settle_hours", "max_emit", "dry_run", "emit_checks",
	},
	Defaults: map[string]interface{}{
		"grace_hours":  48,    // a planned page younger than this may just be queued
		"settle_hours": 2,     // a built page younger than this may be mid-assembly
		"max_emit":     3,     // hard cap per sweep while confidence builds (mirrors triage)
		"dry_run":      false, // when true: report only — no items written, none closed
	},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("diagnose_silent_check", DiagnoseSilentCheckInputSpec)
}

// DiagnoseSilentCheckAction runs one verification sweep.
func DiagnoseSilentCheckAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	logger := params.Logger.With(zap.String("action", "diagnose_silent_check"))
	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("diagnose_silent_check: no DB handle")
	}

	graceH := datahelpers.GetIntField(config, "grace_hours", 48)
	if graceH <= 0 {
		graceH = 48
	}
	settleH := datahelpers.GetIntField(config, "settle_hours", 2)
	if settleH <= 0 {
		settleH = 2
	}
	maxEmit := datahelpers.GetIntField(config, "max_emit", 3)
	dryRun := false
	if d, ok := config["dry_run"].(bool); ok {
		dryRun = d
	}
	emitOverride := configStringSlice(config, "emit_checks", nil)

	type checkResult struct {
		Check      silentCheck
		Emits      bool
		Violations []silentViolation
	}
	var results []checkResult
	violationKeys := map[string]bool{} // every (check,site) key still true this sweep
	for _, c := range silentChecks {
		grace := fmt.Sprintf("%d hours", graceH)
		if c.Name == "deployed_zero_components" {
			grace = fmt.Sprintf("%d hours", settleH)
		}
		vs, err := c.Gather(ctx, params.DB, grace)
		if err != nil {
			return nil, fmt.Errorf("silent-check %s: %w", c.Name, err)
		}
		emits := c.EmitDefault
		if emitOverride != nil {
			emits = false
			for _, name := range emitOverride {
				if strings.TrimSpace(name) == c.Name {
					emits = true
				}
			}
		}
		results = append(results, checkResult{Check: c, Emits: emits, Violations: vs})
		for _, key := range silentGroupKeys(c.Name, vs) {
			violationKeys[key] = true
		}
	}

	// Emit (live, emitting checks only): one item per (check, site), capped,
	// deduped by explicit NOT EXISTS; persisting violations get their
	// updated_at touched so triage's window never ages them out while true.
	created, deduped, capped := 0, 0, 0
	var emittedKeys []string
	if !dryRun {
		for _, r := range results {
			if !r.Emits {
				continue
			}
			for _, g := range silentGroupBySite(r.Check.Name, r.Violations) {
				if created >= maxEmit {
					capped++
					continue
				}
				ins, err := silentInsertItem(ctx, params.DB, r.Check, g, graceH)
				if err != nil {
					logger.Warn("silent-check: emit failed", zap.String("item_key", g.ItemKey), zap.Error(err))
					continue
				}
				if ins {
					created++
					emittedKeys = append(emittedKeys, g.ItemKey)
				} else {
					deduped++
				}
			}
		}
	}

	// Close-out: any of our still-failed items whose violation is no longer
	// observed gets honestly completed (and never blocks a future re-emission).
	closed := 0
	if !dryRun {
		var err error
		closed, err = silentCloseResolved(ctx, params.DB, violationKeys)
		if err != nil {
			logger.Warn("silent-check: close-resolved failed", zap.Error(err))
		}
	}

	now := time.Now().UTC()
	var sections []silentReportSection
	total := 0
	for _, r := range results {
		sections = append(sections, silentReportSection{
			Name: r.Check.Name, Description: r.Check.Description,
			Emits: r.Emits, Violations: r.Violations,
		})
		total += len(r.Violations)
	}
	report := renderSilentCheck(now, graceH, sections, created, deduped, capped, maxEmit, closed, dryRun)

	noteID, nerr := insertDocNote(ctx, params.DB, "pipeline", "diagnose", "",
		report, `["silent-check","fixloop"]`, "diagnose_silent_check", nullSafeAgentType(params), "", "diagnose_silent_check")
	if nerr != nil {
		logger.Warn("silent-check: doc_note persist failed (report still returned)", zap.Error(nerr))
	}

	logger.Info("diagnose_silent_check: swept",
		zap.Int("violations", total),
		zap.Int("emitted", created),
		zap.Int("deduped", deduped),
		zap.Int("capped", capped),
		zap.Int("closed_resolved", closed),
		zap.Bool("dry_run", dryRun),
		zap.String("note_id", noteID),
		zap.String("orchestration_id", orchIDForLog(params)))

	return map[string]interface{}{
		"report":          report,
		"violations":      total,
		"emitted":         created,
		"deduped":         deduped,
		"capped":          capped,
		"closed_resolved": closed,
		"emitted_keys":    emittedKeys,
		"note_id":         noteID,
	}, nil
}

// silentCoverageClause — the "nothing else can see it" filter, shared by every
// check. A page is COVERED (and therefore not silent) if any work item
// references it — by page_id column, by spec->>'page_id', or by item_key
// segment — in any status except the closed ones below. 'complete' does NOT
// cover: a completed remedy with the violation still observable is precisely
// the remediation-ineffective signal. 'cancelled'/'rejected' were retracted,
// not resolved. Everything else (open, failed, unresolved, needs_human_review,
// wont_fix) means the immune system or a human already has it.
const silentCoverageClause = `
	NOT EXISTS (
		SELECT 1 FROM site_work_items w
		WHERE w.site_id = p.site_id
		  AND w.status NOT IN ('complete','cancelled','rejected')
		  AND (w.page_id = p.id
		       OR w.spec->>'page_id' = p.id::text
		       OR w.item_key LIKE '%:' || p.name
		       OR w.item_key LIKE '%:' || p.name || ':%')
	)`

func gatherNavLinkedNeverBuilt(ctx context.Context, db *sql.DB, grace string) ([]silentViolation, error) {
	return silentGatherQuery(ctx, db, `
		SELECT p.site_id, s.domain, p.id, p.name, COALESCE(p.page_type,''), COALESCE(p.build_status,'')
		FROM pages p JOIN sites s ON s.id = p.site_id
		WHERE (p.in_header OR p.in_footer)
		  AND p.build_status = 'planned'
		  AND p.created_at < now() - $1::interval
		  AND `+silentCoverageClause+`
		ORDER BY s.domain, p.name`, grace)
}

func gatherDeployedZeroComponents(ctx context.Context, db *sql.DB, settle string) ([]silentViolation, error) {
	return silentGatherQuery(ctx, db, `
		SELECT p.site_id, s.domain, p.id, p.name, COALESCE(p.page_type,''), COALESCE(p.build_status,'')
		FROM pages p JOIN sites s ON s.id = p.site_id
		WHERE p.build_status IN ('built','deployed')
		  AND COALESCE(p.updated_at, p.created_at) < now() - $1::interval
		  AND NOT EXISTS (SELECT 1 FROM page_components pc WHERE pc.page_id = p.id)
		  AND `+silentCoverageClause+`
		ORDER BY s.domain, p.name`, settle)
}

func silentGatherQuery(ctx context.Context, db *sql.DB, query, grace string) ([]silentViolation, error) {
	rows, err := db.QueryContext(ctx, query, grace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []silentViolation
	for rows.Next() {
		var v silentViolation
		if err := rows.Scan(&v.SiteID, &v.Domain, &v.PageID, &v.PageName, &v.PageType, &v.Build); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

type silentGroup struct {
	SiteID  string
	Domain  string
	ItemKey string
	Pages   []silentViolation
}

// silentItemKey is the STABLE per-(check, site) dedup key.
func silentItemKey(check, siteID string) string {
	id := siteID
	if len(id) > 8 {
		id = id[:8]
	}
	return fmt.Sprintf("silent:%s:%s", check, id)
}

func silentGroupKeys(check string, vs []silentViolation) []string {
	seen := map[string]bool{}
	var out []string
	for _, v := range vs {
		k := silentItemKey(check, v.SiteID)
		if !seen[k] {
			seen[k] = true
			out = append(out, k)
		}
	}
	return out
}

func silentGroupBySite(check string, vs []silentViolation) []silentGroup {
	bySite := map[string]*silentGroup{}
	var order []string
	for _, v := range vs {
		g, ok := bySite[v.SiteID]
		if !ok {
			g = &silentGroup{SiteID: v.SiteID, Domain: v.Domain, ItemKey: silentItemKey(check, v.SiteID)}
			bySite[v.SiteID] = g
			order = append(order, v.SiteID)
		}
		g.Pages = append(g.Pages, v)
	}
	sort.Strings(order)
	var out []silentGroup
	for _, id := range order {
		out = append(out, *bySite[id])
	}
	return out
}

// silentErrorPrefix — the fixed leading signature per check. MUST be at least
// 140 characters (unit-tested), because triage groups failure patterns by
// left(error, 140): with the prefix ≥140 chars, every site's item collapses
// into ONE platform-level pattern and the loop diagnoses the cause once.
func silentErrorPrefix(check silentCheck) string {
	return fmt.Sprintf("silent-check %s: structural invariant violated with NO covering work item — %s; detected by the silent-failure verification checker (fixloop Phase 2)",
		check.Name, check.Description)
}

func silentErrorText(check silentCheck, g silentGroup) string {
	names := make([]string, 0, len(g.Pages))
	for _, p := range g.Pages {
		names = append(names, p.PageName)
	}
	return fmt.Sprintf("%s | site=%s pages=%s", silentErrorPrefix(check), g.Domain, strings.Join(names, ","))
}

func silentSummary(check silentCheck, g silentGroup) string {
	names := make([]string, 0, len(g.Pages))
	for _, p := range g.Pages {
		names = append(names, p.PageName)
	}
	return fmt.Sprintf("SILENT %s: %d page(s) on %s violate the invariant with no covering work item: %s",
		check.Name, len(g.Pages), g.Domain, strings.Join(names, ", "))
}

func silentSpecJSON(check silentCheck, g silentGroup, graceH int) string {
	pages := make([]map[string]string, 0, len(g.Pages))
	for _, p := range g.Pages {
		pages = append(pages, map[string]string{
			"page_id": p.PageID, "name": p.PageName, "page_type": p.PageType, "build_status": p.Build,
		})
	}
	b, _ := json.Marshal(map[string]interface{}{
		"check":         check.Name,
		"invariant":     check.Description,
		"site_domain":   g.Domain,
		"pages":         pages,
		"grace_hours":   graceH,
		"coverage_rule": "no work item referencing the page (page_id / spec page_id / item_key segment) in a status other than complete/cancelled/rejected",
		"source":        silentCheckSource,
	})
	return string(b)
}

// silentInsertItem writes one (check, site) finding as an INERT loud failure
// for triage to route. Explicit NOT EXISTS dedup (idx_swi_dedup excludes
// status='failed', so ON CONFLICT cannot collapse these); when the item
// already exists its updated_at is touched so a persisting violation never
// ages out of triage's sweep window. Returns true only on a fresh insert.
func silentInsertItem(ctx context.Context, db *sql.DB, check silentCheck, g silentGroup, graceH int) (bool, error) {
	res, err := db.ExecContext(ctx, `
		INSERT INTO site_work_items (
			site_id, source, pipeline, item_type, severity, summary,
			spec, priority, handler_agent, status, error, created_by, item_key, max_attempts
		)
		SELECT $1, $2, 'maintenance', 'silent_failure', 'high', $3,
		       $4::jsonb, 40, '', 'failed', $5, $2, $6, 1
		WHERE NOT EXISTS (
			SELECT 1 FROM site_work_items w
			WHERE w.site_id = $1 AND w.item_key = $6 AND w.status = 'failed'
		)`,
		g.SiteID, silentCheckSource, silentSummary(check, g),
		silentSpecJSON(check, g, graceH), silentErrorText(check, g), g.ItemKey)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	if n > 0 {
		return true, nil
	}
	_, err = db.ExecContext(ctx, `
		UPDATE site_work_items SET updated_at = now(), summary = $3, spec = $4::jsonb, error = $5
		WHERE site_id = $1 AND item_key = $2 AND status = 'failed' AND created_by = $6`,
		g.SiteID, g.ItemKey, silentSummary(check, g), silentSpecJSON(check, g, graceH), silentErrorText(check, g), silentCheckSource)
	return false, err
}

// silentCloseResolved completes any of our still-failed items whose
// (check, site) violation was NOT observed this sweep — the minimal honest
// slice of the design's Phase 3 close-out. Closed items leave triage's window
// and, being non-failed, never block a future re-emission if the violation
// returns.
func silentCloseResolved(ctx context.Context, db *sql.DB, stillTrue map[string]bool) (int, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT site_id, item_key FROM site_work_items
		WHERE created_by = $1 AND status = 'failed'`, silentCheckSource)
	if err != nil {
		return 0, err
	}
	type ref struct{ site, key string }
	var stale []ref
	for rows.Next() {
		var r ref
		if err := rows.Scan(&r.site, &r.key); err != nil {
			rows.Close()
			return 0, err
		}
		if !stillTrue[r.key] {
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
			    error = COALESCE(error,'') || ' [silent-check: invariant no longer violated; closed]'
			WHERE site_id = $1 AND item_key = $2 AND status = 'failed' AND created_by = $3`,
			r.site, r.key, silentCheckSource)
		if err != nil {
			return closed, err
		}
		if n, _ := res.RowsAffected(); n > 0 {
			closed++
		}
	}
	return closed, nil
}

type silentReportSection struct {
	Name        string
	Description string
	Emits       bool
	Violations  []silentViolation
}

// renderSilentCheck is pure — tested. The owner-readable sweep artifact.
func renderSilentCheck(now time.Time, graceH int, sections []silentReportSection, created, deduped, capped, maxEmit, closed int, dryRun bool) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Silent-check sweep (generated %s UTC; grace %dh)\n\n", now.Format("2006-01-02 15:04"), graceH)
	if dryRun {
		b.WriteString("_DRY RUN — findings below are report-only this sweep: NO silent_failure items were written, none were closed._\n\n")
	}

	for _, s := range sections {
		mode := "report-only"
		if s.Emits {
			mode = "emits to triage"
		}
		fmt.Fprintf(&b, "## %s — %d violation(s) [%s]\n\n", s.Name, len(s.Violations), mode)
		fmt.Fprintf(&b, "_Invariant: %s._\n\n", s.Description)
		if len(s.Violations) == 0 {
			b.WriteString("None observed.\n\n")
			continue
		}
		byDomain := map[string][]silentViolation{}
		var domains []string
		for _, v := range s.Violations {
			if _, ok := byDomain[v.Domain]; !ok {
				domains = append(domains, v.Domain)
			}
			byDomain[v.Domain] = append(byDomain[v.Domain], v)
		}
		sort.Strings(domains)
		for _, d := range domains {
			vs := byDomain[d]
			names := make([]string, 0, len(vs))
			for _, v := range vs {
				names = append(names, fmt.Sprintf("`%s` (%s, %s)", v.PageName, v.PageType, v.Build))
			}
			fmt.Fprintf(&b, "- **%s** — %d page(s): %s\n", d, len(vs), strings.Join(names, ", "))
		}
		b.WriteString("\n")
	}

	if !dryRun {
		fmt.Fprintf(&b, "## Bookkeeping\n\nEmitted %d, deduped (already open) %d, capped %d (cap=%d); closed as resolved %d.\n",
			created, deduped, capped, maxEmit, closed)
		if capped > 0 {
			fmt.Fprintf(&b, "\n> %d finding(s) NOT emitted this sweep (cap=%d) — coverage was capped, not complete.\n", capped, maxEmit)
		}
		b.WriteString("\n")
	}

	b.WriteString("---\n_Deterministic verification checker: SQL-gathered, Go-rendered, no model. Findings emit as INERT `silent_failure` items (status='failed', unclaimable); the triage sweep routes them under its own cap. A finding only appears here if NO other work item references the page — everything visible to the immune system stays the immune system's._\n")
	return b.String()
}
