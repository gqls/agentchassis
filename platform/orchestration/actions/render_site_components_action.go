// FILE: platform/orchestration/actions/render_site_components_action.go
// RenderSiteComponentsAction renders header/footer/head for a site
// and stores them in site_components table for reuse across all pages

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/storage"
	"go.uber.org/zap"
)

// RenderSiteComponentsAction renders site-level components and stores them
//
// Config:
//   - site_id_field: path to site_id in collected_data (default: "site_record.site_id")
//   - domain_field: path to domain in collected_data (default: "site_record.domain")
//   - slots: which slots to render (default: ["header", "footer", "head"])
//   - force_rerender: re-render even if already exists (default: false)
//
// Returns:
//   - rendered: map of slot_name -> success boolean
//   - site_id: the site ID
func RenderSiteComponentsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("RenderSiteComponentsAction: Starting")

	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	config := params.StepConfig.Config

	// Get site_id using configurable field path (matches UpdateSiteDefaultsAction pattern)
	siteIDField := "site_record.site_id"
	if f, ok := config["site_id_field"].(string); ok && f != "" {
		siteIDField = f
	}
	siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
	if siteIDStr == "" {
		// Fallback: try legacy input_fields approach
		inputFields := []string{"site_id", "domain"}
		if fields, ok := config["input_fields"].([]interface{}); ok {
			inputFields = make([]string, len(fields))
			for i, f := range fields {
				inputFields[i], _ = f.(string)
			}
		}
		extracted := datahelpers.ExtractFields(params.CollectedData, inputFields, params.Logger)
		siteIDStr, _ = extracted["site_id"].(string)
	}

	if siteIDStr == "" {
		return nil, fmt.Errorf("site_id not found at %s", siteIDField)
	}

	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id at %s: %w (got: %q, len: %d)", siteIDField, err, siteIDStr, len(siteIDStr))
	}

	// Get slots to render
	slots := []string{"header", "footer", "head"}
	if s, ok := config["slots"].([]interface{}); ok {
		slots = make([]string, len(s))
		for i, v := range s {
			slots[i], _ = v.(string)
		}
	}

	forceRerender, _ := config["force_rerender"].(bool)

	// Load site data
	siteData, err := loadSiteDataFull(ctx, params.DB, siteID, params.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to load site data: %w", err)
	}

	params.Logger.Info("RenderSiteComponentsAction: Rendering components",
		zap.String("site_id", siteIDStr),
		zap.String("domain", siteData.Domain),
		zap.Strings("slots", slots),
	)

	// Load navigation for header/footer
	/*	navItems := loadNavItems(ctx, params.DB, siteID, 6, params.Logger)
		footerNavItems := loadFooterNavItems(ctx, params.DB, siteID, 10, params.Logger)*/
	// deployedOnly=false: runs during build when pages may not be deployed yet.
	// maxItems=0: no limit — PopulateNavTablesAction controls primary group membership.
	navItems := GetNavItems(ctx, params.DB, siteID, []string{NavGroupPrimary}, false, 0, params.Logger)
	footerNavItems := GetNavItems(ctx, params.DB, siteID, []string{NavGroupPrimary, NavGroupUtility, NavGroupLegal}, false, 0, params.Logger)

	// Build render context
	year := fmt.Sprintf("%d", time.Now().Year())
	copyright := fmt.Sprintf("© %s %s", year, siteData.CompanyName)

	// Build pre-rendered nav HTML for templates that use {{.nav_items_html}}
	// Uses existing buildNavItemsHTML from component_library.go
	// Header templates use nav_items_html (primary only).
	navItemsHTML := buildNavItemsHTML(navItems)

	// Build quick links HTML for footer — includes primary + utility items.
	// Utility items are pages that overflowed from primary or were classified
	// as secondary (FAQ, Approach, Insights etc). They belong in footer nav
	// but not the header. Legal items (privacy, terms) get their own footer section.
	quickLinksItems := GetNavItems(ctx, params.DB, siteID, []string{NavGroupPrimary, NavGroupUtility}, false, 0, params.Logger)
	quickLinksHTML := buildNavItemsHTML(quickLinksItems)

	// Build services HTML for footer "Our Services" column
	// Query pages that represent services (linked from services page or service-named pages)
	servicesHTML := buildServicesHTML(ctx, params.DB, siteID, params.Logger)

	// Convert NavItems to categories format for templates that use {{range .categories}}
	categories := make([]map[string]interface{}, len(navItems))
	for i, item := range navItems {
		categories[i] = map[string]interface{}{
			"name":  item.Label,
			"slug":  strings.TrimSuffix(strings.TrimPrefix(item.URL, "/"), ".html"),
			"url":   item.URL,
			"label": item.Label, // alias
		}
	}

	// Convert footer nav items similarly
	footerLinks := make([]map[string]interface{}, len(footerNavItems))
	for i, item := range footerNavItems {
		footerLinks[i] = map[string]interface{}{
			"name":  item.Label,
			"slug":  strings.TrimSuffix(strings.TrimPrefix(item.URL, "/"), ".html"),
			"url":   item.URL,
			"label": item.Label,
		}
	}

	// Build company links (about, contact, careers).
	// Also capture the real contact page URL for the header CTA, so it points
	// at an existing page instead of the hardcoded phantom /contact.html.
	companyLinks := []map[string]interface{}{}
	ctaURL := ""
	for _, item := range footerNavItems {
		lowerLabel := strings.ToLower(item.Label)
		if lowerLabel == "about" || lowerLabel == "contact" || lowerLabel == "careers" {
			companyLinks = append(companyLinks, map[string]interface{}{
				"name": item.Label,
				"url":  item.URL,
			})
		}
		if lowerLabel == "contact" {
			ctaURL = item.URL
		}
	}

	// Build legal links from real pages classified into the legal nav group.
	// Was a hardcoded {/privacy.html, /terms.html} slice — those pages do not
	// necessarily exist, so it produced phantom links. Now: only pages that
	// actually exist appear; if none, the list is empty and the footer renders
	// no legal links.
	legalNavItems := GetNavItems(ctx, params.DB, siteID, []string{NavGroupLegal}, false, 0, params.Logger)
	legalLinks := make([]map[string]interface{}, 0, len(legalNavItems))
	for _, item := range legalNavItems {
		legalLinks = append(legalLinks, map[string]interface{}{
			"name": item.Label,
			"url":  item.URL,
		})
	}

	// Social links (empty for now - could be populated from site data)
	socialLinks := []map[string]interface{}{}

	renderCtx := &RenderContext{
		Domain:         siteData.Domain,
		CompanyName:    siteData.CompanyName,
		Tagline:        siteData.Tagline,
		Email:          siteData.Email,
		Phone:          siteData.Phone,
		LogoText:       siteData.LogoText,
		LogoURL:        siteData.LogoURL,
		NavItems:       navItems,
		FooterNavItems: quickLinksItems,
		Year:           year,

		// Colors from style collection (RenderContext struct fields feed contextToInterfaceMap defaults)
		PrimaryColor:    siteData.PrimaryColor,
		SecondaryColor:  siteData.SecondaryColor,
		AccentColor:     siteData.AccentColor,
		TextColor:       siteData.TextColor,
		BackgroundColor: siteData.BackgroundColor,

		// Theme CSS from css_themes table
		ThemeCSS: siteData.ThemeCSS,

		ContentData: map[string]interface{}{
			// Core site info
			"company_name":  siteData.CompanyName,
			"brand_name":    siteData.CompanyName, // alias
			"tagline":       siteData.Tagline,
			"domain":        siteData.Domain,
			"email":         siteData.Email,
			"contact_email": siteData.Email, // alias
			"phone":         siteData.Phone,
			"logo_text":     siteData.LogoText,
			"logo_url":      siteData.LogoURL,
			"year":          year,
			"copyright":     copyright,

			// Pre-rendered nav HTML for templates using {{.nav_items_html}}
			"nav_items_html":   navItemsHTML,
			"quick_links_html": quickLinksHTML, // primary + utility items for footer
			"services_html":    servicesHTML,

			// Navigation - multiple formats for different templates
			"categories":       categories,   // for {{range .categories}}
			"nav_items":        categories,   // alias
			"footer_nav_items": footerLinks,  // for footer
			"quick_links":      footerLinks,  // alias for footer
			"company_links":    companyLinks, // about, contact, careers
			"legal_links":      legalLinks,   // privacy, terms
			"social_links":     socialLinks,  // social media (empty for now)

			// CTA defaults — cta_url resolved from the real contact page above
			// (empty when there is no contact page; the gated header template
			// then renders no CTA button rather than a phantom).
			"cta_text":       "Get Started",
			"cta_url":        ctaURL,
			"subscribe_text": "Subscribe",
			"show_subscribe": false,

			// Newsletter defaults (can be overridden)
			"newsletter_title":       "Stay Updated",
			"newsletter_description": "Get the latest news and updates.",
			"email_placeholder":      "Enter your email",
		},
	}

	params.Logger.Info("RenderSiteComponentsAction: Render context built",
		zap.Int("nav_items", len(navItems)),
		zap.Int("nav_items_html_len", len(navItemsHTML)),
		zap.Bool("has_theme_css", siteData.ThemeCSS != ""),
		zap.String("primary_color", siteData.PrimaryColor),
	)

	// Render each slot
	rendered := make(map[string]bool)
	for _, slot := range slots {
		success := renderAndStoreSiteComponent(ctx, params.DB, siteID, slot, renderCtx, forceRerender, params.Logger)
		rendered[slot] = success
	}

	params.Logger.Info("RenderSiteComponentsAction: Complete",
		zap.Any("rendered", rendered),
	)

	return map[string]interface{}{
		"success":  true,
		"site_id":  siteIDStr,
		"domain":   siteData.Domain,
		"rendered": rendered,
	}, nil
}

// SiteDataFull contains all site data needed for rendering
type SiteDataFull struct {
	ID          uuid.UUID
	Domain      string
	Name        string
	CompanyName string
	Tagline     string
	Email       string
	Phone       string
	LogoText    string
	LogoURL     string

	// Style collection data (loaded via sites.style_collection_id)
	PrimaryColor    string
	SecondaryColor  string
	AccentColor     string
	TextColor       string
	TextLightColor  string
	BackgroundColor string
	BackgroundAlt   string
	ThemeCSS        string // from css_themes table
	FontURL         string // Google Fonts URL if set
}

func loadSiteDataFull(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) (*SiteDataFull, error) {
	var s SiteDataFull
	s.ID = siteID

	var colorPaletteJSON sql.NullString
	var themeCSSContent sql.NullString

	err := db.QueryRowContext(ctx, `
		SELECT 
			si.domain,
			COALESCE(si.name, si.domain),
			COALESCE(si.company_name, si.name, si.domain),
			COALESCE(si.tagline, ''),
			COALESCE(si.email, ''),
			COALESCE(si.phone, ''),
			COALESCE(si.logo_text, si.company_name, si.name, si.domain),
			COALESCE(si.logo_url, ''),
			sc.color_palette::text,
			ct.css_content
		FROM sites si
		LEFT JOIN style_collections sc ON si.style_collection_id = sc.id
		LEFT JOIN css_themes ct ON sc.css_theme_id = ct.id
		WHERE si.id = $1
	`, siteID).Scan(
		&s.Domain, &s.Name, &s.CompanyName, &s.Tagline,
		&s.Email, &s.Phone, &s.LogoText, &s.LogoURL,
		&colorPaletteJSON, &themeCSSContent,
	)
	if err != nil {
		return nil, err
	}

	// Parse color palette from style collection
	if colorPaletteJSON.Valid && colorPaletteJSON.String != "" {
		var palette map[string]string
		if jsonErr := json.Unmarshal([]byte(colorPaletteJSON.String), &palette); jsonErr == nil {
			s.PrimaryColor = palette["primary"]
			s.SecondaryColor = palette["secondary"]
			s.AccentColor = palette["accent"]
			s.TextColor = palette["text"]
			s.TextLightColor = palette["text_light"]
			s.BackgroundColor = palette["background"]
			s.BackgroundAlt = palette["background_alt"]
		}
	}

	// Load theme CSS
	if themeCSSContent.Valid {
		s.ThemeCSS = themeCSSContent.String
	}

	// Phase I1 (closes the "logo-in-header" gap): resolve the site logo from
	// the current plan's imagery rows joined to the deployed asset, exactly
	// as plan_sections resolves page heroes. The legacy sites.logo_url value
	// (loaded above) remains the fallback for adopted/guide-less sites. The
	// resolved value is the DERIVED committed git path
	// (storage.DeployedWebPath) — never assets.url, which holds an expiring
	// presigned URL. Only override when the asset row exists AND is active,
	// so headers never reference a file the deployer hasn't committed.
	var logoKey, logoPurpose string
	logoErr := db.QueryRowContext(ctx, `
		SELECT a.asset_key, a.purpose
		  FROM site_plan_imagery spi
		  JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current = true
		  JOIN assets a ON a.site_id = sp.site_id
		               AND a.asset_key = spi.key
		               AND a.status = 'active'
		 WHERE sp.site_id = $1
		   AND spi.scope = 'site'
		   AND spi.kind = 'logo'
		 ORDER BY spi.ordering
		 LIMIT 1
	`, siteID).Scan(&logoKey, &logoPurpose)
	if logoErr == nil && logoKey != "" {
		s.LogoURL = storage.DeployedWebPath(logoKey, logoPurpose)
	} else if logoErr != nil && logoErr != sql.ErrNoRows {
		// Non-fatal: keep the legacy fallback and continue.
		logger.Warn("loadSiteDataFull: plan logo lookup failed — falling back to sites.logo_url",
			zap.Error(logoErr))
	}

	return &s, err
}

func renderAndStoreSiteComponent(
	ctx context.Context,
	db *sql.DB,
	siteID uuid.UUID,
	slot string,
	renderCtx *RenderContext,
	force bool,
	logger *zap.Logger,
) bool {
	// Check if already rendered (unless force)
	if !force {
		var exists bool
		db.QueryRowContext(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM site_components 
				WHERE site_id = $1 AND slot_name = $2 
				AND rendered_html IS NOT NULL AND rendered_html != ''
			)
		`, siteID, slot).Scan(&exists)

		if exists {
			logger.Debug("Site component already rendered, skipping",
				zap.String("slot", slot))
			return true
		}
	}

	// Get component template
	var componentID uuid.UUID
	var htmlTemplate string

	err := db.QueryRowContext(ctx, `
		SELECT sc.component_id, cc.html_template
		FROM site_components sc
		JOIN content_components cc ON sc.component_id = cc.id
		WHERE sc.site_id = $1 AND sc.slot_name = $2
	`, siteID, slot).Scan(&componentID, &htmlTemplate)

	if err != nil {
		// No component assigned, try to find default
		slotToFunction := map[string]string{
			"header": "site-header",
			"footer": "site-footer",
			"head":   "head",
		}
		funcName := slot
		if mapped, ok := slotToFunction[slot]; ok {
			funcName = mapped
		}
		err = db.QueryRowContext(ctx, `
			SELECT id, html_template 
			FROM content_components 
			WHERE function = $1
			ORDER BY name LIMIT 1
		`, funcName).Scan(&componentID, &htmlTemplate)

		if err != nil {
			logger.Warn("No component found for slot",
				zap.String("slot", slot),
				zap.Error(err))
			return false
		}

		// Insert the component assignment
		db.ExecContext(ctx, `
			INSERT INTO site_components (site_id, slot_name, component_id, build_status)
			VALUES ($1, $2, $3, 'pending')
			ON CONFLICT (site_id, slot_name) DO UPDATE SET component_id = $3
		`, siteID, slot, componentID)
	}

	// Render the template
	renderedHTML := RenderTemplate(htmlTemplate, renderCtx, logger)

	if renderedHTML == "" {
		logger.Warn("Template rendered to empty string",
			zap.String("slot", slot))
		return false
	}

	// Store the rendered HTML
	_, err = db.ExecContext(ctx, `
		UPDATE site_components 
		SET rendered_html = $1, build_status = 'rendered', updated_at = now()
		WHERE site_id = $2 AND slot_name = $3
	`, renderedHTML, siteID, slot)

	if err != nil {
		logger.Error("Failed to store rendered component",
			zap.String("slot", slot),
			zap.Error(err))
		return false
	}

	logger.Info("Site component rendered and stored",
		zap.String("slot", slot),
		zap.Int("html_length", len(renderedHTML)))

	return true
}

// loadNavItems loads navigation items for header
// DEPRECATED
func loadNavItems(ctx context.Context, db *sql.DB, siteID uuid.UUID, maxItems int, logger *zap.Logger) []NavItem {
	query := `
		SELECT 
			COALESCE(nav_label, title, name) as label,
			COALESCE(url, '/' || name || '.html') as url
		FROM pages 
		WHERE site_id = $1 
		  AND (in_header = true OR in_header IS NULL)
		  AND status IN ('deployed', 'active')
		  AND name NOT IN ('privacy', 'terms', 'cookies', '404', 'sitemap')
		ORDER BY 
			COALESCE(nav_order, 99),
			CASE name 
				WHEN 'index' THEN 1 
				WHEN 'home' THEN 1
				WHEN 'services' THEN 2
				WHEN 'about' THEN 3
				WHEN 'contact' THEN 10
				ELSE 5 
			END
		LIMIT $2
	`

	rows, err := db.QueryContext(ctx, query, siteID, maxItems)
	if err != nil {
		logger.Warn("loadNavItems: Query failed", zap.Error(err))
		return nil
	}
	defer rows.Close()

	var items []NavItem
	for rows.Next() {
		var label, url string
		if err := rows.Scan(&label, &url); err != nil {
			continue
		}
		// Simplify label
		label = strings.Title(strings.ReplaceAll(label, "-", " "))
		items = append(items, NavItem{Label: label, URL: url})
	}

	return items
}

// loadFooterNavItems loads navigation items for footer
// DEPRECATED
func loadFooterNavItems(ctx context.Context, db *sql.DB, siteID uuid.UUID, maxItems int, logger *zap.Logger) []NavItem {
	query := `
		SELECT 
			COALESCE(nav_label, title, name) as label,
			COALESCE(url, '/' || name || '.html') as url
		FROM pages 
		WHERE site_id = $1 
		  AND (in_footer = true OR in_footer IS NULL)
		  AND status IN ('deployed', 'active')
		  AND name NOT IN ('index', '404', 'sitemap')
		ORDER BY 
			COALESCE(nav_order, 99),
			CASE name 
				WHEN 'services' THEN 1
				WHEN 'about' THEN 2
				WHEN 'contact' THEN 3
				WHEN 'privacy' THEN 8
				WHEN 'terms' THEN 9
				ELSE 5 
			END
		LIMIT $2
	`

	rows, err := db.QueryContext(ctx, query, siteID, maxItems)
	if err != nil {
		logger.Warn("loadFooterNavItems: Query failed", zap.Error(err))
		return nil
	}
	defer rows.Close()

	var items []NavItem
	for rows.Next() {
		var label, url string
		if err := rows.Scan(&label, &url); err != nil {
			continue
		}
		label = strings.Title(strings.ReplaceAll(label, "-", " "))
		items = append(items, NavItem{Label: label, URL: url})
	}

	return items
}

// buildServicesHTML queries service-related pages and builds <li> HTML for the footer services column.
// Looks for pages that represent individual service offerings (excludes structural pages).
// Falls back to an empty string if no service pages found.
func buildServicesHTML(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) string {
	rows, err := db.QueryContext(ctx, `
		SELECT 
			COALESCE(nav_label, title, name) as label,
			COALESCE(url, '/' || name || '.html') as url
		FROM pages
		WHERE site_id = $1
		  AND status IN ('deployed', 'active')
		  AND name NOT IN ('index', 'about', 'contact', 'privacy', 'terms', 'cookies', '404', 'sitemap', 'faq', 'careers', 'insights', 'blog', 'news')
		  AND name != 'services'
		  AND (in_header = true OR in_footer = true)
		ORDER BY COALESCE(nav_order, 99), name
		LIMIT 6
	`, siteID)
	if err != nil {
		logger.Warn("buildServicesHTML: Query failed", zap.Error(err))
		return ""
	}
	defer rows.Close()

	var parts []string
	for rows.Next() {
		var label, url string
		if err := rows.Scan(&label, &url); err != nil {
			continue
		}
		label = strings.ReplaceAll(label, "-", " ")
		words := strings.Fields(label)
		for i, w := range words {
			if len(w) > 0 {
				words[i] = strings.ToUpper(w[:1]) + w[1:]
			}
		}
		label = strings.Join(words, " ")
		parts = append(parts, fmt.Sprintf(`<li><a href="%s">%s</a></li>`, url, label))
	}

	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, "\n                ")
}
