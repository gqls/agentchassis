// FILE: platform/orchestration/actions/deploy_tool_action.go
//
// DeployToolToSiteAction forks a library tool into a site-owned copy,
// creates a page for it, and inserts the page_component.
//
// Fork-on-deploy model:
//   - Library tool: content_components WHERE component_level='tool' AND forked_from IS NULL
//   - Site fork:    new content_components row with forked_from = library tool ID
//   - Changes to the library tool do NOT cascade to existing forks
//
// The forked component + page are then picked up by the normal
// render/deploy pipeline (rerender-pages agent).
//
// Config:
//   - nav_section: string (default "Tools") — nav group for the tool page
//   - in_header:   bool (default true) — show in header nav under Tools
//   - in_footer:   bool (default false)
//
// Data inputs (via ActionInputSpec):
//   - site_id             (required) — target site
//   - tool_component_id   (required) — library tool to fork (from work item spec)
//
// Resolved from work item spec or explicit input:
//   - page_name           (optional) — override auto-generated page name
//   - page_title          (optional) — override auto-generated page title

package actions

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ============================================================================
// ActionInputSpec
// ============================================================================

var DeployToolToSiteInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id", "tool_component_id"},
	Optional:   []string{"page_name", "page_title"},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("deploy_tool_to_site", DeployToolToSiteInputSpec)
}

// ============================================================================
// ACTION: deploy_tool_to_site
// ============================================================================

func DeployToolToSiteAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "deploy_tool_to_site"),
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	logger.Info("DeployToolToSiteAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	// --- Resolve inputs ---
	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		DeployToolToSiteInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	toolIDStr := inputs.Get("tool_component_id")
	toolID, err := uuid.Parse(toolIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid tool_component_id: %w", err)
	}

	pageNameOverride := inputs.Get("page_name")
	pageTitleOverride := inputs.Get("page_title")

	config := params.StepConfig.Config
	navSection := "Tools"
	if v, ok := config["nav_section"].(string); ok && v != "" {
		navSection = v
	}
	inHeader := true
	if v, ok := config["in_header"].(bool); ok {
		inHeader = v
	}
	inFooter := false
	if v, ok := config["in_footer"].(bool); ok {
		inFooter = v
	}

	// --- Load site domain for logging and URL construction ---
	var siteDomain string
	_ = params.DB.QueryRowContext(ctx,
		`SELECT domain FROM sites WHERE id = $1`, siteID,
	).Scan(&siteDomain)

	logger.Info("DeployToolToSiteAction: Deploying tool",
		zap.String("site_id", siteIDStr),
		zap.String("domain", siteDomain),
		zap.String("tool_component_id", toolIDStr),
	)

	// --- 1. Load library tool ---
	var toolName, toolDisplayName, toolFunction, toolCategory string
	var toolDescription, toolHTMLTemplate sql.NullString
	var toolSemanticTags, toolInputSchema sql.NullString

	err = params.DB.QueryRowContext(ctx, `
		SELECT name, display_name, function, category,
		       description, html_template, semantic_tags::text, input_schema::text
		FROM content_components
		WHERE id = $1
		  AND component_level = 'tool'
		  AND is_active = true
	`, toolID).Scan(
		&toolName, &toolDisplayName, &toolFunction, &toolCategory,
		&toolDescription, &toolHTMLTemplate, &toolSemanticTags, &toolInputSchema,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tool component %s not found or not a tool", toolIDStr)
		}
		return nil, fmt.Errorf("load tool: %w", err)
	}

	if !toolHTMLTemplate.Valid || toolHTMLTemplate.String == "" {
		return nil, fmt.Errorf("tool %s has no HTML template", toolFunction)
	}

	// --- 2. Check if already deployed (fork exists for this site) ---
	var existingForkID sql.NullString
	err = params.DB.QueryRowContext(ctx, `
		SELECT cc.id::text
		FROM content_components cc
		JOIN page_components pc ON pc.component_id = cc.id
		JOIN pages p ON pc.page_id = p.id
		WHERE cc.forked_from = $1
		  AND p.site_id = $2
		  AND cc.is_active = true
		LIMIT 1
	`, toolID, siteID).Scan(&existingForkID)

	if err == nil && existingForkID.Valid {
		logger.Info("DeployToolToSiteAction: Tool already deployed to site",
			zap.String("existing_fork_id", existingForkID.String))
		return map[string]interface{}{
			"site_id":          siteIDStr,
			"tool_function":    toolFunction,
			"already_deployed": true,
			"fork_id":          existingForkID.String,
			"needs_rerender":   false,
		}, nil
	}

	// --- 3. Fork the tool (create site-owned copy) ---
	forkID := uuid.New()
	forkName := toolName + "-" + domainSlug(siteDomain)

	_, err = params.DB.ExecContext(ctx, `
		INSERT INTO content_components (
			id, name, display_name, function, category, component_level, render_mode,
			is_dark_section, is_active, description,
			semantic_tags, html_template, input_schema, forked_from
		)
		SELECT
			$1,
			$2,
			display_name,
			function,
			category,
			component_level,
			render_mode,
			is_dark_section,
			true,
			description,
			semantic_tags,
			html_template,
			input_schema,
			$3
		FROM content_components
		WHERE id = $3
	`, forkID, forkName, toolID)
	if err != nil {
		return nil, fmt.Errorf("fork tool: %w", err)
	}

	logger.Info("DeployToolToSiteAction: Tool forked",
		zap.String("fork_id", forkID.String()),
		zap.String("fork_name", forkName),
		zap.String("forked_from", toolIDStr),
	)

	// --- 4. Create tool page ---
	pageName := pageNameOverride
	if pageName == "" {
		// Derive from function: tool-ab-test-calculator → ab-test-calculator
		pageName = strings.TrimPrefix(toolFunction, "tool-")
	}

	pageTitle := pageTitleOverride
	if pageTitle == "" {
		pageTitle = toolDisplayName
	}

	pageURL := "/tools/" + pageName + ".html"

	// Get next nav order for tools section
	var maxNavOrder int
	_ = params.DB.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(nav_order), 0) FROM pages WHERE site_id = $1 AND url LIKE '/tools/%'`,
		siteID,
	).Scan(&maxNavOrder)

	sectionsJSON := fmt.Sprintf(`["%s"]`, toolFunction)

	var pageID uuid.UUID
	err = params.DB.QueryRowContext(ctx, `
		INSERT INTO pages (
			site_id, name, url, title, page_type,
			nav_label, nav_order, in_header, in_footer,
			meta_description, sections,
			build_status, status
		) VALUES (
			$1, $2, $3, $4, 'tool',
			$5, $6, $7, $8,
			$9, $10::jsonb,
			'planned', 'active'
		)
		ON CONFLICT (site_id, name) DO UPDATE SET
			url = EXCLUDED.url,
			title = EXCLUDED.title,
			updated_at = NOW()
		RETURNING id
	`, siteID, pageName, pageURL, pageTitle,
		navSection+" / "+toolDisplayName, maxNavOrder+1, inHeader, inFooter,
		toolDescription.String, sectionsJSON,
	).Scan(&pageID)
	if err != nil {
		return nil, fmt.Errorf("create tool page: %w", err)
	}

	logger.Info("DeployToolToSiteAction: Page created",
		zap.String("page_id", pageID.String()),
		zap.String("page_name", pageName),
		zap.String("url", pageURL),
	)

	// --- 5. Create page_component linking fork to page ---
	var pcID uuid.UUID
	err = params.DB.QueryRowContext(ctx, `
		INSERT INTO page_components (
			page_id, component_id, position, slot_name,
			rendered_html, content_data, build_status
		) VALUES (
			$1, $2, 0, $3,
			$4, '{}'::jsonb, 'pending'
		)
		ON CONFLICT DO NOTHING
		RETURNING id
	`, pageID, forkID, toolFunction,
		toolHTMLTemplate.String,
	).Scan(&pcID)
	if err != nil {
		// May be a conflict — check if already exists
		logger.Warn("DeployToolToSiteAction: page_component insert issue",
			zap.Error(err))
	}

	logger.Info("DeployToolToSiteAction: Tool deployed",
		zap.String("site_id", siteIDStr),
		zap.String("domain", siteDomain),
		zap.String("tool", toolFunction),
		zap.String("fork_id", forkID.String()),
		zap.String("page_id", pageID.String()),
	)

	return map[string]interface{}{
		"site_id":           siteIDStr,
		"domain":            siteDomain,
		"tool_function":     toolFunction,
		"tool_display_name": toolDisplayName,
		"fork_id":           forkID.String(),
		"page_id":           pageID.String(),
		"page_url":          pageURL,
		"already_deployed":  false,
		"needs_rerender":    true,
	}, nil
}

// domainSlug returns a slug-safe version of the full domain.
// The full domain is preserved because .co.uk and .uk are different sites
// and the TLD may be the only differentiator.
// e.g. "website-design.co.uk" → "website-design-co-uk"
//
//	"website-design.uk"    → "website-design-uk"
func domainSlug(domain string) string {
	if domain == "" {
		return "site"
	}
	return strings.ReplaceAll(domain, ".", "-")
}
