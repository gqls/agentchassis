// FILE: platform/orchestration/actions/render_js_snippets_for_site_action.go
// RenderJSSnippetsForSiteAction concatenates active js_snippets that apply to
// the site's components into a single JS bundle for /assets/js/snippets.js.
//
// Mirrors loadComponentCSSSnippets's shape in render_css_from_spec_action.go
// but simpler — no theme composition, just snippet concatenation.
//
// The result is intended to be deployed by a downstream git_commit step
// using files_field: js_snippets_render.files. The head component template
// references /assets/js/snippets.js unconditionally (synchronous <script>
// before </head>), and an empty bundle is written when no active snippets
// apply so the file always exists on the CDN.
//
// Activation gate: only rows with is_active = true contribute. The 9 existing
// js_snippets rows (scroll-reveal, smooth-scroll, etc.) default to is_active
// = false and stay dormant. Operators activate snippets by flipping the flag
// on a per-row basis.

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

// RenderJSSnippetsForSiteAction renders the per-site JS snippet bundle.
//
// Config:
//   - site_id_field: path to site_id in collected data
//     (default: "site_record.site_id")
//   - components_field: path to component function list in collected data
//     (default: "site_context.all_component_functions")
//     If absent or empty, components are loaded from the DB via
//     loadSiteComponentFunctionsForJS.
//
// Returns:
//   - files: { "assets/js/snippets.js": <bundled JS> } — for git_commit
//   - domain: the site's domain (for git_commit's domain_field)
//   - js_content: the bundle string itself (for logging/debugging)
//   - snippet_count: number of active snippets included
//   - snippet_names: names of included snippets (for logging)
func RenderJSSnippetsForSiteAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("RenderJSSnippetsForSiteAction: Starting")

	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	config := params.StepConfig.Config

	// 1. Site ID
	siteIDField := "site_record.site_id"
	if s, ok := config["site_id_field"].(string); ok && s != "" {
		siteIDField = s
	}
	siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
	if siteIDStr == "" {
		return nil, fmt.Errorf("site_id not found at %s", siteIDField)
	}
	siteUUID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// 2. Domain — for the git_commit step
	domain := datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.domain")
	if domain == "" {
		domain = datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.domain")
	}
	if domain == "" {
		// Fall back to the sites table
		_ = params.DB.QueryRowContext(ctx,
			`SELECT COALESCE(domain, '') FROM sites WHERE id = $1`, siteUUID).Scan(&domain)
	}
	if domain == "" {
		return nil, fmt.Errorf("domain not found for site %s", siteIDStr)
	}

	// 3. Component list — collected data first, DB fallback
	componentsField := "site_context.all_component_functions"
	if s, ok := config["components_field"].(string); ok && s != "" {
		componentsField = s
	}
	components := extractComponentFunctionsList(params.CollectedData, componentsField, params.Logger)
	if len(components) == 0 {
		components = loadSiteComponentFunctionsForJS(ctx, params.DB, siteUUID, params.Logger)
		params.Logger.Info("RenderJSSnippetsForSiteAction: components loaded from DB",
			zap.Int("count", len(components)))
	}

	// 4. Empty bundle if no components — file still exists on CDN so the
	//    head's <script src> doesn't 404.
	if len(components) == 0 {
		js := emptyJSSnippetsBundle(siteIDStr)
		return map[string]interface{}{
			"files":         map[string]interface{}{"assets/js/snippets.js": js},
			"domain":        domain,
			"js_content":    js,
			"snippet_count": 0,
			"snippet_names": []string{},
		}, nil
	}

	// 5. Load active snippets that apply
	snippets, err := loadJSSnippetsForSite(ctx, params.DB, components, params.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to load JS snippets: %w", err)
	}

	// 6. Bundle
	js := buildJSSnippetsBundle(siteIDStr, snippets)
	names := make([]string, 0, len(snippets))
	for _, s := range snippets {
		names = append(names, s.Name)
	}

	params.Logger.Info("RenderJSSnippetsForSiteAction: Complete",
		zap.String("site_id", siteIDStr),
		zap.String("domain", domain),
		zap.Int("component_count", len(components)),
		zap.Int("snippet_count", len(snippets)),
		zap.Strings("snippet_names", names),
		zap.Int("js_length", len(js)))

	return map[string]interface{}{
		"files":         map[string]interface{}{"assets/js/snippets.js": js},
		"domain":        domain,
		"js_content":    js,
		"snippet_count": len(snippets),
		"snippet_names": names,
	}, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

type jsSnippetRow struct {
	Name        string
	Description string
	JSContent   string
}

// loadJSSnippetsForSite queries js_snippets where applies_to overlaps the
// site's component list AND is_active = true. Ordered by name for
// deterministic bundle output (so re-renders produce identical files
// when nothing changes — no spurious git commits).
func loadJSSnippetsForSite(ctx context.Context, db *sql.DB, components []string, logger *zap.Logger) ([]jsSnippetRow, error) {
	jsonBytes, err := json.Marshal(components)
	if err != nil {
		return nil, fmt.Errorf("marshal components: %w", err)
	}

	rows, err := db.QueryContext(ctx, `
		SELECT name, COALESCE(description, ''), js_content
		FROM js_snippets
		WHERE is_active = true
		  AND applies_to && $1::jsonb
		ORDER BY name
	`, string(jsonBytes))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snippets []jsSnippetRow
	for rows.Next() {
		var s jsSnippetRow
		if err := rows.Scan(&s.Name, &s.Description, &s.JSContent); err != nil {
			logger.Warn("loadJSSnippetsForSite: row scan failed", zap.Error(err))
			continue
		}
		snippets = append(snippets, s)
	}
	return snippets, nil
}

// extractComponentFunctionsList pulls a string list from collected data.
// Tolerates []interface{} (the usual case) and []string.
func extractComponentFunctionsList(collectedData map[string]interface{}, field string, logger *zap.Logger) []string {
	val := datahelpers.ExtractNestedField(collectedData, field)
	if val == nil {
		return nil
	}
	switch v := val.(type) {
	case []interface{}:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	case []string:
		return v
	}
	return nil
}

// loadSiteComponentFunctionsForJS queries distinct content_components.function
// values used by any page or site-level component on the given site.
// Used as a fallback when component list isn't in collected data.
func loadSiteComponentFunctionsForJS(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) []string {
	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT cc.function
		FROM page_components pc
		JOIN pages p ON p.id = pc.page_id
		JOIN content_components cc ON cc.id = pc.component_id
		WHERE p.site_id = $1
		  AND cc.function IS NOT NULL
		  AND cc.function != ''
		UNION
		SELECT DISTINCT cc.function
		FROM site_components sc
		JOIN content_components cc ON cc.id = sc.component_id
		WHERE sc.site_id = $1
		  AND cc.function IS NOT NULL
		  AND cc.function != ''
	`, siteID)
	if err != nil {
		logger.Warn("loadSiteComponentFunctionsForJS: query failed", zap.Error(err))
		return nil
	}
	defer rows.Close()

	var functions []string
	for rows.Next() {
		var f string
		if err := rows.Scan(&f); err == nil && f != "" {
			functions = append(functions, f)
		}
	}
	return functions
}

func buildJSSnippetsBundle(siteID string, snippets []jsSnippetRow) string {
	var sb strings.Builder
	sb.WriteString("/* ============================================================\n")
	sb.WriteString(fmt.Sprintf(" * JS snippets bundle for site %s\n", siteID))
	sb.WriteString(fmt.Sprintf(" * %d active snippet(s)\n", len(snippets)))
	sb.WriteString(" * Generated by render_js_snippets_for_site action.\n")
	sb.WriteString(" * Do not edit directly — update js_snippets rows in DB.\n")
	sb.WriteString(" * ============================================================ */\n\n")

	for _, s := range snippets {
		sb.WriteString(fmt.Sprintf("/* --- %s", s.Name))
		if s.Description != "" {
			sb.WriteString(fmt.Sprintf(" — %s", s.Description))
		}
		sb.WriteString(" --- */\n")
		sb.WriteString(s.JSContent)
		if !strings.HasSuffix(s.JSContent, "\n") {
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func emptyJSSnippetsBundle(siteID string) string {
	return fmt.Sprintf(
		"/* JS snippets bundle for site %s — no active snippets. */\n",
		siteID)
}
