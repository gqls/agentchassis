// FILE: platform/orchestration/actions/discovery_checks/check_revenue_shape.go
//
// vigilant_designer_offer_analysis Programme B, phase B3 (PLAN 2026-08-02),
// sibling of check_premise_incomplete. Doc 028's rule, made mechanical:
//
//	"The revenue model shapes the site, not the other way round … Mixing shapes
//	 (a tools site with a 'Start a Project' CTA, for example) is a sign that
//	 either the classification is vague or a downstream agent is ignoring it."
//
// The always-on council seat (review_mission) enforces this on PLATFORM CODE
// submissions; nothing enforced it on a LIVE SITE until this check. Population
// measured 2026-08-08: 16/17 sites carry revenue_models.primary_model
// (10× direct_business, 3× saas_tools, 2× display_advertising,
// 1× lead_generation, 1× sponsored_listings).
//
// BRANCHES, by the site's OWN recorded primary_model (PLAN §B3):
//
//	saas_tools / display_advertising —
//	    a service-selling CTA (the lexicon below) on a page is the mixed-shape
//	    defect. Page-component hits → `revenue_shape_cta` → component-template-
//	    fixer, one item per page. Hits in SITE CHROME (site_components) are the
//	    residue: the fixer's remit is page component templates, and an automated
//	    chrome edit is fleet-wide by construction (3 head components serve every
//	    live domain — the lane's Exception 1, fork-first, A4.4 unbuilt) → ONE
//	    capability_gap via remit.go, never a dispatchable item.
//	lead_generation / direct_business —
//	    the site must offer a nav-reachable conversion path: a contact/quote-
//	    shaped deployed page whose rendered components carry a <form, and whose
//	    path appears in the site chrome (header/nav). Missing → ONE
//	    `missing_conversion_path` → content-gap-planner.
//	affiliate —
//	    no affiliate machinery exists on this platform (no link manager, no
//	    disclosure block, no partner config) → ONE capability_gap
//	    (GapHandlerMissing), never a dispatchable item.
//	sponsored_listings, and ANY model with no case above —
//	    no rule is stated, so the check examines nothing. It files ONE
//	    undispatchable capability_gap (GapRuleMissing) naming the model, because
//	    an empty result reads downstream as "examined and clean" and this one is
//	    not. v1 returned silence here and a council seat was right to call it a
//	    swallow (round-3 objection, bug_historian, 2026-08-10). See
//	    runNoRuleForModel for how to make a model permanently silent on purpose.
//	absent / empty primary_model —
//	    not this check's finding; check_premise_incomplete owns it. Skip, and
//	    file nothing: the gap arm above must never fire on a missing premise, or
//	    two checks report one defect.
//
// LEXICON DISCIPLINE: anchor/button TEXT only, never prose — an article ABOUT
// hiring an agency legitimately contains "hire a designer" in a sentence.
// Anchors via the shared datahelpers.ExtractAnchors (the misdirected_cta
// precedent); <button> text via a local scan of the same shape. Phrases are
// tight, case-insensitive, and word-bounded; "our services" is deliberately NOT
// in the list (a tools site may honestly describe what the site provides).
//
// LOCKED components (locked_at IS NOT NULL) are skipped and counted, per the
// lane's autonomy rule. ADOPTED sites: PLAN §B3 wanted needs_human_review-only
// routing, but `sites` has NO adopted/managed column to predicate on (checked
// information_schema 2026-08-09) — deferred, stated in the spec of every item
// this check files, and the locked-component skip is the protection that
// exists today.
//
// RETRACTION (RFC_010, opt-in, narrow): a page whose components were POSITIVELY
// re-scanned clean this run resolves its revenue_shape_cta item by key; a
// conversion path positively observed present resolves the site's
// missing_conversion_path. Never derived from absence; errors return error.
//
// Verifiers (fail-closed, RFC_017): both item types re-run the SAME predicate
// — revenue_shape_cta re-scans the page's unlocked components for the lexicon;
// missing_conversion_path re-runs the site-level conversion-path query.
//
// Registration: automatic via init(). Enable by adding "revenue_shape" to
// quality-discovery-agent's checks array AFTER the image rolls (fatal-if-
// unregistered since bugfix_149 B4); IMP-016 observe-only first.

package discovery_checks

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func init() {
	Register(&RevenueShapeCheck{})
	RegisterVerifier("revenue_shape_cta", VerifyRevenueShapeCTAResolved)
	RegisterVerifier("missing_conversion_path", VerifyConversionPathResolved)
}

type RevenueShapeCheck struct{}

func (c *RevenueShapeCheck) Name() string { return "revenue_shape" }

// serviceCTAPhrases is the service-selling lexicon. Word-bounded, matched
// case-insensitively against anchor/button TEXT only. Additions here widen a
// live detector fleet-wide — keep each phrase unambiguous on its own.
var serviceCTAPhrases = []string{
	"start a project",
	"start your project",
	"get a quote",
	"request a quote",
	"hire us",
	"hire me",
	"book a consultation",
	"free consultation",
	"request a proposal",
	"work with us",
	"schedule a call",
	"book a call",
}

var serviceCTARe = func() *regexp.Regexp {
	quoted := make([]string, len(serviceCTAPhrases))
	for i, p := range serviceCTAPhrases {
		quoted[i] = regexp.QuoteMeta(p)
	}
	return regexp.MustCompile(`(?i)\b(` + strings.Join(quoted, "|") + `)\b`)
}()

// buttonTextRe pulls <button>…</button> inner text — ExtractAnchors covers <a>
// only, and CTA blocks render both shapes. revenueInnerTagRe strips nested
// markup from that inner text (datahelpers keeps its equivalent unexported).
var (
	buttonTextRe      = regexp.MustCompile(`(?is)<button[^>]*>(.*?)</button>`)
	revenueInnerTagRe = regexp.MustCompile(`(?s)<[^>]*>`)
)

// contactishPageRe matches the names a conversion page ships under.
// Schema note (checked live 2026-08-09): pages has name/url/page_type — no
// filename column; url is the deployed path (e.g. /contact.html).
var contactishPageRe = `(name ILIKE '%contact%' OR name ILIKE '%quote%' OR name ILIKE '%enquir%' OR page_type = 'landing')`

// scanHTMLForServiceCTA returns each distinct lexicon hit in anchor/button text.
func scanHTMLForServiceCTA(html string) []string {
	seen := map[string]bool{}
	var hits []string
	record := func(text string) {
		if m := serviceCTARe.FindString(text); m != "" {
			key := strings.ToLower(m)
			if !seen[key] {
				seen[key] = true
				hits = append(hits, m)
			}
		}
	}
	for _, a := range datahelpers.ExtractAnchors(html) {
		record(a.Text)
	}
	for _, b := range buttonTextRe.FindAllStringSubmatch(html, -1) {
		record(revenueInnerTagRe.ReplaceAllString(b[1], " "))
	}
	return hits
}

// sitePrimaryModel reads the site's recorded shape; "" when absent.
func sitePrimaryModel(dctx DiscoveryCheckContext) (domain, primary string, err error) {
	var p sql.NullString
	err = dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT s.domain,
		       (SELECT ss.data->'revenue_models'->>'primary_model' FROM site_specs ss
		         WHERE ss.site_id = s.id AND ss.aspect = 'strategy' AND ss.is_current = true
		         ORDER BY ss.created_at DESC LIMIT 1)
		FROM sites s WHERE s.id = $1
	`, dctx.SiteID).Scan(&domain, &p)
	if p.Valid {
		primary = p.String
	}
	return domain, primary, err
}

// conversionPathState answers the lead-gen/direct question in one query:
// does a deployed contact-shaped page exist, do its components carry a form,
// and does the site chrome link to it?
type conversionPathState struct {
	PageID   string
	PageName string
	HasForm  bool
	InChrome bool
}

func readConversionPath(ctx context.Context, db *sql.DB, siteID uuid.UUID) (*conversionPathState, error) {
	var st conversionPathState
	var pageID, pageName sql.NullString
	var hasForm, inChrome sql.NullBool
	err := db.QueryRowContext(ctx, `
		WITH contactish AS (
			SELECT id, name, url FROM pages
			 WHERE site_id = $1 AND `+datahelpers.PageHasShippedPredicateFor("")+` AND `+contactishPageRe+`
			 ORDER BY (name ILIKE '%contact%') DESC, created_at ASC
			 LIMIT 1
		)
		SELECT c.id::text, c.name,
		       EXISTS (SELECT 1 FROM page_components pc
		                WHERE pc.page_id = c.id AND pc.rendered_html ILIKE '%<form%'),
		       EXISTS (SELECT 1 FROM site_components sc
		                WHERE sc.site_id = $1 AND sc.slot_name IN ('header','head')
		                  AND sc.rendered_html ILIKE '%' || c.url || '%')
		FROM contactish c
	`, siteID).Scan(&pageID, &pageName, &hasForm, &inChrome)
	if err == sql.ErrNoRows {
		return nil, nil // no contact-shaped page at all
	}
	if err != nil {
		return nil, err
	}
	if pageID.Valid {
		st.PageID = pageID.String
	}
	st.PageName = pageName.String
	st.HasForm = hasForm.Valid && hasForm.Bool
	st.InChrome = inChrome.Valid && inChrome.Bool
	return &st, nil
}

func (c *RevenueShapeCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	domain, primary, err := sitePrimaryModel(dctx)
	if err != nil {
		return nil, fmt.Errorf("revenue_shape: strategy read failed: %w", err)
	}
	if primary == "" {
		return &CheckResult{}, nil // premise_incomplete's finding, not ours
	}

	switch primary {
	case "saas_tools", "display_advertising":
		return c.runCTALexicon(dctx, domain, primary)
	case "lead_generation", "direct_business":
		return c.runConversionPath(dctx, domain, primary)
	case "affiliate":
		return &CheckResult{
			WorkItems: []WorkItemSpec{CapabilityGapItem(dctx, CapabilityGap{
				Check:         "revenue_shape",
				Pipeline:      dctx.Pipeline,
				BuilderNeeded: "affiliate-link-manager",
				GapKind:       GapHandlerMissing,
				Capability:    "manage affiliate links/disclosures so an affiliate-shaped site can be checked and fixed",
				Population:    1,
				Residue:       1,
				Examples:      []RemitCandidate{{Key: domain}},
			})},
		}, nil
	default:
		return c.runNoRuleForModel(dctx, domain, primary), nil
	}
}

// runNoRuleForModel is the arm for a recorded model this check states no rule
// for. It files ONE undispatchable capability_gap rather than returning an empty
// result, because the two are indistinguishable to every reader downstream: a
// site whose shape was never examined looks exactly like a site examined and
// found clean. sponsored_listings (vetcomparison.uk) hits this today — doc 028
// states no CTA rule for it and directory machinery already exists, so v1
// deliberately filed nothing, and the cost was a silence nobody could read.
//
// A model is examined by HAVING A CASE above; there is deliberately no parallel
// list of "known models" here to drift out of step with the strategist prompt
// that actually mints these values (six as of 2026-08-11, live
// agent_definitions row for domain-strategist). So a seventh model, or a
// hand-written strategy row carrying a typo, arrives as a gap row rather than as
// silence — the runtime is the lockstep, not a mirror.
//
// TO MAKE A MODEL PERMANENTLY SILENT, give it its own case returning an empty
// result with the reason written there. That is an explicit refusal, visible to
// whoever reviews the diff; this arm is what an omission looks like instead.
//
// One capability_gap per site per check (remit.go's item_key shape), so a site
// that changes from one unruled model to another keeps the first row's wording
// until it is closed. Bounded and stale beats churning.
func (c *RevenueShapeCheck) runNoRuleForModel(dctx DiscoveryCheckContext, domain, primary string) *CheckResult {
	return &CheckResult{
		WorkItems: []WorkItemSpec{CapabilityGapItem(dctx, CapabilityGap{
			Check:         "revenue_shape",
			Pipeline:      dctx.Pipeline,
			BuilderNeeded: "revenue_shape rule for " + primary,
			GapKind:       GapRuleMissing,
			Capability: "state what shape a " + primary + " site must have (doc 028 states none), " +
				"so revenue_shape can examine one instead of returning an empty result",
			Population: 1,
			Residue:    1,
			Examples:   []RemitCandidate{{Key: domain + " (primary_model=" + primary + ")"}},
			CodePointers: []map[string]string{{
				"path": "platform/orchestration/actions/discovery_checks/check_revenue_shape.go",
				"why":  "Run's primary_model switch has no case for " + primary + " — add one, or make the refusal explicit there",
			}},
		})},
	}
}

// runCTALexicon is the tools/advertising branch.
func (c *RevenueShapeCheck) runCTALexicon(dctx DiscoveryCheckContext, domain, primary string) (*CheckResult, error) {
	result := &CheckResult{}

	// Page surface: unlocked components of deployed pages.
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT p.id::text, p.name,
		       COALESCE(string_agg(pc.rendered_html, ' '), ''),
		       COUNT(*) FILTER (WHERE pc.locked_at IS NOT NULL)
		FROM pages p
		LEFT JOIN page_components pc ON pc.page_id = p.id AND pc.locked_at IS NULL
		WHERE p.site_id = $1 AND `+datahelpers.PageHasShippedPredicateFor("p")+`
		GROUP BY p.id, p.name
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("revenue_shape: page scan failed: %w", err)
	}
	defer rows.Close()

	lockedSkipped := 0
	for rows.Next() {
		var pageID, pageName, html string
		var lockedCount int
		if err := rows.Scan(&pageID, &pageName, &html, &lockedCount); err != nil {
			return nil, fmt.Errorf("revenue_shape: page row scan failed: %w", err)
		}
		lockedSkipped += lockedCount
		itemKey := fmt.Sprintf("revenue_shape_cta:%s", pageID)
		hits := scanHTMLForServiceCTA(html)
		if len(hits) == 0 {
			// Positively re-scanned clean → narrow retraction by key.
			result.Resolved = append(result.Resolved, ResolvedFinding{
				ItemType: "revenue_shape_cta",
				ItemKey:  itemKey,
				Reason:   fmt.Sprintf("page %q positively re-scanned: no service-selling CTA text in unlocked components", pageName),
			})
			continue
		}
		pid, perr := uuid.Parse(pageID)
		if perr != nil {
			continue
		}
		spec, _ := json.Marshal(map[string]interface{}{
			"check":         "revenue_shape",
			"primary_model": primary,
			"page_name":     pageName,
			"cta_matches":   hits,
			"rule": "doc 028: a " + primary + " site must not carry service-selling CTAs — " +
				"the revenue model shapes the site",
			"adopted_branch": "[DEFERRED] no adopted/managed column exists on sites to predicate needs_human_review routing on (2026-08-09)",
		})
		result.Findings = append(result.Findings, map[string]interface{}{
			"check": "revenue_shape", "page": pageName, "matches": hits,
		})
		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:   dctx.SiteID,
			PageID:   &pid,
			Source:   "discovery",
			Pipeline: dctx.Pipeline,
			ItemType: "revenue_shape_cta",
			Severity: "medium",
			Summary: fmt.Sprintf(
				"%s site %q carries service-selling CTA text on page %q (%s) — mixed revenue shape (doc 028)",
				primary, domain, pageName, strings.Join(hits, ", ")),
			SpecJSON:     string(spec),
			Priority:     60,
			HandlerAgent: "component-template-fixer",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      itemKey,
			BatchID:      dctx.BatchID,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("revenue_shape: page iteration failed: %w", err)
	}
	if lockedSkipped > 0 {
		result.Findings = append(result.Findings, map[string]interface{}{
			"check": "revenue_shape", "skipped_locked_components": lockedSkipped,
		})
	}

	// Chrome surface: hits here are residue — fleet-wide by construction,
	// fork-first is A4.4's unbuilt handler (lane Exception 1).
	var chromeHTML sql.NullString
	err = dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT string_agg(rendered_html, ' ') FROM site_components
		WHERE site_id = $1 AND slot_name IN ('header','footer','head')
	`, dctx.SiteID).Scan(&chromeHTML)
	if err != nil {
		return nil, fmt.Errorf("revenue_shape: chrome scan failed: %w", err)
	}
	if chromeHTML.Valid {
		if hits := scanHTMLForServiceCTA(chromeHTML.String); len(hits) > 0 {
			result.WorkItems = append(result.WorkItems, CapabilityGapItem(dctx, CapabilityGap{
				Check:         "revenue_shape",
				Pipeline:      dctx.Pipeline,
				BuilderNeeded: "chrome-fork-handler",
				GapKind:       GapHandlerRemit,
				Capability:    "fork shared chrome to site scope before editing (lane Exception 1 / A4.4), so a chrome CTA can be corrected for ONE site",
				Population:    len(hits),
				Residue:       len(hits),
				Examples:      []RemitCandidate{{Key: fmt.Sprintf("%s chrome: %s", domain, strings.Join(hits, ", "))}},
			}))
		}
	}
	return result, nil
}

// runConversionPath is the lead-gen/direct branch.
func (c *RevenueShapeCheck) runConversionPath(dctx DiscoveryCheckContext, domain, primary string) (*CheckResult, error) {
	st, err := readConversionPath(dctx.Ctx, dctx.DB, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("revenue_shape: conversion path read failed: %w", err)
	}
	itemKey := fmt.Sprintf("missing_conversion_path:%s", dctx.SiteID)

	if st != nil && st.HasForm && st.InChrome {
		return &CheckResult{
			Resolved: []ResolvedFinding{{
				ItemType: "missing_conversion_path",
				ItemKey:  itemKey,
				Reason: fmt.Sprintf(
					"conversion path positively observed: page %q has a form and is linked from site chrome", st.PageName),
			}},
		}, nil
	}

	missing := "no deployed contact/quote-shaped page exists"
	if st != nil {
		switch {
		case !st.HasForm && !st.InChrome:
			missing = fmt.Sprintf("page %q has no form and is not linked from the site chrome", st.PageName)
		case !st.HasForm:
			missing = fmt.Sprintf("page %q is nav-reachable but carries no form", st.PageName)
		default:
			missing = fmt.Sprintf("page %q has a form but is not linked from the site chrome", st.PageName)
		}
	}
	spec, _ := json.Marshal(map[string]interface{}{
		"check":          "revenue_shape",
		"primary_model":  primary,
		"missing":        missing,
		"rule":           "doc 028: a " + primary + " site earns through enquiries — a nav-reachable page with a form is the minimum conversion path",
		"adopted_branch": "[DEFERRED] no adopted/managed column exists on sites to predicate needs_human_review routing on (2026-08-09)",
	})
	return &CheckResult{
		Findings: []map[string]interface{}{{
			"check": "revenue_shape", "missing_conversion_path": missing,
		}},
		WorkItems: []WorkItemSpec{{
			SiteID:   dctx.SiteID,
			Source:   "discovery",
			Pipeline: dctx.Pipeline,
			ItemType: "missing_conversion_path",
			Severity: "medium",
			Summary: fmt.Sprintf(
				"%s site %q has no working conversion path: %s (doc 028 — the revenue model shapes the site)",
				primary, domain, missing),
			SpecJSON:     string(spec),
			Priority:     60,
			HandlerAgent: "content-gap-planner",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      itemKey,
			BatchID:      dctx.BatchID,
		}},
	}, nil
}

// VerifyRevenueShapeCTAResolved re-runs the page-scoped half of the detector's
// predicate (verifiers.go contract: the SAME predicate, so they cannot drift —
// both call scanHTMLForServiceCTA).
func VerifyRevenueShapeCTAResolved(ctx context.Context, db *sql.DB, target VerifyTarget, logger *zap.Logger) (VerifyResult, error) {
	if target.PageID == nil {
		return VerifyResult{}, fmt.Errorf("revenue_shape_cta item %s has no page_id", target.ItemID)
	}
	var html sql.NullString
	err := db.QueryRowContext(ctx, `
		SELECT COALESCE(string_agg(rendered_html, ' '), '') FROM page_components
		WHERE page_id = $1 AND locked_at IS NULL
	`, *target.PageID).Scan(&html)
	if err != nil {
		return VerifyResult{}, fmt.Errorf("revenue_shape_cta verify read failed: %w", err)
	}
	hits := scanHTMLForServiceCTA(html.String)
	if len(hits) > 0 {
		return VerifyResult{Resolved: false,
			Detail: fmt.Sprintf("service-selling CTA text still present: %s", strings.Join(hits, ", "))}, nil
	}
	return VerifyResult{Resolved: true,
		Detail: "no service-selling CTA text in the page's unlocked components"}, nil
}

// VerifyConversionPathResolved re-runs the site-level conversion-path predicate.
func VerifyConversionPathResolved(ctx context.Context, db *sql.DB, target VerifyTarget, logger *zap.Logger) (VerifyResult, error) {
	st, err := readConversionPath(ctx, db, target.SiteID)
	if err != nil {
		return VerifyResult{}, fmt.Errorf("missing_conversion_path verify read failed: %w", err)
	}
	if st != nil && st.HasForm && st.InChrome {
		return VerifyResult{Resolved: true,
			Detail: fmt.Sprintf("page %q has a form and is linked from site chrome", st.PageName)}, nil
	}
	detail := "no deployed contact/quote-shaped page exists"
	if st != nil {
		detail = fmt.Sprintf("page %q: has_form=%v, in_chrome=%v", st.PageName, st.HasForm, st.InChrome)
	}
	return VerifyResult{Resolved: false, Detail: detail}, nil
}
