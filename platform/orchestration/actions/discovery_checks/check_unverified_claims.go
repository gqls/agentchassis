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
// Opt-in: sites without a current site_specs 'evidence_base' aspect are
// skipped entirely. banned_claims starts empty elsewhere — never block or
// flag fleet-wide on a layer only one site has data for.
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

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func init() { Register(&UnverifiedClaimsCheck{}) }

type UnverifiedClaimsCheck struct{}

func (c *UnverifiedClaimsCheck) Name() string { return "unverified_claims" }

func (c *UnverifiedClaimsCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	eb, err := loadEvidenceBaseForCheck(dctx)
	if err != nil {
		return nil, err
	}
	if eb == nil {
		return &CheckResult{}, nil // site not opted in — no evidence base
	}

	pageFindings, siteFindings, err := scanDeployedClaims(dctx, eb)
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

	// One work item per affected page.
	for _, pf := range pageFindings {
		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":     "unverified_claims",
			"page_id":   pf.PageID,
			"page_name": pf.PageName,
			"findings":  pf.Findings,
			"audit_doc": eb.AuditDoc,
			"fix": "Human review required. Banned claims are audited-out fabrications for this " +
				"site (see reasons and the audit doc); unregistered numbers need either an " +
				"evidence_base fact row (if true and provable) or removal from the copy. " +
				"Truth decisions are human — do not auto-rewrite.",
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
			"audit_doc": eb.AuditDoc,
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
type unverifiedClaimFinding struct {
	Check       string `json:"check"` // banned_claim | unregistered_number
	SlotName    string `json:"slot_name"`
	Matched     string `json:"matched"`
	Pattern     string `json:"pattern,omitempty"`
	Reason      string `json:"reason"`
	Snippet     string `json:"snippet"`
	Occurrences int    `json:"occurrences"`
}

// pageClaimFindings groups findings for one page (one work item per page).
type pageClaimFindings struct {
	PageID   string
	PageName string
	Findings []unverifiedClaimFinding
}

// loadEvidenceBaseForCheck reads the site's current evidence_base spec.
// nil, nil means the site has not opted in.
func loadEvidenceBaseForCheck(dctx DiscoveryCheckContext) (*datahelpers.EvidenceBase, error) {
	var data []byte
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT data FROM site_specs
		WHERE site_id = $1 AND aspect = 'evidence_base' AND is_current = true
	`, dctx.SiteID).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("unverified_claims evidence_base query failed: %w", err)
	}
	eb, err := datahelpers.ParseEvidenceBase(data)
	if err != nil {
		// A malformed evidence base is a real defect on an opted-in site —
		// surface it as a check error (logged, not fatal to the run).
		return nil, fmt.Errorf("unverified_claims evidence_base parse failed: %w", err)
	}
	return eb, nil
}

// scanDeployedClaims runs both scans over deployed page and site components,
// skipping locked components (human-pinned content is never flagged).
func scanDeployedClaims(dctx DiscoveryCheckContext, eb *datahelpers.EvidenceBase) ([]pageClaimFindings, []unverifiedClaimFinding, error) {
	// page_components — body sections.
	pageRows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT pc.page_id::text, p.name, COALESCE(pc.slot_name, ''), pc.rendered_html
		FROM page_components pc
		JOIN pages p ON p.id = pc.page_id
		WHERE p.site_id = $1
		  AND pc.rendered_html IS NOT NULL AND pc.rendered_html <> ''
		  AND pc.locked_at IS NULL
		ORDER BY p.name, pc.position
	`, dctx.SiteID)
	if err != nil {
		return nil, nil, fmt.Errorf("unverified_claims page query failed: %w", err)
	}
	defer pageRows.Close()

	byPage := map[string]*pageClaimFindings{}
	var pageOrder []string

	for pageRows.Next() {
		var pageID, pageName, slotName, html string
		if err := pageRows.Scan(&pageID, &pageName, &slotName, &html); err != nil {
			dctx.Logger.Warn("unverified_claims: page scan error", zap.Error(err))
			continue
		}
		findings := scanComponentClaims(html, slotName, eb)
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

	// site_components — header/footer/head chrome.
	siteRows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT COALESCE(sc.slot_name, ''), sc.rendered_html
		FROM site_components sc
		WHERE sc.site_id = $1
		  AND sc.rendered_html IS NOT NULL AND sc.rendered_html <> ''
		  AND sc.locked_at IS NULL
	`, dctx.SiteID)
	if err != nil {
		return nil, nil, fmt.Errorf("unverified_claims site query failed: %w", err)
	}
	defer siteRows.Close()

	var siteFindings []unverifiedClaimFinding
	for siteRows.Next() {
		var slotName, html string
		if err := siteRows.Scan(&slotName, &html); err != nil {
			dctx.Logger.Warn("unverified_claims: site scan error", zap.Error(err))
			continue
		}
		siteFindings = append(siteFindings, scanComponentClaims(html, slotName, eb)...)
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

// scanComponentClaims runs both shared scans over one component's HTML.
func scanComponentClaims(html, slotName string, eb *datahelpers.EvidenceBase) []unverifiedClaimFinding {
	blocks := datahelpers.ExtractAssertionText(html)
	var out []unverifiedClaimFinding
	for _, f := range eb.ScanBannedClaims(blocks) {
		out = append(out, unverifiedClaimFinding{
			Check: f.Check, SlotName: slotName, Matched: f.Matched,
			Pattern: f.Pattern, Reason: f.Reason, Snippet: f.Snippet,
			Occurrences: f.Occurrences,
		})
	}
	for _, f := range eb.ScanUnregisteredNumbers(blocks) {
		out = append(out, unverifiedClaimFinding{
			Check: f.Check, SlotName: slotName, Matched: f.Matched,
			Reason: f.Reason, Snippet: f.Snippet, Occurrences: f.Occurrences,
		})
	}
	return out
}

// claimsSeverity: a banned claim is a known falsehood live on the site — high.
// Unregistered numbers alone are medium (extraction has false positives).
func claimsSeverity(findings []unverifiedClaimFinding) string {
	for _, f := range findings {
		if f.Check == "banned_claim" {
			return "high"
		}
	}
	return "medium"
}

// claimsPriority mirrors severity: banned claims outrank placeholder_contact
// (30); number-only findings sit behind it.
func claimsPriority(findings []unverifiedClaimFinding) int {
	if claimsSeverity(findings) == "high" {
		return 25
	}
	return 35
}

func claimsSummary(where string, findings []unverifiedClaimFinding) string {
	banned, numbers := 0, 0
	for _, f := range findings {
		if f.Check == "banned_claim" {
			banned++
		} else {
			numbers++
		}
	}
	return fmt.Sprintf("Unverified claims on %s: %d banned claim(s), %d unregistered number(s)",
		where, banned, numbers)
}
