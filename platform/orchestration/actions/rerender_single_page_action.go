// FILE: platform/orchestration/actions/rerender_single_page_action.go
// RerenderSinglePageAction renders ONE page from page_components.rendered_html
// Used by page-rerender agent - returns HTML for git_commit

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
//
// Config:
//   - input_fields: fields to extract (default: ["page_id", "site_id", "domain"])
//   - max_nav_items: int (default: 6)
//
// Expects (via input_mapping from caller):
//   - page_id: the page to render
//   - site_id: site identifier
//   - domain: site domain
//
// Returns:
//   - html: rendered page HTML
//   - domain: site domain
//   - filename: page filename (e.g., "about.html")
//   - page_id: the page ID
//   - page_name: the page name
//   - skipped: true if page had no sections (caller should skip deploy)
func RerenderSinglePageAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("RerenderSinglePageAction: Starting",
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	config := params.StepConfig.Config

	// Use input_fields pattern - simple flat fields
	inputFields := []string{"page_id", "site_id", "domain"}
	if fields, ok := config["input_fields"].([]interface{}); ok {
		inputFields = make([]string, len(fields))
		for i, f := range fields {
			inputFields[i], _ = f.(string)
		}
	}

	// ExtractFields handles finding fields regardless of input_data prefix
	extracted := datahelpers.ExtractFields(params.CollectedData, inputFields, params.Logger)

	params.Logger.Debug("RerenderSinglePageAction: Extracted fields",
		zap.Any("fields", inputFields),
		zap.Any("extracted_keys", getMapKeys(extracted)),
	)

	// Get page_id - try direct string first, then nested object (backward compat)
	var pageIDStr string
	if s, ok := extracted["page_id"].(string); ok && s != "" {
		pageIDStr = s
	} else if currentPage, ok := extracted["current_page"].(map[string]interface{}); ok {
		// Backward compatibility with old workflow structure
		pageIDStr, _ = currentPage["page_id"].(string)
	}
	if pageIDStr == "" {
		return nil, fmt.Errorf("page_id not found in input")
	}

	pageID, err := uuid.Parse(pageIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid page_id: %w", err)
	}

	// Get site_id - try direct string first, then nested object
	var siteIDStr string
	if s, ok := extracted["site_id"].(string); ok && s != "" {
		siteIDStr = s
	} else if rerenderPages, ok := extracted["rerender_pages"].(map[string]interface{}); ok {
		// Backward compatibility
		siteIDStr, _ = rerenderPages["site_id"].(string)
	}
	if siteIDStr == "" {
		return nil, fmt.Errorf("site_id not found in input")
	}

	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// Get domain - try direct string first, then nested object
	var domain string
	if s, ok := extracted["domain"].(string); ok && s != "" {
		domain = s
	} else if rerenderPages, ok := extracted["rerender_pages"].(map[string]interface{}); ok {
		// Backward compatibility
		domain, _ = rerenderPages["domain"].(string)
	}

	// Fallback domain lookup from DB
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
		// No sections stored - skip this page gracefully
		params.Logger.Warn("RerenderSinglePageAction: No sections found, skipping page",
			zap.String("page_name", pageInfo.Name),
			zap.String("page_id", pageIDStr),
		)
		return map[string]interface{}{
			"success":   false,
			"skipped":   true,
			"reason":    "no sections stored for this page",
			"html":      "",
			"domain":    domain,
			"filename":  pageInfo.Filename,
			"page_id":   pageIDStr,
			"page_name": pageInfo.Name,
		}, nil
	}

	// Load site data for template rendering
	siteData := loadSiteData(ctx, params.DB, siteID, params.Logger)

	// Build render context with full site data
	// IMPORTANT: Set BOTH struct fields (for RenderTemplate/contextToMap)
	// AND ContentData (for any direct map access)
	renderCtx := &RenderContext{
		// Struct fields - these are what contextToMap reads
		Domain:      domain,
		Title:       pageInfo.Title,
		Description: pageInfo.MetaDesc,
		CompanyName: siteData.CompanyName,
		Tagline:     siteData.Tagline,
		Email:       siteData.Email,
		Phone:       siteData.Phone,
		CurrentPage: pageInfo.Name,
		LogoText:    siteData.CompanyName,

		// Also set ContentData for any actions that read from it directly
		ContentData: map[string]interface{}{
			"Title":           pageInfo.Title,
			"MetaDescription": pageInfo.MetaDesc,
			"PageName":        pageInfo.Name,
			"CompanyName":     siteData.CompanyName,
			"Tagline":         siteData.Tagline,
			"Email":           siteData.Email,
			"Phone":           siteData.Phone,
			"Domain":          domain,
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

	// Get site contact info for injection
	siteInfo := SiteContactInfo{Email: siteData.Email, Phone: siteData.Phone}

	// Load templates
	headHTML := loadHeadTemplate(ctx, params.DB, siteID, renderCtx, params.Logger)
	headerHTML := loadHeaderTemplate(ctx, params.DB, siteID, renderCtx, params.Logger)
	footerHTML := loadFooterTemplate(ctx, params.DB, siteID, renderCtx, params.Logger)

	// Render head with page-specific data
	renderedHead := executeSimpleTemplate(headHTML, renderCtx.ContentData, params.Logger)

	// Strip any DOCTYPE/html tags from head template (avoid duplicates)
	renderedHead = stripDoctype(renderedHead)

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

// Helper to get map keys for logging
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// SiteContactInfo holds site contact details
type SiteContactInfo struct {
	Email string
	Phone string
}

// SiteData holds full site data for template rendering
type SiteData struct {
	CompanyName string
	Tagline     string
	Email       string
	Phone       string
	Domain      string
}

func loadSiteData(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) SiteData {
	if db == nil {
		return SiteData{}
	}

	var data SiteData
	var contentDataStr sql.NullString

	// Query site table fields and content_data
	err := db.QueryRowContext(ctx, `
		SELECT 
			COALESCE(name, ''), 
			COALESCE(email, ''), 
			COALESCE(phone, ''),
			COALESCE(domain, ''),
			COALESCE(content_data::text, '{}')
		FROM sites WHERE id = $1
	`, siteID).Scan(&data.CompanyName, &data.Email, &data.Phone, &data.Domain, &contentDataStr)

	if err != nil {
		logger.Warn("loadSiteData: Query failed", zap.Error(err))
		return SiteData{}
	}

	// Extract from content_data JSON if main fields are empty
	if contentDataStr.Valid && contentDataStr.String != "" {
		contentData := contentDataStr.String

		// Company name fallback
		if data.CompanyName == "" {
			data.CompanyName = extractJSONString(contentData, "company_name")
		}

		// Tagline
		if data.Tagline == "" {
			data.Tagline = extractJSONString(contentData, "tagline")
		}

		// Email fallback
		if data.Email == "" {
			data.Email = extractJSONString(contentData, "contact_email")
		}

		// Phone fallback
		if data.Phone == "" {
			data.Phone = extractJSONString(contentData, "contact_phone")
		}
	}

	logger.Debug("loadSiteData: Loaded",
		zap.String("company_name", data.CompanyName),
		zap.String("email", data.Email),
		zap.String("phone", data.Phone),
	)

	return data
}

// extractJSONString extracts a string value from JSON using simple pattern matching
// This avoids unmarshaling the entire JSON for just a few fields
func extractJSONString(jsonStr, key string) string {
	// Look for "key": "value" or "key":"value"
	patterns := []string{
		fmt.Sprintf(`"%s"\s*:\s*"([^"]*)"`, key),
		fmt.Sprintf(`"%s":"([^"]*)"`, key),
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(jsonStr); len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
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
	// Remove " | Company Name" suffix
	if idx := strings.Index(label, " | "); idx > 0 {
		label = strings.TrimSpace(label[:idx])
	}
	// Remove " - Company Name" suffix
	if idx := strings.Index(label, " - "); idx > 0 {
		label = strings.TrimSpace(label[:idx])
	}

	// If still too long, derive from URL
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

// stripDoctype removes DOCTYPE, html, body tags from template output
func stripDoctype(html string) string {
	html = regexp.MustCompile(`(?i)<!DOCTYPE[^>]*>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`(?i)<html[^>]*>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`(?i)</html>`).ReplaceAllString(html, "")
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

		// Replace emails in contact/footer sections
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
