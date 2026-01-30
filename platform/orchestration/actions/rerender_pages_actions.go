// FILE: platform/orchestration/actions/rerender_pages_action.go
// RerenderSitePagesAction re-assembles all deployed pages with current components
// Uses existing InjectHeader/InjectFooter from component_library.go
// Adds proper head component (for stylesheet links, meta tags)

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

// RerenderSitePagesAction re-renders all deployed pages with current components
// Config:
//   - site_id_field: path to site_id in collected_data (default: "input_data.site_id")
//   - domain_field: path to domain (fallback if no site_id)
//   - include_statuses: array of page statuses to include (default: ["deployed", "active"])
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

	params.Logger.Info("RerenderSitePagesAction: Loading pages",
		zap.String("site_id", siteID.String()),
		zap.String("domain", domain),
		zap.Strings("include_statuses", includeStatuses),
	)

	// Load pages
	pages, err := rerenderLoadPages(ctx, params.DB, siteID, includeStatuses)
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

	// Build navigation items from pages
	baseRenderCtx.NavItems = rerenderBuildNavItems(pages)

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

func rerenderLoadPages(ctx context.Context, db *sql.DB, siteID uuid.UUID, statuses []string) ([]RerenderPageInfo, error) {
	// Build status filter
	statusPlaceholders := make([]string, len(statuses))
	statusArgs := make([]interface{}, len(statuses)+1)
	statusArgs[0] = siteID
	for i, s := range statuses {
		statusPlaceholders[i] = fmt.Sprintf("$%d", i+2)
		statusArgs[i+1] = s
	}

	query := fmt.Sprintf(`
		SELECT p.id, p.name, COALESCE(p.title, p.name) as title, p.url,
		       COALESCE(p.meta_description, '') as meta_desc,
		       COALESCE(p.nav_label, p.name) as nav_label,
		       COALESCE(p.nav_order, 100) as nav_order,
		       COALESCE(p.in_header, true) as in_header
		FROM pages p
		WHERE p.site_id = $1 AND p.status IN (%s)
		ORDER BY p.nav_order, p.created_at
	`, strings.Join(statusPlaceholders, ","))

	rows, err := db.QueryContext(ctx, query, statusArgs...)
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

	// Clean up any double DOCTYPE that might have been in sections
	finalHTML = rerenderCleanDoubleDoctype(finalHTML)

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
	// Remove entire <head>...</head>
	html = regexp.MustCompile(`(?is)<head[^>]*>.*?</head>`).ReplaceAllString(html, "")
	// Remove <body> tags but keep content
	html = regexp.MustCompile(`(?i)<body[^>]*>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`(?i)</body>`).ReplaceAllString(html, "")

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
