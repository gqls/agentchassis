// ===========================================================================
// COMPONENT-BASED HEADER RENDERING
// FILE: platform/orchestration/actions/component_renderer.go
// ===========================================================================
// Renders headers (and other components) from database templates.
// Uses Handlebars-style templating for variable substitution.
// ===========================================================================

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// StyleCollection represents a bundle of components for a site
type StyleCollection struct {
	ID                    uuid.UUID         `json:"id"`
	Name                  string            `json:"name"`
	DisplayName           string            `json:"display_name"`
	HeaderComponentID     *uuid.UUID        `json:"header_component_id"`
	HeaderHomeComponentID *uuid.UUID        `json:"header_home_component_id"`
	FooterComponentID     *uuid.UUID        `json:"footer_component_id"`
	CSSThemeID            *uuid.UUID        `json:"css_theme_id"`
	ColorPalette          map[string]string `json:"color_palette"`
	Typography            map[string]string `json:"typography"`
	Category              string            `json:"category"`
}

// HeaderRenderInput contains all data needed to render a header
type HeaderRenderInput struct {
	LogoText     string    `json:"logo_text"`
	LogoAccent   string    `json:"logo_accent,omitempty"`
	PrimaryColor string    `json:"primary_color"`
	AccentColor  string    `json:"accent_color"`
	TextColor    string    `json:"text_color,omitempty"`
	Background   string    `json:"background_color,omitempty"`
	NavItems     []NavItem `json:"nav_items"`
	CurrentPage  string    `json:"current_page,omitempty"`
	IsHomePage   bool      `json:"is_home_page,omitempty"`
}

// ===========================================================================
// COMPONENT LOADING
// ===========================================================================

// GetStyleCollectionForSite retrieves the style collection for a site
func GetStyleCollectionForSite(ctx context.Context, db interface{}, siteID uuid.UUID, logger *zap.Logger) (*StyleCollection, error) {
	query := `
		SELECT 
			sc.id, sc.name, sc.display_name,
			sc.header_component_id, sc.header_home_component_id,
			sc.footer_component_id, sc.css_theme_id,
			sc.color_palette, sc.typography, sc.category
		FROM sites s
		JOIN style_collections sc ON s.style_collection_id = sc.id
		WHERE s.id = $1
	`

	var collection StyleCollection
	var colorPaletteJSON, typographyJSON []byte
	var headerID, headerHomeID, footerID, themeID *uuid.UUID

	var err error
	switch d := db.(type) {
	case *sql.DB:
		err = d.QueryRowContext(ctx, query, siteID).Scan(
			&collection.ID, &collection.Name, &collection.DisplayName,
			&headerID, &headerHomeID, &footerID, &themeID,
			&colorPaletteJSON, &typographyJSON, &collection.Category,
		)
	case *pgxpool.Pool:
		err = d.QueryRow(ctx, query, siteID).Scan(
			&collection.ID, &collection.Name, &collection.DisplayName,
			&headerID, &headerHomeID, &footerID, &themeID,
			&colorPaletteJSON, &typographyJSON, &collection.Category,
		)
	default:
		return nil, fmt.Errorf("unsupported database type: %T", db)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			logger.Info("No style collection for site, will use defaults",
				zap.String("site_id", siteID.String()))
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get style collection: %w", err)
	}

	collection.HeaderComponentID = headerID
	collection.HeaderHomeComponentID = headerHomeID
	collection.FooterComponentID = footerID
	collection.CSSThemeID = themeID

	// Parse JSON fields
	if len(colorPaletteJSON) > 0 {
		json.Unmarshal(colorPaletteJSON, &collection.ColorPalette)
	}
	if len(typographyJSON) > 0 {
		json.Unmarshal(typographyJSON, &collection.Typography)
	}

	return &collection, nil
}

// GetComponentTemplate retrieves a component template by ID or name
func GetComponentTemplate(ctx context.Context, db interface{}, identifier string, logger *zap.Logger) (*ComponentTemplate, error) {
	// Try as UUID first, then as name
	var query string
	var arg interface{}

	if id, err := uuid.Parse(identifier); err == nil {
		query = `SELECT id, name, html_template, input_schema, category FROM content_components WHERE id = $1`
		arg = id
	} else {
		query = `SELECT id, name, html_template, input_schema, category FROM content_components WHERE name = $1`
		arg = identifier
	}

	var component ComponentTemplate
	var schemaJSON []byte

	var err error
	switch d := db.(type) {
	case *sql.DB:
		err = d.QueryRowContext(ctx, query, arg).Scan(
			&component.ID, &component.Name, &component.HTMLTemplate, &schemaJSON, &component.Category,
		)
	case *pgxpool.Pool:
		err = d.QueryRow(ctx, query, arg).Scan(
			&component.ID, &component.Name, &component.HTMLTemplate, &schemaJSON, &component.Category,
		)
	default:
		return nil, fmt.Errorf("unsupported database type: %T", db)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get component: %w", err)
	}

	if len(schemaJSON) > 0 {
		json.Unmarshal(schemaJSON, &component.InputSchema)
	}

	return &component, nil
}

// GetHeaderComponentForSite gets the appropriate header component for a site
func GetHeaderComponentForSite(ctx context.Context, db interface{}, siteID uuid.UUID, isHomePage bool, logger *zap.Logger) (*ComponentTemplate, error) {
	collection, err := GetStyleCollectionForSite(ctx, db, siteID, logger)
	if err != nil {
		return nil, err
	}

	// If no collection, return nil (caller should use default)
	if collection == nil {
		return nil, nil
	}

	// Choose header variant
	var componentID *uuid.UUID
	if isHomePage && collection.HeaderHomeComponentID != nil {
		componentID = collection.HeaderHomeComponentID
	} else {
		componentID = collection.HeaderComponentID
	}

	if componentID == nil {
		return nil, nil
	}

	return GetComponentTemplate(ctx, db, componentID.String(), logger)
}

// ===========================================================================
// TEMPLATE RENDERING
// ===========================================================================

// RenderComponent renders a component template with the given data
func RenderComponent(template string, data interface{}, logger *zap.Logger) (string, error) {
	// Convert data to map for template substitution
	dataMap, err := toMap(data)
	if err != nil {
		return "", fmt.Errorf("failed to convert data: %w", err)
	}

	result := template

	// Handle {{#each nav_items}}...{{/each}} blocks
	result = renderEachBlocks(result, dataMap)

	// Handle {{#if field}}...{{/if}} blocks
	result = renderIfBlocks(result, dataMap)

	// Handle simple {{field}} substitutions
	result = renderSimpleSubstitutions(result, dataMap)

	// Handle {{this.field}} references (inside each blocks, already processed)
	// These should be cleaned up if any remain
	result = regexp.MustCompile(`\{\{this\.\w+\}\}`).ReplaceAllString(result, "")

	return result, nil
}

// renderEachBlocks handles {{#each array}}...{{/each}} blocks
func renderEachBlocks(template string, data map[string]interface{}) string {
	eachRe := regexp.MustCompile(`(?s)\{\{#each\s+(\w+)\}\}(.*?)\{\{/each\}\}`)

	return eachRe.ReplaceAllStringFunc(template, func(match string) string {
		matches := eachRe.FindStringSubmatch(match)
		if len(matches) < 3 {
			return match
		}

		arrayName := matches[1]
		itemTemplate := matches[2]

		arrayData, ok := data[arrayName]
		if !ok {
			return ""
		}

		items, ok := arrayData.([]interface{})
		if !ok {
			// Try []NavItem
			if navItems, ok := arrayData.([]NavItem); ok {
				var result strings.Builder
				for _, item := range navItems {
					itemStr := itemTemplate
					itemStr = strings.ReplaceAll(itemStr, "{{this.url}}", item.URL)
					itemStr = strings.ReplaceAll(itemStr, "{{this.label}}", item.Label)
					if item.IsActive {
						itemStr = strings.ReplaceAll(itemStr, `{{#if this.is_active}} class="active"{{/if}}`, ` class="active"`)
					} else {
						itemStr = strings.ReplaceAll(itemStr, `{{#if this.is_active}} class="active"{{/if}}`, "")
					}
					result.WriteString(itemStr)
				}
				return result.String()
			}
			return ""
		}

		var result strings.Builder
		for _, item := range items {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			itemStr := itemTemplate
			for key, val := range itemMap {
				placeholder := fmt.Sprintf("{{this.%s}}", key)
				itemStr = strings.ReplaceAll(itemStr, placeholder, fmt.Sprintf("%v", val))
			}

			// Handle {{#if this.field}} blocks within the item
			itemStr = renderIfBlocksForItem(itemStr, itemMap)

			result.WriteString(itemStr)
		}

		return result.String()
	})
}

// renderIfBlocks handles {{#if field}}...{{/if}} blocks
func renderIfBlocks(template string, data map[string]interface{}) string {
	ifRe := regexp.MustCompile(`(?s)\{\{#if\s+(\w+)\}\}(.*?)\{\{/if\}\}`)

	return ifRe.ReplaceAllStringFunc(template, func(match string) string {
		matches := ifRe.FindStringSubmatch(match)
		if len(matches) < 3 {
			return match
		}

		fieldName := matches[1]
		content := matches[2]

		value, exists := data[fieldName]
		if !exists || value == nil || value == "" || value == false {
			return ""
		}

		return content
	})
}

// renderIfBlocksForItem handles {{#if this.field}} within each loops
func renderIfBlocksForItem(template string, itemData map[string]interface{}) string {
	ifRe := regexp.MustCompile(`(?s)\{\{#if\s+this\.(\w+)\}\}(.*?)\{\{/if\}\}`)

	return ifRe.ReplaceAllStringFunc(template, func(match string) string {
		matches := ifRe.FindStringSubmatch(match)
		if len(matches) < 3 {
			return match
		}

		fieldName := matches[1]
		content := matches[2]

		value, exists := itemData[fieldName]
		if !exists || value == nil || value == "" || value == false {
			return ""
		}

		return content
	})
}

// renderSimpleSubstitutions handles {{field}} placeholders
func renderSimpleSubstitutions(template string, data map[string]interface{}) string {
	simpleRe := regexp.MustCompile(`\{\{(\w+)\}\}`)

	return simpleRe.ReplaceAllStringFunc(template, func(match string) string {
		matches := simpleRe.FindStringSubmatch(match)
		if len(matches) < 2 {
			return match
		}

		fieldName := matches[1]
		value, exists := data[fieldName]
		if !exists || value == nil {
			return match // Keep placeholder if no value
		}

		return fmt.Sprintf("%v", value)
	})
}

// toMap converts a struct to map[string]interface{}
func toMap(data interface{}) (map[string]interface{}, error) {
	if m, ok := data.(map[string]interface{}); ok {
		return m, nil
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	return result, err
}

// ===========================================================================
// HIGH-LEVEL RENDERING FUNCTIONS
// ===========================================================================

// RenderHeaderForSite renders the header for a specific site and page
func RenderHeaderForSite(ctx context.Context, db interface{}, siteID uuid.UUID, input *HeaderRenderInput, logger *zap.Logger) (string, error) {
	// Try to get site's header component
	component, err := GetHeaderComponentForSite(ctx, db, siteID, input.IsHomePage, logger)
	if err != nil {
		logger.Warn("Failed to get header component, using default",
			zap.Error(err))
	}

	// If no component found, use default
	if component == nil {
		return RenderDefaultHeader(input), nil
	}

	// Set active states on nav items
	for i := range input.NavItems {
		urlPage := strings.TrimSuffix(strings.TrimPrefix(input.NavItems[i].URL, "/"), ".html")
		input.NavItems[i].IsActive = urlPage == input.CurrentPage ||
			(input.CurrentPage == "index" && (urlPage == "home" || urlPage == "index"))
	}

	// Render the component
	rendered, err := RenderComponent(component.HTMLTemplate, input, logger)
	if err != nil {
		logger.Warn("Failed to render header component, using default",
			zap.Error(err))
		return RenderDefaultHeader(input), nil
	}

	return rendered, nil
}

// RenderDefaultHeader renders the built-in default header
func RenderDefaultHeader(input *HeaderRenderInput) string {
	// This is the fallback when no component is found
	var navLinks []string
	for _, item := range input.NavItems {
		activeClass := ""
		if item.IsActive {
			activeClass = ` class="active"`
		}
		navLinks = append(navLinks, fmt.Sprintf(
			`<li><a href="%s"%s>%s</a></li>`,
			item.URL, activeClass, item.Label,
		))
	}

	navHTML := strings.Join(navLinks, "\n                ")

	if input.PrimaryColor == "" {
		input.PrimaryColor = "#1a1a2e"
	}
	if input.AccentColor == "" {
		input.AccentColor = "#16a085"
	}

	return fmt.Sprintf(`<header class="site-header">
    <div class="header-container">
        <a href="/index.html" class="logo">%s</a>
        <nav class="main-nav">
            <ul>
                %s
            </ul>
        </nav>
    </div>
</header>
<style>
.site-header { background: %s; padding: 1rem 0; position: sticky; top: 0; z-index: 1000; }
.header-container { max-width: 1200px; margin: 0 auto; padding: 0 2rem; display: flex; align-items: center; justify-content: space-between; }
.logo { text-decoration: none; font-size: 1.5rem; font-weight: 700; color: white; }
.main-nav ul { display: flex; list-style: none; margin: 0; padding: 0; gap: 2rem; }
.main-nav a { color: rgba(255,255,255,0.9); text-decoration: none; font-weight: 500; }
.main-nav a:hover, .main-nav a.active { color: %s; }
</style>`, input.LogoText, navHTML, input.PrimaryColor, input.AccentColor)
}

// ===========================================================================
// INTEGRATION: Update injectConsistentHeader to use components
// ===========================================================================

// injectConsistentHeaderFromComponent replaces header using component system
func injectConsistentHeaderFromComponent(ctx context.Context, db interface{}, html string, siteID uuid.UUID, input *HeaderRenderInput, logger *zap.Logger) string {
	// Render header from component or default
	headerHTML, err := RenderHeaderForSite(ctx, db, siteID, input, logger)
	if err != nil {
		logger.Warn("Failed to render header from component",
			zap.Error(err))
		headerHTML = RenderDefaultHeader(input)
	}

	// Remove existing header
	headerRe := regexp.MustCompile(`(?is)<header[^>]*>.*?</header>`)
	html = headerRe.ReplaceAllString(html, "<!-- HEADER_PLACEHOLDER -->")

	// Insert new header after <body>
	bodyRe := regexp.MustCompile(`(?i)(<body[^>]*>)`)
	if bodyRe.MatchString(html) {
		html = bodyRe.ReplaceAllString(html, "$1\n"+headerHTML)
		html = strings.Replace(html, "<!-- HEADER_PLACEHOLDER -->", "", 1)
	} else {
		html = strings.Replace(html, "<!-- HEADER_PLACEHOLDER -->", headerHTML, 1)
	}

	return html
}
