// FILE: platform/orchestration/actions/rerender_single_page_action.go
// RerenderSinglePageAction assembles a page from stored components
// Uses site_components for header/footer, page_components for sections
// Simple concatenation - no template re-rendering

package actions

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// RerenderSinglePageAction assembles a page from pre-rendered components
//
// Config:
//   - input_fields: fields to extract (default: ["page_id", "site_id", "domain"])
//
// Returns:
//   - html: assembled page HTML
//   - domain: site domain
//   - filename: page filename (e.g., "about.html")
//   - page_id: the page ID
//   - page_name: the page name
//   - skipped: true if page had no sections
func RerenderSinglePageAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("RerenderSinglePageAction: Starting")

	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	config := params.StepConfig.Config

	// Extract input fields
	inputFields := []string{"page_id", "site_id", "domain"}
	if fields, ok := config["input_fields"].([]interface{}); ok {
		inputFields = make([]string, len(fields))
		for i, f := range fields {
			inputFields[i], _ = f.(string)
		}
	}

	extracted := datahelpers.ExtractFields(params.CollectedData, inputFields, params.Logger)

	// Get page_id
	pageIDStr, _ := extracted["page_id"].(string)
	if pageIDStr == "" {
		return nil, fmt.Errorf("page_id not found in input")
	}
	pageID, err := uuid.Parse(pageIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid page_id: %w", err)
	}

	// Get page info (includes site_id, area_id)
	pageInfo, err := getPageInfo(ctx, params.DB, pageID)
	if err != nil {
		return nil, fmt.Errorf("failed to load page info: %w", err)
	}

	params.Logger.Info("RerenderSinglePageAction: Assembling page",
		zap.String("page_id", pageIDStr),
		zap.String("page_name", pageInfo.Name),
		zap.String("domain", pageInfo.Domain),
	)

	// Assemble the page from stored components
	html, err := assemblePage(ctx, params.DB, pageInfo, params.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to assemble page: %w", err)
	}

	if html == "" {
		params.Logger.Warn("RerenderSinglePageAction: No content, skipping",
			zap.String("page_name", pageInfo.Name))
		return map[string]interface{}{
			"success":   false,
			"skipped":   true,
			"reason":    "no components found for page",
			"html":      "",
			"domain":    pageInfo.Domain,
			"filename":  pageInfo.Filename,
			"page_id":   pageIDStr,
			"page_name": pageInfo.Name,
		}, nil
	}

	params.Logger.Info("RerenderSinglePageAction: Complete",
		zap.String("page_name", pageInfo.Name),
		zap.Int("html_length", len(html)),
	)

	return map[string]interface{}{
		"success":   true,
		"html":      html,
		"domain":    pageInfo.Domain,
		"filename":  pageInfo.Filename,
		"page_id":   pageIDStr,
		"page_name": pageInfo.Name,
	}, nil
}

// getPageInfo loads page metadata including site and area
func getPageInfo(ctx context.Context, db *sql.DB, pageID uuid.UUID) (*PageInfo, error) {
	var p PageInfo
	var areaID sql.NullString

	err := db.QueryRowContext(ctx, `
		SELECT 
			p.id, p.site_id, p.site_area_id, 
			p.name, COALESCE(p.title, p.name), p.url,
			s.domain
		FROM pages p
		JOIN sites s ON p.site_id = s.id
		WHERE p.id = $1
	`, pageID).Scan(
		&p.ID, &p.SiteID, &areaID,
		&p.Name, &p.Title, &p.URL,
		&p.Domain,
	)
	if err != nil {
		return nil, err
	}

	if areaID.Valid {
		id, _ := uuid.Parse(areaID.String)
		p.AreaID = &id
	}

	// Derive filename from URL
	if p.URL == "/" || p.URL == "" || p.Name == "index" {
		p.Filename = "index.html"
	} else {
		p.Filename = strings.TrimPrefix(p.URL, "/")
		if !strings.HasSuffix(p.Filename, ".html") {
			p.Filename = p.Filename + ".html"
		}
	}

	return &p, nil
}

// assemblePage combines site/area/page components into full HTML
func assemblePage(ctx context.Context, db *sql.DB, page *PageInfo, logger *zap.Logger) (string, error) {
	// 1. Get site-level components
	siteComponents, err := getSiteComponents(ctx, db, page.SiteID)
	if err != nil {
		logger.Warn("Failed to load site components", zap.Error(err))
	}

	// 2. Get area-level overrides (if page is in an area)
	areaComponents := map[string]string{}
	if page.AreaID != nil {
		areaComponents, err = getAreaComponents(ctx, db, *page.AreaID)
		if err != nil {
			logger.Warn("Failed to load area components", zap.Error(err))
		}
	}

	// 3. Get page sections
	sections, err := getPageSections(ctx, db, page.ID)
	if err != nil {
		logger.Warn("Failed to load page sections", zap.Error(err))
	}

	// No content at all?
	if len(siteComponents) == 0 && len(sections) == 0 {
		return "", nil
	}

	// 4. Resolve components (area overrides site)
	head := resolveComponent(areaComponents, siteComponents, "head")
	header := resolveComponent(areaComponents, siteComponents, "header")
	footer := resolveComponent(areaComponents, siteComponents, "footer")

	// 5. Build page-specific head if we don't have one stored
	if head == "" {
		head = buildDefaultHead(page)
	} else {
		// Inject page-specific title into stored head component
		// The site-level head has <title></title> — replace with this page's title
		if page.Title != "" {
			titleRe := regexp.MustCompile(`<title>[^<]*</title>`)
			head = titleRe.ReplaceAllString(head, fmt.Sprintf("<title>%s</title>", page.Title))
		}
		// Inject meta description if the page has one and the head has an empty content=""
		if page.MetaDesc != "" {
			head = strings.Replace(head,
				`content="">`,
				fmt.Sprintf(`content="%s">`, page.MetaDesc), 1)
		}
	}

	// 6. Assemble
	var html strings.Builder
	html.WriteString("<!DOCTYPE html>\n<html lang=\"en\">\n")
	html.WriteString(head)
	html.WriteString("\n<body>\n")

	if header != "" {
		html.WriteString(header)
		html.WriteString("\n")
	}

	html.WriteString("<main>\n")
	html.WriteString(sections)
	html.WriteString("\n</main>\n")

	if footer != "" {
		html.WriteString(footer)
		html.WriteString("\n")
	}

	html.WriteString("</body>\n</html>")

	logger.Debug("assemblePage: Complete",
		zap.String("page", page.Name),
		zap.Bool("has_header", header != ""),
		zap.Bool("has_footer", footer != ""),
		zap.Int("sections_length", len(sections)),
	)

	return html.String(), nil
}

// getSiteComponents loads site-level components (header, footer, head)
func getSiteComponents(ctx context.Context, db *sql.DB, siteID uuid.UUID) (map[string]string, error) {
	components := make(map[string]string)

	rows, err := db.QueryContext(ctx, `
		SELECT slot_name, rendered_html 
		FROM site_components 
		WHERE site_id = $1 AND rendered_html IS NOT NULL AND rendered_html != ''
	`, siteID)
	if err != nil {
		return components, err
	}
	defer rows.Close()

	for rows.Next() {
		var slot, html string
		if err := rows.Scan(&slot, &html); err != nil {
			continue
		}
		components[slot] = html
	}

	return components, nil
}

// getAreaComponents loads area-level component overrides
func getAreaComponents(ctx context.Context, db *sql.DB, areaID uuid.UUID) (map[string]string, error) {
	components := make(map[string]string)

	rows, err := db.QueryContext(ctx, `
		SELECT slot_name, rendered_html 
		FROM area_components 
		WHERE area_id = $1 AND rendered_html IS NOT NULL AND rendered_html != ''
	`, areaID)
	if err != nil {
		return components, err
	}
	defer rows.Close()

	for rows.Next() {
		var slot, html string
		if err := rows.Scan(&slot, &html); err != nil {
			continue
		}
		components[slot] = html
	}

	return components, nil
}

// getPageSections loads and concatenates page sections in order
func getPageSections(ctx context.Context, db *sql.DB, pageID uuid.UUID) (string, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT COALESCE(rendered_html, '') 
		FROM page_components 
		WHERE page_id = $1 
		  AND rendered_html IS NOT NULL 
		  AND rendered_html != ''
		ORDER BY position ASC
	`, pageID)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var sections strings.Builder
	for rows.Next() {
		var html string
		if err := rows.Scan(&html); err != nil {
			continue
		}
		sections.WriteString(html)
		sections.WriteString("\n")
	}

	return sections.String(), nil
}

// resolveComponent returns area component if exists, otherwise site component
func resolveComponent(area, site map[string]string, slot string) string {
	if html, ok := area[slot]; ok && html != "" {
		return html
	}
	return site[slot]
}

// buildDefaultHead creates a basic head section for a page
// Used as fallback if no head stored in site_components
func buildDefaultHead(page *PageInfo) string {
	return fmt.Sprintf(`<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <link rel="stylesheet" href="/assets/css/styles.css">
</head>`, page.Title)
}
