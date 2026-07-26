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
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func init() { Register(&UnverifiedClaimsCheck{}) }

type UnverifiedClaimsCheck struct{}

func (c *UnverifiedClaimsCheck) Name() string { return "unverified_claims" }

func (c *UnverifiedClaimsCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	eb, rowExists, err := loadEvidenceBaseForCheck(dctx)
	if err != nil {
		return nil, err
	}

	// NO early return on eb == nil any more (bugs_open/093). The HTML scans
	// still need a register and are gated on eb below, but the stat-unit lint
	// compares a component against itself and runs fleet-wide — the same scope
	// it has in the build gate. Returning here would have kept the sites with
	// no register unaudited on BOTH paths, which is the gap this bug is about.
	pageFindings, siteFindings, err := scanDeployedClaims(dctx, eb, rowExists)
	if err != nil {
		return nil, err
	}
	if len(pageFindings) == 0 && len(siteFindings) == 0 {
		return &CheckResult{}, nil
	}

	total := 0
	for _, pf := range pageFindings {
		total += len(pf.Findings)
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

// pageClaimFindings groups findings for one page (one work item per page).
type pageClaimFindings struct {
	PageID   string
	PageName string
	Findings []unverifiedClaimFinding
}

// loadEvidenceBaseForCheck reads the site's current evidence_base spec.
//
// The second return is whether the ROW EXISTS, which is NOT the same question
// as whether eb is non-nil, and the difference is load-bearing (bugs_open/093,
// and the landmine that made the build gate's check 9 key on the row too):
// ParseEvidenceBase returns nil for a row with no facts[] and no
// banned_claims[], so a writer_block-only row reads as "not opted in" while the
// writer block itself goes on working. Three sites sat like that for two days
// with both claims checkers silently off. The HTML scans still need eb (they
// have nothing to scan against without it); the stat scans use rowExists.
func loadEvidenceBaseForCheck(dctx DiscoveryCheckContext) (*datahelpers.EvidenceBase, bool, error) {
	var data []byte
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT data FROM site_specs
		WHERE site_id = $1 AND aspect = 'evidence_base' AND is_current = true
	`, dctx.SiteID).Scan(&data)
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

// scanDeployedClaims runs the scans over deployed page and site components,
// skipping locked components (human-pinned content is never flagged).
//
// TWO SURFACES PER COMPONENT, and each scan is gated on the one it reads:
// the HTML scans on rendered_html, the stat scans on content_data
// (bugs_open/093). The row predicate is therefore an OR — a component with
// stored content_data but no rendered_html has never been published, but it is
// exactly what the next re-render WILL publish, unchecked.
func scanDeployedClaims(dctx DiscoveryCheckContext, eb *datahelpers.EvidenceBase, rowExists bool) ([]pageClaimFindings, []unverifiedClaimFinding, error) {
	// page_components — body sections.
	pageRows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT pc.page_id::text, p.name, COALESCE(pc.slot_name, ''),
		       COALESCE(pc.rendered_html, ''), COALESCE(cc.name, ''),
		       COALESCE(pc.content_data::text, '')
		FROM page_components pc
		JOIN pages p ON p.id = pc.page_id
		LEFT JOIN content_components cc ON cc.id = pc.component_id
		WHERE p.site_id = $1
		  AND pc.locked_at IS NULL
		  AND ( (pc.rendered_html IS NOT NULL AND pc.rendered_html <> '')
		     OR (pc.content_data IS NOT NULL AND pc.content_data::text <> '{}') )
		ORDER BY p.name, pc.position
	`, dctx.SiteID)
	if err != nil {
		return nil, nil, fmt.Errorf("unverified_claims page query failed: %w", err)
	}
	defer pageRows.Close()

	byPage := map[string]*pageClaimFindings{}
	var pageOrder []string

	for pageRows.Next() {
		var pageID, pageName, slotName, html, component, contentJSON string
		if err := pageRows.Scan(&pageID, &pageName, &slotName, &html, &component, &contentJSON); err != nil {
			dctx.Logger.Warn("unverified_claims: page scan error", zap.Error(err))
			continue
		}
		findings := scanComponentClaims(html, slotName, eb)
		findings = append(findings, scanStoredStatClaims(
			component, slotName, decodeContentData(contentJSON, dctx.Logger), eb, rowExists)...)
		if len(findings) == 0 {
			continue
		}
		pf, ok := byPage[pageID]
		if !ok {
			pf = &pageClaimFindings{PageID: pageID, PageName: pageName}
			byPage[pageID] = pf
			pageOrder = append(pageOrder, pageID)
		}
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
	siteRows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT COALESCE(sc.slot_name, ''), COALESCE(sc.rendered_html, ''),
		       COALESCE(cc.name, ''), COALESCE(sc.content_data::text, '')
		FROM site_components sc
		LEFT JOIN content_components cc ON cc.id = sc.component_id
		WHERE sc.site_id = $1
		  AND sc.locked_at IS NULL
		  AND ( (sc.rendered_html IS NOT NULL AND sc.rendered_html <> '')
		     OR (sc.content_data IS NOT NULL AND sc.content_data::text <> '{}') )
	`, dctx.SiteID)
	if err != nil {
		return nil, nil, fmt.Errorf("unverified_claims site query failed: %w", err)
	}
	defer siteRows.Close()

	var siteFindings []unverifiedClaimFinding
	for siteRows.Next() {
		var slotName, html, component, contentJSON string
		if err := siteRows.Scan(&slotName, &html, &component, &contentJSON); err != nil {
			dctx.Logger.Warn("unverified_claims: site scan error", zap.Error(err))
			continue
		}
		siteFindings = append(siteFindings, scanComponentClaims(html, slotName, eb)...)
		siteFindings = append(siteFindings, scanStoredStatClaims(
			component, slotName, decodeContentData(contentJSON, dctx.Logger), eb, rowExists)...)
	}
	if err := siteRows.Err(); err != nil {
		return nil, nil, fmt.Errorf("unverified_claims site iteration failed: %w", err)
	}

	ordered := make([]pageClaimFindings, 0, len(pageOrder))
	for _, id := range pageOrder {
		ordered = append(ordered, *byPage[id])
	}
	return ordered, siteFindings, nil
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
// Gated on eb: with no register there is nothing to compare prose against, and
// flagging every number on a site that never opted in would be noise, not a
// finding. (The stat scans are different — see scanStoredStatClaims.)
func scanComponentClaims(html, slotName string, eb *datahelpers.EvidenceBase) []unverifiedClaimFinding {
	if eb == nil || html == "" {
		return nil
	}
	blocks := datahelpers.ExtractAssertionText(html)
	var out []unverifiedClaimFinding
	for _, f := range eb.ScanBannedClaims(blocks) {
		out = append(out, unverifiedClaimFinding{
			Check: f.Check, SlotName: slotName, Matched: f.Matched,
			Pattern: f.Pattern, Reason: f.Reason, Snippet: f.Snippet,
			Occurrences: f.Occurrences,
			// A banned claim is a known falsehood, live on the site.
			Severity: "high", Source: "rendered_html",
		})
	}
	for _, f := range eb.ScanUnregisteredNumbers(blocks) {
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
