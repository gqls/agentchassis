// FILE: platform/orchestration/actions/load_page_record_action.go
//
// Loads a page record from the pages table by site_id + page_name.
// Replaces inline query_database SQL in workflows — column names are
// compiled Go, not fragile JSON strings.
//
// Handles:
//   - Page found: returns full page record with parsed sections
//   - Page not found: returns {found: false} (not an error)
//   - Page name is "new page needed" / "site-wide" / empty: returns {found: false}
//
// Registration:
//   "load_page_record": {
//       Handler:     LoadPageRecordAction,
//       Category:    "site",
//       Description: "Load a page record by site_id and page name",
//       IsLocal:     true,
//   },
//
// Workflow config example:
//
//   "load_page_record": {
//       "action": "load_page_record",
//       "config": {
//           "site_id": "site_record.site_id",
//           "page_name": "input_data.page_name"
//       },
//       "next_step": "spawn_content_writer",
//       "output_field": "page_record"
//   }

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

var LoadPageRecordInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id", "page_name"},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("load_page_record", LoadPageRecordInputSpec)
}

// Names that indicate the audit finding is about a new page or site-wide issue,
// not an existing page in the database.
var nonPageNames = map[string]bool{
	"new page needed": true,
	"site-wide":       true,
	"new page":        true,
	"general":         true,
	"":                true,
}

func LoadPageRecordAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "load_page_record"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		LoadPageRecordInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id %q: %w", siteIDStr, err)
	}

	pageName := strings.TrimSpace(inputs.Get("page_name"))

	// Check for non-page names (audit findings that describe gaps, not existing pages)
	if nonPageNames[strings.ToLower(pageName)] {
		logger.Info("LoadPageRecordAction: page_name is not a real page",
			zap.String("page_name", pageName))
		return map[string]interface{}{
			"found":     false,
			"page_name": pageName,
			"reason":    "page_name is a description, not an existing page",
		}, nil
	}

	// Query the page
	var (
		pageID       string
		name         string
		title        sql.NullString
		pageType     sql.NullString
		sectionsJSON sql.NullString
		url          sql.NullString
		buildStatus  sql.NullString
		navLabel     sql.NullString
		navOrder     sql.NullInt32
	)

	err = params.DB.QueryRowContext(ctx, `
		SELECT id::text, name, title, page_type, sections::text, url, build_status, nav_label, nav_order
		FROM pages
		WHERE site_id = $1 AND name = $2
		LIMIT 1
	`, siteID, pageName).Scan(
		&pageID, &name, &title, &pageType,
		&sectionsJSON, &url, &buildStatus, &navLabel, &navOrder,
	)

	if err == sql.ErrNoRows {
		logger.Info("LoadPageRecordAction: page not found",
			zap.String("site_id", siteID.String()),
			zap.String("page_name", pageName))
		return map[string]interface{}{
			"found":     false,
			"page_name": pageName,
			"site_id":   siteID.String(),
			"reason":    "page not found in database",
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query page: %w", err)
	}

	// Parse sections JSON into an array
	var sections []interface{}
	if sectionsJSON.Valid && sectionsJSON.String != "" {
		if err := json.Unmarshal([]byte(sectionsJSON.String), &sections); err != nil {
			logger.Warn("LoadPageRecordAction: failed to parse sections JSON",
				zap.String("sections_raw", sectionsJSON.String), zap.Error(err))
			sections = []interface{}{}
		}
	}

	result := map[string]interface{}{
		"found":         true,
		"id":            pageID,
		"name":          name,
		"site_id":       siteID.String(),
		"sections":      sections,
		"section_count": len(sections),
	}

	if title.Valid {
		result["title"] = title.String
	}
	if pageType.Valid {
		result["page_type"] = pageType.String
	}
	if url.Valid {
		result["url"] = url.String
	}
	if buildStatus.Valid {
		result["build_status"] = buildStatus.String
	}
	if navLabel.Valid {
		result["nav_label"] = navLabel.String
	}
	if navOrder.Valid {
		result["nav_order"] = navOrder.Int32
	}

	logger.Info("LoadPageRecordAction: page loaded",
		zap.String("page_id", pageID),
		zap.String("page_name", name),
		zap.String("page_type", pageType.String),
		zap.Int("sections", len(sections)),
	)

	return result, nil
}
