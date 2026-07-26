// FILE: platform/orchestration/actions/rerender_pages_action.go
// RerenderSitePagesAction re-assembles all deployed pages with current components
// Uses existing InjectHeader/InjectFooter from component_library.go
// Adds proper head component (for stylesheet links, meta tags)
// DEPRECATED
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
	"github.com/gqls/agentchassis/platform/content"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// RerenderSitePagesAction re-renders all deployed pages with current components
// Config:
//   - site_id_field: path to site_id in collected_data (default: "input_data.site_id")
//   - domain_field: path to domain (fallback if no site_id)
//   - include_statuses: array of page statuses to include (default: ["deployed", "active"])
//   - page_id_field: optional - path to specific page_id to rerender (single page mode)
//   - page_name_field: optional - path to specific page name to rerender (single page mode)
//
// Returns: array of pages ready for deployment with fresh HTML
func RerenderSitePagesAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("RerenderSitePagesAction: Starting",
		zap.String("step_name", params.ExecutionContext.StepName),
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
	)

	config := params.StepConfig.Config

	// Get site_id
	siteIDField := "input_data.site_id"
	if f, ok := config["site_id_field"].(string); ok && f != "" {
		siteIDField = f
	}

	siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
	var siteID uuid.UUID
	var err error

	if siteIDStr != "" {
		siteID, err = uuid.Parse(siteIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid site_id: %w", err)
		}
	}

	// Fallback to domain lookup
	var domain string
	if siteID == uuid.Nil {
		domainField := "input_data.domain"
		if f, ok := config["domain_field"].(string); ok && f != "" {
			domainField = f
		}
		domain = datahelpers.ExtractNestedFieldString(params.CollectedData, domainField)

		if domain != "" {
			siteID, domain, err = rerenderLookupSiteByDomain(ctx, params.DB, domain)
			if err != nil {
				return nil, fmt.Errorf("failed to lookup site by domain %s: %w", domain, err)
			}
		}
	} else {
		// Get domain from site
		domain, _ = rerenderGetDomainForSite(ctx, params.DB, siteID)
	}

	if siteID == uuid.Nil {
		return nil, fmt.Errorf("site_id not found - check site_id_field or domain_field config")
	}

	// Get statuses to include
	includeStatuses := []string{"deployed", "active"}
	if statuses, ok := config["include_statuses"].([]interface{}); ok {
		includeStatuses = []string{}
		for _, s := range statuses {
			if str, ok := s.(string); ok {
				includeStatuses = append(includeStatuses, str)
			}
		}
	}

	// Check for single page filter
	var pageFilter *PageFilter
	if pageIDField, ok := config["page_id_field"].(string); ok && pageIDField != "" {
		if pageIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, pageIDField); pageIDStr != "" {
			pageFilter = &PageFilter{ByID: pageIDStr}
		}
	}
	if pageFilter == nil {
		if pageNameField, ok := config["page_name_field"].(string); ok && pageNameField != "" {
			if pageName := datahelpers.ExtractNestedFieldString(params.CollectedData, pageNameField); pageName != "" {
				pageFilter = &PageFilter{ByName: pageName}
			}
		}
	}

	params.Logger.Info("RerenderSitePagesAction: Loading pages",
		zap.String("site_id", siteID.String()),
		zap.String("domain", domain),
		zap.Strings("include_statuses", includeStatuses),
		zap.Any("page_filter", pageFilter),
	)

	// Load pages
	pages, err := rerenderLoadPages(ctx, params.DB, siteID, includeStatuses, pageFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to load pages: %w", err)
	}

	if len(pages) == 0 {
		return map[string]interface{}{
			"success":        true,
			"site_id":        siteID.String(),
			"domain":         domain,
			"pages_rendered": 0,
			"pages":          []map[string]interface{}{},
			"message":        "no pages found matching criteria",
		}, nil
	}

	params.Logger.Info("RerenderSitePagesAction: Loaded pages",
		zap.Int("page_count", len(pages)),
	)

	// Build base RenderContext using existing function from multipage_actions.go
	// We construct a minimal collectedData structure for compatibility
	collectedData := map[string]interface{}{
		"input_data": map[string]interface{}{
			"domain": domain,
		},
		"site_record": map[string]interface{}{
			"site_id": siteID.String(),
		},
	}

	// Load site content_data for company info
	var companyName, tagline sql.NullString
	params.DB.QueryRowContext(ctx, `
		SELECT 
			COALESCE(content_data->>'company_name', content_data->'reviewed_brief'->>'company_name', '') as company_name,
			COALESCE(content_data->>'tagline', content_data->'reviewed_brief'->>'tagline', '') as tagline
		FROM sites WHERE id = $1
	`, siteID).Scan(&companyName, &tagline)

	if companyName.Valid && companyName.String != "" {
		collectedData["input_data"].(map[string]interface{})["reviewed_brief"] = map[string]interface{}{
			"company_name": companyName.String,
			"tagline":      tagline.String,
		}
	}

	// Use existing buildRenderContextFromCollectedData from multipage_actions.go
	baseRenderCtx := buildRenderContextFromCollectedData(collectedData, params.Logger)
	baseRenderCtx.SiteID = siteID

	// Build navigation items from DB (deployed pages only)
	// This ensures nav matches actual deployed pages, not planned pages
	/*dbNav := rerenderGetHeaderNavFromDB(ctx, params.DB, siteID, 6, params.Logger)*/
	// NavFetchableOnly for rerender — this nav is baked into the page files.
	// The cap of 6 is now applied AFTER filtering, so it means six usable items;
	// under the old SQL LIMIT an unfetchable item inside the first six silently
	// shortened the header.
	dbNav := GetNavItems(ctx, params.DB, siteID, []string{NavGroupPrimary}, NavFetchableOnly, 6, params.Logger)
	if len(dbNav) > 0 {
		baseRenderCtx.NavItems = dbNav
	} else {
		// Fallback to building from provided pages array
		baseRenderCtx.NavItems = rerenderBuildNavItems(pages)
	}

	// Load head component template
	headTemplate := rerenderLoadHeadTemplate(ctx, params.DB, siteID, params.Logger)

	// Re-render each page
	renderedPages := make([]map[string]interface{}, 0, len(pages))
	for _, page := range pages {
		// Build page-specific context
		pageRenderCtx := copyRenderContext(baseRenderCtx)
		currentPage := strings.TrimSuffix(page.Filename, ".html")
		pageRenderCtx.CurrentPage = currentPage
		pageRenderCtx.Title = page.Title
		pageRenderCtx.Description = page.MetaDesc
		pageRenderCtx.NavItems = setActiveNavItems(baseRenderCtx.NavItems, currentPage)

		rendered, err := rerenderSinglePage(ctx, params.DB, page, headTemplate, siteID, pageRenderCtx, params.Logger)
		if err != nil {
			params.Logger.Warn("RerenderSitePagesAction: Failed to render page",
				zap.String("page_name", page.Name),
				zap.Error(err),
			)
			continue
		}
		renderedPages = append(renderedPages, rendered)
	}

	params.Logger.Info("RerenderSitePagesAction: Complete",
		zap.Int("pages_rendered", len(renderedPages)),
	)

	return map[string]interface{}{
		"success":        true,
		"site_id":        siteID.String(),
		"domain":         domain,
		"pages_rendered": len(renderedPages),
		"pages":          renderedPages,
	}, nil
}

// RerenderPageInfo holds page data for re-rendering
type RerenderPageInfo struct {
	ID       string
	Name     string
	Title    string
	URL      string
	Filename string
	MetaDesc string
	NavLabel string
	NavOrder int
	InHeader bool
}

// PageFilter for single page mode
type PageFilter struct {
	ByID   string
	ByName string
}

func rerenderLookupSiteByDomain(ctx context.Context, db *sql.DB, domain string) (uuid.UUID, string, error) {
	var siteID uuid.UUID
	var foundDomain string
	err := db.QueryRowContext(ctx,
		`SELECT id, domain FROM sites WHERE domain = $1 LIMIT 1`,
		domain,
	).Scan(&siteID, &foundDomain)
	return siteID, foundDomain, err
}

func rerenderGetDomainForSite(ctx context.Context, db *sql.DB, siteID uuid.UUID) (string, error) {
	var domain string
	err := db.QueryRowContext(ctx, `SELECT domain FROM sites WHERE id = $1`, siteID).Scan(&domain)
	return domain, err
}

func rerenderLoadPages(ctx context.Context, db *sql.DB, siteID uuid.UUID, statuses []string, filter *PageFilter) ([]RerenderPageInfo, error) {
	// Build base query
	var queryBuilder strings.Builder
	queryBuilder.WriteString(`
		SELECT p.id, p.name, COALESCE(p.title, p.name) as title, p.url,
		       COALESCE(p.meta_description, '') as meta_desc,
		       COALESCE(p.nav_label, p.name) as nav_label,
		       COALESCE(p.nav_order, 100) as nav_order,
		       COALESCE(p.in_header, true) as in_header
		FROM pages p
		WHERE p.site_id = $1
	`)

	args := []interface{}{siteID}
	argIndex := 2

	// Add status filter
	if len(statuses) > 0 {
		statusPlaceholders := make([]string, len(statuses))
		for i, s := range statuses {
			statusPlaceholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, s)
			argIndex++
		}
		queryBuilder.WriteString(fmt.Sprintf(" AND p.status IN (%s)", strings.Join(statusPlaceholders, ",")))
	}

	// Add page filter if specified
	if filter != nil {
		if filter.ByID != "" {
			queryBuilder.WriteString(fmt.Sprintf(" AND p.id = $%d", argIndex))
			args = append(args, filter.ByID)
			argIndex++
		} else if filter.ByName != "" {
			queryBuilder.WriteString(fmt.Sprintf(" AND p.name = $%d", argIndex))
			args = append(args, filter.ByName)
			argIndex++
		}
	}

	queryBuilder.WriteString(" ORDER BY p.nav_order, p.created_at")

	rows, err := db.QueryContext(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pages []RerenderPageInfo
	for rows.Next() {
		var p RerenderPageInfo
		if err := rows.Scan(&p.ID, &p.Name, &p.Title, &p.URL, &p.MetaDesc, &p.NavLabel, &p.NavOrder, &p.InHeader); err != nil {
			continue
		}

		// Determine filename from URL
		p.Filename = p.URL
		if p.Filename == "/" || p.Filename == "" {
			p.Filename = "index.html"
		} else {
			p.Filename = strings.TrimPrefix(p.Filename, "/")
			if !strings.HasSuffix(p.Filename, ".html") {
				p.Filename = p.Filename + ".html"
			}
		}

		pages = append(pages, p)
	}

	return pages, nil
}

func rerenderBuildNavItems(pages []RerenderPageInfo) []NavItem {
	var items []NavItem
	for _, p := range pages {
		if p.InHeader {
			url := "/" + p.Filename
			if p.Filename == "index.html" {
				url = "/index.html"
			}
			items = append(items, NavItem{
				Label: p.NavLabel,
				URL:   url,
			})
		}
	}
	return items
}

// rerenderGetHeaderNavFromDB queries pages table for header nav (deployed pages only)
// DEPRECATED
func rerenderGetHeaderNavFromDB(ctx context.Context, db *sql.DB, siteID uuid.UUID, maxItems int, logger *zap.Logger) []NavItem {
	if db == nil || siteID == uuid.Nil {
		return nil
	}

	if maxItems <= 0 {
		maxItems = 6
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
		logger.Warn("rerenderGetHeaderNavFromDB: Query failed", zap.Error(err))
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
		// Simplify verbose labels
		label = rerenderSimplifyNavLabel(label, url)
		items = append(items, NavItem{Label: label, URL: url})
	}

	logger.Debug("rerenderGetHeaderNavFromDB: Built nav from DB",
		zap.Int("items", len(items)),
	)

	return items
}

// rerenderGetFooterNavFromDB queries pages table for footer nav
// DEPRECATED
func rerenderGetFooterNavFromDB(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) []NavItem {
	if db == nil || siteID == uuid.Nil {
		return nil
	}

	query := `
		SELECT 
			COALESCE(nav_label, title, name) as label,
			COALESCE(url, '/' || name || '.html') as url
		FROM pages 
		WHERE site_id = $1 
		  AND (in_footer = true OR LOWER(name) LIKE '%privacy%' OR LOWER(name) LIKE '%terms%')
		  AND status IN ('deployed', 'active')
		  AND deleted_at IS NULL
		ORDER BY nav_order ASC
	`

	rows, err := db.QueryContext(ctx, query, siteID)
	if err != nil {
		logger.Warn("rerenderGetFooterNavFromDB: Query failed", zap.Error(err))
		return nil
	}
	defer rows.Close()

	var items []NavItem
	for rows.Next() {
		var label, url string
		if err := rows.Scan(&label, &url); err != nil {
			continue
		}
		label = rerenderSimplifyNavLabel(label, url)
		items = append(items, NavItem{Label: label, URL: url})
	}

	return items
}

// rerenderSimplifyNavLabel cleans up verbose labels
func rerenderSimplifyNavLabel(label, url string) string {
	if len(label) <= 15 {
		return label
	}

	// Extract name from URL for mapping
	name := strings.TrimPrefix(url, "/")
	name = strings.TrimSuffix(name, ".html")
	nameLower := strings.ToLower(name)

	switch {
	case nameLower == "index" || nameLower == "home":
		return "Home"
	case strings.HasPrefix(nameLower, "about"):
		return "About"
	case strings.HasPrefix(nameLower, "service"):
		return "Services"
	case strings.HasPrefix(nameLower, "contact"):
		return "Contact"
	case strings.HasPrefix(nameLower, "work") || strings.HasPrefix(nameLower, "portfolio") || strings.HasPrefix(nameLower, "case"):
		return "Work"
	case strings.HasPrefix(nameLower, "team"):
		return "Team"
	case strings.HasPrefix(nameLower, "privacy"):
		return "Privacy"
	case strings.HasPrefix(nameLower, "terms"):
		return "Terms"
	}

	// Take first word if still long
	if len(label) > 20 {
		words := strings.Fields(label)
		if len(words) > 0 {
			return words[0]
		}
	}

	return label
}

func rerenderLoadHeadTemplate(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) string {
	// Try to get head component name from site defaults
	var headComponentName sql.NullString
	db.QueryRowContext(ctx, `
		SELECT defaults->>'head' FROM sites WHERE id = $1
	`, siteID).Scan(&headComponentName)

	componentName := "head-seo-standard"
	if headComponentName.Valid && headComponentName.String != "" {
		componentName = headComponentName.String
	}

	// Load the component template
	var htmlTemplate string
	err := db.QueryRowContext(ctx, `
		SELECT html_template FROM content_components 
		WHERE name = $1 OR function = $1
		LIMIT 1
	`, componentName).Scan(&htmlTemplate)

	if err != nil {
		logger.Warn("Failed to load head component, will use fallback",
			zap.String("component_name", componentName),
			zap.Error(err))
		return ""
	}

	return htmlTemplate
}

func rerenderSinglePage(ctx context.Context, db *sql.DB, page RerenderPageInfo, headTemplate string, siteID uuid.UUID, renderCtx *RenderContext, logger *zap.Logger) (map[string]interface{}, error) {
	// Load section content for this page from page_components
	sectionHTML := rerenderLoadSections(ctx, db, page.ID, logger)

	if sectionHTML == "" {
		return nil, fmt.Errorf("no section content found for page %s", page.Name)
	}

	// Build the complete page
	var html strings.Builder

	html.WriteString("<!DOCTYPE html>\n<html lang=\"en\">\n")

	// Head section - use component or fallback
	if headTemplate != "" {
		// Render head template with page data
		renderedHead := RenderTemplate(headTemplate, renderCtx, logger)
		// Ensure title is set correctly for this page
		if strings.Contains(renderedHead, "<title>") {
			// Replace any placeholder title with actual page title
			renderedHead = regexp.MustCompile(`<title>[^<]*</title>`).ReplaceAllString(
				renderedHead, fmt.Sprintf("<title>%s</title>", page.Title))
		} else {
			renderedHead = strings.Replace(renderedHead, "</head>",
				fmt.Sprintf("    <title>%s</title>\n</head>", page.Title), 1)
		}
		// Ensure meta description is set
		if page.MetaDesc != "" && !strings.Contains(renderedHead, `name="description"`) {
			renderedHead = strings.Replace(renderedHead, "</head>",
				fmt.Sprintf(`    <meta name="description" content="%s">`+"\n</head>", page.MetaDesc), 1)
		}
		html.WriteString(renderedHead)
	} else {
		// Fallback head with essential elements
		html.WriteString("<head>\n")
		html.WriteString("    <meta charset=\"UTF-8\">\n")
		html.WriteString("    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
		html.WriteString(fmt.Sprintf("    <title>%s</title>\n", page.Title))
		if page.MetaDesc != "" {
			html.WriteString(fmt.Sprintf("    <meta name=\"description\" content=\"%s\">\n", page.MetaDesc))
		}
		html.WriteString("    <link rel=\"stylesheet\" href=\"/assets/css/styles.css\">\n")
		html.WriteString("</head>\n")
	}

	html.WriteString("<body>\n")

	// Section content
	html.WriteString(sectionHTML)
	html.WriteString("\n")

	html.WriteString("</body>\n</html>")

	// Inject header and footer using existing component_library functions
	finalHTML := html.String()
	finalHTML = InjectHeader(ctx, db, finalHTML, siteID, renderCtx, logger)
	finalHTML = InjectFooter(ctx, db, finalHTML, siteID, renderCtx, logger)

	// Also inject contact-info from template to ensure correct email/phone
	finalHTML = rerenderInjectContactInfo(ctx, db, finalHTML, siteID, renderCtx, logger)

	// Clean up any double DOCTYPE that might have been in sections
	finalHTML = rerenderCleanDoubleDoctype(finalHTML)

	// Strip tool-doc headers from the outbound HTML (019 §Tool Doc Header) —
	// this bulk path deploys independently of RerenderSinglePageAction, so it
	// needs its own strip. No-op when absent.
	// NOTE (observed, unchanged): unlike the single-page path, this bulk path
	// has no collectJSAssets equivalent — js_content assets are only emitted
	// by single-page rerenders.
	finalHTML = content.StripToolDocHeader(finalHTML)

	logger.Debug("Re-rendered page",
		zap.String("page", page.Name),
		zap.String("filename", page.Filename),
		zap.Int("html_length", len(finalHTML)),
	)

	// Include slug (filename without extension) for git_commit's determinePageFilename
	slug := strings.TrimSuffix(page.Filename, ".html")

	return map[string]interface{}{
		"page_id":  page.ID,
		"title":    page.Title,
		"name":     page.Name,
		"slug":     slug, // For git_commit's determinePageFilename (priority: slug > name > id)
		"filename": page.Filename,
		"html":     finalHTML,
	}, nil
}

func rerenderLoadSections(ctx context.Context, db *sql.DB, pageID string, logger *zap.Logger) string {
	rows, err := db.QueryContext(ctx, `
		SELECT COALESCE(pc.rendered_html, '') as html
		FROM page_components pc
		WHERE pc.page_id = $1
		ORDER BY pc.position
	`, pageID)
	if err != nil {
		logger.Warn("Failed to load page sections", zap.Error(err))
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
			// Strip any existing DOCTYPE/html/head/body wrapper from section
			html = rerenderStripWrapper(html)
			if html != "" {
				sections = append(sections, html)
			}
		}
	}

	return strings.Join(sections, "\n")
}

// rerenderStripWrapper removes DOCTYPE, html, head, body tags from section HTML
// This handles cases where page_components.rendered_html contains full page markup
func rerenderStripWrapper(html string) string {
	// Remove DOCTYPE
	html = regexp.MustCompile(`(?i)<!DOCTYPE[^>]*>`).ReplaceAllString(html, "")
	// Remove <html> tags
	html = regexp.MustCompile(`(?i)</?html[^>]*>`).ReplaceAllString(html, "")
	// Remove entire <head>...</head> (not <header>)
	html = regexp.MustCompile(`(?is)<head(?:\s[^>]*)?>[EXISTING PATTERN]`).ReplaceAllString(html, "")
	// Remove <body> tags
	html = regexp.MustCompile(`(?i)</?body[^>]*>`).ReplaceAllString(html, "")

	// Remove injected site-level headers (with associated style/script blocks)
	// Only strips headers with class="site-header" or preceded by HEADER SOURCE comment
	// to avoid stripping semantic <header> tags inside section content
	html = regexp.MustCompile(
		`(?is)(?:<!--\s*HEADER\s+SOURCE:[^>]*-->\s*)*`+
			`<header\s[^>]*class="site-header[^>]*>.*?</header>`+
			`\s*(?:<style[^>]*>.*?</style>\s*)*`+
			`(?:<script[^>]*>.*?</script>\s*)*`,
	).ReplaceAllString(html, "")

	// NEW: Remove injected site-level footers (with associated style blocks)
	html = regexp.MustCompile(
		`(?is)(?:<!--\s*FOOTER\s+SOURCE:[^>]*-->\s*)*`+
			`<footer\s[^>]*class="site-footer[^>]*>.*?</footer>`+
			`\s*(?:<style[^>]*>.*?</style>\s*)*`,
	).ReplaceAllString(html, "")

	return strings.TrimSpace(html)
}

// rerenderCleanDoubleDoctype removes duplicate DOCTYPE declarations
func rerenderCleanDoubleDoctype(html string) string {
	count := 0
	return regexp.MustCompile(`(?i)<!DOCTYPE[^>]*>`).ReplaceAllStringFunc(html, func(match string) string {
		count++
		if count == 1 {
			return match
		}
		return ""
	})
}

// rerenderInjectContactInfo replaces content-writer generated contact-info with template version
// This ensures correct email/phone from site data rather than hallucinated values
func rerenderInjectContactInfo(ctx context.Context, db *sql.DB, html string, siteID uuid.UUID, renderCtx *RenderContext, logger *zap.Logger) string {
	// Check if page has contact-info section
	// Match section AND any immediately following style block (the component outputs both)
	contactInfoRe := regexp.MustCompile(`(?is)<section[^>]*data-component="contact-info"[^>]*>.*?</section>\s*(?:<style>.*?</style>)?`)
	if !contactInfoRe.MatchString(html) {
		return html // No contact-info section
	}

	// Load contact-info component template
	var htmlTemplate string
	err := db.QueryRowContext(ctx, `
		SELECT html_template FROM content_components 
		WHERE function = 'contact-info' AND is_active = true
		LIMIT 1
	`).Scan(&htmlTemplate)

	if err != nil {
		logger.Warn("Could not load contact-info template", zap.Error(err))
		return html
	}

	// Ensure RenderContext has email/phone from site data
	if renderCtx.Email == "" || renderCtx.Phone == "" {
		rerenderLoadContactFromSite(ctx, db, siteID, renderCtx, logger)
	}

	// Render the template with lowercase keys for Go template compatibility
	// The template uses {{.email}} not {{.Email}}
	templateData := map[string]interface{}{
		"email":         renderCtx.Email,
		"phone":         renderCtx.Phone,
		"phone_display": renderCtx.Phone, // Same value, template may use either
		"title":         "Contact Information",
		"hours":         "Monday – Friday, 9am – 6pm GMT",
	}
	renderedContactInfo := RenderTemplateWithMap(htmlTemplate, templateData, logger)

	// Replace the existing contact-info section (and its style block)
	html = contactInfoRe.ReplaceAllString(html, renderedContactInfo)

	logger.Debug("Injected contact-info from template",
		zap.String("email", renderCtx.Email),
		zap.String("phone", renderCtx.Phone),
	)

	return html
}

// rerenderLoadContactFromSite loads email/phone from site content_data
func rerenderLoadContactFromSite(ctx context.Context, db *sql.DB, siteID uuid.UUID, renderCtx *RenderContext, logger *zap.Logger) {
	var email, phone sql.NullString
	err := db.QueryRowContext(ctx, `
		SELECT 
			COALESCE(
				content_data->>'contact_email',
				content_data->'reviewed_brief'->>'contact_email',
				content_data->'brief'->>'contact_email',
				''
			) as email,
			COALESCE(
				content_data->>'contact_phone',
				content_data->'reviewed_brief'->>'contact_phone',
				content_data->'brief'->>'contact_phone',
				''
			) as phone
		FROM sites WHERE id = $1
	`, siteID).Scan(&email, &phone)

	if err != nil {
		logger.Warn("Could not load contact info from site", zap.Error(err))
		return
	}

	if email.Valid && email.String != "" {
		renderCtx.Email = email.String
	}
	if phone.Valid && phone.String != "" {
		renderCtx.Phone = phone.String
	}

	logger.Debug("Loaded contact info from site",
		zap.String("email", renderCtx.Email),
		zap.String("phone", renderCtx.Phone),
	)
}

// RenderTemplateWithMap renders a Go template with a map of data
// Used for component templates where we need explicit control over field names.
//
// It shares component_library.go's silent-drop discipline: after execution it
// names the bare root-scope fields that rendered empty (escalating a blanked
// href/src to a dead-control Error via missingBareFields) and strips Go's
// "<no value>" artefact. This is the SECOND independent render path the council's
// round-3 audit of the idea.uk chrome fix warned must not be left untouched
// (bugs_open/018) — before this it left "<no value>" visible and logged only on
// parse/execute error.
func RenderTemplateWithMap(templateStr string, data map[string]interface{}, logger *zap.Logger) string {
	tmpl, err := template.New("component").Parse(templateStr)
	if err != nil {
		logger.Warn("Template parse error in RenderTemplateWithMap", zap.Error(err))
		return ""
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		logger.Warn("Template execute error in RenderTemplateWithMap", zap.Error(err))
		return ""
	}

	result := buf.String()

	missing, inURLAttr := missingBareFields(templateStr, data)
	if strings.Contains(result, "<no value>") {
		result = strings.ReplaceAll(result, "<no value>", "")
	}
	if len(inURLAttr) > 0 {
		logger.Error("RenderTemplateWithMap: URL attribute rendered empty — dead control",
			zap.Strings("fields", inURLAttr),
			zap.Strings("all_missing", missing),
			zap.String("template_preview", datahelpers.TruncateString(templateStr, 100)),
		)
	} else if len(missing) > 0 {
		logger.Warn("RenderTemplateWithMap: fields rendered empty",
			zap.Strings("fields", missing),
			zap.String("template_preview", datahelpers.TruncateString(templateStr, 100)),
		)
	}

	return result
}
