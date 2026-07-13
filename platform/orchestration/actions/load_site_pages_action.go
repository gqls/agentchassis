// FILE: platform/orchestration/actions/load_site_pages_action.go
//
// Loads all active pages for a site. Used by planners and auditors
// that need the full page inventory for context.
//
// Returns a list of page summaries (not full content) plus
// convenience fields: page_count, page_names, has_blog.
//
// Registration:
//   "load_site_pages": {
//       Handler:     LoadSitePagesAction,
//       Category:    "site",
//       Description: "Load all pages for a site",
//       IsLocal:     true,
//   },
//
// Workflow config:
//   "load_existing_pages": {
//       "action": "load_site_pages",
//       "config": { "site_id": "site_record.site_id" },
//       "next_step": "plan_gaps",
//       "output_field": "existing_pages"
//   }

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var LoadSitePagesInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id"},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("load_site_pages", LoadSitePagesInputSpec)
}

func LoadSitePagesAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "load_site_pages"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		LoadSitePagesInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id %q: %w", siteIDStr, err)
	}

	rows, err := params.DB.QueryContext(ctx, `
		SELECT id::text, name, COALESCE(title, ''), COALESCE(page_type, 'content'),
		       COALESCE(sections::text, '[]'), COALESCE(url, ''),
		       COALESCE(build_status, 'pending'), COALESCE(nav_label, ''),
		       COALESCE(nav_order, 100), in_header, in_footer
		FROM pages
		WHERE site_id = $1 AND status = 'active'
		ORDER BY nav_order, name
	`, siteID)
	if err != nil {
		return nil, fmt.Errorf("query pages: %w", err)
	}
	defer rows.Close()

	var pages []map[string]interface{}
	var pageNames []string
	hasBlog := false

	for rows.Next() {
		var (
			pageID      string
			name        string
			title       string
			pageType    string
			sectionsRaw string
			url         string
			buildStatus string
			navLabel    string
			navOrder    int
			inHeader    sql.NullBool
			inFooter    sql.NullBool
		)

		if err := rows.Scan(&pageID, &name, &title, &pageType,
			&sectionsRaw, &url, &buildStatus, &navLabel,
			&navOrder, &inHeader, &inFooter); err != nil {
			logger.Warn("LoadSitePagesAction: failed to scan row", zap.Error(err))
			continue
		}

		// Parse sections
		var sections []interface{}
		if err := json.Unmarshal([]byte(sectionsRaw), &sections); err != nil {
			sections = []interface{}{}
		}

		page := map[string]interface{}{
			"id":            pageID,
			"name":          name,
			"title":         title,
			"page_type":     pageType,
			"sections":      sections,
			"section_count": len(sections),
			"url":           url,
			"build_status":  buildStatus,
			"nav_label":     navLabel,
			"nav_order":     navOrder,
			"in_header":     inHeader.Valid && inHeader.Bool,
			"in_footer":     inFooter.Valid && inFooter.Bool,
		}

		pages = append(pages, page)
		pageNames = append(pageNames, name)

		if pageType == "blog" || pageType == "blog-index" || name == "blog" {
			hasBlog = true
		}
	}

	logger.Info("LoadSitePagesAction: loaded pages",
		zap.String("site_id", siteIDStr),
		zap.Int("page_count", len(pages)),
		zap.Bool("has_blog", hasBlog))

	return map[string]interface{}{
		"pages":      pages,
		"page_names": pageNames,
		"page_count": len(pages),
		"has_blog":   hasBlog,
		"site_id":    siteIDStr,
	}, nil
}
