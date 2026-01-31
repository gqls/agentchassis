// FILE: platform/orchestration/actions/rerender_single_page_action.go
// RerenderSinglePageAction renders ONE page from page_components.rendered_html
// Used in rerender loop - returns HTML for git_commit

package actions

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// RerenderSinglePageAction renders a single page from stored sections
// Config:
//   - page_id_field: path to page_id (default: "current_page.page_id")
//   - site_id_field: path to site_id (default: "rerender_pages.site_id")
//   - domain_field: path to domain (default: "rerender_pages.domain")
//   - max_nav_items: int (default: 6)
//
// Returns:
//   - html: rendered page HTML
//   - domain: site domain
//   - filename: page filename (e.g., "about.html")
//   - page_id: the page ID
//   - page_name: the page name
func RerenderSinglePageAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("RerenderSinglePageAction: Starting",
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	config := params.StepConfig.Config

	// Extract page_id
	pageIDField := "current_page.page_id"
	if f, ok := config["page_id_field"].(string); ok && f != "" {
		pageIDField = f
	}
	pageIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, pageIDField)
	if pageIDStr == "" {
		return nil, fmt.Errorf("page_id not found at %s", pageIDField)
	}

	pageID, err := uuid.Parse(pageIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid page_id: %w", err)
	}

	// Extract site_id
	siteIDField := "rerender_pages.site_id"
	if f, ok := config["site_id_field"].(string); ok && f != "" {
		siteIDField = f
	}
	siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
	if siteIDStr == "" {
		return nil, fmt.Errorf("site_id not found at %s", siteIDField)
	}

	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// Extract domain
	domainField := "rerender_pages.domain"
	if f, ok := config["domain_field"].(string); ok && f != "" {
		domainField = f
	}
	domain := datahelpers.ExtractNestedFieldString(params.CollectedData, domainField)
	if domain == "" {
		domain, _ = getDomainForSite(ctx, params.DB, siteID)
	}

	// Get max nav items
	maxNavItems := 6
	if m, ok := config["max_nav_items"].(float64); ok {
		maxNavItems = int(m)
	}

	// Load page info
	pageInfo, err := loadPageInfo(ctx, params.DB, pageID)
	if err != nil {
		return nil, fmt.Errorf("failed to load page info: %w", err)
	}

	params.Logger.Info("RerenderSinglePageAction: Rendering page",
		zap.String("page_id", pageIDStr),
		zap.String("page_name", pageInfo.Name),
		zap.String("domain", domain),
	)

	// Load sections from page_components
	sectionsHTML := loadPageSections(ctx, params.DB, pageID, params.Logger)
	if sectionsHTML == "" {
		return nil, fmt.Errorf("no sections found for page %s", pageInfo.Name)
	}

	// Build render context
	renderCtx := &RenderContext{
		ContentData: map[string]interface{}{
			"Title":           pageInfo.Title,
			"MetaDescription": pageInfo.MetaDesc,
			"PageName":        pageInfo.Name,
		},
		NavItems:       []NavItem{},
		FooterNavItems: []NavItem{},
	}

	// Get nav from deployed pages
	headerNav := getHeaderNavFromDB(ctx, params.DB, siteID, maxNavItems, params.Logger)
	footerNav := getFooterNavFromDB(ctx, params.DB, siteID, 10, params.Logger)
	if len(headerNav) > 0 {
		renderCtx.NavItems = headerNav
	}
	if len(footerNav) > 0 {
		renderCtx.FooterNavItems = footerNav
	}

	// Get site contact info
	siteInfo := getSiteContactInfo(ctx, params.DB, siteID, params.Logger)

	// Load templates
	headHTML := loadHeadTemplate(ctx, params.DB, siteID, renderCtx, params.Logger)
	headerHTML := loadHeaderTemplate(ctx, params.DB, siteID, renderCtx, params.Logger)
	footerHTML := loadFooterTemplate(ctx, params.DB, siteID, renderCtx, params.Logger)

	// Render head with page-specific data
	renderedHead := executeSimpleTemplate(headHTML, renderCtx.ContentData, params.Logger)

	// Assemble full page
	var html strings.Builder
	html.WriteString("<!DOCTYPE html>\n<html lang=\"en\">\n")
	html.WriteString(renderedHead)
	html.WriteString("\n<body>\n")

	if headerHTML != "" {
		html.WriteString(headerHTML)
		html.WriteString("\n")
	}

	html.WriteString("<main>\n")
	html.WriteString(sectionsHTML)
	html.WriteString("\n</main>\n")

	if footerHTML != "" {
		html.WriteString(footerHTML)
		html.WriteString("\n")
	}

	html.WriteString("</body>\n</html>")

	// Inject correct contact info
	finalHTML := injectContactInfo(html.String(), siteInfo, params.Logger)

	params.Logger.Info("RerenderSinglePageAction: Complete",
		zap.String("page_name", pageInfo.Name),
		zap.Int("html_length", len(finalHTML)),
	)

	return map[string]interface{}{
		"success":   true,
		"html":      finalHTML,
		"domain":    domain,
		"filename":  pageInfo.Filename,
		"page_id":   pageIDStr,
		"page_name": pageInfo.Name,
	}, nil
}

// PageInfo holds page details for rendering
type PageInfo struct {
	ID       uuid.UUID
	Name     string
	Title    string
	URL      string
	Filename string
	MetaDesc string
}

// SiteContactInfo holds site contact details
type SiteContactInfo struct {
	Email string
	Phone string
}

func loadPageInfo(ctx context.Context, db *sql.DB, pageID uuid.UUID) (*PageInfo, error) {
	var p PageInfo
	p.ID = pageID

	err := db.QueryRowContext(ctx, `
		SELECT name, COALESCE(title, name), url, COALESCE(meta_description, '')
		FROM pages WHERE id = $1
	`, pageID).Scan(&p.Name, &p.Title, &p.URL, &p.MetaDesc)

	if err != nil {
		return nil, err
	}

	// Derive filename
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

func loadPageSections(ctx context.Context, db *sql.DB, pageID uuid.UUID, logger *zap.Logger) string {
	rows, err := db.QueryContext(ctx, `
		SELECT COALESCE(rendered_html, '') as html
		FROM page_components
		WHERE page_id = $1
		ORDER BY position
	`, pageID)
	if err != nil {
		logger.Warn("loadPageSections: Query failed", zap.Error(err))
		return ""
	}
	defer rows.Close()

	var sections []string
	for rows.Next() {
		var html string
		if err := rows.Scan(&html); err != nil {
			continue
		}
		if html != "" {
			html = stripPageWrapper(html)
			if html != "" {
				sections = append(sections, html)
			}
		}
	}

	logger.Debug("loadPageSections: Loaded",
		zap.String("page_id", pageID.String()),
		zap.Int("section_count", len(sections)),
	)

	return strings.Join(sections, "\n")
}

func getHeaderNavFromDB(ctx context.Context, db *sql.DB, siteID uuid.UUID, maxItems int, logger *zap.Logger) []NavItem {
	if db == nil || siteID == uuid.Nil {
		return nil
	}

	query := `
		SELECT 
			COALESCE(nav_label, title, name) as label,
			COALESCE(url, '/' || name || '.html') as url,
			COALESCE(nav_order, 0) as nav_order
		FROM pages 
		WHERE site_id = $1 
		  AND in_header = true
		  AND status IN ('deployed', 'active')
		  AND deleted_at IS NULL
		ORDER BY nav_order ASC, created_at ASC
		LIMIT $2
	`

	rows, err := db.QueryContext(ctx, query, siteID, maxItems)
	if err != nil {
		logger.Warn("getHeaderNavFromDB: Query failed", zap.Error(err))
		return nil
	}
	defer rows.Close()

	var items []NavItem
	for rows.Next() {
		var label, url string
		var navOrder int
		if err := rows.Scan(&label, &url, &navOrder); err != nil {
			continue
		}
		label = simplifyNavLabel(label, url)
		items = append(items, NavItem{Label: label, URL: url})
	}

	return items
}

func getFooterNavFromDB(ctx context.Context, db *sql.DB, siteID uuid.UUID, maxItems int, logger *zap.Logger) []NavItem {
	if db == nil || siteID == uuid.Nil {
		return nil
	}

	query := `
		SELECT 
			COALESCE(nav_label, title, name) as label,
			COALESCE(url, '/' || name || '.html') as url
		FROM pages 
		WHERE site_id = $1 
		  AND in_footer = true
		  AND status IN ('deployed', 'active')
		  AND deleted_at IS NULL
		ORDER BY nav_order ASC, created_at ASC
		LIMIT $2
	`

	rows, err := db.QueryContext(ctx, query, siteID, maxItems)
	if err != nil {
		logger.Warn("getFooterNavFromDB: Query failed", zap.Error(err))
		return nil
	}
	defer rows.Close()

	var items []NavItem
	for rows.Next() {
		var label, url string
		if err := rows.Scan(&label, &url); err != nil {
			continue
		}
		label = simplifyNavLabel(label, url)
		items = append(items, NavItem{Label: label, URL: url})
	}

	return items
}

func simplifyNavLabel(label, url string) string {
	if idx := strings.Index(label, " | "); idx > 0 {
		label = strings.TrimSpace(label[:idx])
	}
	if idx := strings.Index(label, " - "); idx > 0 {
		label = strings.TrimSpace(label[:idx])
	}

	if len(label) > 20 {
		urlParts := strings.Split(strings.Trim(url, "/"), "/")
		if len(urlParts) > 0 {
			pageName := strings.TrimSuffix(urlParts[len(urlParts)-1], ".html")
			if pageName != "" && pageName != "index" {
				label = strings.Title(strings.ReplaceAll(pageName, "-", " "))
			}
		}
	}
	return label
}

func getSiteContactInfo(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) SiteContactInfo {
	if db == nil {
		return SiteContactInfo{}
	}

	var info SiteContactInfo
	err := db.QueryRowContext(ctx, `
		SELECT COALESCE(email, ''), COALESCE(phone, '')
		FROM sites WHERE id = $1
	`, siteID).Scan(&info.Email, &info.Phone)

	if err != nil {
		logger.Warn("getSiteContactInfo: Query failed", zap.Error(err))
	}
	return info
}

func loadHeadTemplate(ctx context.Context, db *sql.DB, siteID uuid.UUID, renderCtx *RenderContext, logger *zap.Logger) string {
	if db == nil {
		return fallbackHead()
	}

	var headComponent string
	err := db.QueryRowContext(ctx, `
		SELECT COALESCE(content_data->>'head', 'head-seo-standard')
		FROM sites WHERE id = $1
	`, siteID).Scan(&headComponent)
	if err != nil {
		return fallbackHead()
	}

	var htmlTemplate string
	err = db.QueryRowContext(ctx, `
		SELECT html_template FROM content_components 
		WHERE name = $1 OR function = $1
		LIMIT 1
	`, headComponent).Scan(&htmlTemplate)
	if err != nil {
		return fallbackHead()
	}

	return RenderTemplate(htmlTemplate, renderCtx, logger)
}

func fallbackHead() string {
	return `<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <link rel="stylesheet" href="/assets/css/styles.css">
</head>`
}

func loadHeaderTemplate(ctx context.Context, db *sql.DB, siteID uuid.UUID, renderCtx *RenderContext, logger *zap.Logger) string {
	if db == nil {
		return ""
	}

	var headerComponent string
	err := db.QueryRowContext(ctx, `
		SELECT COALESCE(content_data->>'header', 'header-standard')
		FROM sites WHERE id = $1
	`, siteID).Scan(&headerComponent)
	if err != nil {
		return ""
	}

	var htmlTemplate string
	err = db.QueryRowContext(ctx, `
		SELECT html_template FROM content_components 
		WHERE name = $1 OR function = $1
		LIMIT 1
	`, headerComponent).Scan(&htmlTemplate)
	if err != nil {
		return ""
	}

	return RenderTemplate(htmlTemplate, renderCtx, logger)
}

func loadFooterTemplate(ctx context.Context, db *sql.DB, siteID uuid.UUID, renderCtx *RenderContext, logger *zap.Logger) string {
	if db == nil {
		return ""
	}

	var footerComponent string
	err := db.QueryRowContext(ctx, `
		SELECT COALESCE(content_data->>'footer', 'footer-standard')
		FROM sites WHERE id = $1
	`, siteID).Scan(&footerComponent)
	if err != nil {
		return ""
	}

	var htmlTemplate string
	err = db.QueryRowContext(ctx, `
		SELECT html_template FROM content_components 
		WHERE name = $1 OR function = $1
		LIMIT 1
	`, footerComponent).Scan(&htmlTemplate)
	if err != nil {
		return ""
	}

	return RenderTemplate(htmlTemplate, renderCtx, logger)
}

func stripPageWrapper(html string) string {
	html = regexp.MustCompile(`(?i)<!DOCTYPE[^>]*>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`(?i)</?html[^>]*>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`(?is)<head>.*?</head>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`(?i)<body[^>]*>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`(?i)</body>`).ReplaceAllString(html, "")
	return strings.TrimSpace(html)
}

func injectContactInfo(html string, siteInfo SiteContactInfo, logger *zap.Logger) string {
	if siteInfo.Email == "" && siteInfo.Phone == "" {
		return html
	}

	if siteInfo.Email != "" {
		// Replace mailto: links
		emailPattern := regexp.MustCompile(`mailto:([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})`)
		html = emailPattern.ReplaceAllString(html, "mailto:"+siteInfo.Email)

		// Replace emails in contact sections
		contactSectionPattern := regexp.MustCompile(`(?is)(<section[^>]*(?:contact|footer)[^>]*>.*?)([\w.+-]+@[\w.-]+\.[a-z]{2,})(.*?</section>)`)
		html = contactSectionPattern.ReplaceAllStringFunc(html, func(match string) string {
			parts := contactSectionPattern.FindStringSubmatch(match)
			if len(parts) == 4 {
				return parts[1] + siteInfo.Email + parts[3]
			}
			return match
		})
	}

	if siteInfo.Phone != "" {
		phonePattern := regexp.MustCompile(`tel:([+\d\s()-]+)`)
		cleanPhone := regexp.MustCompile(`[^\d+]`).ReplaceAllString(siteInfo.Phone, "")
		html = phonePattern.ReplaceAllString(html, "tel:"+cleanPhone)
	}

	return html
}

func executeSimpleTemplate(tmpl string, data map[string]interface{}, logger *zap.Logger) string {
	if tmpl == "" {
		return ""
	}

	t, err := template.New("page").Parse(tmpl)
	if err != nil {
		logger.Warn("executeSimpleTemplate: Parse error", zap.Error(err))
		return tmpl
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		logger.Warn("executeSimpleTemplate: Execute error", zap.Error(err))
		return tmpl
	}

	return buf.String()
}
