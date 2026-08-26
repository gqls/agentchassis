// FILE: platform/orchestration/actions/request_render_audit_cursor.go
//
// The coverage cursor for `request_render_audit` — bugs_open/394.
//
// THE DEFECT THIS EXISTS FOR. The audit selects a site's live pages
// `ORDER BY COALESCE(nav_order,999), name` and takes the first `max_pages`. It
// takes the SAME deterministic prefix on every run, so pages past the cap are
// not audited less often — they have never been audited and never will be.
// `bugs_closed/242` made that loud (the durable RENDER_AUDIT_TRUNCATED row) and
// raised the cap 25→60 as a stated mitigation. Measured 2026-08-26,
// webdesign.co.uk is 146 live pages / 60 audited / 86 never, and the tail is a
// CLASS rather than a count: its nav_order bands are 0..90 (6 nav pages), 100
// (94 tools, alphabetical), 200 (48 `tool-*-guide`), 201 (1). The cap cuts
// between `tool-head-architect` and `tool-html-minifier`, so all 45 remaining
// guide pages are unreachable at ANY cap below 98 — on a site that grew 15 pages
// in two days. A constant cannot chase that, which is why the bug's "raise the
// cap" candidate is rejected on a measurement rather than a preference.
//
// ── WHY A CURSOR IS NOT THE RANDOM SWEEP 242 FORBADE ───────────────────────
//
// `bugs_closed/242` §5 says in terms: do NOT sweep in random page order to
// spread the misses. This is not that. A keyset cursor is DETERMINISTIC rotation
// with a recorded window: within any single run the gap is contiguous and
// reportable (`window_first`/`window_last`), and across runs the union
// converges. Randomisation destroys both properties; this preserves both.
//
// ── THE PRIORITY SET, AND THE OWNER RULING IT PROTECTS ─────────────────────
//
// A PLAIN cursor would take webdesign's per-page re-measurement latency to
// 3 days × ceil(146/60) ≈ 9 days. Migration 469 is an OWNER INSTRUCTION of
// 2026-08-18 that cut the window 7d→3d, and its stated why is that the render
// audit is the only thing that GRADES a contrast repair, so its eligibility
// window IS the confirmation latency of the whole repair loop — seven days was
// ruled unacceptable. A plain cursor therefore would not merely underperform
// that ruling; it would EXCEED the condition the owner ordered removed.
//
// It dissolves rather than trades, because the population 469 protects is not
// the site — it is the pages awaiting a grade. Measured 2026-08-26 with the
// grader's own predicate, webdesign.co.uk has 3 open contrast_failure items
// across 3 paths (fleet maximum 17, on a 37-page site that never truncates at
// cap 60). So the window is a UNION: the finding-bearing pages ride in every
// run, the cursor covers the rest, and the cost is 3 of 60 slots.
//
// ⚠ THE PRIORITY SET HAS EXACTLY ONE ITEM TYPE, BY CONSTRUCTION rather than by
// choice. It is defined as "the pages whose open findings THIS reply can close",
// and `write_render_audit_findings_action.go:791` makes exactly one retraction
// call on this payload's path — `loadAuditRetractionCandidates(…,
// "contrast_failure")`. `undeployed_asset` is excluded for three independent
// reasons, any one sufficient: nothing on this path grades it; its key is
// `undeployed_asset:<asset_id>` so there is no page to prioritise (measured
// 2026-08-26: all 190 open rows carry a NULL page_id, against 111 of 111 for
// contrast_failure); and its verifier is the asset being deployed, observed
// elsewhere. Do not widen this set without first finding a second retraction
// call on this path.
//
// ── THE BOUND IS THE DESIGN ────────────────────────────────────────────────
//
// The priority set is capped at `max_pages/2`, so the rotation always keeps at
// least `ceil(max_pages/2)` slots and coverage can NEVER stall. Without that, a
// site with more finding-bearing pages than the cap would spend every run on the
// same priority pages and never advance — the deterministic-prefix disease
// reappearing one level down, inside the fix for it. When the bound bites, the
// dropped paths are NAMED in the log and counted in the durable context, and the
// selection is cyclic from the cursor so the dropped excess is not the same
// pages every run.
//
// ── WHAT THIS FILE DOES NOT DECIDE ─────────────────────────────────────────
//
// Advancement TIMING lives at the call site, and is deliberate: the cursor is
// written AFTER a successful produce, which is the opposite ordering from the
// truncation row (written BEFORE the send, so a failed dispatch cannot unrecord
// it — bugs_open/242). A cursor is a commitment about the NEXT run; written
// before a produce that then fails, it would skip a window nothing ever
// requested. And advancement is at DISPATCH, not at reply: that is this estate's
// existing ruling for this rotation family (migration 346 — "a site whose run
// fails must not pin the rotation head and starve the fleet"), and the failure
// modes are not symmetric. Dispatch-advance skips one window for one cycle,
// visibly. Reply-advance would retry a page that reliably wedges Chromium FOR
// EVER and starve everything behind it — the unbounded silent stall, strictly
// worse than the prefix this bug is about.
package actions

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// auditPageRow is one live, shipped page in the audit's own ordering.
type auditPageRow struct {
	URL  string // absolute, as sent to the adapter
	Path string // pages.url as recorded — what item_key carries
	Ord  int    // COALESCE(nav_order, 999): the COALESCED value, never the raw one
	Name string
}

// auditCursor is a keyset position in `ORDER BY COALESCE(nav_order,999), name`.
//
// It stores the ORDERING TUPLE, never an index and never the cap. Storing an
// index would break the moment a page is added or removed; storing the cap would
// corrupt on a per-dispatch override, and `max_pages` IS per-dispatch — the
// 2026-08-11 `5 of 26` row carries {"max_pages": 5} under render-audit-agent,
// which is what resolves bugs_open/394's own [UNEXPLAINED] note.
type auditCursor struct {
	Ord  int
	Name string
}

// priorityResult is the regrade set actually assembled for one run.
type priorityResult struct {
	taken    []auditPageRow // in the run, ahead of the rotation slice
	dropped  []string       // beyond the reserve THIS run; rotates with the cursor
	notLive  int            // open findings whose page is no longer live/shipped
	openRows int            // open contrast_failure rows backing the set
}

func (p priorityResult) paths() []string {
	out := make([]string, 0, len(p.taken))
	for _, r := range p.taken {
		out = append(out, r.Path)
	}
	return out
}

// afterCursor reports whether r sorts strictly after cur in the audit ordering.
func afterCursor(r auditPageRow, cur auditCursor) bool {
	if r.Ord != cur.Ord {
		return r.Ord > cur.Ord
	}
	return r.Name > cur.Name
}

// selectAuditWindow returns ONE CONTIGUOUS slice of the ordering starting past
// `cur`, skipping anything already in `skip`, plus the cursor for the next run.
//
// A nil `next` means the cycle completed and the caller must DELETE the row, so
// the following capped run starts from the top again.
//
// THE BOUNDARY IS FOUND BY EXACT MATCH FIRST, then by first-strictly-greater.
// That ordering is load-bearing: exact match is collation-proof, and the
// fallback is what stops a DELETED or RENAMED cursor page either stalling the
// rotation for ever or restarting it from the top. Both failure modes are
// silent, and both look exactly like a working cursor from outside.
func selectAuditWindow(rows []auditPageRow, cur *auditCursor, n int, skip map[string]bool) ([]auditPageRow, *auditCursor) {
	if n <= 0 || len(rows) == 0 {
		return nil, nil
	}

	start := 0
	if cur != nil {
		start = len(rows) // past-the-end unless a successor is found
		for i, r := range rows {
			if r.Ord == cur.Ord && r.Name == cur.Name {
				start = i + 1
				break
			}
			if afterCursor(r, *cur) {
				start = i // the cursor's own page is gone: take its successor's slot
				break
			}
		}
		if start >= len(rows) {
			// The cursor sorts past everything left (the site shrank, or renames
			// moved the tail). Restart from the top THIS run rather than
			// returning nothing.
			start = 0
			cur = nil
		}
	}

	out := make([]auditPageRow, 0, n)
	i := start
	for ; i < len(rows) && len(out) < n; i++ {
		if skip[rows[i].Path] {
			continue // already carried by the priority set: never send a URL twice
		}
		out = append(out, rows[i])
	}

	if len(out) == 0 {
		// Everything past the boundary was already in the priority set. Falling
		// through with an empty window would reach the caller's len(urls)==0
		// branch, which reports `no_deployed_pages` — a FALSE SKIP on a site that
		// is merely fully covered this run. Clear the cursor instead: the next
		// run starts at the top.
		return nil, nil
	}

	if i >= len(rows) {
		return out, nil // final window of the cycle
	}
	last := out[len(out)-1]
	return out, &auditCursor{Ord: last.Ord, Name: last.Name}
}

// cyclicFrom returns rows rotated so that the first element is the one just past
// `cur`. Used only to ORDER the priority candidates, so that when the reserve
// bound bites the dropped excess is not the same pages on every run — the
// deterministic-prefix disease this whole change exists to remove, in miniature.
func cyclicFrom(rows []auditPageRow, cur *auditCursor) []auditPageRow {
	if cur == nil || len(rows) == 0 {
		return rows
	}
	split := 0
	for i, r := range rows {
		if afterCursor(r, *cur) {
			split = i
			break
		}
	}
	if split == 0 {
		return rows
	}
	out := make([]auditPageRow, 0, len(rows))
	out = append(out, rows[split:]...)
	out = append(out, rows[:split]...)
	return out
}

// openContrastPagePaths returns the page paths carrying an OPEN contrast_failure
// row — the exact population `retractResolvedContrastFindings` can grade.
//
// The status predicate interpolates sqlInList(workItemClosedStatuses), the SAME
// constant `loadAuditRetractionCandidates` uses, rather than restating a status
// list. That is not tidiness: `workItemClosedStatuses` is deliberately NOT
// `workItemTerminalStatuses` — `unresolved` and `failed` are OPEN (RFC_010,
// owner ruling 2026-08-02 "Decision 2") — and a hand-written copy of the list
// would agree with it until the day it did not. A session censusing this
// population by hand on 2026-08-26 omitted `verified` and `wont_fix` and got the
// right answer anyway, by site-specific coincidence (WRONG_CALLS.md).
//
// The path is recovered with pagePathFromContrastKey, the inverse of the
// composer, because the grader identifies a page by the path inside item_key and
// has never read page_id.
func openContrastPagePaths(ctx context.Context, db *sql.DB, siteID string) (map[string]bool, int, error) {
	q := `SELECT COALESCE(item_key, '')
	        FROM site_work_items
	       WHERE site_id = $1::uuid
	         AND item_type = 'contrast_failure'
	         AND status NOT IN (` + sqlInList(workItemClosedStatuses) + `)`
	rows, err := db.QueryContext(ctx, q, siteID)
	if err != nil {
		return nil, 0, fmt.Errorf("open contrast_failure lookup failed: %w", err)
	}
	defer rows.Close()

	paths := map[string]bool{}
	n := 0
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, 0, fmt.Errorf("open contrast_failure scan: %w", err)
		}
		n++
		if p := pagePathFromContrastKey(key); p != "" {
			paths[p] = true
		}
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("open contrast_failure rows: %w", err)
	}
	return paths, n, nil
}

// selectPriorityRegradeSet picks the finding-bearing pages to carry in this run.
//
// It INTERSECTS with the live page set rather than trusting the work items: an
// open finding on a page that has since been archived or was never shipped must
// not be injected, because it would hand the adapter a dead navigation and put a
// page outside the audit's own population into the request. Those are counted
// (`notLive`) rather than dropped silently — a finding whose page has left the
// population can never self-grade, which is worth surfacing even though this
// change does not act on it.
func selectPriorityRegradeSet(
	ctx context.Context,
	db *sql.DB,
	siteID string,
	live []auditPageRow,
	cur *auditCursor,
	budget int,
	logger *zap.Logger,
) priorityResult {
	var res priorityResult
	if db == nil || budget <= 0 {
		return res
	}

	paths, openRows, err := openContrastPagePaths(ctx, db, siteID)
	if err != nil {
		// Fail OPEN, loudly: a priority set we could not compute must degrade to
		// a plain rotation, never block the audit. The run still covers a window.
		logger.Warn("request_render_audit: could not compute the priority regrade set — this run rotates without it",
			zap.Error(err))
		return res
	}
	res.openRows = openRows
	if len(paths) == 0 {
		return res
	}

	matched := 0
	for _, r := range cyclicFrom(live, cur) {
		if !paths[r.Path] {
			continue
		}
		matched++
		if len(res.taken) >= budget {
			res.dropped = append(res.dropped, r.Path)
			continue
		}
		res.taken = append(res.taken, r)
	}
	res.notLive = len(paths) - matched

	if len(res.dropped) > 0 {
		logger.Warn("request_render_audit: priority regrade set exceeds its reserve — the excess is NOT audited this run; the drop rotates with the cursor, so no page is dropped permanently",
			zap.Int("budget", budget),
			zap.Int("finding_bearing_live_pages", matched),
			zap.Strings("dropped_paths", res.dropped))
	}
	if res.notLive > 0 {
		logger.Info("request_render_audit: open contrast_failure findings name pages that are no longer live — they cannot self-grade",
			zap.Int("not_live", res.notLive))
	}
	return res
}

// --- cursor persistence -----------------------------------------------------
//
// The cursor lives in its own table rather than as columns on
// site_discovery_rotation. That table's own COMMENT declares it "Written by the
// site-discovery-rotation-* pre_queries; safe to TRUNCATE" and it has never had
// a Go writer; its last_selected_at is NOT NULL with ruled semantics ("when the
// scheduler last PICKED the site") that an action-side UPSERT would have to
// invent a value for; and its agent_type means "a scheduled task's target",
// which coincides with "the agent that dispatched with rotation on" today and is
// not guaranteed to.

func loadAuditCursor(ctx context.Context, db *sql.DB, siteID, agentType string) (*auditCursor, error) {
	var c auditCursor
	err := db.QueryRowContext(ctx, `
		SELECT after_nav_order, after_name
		  FROM render_audit_page_cursor
		 WHERE site_id = $1::uuid AND agent_type = $2`, siteID, agentType).Scan(&c.Ord, &c.Name)
	if err == sql.ErrNoRows {
		return nil, nil // no row = start of a cycle, which is today's behaviour
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func saveAuditCursor(ctx context.Context, db *sql.DB, siteID, agentType string, c *auditCursor) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO render_audit_page_cursor (site_id, agent_type, after_nav_order, after_name, updated_at)
		VALUES ($1::uuid, $2, $3, $4, now())
		ON CONFLICT (site_id, agent_type)
		DO UPDATE SET after_nav_order = EXCLUDED.after_nav_order,
		              after_name      = EXCLUDED.after_name,
		              updated_at      = now()`, siteID, agentType, c.Ord, c.Name)
	return err
}

func deleteAuditCursor(ctx context.Context, db *sql.DB, siteID, agentType string) error {
	_, err := db.ExecContext(ctx, `
		DELETE FROM render_audit_page_cursor WHERE site_id = $1::uuid AND agent_type = $2`,
		siteID, agentType)
	return err
}

// --- the durable truncation record -----------------------------------------

// truncationMessage is MODE-SPLIT, and the prefix branch keeps its original
// wording deliberately.
//
// "the unaudited tail is the SAME pages every run" remains LITERALLY TRUE for a
// caller that does not rotate, and design-critique-agent is such a caller by
// design. Rewording it fleet-wide would be the stale-assertion error in the
// other direction — a message asserting a property the code no longer has is
// exactly what this estate keeps logging, and so is a message that has stopped
// asserting a property the code still has.
func truncationMessage(rotate bool, audited, total int, domain string, pri priorityResult, rotationSize int) string {
	if !rotate {
		return fmt.Sprintf("render audit truncated by max_pages: %d of %d live pages audited for %s — the unaudited tail is the SAME pages every run",
			audited, total, domain)
	}
	cycle := 1
	if rotationSize > 0 {
		cycle = (total + rotationSize - 1) / rotationSize
	}
	msg := fmt.Sprintf("render audit truncated by max_pages: %d of %d live pages audited for %s — coverage rotates, this run covering a %d-page window; the whole site is covered over ~%d runs",
		audited, total, domain, rotationSize, cycle)
	if len(pri.taken) > 0 {
		msg += fmt.Sprintf(", plus %d finding-bearing page(s) held in EVERY run so repair grading keeps its cadence", len(pri.taken))
	}
	return msg
}

// truncationContext carries what a reader keys on. The prose above is for
// humans; every machine consumer reads these.
//
// ⚠ THE priority_* KEYS ARE PRESENT ON EVERY CURSOR-MODE ROW, ZEROS AND EMPTY
// ARRAYS INCLUDED. A key that appears only on a bad run makes its absence
// ambiguous between "none happened" and "the binary is too old to count", and
// this estate has already paid for that ambiguity once
// (write_render_audit_findings_action.go:593-596 states the same rule for the
// same reason). The three original keys keep their names and meanings so every
// existing consumer is unaffected.
func truncationContext(rotate bool, audited, total, maxPages int, paths []string, next *auditCursor, window []auditPageRow, pri priorityResult) map[string]interface{} {
	ctx := map[string]interface{}{
		"pages_total":   total,
		"pages_audited": audited,
		"max_pages":     maxPages,
	}
	if !rotate {
		ctx["coverage_mode"] = "prefix"
		return ctx
	}
	ctx["coverage_mode"] = "cursor"
	ctx["audited_paths"] = paths
	ctx["cursor_cleared"] = next == nil
	// window_first/window_last describe the ROTATION slice only — never the
	// priority set, which is out-of-band and reported separately. Folding the two
	// together would make the union-reconstruction query in the bug's acceptance
	// arm silently wrong.
	if len(window) > 0 {
		ctx["window_first"] = window[0].Name
		ctx["window_last"] = window[len(window)-1].Name
	} else {
		ctx["window_first"] = ""
		ctx["window_last"] = ""
	}
	ctx["priority_paths"] = pri.paths()
	ctx["priority_open_items"] = pri.openRows
	ctx["priority_dropped"] = len(pri.dropped)
	ctx["priority_not_live"] = pri.notLive
	return ctx
}

// absoluteAuditURL builds the URL exactly as the dispatch loop always has, so
// the cursor's population and the request's population cannot differ.
func absoluteAuditURL(domain, recorded string) string {
	if strings.HasPrefix(recorded, "http") {
		return recorded
	}
	return "https://" + domain + "/" + strings.TrimPrefix(recorded, "/")
}

// auditedPaths is the full list of page paths this run REQUESTED — priority set
// first, then the rotation window, in the order they were sent.
//
// It is what the bug's acceptance arm integrates over consecutive runs to prove
// the union converges on the whole site. It deliberately reports what was
// REQUESTED, not what the adapter later measured: the cursor governs the
// request, and the measured set (`summary.pages_audited`, which is the retraction
// scope) can be smaller when a navigation fails. Grading the cursor against the
// measured set would blame it for the browser's failures.
func auditedPaths(priority, window []auditPageRow) []string {
	out := make([]string, 0, len(priority)+len(window))
	for _, r := range priority {
		out = append(out, r.Path)
	}
	for _, r := range window {
		out = append(out, r.Path)
	}
	return out
}

// firstWindowName names a window for a log line without the caller having to
// guard the empty case at every site.
func firstWindowName(window []auditPageRow) string {
	if len(window) == 0 {
		return ""
	}
	return window[0].Name
}
