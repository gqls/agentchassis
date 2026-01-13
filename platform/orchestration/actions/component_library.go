// ===========================================================================
// UNIFIED COMPONENT LIBRARY
// FILE: platform/orchestration/actions/component_library.go
// ===========================================================================
// Shared code for component loading, rendering, and theming.
// Used by:
//   - assemble_from_library.go (full page assembly)
//   - Header/footer injection in multipage assembly
// ===========================================================================

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// ===========================================================================
// CORE TYPES
// ===========================================================================

// Component represents a content component from the database
type Component struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Function     string                 `json:"function"`
	Category     string                 `json:"category"`
	HTMLTemplate string                 `json:"html_template"`
	InputSchema  map[string]interface{} `json:"input_schema"`
}

// StyleCollection bundles components + colors for a site
type StyleCollection struct {
	ID                uuid.UUID         `json:"id"`
	Name              string            `json:"name"`
	DisplayName       string            `json:"display_name"`
	HeaderComponentID *uuid.UUID        `json:"header_component_id"`
	FooterComponentID *uuid.UUID        `json:"footer_component_id"`
	CSSThemeID        *uuid.UUID        `json:"css_theme_id"`
	ColorPalette      map[string]string `json:"color_palette"`
	Typography        map[string]string `json:"typography"`
	Category          string            `json:"category"`
	IndustryTags      []string          `json:"industry_tags"`
}

// Theme represents a CSS theme
type Theme struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	CSSContent   string            `json:"css_content"`
	ColorPalette map[string]string `json:"color_palette"`
}

// RenderContext holds all data needed to render components
type RenderContext struct {
	// Site info
	Domain      string `json:"domain"`
	SiteID      uuid.UUID
	LogoText    string `json:"logo_text"`
	CompanyName string `json:"company_name"`
	Tagline     string `json:"tagline"`

	// Navigation
	NavItems    []NavItem `json:"nav_items"`
	CurrentPage string    `json:"current_page"`

	// Colors (from style collection or extracted from brief)
	PrimaryColor    string `json:"primary_color"`
	SecondaryColor  string `json:"secondary_color"`
	AccentColor     string `json:"accent_color"`
	TextColor       string `json:"text_color"`
	BackgroundColor string `json:"background_color"`

	// Theme CSS (for full page assembly)
	ThemeCSS string `json:"theme_css"`

	// Page-specific
	Title       string `json:"title"`
	Description string `json:"description"`

	// Contact
	Email string `json:"email"`
	Phone string `json:"phone"`

	// CTA
	CTAText string `json:"cta_text"`
	CTAUrl  string `json:"cta_url"`

	// Metadata
	Year string `json:"year"`

	// Content generation
	Industry       string
	Tone           string
	TargetAudience string
	Services       []string

	// ContentData holds arbitrary content fields from LLM generation
	// Examples: headline, subheadline, features[], testimonials[], body, etc.
	// These flow through to template substitution
	ContentData map[string]interface{} `json:"content_data"`

	// SchemaMode controls validation strictness
	// "flexible" (default): best-effort rendering, warn on missing fields
	// "strict": fail if content doesn't match component's input_schema
	SchemaMode string `json:"schema_mode"`

	// SchemaSnapshot is the locked input_schema (only used in strict mode)
	SchemaSnapshot map[string]interface{} `json:"schema_snapshot,omitempty"`
}

// RenderOptions controls rendering behavior
type RenderOptions struct {
	SchemaMode     string                 // "flexible" or "strict"
	SchemaSnapshot map[string]interface{} // Locked schema (for strict mode)
	Logger         *zap.Logger
}

// NavItem represents a navigation link
type NavItem struct {
	Label    string `json:"label"`
	URL      string `json:"url"`
	IsActive bool   `json:"is_active"`
}

// ===========================================================================
// DATABASE QUERIES - Generic interface for sql.DB and pgxpool.Pool
// ===========================================================================

// dbQuerier abstracts database query methods
type dbQuerier interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// queryRow handles both *sql.DB and *pgxpool.Pool
func queryRow(ctx context.Context, db interface{}, query string, args ...interface{}) (interface{}, error) {
	switch d := db.(type) {
	case *sql.DB:
		return d.QueryRowContext(ctx, query, args...), nil
	case *pgxpool.Pool:
		return d.QueryRow(ctx, query, args...), nil
	default:
		return nil, fmt.Errorf("unsupported database type: %T", db)
	}
}

// ===========================================================================
// COMPONENT QUERIES
// ===========================================================================

// GetComponentByFunction retrieves a component by its function name
func GetComponentByFunction(ctx context.Context, db interface{}, function string, logger *zap.Logger) (*Component, error) {
	query := `
		SELECT 
			id, 
			name, 
			function, 
			COALESCE(category, '') as category,  -- Handle NULL category
			html_template, 
			input_schema
		FROM content_components
		WHERE function = $1 AND is_active = true
		LIMIT 1
	`
	return queryComponent(ctx, db, query, function, logger)
}

// GetComponentByName retrieves a component by its name
func GetComponentByName(ctx context.Context, db interface{}, name string, logger *zap.Logger) (*Component, error) {
	query := `
		SELECT id, name, "function", category, html_template, input_schema
		FROM content_components
		WHERE name = $1
		LIMIT 1
	`
	return queryComponent(ctx, db, query, name, logger)
}

// GetComponentByID retrieves a component by its UUID
func GetComponentByID(ctx context.Context, db interface{}, id uuid.UUID, logger *zap.Logger) (*Component, error) {
	query := `
		SELECT 
			id, 
			name, 
			function, 
			COALESCE(category, '') as category,  -- Handle NULL category
			html_template, 
			input_schema
		FROM content_components
		WHERE id = $1
		LIMIT 1
	`
	return queryComponent(ctx, db, query, id, logger)
}

// queryComponent executes a component query
func queryComponent(ctx context.Context, db interface{}, query string, arg interface{}, logger *zap.Logger) (*Component, error) {
	var comp Component
	var schemaJSON []byte
	var category sql.NullString // Use NullString for nullable field

	var err error
	switch d := db.(type) {
	case *sql.DB:
		err = d.QueryRowContext(ctx, query, arg).Scan(
			&comp.ID, &comp.Name, &comp.Function, &category,
			&comp.HTMLTemplate, &schemaJSON,
		)
	case *pgxpool.Pool:
		err = d.QueryRow(ctx, query, arg).Scan(
			&comp.ID, &comp.Name, &comp.Function, &category,
			&comp.HTMLTemplate, &schemaJSON,
		)
	default:
		return nil, fmt.Errorf("unsupported database type: %T", db)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("component not found: %v", arg)
		}
		return nil, fmt.Errorf("failed to query component: %w", err)
	}

	// Handle nullable category
	if category.Valid {
		comp.Category = category.String
	} else {
		comp.Category = "" // Default to empty string
	}

	if len(schemaJSON) > 0 {
		json.Unmarshal(schemaJSON, &comp.InputSchema)
	}

	return &comp, nil
}

// GetComponentWithFallback tries to get a component, falling back to generic
func GetComponentWithFallback(ctx context.Context, db interface{}, function string, logger *zap.Logger) (*Component, error) {
	comp, err := GetComponentByFunction(ctx, db, function, logger)
	if err == nil {
		return comp, nil
	}

	logger.Warn("Component not found, using fallback",
		zap.String("requested", function),
		zap.String("fallback", "generic-text-block"))

	return GetComponentByFunction(ctx, db, "generic-text-block", logger)
}

// ===========================================================================
// THEME QUERIES
// ===========================================================================

// GetThemeByName retrieves a CSS theme by name
func GetThemeByName(ctx context.Context, db interface{}, name string, logger *zap.Logger) (*Theme, error) {
	query := `
		SELECT id, name, css_content, color_palette
		FROM css_themes
		WHERE name = $1 AND is_active = true
		LIMIT 1
	`

	var theme Theme
	var colorPaletteJSON []byte

	var err error
	switch d := db.(type) {
	case *sql.DB:
		err = d.QueryRowContext(ctx, query, name).Scan(
			&theme.ID, &theme.Name, &theme.CSSContent, &colorPaletteJSON,
		)
	case *pgxpool.Pool:
		err = d.QueryRow(ctx, query, name).Scan(
			&theme.ID, &theme.Name, &theme.CSSContent, &colorPaletteJSON,
		)
	default:
		return nil, fmt.Errorf("unsupported database type: %T", db)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			// Try default theme
			if name != "default" {
				logger.Warn("Theme not found, trying default", zap.String("requested", name))
				return GetThemeByName(ctx, db, "default", logger)
			}
			return nil, fmt.Errorf("theme not found: %s", name)
		}
		return nil, fmt.Errorf("failed to query theme: %w", err)
	}

	if len(colorPaletteJSON) > 0 {
		json.Unmarshal(colorPaletteJSON, &theme.ColorPalette)
	}

	return &theme, nil
}

// ===========================================================================
// STYLE COLLECTION QUERIES
// ===========================================================================

// GetStyleCollectionForSite retrieves the style collection assigned to a site
func GetStyleCollectionForSite(ctx context.Context, db interface{}, siteID uuid.UUID, logger *zap.Logger) (*StyleCollection, error) {
	query := `
		SELECT 
			sc.id, sc.name, sc.display_name,
			sc.header_component_id, sc.footer_component_id, sc.css_theme_id,
			sc.color_palette, sc.typography, sc.category, sc.industry_tags
		FROM sites s
		JOIN style_collections sc ON s.style_collection_id = sc.id
		WHERE s.id = $1
	`

	var coll StyleCollection
	var colorPaletteJSON, typographyJSON []byte
	var industryTags []string

	var err error
	switch d := db.(type) {
	case *sql.DB:
		err = d.QueryRowContext(ctx, query, siteID).Scan(
			&coll.ID, &coll.Name, &coll.DisplayName,
			&coll.HeaderComponentID, &coll.FooterComponentID, &coll.CSSThemeID,
			&colorPaletteJSON, &typographyJSON, &coll.Category, &industryTags,
		)
	case *pgxpool.Pool:
		err = d.QueryRow(ctx, query, siteID).Scan(
			&coll.ID, &coll.Name, &coll.DisplayName,
			&coll.HeaderComponentID, &coll.FooterComponentID, &coll.CSSThemeID,
			&colorPaletteJSON, &typographyJSON, &coll.Category, &industryTags,
		)
	default:
		return nil, fmt.Errorf("unsupported database type: %T", db)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No collection assigned
		}
		return nil, fmt.Errorf("failed to query style collection: %w", err)
	}

	if len(colorPaletteJSON) > 0 {
		json.Unmarshal(colorPaletteJSON, &coll.ColorPalette)
	}
	if len(typographyJSON) > 0 {
		json.Unmarshal(typographyJSON, &coll.Typography)
	}
	coll.IndustryTags = industryTags

	return &coll, nil
}

// GetStyleCollectionByName retrieves a style collection by name
func GetStyleCollectionByName(ctx context.Context, db interface{}, name string, logger *zap.Logger) (*StyleCollection, error) {
	query := `
		SELECT 
			id, name, display_name,
			header_component_id, footer_component_id, css_theme_id,
			color_palette, typography, category, industry_tags
		FROM style_collections
		WHERE name = $1 AND is_active = true
		LIMIT 1
	`

	var coll StyleCollection
	var colorPaletteJSON, typographyJSON []byte
	var industryTags []string

	var err error
	switch d := db.(type) {
	case *sql.DB:
		err = d.QueryRowContext(ctx, query, name).Scan(
			&coll.ID, &coll.Name, &coll.DisplayName,
			&coll.HeaderComponentID, &coll.FooterComponentID, &coll.CSSThemeID,
			&colorPaletteJSON, &typographyJSON, &coll.Category, &industryTags,
		)
	case *pgxpool.Pool:
		err = d.QueryRow(ctx, query, name).Scan(
			&coll.ID, &coll.Name, &coll.DisplayName,
			&coll.HeaderComponentID, &coll.FooterComponentID, &coll.CSSThemeID,
			&colorPaletteJSON, &typographyJSON, &coll.Category, &industryTags,
		)
	default:
		return nil, fmt.Errorf("unsupported database type: %T", db)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("style collection not found: %s", name)
		}
		return nil, fmt.Errorf("failed to query style collection: %w", err)
	}

	if len(colorPaletteJSON) > 0 {
		json.Unmarshal(colorPaletteJSON, &coll.ColorPalette)
	}
	if len(typographyJSON) > 0 {
		json.Unmarshal(typographyJSON, &coll.Typography)
	}
	coll.IndustryTags = industryTags

	return &coll, nil
}

// SelectStyleCollectionByDomain chooses a style collection based on domain keywords
// This replaces the old selectTheme function but returns a full style collection
func SelectStyleCollectionByDomain(ctx context.Context, db interface{}, domain string, logger *zap.Logger) (*StyleCollection, error) {
	domainLower := strings.ToLower(domain)

	// Map domain keywords to style collection names
	var collectionName string

	switch {
	case containsAny(domainLower, "tech", "software", "app", "ai", "cloud", "dev", "code", "data", "cyber", "saas"):
		collectionName = "bold-gradient"
	case containsAny(domainLower, "law", "legal", "finance", "invest", "consult", "advisor", "capital", "bank"):
		collectionName = "professional-dark"
	case containsAny(domainLower, "design", "creative", "studio", "agency", "portfolio", "art"):
		collectionName = "minimal-light"
	case containsAny(domainLower, "box", "fight", "sport", "gym", "fitness", "martial"):
		collectionName = "bold-gradient" // energetic
	case containsAny(domainLower, "bak", "food", "cafe", "restaurant", "cook", "chef", "bistro"):
		collectionName = "minimal-light" // clean for food
	default:
		collectionName = "professional-dark"
	}

	logger.Info("Selected style collection by domain",
		zap.String("domain", domain),
		zap.String("collection", collectionName))

	coll, err := GetStyleCollectionByName(ctx, db, collectionName, logger)
	if err != nil {
		// Fallback to professional-dark
		logger.Warn("Style collection not found, using professional-dark",
			zap.String("requested", collectionName))
		return GetStyleCollectionByName(ctx, db, "professional-dark", logger)
	}

	return coll, nil
}

// containsAny checks if s contains any of the substrings
func containsAny(s string, substrings ...string) bool {
	for _, sub := range substrings {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// ===========================================================================
// TEMPLATE RENDERING - Unified for both Go-style and Handlebars-style
// ===========================================================================

// RenderTemplate renders a component template with the given context
// Supports both {{.field}} (Go-style) and {{field}} (Handlebars-style)
func RenderTemplate(template string, ctx *RenderContext, logger *zap.Logger) string {
	result := template

	// Convert context to map for flexible access
	data := contextToMap(ctx)

	// Step 1: Handle {{#each nav_items}}...{{/each}} blocks
	result = renderEachBlocks(result, ctx.NavItems)

	// Step 2: Handle {{#if field}}...{{/if}} blocks
	result = renderIfBlocks(result, data)

	// Step 3: Handle Go-style {{.field}} substitutions
	result = renderGoStyleSubstitutions(result, data)

	// Step 4: Handle Handlebars-style {{field}} substitutions
	result = renderHandlebarsSubstitutions(result, data)

	// Step 5: Build nav items HTML for {{nav_items_html}} placeholder
	navItemsHTML := buildNavItemsHTML(ctx.NavItems)
	result = strings.ReplaceAll(result, "{{nav_items_html}}", navItemsHTML)
	result = strings.ReplaceAll(result, "{{.nav_items_html}}", navItemsHTML)

	return result
}

// contextToMap converts RenderContext to a map for template substitution
// Includes field aliasing to handle common naming variations
func contextToMap(ctx *RenderContext) map[string]string {
	if ctx.Year == "" {
		ctx.Year = fmt.Sprintf("%d", time.Now().Year())
	}

	// Fallback for logo text - extract from domain if empty
	logoText := ctx.LogoText
	if logoText == "" && ctx.CompanyName != "" {
		logoText = ctx.CompanyName
	}
	if logoText == "" && ctx.Domain != "" {
		// Extract from domain: "leopardessconsulting.co.uk" -> "Leopardessconsulting"
		parts := strings.Split(ctx.Domain, ".")
		if len(parts) > 0 && len(parts[0]) > 0 {
			name := parts[0]
			logoText = strings.ToUpper(name[:1]) + name[1:]
		}
	}
	if logoText == "" {
		logoText = "Company"
	}

	result := map[string]string{
		"domain":           ctx.Domain,
		"logo_text":        logoText,
		"company_name":     defaultString(ctx.CompanyName, logoText),
		"tagline":          ctx.Tagline,
		"current_page":     ctx.CurrentPage,
		"primary_color":    defaultString(ctx.PrimaryColor, "#1a1a2e"),
		"secondary_color":  defaultString(ctx.SecondaryColor, "#2d2d44"),
		"accent_color":     defaultString(ctx.AccentColor, "#16a085"),
		"text_color":       defaultString(ctx.TextColor, "#333333"),
		"background_color": defaultString(ctx.BackgroundColor, "#ffffff"),
		"theme_css":        ctx.ThemeCSS,
		"title":            ctx.Title,
		"description":      ctx.Description,
		"email":            ctx.Email,
		"contact_email":    ctx.Email,
		"phone":            ctx.Phone,
		"cta_text":         defaultString(ctx.CTAText, "Get Started"),
		"cta_url":          defaultString(ctx.CTAUrl, "/contact.html"),
		"year":             ctx.Year,
		"industry":         ctx.Industry,
		"tone":             ctx.Tone,
		"target_audience":  ctx.TargetAudience,
	}

	// Add all content data fields
	for key, value := range ctx.ContentData {
		// Don't override known fields - they have priority
		if _, exists := result[key]; exists {
			continue
		}
		result[key] = datahelpers.InterfaceToString(value)
	}

	// =========================================================================
	// FIELD ALIASING - Map common variations to expected template names
	// =========================================================================
	aliases := map[string][]string{
		// CTA variations
		"primary_cta":       {"cta_text", "cta", "button_text", "action_text"},
		"primary_cta_url":   {"cta_url", "cta_link", "button_url", "action_url"},
		"secondary_cta":     {"secondary_button", "alt_cta", "secondary_text"},
		"secondary_cta_url": {"secondary_url", "alt_cta_url", "secondary_link"},

		// Content variations
		"subheadline": {"subtitle", "sub_headline", "lead"},
		"headline":    {"main_title", "header"},
		"body":        {"content", "text", "paragraph"},
		"heading":     {"section_title", "section_heading"},
	}

	// Apply aliases - if target field is empty, try source fields
	for targetField, sourceFields := range aliases {
		// Skip if target already has a value
		if result[targetField] != "" {
			continue
		}
		// Try each source field
		for _, sourceField := range sourceFields {
			if val, exists := result[sourceField]; exists && val != "" {
				result[targetField] = val
				break
			}
		}
	}

	// =========================================================================
	// DEFAULT VALUES for common template fields to prevent raw {{.field}}
	// =========================================================================
	defaults := map[string]string{
		"primary_cta":       "Get Started",
		"primary_cta_url":   "/contact.html",
		"secondary_cta":     "Learn More",
		"secondary_cta_url": "/about.html",
	}

	for field, defaultVal := range defaults {
		if result[field] == "" {
			result[field] = defaultVal
		}
	}

	return result
}

// RenderTemplateWithValidation renders a template with optional schema validation
func RenderTemplateWithValidation(
	template string,
	ctx *RenderContext,
	opts RenderOptions,
) (string, error) {

	logger := opts.Logger
	if logger == nil {
		logger = zap.NewNop()
	}

	// Default to flexible mode
	schemaMode := opts.SchemaMode
	if schemaMode == "" {
		schemaMode = "flexible"
	}

	// Convert context to template data
	data := contextToMap(ctx)

	// In strict mode, validate against schema
	if schemaMode == "strict" && opts.SchemaSnapshot != nil {
		if err := validateContentAgainstSchema(data, opts.SchemaSnapshot, logger); err != nil {
			return "", fmt.Errorf("schema validation failed: %w", err)
		}
	}

	// Perform template substitution
	result := template
	result = renderEachBlocks(result, ctx.NavItems)
	result = renderIfBlocks(result, data)
	result = renderGoStyleSubstitutions(result, data)
	result = renderHandlebarsSubstitutions(result, data)

	// Build nav items HTML
	navItemsHTML := buildNavItemsHTML(ctx.NavItems)
	result = strings.ReplaceAll(result, "{{nav_items_html}}", navItemsHTML)
	result = strings.ReplaceAll(result, "{{.nav_items_html}}", navItemsHTML)

	// Check for unsubstituted placeholders
	unsubstituted := findUnsubstitutedPlaceholders(result)
	if len(unsubstituted) > 0 {
		if schemaMode == "strict" {
			return "", fmt.Errorf("unsubstituted placeholders in strict mode: %v", unsubstituted)
		}
		// In flexible mode, just warn
		logger.Warn("Template has unsubstituted placeholders (flexible mode)",
			zap.Strings("placeholders", unsubstituted),
			zap.Int("count", len(unsubstituted)))
	}

	return result, nil
}

// validateContentAgainstSchema checks if data has all required fields from schema
func validateContentAgainstSchema(data map[string]string, schema map[string]interface{}, logger *zap.Logger) error {
	// Check required fields
	if required, ok := schema["required"].([]interface{}); ok {
		var missing []string
		for _, field := range required {
			fieldName, ok := field.(string)
			if !ok {
				continue
			}
			if val, exists := data[fieldName]; !exists || val == "" {
				missing = append(missing, fieldName)
			}
		}
		if len(missing) > 0 {
			return fmt.Errorf("missing required fields: %v", missing)
		}
	}

	return nil
}

// findUnsubstitutedPlaceholders finds any remaining {{...}} or {{.xxx}} in template
func findUnsubstitutedPlaceholders(template string) []string {
	var placeholders []string

	// Simple pattern matching for remaining placeholders
	// This catches both {{field}} and {{.field}} patterns
	inPlaceholder := false
	start := 0

	for i := 0; i < len(template)-1; i++ {
		if template[i] == '{' && template[i+1] == '{' {
			inPlaceholder = true
			start = i
		} else if inPlaceholder && template[i] == '}' && template[i+1] == '}' {
			placeholder := template[start : i+2]
			// Skip block helpers ({{#if}}, {{/if}}, {{#each}}, etc.)
			if !strings.Contains(placeholder, "#") && !strings.Contains(placeholder, "/") {
				placeholders = append(placeholders, placeholder)
			}
			inPlaceholder = false
		}
	}

	return placeholders
}

// renderEachBlocks handles {{#each nav_items}}...{{/each}}
func renderEachBlocks(template string, navItems []NavItem) string {
	eachRe := regexp.MustCompile(`(?s)\{\{#each\s+nav_items\}\}(.*?)\{\{/each\}\}`)

	return eachRe.ReplaceAllStringFunc(template, func(match string) string {
		matches := eachRe.FindStringSubmatch(match)
		if len(matches) < 2 {
			return match
		}

		itemTemplate := matches[1]
		var result strings.Builder

		for _, item := range navItems {
			itemStr := itemTemplate
			itemStr = strings.ReplaceAll(itemStr, "{{this.url}}", item.URL)
			itemStr = strings.ReplaceAll(itemStr, "{{this.label}}", item.Label)

			// Handle {{#if this.is_active}}...{{/if}}
			activeRe := regexp.MustCompile(`(?s)\{\{#if\s+this\.is_active\}\}(.*?)\{\{/if\}\}`)
			itemStr = activeRe.ReplaceAllStringFunc(itemStr, func(m string) string {
				matches := activeRe.FindStringSubmatch(m)
				if len(matches) < 2 {
					return m
				}
				if item.IsActive {
					return matches[1]
				}
				return ""
			})

			result.WriteString(itemStr)
		}

		return result.String()
	})
}

// renderIfBlocks handles {{#if field}}...{{/if}}
func renderIfBlocks(template string, data map[string]string) string {
	ifRe := regexp.MustCompile(`(?s)\{\{#if\s+(\w+)\}\}(.*?)\{\{/if\}\}`)

	return ifRe.ReplaceAllStringFunc(template, func(match string) string {
		matches := ifRe.FindStringSubmatch(match)
		if len(matches) < 3 {
			return match
		}

		fieldName := matches[1]
		content := matches[2]

		value, exists := data[fieldName]
		if !exists || value == "" {
			return ""
		}
		return content
	})
}

// renderGoStyleSubstitutions handles {{.field}} placeholders
func renderGoStyleSubstitutions(template string, data map[string]string) string {
	goRe := regexp.MustCompile(`\{\{\.(\w+)\}\}`)

	return goRe.ReplaceAllStringFunc(template, func(match string) string {
		matches := goRe.FindStringSubmatch(match)
		if len(matches) < 2 {
			return match
		}

		fieldName := matches[1]
		if value, ok := data[fieldName]; ok {
			return value
		}
		return match // Keep placeholder if no value
	})
}

// renderHandlebarsSubstitutions handles {{field}} placeholders
func renderHandlebarsSubstitutions(template string, data map[string]string) string {
	// Match all {{...}} patterns, then filter in the replacement function
	// Go's regexp doesn't support negative lookahead, so we check manually
	hbRe := regexp.MustCompile(`\{\{([^}]+)\}\}`)

	return hbRe.ReplaceAllStringFunc(template, func(match string) string {
		// Extract content between {{ and }}
		inner := match[2 : len(match)-2]

		// Skip special patterns: #if, /if, #each, /each, this.field
		if strings.HasPrefix(inner, "#") ||
			strings.HasPrefix(inner, "/") ||
			strings.HasPrefix(inner, "this.") {
			return match
		}

		// Skip if contains spaces (likely a block expression we missed)
		if strings.Contains(inner, " ") {
			return match
		}

		// Look up simple field name
		fieldName := strings.TrimSpace(inner)
		if value, ok := data[fieldName]; ok {
			return value
		}
		return match
	})
}

// buildNavItemsHTML creates pre-rendered nav items HTML
func buildNavItemsHTML(items []NavItem) string {
	var parts []string
	for _, item := range items {
		activeClass := ""
		if item.IsActive {
			activeClass = ` class="active"`
		}
		// Simplify the label at render time (defense in depth)
		label := simplifyNavLabelForRender(item.Label, item.URL)
		parts = append(parts, fmt.Sprintf(
			`<li><a href="%s"%s>%s</a></li>`,
			item.URL, activeClass, label,
		))
	}
	return strings.Join(parts, "\n                ")
}

// simplifyNavLabelForRender creates a clean nav label
// Handles cases like "About Us | Leopardess Consulting" -> "About"
func simplifyNavLabelForRender(label, url string) string {
	// Strip "|" and everything after
	if idx := strings.Index(label, "|"); idx > 0 {
		label = strings.TrimSpace(label[:idx])
	}

	// Strip " - " and everything after
	if idx := strings.Index(label, " - "); idx > 0 {
		label = strings.TrimSpace(label[:idx])
	}

	// Extract page name from URL for mapping
	pageName := strings.TrimSuffix(strings.TrimPrefix(url, "/"), ".html")
	pageNameLower := strings.ToLower(pageName)

	// Simple labels by page name
	simpleLabels := map[string]string{
		"index":     "Home",
		"home":      "Home",
		"about":     "About",
		"services":  "Services",
		"contact":   "Contact",
		"insights":  "Insights",
		"blog":      "Blog",
		"careers":   "Careers",
		"team":      "Team",
		"pricing":   "Pricing",
		"faq":       "FAQ",
		"support":   "Support",
		"features":  "Features",
		"products":  "Products",
		"portfolio": "Portfolio",
		"work":      "Work",
		"clients":   "Clients",
		"resources": "Resources",
	}

	// First check if page name matches a known simple label
	if simple, ok := simpleLabels[pageNameLower]; ok {
		return simple
	}

	// Check if the label starts with a known nav word
	labelLower := strings.ToLower(label)
	for pagePart, simple := range simpleLabels {
		if strings.HasPrefix(labelLower, pagePart) {
			return simple
		}
	}

	// Return cleaned label (first meaningful part)
	words := strings.Fields(label)
	if len(words) >= 1 && len(words) <= 3 {
		return label // Already simple enough
	}
	if len(words) > 3 {
		// Take first 2 words if they make sense
		return strings.Join(words[:2], " ")
	}

	return label
}

// defaultString returns the default if s is empty
func defaultString(s, defaultVal string) string {
	if s == "" {
		return defaultVal
	}
	return s
}

// ===========================================================================
// HIGH-LEVEL RENDERING FUNCTIONS
// ===========================================================================

// RenderHeader renders the header component for a site
func RenderHeader(ctx context.Context, db interface{}, siteID uuid.UUID, renderCtx *RenderContext, logger *zap.Logger) (string, error) {
	// Try to get site's style collection
	var coll *StyleCollection
	var err error
	var source string = "fallback"

	if siteID != uuid.Nil {
		coll, err = GetStyleCollectionForSite(ctx, db, siteID, logger)
		if err != nil {
			logger.Warn("Failed to get style collection", zap.Error(err))
		}
	}

	// Fallback: select by domain
	if coll == nil && renderCtx.Domain != "" {
		coll, err = SelectStyleCollectionByDomain(ctx, db, renderCtx.Domain, logger)
		if err != nil {
			logger.Warn("Failed to select style collection by domain", zap.Error(err))
		}
	}

	// Apply colors from collection
	if coll != nil && coll.ColorPalette != nil {
		if renderCtx.PrimaryColor == "" {
			renderCtx.PrimaryColor = coll.ColorPalette["primary"]
		}
		if renderCtx.AccentColor == "" {
			renderCtx.AccentColor = coll.ColorPalette["accent"]
		}
		if renderCtx.SecondaryColor == "" {
			renderCtx.SecondaryColor = coll.ColorPalette["secondary"]
		}
	}

	// Get header component
	var comp *Component
	if coll != nil && coll.HeaderComponentID != nil {
		comp, err = GetComponentByID(ctx, db, *coll.HeaderComponentID, logger)
		if err != nil {
			logger.Warn("Failed to get header component", zap.Error(err))
		} else {
			source = fmt.Sprintf("component-db:%s", coll.Name)
		}
	}

	// Fallback: try by function name
	if comp == nil {
		comp, err = GetComponentByFunction(ctx, db, "site-header", logger)
		if err != nil {
			logger.Warn("No header component found, using fallback")
			header := RenderFallbackHeader(renderCtx)
			return fmt.Sprintf("<!-- HEADER SOURCE: fallback -->\n%s", header), nil
		}
		source = "component-db:site-header"
	}

	// Render template
	rendered := RenderTemplate(comp.HTMLTemplate, renderCtx, logger)
	return fmt.Sprintf("<!-- HEADER SOURCE: %s -->\n%s", source, rendered), nil
}

// RenderFooter renders the footer component for a site
func RenderFooter(ctx context.Context, db interface{}, siteID uuid.UUID, renderCtx *RenderContext, logger *zap.Logger) (string, error) {
	var coll *StyleCollection
	var err error
	var source string = "fallback"

	if siteID != uuid.Nil {
		coll, err = GetStyleCollectionForSite(ctx, db, siteID, logger)
		if err != nil {
			logger.Warn("Failed to get style collection", zap.Error(err))
		}
	}

	if coll == nil && renderCtx.Domain != "" {
		coll, err = SelectStyleCollectionByDomain(ctx, db, renderCtx.Domain, logger)
	}

	if coll != nil && coll.ColorPalette != nil {
		if renderCtx.PrimaryColor == "" {
			renderCtx.PrimaryColor = coll.ColorPalette["primary"]
		}
		if renderCtx.AccentColor == "" {
			renderCtx.AccentColor = coll.ColorPalette["accent"]
		}
	}

	var comp *Component
	if coll != nil && coll.FooterComponentID != nil {
		comp, err = GetComponentByID(ctx, db, *coll.FooterComponentID, logger)
		if err != nil {
			logger.Warn("Failed to get footer component", zap.Error(err))
		} else {
			source = fmt.Sprintf("component-db:%s", coll.Name)
		}
	}

	if comp == nil {
		comp, err = GetComponentByFunction(ctx, db, "site-footer", logger)
		if err != nil {
			footer := RenderFallbackFooter(renderCtx)
			return fmt.Sprintf("<!-- FOOTER SOURCE: fallback -->\n%s", footer), nil
		}
		source = "component-db:site-footer"
	}

	rendered := RenderTemplate(comp.HTMLTemplate, renderCtx, logger)
	return fmt.Sprintf("<!-- FOOTER SOURCE: %s -->\n%s", source, rendered), nil
}

// RenderFallbackHeader creates a basic header when no component is available
func RenderFallbackHeader(ctx *RenderContext) string {
	navHTML := buildNavItemsHTML(ctx.NavItems)
	primary := defaultString(ctx.PrimaryColor, "#1a1a2e")
	accent := defaultString(ctx.AccentColor, "#16a085")

	return fmt.Sprintf(`<header class="site-header">
    <div class="header-container">
        <a href="/index.html" class="logo">%s</a>
        <button class="mobile-menu-toggle" aria-label="Toggle menu"><span></span><span></span><span></span></button>
        <nav class="main-nav">
            <ul>%s</ul>
        </nav>
    </div>
</header>
<style>
.site-header{background:%s;padding:1rem 0;position:sticky;top:0;z-index:1000;box-shadow:0 2px 10px rgba(0,0,0,.1)}
.header-container{max-width:1200px;margin:0 auto;padding:0 2rem;display:flex;align-items:center;justify-content:space-between}
.logo{text-decoration:none;font-size:1.5rem;font-weight:700;color:#fff}
.main-nav ul{display:flex;list-style:none;margin:0;padding:0;gap:2rem}
.main-nav a{color:rgba(255,255,255,.9);text-decoration:none;font-weight:500;transition:color .2s}
.main-nav a:hover,.main-nav a.active{color:%s}
.mobile-menu-toggle{display:none;background:none;border:none;cursor:pointer;padding:.5rem}
.mobile-menu-toggle span{display:block;width:24px;height:2px;background:#fff;margin:5px 0}
@media(max-width:768px){.mobile-menu-toggle{display:block}.main-nav{position:absolute;top:100%%;left:0;right:0;background:%s;padding:1rem;display:none}.main-nav.active{display:block}.main-nav ul{flex-direction:column;gap:0}.main-nav a{display:block;padding:.75rem 0;border-bottom:1px solid rgba(255,255,255,.1)}}
</style>
<script>document.addEventListener("DOMContentLoaded",function(){var t=document.querySelector(".mobile-menu-toggle"),n=document.querySelector(".main-nav");t&&n&&t.addEventListener("click",function(){n.classList.toggle("active")})});</script>`,
		ctx.LogoText, navHTML, primary, accent, primary)
}

// RenderFallbackFooter creates a basic footer when no component is available
func RenderFallbackFooter(ctx *RenderContext) string {
	navHTML := buildNavItemsHTML(ctx.NavItems)
	primary := defaultString(ctx.PrimaryColor, "#1a1a2e")
	year := ctx.Year
	if year == "" {
		year = fmt.Sprintf("%d", time.Now().Year())
	}

	return fmt.Sprintf(`<footer class="site-footer">
    <div class="footer-container">
        <div class="footer-brand"><h3>%s</h3><p>%s</p></div>
        <div class="footer-links"><h4>Links</h4><ul>%s</ul></div>
        <div class="footer-contact"><h4>Contact</h4><p>%s</p></div>
    </div>
    <div class="footer-bottom"><p>&copy; %s %s. All rights reserved.</p></div>
</footer>
<style>
.site-footer{background:%s;color:rgba(255,255,255,.9);padding:3rem 0 0;margin-top:auto}
.footer-container{max-width:1200px;margin:0 auto;padding:0 2rem;display:grid;grid-template-columns:2fr 1fr 1fr;gap:2rem}
.footer-brand h3,.footer-links h4,.footer-contact h4{color:#fff;margin:0 0 1rem}
.footer-links ul{list-style:none;padding:0;margin:0}
.footer-links li{margin-bottom:.5rem}
.footer-links a{color:rgba(255,255,255,.7);text-decoration:none}
.footer-links a:hover{color:#fff}
.footer-bottom{margin-top:2rem;padding:1.5rem 0;border-top:1px solid rgba(255,255,255,.1);text-align:center}
.footer-bottom p{margin:0;color:rgba(255,255,255,.6);font-size:.9rem}
@media(max-width:768px){.footer-container{grid-template-columns:1fr}}
</style>`, ctx.LogoText, ctx.Tagline, navHTML, ctx.Email, year, ctx.CompanyName, primary)
}

// ===========================================================================
// HTML INJECTION FUNCTIONS
// ===========================================================================

// InjectHeader replaces an existing header in HTML with a rendered component
func InjectHeader(ctx context.Context, db interface{}, html string, siteID uuid.UUID, renderCtx *RenderContext, logger *zap.Logger) string {
	headerHTML, err := RenderHeader(ctx, db, siteID, renderCtx, logger)
	if err != nil {
		logger.Warn("Failed to render header", zap.Error(err))
		headerHTML = RenderFallbackHeader(renderCtx)
	}

	// Remove existing header
	headerRe := regexp.MustCompile(`(?is)<header[^>]*>.*?</header>`)
	html = headerRe.ReplaceAllString(html, "<!-- HEADER_REPLACED -->")

	// Insert after <body>
	bodyRe := regexp.MustCompile(`(?i)(<body[^>]*>)`)
	if bodyRe.MatchString(html) {
		html = bodyRe.ReplaceAllString(html, "$1\n"+headerHTML)
		html = strings.Replace(html, "<!-- HEADER_REPLACED -->", "", 1)
	} else {
		html = strings.Replace(html, "<!-- HEADER_REPLACED -->", headerHTML, 1)
	}

	return html
}

// InjectFooter replaces an existing footer in HTML with a rendered component
func InjectFooter(ctx context.Context, db interface{}, html string, siteID uuid.UUID, renderCtx *RenderContext, logger *zap.Logger) string {
	footerHTML, err := RenderFooter(ctx, db, siteID, renderCtx, logger)
	if err != nil {
		logger.Warn("Failed to render footer", zap.Error(err))
		footerHTML = RenderFallbackFooter(renderCtx)
	}

	// Remove existing footer
	footerRe := regexp.MustCompile(`(?is)<footer[^>]*>.*?</footer>`)
	html = footerRe.ReplaceAllString(html, "<!-- FOOTER_REPLACED -->")

	// Insert before </body>
	bodyCloseRe := regexp.MustCompile(`(?i)(</body>)`)
	if bodyCloseRe.MatchString(html) {
		html = bodyCloseRe.ReplaceAllString(html, footerHTML+"\n$1")
		html = strings.Replace(html, "<!-- FOOTER_REPLACED -->", "", 1)
	} else {
		html = strings.Replace(html, "<!-- FOOTER_REPLACED -->", footerHTML, 1)
	}

	return html
}

// ===========================================================================
// BUILD METADATA (for assemble_from_library compatibility)
// ===========================================================================

// BuildThemeMetadata creates a CSS comment with build info
func BuildThemeMetadata(themeName string, componentFunctions []string, domain string) string {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	components := strings.Join(componentFunctions, ", ")

	return fmt.Sprintf(`/*
 * ============================================
 * SITE BUILD METADATA
 * ============================================
 * Theme: %s
 * Domain: %s
 * Components: %s
 * Generated: %s
 * Source: component-library
 * ============================================
 */
`, themeName, domain, components, timestamp)
}
