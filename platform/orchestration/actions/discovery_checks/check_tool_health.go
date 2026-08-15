// FILE: platform/orchestration/actions/discovery_checks/check_tool_health.go
//
// Discovery check: tool_health
//
// Audits deployed tools on a site for structural and quality issues.
// Does NOT judge whether a tool is appropriate for the site — that's
// an admin decision. This check asks: "is the tool actually working?"
//
// WHICH TOOLS (widened 2026-08-15, bugs_open/281). The population is the
// tool-acceptance ladder's — toolEligibilityWhere in tool_eligibility.go — so
// a tool is either
//
//	(a) a real tool component (component_level='tool', a per-site fork whose
//	    contract is its html_template), keyed by cc.function; or
//	(b) a PORTED tool: the sole component on a page_type='tool' page, one
//	    instance of a shared section-level component whose actual tool code
//	    lives in that page's page_components.rendered_html, keyed by the page
//	    name minus its 'tool-' prefix.
//
// Until 2026-08-15 this check scoped on `cc.component_level='tool'` alone and
// so saw 4 of webdesign.co.uk's 67 tools; the other 63 (one shared
// 'ported-page' row, 115 page instances fleet-wide) were never examined, which
// is how the Mind Map Studio's illegible controls waited for the owner to find
// them by hand. tool_eligibility.go recorded the exclusion as deliberate
// (noise: ~71 improve_tool items in one pass, and no PLANs to judge them by).
// That objection is answered STRUCTURALLY here rather than dismissed:
//
//   - a ported instance's findings are filed as `ported_tool_fix` with NO
//     handler (needs_human_review), never as improve_tool: the only fixer,
//     tool-improver, rewrites content_components.html_template, which for a
//     ported instance is the SHARED wrapper — that write fanned out to every
//     ported page on three sites on 2026-08-05 and again on 2026-08-14. Same
//     posture as check_orphan_element_refs (no fixer until the tool has a PLAN
//     and a per-instance repair path exists).
//   - identity is per INSTANCE (component_id + page_id): item_keys carry the
//     subject key (all 63 ported instances share cc.function='ported-page', so a
//     function-keyed item_key would collapse them onto ONE row via
//     idx_swi_dedup), and the cooldowns are instance-scoped for ported tools.
//     A real fork's keys and cooldowns are byte-for-byte what they were.
//   - Tier-2 audit_tool queueing is capped per run so the first sweep does not
//     dump 60 LLM audits into the queue at once.
//
// Checks per tool (the ones marked TEMPLATE run only for real forks — a ported
// instance's html_template is the shared passthrough wrapper, not its own):
//   1. Page deployed    — build_status = 'deployed' (blocker if not)
//   2. HTML present     — page_component.rendered_html non-empty (blocker)
//   3. Template present — fork html_template non-empty (blocker)      TEMPLATE
//   4. Has script       — <script> block exists (error — tool isn't interactive)
//   5. Has style        — <style> block exists (warning — may have no layout)
//   6. Mobile-ready     — CSS contains @media breakpoint (warning)
//   7. CSS variables    — no bare hex outside var() fallbacks (warning)
//   8. Self-contained   — no fetch(), no external src= (warning)
//   9. Doc header       — tool-doc header present in html_template   TEMPLATE
//                         (warning; 019 §Tool Doc Header — template-only: the
//                         strip runs at deploy assembly, so rendered_html may
//                         legitimately retain the header)
//  10. Doc header shape — opener without closer (error — the malformed  TEMPLATE
//                         block would SHIP; StripToolDocHeader leaves it untouched)
//
// Creates improve_tool work items (forks) / ported_tool_fix items (ported) for
// structural issues, and queues audit_tool for tool-auditor's LLM review.
// Cooldown: skips tools that had an item in the last 7 days.

package discovery_checks

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	// Sentinel consts + HasToolDocHeader (019 §Tool Doc Header). Reconcile the
	// import path with wherever tool_doc_header.go lands (drafted as
	// platform/content, beside the render path that calls StripToolDocHeader).
	"github.com/gqls/agentchassis/platform/content"
)

func init() { Register(&ToolHealthCheck{}) }

// maxNewAuditItemsPerRun bounds how many audit_tool items one sweep may queue.
// tool-auditor is a Sonnet review per item; with the ported population in
// scope a first pass would otherwise queue ~60 at once. Deterministic ORDER BY
// p.name plus the 30-day audit cooldown means a capped sweep converges over
// successive rotation passes without re-queueing anything. It never binds for
// a site with only real forks.
const maxNewAuditItemsPerRun = 12

type ToolHealthCheck struct{}

func (c *ToolHealthCheck) Name() string { return "tool_health" }

type deployedTool struct {
	ComponentID    string
	Function       string
	ComponentLevel string
	SubjectKey     string
	DisplayName    string
	TemplateHTML   string
	RenderedHTML   string
	PageID         string
	PageName       string
	BuildStatus    string
}

// isFork: a real tool component whose contract is its own html_template. A
// ported instance (anything else the ladder admits) is one page's blob in a
// shared wrapper, and its truth is rendered_html.
func (t deployedTool) isFork() bool { return t.ComponentLevel == "tool" }

func (c *ToolHealthCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	result := &CheckResult{}

	// Load every tool the ladder admits for this site (tool_eligibility.go).
	// display_name falls back to the subject key for a ported instance, whose
	// component row is the generic shared blob and would otherwise label 63
	// different tools identically (check_tool_acceptance precedent).
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT
			cc.id::text             AS component_id,
			cc.function,
			cc.component_level,
			`+toolSubjectKeyExpr+` AS subject_key,
			CASE WHEN cc.component_level = 'tool'
			     THEN COALESCE(cc.display_name, cc.function)
			     ELSE `+toolSubjectKeyExpr+`
			END AS display_name,
			COALESCE(cc.html_template, '') AS template_html,
			COALESCE(pc.rendered_html, '') AS rendered_html,
			p.id::text              AS page_id,
			p.name                  AS page_name,
			COALESCE(p.build_status, '') AS build_status
		FROM content_components cc
		JOIN page_components pc ON pc.component_id = cc.id
		JOIN pages p ON pc.page_id = p.id
		WHERE p.site_id = $1`+toolEligibilityWhere+`
		ORDER BY p.name
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("tool_health: query failed: %w", err)
	}
	defer rows.Close()

	var tools []deployedTool
	for rows.Next() {
		var t deployedTool
		if err := rows.Scan(&t.ComponentID, &t.Function, &t.ComponentLevel, &t.SubjectKey,
			&t.DisplayName, &t.TemplateHTML, &t.RenderedHTML,
			&t.PageID, &t.PageName, &t.BuildStatus); err != nil {
			dctx.Logger.Warn("tool_health: scan error", zap.Error(err))
			continue
		}
		tools = append(tools, t)
	}

	if len(tools) == 0 {
		return result, nil // No tools deployed — nothing to check
	}

	// Recent structural items, to avoid duplicate work. Two scopes from one
	// query: a real fork is cooled by ANY item on its component (shared with
	// check_tool_acceptance, so the two checks never double-team tool-improver);
	// a ported instance is cooled only by an item on ITS page — the component
	// is shared by every ported page on the site, so a component-keyed cooldown
	// would let one item silence all of them.
	recentComponent := map[string]bool{}
	recentInstance := map[string]bool{}
	itemRows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT spec->>'component_id', COALESCE(spec->>'page_id', '')
		FROM site_work_items
		WHERE site_id = $1
		  AND item_type IN ('improve_tool', 'ported_tool_fix')
		  AND status <> 'cancelled'
		  AND created_at > NOW() - INTERVAL '7 days'
	`, dctx.SiteID)
	if err == nil {
		defer itemRows.Close()
		for itemRows.Next() {
			var compID, pageID sql.NullString
			if itemRows.Scan(&compID, &pageID) == nil && compID.Valid {
				recentComponent[compID.String] = true
				recentInstance[compID.String+":"+pageID.String] = true
			}
		}
	}
	onCooldown := func(t deployedTool) bool {
		if t.isFork() {
			return recentComponent[t.ComponentID]
		}
		return recentInstance[t.ComponentID+":"+t.PageID]
	}

	// ── Tier 1: structural audit ──
	for _, tool := range tools {
		if onCooldown(tool) {
			continue
		}

		issues := auditTool(tool.TemplateHTML, tool.RenderedHTML, tool.BuildStatus, tool.isFork())
		if len(issues) == 0 {
			continue
		}

		// Determine highest severity
		highestSeverity := "low"
		for _, issue := range issues {
			if issue.severity == "blocker" {
				highestSeverity = "high"
				break
			}
			if issue.severity == "error" && highestSeverity != "high" {
				highestSeverity = "medium"
			}
		}

		var issueDescriptions []string
		for _, issue := range issues {
			issueDescriptions = append(issueDescriptions,
				fmt.Sprintf("[%s] %s", issue.severity, issue.description))
		}
		issueText := strings.Join(issueDescriptions, "; ")

		result.Findings = append(result.Findings, map[string]interface{}{
			"check":        "tool_health",
			"tool":         tool.SubjectKey,
			"display_name": tool.DisplayName,
			"ported":       !tool.isFork(),
			"issues":       len(issues),
			"severity":     highestSeverity,
		})

		// Map severity to work item priority
		priority := 80
		switch highestSeverity {
		case "high":
			priority = 30
		case "medium":
			priority = 60
		}

		spec := map[string]interface{}{
			"component_id": tool.ComponentID,
			"issue":        issueText,
			"check":        "tool_health",
			"page_id":      tool.PageID,
			"page_name":    tool.PageName,
			"subject_key":  tool.SubjectKey,
		}
		specJSON, _ := json.Marshal(spec)

		if tool.isFork() {
			result.WorkItems = append(result.WorkItems, WorkItemSpec{
				SiteID:       dctx.SiteID,
				PageID:       parsePageUUID(tool.PageID),
				Source:       "discovery",
				Pipeline:     "build",
				ItemType:     "improve_tool",
				Severity:     highestSeverity,
				Summary:      fmt.Sprintf("Fix %s: %s", tool.DisplayName, firstIssue(issues)),
				SpecJSON:     string(specJSON),
				Priority:     priority,
				HandlerAgent: "tool-improver",
				Status:       "detected",
				CreatedBy:    dctx.AgentType,
				ItemKey:      fmt.Sprintf("tool_health:%s:%s", tool.SubjectKey, dctx.SiteID),
				BatchID:      dctx.BatchID,
			})
		} else {
			// NO handler, deliberately. tool-improver's writeback targets the
			// component's html_template — for a ported instance that is the
			// wrapper shared by every ported page (bugs_open/281 mechanism 2;
			// clobbered fleet-wide 2026-08-05 and 2026-08-14). The per-instance
			// artefact is page_components.rendered_html and no automated fixer
			// edits it from a finding today, so this is a human's to route.
			result.WorkItems = append(result.WorkItems, WorkItemSpec{
				SiteID:    dctx.SiteID,
				PageID:    parsePageUUID(tool.PageID),
				Source:    "discovery",
				Pipeline:  "build",
				ItemType:  "ported_tool_fix",
				Severity:  highestSeverity,
				Summary:   fmt.Sprintf("Ported tool %s: %s", tool.DisplayName, firstIssue(issues)),
				SpecJSON:  string(specJSON),
				Priority:  priority,
				Status:    "needs_human_review",
				CreatedBy: dctx.AgentType,
				ItemKey:   fmt.Sprintf("ported_tool_fix:tool_health:%s:%s", tool.SubjectKey, dctx.SiteID),
				BatchID:   dctx.BatchID,
			})
		}

		dctx.Logger.Info("tool_health: issues found",
			zap.String("tool", tool.SubjectKey),
			zap.Bool("ported", !tool.isFork()),
			zap.Int("issue_count", len(issues)),
			zap.String("severity", highestSeverity),
			zap.String("issues", issueText))
	}

	// ── Tier 2: Queue LLM audit for tools that pass structural checks ──
	// Tools with blockers need fixing first; tools without blockers
	// get queued for deeper code review by tool-auditor.
	queuedAudits := 0
	for _, tool := range tools {
		if queuedAudits >= maxNewAuditItemsPerRun {
			dctx.Logger.Info("tool_health: audit queue cap reached for this run",
				zap.Int("cap", maxNewAuditItemsPerRun))
			break
		}
		if onCooldown(tool) {
			continue
		}
		// Skip if this tool had structural blockers
		issues := auditTool(tool.TemplateHTML, tool.RenderedHTML, tool.BuildStatus, tool.isFork())
		hasBlocker := false
		for _, issue := range issues {
			if issue.severity == "blocker" {
				hasBlocker = true
				break
			}
		}
		if hasBlocker {
			continue // Tier 1 item already created above
		}

		// Check if recently audited by tool-auditor (30-day cooldown) — per
		// component for a fork, per instance for a ported tool.
		var recentAudit bool
		_ = dctx.DB.QueryRowContext(dctx.Ctx, `
			SELECT EXISTS (
				SELECT 1 FROM site_work_items
				WHERE site_id = $1
				  AND item_type = 'audit_tool'
				  AND spec->>'component_id' = $2
				  AND ($3 = 'tool' OR spec->>'page_id' = $4)
				  AND created_at > NOW() - INTERVAL '30 days'
			)
		`, dctx.SiteID, tool.ComponentID, tool.ComponentLevel, tool.PageID).Scan(&recentAudit)

		if recentAudit {
			continue
		}

		spec := map[string]interface{}{
			"component_id": tool.ComponentID,
			"check":        "tool_health",
			"page_id":      tool.PageID,
			"page_name":    tool.PageName,
			"subject_key":  tool.SubjectKey,
		}
		specJSON, _ := json.Marshal(spec)

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			PageID:       parsePageUUID(tool.PageID),
			Source:       "discovery",
			Pipeline:     "build",
			ItemType:     "audit_tool",
			Severity:     "low",
			Summary:      fmt.Sprintf("LLM code review: %s", tool.DisplayName),
			SpecJSON:     string(specJSON),
			Priority:     140,
			HandlerAgent: "tool-auditor",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("audit_tool:%s:%s", tool.SubjectKey, dctx.SiteID),
			BatchID:      dctx.BatchID,
		})
		queuedAudits++

		dctx.Logger.Info("tool_health: queued LLM audit",
			zap.String("tool", tool.SubjectKey),
			zap.Bool("ported", !tool.isFork()),
			zap.String("component_id", tool.ComponentID),
			zap.String("page_id", tool.PageID))
	}

	return result, nil
}

// ============================================================================
// Audit checks
// ============================================================================

type toolIssue struct {
	check       string
	severity    string // "blocker", "error", "warning"
	description string
}

// Regex patterns — compiled once
var (
	scriptTagRe      = regexp.MustCompile(`(?i)<script[\s>]`)
	styleTagRe       = regexp.MustCompile(`(?i)<style[\s>]`)
	mediaQueryRe     = regexp.MustCompile(`@media\s*\(`)
	bareHexRe        = regexp.MustCompile(`(?i)(?:background|color|border)(?:-\w+)*\s*:\s*#[0-9a-f]{3,8}`)
	cssVarFallbackRe = regexp.MustCompile(`var\([^)]+,\s*#[0-9a-f]{3,8}\)`)
	externalFetchRe  = regexp.MustCompile(`(?i)fetch\s*\(`)
	externalSrcRe    = regexp.MustCompile(`(?i)src\s*=\s*["']https?://`)
	cdnLinkRe        = regexp.MustCompile(`(?i)cdn\.|cdnjs\.|unpkg\.com|jsdelivr\.net`)
)

// auditTool runs the structural checks. templateIsFork says whether
// templateHTML is the tool's OWN contract (a real fork) or the shared wrapper a
// ported instance sits in; the TEMPLATE-marked checks in the file header only
// run for a fork, and a ported instance is judged on renderedHTML alone (no
// fallback to the wrapper, which has no script/style of its own and would draw
// every content warning at once).
func auditTool(templateHTML, renderedHTML, buildStatus string, templateIsFork bool) []toolIssue {
	var issues []toolIssue

	// Use the best available HTML for content checks
	html := renderedHTML
	if html == "" && templateIsFork {
		html = templateHTML
	}

	// 1. Page not deployed
	if buildStatus != "deployed" && buildStatus != "active" {
		issues = append(issues, toolIssue{
			check:       "not_deployed",
			severity:    "blocker",
			description: fmt.Sprintf("Tool page has build_status '%s' — not yet rendered and deployed", buildStatus),
		})
	}

	// 2. No rendered HTML
	if renderedHTML == "" {
		issues = append(issues, toolIssue{
			check:       "no_rendered_html",
			severity:    "blocker",
			description: "Page component has no rendered HTML — the render pipeline may have skipped this tool",
		})
	}

	// 3. No template HTML (fork is empty) — a fork's contract only.
	if templateIsFork && templateHTML == "" {
		issues = append(issues, toolIssue{
			check:       "empty_template",
			severity:    "blocker",
			description: "Tool fork has no html_template — was likely forked from an empty library tool",
		})
	}

	// If no HTML at all, skip content checks
	if html == "" {
		return issues
	}

	// 4. No script tag (not interactive)
	if !scriptTagRe.MatchString(html) {
		issues = append(issues, toolIssue{
			check:       "no_script",
			severity:    "error",
			description: "Tool has no <script> block — it cannot be interactive without JavaScript",
		})
	}

	// 5. No style tag
	if !styleTagRe.MatchString(html) {
		issues = append(issues, toolIssue{
			check:       "no_style",
			severity:    "warning",
			description: "Tool has no <style> block — layout may depend entirely on global CSS",
		})
	}

	// 6. No responsive breakpoint
	if !mediaQueryRe.MatchString(html) {
		issues = append(issues, toolIssue{
			check:       "no_responsive",
			severity:    "warning",
			description: "Tool CSS has no @media breakpoint — may not work well on mobile devices",
		})
	}

	// 7. Hardcoded colours (bare hex not inside var() fallback)
	bareHexMatches := bareHexRe.FindAllString(html, -1)
	if len(bareHexMatches) > 0 {
		// Filter out ones that are inside var() fallbacks
		nonFallbackCount := 0
		for _, match := range bareHexMatches {
			// Check if this hex appears inside a var() context
			if !cssVarFallbackRe.MatchString(match) {
				nonFallbackCount++
			}
		}
		if nonFallbackCount > 3 {
			issues = append(issues, toolIssue{
				check:       "hardcoded_colors",
				severity:    "warning",
				description: fmt.Sprintf("Tool has %d hardcoded colour values outside CSS variable fallbacks — should use var(--color-*)", nonFallbackCount),
			})
		}
	}

	// 8. External dependencies
	if externalFetchRe.MatchString(html) {
		issues = append(issues, toolIssue{
			check:       "external_fetch",
			severity:    "warning",
			description: "Tool uses fetch() — tools should be self-contained with no API calls",
		})
	}
	if cdnLinkRe.MatchString(html) {
		issues = append(issues, toolIssue{
			check:       "external_cdn",
			severity:    "warning",
			description: "Tool references an external CDN — tools should have no external dependencies",
		})
	}

	// 9./10. Tool-doc header (019 §Tool Doc Header). Checked on the TEMPLATE
	// only — the template is the contract's home; the strip runs at deploy
	// assembly (page HTML + JS assets), so rendered_html may legitimately
	// retain the header. Guarded on templateHTML: when the fork is empty,
	// empty_template (blocker) has already fired. A ported instance's template
	// is the shared wrapper, which is nobody's contract — skipped.
	if templateIsFork && templateHTML != "" && !content.HasToolDocHeader(templateHTML) {
		if strings.Contains(templateHTML, content.ToolDocOpen) {
			// Opener without closer: StripToolDocHeader deliberately leaves a
			// malformed block untouched (truncating a script would be worse),
			// so until fixed the block SHIPS to the public page.
			issues = append(issues, toolIssue{
				check:       "malformed_doc_header",
				severity:    "error",
				description: "Tool-doc header opener present but closer missing — the block will ship to the public page until fixed (019 §Tool Doc Header)",
			})
		} else {
			issues = append(issues, toolIssue{
				check:       "no_doc_header",
				severity:    "warning",
				description: "Tool has no tool-doc header — add the sentinel block stating purpose and behavioural invariants (019 §Tool Doc Header)",
			})
		}
	}

	return issues
}

// ============================================================================
// Helpers
// ============================================================================

func firstIssue(issues []toolIssue) string {
	if len(issues) == 0 {
		return "unknown issue"
	}
	return issues[0].description
}

func parsePageUUID(s string) *uuid.UUID {
	if pu, err := uuid.Parse(s); err == nil {
		return &pu
	}
	return nil
}

// escapeJSON is a package helper (check_backend_unreachable,
// check_site_unreachable still build their specs with it).
func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}
