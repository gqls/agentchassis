// FILE: platform/orchestration/actions/create_tool_component_action.go
//
// Creates a new tool component from LLM-generated HTML, creates a tool page,
// and links them via page_components. Used by tool-generator for tools
// that don't exist in the library.
//
// This is the "create from scratch" counterpart to deploy_tool_to_site
// (which forks existing library tools).
//
// Registration:
//   "create_tool_component": {
//       Handler:     CreateToolComponentAction,
//       Category:    "site",
//       Description: "Create a new tool component from generated HTML and set up its page",
//       IsLocal:     true,
//   },

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
			is_active, is_dark_section
		) VALUES ($1, $2, $3, $4, 'tool', $5, $6, $7, true, false)
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
	_, err = params.DB.ExecContext(ctx, `
		INSERT INTO page_components (
			page_id, component_id, position, build_status
		) VALUES ($1, $2, 1, 'pending')
	`, pageID, componentID)
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
		"needs_rerender": true,
		"generated":      true,
	}, nil
}

// addToolToNav inserts the tool page into the site's navigation groups.
// Best-effort — nav failures don't block tool creation.
func addToolToNav(ctx context.Context, db *sql.DB, siteID, pageID uuid.UUID,
	label, url, navSection string, inHeader, inFooter bool, logger *zap.Logger) {

	// Find or create the "Tools" nav group
	var groupID uuid.UUID
	err := db.QueryRowContext(ctx, `
		SELECT id FROM site_nav_groups
		WHERE site_id = $1 AND label = $2
		LIMIT 1
	`, siteID, navSection).Scan(&groupID)

	if err != nil {
		// Create the group
		groupID = uuid.New()
		_, err = db.ExecContext(ctx, `
			INSERT INTO site_nav_groups (id, site_id, label, sort_order, in_header, in_footer)
			VALUES ($1, $2, $3, 100, $4, $5)
			ON CONFLICT DO NOTHING
		`, groupID, siteID, navSection, inHeader, inFooter)
		if err != nil {
			logger.Warn("Failed to create nav group for tools", zap.Error(err))
			return
		}
	}

	// Add nav item
	_, err = db.ExecContext(ctx, `
		INSERT INTO site_nav_items (id, group_id, page_id, label, url, sort_order)
		VALUES ($1, $2, $3, $4, $5, 100)
		ON CONFLICT DO NOTHING
	`, uuid.New(), groupID, pageID, label, url)
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
