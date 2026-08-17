// FILE: platform/orchestration/actions/retract_page_deployment_action.go
//
// UNPUBLISH — the page-level half of `bugs_open/098`.
//
// THE GAP THIS CLOSES, stated the way `bugs_closed/125` stated it: *the platform
// can publish a page but has no implemented way to unpublish one.* Archiving a
// page (`pages.status='archived'`) correctly removes it from every derivation
// AND from re-rendering — so the last HTML it ever rendered is frozen and keeps
// being served, and no repair path can reach it because every repair path skips
// archived rows on purpose. The live instance the bug was filed on:
// robot-hands.com/learning-center/index.html serves 200 with a listing written
// 2026-07-03 that links to a page serving 404.
//
// Retraction is possible AT ALL because the deploy chain is already reconciling
// on the far side: gqls/sites' deploy-to-b2 workflow runs
// `b2 sync --delete … && purge the Cloudflare zone` on every push. So removing
// the file from the repo removes it from B2 and from the edge. Nothing was ever
// missing downstream — only the ability to remove the file, which is
// git-adapter `delete_file` (added with this).
//
// WHY THIS IS AN ACTION AND NOT A STEP IN AN ARCHIVE HANDLER: nothing in this
// codebase archives a page. Checked across Go and all four frontends — there is
// no writer of `status='archived'` at all; archiving is a hand-run SQL
// operation (052's containment route R6, and the relojistas `contacto` repair).
// So there is no existing seam to hook, and the honest shape is a callable
// platform capability that an operator run, a work-item handler, or a future
// archive path can all reach.
//
// ────────────────────────────────────────────────────────────────────────────
// THE GUARDS ARE THE DESIGN. This deletes live files, so what it CANNOT do
// matters more than what it does.
//
//  1. Paths are only ever derived from `pages.url` through the shared
//     datahelpers.PageFilePathFromURL — the same function the publish side has
//     used since bugs_closed/125. So a retraction can only ever name a file a
//     publish could have written: never an asset, never a stylesheet, never
//     .github/. A page whose url does not designate a file of its own (a
//     fragment, a query, an off-origin url) is DECLINED, never sanitised —
//     125's live case is idea.uk's "/tools.html#audience-check", where
//     /tools.html is a DIFFERENT page's canonical url.
//
//  2. The action re-reads the page rows itself. A caller names pages; it never
//     names files. There is no config path by which a caller-supplied string
//     reaches the adapter as a deletion.
//
//  3. A page is REFUSED if its derived file path is also derived by an `active`
//     page on the same site. Two urls can collide onto one file ("/foo/" and
//     "/foo/index.html"), and retracting the archived one would delete the live
//     one's artefact. Measured fleet-wide 2026-08-02: 0 such collisions exist,
//     so this guard is INERT today — it ships because "no collision exists" and
//     "a collision cannot cause damage" are different properties, and RFC_010 is
//     open on precisely this class.
//
//  4. Only pages the platform says must not be served are eligible:
//     `status <> 'active'`. Never "delete every file with no matching page row"
//     — the repo legitimately holds files `pages` does not model (assets,
//     sitemap.xml, robots.txt, the 404 page).
//
// ────────────────────────────────────────────────────────────────────────────
// WHAT THIS DELIBERATELY DOES NOT DO: it does not clear `pages.deployed_at`.
// 098's candidate 2 asks for that, and the census says the premise behind it is
// wrong: of 49 non-test references, three read the column as HISTORY rather
// than as a boolean — reconcile_superseded_reviews_action.go:98
// (`deployed_at > GREATEST(wi.created_at, …)`), maintenance_actions.go:725
// (findStalePages), and page_admin_handlers.go:101 (shown to a human). Clearing
// it changes what all three return, which is a change to what a shared column
// GUARANTEES — architecture scope by the owner ruling of 2026-07-29 §1, and
// exactly the ground bugs_closed/124 drew a REJECTED verdict on. It is recorded
// on the bug as measured-and-deferred, with the census.
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var RetractPageDeploymentInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{},
	Optional:   []string{"site_id", "site_id_field", "domain_field", "page_ids_field", "repo_name", "dry_run"},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("retract_page_deployment", RetractPageDeploymentInputSpec)
}

// retractionCandidate is one page considered for retraction, and the reason if
// it was not retracted. Refusals are RETURNED, RECORDED in collected_data
// under retractionAuditKey, and — on a real run — persisted as
// agent_error_log rows: a page the platform declines to retract is a page
// that keeps serving, and an operator reading a green result must be able to
// see which ones those are. Returning alone was tried and is NOT enough
// (debt 5, 2026-08-03): this action awaits the adapter's reply, and the await
// machinery replaces the step's own collected_data keys with that reply, so
// the returned refusals survived nowhere but pod logs.
type retractionCandidate struct {
	PageID   string `json:"page_id"`
	PageName string `json:"page_name"`
	URL      string `json:"url"`
	Status   string `json:"status"`
	FilePath string `json:"file_path,omitempty"`
	Refused  string `json:"refused,omitempty"`
}

// RetractPageDeploymentAction removes the deployed artefacts of pages that must
// no longer be served, via the git-adapter's delete_file verb.
//
// Selection is either an explicit page-id list (page_ids_field) or, failing
// that, every non-active page on the site that still carries a deployed_at
// stamp — the exact population 098 measures (13 fleet-wide on 2026-08-02).
func RetractPageDeploymentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	logger := params.Logger.With(zap.String("action", "retract_page_deployment"))
	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("retract_page_deployment requires a database handle")
	}

	siteID, err := resolveRetractionSiteID(params, config)
	if err != nil {
		return nil, err
	}

	var domain string
	if err := params.DB.QueryRowContext(ctx,
		`SELECT domain FROM sites WHERE id = $1`, siteID).Scan(&domain); err != nil {
		return nil, fmt.Errorf("resolve domain for site %s: %w", siteID, err)
	}
	if strings.TrimSpace(domain) == "" {
		return nil, fmt.Errorf("site %s has no domain; a retraction cannot be scoped", siteID)
	}

	pageIDs := extractRetractionPageIDs(params, config)
	candidates, err := loadRetractionCandidates(ctx, params.DB, siteID, pageIDs)
	if err != nil {
		return nil, err
	}

	// Guard 3: the file paths an ACTIVE page on this site owns are off limits,
	// whatever any other row says. Built once, from the same derivation.
	activePaths, err := loadActivePageFilePaths(ctx, params.DB, siteID)
	if err != nil {
		return nil, err
	}

	var paths []string
	var eligible []retractionCandidate
	seen := map[string]bool{}
	filePathToURL := map[string]string{}
	for i := range candidates {
		c := &candidates[i]

		// Guard 4, restated at the row: only a page the platform says must not
		// be served. loadRetractionCandidates already filters, but an explicit
		// page-id list is caller input and gets checked again here.
		if c.Status == "active" {
			c.Refused = "page is active — retracting a live page is not what archiving means"
			continue
		}

		fp, ok := datahelpers.PageFilePathFromURL(c.URL)
		if !ok {
			// Guard 1. Declined, never guessed: a page whose url does not name
			// a file of its own has no artefact this action can identify, and
			// falling back to its NAME is what published 125's orphans.
			c.Refused = "url does not designate a file of its own (fragment, query, or off-origin) — no path can be derived safely"
			continue
		}
		c.FilePath = fp
		filePathToURL[fp] = c.URL

		if owner, clash := activePaths[fp]; clash {
			c.Refused = fmt.Sprintf("file path is also owned by ACTIVE page %q — retracting it would delete a live page's artefact", owner)
			continue
		}
		if seen[fp] {
			c.Refused = "duplicate of another candidate's file path in this batch"
			continue
		}
		seen[fp] = true
		paths = append(paths, fp)
		eligible = append(eligible, *c)
	}

	// GUARD 5 — the graph. A retraction that only removes the artefact trades
	// one broken state for another: 098's second surface is relojistas, where
	// the file was correctly deleted and every page then advertised the 404
	// from its own footer because the nav row was still active. See
	// retract_page_graph.go for why editorial and nav references are handled
	// differently, and why stranded targets are reported rather than repaired.
	graph, err := auditRetraction(ctx, params.DB, siteID, eligible, logger)
	if err != nil {
		return nil, err
	}

	// Editorial references BLOCK. A page something still links to may not be
	// retracted, so "a retraction created a dead link" is unreachable rather
	// than merely detectable. The referrers are named, because a refusal an
	// operator cannot act on is the defect bugs_open/033 and /071 describe.
	if len(graph.Editorial) > 0 {
		blocked := map[string]bool{}
		for _, r := range graph.Editorial {
			blocked[r.TargetURL] = true
		}
		kept := paths[:0]
		for _, fp := range paths {
			if u, ok := filePathToURL[fp]; ok && blocked[u] {
				continue
			}
			kept = append(kept, fp)
		}
		paths = kept
		for i := range candidates {
			if candidates[i].FilePath != "" && blocked[candidates[i].URL] {
				candidates[i].Refused = "still linked from live content — repair or remove those links first (see editorial_inbound)"
			}
		}
	}
	sort.Strings(paths)

	// DEBT 5 (bugs_open/098, found on the first real retraction, 2026-08-03):
	// this action's RETURN VALUE does not survive the await. The step returns
	// its findings AND awaits the adapter's reply, and when the reply lands
	// the await machinery replaces the step-name and output_field keys with it
	// wholesale (coordinator.go, applyResponseToState, default branch) — so
	// the one real retraction's record held only {paths, success, …} and every
	// refusal survived nowhere but pod logs, which this repo documents as
	// ephemeral across rollouts. A retraction that refused a page had refused
	// it silently — the opposite of what the guards above promise. Two sinks,
	// two readers — CORRECTED 2026-08-04: only the second is durable for an
	// awaited run (see retractionAuditKey's comment):
	//   - collected_data[retractionAuditKey]: survives NON-awaited paths only —
	//     the await park discards all CollectedData mutations;
	//   - agent_error_log rows (real runs only, below): refusals as warnings,
	//     the full audit as one RETRACTION_AUDIT info row — must outlive the
	//     orchestration row itself, which is retention-clocked (~24h once
	//     terminal), and that table is where the dashboards and the
	//     immune-system sweep already look.
	// `audit` and `result` share the same nested values on purpose; the audit
	// map is enriched in place below (nav_retired, request_id) and the
	// coordinator persists it when the step parks to await.
	audit := map[string]interface{}{
		"site_id":           siteID.String(),
		"domain":            domain,
		"considered":        len(candidates),
		"candidates":        candidates,
		"paths":             paths,
		"retracted":         len(paths),
		"editorial_inbound": graph.Editorial,
		"nav_inbound":       graph.Nav,
		"stranded_targets":  graph.Stranded,
		"audited_at":        time.Now().UTC().Format(time.RFC3339),
		"dispatched":        false,
	}
	attachRetractionAudit(params, audit, logger)

	result := map[string]interface{}{
		"site_id":           siteID.String(),
		"domain":            domain,
		"considered":        len(candidates),
		"candidates":        candidates,
		"paths":             paths,
		"retracted":         len(paths),
		"editorial_inbound": graph.Editorial,
		"nav_inbound":       graph.Nav,
		"stranded_targets":  graph.Stranded,
	}

	if dry, _ := config["dry_run"].(bool); dry {
		result["dry_run"] = true
		result["await_response"] = false
		audit["dry_run"] = true
		logger.Info("retract_page_deployment: dry run",
			zap.String("domain", domain), zap.Strings("paths", paths))
		return result, nil
	}

	// Refusals and stranded targets become durable rows on every REAL run —
	// deliberately BEFORE the empty-batch return below, because a run that
	// dispatches nothing since every candidate was refused is the loudest of
	// the silent cases, and BEFORE the dispatch, so a failed send cannot
	// unrecord what the audit found. The counts go into the audit so the
	// record says how many durable rows actually back it — recorded-but-lost
	// must not read as recorded.
	attempted, recorded := recordRetractionConditions(ctx, params, siteID, domain, candidates, graph, paths, logger)
	audit["conditions_recorded"] = recorded
	if attempted != recorded {
		audit["conditions_lost"] = attempted - recorded
	}

	if len(paths) == 0 {
		// Nothing to do is a real, common, correct outcome — say so rather than
		// dispatching an empty request the adapter would have to reject.
		result["await_response"] = false
		result["status"] = "nothing_to_retract"
		audit["status"] = "nothing_to_retract"
		recordRetractionAudit(ctx, params, siteID, domain, audit, logger)
		logger.Info("retract_page_deployment: nothing to retract",
			zap.String("domain", domain), zap.Int("considered", len(candidates)))
		return result, nil
	}

	// Nav rows are retired BEFORE the artefact is removed, not after. The
	// window between the two is the relojistas state — a live nav entry
	// pointing at a file that is already gone — and ordering it this way makes
	// the window the harmless one instead: a nav entry disappears slightly
	// before the page it pointed at does. It is also the durable half: the
	// delete is dispatched asynchronously and the DB write is not, so doing the
	// DB write on a response that may never arrive would leave the nav live
	// after a successful retraction.
	navRetired, err := deactivateNavItems(ctx, params.DB, graph.Nav, logger)
	if err != nil {
		return nil, err
	}
	result["nav_retired"] = navRetired
	audit["nav_retired"] = navRetired

	// The audit row is written AFTER the nav retire (so it carries nav_retired)
	// and BEFORE the dispatch (so a failed send cannot unrecord it). It cannot
	// carry request_id/dispatched — those happen later and this is
	// deliberately INSERT-only — but the row's orchestration_id column joins it
	// to the adapter reply and everything else about the run.
	recordRetractionAudit(ctx, params, siteID, domain, audit, logger)

	requestID, err := sendGitDeleteFileRequest(ctx, params, config, domain, paths,
		fmt.Sprintf("Retract %d retired page(s) from %s (bugs_open/098)", len(paths), domain), logger)
	if err != nil {
		return nil, err
	}
	result["request_id"] = requestID
	result["await_response"] = true
	audit["request_id"] = requestID
	audit["dispatched"] = true
	return result, nil
}

// retractionAuditKey is the collected_data key the audit is attached under.
// It is a TOP-LEVEL SIBLING of the step's own keys because
// applyResponseToState overwrites the step-name and output_field keys
// wholesale when the awaited reply lands. ~~A sibling key … is still there
// when the record is read afterwards~~ — **CORRECTED 2026-08-04 by the first
// live batch run: on an AWAITED step this key never reaches the DB at all.**
// persistAwaitingStateWithRetry parks the step onto a fresh DB load carrying
// only awaited-request bookkeeping, discarding every CollectedData mutation
// (RFC_012 addendum 2). The attach stays because it is correct for
// non-awaited paths (dry_run, nothing_to_retract) and for whenever RFC_012
// fixes the park; the DURABLE record of a real run is the RETRACTION_AUDIT
// agent_error_log row (recordRetractionAudit). The seed's output_field is
// 'retraction' (sql_for_agents/215); keep any caller's output_field distinct
// from this key.
const retractionAuditKey = "retraction_audit"

// attachRetractionAudit stores the audit under retractionAuditKey.
// params.CollectedData is the coordinator's live state map, passed by
// reference (coordinator.go, ActionParams construction), so a key written
// here is persisted when the step parks to await — writing side keys this way
// is an established action pattern (image_result, final_html, __spawn_input_data__).
func attachRetractionAudit(params ActionParams, audit map[string]interface{}, logger *zap.Logger) {
	if params.CollectedData == nil {
		// Never true under the coordinator — the step could not have resolved
		// its site without collected data — but a panic in bookkeeping must
		// not fail a retraction that is otherwise sound.
		logger.Warn("retract_page_deployment: no collected_data map — audit not attached; the step result is the only in-record copy and it will NOT survive the await")
		return
	}
	params.CollectedData[retractionAuditKey] = audit
}

// recordRetractionConditions persists every refusal, and the stranded-target
// report, as agent_error_log rows (severity 'warning') — the table the
// dashboards and the immune-system sweep already read, and the only sink here
// that outlives the orchestration record's retention. Real runs only; the
// caller skips it for dry runs. Row volume is bounded by the candidate set,
// which is at most the non-active stamped pages of ONE site.
//
// Best-effort, like every recorder against this table: a failure to record
// must never change a disposition already decided — the refusal stands either
// way, it is only the durable trace that is lost, and that loss is logged.
// Returns (rows attempted, rows written) so the caller can put the difference
// in the audit rather than letting a lost row read as a recorded one.
func recordRetractionConditions(ctx context.Context, params ActionParams, siteID uuid.UUID, domain string, candidates []retractionCandidate, graph *retractionGraph, paths []string, logger *zap.Logger) (attempted, recorded int) {
	// Editorial referrers, grouped by target, so a refusal row carries the
	// links an operator must repair — a refusal nobody can act on is the
	// defect bugs_open/033 and /071 describe.
	referrersByURL := map[string][]inboundRef{}
	for _, r := range graph.Editorial {
		referrersByURL[r.TargetURL] = append(referrersByURL[r.TargetURL], r)
	}

	for _, c := range candidates {
		if c.Refused == "" {
			continue
		}
		payload := map[string]interface{}{
			"page_id":     c.PageID,
			"url":         c.URL,
			"page_status": c.Status,
			"refused":     c.Refused,
		}
		if c.FilePath != "" {
			payload["file_path"] = c.FilePath
		}
		if refs := referrersByURL[c.URL]; len(refs) > 0 {
			payload["editorial_referrers"] = refs
		}
		attempted++
		if insertRetractionConditionRow(ctx, params, siteID, domain,
			"RETRACTION_REFUSED", "warning",
			fmt.Sprintf("retraction refused for page %q (%s): %s", c.PageName, c.URL, c.Refused),
			payload, logger) {
			recorded++
		}
	}

	if len(graph.Stranded) > 0 {
		attempted++
		if insertRetractionConditionRow(ctx, params, siteID, domain,
			"RETRACTION_STRANDED_TARGETS", "warning",
			fmt.Sprintf("%d page(s) lose their last inbound referrer when this retraction lands", len(graph.Stranded)),
			map[string]interface{}{
				"stranded":        graph.Stranded,
				"retracted_paths": paths,
			}, logger) {
			recorded++
		}
	}
	return attempted, recorded
}

// recordRetractionAudit persists the FULL audit as one agent_error_log row
// (`RETRACTION_AUDIT`, severity 'info') on every real run — clean ones
// included, which is the point: after the first live batch run proved the
// collected_data copy dies at the await park (debt 5b, RFC_012 addendum 2),
// this row is the ONLY durable record of what a retraction considered,
// refused and stranded. Dry runs are the caller's concern and never reach
// here. Best-effort like every recorder in this file; the sibling-key attach
// stays because it is correct for non-awaited reads and costs nothing.
func recordRetractionAudit(ctx context.Context, params ActionParams, siteID uuid.UUID, domain string, audit map[string]interface{}, logger *zap.Logger) bool {
	considered, _ := audit["considered"].(int)
	retracted := 0
	if p, ok := audit["paths"].([]string); ok {
		retracted = len(p)
	}
	ok := insertRetractionConditionRow(ctx, params, siteID, domain,
		"RETRACTION_AUDIT", "info",
		fmt.Sprintf("retraction audit for %s: %d considered, %d dispatched for deletion", domain, considered, retracted),
		audit, logger)
	audit["audit_row_recorded"] = ok
	return ok
}

// insertRetractionConditionRow writes one agent_error_log row, reporting
// whether the write landed.
//
// Was this package's exemplar hand-copied INSERT (the cycle that forced the
// copies is retired — RFC_012 option B moved the ONE writer into the
// agenterrors leaf package, and LogActionError is this package's door onto
// it). Kept as a named function because its call sites read better and its
// doc carries the doctrine:
//
// THIS TABLE IS THE ONLY SINK THAT SURVIVES AN AWAITED STEP (debt 5b,
// 2026-08-04): the collected_data sibling key this file also writes was
// refuted by the first live batch run — persistAwaitingStateWithRetry parks
// the step onto a fresh DB load and discards every CollectedData mutation
// (RFC_012 addendum 2). A direct row, written before dispatch, is durable
// regardless.
func insertRetractionConditionRow(ctx context.Context, params ActionParams, siteID uuid.UUID, domain, code, severity, message string, contextPayload map[string]interface{}, logger *zap.Logger) bool {
	return LogActionError(ctx, params, siteID.String(), domain, "retract_page_deployment",
		code, severity, message, contextPayload, logger)
}

// resolveRetractionSiteID takes the site from config, from a configured field,
// or from the usual collected-data locations. A retraction with no site is
// refused rather than defaulted: an unscoped deletion in a repo that holds every
// site is the one mistake this action must not be able to make.
func resolveRetractionSiteID(params ActionParams, config map[string]interface{}) (uuid.UUID, error) {
	if raw, ok := config["site_id"].(string); ok && raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return uuid.Nil, fmt.Errorf("config site_id %q is not a uuid: %w", raw, err)
		}
		return id, nil
	}
	fields := []string{"site_record.id", "input_data.site_id", "site_id"}
	if f, ok := config["site_id_field"].(string); ok && f != "" {
		fields = append([]string{f}, fields...)
	}
	for _, f := range fields {
		if v := datahelpers.ExtractNestedFieldString(params.CollectedData, f); v != "" {
			id, err := uuid.Parse(v)
			if err != nil {
				continue
			}
			return id, nil
		}
	}
	return uuid.Nil, fmt.Errorf("no site_id resolved — a retraction must be scoped to one site")
}

// extractRetractionPageIDs reads an optional explicit page list. Empty means
// "every non-active page on the site that is still stamped as deployed".
func extractRetractionPageIDs(params ActionParams, config map[string]interface{}) []string {
	field, _ := config["page_ids_field"].(string)
	if field == "" {
		return nil
	}
	raw := datahelpers.ExtractNestedField(params.CollectedData, field)
	var out []string
	switch v := raw.(type) {
	case []string:
		out = append(out, v...)
	case []interface{}:
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				out = append(out, s)
			}
		}
	case string:
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

func loadRetractionCandidates(ctx context.Context, db *sql.DB, siteID uuid.UUID, pageIDs []string) ([]retractionCandidate, error) {
	// The default population is 098's own measurement query: not active, and
	// still carrying a deploy stamp. `deployed_at IS NOT NULL` is used here as
	// the bug documents it — "a deploy happened once", i.e. an artefact may
	// exist — NOT as proof it is fetchable. Whether it is actually still there
	// is settled by the adapter's per-path existence check, which reports an
	// absent path as a success rather than a failure.
	query := `
		SELECT id::text, COALESCE(name,''), COALESCE(url,''), COALESCE(status,'')
		  FROM pages
		 WHERE site_id = $1
		   AND COALESCE(status,'') <> 'active'
		   AND deployed_at IS NOT NULL`
	args := []interface{}{siteID}
	if len(pageIDs) > 0 {
		// An explicit list still may not reach outside the site or pick up an
		// active page — it narrows the population, it never widens it.
		query = `
			SELECT id::text, COALESCE(name,''), COALESCE(url,''), COALESCE(status,'')
			  FROM pages
			 WHERE site_id = $1
			   AND id::text = ANY($2)`
		args = append(args, pageIDs)
	}
	query += ` ORDER BY url`

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("load retraction candidates: %w", err)
	}
	defer rows.Close()

	var out []retractionCandidate
	for rows.Next() {
		var c retractionCandidate
		if err := rows.Scan(&c.PageID, &c.PageName, &c.URL, &c.Status); err != nil {
			return nil, fmt.Errorf("scan retraction candidate: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// loadActivePageFilePaths maps every file path an ACTIVE page on this site
// derives, to that page's name. The derivation is the shared one, so this
// answers the question that matters — "would deleting this file remove a live
// page's artefact?" — rather than the weaker url-equality question.
func loadActivePageFilePaths(ctx context.Context, db *sql.DB, siteID uuid.UUID) (map[string]string, error) {
	// The lifecycle arm is the shared helper; for the `=` direction the
	// COALESCE form it replaces was NULL-identical (both reject a NULL
	// status). Only the `<>` COMPLEMENT in loadRetractionCandidates differs
	// on NULL and deliberately keeps its COALESCE spelling.
	rows, err := db.QueryContext(ctx, `
		SELECT COALESCE(name,''), COALESCE(url,'')
		  FROM pages
		 WHERE site_id = $1 AND `+datahelpers.PageWantedLivePredicateFor("")+``, siteID)
	if err != nil {
		return nil, fmt.Errorf("load active page paths: %w", err)
	}
	defer rows.Close()

	out := map[string]string{}
	for rows.Next() {
		var name, url string
		if err := rows.Scan(&name, &url); err != nil {
			return nil, fmt.Errorf("scan active page: %w", err)
		}
		if fp, ok := datahelpers.PageFilePathFromURL(url); ok {
			out[fp] = name
		}
	}
	return out, rows.Err()
}

// sendGitDeleteFileRequest dispatches the retraction to the git-adapter. The
// envelope mirrors GitCommitAction's — same topic, same headers, same
// deploy-target resolution — because a retraction must land in exactly the repo
// the publish landed in.
//
// Branch is NOT set. gqls/sites has no `main` (it carries `master`, its default,
// and `750start`) while sites.github_branch says 'main' for most rows, and the
// B2 deploy workflow triggers on `master`. Deploys work because the commit path
// leaves the branch unset and CommitToRepo falls back to the repo default;
// a retraction that helpfully passed the column would target a branch that does
// not exist.
func sendGitDeleteFileRequest(ctx context.Context, params ActionParams, config map[string]interface{}, domain string, paths []string, commitMessage string, logger *zap.Logger) (string, error) {
	if params.Producer == nil {
		return "", fmt.Errorf("kafka producer not available")
	}

	correlationID := getExecutionContextValue(params, "correlation_id", params.ExecutionContext.CorrelationID)
	orchestrationID := getExecutionContextValue(params, "orchestration_id", params.ExecutionContext.OrchestrationID)
	clientID := getExecutionContextValue(params, "client_id", params.ExecutionContext.ClientID)
	responsesTopic := getExecutionContextValue(params, "responses_topic", params.ExecutionContext.ResponsesTopic)
	if responsesTopic == "" {
		if alt, ok := params.CollectedData["__my_responses_topic__"].(string); ok && alt != "" {
			responsesTopic = alt
		}
	}
	if responsesTopic == "" {
		if parent, ok := params.CollectedData["__parent_responses_topic__"].(string); ok && parent != "" {
			responsesTopic = parent
		}
	}

	repoName := resolveGitRepoNameDB(ctx, params.DB, config, params.CollectedData, domain, logger)
	newRequestID := uuid.New().String()

	adapterRequest := map[string]interface{}{
		"headers": map[string]interface{}{
			"correlation_id":         correlationID,
			"orchestration_id":       orchestrationID,
			"client_id":              clientID,
			"step_name":              params.ExecutionContext.StepName,
			"step_id":                params.ExecutionContext.StepID,
			"request_id":             newRequestID,
			"message_type":           "request",
			"sender_agent_type":      params.ExecutionContext.Sender.AgentType,
			"sender_agent_id":        orchestrationID,
			"sender_pod_name":        params.ExecutionContext.Sender.PodName,
			"responses_topic":        responsesTopic,
			"parent_responses_topic": responsesTopic,
			"timestamp":              time.Now().UTC().Format(time.RFC3339),
			"action":                 "delete_file",
		},
		"body": map[string]interface{}{
			"action": "delete_file",
			"data": map[string]interface{}{
				"repo_name":      repoName,
				"domain":         domain,
				"paths":          paths,
				"commit_message": commitMessage,
			},
			"reply_to_topic":         responsesTopic,
			"parent_responses_topic": responsesTopic,
			"metadata": map[string]interface{}{
				"requesting_agent_id":   orchestrationID,
				"requesting_agent_type": params.ExecutionContext.Sender.AgentType,
				"requesting_step":       params.ExecutionContext.StepName,
				"client_id":             clientID,
				"domain":                domain,
				"paths_count":           len(paths),
			},
			"request_context": map[string]interface{}{
				"correlation_id":   correlationID,
				"orchestration_id": orchestrationID,
				"request_id":       newRequestID,
			},
		},
	}

	messageBytes, err := json.Marshal(adapterRequest)
	if err != nil {
		return "", fmt.Errorf("marshal delete_file request: %w", err)
	}
	rawHeaders := adapterRequest["headers"].(map[string]interface{})
	headers := make(map[string]string, len(rawHeaders))
	for k, v := range rawHeaders {
		headers[k] = fmt.Sprintf("%v", v)
	}

	if err := params.Producer.ProduceWithValidation(ctx, "system.adapter.git.requests",
		headers, []byte(correlationID), messageBytes); err != nil {
		return "", fmt.Errorf("send delete_file to git-adapter: %w", err)
	}

	logger.Info("retract_page_deployment: delete_file sent, awaiting response",
		zap.String("repo_name", repoName),
		zap.String("domain", domain),
		zap.Strings("paths", paths),
		zap.String("request_id", newRequestID))

	return newRequestID, nil
}
