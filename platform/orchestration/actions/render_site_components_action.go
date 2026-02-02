// FILE: platform/orchestration/actions/render_site_components_action.go
// RenderSiteComponentsAction renders header/footer/head for a site
// and stores them in site_components table for reuse across all pages

package actions

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// RenderSiteComponentsAction renders site-level components and stores them
//
// Config:
//   - input_fields: fields to extract (default: ["site_id", "domain"])
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

	// Extract inputs
	inputFields := []string{"site_id", "domain"}
	if fields, ok := config["input_fields"].([]interface{}); ok {
		inputFields = make([]string, len(fields))
		for i, f := range fields {
			inputFields[i], _ = f.(string)
		}
	}

	extracted := datahelpers.ExtractFields(params.CollectedData, inputFields, params.Logger)

	// Get site_id
	siteIDStr, _ := extracted["site_id"].(string)
	if siteIDStr == "" {
		return nil, fmt.Errorf("site_id not found in input")
	}
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
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
	siteData, err := loadSiteDataFull(ctx, params.DB, siteID)
	if err != nil {
		return nil, fmt.Errorf("failed to load site data: %w", err)
	}

	params.Logger.Info("RenderSiteComponentsAction: Rendering components",
		zap.String("site_id", siteIDStr),
		zap.String("domain", siteData.Domain),
		zap.Strings("slots", slots),
	)

	// Load navigation for header/footer
	navItems := loadNavItems(ctx, params.DB, siteID, 6, params.Logger)
	footerNavItems := loadFooterNavItems(ctx, params.DB, siteID, 10, params.Logger)

	// Build render context
	year := fmt.Sprintf("%d", time.Now().Year())
	copyright := fmt.Sprintf("© %s %s", year, siteData.CompanyName)

	renderCtx := &RenderContext{
		Domain:      siteData.Domain,
		CompanyName: siteData.CompanyName,
		Tagline:     siteData.Tagline,
		Email:       siteData.Email,
		Phone:       siteData.Phone,
		LogoText:    siteData.LogoText,
		NavItems:    navItems,
		Year:        year,
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

			// Navigation for footer (header uses NavItems field directly)
			"footer_nav_items": footerNavItems,

			// CTA defaults
			"cta_text":       "Get Started",
			"cta_url":        "/contact.html",
			"subscribe_text": "Subscribe",

			// Newsletter defaults (can be overridden)
			"newsletter_title":       "Stay Updated",
			"newsletter_description": "Get the latest news and updates.",
			"email_placeholder":      "Enter your email",
		},
	}

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
}

func loadSiteDataFull(ctx context.Context, db *sql.DB, siteID uuid.UUID) (*SiteDataFull, error) {
	var s SiteDataFull
	s.ID = siteID

	err := db.QueryRowContext(ctx, `
		SELECT 
			domain,
			COALESCE(name, domain),
			COALESCE(company_name, name, domain),
			COALESCE(tagline, ''),
			COALESCE(email, ''),
			COALESCE(phone, ''),
			COALESCE(logo_text, company_name, name, domain),
			COALESCE(logo_url, '')
		FROM sites WHERE id = $1
	`, siteID).Scan(
		&s.Domain, &s.Name, &s.CompanyName, &s.Tagline,
		&s.Email, &s.Phone, &s.LogoText, &s.LogoURL,
	)

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
		err = db.QueryRowContext(ctx, `
			SELECT id, html_template 
			FROM content_components 
			WHERE function = $1
			ORDER BY name
			LIMIT 1
		`, slot).Scan(&componentID, &htmlTemplate)

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
