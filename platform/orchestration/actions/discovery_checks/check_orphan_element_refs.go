// FILE: platform/orchestration/actions/discovery_checks/check_orphan_element_refs.go
//
// Detects a deployed page whose own JavaScript addresses elements the page does
// not contain — the script names an id as a string literal, no element with
// that id is anywhere in the page or its chrome, and the script never creates
// one. The visitor gets a tool with no inputs, or a control that throws on the
// first click.
//
// WHY NOTHING CAUGHT THIS. webdesign.co.uk shipped 63 tools, ten of them in
// exactly this state, and every existing gate passed them:
//
//   - the page renders, deploys and returns 200, so the deploy checks are happy;
//   - tool_health and the whole Tier 1-4 acceptance ladder require
//     cc.component_level = 'tool', and these are ported-page SECTIONS, so the
//     ladder's eligibility query has never returned one of them;
//   - dead_controls deliberately does not judge JS binding statically;
//   - a browser probe can miss it too — css-filter-playground drives one
//     surviving button, something changes, and it scores as working while all
//     six of its sliders are absent.
//
// That last point is why this check earns its place beside the browser tier
// rather than being made redundant by it.
//
// The judgement itself is a fact about the artefact, not a prediction about
// runtime: the id is a literal, the document text either contains it or does
// not. That is what separates this from the case check_dead_controls refuses
// (a <button> with no visible handler, where JS may bind at runtime).
// datahelpers.OrphanElementRefs holds the rules and the reasoning for each
// conservatism; all of them exist to avoid a false positive.
//
// MEASURED BEFORE SHIPPING (2026-07-29, live DB): 98 deployed script-carrying
// pages fleet-wide, 10 flagged, all 10 on webdesign.co.uk, all 10 confirmed
// broken in a real browser. No other site produced a finding.
//
// Routing: needs_human_review with NO handler, following the check_dead_controls
// precedent. The defect is unambiguous but the repair is not mechanical — the
// missing markup has to be rebuilt to the contract the surviving CSS and JS
// describe, and the ids in the spec ARE that contract. Once a ported tool has a
// PLAN, tool-improver becomes the right handler; wiring it before then would
// hand the fixer a subject it has no document for.
//
// Registration: automatic via init(). Enable by adding "orphan_element_refs" to
// a discovery agent's checks array AFTER the image carrying this file is live.

package discovery_checks

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

func init() {
	Register(&OrphanElementRefsCheck{})
	RegisterVerifier("orphan_element_refs", VerifyOrphanElementRefsResolved)
}

type OrphanElementRefsCheck struct{}

func (c *OrphanElementRefsCheck) Name() string { return "orphan_element_refs" }

type orphanRefFinding struct {
	PageName string
	PageID   string
	PageURL  string
	Missing  []string
}

func (c *OrphanElementRefsCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	// The site chrome is part of every page. A script legitimately reaching a
	// nav toggle in the header must not be reported, so chrome is loaded once
	// and appended to every page's text before judging.
	var chrome strings.Builder
	chromeRows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT COALESCE(rendered_html, '')
		FROM site_components
		WHERE site_id = $1
	`, dctx.SiteID)
	if err == nil {
		defer chromeRows.Close()
		for chromeRows.Next() {
			var html string
			if chromeRows.Scan(&html) == nil {
				chrome.WriteString("\n")
				chrome.WriteString(html)
			}
		}
	} else {
		// Non-fatal, but it makes a false positive possible, so it is loud.
		dctx.Logger.Warn("orphan_element_refs: chrome load failed; a script "+
			"reaching a header element could now be reported",
			zap.Error(err), zap.String("site_id", dctx.SiteID.String()))
	}

	// Liveness is the COMPONENT's deploy state, not p.build_status — the
	// check_dead_controls precedent, and for the same measured reason: pages
	// routinely serve their deployed artefact while p.build_status has drifted
	// to 'needs_rebuild' (bugs_open/049/052/053). Filtering on the page status
	// would blind this check to live, broken pages.
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT p.id::text, p.name, COALESCE(p.url, ''),
		       string_agg(COALESCE(pc.rendered_html, ''), E'\n' ORDER BY pc.position)
		FROM pages p
		JOIN page_components pc ON pc.page_id = p.id
		WHERE p.site_id = $1
		  AND p.status = 'active'
		  AND pc.build_status = 'deployed'
		  AND pc.rendered_html IS NOT NULL
		  AND pc.rendered_html <> ''
		GROUP BY p.id, p.name, p.url
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("orphan_element_refs: page query failed: %w", err)
	}
	defer rows.Close()

	var findings []orphanRefFinding
	for rows.Next() {
		var pageID, pageName, pageURL, html string
		if err := rows.Scan(&pageID, &pageName, &pageURL, &html); err != nil {
			dctx.Logger.Warn("orphan_element_refs: scan error", zap.Error(err))
			continue
		}
		missing := datahelpers.OrphanElementRefs(html + chrome.String())
		if len(missing) == 0 {
			continue
		}
		findings = append(findings, orphanRefFinding{
			PageName: pageName, PageID: pageID, PageURL: pageURL, Missing: missing,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("orphan_element_refs: row scan failed: %w", err)
	}
	if len(findings) == 0 {
		return &CheckResult{}, nil
	}

	sort.Slice(findings, func(i, j int) bool { return findings[i].PageName < findings[j].PageName })

	result := &CheckResult{
		Findings: []map[string]interface{}{{
			"check":    "orphan_element_refs",
			"count":    len(findings),
			"findings": findings,
		}},
	}

	for _, f := range findings {
		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":         "orphan_element_refs",
			"page_name":     f.PageName,
			"page_id":       f.PageID,
			"page_url":      f.PageURL,
			"missing_ids":   f.Missing,
			"missing_count": len(f.Missing),
			"fix": "The page's script addresses these element ids and the page contains none of them, " +
				"so the script throws and the controls never appear. This is usually markup dropped " +
				"in a port, with the CSS and JS left intact. Rebuild the missing markup to the " +
				"contract the surviving code already describes: the ids listed here are the " +
				"specification, and the stylesheet names the classes they belong to. Do not " +
				"redesign the tool and do not delete the script to silence the error.",
		})

		var pageIDPtr *uuid.UUID
		if parsed, perr := uuid.Parse(f.PageID); perr == nil {
			pageIDPtr = &parsed
		}

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:   dctx.SiteID,
			PageID:   pageIDPtr,
			Source:   "discovery",
			Pipeline: "build",
			ItemType: "orphan_element_refs",
			Severity: "high",
			Summary: fmt.Sprintf("%s addresses %d element(s) it does not contain: %s",
				f.PageName, len(f.Missing), strings.Join(f.Missing, ", ")),
			SpecJSON:  string(specJSON),
			Priority:  35,
			Status:    "needs_human_review",
			CreatedBy: dctx.AgentType,
			ItemKey:   fmt.Sprintf("orphan_element_refs:%s:%s", f.PageName, dctx.SiteID),
			BatchID:   dctx.BatchID,
		})
	}

	dctx.Logger.Warn("orphan_element_refs: deployed pages address elements they do not contain",
		zap.Int("pages", len(findings)),
		zap.String("site_id", dctx.SiteID.String()))

	return result, nil
}

// VerifyOrphanElementRefsResolved answers, at completion time, whether the
// elements this item named are now actually on the page.
//
// It verifies THE IDS THE ITEM NAMED, not the whole-page predicate — the
// distinction the verifier registry was burned by before (verifiers.go, and the
// page_rerender entry in verifier_coverage_test.go: a verifier stricter than the
// fixer's remit strands correctly-handled items in 'failed'). The remit here is
// "make these ids exist"; a NEW orphan introduced elsewhere on the page is a new
// defect for the next discovery pass to file, not grounds to refuse this
// completion. It is reported in the detail either way, so it is never silent.
//
// Fails open on a missing page_id or an unreadable page: a verifier that cannot
// look has no opinion, and inventing 'unresolved' from an infrastructure problem
// would burn the item's attempts.
func VerifyOrphanElementRefsResolved(ctx context.Context, db *sql.DB, target VerifyTarget, logger *zap.Logger) (VerifyResult, error) {
	claimed, _ := target.Spec["missing_ids"].([]interface{})
	if len(claimed) == 0 {
		return VerifyResult{Resolved: true, Detail: "item names no missing ids; nothing to re-check"}, nil
	}
	if target.PageID == nil {
		return VerifyResult{}, fmt.Errorf("orphan_element_refs verifier: item carries no page_id")
	}

	var html string
	err := db.QueryRowContext(ctx, `
		SELECT COALESCE(string_agg(COALESCE(pc.rendered_html, ''), E'\n' ORDER BY pc.position), '')
		FROM page_components pc
		WHERE pc.page_id = $1
	`, *target.PageID).Scan(&html)
	if err != nil {
		return VerifyResult{}, fmt.Errorf("orphan_element_refs verifier: page load failed: %w", err)
	}

	var chrome strings.Builder
	chromeRows, cErr := db.QueryContext(ctx, `
		SELECT COALESCE(rendered_html, '') FROM site_components WHERE site_id = $1
	`, target.SiteID)
	if cErr == nil {
		defer chromeRows.Close()
		for chromeRows.Next() {
			var c string
			if chromeRows.Scan(&c) == nil {
				chrome.WriteString("\n")
				chrome.WriteString(c)
			}
		}
	}

	stillOrphaned := map[string]struct{}{}
	for _, id := range datahelpers.OrphanElementRefs(html + chrome.String()) {
		stillOrphaned[id] = struct{}{}
	}

	var unresolved, newlyBroken []string
	claimedSet := map[string]struct{}{}
	for _, raw := range claimed {
		id, ok := raw.(string)
		if !ok {
			continue
		}
		claimedSet[id] = struct{}{}
		if _, still := stillOrphaned[id]; still {
			unresolved = append(unresolved, id)
		}
	}
	for id := range stillOrphaned {
		if _, wasClaimed := claimedSet[id]; !wasClaimed {
			newlyBroken = append(newlyBroken, id)
		}
	}
	sort.Strings(unresolved)
	sort.Strings(newlyBroken)

	detail := fmt.Sprintf("%d of %d named element(s) now present on the page",
		len(claimedSet)-len(unresolved), len(claimedSet))
	if len(unresolved) > 0 {
		detail += "; still absent: " + strings.Join(unresolved, ", ")
	}
	if len(newlyBroken) > 0 {
		// Out of remit for this verdict, but a fixer that silences one error by
		// creating another must not do it invisibly.
		detail += "; NOT counted against this item, but the page now also addresses: " +
			strings.Join(newlyBroken, ", ")
		logger.Warn("orphan_element_refs verifier: page carries orphan refs this item did not name",
			zap.String("item_id", target.ItemID.String()),
			zap.Strings("new_orphans", newlyBroken))
	}

	return VerifyResult{Resolved: len(unresolved) == 0, Detail: detail}, nil
}
