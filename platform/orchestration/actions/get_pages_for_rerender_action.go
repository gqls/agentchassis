// FILE: platform/orchestration/actions/get_pages_for_rerender_action.go
// GetPagesForRerenderAction returns page metadata for rerender loop
// Returns small payload (no HTML) suitable for loop iteration

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
// Config:
//   - site_id_field: path to site_id (default: "input_data.site_id")
//   - domain_field: path to domain (default: "input_data.domain")
//   - include_statuses: array of page statuses (default: ["deployed", "active"])
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

	// Get site_id
	siteIDField := "input_data.site_id"
	if f, ok := config["site_id_field"].(string); ok && f != "" {
		siteIDField = f
	}
	siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)

	var siteID uuid.UUID
	var err error
	var domain string

	if siteIDStr != "" {
		siteID, err = uuid.Parse(siteIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid site_id: %w", err)
		}
	}

	// Fallback to domain lookup
	if siteID == uuid.Nil {
		domainField := "input_data.domain"
		if f, ok := config["domain_field"].(string); ok && f != "" {
			domainField = f
		}
		domain = datahelpers.ExtractNestedFieldString(params.CollectedData, domainField)
		if domain != "" {
			siteID, err = lookupSiteByDomain(ctx, params.DB, domain)
			if err != nil {
				return nil, fmt.Errorf("failed to lookup site by domain %s: %w", domain, err)
			}
		}
	}

	if siteID == uuid.Nil {
		return nil, fmt.Errorf("no valid site_id or domain provided")
	}

	// Get domain if we don't have it
	if domain == "" {
		domain, _ = getDomainForSite(ctx, params.DB, siteID)
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

func getDomainForSite(ctx context.Context, db *sql.DB, siteID uuid.UUID) (string, error) {
	var domain string
	err := db.QueryRowContext(ctx, `SELECT domain FROM sites WHERE id = $1`, siteID).Scan(&domain)
	return domain, err
}
