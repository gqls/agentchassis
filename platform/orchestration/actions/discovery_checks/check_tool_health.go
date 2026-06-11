// FILE: platform/orchestration/actions/discovery_checks/check_tool_health.go
//
// Discovery check: tool_health
//
// Audits deployed tools on a site for structural and quality issues.
// Does NOT judge whether a tool is appropriate for the site — that's
// an admin decision. This check asks: "is the tool actually working?"
//
// Checks per tool:
//   1. Page deployed    — build_status = 'deployed' (blocker if not)
//   2. HTML present     — page_component.rendered_html non-empty (blocker)
//   3. Template present — fork html_template non-empty (blocker)
//   4. Has script       — <script> block exists (error — tool isn't interactive)
//   5. Has style        — <style> block exists (warning — may have no layout)
//   6. Mobile-ready     — CSS contains @media breakpoint (warning)
//   7. CSS variables    — no bare hex outside var() fallbacks (warning)
//   8. Self-contained   — no fetch(), no external src= (warning)
//   9. Doc header       — tool-doc header present in html_template (warning;
//                         019 §Tool Doc Header — template-only: rendered_html
//                         is post-strip)
//  10. Doc header shape — opener without closer (error — the malformed block
//                         would SHIP; StripToolDocHeader leaves it untouched)
//
// Creates improve_tool work items for tool-improver to fix.
// Cooldown: skips tools that had an improve_tool item in the last 7 days.

package discovery_checks

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"go.uber.org/zap"

	// Sentinel consts + HasToolDocHeader (019 §Tool Doc Header). Reconcile the
	// import path with wherever tool_doc_header.go lands (drafted as
	// platform/content, beside the render path that calls StripToolDocHeader).
	"github.com/gqls/agentchassis/platform/content"
)

func init() { Register(&ToolHealthCheck{}) }

type ToolHealthCheck struct{}

func (c *ToolHealthCheck) Name() string { return "tool_health" }

func (c *ToolHealthCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	result := &CheckResult{}

	// Load all deployed tools for this site
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT
			cc.id::text             AS component_id,
			cc.function,
			cc.display_name,
			COALESCE(cc.html_template, '') AS template_html,
			COALESCE(pc.rendered_html, '') AS rendered_html,
			p.id::text              AS page_id,
			p.name                  AS page_name,
			COALESCE(p.build_status, '') AS build_status
		FROM content_components cc
		JOIN page_components pc ON pc.component_id = cc.id
		JOIN pages p ON pc.page_id = p.id
		WHERE p.site_id = $1
		  AND cc.component_level = 'tool'
		  AND cc.is_active = true
		  AND p.status = 'active'
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("tool_health: query failed: %w", err)
	}
	defer rows.Close()

	type deployedTool struct {
		ComponentID  string
		Function     string
		DisplayName  string
		TemplateHTML string
		RenderedHTML string
		PageID       string
		PageName     string
		BuildStatus  string
	}

	var tools []deployedTool
	for rows.Next() {
		var t deployedTool
		if err := rows.Scan(&t.ComponentID, &t.Function, &t.DisplayName,
			&t.TemplateHTML, &t.RenderedHTML,
			&t.PageID, &t.PageName, &t.BuildStatus); err != nil {
			dctx.Logger.Warn("tool_health: scan error", zap.Error(err))
			continue
		}
		tools = append(tools, t)
	}

	if len(tools) == 0 {
		return result, nil // No tools deployed — nothing to check
	}

	// Load recent improve_tool items to avoid duplicate work items
	recentItems := map[string]bool{}
	itemRows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT spec->>'component_id'
		FROM site_work_items
		WHERE site_id = $1
		  AND item_type = 'improve_tool'
		  AND created_at > NOW() - INTERVAL '7 days'
	`, dctx.SiteID)
	if err == nil {
		defer itemRows.Close()
		for itemRows.Next() {
			var compID sql.NullString
			if itemRows.Scan(&compID) == nil && compID.Valid {
				recentItems[compID.String] = true
			}
		}
	}

	// Check each tool
	for _, tool := range tools {
		if recentItems[tool.ComponentID] {
			continue // Already has a recent improve_tool item
		}

		issues := auditTool(tool.TemplateHTML, tool.RenderedHTML, tool.BuildStatus)
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

		// Build issue description for the improve_tool spec
		var issueDescriptions []string
		for _, issue := range issues {
			issueDescriptions = append(issueDescriptions,
				fmt.Sprintf("[%s] %s", issue.severity, issue.description))
		}
		issueText := strings.Join(issueDescriptions, "; ")

		result.Findings = append(result.Findings, map[string]interface{}{
			"check":        "tool_health",
			"tool":         tool.Function,
			"display_name": tool.DisplayName,
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

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:   dctx.SiteID,
			Source:   "discovery",
			Pipeline: "build",
			ItemType: "improve_tool",
			Severity: highestSeverity,
			Summary:  fmt.Sprintf("Fix %s: %s", tool.DisplayName, firstIssue(issues)),
			SpecJSON: fmt.Sprintf(
				`{"component_id": "%s", "issue": "%s", "check": "tool_health", "page_id": "%s", "page_name": "%s"}`,
				tool.ComponentID,
				escapeJSON(issueText),
				tool.PageID,
				tool.PageName,
			),
			Priority:     priority,
			HandlerAgent: "tool-improver",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("tool_health:%s:%s", tool.Function, dctx.SiteID),
			BatchID:      dctx.BatchID,
		})

		dctx.Logger.Info("tool_health: issues found",
			zap.String("tool", tool.Function),
			zap.Int("issue_count", len(issues)),
			zap.String("severity", highestSeverity),
			zap.String("issues", issueText))
	}

	// ── Tier 2: Queue LLM audit for tools that pass structural checks ──
	// Tools with blockers need fixing first; tools without blockers
	// get queued for deeper code review by tool-auditor.
	for _, tool := range tools {
		if recentItems[tool.ComponentID] {
			continue
		}
		// Skip if this tool had structural blockers
		issues := auditTool(tool.TemplateHTML, tool.RenderedHTML, tool.BuildStatus)
		hasBlocker := false
		for _, issue := range issues {
			if issue.severity == "blocker" {
				hasBlocker = true
				break
			}
		}
		if hasBlocker {
			continue // Tier 1 improve_tool item already created above
		}

		// Check if recently audited by tool-auditor (30-day cooldown)
		var recentAudit bool
		_ = dctx.DB.QueryRowContext(dctx.Ctx, `
			SELECT EXISTS (
				SELECT 1 FROM site_work_items
				WHERE site_id = $1
				  AND item_type = 'audit_tool'
				  AND spec->>'component_id' = $2
				  AND created_at > NOW() - INTERVAL '30 days'
			)
		`, dctx.SiteID, tool.ComponentID).Scan(&recentAudit)

		if recentAudit {
			continue
		}

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:   dctx.SiteID,
			Source:   "discovery",
			Pipeline: "build",
			ItemType: "audit_tool",
			Severity: "low",
			Summary:  fmt.Sprintf("LLM code review: %s", tool.DisplayName),
			SpecJSON: fmt.Sprintf(
				`{"component_id": "%s", "check": "tool_health", "page_id": "%s", "page_name": "%s"}`,
				tool.ComponentID, tool.PageID, tool.PageName,
			),
			Priority:     140,
			HandlerAgent: "tool-auditor",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("audit_tool:%s:%s", tool.Function, dctx.SiteID),
			BatchID:      dctx.BatchID,
		})

		dctx.Logger.Info("tool_health: queued LLM audit",
			zap.String("tool", tool.Function),
			zap.String("component_id", tool.ComponentID))
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

func auditTool(templateHTML, renderedHTML, buildStatus string) []toolIssue {
	var issues []toolIssue

	// Use the best available HTML for content checks
	html := renderedHTML
	if html == "" {
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

	// 3. No template HTML (fork is empty)
	if templateHTML == "" {
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
	// only — never on rendered_html, which is post-strip once
	// StripToolDocHeader ships. Guarded on templateHTML: when the fork is
	// empty, empty_template (blocker) has already fired.
	if templateHTML != "" && !content.HasToolDocHeader(templateHTML) {
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

func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}
