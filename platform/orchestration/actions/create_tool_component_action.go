// FILE: platform/orchestration/actions/create_tool_component_action.go
//
// Creates a new tool component from LLM-generated HTML, creates a tool page,
// and links them via page_components. Used by tool-generator for tools
// that don't exist in the library.
//
// This is the "create from scratch" counterpart to deploy_tool_to_site
// (which forks existing library tools).
//
// After creation, this action also:
//   - Sets rendered_html on the page_component (so the tool is visible immediately)
//   - Creates a needs_content_page work item (so hero/intro/CTA sections get written)
//   - Creates a companion guide page + work item (same pattern as deploy_tool_action)
//
// Registration:
//
//	"create_tool_component": {
//	    Handler:     CreateToolComponentAction,
//	    Category:    "site",
//	    Description: "Create a new tool component from generated HTML and set up its page",
//	    IsLocal:     true,
//	},

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ============================================================================
// ActionInputSpec
// ============================================================================

var CreateToolComponentInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id", "html_content", "function", "display_name"},
	Optional:   []string{"description", "category"},
	Defaults:   map[string]interface{}{"category": "interactive"},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("create_tool_component", CreateToolComponentInputSpec)
}

// ============================================================================
// ACTION: create_tool_component
// ============================================================================

func CreateToolComponentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "create_tool_component"),
	)

	logger.Info("CreateToolComponentAction: Starting")

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
		CreateToolComponentInputSpec,
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

	htmlContent := inputs.Get("html_content")
	if htmlContent == "" {
		return nil, fmt.Errorf("html_content is empty — LLM generation may have failed")
	}

	// Strip markdown fences if LLM included them despite instructions
	htmlContent = datahelpers.StripCodeFences(htmlContent)

	function := inputs.Get("function")
	displayName := inputs.Get("display_name")
	description := inputs.Get("description")
	category := inputs.Get("category")
	if category == "" {
		category = "interactive"
	}

	// Read config for nav/page settings
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

	// --- Load site domain ---
	var siteDomain string
	err = params.DB.QueryRowContext(ctx,
		`SELECT domain FROM sites WHERE id = $1`, siteID,
	).Scan(&siteDomain)
	if err != nil {
		return nil, fmt.Errorf("failed to load site domain: %w", err)
	}

	// Sanitise function name
	function = sanitiseFunction(function)

	// Build component name (unique per site)
	domainSlug := strings.ReplaceAll(siteDomain, ".", "-")
	componentName := fmt.Sprintf("%s-%s", function, domainSlug)

	logger.Info("CreateToolComponentAction: Creating tool",
		zap.String("site_id", siteIDStr),
		zap.String("domain", siteDomain),
		zap.String("function", function),
		zap.String("display_name", displayName),
		zap.String("component_name", componentName),
	)

	// --- Check if already exists ---
	var existingID string
	err = params.DB.QueryRowContext(ctx, `
		SELECT cc.id::text FROM content_components cc
		JOIN page_components pc ON pc.component_id = cc.id
		JOIN pages p ON pc.page_id = p.id
		WHERE cc.function = $1
		  AND cc.component_level = 'tool'
		  AND p.site_id = $2
		  AND cc.is_active = true
		LIMIT 1
	`, function, siteID).Scan(&existingID)

	if err == nil && existingID != "" {
		logger.Info("CreateToolComponentAction: Tool already exists for this site",
			zap.String("existing_id", existingID))
		return map[string]interface{}{
			"already_exists": true,
			"component_id":   existingID,
			"function":       function,
		}, nil
	}

	// --- Create component ---
	componentID := uuid.New()

	_, err = params.DB.ExecContext(ctx, `
		INSERT INTO content_components (
			id, name, display_name, function, component_level,
			category, description, html_template,
			is_active, is_dark_section, created_from
		) VALUES ($1, $2, $3, $4, 'tool', $5, $6, $7, true, false, 'tool-generator')
	`, componentID, componentName, displayName, function,
		category, description, htmlContent)
	if err != nil {
		return nil, fmt.Errorf("failed to create tool component: %w", err)
	}

	logger.Info("CreateToolComponentAction: Component created",
		zap.String("component_id", componentID.String()))

	// --- Create page ---
	pageID := uuid.New()
	pageName := function
	pageSlug := fmt.Sprintf("tools/%s", function)
	pageURL := fmt.Sprintf("/%s.html", pageSlug)
	pageTitle := fmt.Sprintf("%s | %s", displayName, navSection)

	_, err = params.DB.ExecContext(ctx, `
		INSERT INTO pages (
			id, site_id, name, url, title,
			page_type, status, build_status, nav_order, meta_description
		) VALUES ($1, $2, $3, $4, $5, 'tool', 'active', 'planned', 200, $6)
	`, pageID, siteID, pageName, pageURL, pageTitle, description)
	if err != nil {
		// If page creation fails, clean up the component
		params.DB.ExecContext(ctx, `DELETE FROM content_components WHERE id = $1`, componentID)
		return nil, fmt.Errorf("failed to create tool page: %w", err)
	}

	logger.Info("CreateToolComponentAction: Page created",
		zap.String("page_id", pageID.String()),
		zap.String("url", pageURL))

	// --- Link component to page ---
	// Position 2 (same as deploy_tool_action): tool widget sits between intro and CTA
	// Set rendered_html so the tool is visible immediately
	// Set slot_name to the function (component naming contract)
	_, err = params.DB.ExecContext(ctx, `
		INSERT INTO page_components (
			page_id, component_id, position, slot_name,
			rendered_html, content_data, build_status
		) VALUES ($1, $2, 2, $3, $4, '{}'::jsonb, 'deployed')
	`, pageID, componentID, function, htmlContent)
	if err != nil {
		// Clean up on failure
		params.DB.ExecContext(ctx, `DELETE FROM pages WHERE id = $1`, pageID)
		params.DB.ExecContext(ctx, `DELETE FROM content_components WHERE id = $1`, componentID)
		return nil, fmt.Errorf("failed to link component to page: %w", err)
	}

	// --- Add to nav (if configured) ---
	if inHeader || inFooter {
		addToolToNav(ctx, params.DB, siteID, pageID, displayName, pageURL, navSection, inHeader, inFooter, logger)
	}

	// --- Create needs_content_page work item for tool page ---
	// This triggers page-build-handler to write hero, intro, and CTA sections
	// around the tool (the tool widget at position 2 is already deployed).
	toolContentSpec, _ := json.Marshal(map[string]interface{}{
		"page_name":         pageName,
		"page_id":           pageID.String(),
		"page_type":         "tool",
		"tool_function":     function,
		"tool_display_name": displayName,
		"tool_description":  description,
		"tool_page_url":     pageURL,
		"source":            "tool-generator",
		"content_guidance": fmt.Sprintf(
			"This is a tool page for '%s'. Generate: (1) a hero section with the tool name and a one-line benefit statement, "+
				"(2) an educational guide section explaining the concept behind the tool — what it calculates, why it matters, "+
				"how users should interpret the results — written for the site's target audience, "+
				"(3) a CTA section encouraging users to try the tool and linking to related content. "+
				"Do NOT regenerate the tool widget itself — it is already deployed at position 2.",
			displayName,
		),
	})

	_, err = params.DB.ExecContext(ctx, `
		INSERT INTO site_work_items (
			site_id, source, pipeline, item_type, severity, summary,
			spec, page_id, priority, handler_agent, status, created_by, item_key
		) VALUES (
			$1, 'tool-generator', 'build', 'needs_content_page', 'medium',
			$2, $3::jsonb, $4, 50, 'page-build-handler', 'triaged', 'tool-generator', $5
		) ON CONFLICT DO NOTHING
	`, siteID,
		fmt.Sprintf("Write content for tool page: %s", displayName),
		string(toolContentSpec), pageID,
		fmt.Sprintf("tool_content:%s:%s", function, siteID),
	)
	if err != nil {
		logger.Warn("CreateToolComponentAction: Failed to create tool content work item (non-fatal)", zap.Error(err))
	}

	// --- Create companion guide page ---
	guideName := pageName + "-guide"
	guideURL := fmt.Sprintf("/guides/%s.html", guideName)
	guideTitle := fmt.Sprintf("Understanding %s | Guide", displayName)

	var guidePageID uuid.UUID
	err = params.DB.QueryRowContext(ctx, `
		INSERT INTO pages (
			site_id, name, url, title, page_type,
			nav_order, in_header, in_footer,
			meta_description, sections,
			build_status, status
		) VALUES (
			$1, $2, $3, $4, 'blog-post',
			200, false, false,
			$5, '["hero", "article-body", "call-to-action"]'::jsonb,
			'planned', 'active'
		)
		ON CONFLICT (site_id, name) DO UPDATE SET
			title = EXCLUDED.title,
			updated_at = NOW()
		RETURNING id
	`, siteID, guideName, guideURL, guideTitle,
		fmt.Sprintf("A practical guide to %s — what it means, how it works, and how to use our interactive %s.",
			strings.TrimPrefix(displayName, "UK "),
			strings.ToLower(displayName)),
	).Scan(&guidePageID)

	if err != nil {
		logger.Warn("CreateToolComponentAction: Failed to create companion guide page (non-fatal)", zap.Error(err))
	} else {
		// Create work item for the guide article
		guideSpec, _ := json.Marshal(map[string]interface{}{
			"page_name":         guideName,
			"page_id":           guidePageID.String(),
			"page_type":         "blog-post",
			"tool_function":     function,
			"tool_display_name": displayName,
			"tool_page_url":     pageURL,
			"source":            "tool-generator",
			"content_guidance": fmt.Sprintf(
				"Write an in-depth guide about %s. Explain the concept, why it matters, common mistakes people make, "+
					"and practical tips. At relevant points, reference the interactive %s tool at %s — "+
					"e.g. 'Use our %s to see how this applies to your situation.' "+
					"The article should stand alone as useful content even without the tool.",
				strings.TrimPrefix(displayName, "UK "),
				strings.ToLower(displayName), pageURL,
				strings.ToLower(displayName),
			),
		})

		_, err = params.DB.ExecContext(ctx, `
			INSERT INTO site_work_items (
				site_id, source, pipeline, item_type, severity, summary,
				spec, page_id, priority, handler_agent, status, created_by, item_key
			) VALUES (
				$1, 'tool-generator', 'build', 'needs_content_page', 'low',
				$2, $3::jsonb, $4, 70, 'page-build-handler', 'triaged', 'tool-generator', $5
			) ON CONFLICT DO NOTHING
		`, siteID,
			fmt.Sprintf("Write companion guide: %s", guideTitle),
			string(guideSpec), guidePageID,
			fmt.Sprintf("tool_guide:%s:%s", function, siteID),
		)
		if err != nil {
			logger.Warn("CreateToolComponentAction: Failed to create guide work item (non-fatal)", zap.Error(err))
		} else {
			logger.Info("CreateToolComponentAction: Companion guide created",
				zap.String("guide_page_id", guidePageID.String()),
				zap.String("guide_url", guideURL))
		}
	}

	logger.Info("CreateToolComponentAction: Complete",
		zap.String("component_id", componentID.String()),
		zap.String("page_id", pageID.String()),
		zap.String("page_url", pageURL),
		zap.String("function", function))

	return map[string]interface{}{
		"component_id":   componentID.String(),
		"page_id":        pageID.String(),
		"page_url":       pageURL,
		"function":       function,
		"display_name":   displayName,
		"guide_url":      guideURL,
		"needs_rerender": true,
		"generated":      true,
	}, nil
}

// addToolToNav inserts the tool page into the site's navigation groups.
// Best-effort — nav failures don't block tool creation.
//
// Schema:
//   site_nav_groups: id, site_id, group_key, group_label, group_type, position
//   site_nav_items:  id, site_id, group_id, label, url, page_id, item_type, position, status
func addToolToNav(ctx context.Context, db *sql.DB, siteID, pageID uuid.UUID,
	label, url, navSection string, inHeader, inFooter bool, logger *zap.Logger) {

	// Find or create the "Tools" nav group
	// group_key = "tools", group_type = "primary" so it appears in the header
	var groupID uuid.UUID
	err := db.QueryRowContext(ctx, `
		SELECT id FROM site_nav_groups
		WHERE site_id = $1 AND group_key = 'tools'
		LIMIT 1
	`, siteID).Scan(&groupID)

	if err != nil {
		// Create the group using ON CONFLICT on (site_id, group_key)
		err = db.QueryRowContext(ctx, `
			INSERT INTO site_nav_groups (site_id, group_key, group_label, group_type, position)
			VALUES ($1, 'tools', $2, 'primary', 100)
			ON CONFLICT (site_id, group_key) DO UPDATE SET
				group_label = EXCLUDED.group_label,
				updated_at = NOW()
			RETURNING id
		`, siteID, navSection).Scan(&groupID)
		if err != nil {
			logger.Warn("Failed to create nav group for tools", zap.Error(err))
			return
		}
	}

	// Add nav item (skip if already exists for this page)
	_, err = db.ExecContext(ctx, `
		INSERT INTO site_nav_items (site_id, group_id, label, url, page_id, item_type, position, status)
		SELECT $1, $2, $3, $4, $5, 'page_link', 100, 'active'
		WHERE NOT EXISTS (
			SELECT 1 FROM site_nav_items
			WHERE site_id = $1 AND group_id = $2 AND page_id = $5
		)
	`, siteID, groupID, label, url, pageID)
	if err != nil {
		logger.Warn("Failed to add tool to nav", zap.Error(err))
	}
}

// sanitiseFunction ensures function name is valid kebab-case with tool- prefix
func sanitiseFunction(function string) string {
	function = strings.ToLower(strings.TrimSpace(function))
	function = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' {
			return r
		}
		if r == ' ' || r == '_' {
			return '-'
		}
		return -1
	}, function)
	if !strings.HasPrefix(function, "tool-") {
		function = "tool-" + function
	}
	return function
}
