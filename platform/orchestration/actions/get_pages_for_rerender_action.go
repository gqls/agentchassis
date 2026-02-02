// FILE: platform/orchestration/actions/get_pages_for_rerender_action.go
// GetPagesForRerenderAction returns page metadata for rerender loop
// Returns small payload (no HTML) suitable for loop iteration
//
// Uses input_fields pattern - caller passes site_id/domain via input_mapping,
// action uses ExtractFields to get them regardless of structure

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

// GetPagesForRerenderAction queries pages table and returns metadata for loop
//
// Config:
//   - input_fields: array of field names to extract (default: ["site_id", "domain"])
//   - include_statuses: array of page statuses (default: ["deployed", "active"])
//
// The caller provides site_id and/or domain via input_mapping.
// ExtractFields handles finding them regardless of input_data prefix.
//
// Returns:
//   - pages: array of {page_id, name, url, filename, title}
//   - site_id: string
//   - domain: string
//   - page_count: int
func GetPagesForRerenderAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("GetPagesForRerenderAction: Starting",
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	config := params.StepConfig.Config

	// Use input_fields pattern with sensible defaults
	inputFields := []string{"site_id", "domain"}
	if fields, ok := config["input_fields"].([]interface{}); ok {
		inputFields = make([]string, len(fields))
		for i, f := range fields {
			inputFields[i], _ = f.(string)
		}
	}

	// ExtractFields handles checking both CollectedData and CollectedData["input_data"]
	extracted := datahelpers.ExtractFields(params.CollectedData, inputFields, params.Logger)

	params.Logger.Debug("GetPagesForRerenderAction: Extracted fields",
		zap.Any("extracted", extracted),
	)

	// Get site_id
	var siteID uuid.UUID
	var err error
	if siteIDStr, ok := extracted["site_id"].(string); ok && siteIDStr != "" {
		siteID, err = uuid.Parse(siteIDStr)
		if err != nil {
			params.Logger.Warn("GetPagesForRerenderAction: Invalid site_id format",
				zap.String("site_id", siteIDStr),
				zap.Error(err),
			)
		}
	}

	// Get domain
	domain, _ := extracted["domain"].(string)

	// Fallback: if we have site_id but no domain, look it up
	if siteID != uuid.Nil && domain == "" {
		domain, _ = getDomainForSite(ctx, params.DB, siteID)
	}

	// Fallback: if we have domain but no site_id, look it up
	if siteID == uuid.Nil && domain != "" {
		siteID, err = lookupSiteByDomain(ctx, params.DB, domain)
		if err != nil {
			return nil, fmt.Errorf("failed to lookup site by domain %s: %w", domain, err)
		}
	}

	if siteID == uuid.Nil {
		return nil, fmt.Errorf("no valid site_id or domain provided - check input_mapping in caller")
	}

	// Get statuses to include
	statuses := []string{"deployed", "active"}
	if statusList, ok := config["include_statuses"].([]interface{}); ok {
		statuses = make([]string, len(statusList))
		for i, s := range statusList {
			statuses[i], _ = s.(string)
		}
	}

	// Query pages - metadata only
	pages, err := queryPagesMetadata(ctx, params.DB, siteID, statuses, params.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to query pages: %w", err)
	}

	params.Logger.Info("GetPagesForRerenderAction: Complete",
		zap.String("site_id", siteID.String()),
		zap.String("domain", domain),
		zap.Int("page_count", len(pages)),
	)

	return map[string]interface{}{
		"success":    true,
		"pages":      pages,
		"site_id":    siteID.String(),
		"domain":     domain,
		"page_count": len(pages),
	}, nil
}

// PageMetadata holds minimal page info for loop iteration
type PageMetadata struct {
	PageID   string `json:"page_id"`
	Name     string `json:"name"`
	Title    string `json:"title"`
	URL      string `json:"url"`
	Filename string `json:"filename"`
}

func queryPagesMetadata(ctx context.Context, db *sql.DB, siteID uuid.UUID, statuses []string, logger *zap.Logger) ([]PageMetadata, error) {
	var queryBuilder strings.Builder
	queryBuilder.WriteString(`
		SELECT p.id, p.name, COALESCE(p.title, p.name) as title, p.url
		FROM pages p
		WHERE p.site_id = $1
	`)

	args := []interface{}{siteID}
	argIndex := 2

	if len(statuses) > 0 {
		placeholders := make([]string, len(statuses))
		for i, s := range statuses {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, s)
			argIndex++
		}
		queryBuilder.WriteString(fmt.Sprintf(" AND p.status IN (%s)", strings.Join(placeholders, ",")))
	}

	queryBuilder.WriteString(" ORDER BY p.nav_order, p.created_at")

	rows, err := db.QueryContext(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pages []PageMetadata
	for rows.Next() {
		var p PageMetadata
		var url string
		if err := rows.Scan(&p.PageID, &p.Name, &p.Title, &url); err != nil {
			logger.Warn("queryPagesMetadata: Scan error", zap.Error(err))
			continue
		}

		p.URL = url

		// Derive filename from URL
		if url == "/" || url == "" || p.Name == "index" {
			p.Filename = "index.html"
		} else {
			p.Filename = strings.TrimPrefix(url, "/")
			if !strings.HasSuffix(p.Filename, ".html") {
				p.Filename = p.Filename + ".html"
			}
		}

		pages = append(pages, p)
	}

	return pages, nil
}

func lookupSiteByDomain(ctx context.Context, db *sql.DB, domain string) (uuid.UUID, error) {
	var siteID uuid.UUID
	err := db.QueryRowContext(ctx, `SELECT id FROM sites WHERE domain = $1`, domain).Scan(&siteID)
	return siteID, err
}
