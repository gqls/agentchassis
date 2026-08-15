// FILE: platform/orchestration/actions/discovery_checks/check_unverified_claims.go
//
// Post-deploy claims audit (SPEC_claims_verification V1b). Runs the same two
// deterministic scans as the build-time gate — banned claims and unregistered
// numbers — over DEPLOYED page_components and site_components HTML.
//
// Why this exists when the gate already checks at build time: the deployed
// surface catches drift, hand-edits, and pages that PREDATE the gate. The
// motivating case: "eight departments" was audited out of the leopardess site,
// and was found weeks later alive mid-paragraph on an orphan page
// (for-engineering-teams) — nothing in the platform knew the claim was on the
// banned list. This check is what would have caught it sleeping there. (It
// did: the check's first live run on 2026-07-16 found that same page plus a
// guide written the day before both carrying "70+ agents across eight
// functional departments" — the writer had leaked the banned claim again
// after the 2026-07-10 fabrication sweep.)
//
// Scan engine: the SHARED datahelpers claims implementation (ExtractAssertionText
// / ScanBannedClaims / ScanUnregisteredNumbers) — the same functions the deploy
// gate (validate_page_content check 8) uses, so gate and audit agree by one
// literal implementation on what counts as an asserted claim. Assertion TEXT
// NODES only, never raw HTML/attributes (landmine: placeholder="jane@…" is an
// example, not a claim).
//
// Opt-in: the two HTML scans need a register to compare against, so sites
// without a current site_specs 'evidence_base' aspect are skipped for those.
// banned_claims starts empty elsewhere — never block or flag fleet-wide on a
// layer only one site has data for.
//
// SECOND SURFACE, added 2026-07-26 for bugs_open/093: stored content_data is
// now audited alongside rendered_html, because the stat audit had exactly one
// call site (the build gate) and the re-render path republishes stored figures
// with no check at all. The stat scans have their own scope rules — the unit
// lint runs fleet-wide, the register comparison keys on the site_specs row
// EXISTING rather than on ParseEvidenceBase returning non-nil. Both are argued
// in check_unverified_claims_stats.go; read that header before changing the
// gating here.
//
// Routing: findings terminate at HUMAN review. Truth decisions are human —
// auditors raise work items, they never rewrite content (content-governance
// rule). One work item per page (item_key 'claims:<page_id>', mirroring
// page_rerender:<page> dedup), status needs_human_review with no handler
// agent — the same HITL-terminal shape as check_required_fields_missing and
// check_section_source_drift. Site-level components (header/footer chrome)
// group under a single 'claims:site_components' item.
//
// Locked components (locked_at IS NOT NULL) are skipped by precedent
// (check_placeholder_contact): a human explicitly pinned that content.
//
// Registration: automatic via init() -> Register(&UnverifiedClaimsCheck{}).
// Enable by adding "unverified_claims" to a discovery agent's checks array.

package discovery_checks

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func init() { Register(&UnverifiedClaimsCheck{}) }

type UnverifiedClaimsCheck struct{}

func (c *UnverifiedClaimsCheck) Name() string { return "unverified_claims" }

func (c *UnverifiedClaimsCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	eb, rowExists, err := LoadEvidenceBase(dctx.Ctx, dctx.DB, dctx.SiteID)
	if err != nil {
		return nil, err
	}

	// NO early return on eb == nil any more (bugs_open/093). The HTML scans
	// still need a register and are gated on eb below, but the stat-unit lint
	// compares a component against itself and runs fleet-wide — the same scope
	// it has in the build gate. Returning here would have kept the sites with
	// no register unaudited on BOTH paths, which is the gap this bug is about.
	scans, siteFindings, err := ScanDeployedClaims(
		dctx.Ctx, dctx.DB, dctx.SiteID, "", eb, rowExists, dctx.Logger)
	if err != nil {
		return nil, err
	}

	// ScanDeployedClaims reports every page it examined, including the clean ones,
	// so the revalidator can tell "examined and clean" from "examined nothing".
	// Only the pages that actually carry findings become work items — filtering
	// here keeps this check's emitted output exactly what it was before the scan
	// was shared.
	pageFindings := make([]*ClaimsPageScan, 0, len(scans))
	total := 0
	for _, pf := range scans {
		if len(pf.Findings) == 0 {
			continue
		}
		pageFindings = append(pageFindings, pf)
		total += len(pf.Findings)
	}
	if len(pageFindings) == 0 && len(siteFindings) == 0 {
		return &CheckResult{}, nil
	}
	total += len(siteFindings)

	result := &CheckResult{
		Findings: []map[string]interface{}{{
			"check":         "unverified_claims",
			"count":         total,
			"pages":         len(pageFindings),
			"site_surfaces": len(siteFindings) > 0,
		}},
	}

	// eb is nil on a site with no register, or with a writer_block-only row.
	auditDoc := ""
	if eb != nil {
		auditDoc = eb.AuditDoc
	}

	// One work item per affected page.
	for _, pf := range pageFindings {
		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":     "unverified_claims",
			"page_id":   pf.PageID,
			"page_name": pf.PageName,
			"findings":  pf.Findings,
			"audit_doc": auditDoc,
			"fix": "Human review required. Banned claims are audited-out fabrications for this " +
				"site (see reasons and the audit doc); unregistered numbers and unregistered stat " +
				"fields need either an evidence_base fact row (if true and provable) or removal " +
				"from the copy; an impossible stat unit (a magnitude marker given a dimension, e.g. " +
				"\"2,400+%\") is an unedited component-schema fallback and should be cleared in " +
				"page_components.content_data, NOT only in the rendered HTML — a re-render would " +
				"otherwise reprint it. Truth decisions are human — do not auto-rewrite.",
		})

		var pageIDPtr *uuid.UUID
		if parsed, perr := uuid.Parse(pf.PageID); perr == nil {
			pageIDPtr = &parsed
		}

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			PageID:       pageIDPtr,
			Source:       "discovery",
			Pipeline:     "content",
			ItemType:     "claims_unverified",
			Severity:     claimsSeverity(pf.Findings),
			Summary:      claimsSummary(pf.PageName, pf.Findings),
			SpecJSON:     string(specJSON),
			Priority:     claimsPriority(pf.Findings),
			HandlerAgent: "", // HITL-terminal: no automated handler, ever
			Status:       "needs_human_review",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("claims:%s", pf.PageID),
			BatchID:      dctx.BatchID,
		})
	}

	// Site-level components (header/footer/head chrome) — one grouped item.
	if len(siteFindings) > 0 {
		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":     "unverified_claims",
			"surface":   "site_components",
			"findings":  siteFindings,
			"audit_doc": auditDoc,
			"fix":       "Human review required — claims in site-level chrome.",
		})
		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Pipeline:     "content",
			ItemType:     "claims_unverified",
			Severity:     claimsSeverity(siteFindings),
			Summary:      claimsSummary("site components", siteFindings),
			SpecJSON:     string(specJSON),
			Priority:     claimsPriority(siteFindings),
			HandlerAgent: "",
			Status:       "needs_human_review",
			CreatedBy:    dctx.AgentType,
			ItemKey:      "claims:site_components",
			BatchID:      dctx.BatchID,
		})
	}

	dctx.Logger.Warn("unverified_claims: found claims without evidence support",
		zap.Int("finding_count", total),
		zap.Int("pages_affected", len(pageFindings)),
		zap.String("site_id", dctx.SiteID.String()))

	return result, nil
}

// unverifiedClaimFinding is one scan finding located on a deployed surface.
//
// Severity is carried PER FINDING rather than re-derived from Check by the
// grader. The stat scans (bugs_open/093) grade on facts the grader cannot see
// from the check name alone — whether the site registered any countables at
// all — and a finding that could not actually be verified must not be able to
// present itself as one that was.
type unverifiedClaimFinding struct {
	// banned_claim | unregistered_number | unregistered_stat |
	// unregistered_stat_unpaired | stat_unit_impossible | stat_unit_mismatch
	Check       string `json:"check"`
	SlotName    string `json:"slot_name"`
	Matched     string `json:"matched"`
	Pattern     string `json:"pattern,omitempty"`
	Reason      string `json:"reason"`
	Snippet     string `json:"snippet"`
	Occurrences int    `json:"occurrences"`
	Severity    string `json:"severity"`
	Source      string `json:"source"` // rendered_html | content_data
}

// ClaimFinding is the exported name for a single claims finding, so the
// review-queue revalidator in `actions` can name the type it iterates. Alias
// rather than rename: the unexported spelling is used at ~20 sites including
// this package's tests, and churning them buys nothing.
type ClaimFinding = unverifiedClaimFinding

// ClaimsPageScan is one page's claims audit. It carries the findings AND the two
// counts that say how the scan reached them, because the review-queue
// revalidator has to distinguish three states a bare `len(Findings) == 0`
// collapses into one:
//
//   - examined some components, found no claims     -> the copy was fixed
//   - examined NOTHING (page gone, unbuilt, empty)  -> we cannot answer
//   - the only components are human-LOCKED          -> we cannot answer
//
// Only the first licenses closing a live work item. The other two look
// identical from the findings list alone, which is exactly the no-op case that
// reads like a success. Same shape and same reasoning as VoicePageScan.
type ClaimsPageScan struct {
	PageID   string
	PageName string
	Findings []unverifiedClaimFinding
	// ComponentsExamined counts components actually passed to the scans.
	ComponentsExamined int
	// ComponentsSkippedLocked counts human-pinned components, which the emit
	// side has always skipped and this scan still does — counted rather than
	// silently dropped so a page whose only claims sit in a locked component
	// cannot be read as "the copy was fixed".
	ComponentsSkippedLocked int
	// NewestComponentUpdate is the latest page_components.updated_at across the
	// components actually EXAMINED (locked rows are excluded, since their content
	// was not read and cannot be what changed).
	//
	// It exists for one reason: the register this scan measures against is DATA,
	// so a page can stop tripping the check because somebody edited the REGISTER
	// rather than the PAGE. Findings alone cannot tell those apart. Comparing this
	// against the work item's created_at can. Zero value means no examined
	// component carried a timestamp, which the ladder must treat as "cannot show
	// the copy moved" — never as "it did".
	NewestComponentUpdate time.Time
	// ExaminedTextBySlot is the copy this scan ACTUALLY READ, keyed by slot_name,
	// for EXAMINED components only — locked rows are absent by the same reasoning
	// as NewestComponentUpdate: their content was not read, so nothing may be
	// concluded from its presence or absence.
	//
	// It exists for the CLAIM-GRANULAR gate in the review-queue revalidator. The
	// timestamp gate above answers "did the page move?"; it cannot answer "did the
	// thing we flagged go away?", and an edit to an unrelated slot satisfies it.
	// Holding the examined text per slot lets the revalidator ask whether the exact
	// text a finding cited is still sitting in the component it was cited from —
	// without a second query and without a second predicate.
	//
	// Keyed by slot rather than concatenated page-wide DELIBERATELY, and this is
	// measured rather than assumed: page-wide, a bare `unregistered_number` token
	// such as "5" or "26" matches somewhere in almost any page's markup. On the 8
	// items closed to 2026-08-11 a page-wide search reported 7 of 18 flagged texts
	// still present; scoped to the slot, 2. Every one of the 7 was 1-4 characters
	// long. Slot scoping is what makes the token test discriminate rather than
	// merely fire.
	//
	// Components sharing a slot_name are concatenated, so a finding is judged
	// against everything read under its own slot.
	ExaminedTextBySlot map[string]string
}

// LoadEvidenceBase reads the site's current evidence_base spec.
//
// Exported for the review-queue revalidator (mirrors LoadVoiceGate): whether a
// site is opted in has to be answerable from the `actions` package, because a
// site with no register cannot have its parked claims items re-judged on the
// half of the audit that needs one.
//
// The second return is whether the ROW EXISTS, which is NOT the same question
// as whether eb is non-nil, and the difference is load-bearing (bugs_open/093,
// and the landmine that made the build gate's check 9 key on the row too):
// ParseEvidenceBase returns nil for a row with no facts[] and no
// banned_claims[], so a writer_block-only row reads as "not opted in" while the
// writer block itself goes on working. Three sites sat like that for two days
// with both claims checkers silently off. The stat scans use rowExists.
//
// UPDATED 2026-07-28 (bugs_open/104): the HTML scans no longer split the same
// way. The NUMBER scan still needs eb — it has nothing to compare a figure
// against without one — but the BANNED-CLAIM scan runs on every site, register
// or not, because the fleet-wide set is not per-site data. So eb == nil now
// means "no site-specific patterns and no facts", not "skip the HTML scans".
func LoadEvidenceBase(ctx context.Context, db *sql.DB, siteID uuid.UUID) (*datahelpers.EvidenceBase, bool, error) {
	var data []byte
	err := db.QueryRowContext(ctx, `
		SELECT data FROM site_specs
		WHERE site_id = $1 AND aspect = 'evidence_base' AND is_current = true
	`, siteID).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("unverified_claims evidence_base query failed: %w", err)
	}
	eb, err := datahelpers.ParseEvidenceBase(data)
	if err != nil {
		// A malformed evidence base is a real defect on an opted-in site —
		// surface it as a check error (logged, not fatal to the run).
		return nil, true, fmt.Errorf("unverified_claims evidence_base parse failed: %w", err)
	}
	return eb, true, nil
}

// ScanDeployedClaims runs the scans over deployed page and site components,
// skipping locked components (human-pinned content is never flagged) and
// returning one ClaimsPageScan per page examined, INCLUDING pages with no
// findings.
//
// It is shared by the emit side (UnverifiedClaimsCheck.Run, pageIDFilter "")
// and the review-queue revalidator (one page). That sharing is the point: the
// two ends of a claims_unverified item's life must not be able to answer "does
// this page still assert something the register does not support?" differently.
// Same precedent as ScanVoiceTells and needs_page.
//
// ⚠ ONE PAGE-STATUS EXCLUSION, added 2026-08-15 on the owner's instruction, and
// it is a CONJUNCTION: archived AND never deployed. Until then this check had no
// page filter at all, deliberately — which meant a rejected draft no visitor can
// reach was re-scanned on every daily sweep, filed findings, and had no way to
// close (webdesign.uk/index-rejected-v1-20260806, 19 banned-claim findings,
// review-queue item a355d78b). It remains narrower than ScanVoiceTells, which
// restricts to p.status IN ('active','deployed').
//
// BOTH HALVES ARE LOAD-BEARING, and that is measured rather than argued.
// `p.status = 'archived'` ALONE IS UNSAFE: of the 11 archived pages carrying
// scannable components on 2026-08-15, FIVE were serving HTTP 200 —
// fundamentallyai.com/blog/ai-readiness-checker-guide.html (25,861 B),
// fundamentallyai.com/tools/llm-cost-calculator/index.html (35,331 B),
// leopardessconsulting.co.uk/our-approach.html (28,948 B),
// robot-hands.com/gripper-catalog.html (30,997 B) and robot-hands.com/news.html
// (48,047 B) — each measured against a fabricated-URL control on its own domain
// that returned 404 at a different byte count. Excluding on status alone would
// silently stop claim-checking five live, publicly-served pages.
//
// ⚠ THE ESTATE'S USUAL LIVENESS PREDICATE IS THEREFORE WRONG FOR THIS QUERY.
// `p.status = 'active' AND p.build_status = 'deployed'` (load_site_pages_action.go:80
// and four sibling call sites) is the right test for "should we RE-PUBLISH this
// page". It is the wrong test for "is there copy a visitor can read", which is
// what an audit asks: archiving a page sets a database column, it does not
// unpublish the artefact already sitting in the deploy repo.
//
// The conjunction excludes exactly 1 page / 7 components fleet-wide, and the URL
// it drops stays covered — an active, deployed row serves the same /index.html.
//
// The revalidator must judge by exactly the predicate the emit side uses or the
// two ends disagree, which is why this scan is shared rather than reimplemented.
// An excluded page returns no row, which unverifiedClaimsVerdict reads as
// armPageAbsent — a REFUSAL, not a closure. So this stops the CLASS; it does not
// dispose of an item already filed against such a page.
//
// TWO SURFACES PER COMPONENT, and each scan is gated on the one it reads:
// the HTML scans on rendered_html, the stat scans on content_data
// (bugs_open/093). The row predicate is therefore an OR — a component with
// stored content_data but no rendered_html has never been published, but it is
// exactly what the next re-render WILL publish, unchecked.
//
// LOCKED COMPONENTS MOVED FROM SQL TO GO (2026-08-09). The page query used to
// carry `AND pc.locked_at IS NULL`, which dropped pinned rows before anything
// could count them — so a page whose only claims sit in a locked component was
// indistinguishable from a page that was cleaned up. The rows now come back and
// are skipped in the loop, which emits exactly what it emitted before and
// additionally reports what it did not read.
func ScanDeployedClaims(
	ctx context.Context,
	db *sql.DB,
	siteID uuid.UUID,
	pageIDFilter string,
	eb *datahelpers.EvidenceBase,
	rowExists bool,
	logger *zap.Logger,
) ([]*ClaimsPageScan, []unverifiedClaimFinding, error) {
	// page_components — body sections.
	//
	// p.page_type joins the select for bugs_open/102: it gates the prose number
	// heuristic, which cannot tell a guide's worked example from a sales claim.
	// The column was always in the row this query already reads — the layer
	// simply never asked for it.
	pageQuery := `
		SELECT pc.page_id::text, p.name, COALESCE(pc.slot_name, ''),
		       COALESCE(pc.rendered_html, ''), COALESCE(cc.name, ''),
		       COALESCE(pc.content_data::text, ''), COALESCE(p.page_type, ''),
		       pc.locked_at, pc.updated_at
		FROM page_components pc
		JOIN pages p ON p.id = pc.page_id
		LEFT JOIN content_components cc ON cc.id = pc.component_id
		WHERE p.site_id = $1
		  AND NOT (p.status = 'archived' AND p.deployed_at IS NULL)
		  AND ( (pc.rendered_html IS NOT NULL AND pc.rendered_html <> '')
		     OR (pc.content_data IS NOT NULL AND pc.content_data::text <> '{}') )`
	pageArgs := []interface{}{siteID}
	if strings.TrimSpace(pageIDFilter) != "" {
		pageQuery += ` AND pc.page_id = $2::uuid`
		pageArgs = append(pageArgs, strings.TrimSpace(pageIDFilter))
	}
	pageQuery += ` ORDER BY p.name, pc.position`

	pageRows, err := db.QueryContext(ctx, pageQuery, pageArgs...)
	if err != nil {
		return nil, nil, fmt.Errorf("unverified_claims page query failed: %w", err)
	}
	defer pageRows.Close()

	byPage := map[string]*ClaimsPageScan{}
	var pageOrder []*ClaimsPageScan

	for pageRows.Next() {
		var pageID, pageName, slotName, html, component, contentJSON, pageType string
		var lockedAt, updatedAt sql.NullTime
		if err := pageRows.Scan(&pageID, &pageName, &slotName, &html, &component,
			&contentJSON, &pageType, &lockedAt, &updatedAt); err != nil {
			logger.Warn("unverified_claims: page scan error", zap.Error(err))
			continue
		}
		pf, ok := byPage[pageID]
		if !ok {
			pf = &ClaimsPageScan{PageID: pageID, PageName: pageName}
			byPage[pageID] = pf
			pageOrder = append(pageOrder, pf)
		}
		if lockedAt.Valid {
			pf.ComponentsSkippedLocked++
			continue
		}
		pf.ComponentsExamined++
		// Tracked only for EXAMINED rows: a locked component's timestamp says
		// something changed in content this scan deliberately did not read, which
		// is not evidence about the copy the finding was raised against.
		if updatedAt.Valid && updatedAt.Time.After(pf.NewestComponentUpdate) {
			pf.NewestComponentUpdate = updatedAt.Time
		}
		// Same rule, same reason: EXAMINED rows only. Recorded before the scans run
		// so the text held is exactly the text the scans were given — if these two
		// ever diverge, the revalidator is judging copy the emit side never read.
		if pf.ExaminedTextBySlot == nil {
			pf.ExaminedTextBySlot = map[string]string{}
		}
		pf.ExaminedTextBySlot[slotName] += html + contentJSON

		findings := scanComponentClaims(html, slotName, eb,
			datahelpers.ClaimSurface{PageType: pageType})
		findings = append(findings, scanStoredStatClaims(
			component, slotName, decodeContentData(contentJSON, logger), eb, rowExists)...)
		pf.Findings = append(pf.Findings, findings...)
	}
	if err := pageRows.Err(); err != nil {
		return nil, nil, fmt.Errorf("unverified_claims page iteration failed: %w", err)
	}

	// site_components — header/footer/head chrome. Audited on both surfaces
	// too: site chrome is re-rendered by the same content_data-only route, and
	// leaving it on one surface would reproduce, one level down, the exact
	// one-guarded-call-site shape bugs_open/093 is about. It carries zero stat
	// fields fleet-wide today (measured 2026-07-26), so this costs nothing now
	// and forecloses the gap if a stat component is ever slotted into chrome.
	siteRows, err := db.QueryContext(ctx, `
		SELECT COALESCE(sc.slot_name, ''), COALESCE(sc.rendered_html, ''),
		       COALESCE(cc.name, ''), COALESCE(sc.content_data::text, '')
		FROM site_components sc
		LEFT JOIN content_components cc ON cc.id = sc.component_id
		WHERE sc.site_id = $1
		  AND sc.locked_at IS NULL
		  AND ( (sc.rendered_html IS NOT NULL AND sc.rendered_html <> '')
		     OR (sc.content_data IS NOT NULL AND sc.content_data::text <> '{}') )
	`, siteID)
	if err != nil {
		return nil, nil, fmt.Errorf("unverified_claims site query failed: %w", err)
	}
	defer siteRows.Close()

	var siteFindings []unverifiedClaimFinding
	for siteRows.Next() {
		var slotName, html, component, contentJSON string
		if err := siteRows.Scan(&slotName, &html, &component, &contentJSON); err != nil {
			logger.Warn("unverified_claims: site scan error", zap.Error(err))
			continue
		}
		// Site chrome belongs to no page, so its surface is the zero value —
		// UNKNOWN, which scans (bugs_open/102). A footer asserting "170 agents"
		// is a business claim, and there is no page type that could say
		// otherwise.
		siteFindings = append(siteFindings, scanComponentClaims(html, slotName, eb,
			datahelpers.ClaimSurface{})...)
		siteFindings = append(siteFindings, scanStoredStatClaims(
			component, slotName, decodeContentData(contentJSON, logger), eb, rowExists)...)
	}
	if err := siteRows.Err(); err != nil {
		return nil, nil, fmt.Errorf("unverified_claims site iteration failed: %w", err)
	}

	return pageOrder, siteFindings, nil
}

// decodeContentData turns the JSONB text into a map. A row that will not decode
// is skipped rather than failing the check — one malformed component must not
// take the whole site's audit down with it.
func decodeContentData(contentJSON string, logger *zap.Logger) map[string]interface{} {
	if contentJSON == "" || contentJSON == "{}" || contentJSON == "null" {
		return nil
	}
	var cd map[string]interface{}
	if err := json.Unmarshal([]byte(contentJSON), &cd); err != nil {
		logger.Debug("unverified_claims: content_data did not decode as an object — skipping its stat audit",
			zap.Error(err))
		return nil
	}
	return cd
}

// scanComponentClaims runs both shared HTML scans over one component's markup.
//
// The banned-claim half is NOT gated on eb (bugs_open/104): the fleet-wide set
// applies to every site, armed or not, and ScanAllBannedClaims is nil-safe. The
// NUMBER half still is — with no register there is nothing to compare prose
// against, and flagging every number on a site that never opted in would be
// noise, not a finding. (The stat scans are different — see
// scanStoredStatClaims.)
//
// This deliberately mirrors the build gate (validate_page_content check 8) rather
// than gating differently: the two surfaces are documented to agree by one
// literal implementation, and a fix applied to one branch of a two-branch
// enforcement path is how 016b §9's "reads as done, other branch keeps the bug"
// entry got written. Note this half of the layer is currently unreachable —
// improvement-sweep has been disabled since 2026-05-02 (bugs_open/083) — so the
// change is inert here today and correct when that is re-enabled.
//
// surface gates the NUMBER scan only (bugs_open/102). The banned scan below is
// deliberately not gated by it: this check's own motivating case was a banned
// claim living on a guide page (see the file header), and a pattern a human
// audited out is a falsehood on any page type.
func scanComponentClaims(html, slotName string, eb *datahelpers.EvidenceBase,
	surface datahelpers.ClaimSurface) []unverifiedClaimFinding {
	if html == "" {
		return nil
	}
	blocks := datahelpers.ExtractAssertionText(html)
	var out []unverifiedClaimFinding
	for _, f := range datahelpers.ScanAllBannedClaims(blocks, eb) {
		out = append(out, unverifiedClaimFinding{
			Check: f.Check, SlotName: slotName, Matched: f.Matched,
			Pattern: f.Pattern, Reason: f.Reason, Snippet: f.Snippet,
			Occurrences: f.Occurrences,
			// A banned claim is a known falsehood, live on the site.
			Severity: "high", Source: "rendered_html",
		})
	}
	for _, f := range eb.ScanUnregisteredNumbers(blocks, surface) {
		out = append(out, unverifiedClaimFinding{
			Check: f.Check, SlotName: slotName, Matched: f.Matched,
			Reason: f.Reason, Snippet: f.Snippet, Occurrences: f.Occurrences,
			// Extraction has false positives, so never above medium.
			Severity: "medium", Source: "rendered_html",
		})
	}
	return out
}

// claimsSeverity is the worst severity any finding on the surface carries.
//
// Previously derived from Check alone (banned_claim -> high, else medium); it
// now reads the per-finding grade, because the stat scans grade on whether the
// site registered any facts at all and that is invisible from the check name.
// Behaviour for the two HTML checks is unchanged by construction.
func claimsSeverity(findings []unverifiedClaimFinding) string {
	worst := "low"
	for _, f := range findings {
		switch f.Severity {
		case "high":
			return "high"
		case "medium":
			worst = "medium"
		case "":
			// Defensive: an ungraded finding must not silently become `low`.
			worst = "medium"
		}
	}
	return worst
}

// claimsPriority mirrors severity: banned claims and impossible stat units
// outrank placeholder_contact (30); unregistered figures sit behind it; and
// findings we could not actually verify sit behind those.
func claimsPriority(findings []unverifiedClaimFinding) int {
	switch claimsSeverity(findings) {
	case "high":
		return 25
	case "medium":
		return 35
	default:
		return 45
	}
}

func claimsSummary(where string, findings []unverifiedClaimFinding) string {
	banned, numbers, stats, units := 0, 0, 0, 0
	for _, f := range findings {
		switch f.Check {
		case "banned_claim":
			banned++
		case "unregistered_number":
			numbers++
		case "unregistered_stat", "unregistered_stat_unpaired":
			stats++
		case "stat_unit_impossible", "stat_unit_mismatch":
			units++
		default:
			numbers++
		}
	}
	// Only name the categories that actually fired — a summary reading
	// "0 banned claim(s), 0 …" for every item is what makes a queue unreadable.
	parts := make([]string, 0, 4)
	for _, p := range []struct {
		n    int
		noun string
	}{
		{banned, "banned claim(s)"},
		{numbers, "unregistered number(s)"},
		{stats, "unregistered stat field(s)"},
		{units, "suspect stat unit(s)"},
	} {
		if p.n > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", p.n, p.noun))
		}
	}
	if len(parts) == 0 {
		return fmt.Sprintf("Unverified claims on %s", where)
	}
	return fmt.Sprintf("Unverified claims on %s: %s", where, strings.Join(parts, ", "))
}
