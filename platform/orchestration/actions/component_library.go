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
	"sort"
	"strings"
	"text/template"
	"text/template/parse"
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
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Function      string                 `json:"function"`
	Category      string                 `json:"category"`
	HTMLTemplate  string                 `json:"html_template"`
	InputSchema   map[string]interface{} `json:"input_schema"`
	IsDarkSection bool                   `json:"is_dark_section"`
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
	LogoURL     string `json:"logo_url"`
	CompanyName string `json:"company_name"`
	Tagline     string `json:"tagline"`

	// Navigation
	NavItems       []NavItem `json:"nav_items"`
	FooterNavItems []NavItem `json:"footer_nav_items"`
	CurrentPage    string    `json:"current_page"`

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
	//
	// These carry json tags like every other field because the tags are the
	// SINGLE DECLARATION of a field's template key (bugs_open/109) — see
	// renderContextScalarFields. Untagged, they were invisible to any
	// mechanical check and had to be remembered by hand in each of four maps.
	// Nothing marshals a RenderContext, so adding the tags changes no wire
	// format; it only makes the declaration complete.
	Industry       string   `json:"industry"`
	Tone           string   `json:"tone"`
	TargetAudience string   `json:"target_audience"`
	Services       []string `json:"services"`

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

// GetComponentByFunction retrieves a library/template component by its function name.
// Excludes forks (forked_from IS NOT NULL) — forks share their parent's function
// and should only be accessed by component_id through page_components.
func GetComponentByFunction(ctx context.Context, db interface{}, function string, logger *zap.Logger) (*Component, error) {
	query := `
		SELECT 
			id, 
			name, 
			function, 
			COALESCE(category, '') as category,
			html_template, 
			input_schema,
			COALESCE(is_dark_section, false) as is_dark_section
		FROM content_components
		WHERE function = $1 AND is_active = true AND forked_from IS NULL
		LIMIT 1
	`
	return queryComponent(ctx, db, query, function, logger)
}

// GetComponentByName retrieves a component by its name
func GetComponentByName(ctx context.Context, db interface{}, name string, logger *zap.Logger) (*Component, error) {
	query := `
		SELECT id, name, "function", COALESCE(category, '') as category,
			html_template, input_schema,
			COALESCE(is_dark_section, false) as is_dark_section
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
			COALESCE(category, '') as category,
			html_template, 
			input_schema,
			COALESCE(is_dark_section, false) as is_dark_section
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
			&comp.HTMLTemplate, &schemaJSON, &comp.IsDarkSection,
		)
	case *pgxpool.Pool:
		err = d.QueryRow(ctx, query, arg).Scan(
			&comp.ID, &comp.Name, &comp.Function, &category,
			&comp.HTMLTemplate, &schemaJSON, &comp.IsDarkSection,
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
	// Try exact match first
	comp, err := GetComponentByFunction(ctx, db, function, logger)
	if err == nil {
		return comp, nil
	}

	// Try normalized form (underscore â†’ hyphen, lowercase)
	normalized := NormalizeComponentFunction(function)
	if normalized != function {
		logger.Info("GetComponentWithFallback: Trying normalized function name",
			zap.String("original", function),
			zap.String("normalized", normalized),
		)
		comp, err = GetComponentByFunction(ctx, db, normalized, logger)
		if err == nil {
			return comp, nil
		}
	}

	logger.Warn("GetComponentWithFallback: Component not found, using fallback",
		zap.String("requested", function),
		zap.String("also_tried", normalized),
		zap.String("fallback", "generic-text-block"),
	)

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
	var industryTagsJSON []byte

	var err error
	switch d := db.(type) {
	case *sql.DB:
		err = d.QueryRowContext(ctx, query, siteID).Scan(
			&coll.ID, &coll.Name, &coll.DisplayName,
			&coll.HeaderComponentID, &coll.FooterComponentID, &coll.CSSThemeID,
			&colorPaletteJSON, &typographyJSON, &coll.Category, &industryTagsJSON,
		)
	case *pgxpool.Pool:
		err = d.QueryRow(ctx, query, siteID).Scan(
			&coll.ID, &coll.Name, &coll.DisplayName,
			&coll.HeaderComponentID, &coll.FooterComponentID, &coll.CSSThemeID,
			&colorPaletteJSON, &typographyJSON, &coll.Category, &industryTagsJSON,
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
	if len(industryTagsJSON) > 0 {
		json.Unmarshal(industryTagsJSON, &coll.IndustryTags)
	}

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
	var industryTagsJSON []byte

	var err error
	switch d := db.(type) {
	case *sql.DB:
		err = d.QueryRowContext(ctx, query, name).Scan(
			&coll.ID, &coll.Name, &coll.DisplayName,
			&coll.HeaderComponentID, &coll.FooterComponentID, &coll.CSSThemeID,
			&colorPaletteJSON, &typographyJSON, &coll.Category, &industryTagsJSON,
		)
	case *pgxpool.Pool:
		err = d.QueryRow(ctx, query, name).Scan(
			&coll.ID, &coll.Name, &coll.DisplayName,
			&coll.HeaderComponentID, &coll.FooterComponentID, &coll.CSSThemeID,
			&colorPaletteJSON, &typographyJSON, &coll.Category, &industryTagsJSON,
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
	if len(industryTagsJSON) > 0 {
		json.Unmarshal(industryTagsJSON, &coll.IndustryTags)
	}

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

// bareFieldRe matches a bare OUTPUT placeholder — {{.Name}} / {{ .Name }} —
// only. Used solely by the missingBareFieldsRegex FALLBACK below (when a
// template will not parse as a Go template); the primary detector is the
// scope-aware parse-tree walk in missingBareFields.
var bareFieldRe = regexp.MustCompile(`\{\{\s*\.([A-Za-z_][A-Za-z0-9_]*)\s*\}\}`)

// urlAttrRe tests whether the text immediately preceding a placeholder ends in
// an href=/src= attribute opener, in EITHER quote style with optional
// whitespace, so a blanked href="{{.x}}" or href='{{.x}}' is recognised as a
// dead control rather than a merely-missing word.
var urlAttrRe = regexp.MustCompile(`(?:href|src)\s*=\s*["']?$`)

// scanTemplateFuncs registers, by NAME only, every function a component template
// may call, so a parse-only scan succeeds on the same templates executeGoTemplate
// (call_agent.go) renders. The bodies are irrelevant here — the template is
// PARSED, never executed — but Funcs still validates each value is a function,
// so trivial stubs with the right arity are supplied. A template that calls a
// name not listed here fails to parse and falls back to the regex scan, which is
// exactly what its render path does too. Keep this list in step with
// executeGoTemplate's FuncMap.
var scanTemplateFuncs = template.FuncMap{
	"default": func(_, v interface{}) interface{} { return v },
	"eq":      func(_, _ interface{}) bool { return false },
	"ne":      func(_, _ interface{}) bool { return false },
	"lower":   strings.ToLower,
	"upper":   strings.ToUpper,
	"isset":   func(_ interface{}) bool { return false },
	"safe":    func(_ interface{}) string { return "" },
}

// missingBareFields returns the names of bare output placeholders ({{.Name}})
// that the data map will render empty, and the subset of those that sit inside
// an href=/src= URL attribute. It walks the parsed template's SYNTAX TREE and
// reports only ROOT-SCOPE bare fields, so it is scope-aware:
//   - a bare {{.Name}} nested inside {{range .Items}}…{{end}} refers to the
//     per-item dot, NOT the top-level data map, and is NOT reported;
//   - a bare {{.Name}} nested inside {{if .x}}…{{end}} (or {{with}}) is
//     author-gated — its empty case is deliberately handled — and is NOT
//     reported.
//
// Reporting either produced the false-positive Error noise the council flagged
// (a noisy channel bugs_open/054 is meant to escalate on). Only an ungated,
// root-scope bare field that renders empty is a dead control. Detection is
// deterministic and independent of the "<no value>" output string, so the same
// set is reported whichever render path ran. If the template will not parse as a
// Go template (executeGoTemplate would have used its regex fallback too), this
// degrades to the flat regex scan in missingBareFieldsRegex.
func missingBareFields(tpl string, data map[string]interface{}) (missing, inURLAttr []string) {
	t, err := template.New("scan").Funcs(scanTemplateFuncs).Option("missingkey=zero").Parse(tpl)
	if err != nil || t.Tree == nil || t.Tree.Root == nil {
		return missingBareFieldsRegex(tpl, data)
	}

	seen := map[string]bool{}
	// Walk ONLY the root node list. We deliberately do not descend into the
	// bodies of {{if}}/{{range}}/{{with}}/{{template}} nodes (they are not
	// ActionNodes), which is precisely how a field is judged root-scope-and-
	// ungated versus per-item-or-gated.
	for _, n := range t.Tree.Root.Nodes {
		an, ok := n.(*parse.ActionNode)
		if !ok {
			continue
		}
		name, isBare := bareFieldName(an.Pipe)
		if !isBare || seen[name] {
			continue
		}
		if v, present := data[name]; present && v != nil && v != "" {
			continue // field is filled — not missing
		}
		seen[name] = true
		missing = append(missing, name)
		// Attribute context: does the placeholder sit right after href=/src=?
		// an.Pos is the offset just INSIDE the action (after "{{" and any trim
		// marker/whitespace), so back up to the "{{" delimiter and test the
		// source text preceding it.
		if int(an.Pos) <= len(tpl) {
			if open := strings.LastIndex(tpl[:an.Pos], "{{"); open >= 0 && urlAttrRe.MatchString(tpl[:open]) {
				inURLAttr = append(inURLAttr, name)
			}
		}
	}
	sort.Strings(missing)
	sort.Strings(inURLAttr)
	return
}

// bareFieldName reports whether a pipe is a bare single-segment output field —
// {{.Name}} — and returns "Name". It excludes: variable declarations
// ({{$x := …}}), pipelines through functions ({{.Name | safe}}), commands with
// arguments, and nested access ({{.Foo.Bar}}, whose top-level presence says
// nothing about whether the leaf renders empty). This matches the class the old
// regex targeted, now decided structurally rather than textually.
func bareFieldName(pipe *parse.PipeNode) (string, bool) {
	if pipe == nil || len(pipe.Decl) != 0 || len(pipe.Cmds) != 1 {
		return "", false
	}
	cmd := pipe.Cmds[0]
	if len(cmd.Args) != 1 {
		return "", false
	}
	fn, ok := cmd.Args[0].(*parse.FieldNode)
	if !ok || len(fn.Ident) != 1 {
		return "", false
	}
	return fn.Ident[0], true
}

// missingBareFieldsRegex is the pre-parse-tree flat-scan fallback, used only
// when a template will not parse as a Go template. It is control-flow-blind (it
// matches {{.Name}} textually wherever it appears, including inside {{range}}/
// {{if}} bodies), so it can over-report; that is acceptable only as a
// last-resort signal for templates the structured walk cannot read.
func missingBareFieldsRegex(tpl string, data map[string]interface{}) (missing, inURLAttr []string) {
	seen := map[string]bool{}
	for _, m := range bareFieldRe.FindAllStringSubmatchIndex(tpl, -1) {
		name := tpl[m[2]:m[3]]
		if seen[name] {
			continue
		}
		if v, present := data[name]; present && v != nil && v != "" {
			continue // field is filled — not missing
		}
		seen[name] = true
		missing = append(missing, name)
		if urlAttrRe.MatchString(tpl[:m[0]]) {
			inURLAttr = append(inURLAttr, name)
		}
	}
	sort.Strings(missing)
	sort.Strings(inURLAttr)
	return
}

// RenderTemplate renders a component template with the given context.
// It keeps its original signature; every existing caller is unchanged. It is a
// thin wrapper over RenderTemplateReportingMissing, which also returns the set
// of placeholders that rendered empty (consumed by the site-chrome renderer to
// name dead controls — bugs_open/018).
func RenderTemplate(templateStr string, ctx *RenderContext, logger *zap.Logger) string {
	out, _, _ := RenderTemplateReportingMissing(templateStr, ctx, logger)
	return out
}

// RenderTemplateReportingMissing renders a component template and additionally
// reports which bare output placeholders rendered empty (`missing`) and which
// of those sat inside an href=/src= attribute (`inURLAttr` — a dead control on
// a live page). A blanked URL attribute logs at Error with its field names; any
// other blanked field logs at Warn. This replaces the previous count-only
// <no value> log, which named nothing and let 30 dead controls ship silently on
// idea.uk (bugs_open/018). Uses Go's text/template for full support of {{if}},
// {{range}}, {{with}}, etc.
func RenderTemplateReportingMissing(templateStr string, ctx *RenderContext, logger *zap.Logger) (string, []string, []string) {
	if templateStr == "" {
		return "", nil, nil
	}

	// Convert context to map[string]interface{} - preserves nested structures
	data := contextToInterfaceMap(ctx)

	// Log what we're about to render (debug)
	logger.Debug("RenderTemplate: executing",
		zap.Int("data_fields", len(data)),
		zap.String("template_preview", datahelpers.TruncateString(templateStr, 100)),
	)

	// Try Go template execution first
	result, err := executeGoTemplate(templateStr, data, logger)
	if err != nil {
		logger.Warn("Go template execution failed, using regex fallback",
			zap.Error(err),
			zap.String("template_preview", datahelpers.TruncateString(templateStr, 100)),
		)

		// Fallback to old regex-based method
		result = templateStr
		stringData := contextToMap(ctx) // Use existing function for fallback

		result = renderEachBlocks(result, ctx)
		result = renderIfBlocks(result, stringData)
		result = renderGoStyleSubstitutions(result, stringData)
		result = renderHandlebarsSubstitutions(result, stringData)

		navItemsHTML := buildNavItemsHTML(ctx.NavItems)
		result = strings.ReplaceAll(result, "{{nav_items_html}}", navItemsHTML)
		result = strings.ReplaceAll(result, "{{.nav_items_html}}", navItemsHTML)

		// Quick links for footer — primary + utility items (set by render_site_components)
		if len(ctx.FooterNavItems) > 0 {
			quickLinksHTML := buildNavItemsHTML(ctx.FooterNavItems)
			result = strings.ReplaceAll(result, "{{quick_links_html}}", quickLinksHTML)
			result = strings.ReplaceAll(result, "{{.quick_links_html}}", quickLinksHTML)
		}
	}

	// =====================================================================
	// Report fields that rendered empty, then strip Go's "<no value>"
	// artefact. Detection is deterministic — missingBareFields scans the
	// source template and tests the data map, independent of the
	// "<no value>" string. A blanked href=/src= is a DEAD CONTROL shipped
	// to a live page: logged at Error with its field names so it is
	// greppable and alertable, not a merely-missing word (Warn). This
	// replaced the previous count-only Warn that named nothing, under which
	// 30 dead controls shipped silently on idea.uk (bugs_open/018).
	// =====================================================================
	missing, inURLAttr := missingBareFields(templateStr, data)
	if strings.Contains(result, "<no value>") {
		result = strings.ReplaceAll(result, "<no value>", "")
	}
	if len(inURLAttr) > 0 {
		logger.Error("RenderTemplate: URL attribute rendered empty — dead control",
			zap.Strings("fields", inURLAttr),
			zap.Strings("all_missing", missing),
			zap.String("template_preview", datahelpers.TruncateString(templateStr, 100)),
		)
	} else if len(missing) > 0 {
		logger.Warn("RenderTemplate: fields rendered empty",
			zap.Strings("fields", missing),
			zap.String("template_preview", datahelpers.TruncateString(templateStr, 100)),
		)
	}

	return result, missing, inURLAttr
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

	// Same post-merge sanitisation as the Go-template path. This is the regex
	// fallback RenderTemplate drops to when template execution fails; without
	// this it would render the dead form the other path now repairs.
	sanitiseFormActionStrings(result, ctx)

	return result
}

// contextToInterfaceMap converts RenderContext to map[string]interface{}
// This preserves nested structures (slices, maps) required for {{range}}
func contextToInterfaceMap(ctx *RenderContext) map[string]interface{} {
	if ctx.Year == "" {
		ctx.Year = fmt.Sprintf("%d", time.Now().Year())
	}

	logoText := ctx.LogoText
	if logoText == "" && ctx.CompanyName != "" {
		logoText = ctx.CompanyName
	}
	if logoText == "" && ctx.Domain != "" {
		parts := strings.Split(ctx.Domain, ".")
		if len(parts) > 0 && len(parts[0]) > 0 {
			logoText = strings.ToUpper(parts[0][:1]) + parts[0][1:]
		}
	}
	if logoText == "" {
		logoText = "Company"
	}

	result := map[string]interface{}{
		// Site info
		"domain":       ctx.Domain,
		"site_id":      ctx.SiteID,
		"logo_text":    logoText,
		"logo_url":     ctx.LogoURL,
		"company_name": defaultString(ctx.CompanyName, logoText),
		"tagline":      ctx.Tagline,

		// Navigation - keep as slice for {{range}}
		"nav_items":    ctx.NavItems,
		"current_page": ctx.CurrentPage,

		// Colors
		"primary_color":    defaultString(ctx.PrimaryColor, "#1a1a2e"),
		"secondary_color":  defaultString(ctx.SecondaryColor, "#2d2d44"),
		"accent_color":     defaultString(ctx.AccentColor, "#16a085"),
		"text_color":       defaultString(ctx.TextColor, "#333333"),
		"background_color": defaultString(ctx.BackgroundColor, "#ffffff"),
		"theme_css":        ctx.ThemeCSS,

		// Page
		"title":       ctx.Title,
		"description": ctx.Description,

		// Contact
		"email":         ctx.Email,
		"contact_email": ctx.Email,
		"phone":         ctx.Phone,

		// CTA - use defaults if not set
		"cta_text": defaultString(ctx.CTAText, "Get Started"),
		"cta_url":  defaultString(ctx.CTAUrl, "/contact.html"),

		// Metadata
		"year":            ctx.Year,
		"industry":        ctx.Industry,
		"tone":            ctx.Tone,
		"target_audience": ctx.TargetAudience,
		"services":        ctx.Services,
	}

	// Merge ContentData - this contains LLM-generated content
	// IMPORTANT: ContentData fields take priority and should NOT be aliased
	// because each section has its own specific content
	for key, value := range ctx.ContentData {
		result[key] = value
	}

	// REMOVED: applyBidirectionalAliases(result)
	// This was causing headlines to bleed across sections!

	// Only apply SAFE aliases that don't affect content fields
	applySafeAliases(result)

	// form_action MUST be sanitised after the ContentData merge, not defaulted
	// in the base map above. The broken values are ones the LLM actively wrote
	// ("#contact" on 8 live sites), so a base-map default would be overwritten
	// by the merge and would only ever fix the empty case. See bugs_open/006 §B.
	sanitiseFormAction(result, ctx)

	return result
}

// nonDeliveringFormActions are the values observed in live content_data that
// render a form that submits nowhere. "#contact" and friends POST to the
// current URL, which on a static host is a 405/404: the visitor gets no error
// they can act on and the message is silently lost.
var nonDeliveringFormActions = map[string]bool{
	"":              true,
	"#":             true,
	"#contact":      true,
	"#contact-form": true,
	"/contact":      true, // no such backend has ever existed in the chassis
}

// sanitiseFormAction replaces a non-delivering form_action with a mailto: built
// from the site's real contact address — the pattern the owner chose on
// 2026-07-17 (idea_uk_vm_site/RUNNING_NOTES §Q) and which idea.uk already uses.
//
// It deliberately does NOT fall back to a synthesised "info@<domain>" the way
// some render paths do for display-only contact blocks. An address nobody reads
// makes the form look repaired while still losing the message, which is worse
// than the visible breakage: the failure stops being detectable from outside.
// Where no real address is resolvable the action is left as-is for the
// contact_form_undeliverable discovery check to raise for a human.
func sanitiseFormAction(data map[string]interface{}, ctx *RenderContext) {
	raw, present := data["form_action"]
	if !present {
		// Absent is only a defect if this component actually has a form to
		// point somewhere; templates without one must not gain the field.
		return
	}
	if replacement, ok := deliverableFormAction(fmt.Sprintf("%v", raw), ctx); ok {
		data["form_action"] = replacement
	}
}

// sanitiseFormActionStrings is the same rule for the regex fallback path, whose
// context map is map[string]string. Both render paths in RenderTemplate merge
// ContentData, so both can carry a "#contact" the LLM wrote; fixing only the Go
// template path would leave the fallback rendering a dead form and still read
// as fixed. One rule, two callers — deliberately not two copies of the list.
func sanitiseFormActionStrings(data map[string]string, ctx *RenderContext) {
	current, present := data["form_action"]
	if !present {
		return
	}
	if replacement, ok := deliverableFormAction(current, ctx); ok {
		data["form_action"] = replacement
	}
}

// deliverableFormAction reports the mailto: that should replace a form_action
// which submits nowhere, or ok=false to leave the existing value alone.
func deliverableFormAction(current string, ctx *RenderContext) (string, bool) {
	if !nonDeliveringFormActions[strings.ToLower(strings.TrimSpace(current))] {
		return "", false // already points somewhere real (a mailto:, a live handler)
	}

	email := strings.TrimSpace(ctx.Email)
	if email == "" || !strings.Contains(email, "@") {
		return "", false // nothing honest to substitute — leave it for the check
	}

	// Two render paths synthesise ctx.Email = "info@" + Domain as a DISPLAY
	// fallback before rendering when the site has no configured address
	// (section_editor_actions.go:452, multipage_actions.go:333). That value is
	// not a real inbox — the 4 address-less sites (robot-hands, relojistas,
	// vetcomparison, vonc) all have an empty sites.email, so any
	// "info@<their own domain>" reaching here IS that fallback. Building a mailto
	// to it would fabricate exactly the address this function refuses to invent,
	// silently and only on those two paths. The struct field is the sole input
	// the sanitiser sees, so guarding it here closes the gap on every caller at
	// once. A site that has genuinely configured info@<its own domain> as a
	// monitored inbox is the one false refusal; the form is then left for the
	// contact_form_undeliverable check to raise for a human, which is safe.
	// A real info@ on a DIFFERENT domain (e.g. a shared CRM inbox) is honoured.
	if strings.EqualFold(email, "info@"+strings.TrimSpace(ctx.Domain)) {
		return "", false
	}

	subject := ctx.Domain
	if subject == "" {
		subject = "website"
	}
	return fmt.Sprintf("mailto:%s?subject=%s enquiry", email, subject), true
}

// applySafeAliases only adds aliases for non-content fields
// It does NOT touch headline, subheadline, section_title, etc.
func applySafeAliases(data map[string]interface{}) {
	// CTA aliases - these are safe because they're typically set once per page
	if v, ok := data["primary_cta"]; ok && v != nil && v != "" {
		if _, exists := data["cta_text"]; !exists {
			data["cta_text"] = v
		}
	}
	if v, ok := data["primary_cta_url"]; ok && v != nil && v != "" {
		if _, exists := data["cta_url"]; !exists {
			data["cta_url"] = v
		}
	}

	// Contact email alias
	if v, ok := data["email"]; ok && v != nil && v != "" {
		if _, exists := data["contact_email"]; !exists {
			data["contact_email"] = v
		}
	}

	// DO NOT alias:
	// - headline <-> heading <-> title <-> section_title
	// - subheadline <-> subtitle <-> description
	// - content <-> body <-> text
	// These should be set explicitly by each section's LLM content
}

// applyBidirectionalAliases ensures templates work regardless of field naming
func applyBidirectionalAliases(data map[string]interface{}) {
	// Map of field -> aliases (bidirectional)
	aliasGroups := [][]string{
		{"headline", "heading", "title", "section_title"},
		{"subheadline", "subtitle", "sub_headline", "lead", "description"},
		{"cta_text", "primary_cta", "button_text"},
		{"cta_url", "primary_cta_url", "button_url"},
		{"content", "body", "text", "paragraph"},
	}

	for _, group := range aliasGroups {
		// Find first non-empty value in group
		var foundValue interface{}
		for _, field := range group {
			if val, exists := data[field]; exists && val != nil && val != "" {
				foundValue = val
				break
			}
		}

		// If found, ensure all aliases have this value
		if foundValue != nil {
			for _, field := range group {
				if _, exists := data[field]; !exists {
					data[field] = foundValue
				}
			}
		}
	}
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
	result = renderEachBlocks(result, ctx)
	result = renderIfBlocks(result, data)
	result = renderGoStyleSubstitutions(result, data)
	result = renderHandlebarsSubstitutions(result, data)

	// Build nav items HTML
	navItemsHTML := buildNavItemsHTML(ctx.NavItems)
	result = strings.ReplaceAll(result, "{{nav_items_html}}", navItemsHTML)
	result = strings.ReplaceAll(result, "{{.nav_items_html}}", navItemsHTML)

	// Quick links for footer — primary + utility items
	if len(ctx.FooterNavItems) > 0 {
		quickLinksHTML := buildNavItemsHTML(ctx.FooterNavItems)
		result = strings.ReplaceAll(result, "{{quick_links_html}}", quickLinksHTML)
		result = strings.ReplaceAll(result, "{{.quick_links_html}}", quickLinksHTML)
	}

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

// renderEachBlocks handles {{#each <collection>}}...{{/each}} for known collections
// Supported collections: nav_items, services, quick_links
func renderEachBlocks(template string, ctx *RenderContext) string {
	// Pattern to match any {{#each <name>}}...{{/each}}
	eachRe := regexp.MustCompile(`(?s)\{\{#each\s+(\w+)\}\}(.*?)\{\{/each\}\}`)

	return eachRe.ReplaceAllStringFunc(template, func(match string) string {
		matches := eachRe.FindStringSubmatch(match)
		if len(matches) < 3 {
			return match
		}

		collectionName := matches[1]
		itemTemplate := matches[2]

		var result strings.Builder

		switch collectionName {
		case "nav_items":
			for _, item := range ctx.NavItems {
				itemStr := itemTemplate
				itemStr = strings.ReplaceAll(itemStr, "{{this.url}}", item.URL)
				itemStr = strings.ReplaceAll(itemStr, "{{this.label}}", item.Label)

				// Handle {{#if this.is_active}}...{{/if}}
				activeRe := regexp.MustCompile(`(?s)\{\{#if\s+this\.is_active\}\}(.*?)\{\{/if\}\}`)
				itemStr = activeRe.ReplaceAllStringFunc(itemStr, func(m string) string {
					innerMatches := activeRe.FindStringSubmatch(m)
					if len(innerMatches) < 2 {
						return m
					}
					if item.IsActive {
						return innerMatches[1]
					}
					return ""
				})

				result.WriteString(itemStr)
			}

		case "services":
			// Get services from ContentData
			services := extractServicesFromContext(ctx)
			for _, svc := range services {
				itemStr := itemTemplate
				itemStr = strings.ReplaceAll(itemStr, "{{this.name}}", svc.Name)
				itemStr = strings.ReplaceAll(itemStr, "{{this.slug}}", svc.Slug)
				itemStr = strings.ReplaceAll(itemStr, "{{this.description}}", svc.Description)
				result.WriteString(itemStr)
			}

		case "quick_links":
			// Quick links for footer - prefer FooterNavItems if available
			navItems := ctx.NavItems
			if len(ctx.FooterNavItems) > 0 {
				navItems = ctx.FooterNavItems
			}
			for _, item := range navItems {
				itemStr := itemTemplate
				itemStr = strings.ReplaceAll(itemStr, "{{this.url}}", item.URL)
				itemStr = strings.ReplaceAll(itemStr, "{{this.label}}", item.Label)
				result.WriteString(itemStr)
			}

		default:
			// Unknown collection - try to get from ContentData as generic array
			if items := extractGenericArray(ctx.ContentData, collectionName); len(items) > 0 {
				for _, item := range items {
					itemStr := renderGenericItem(itemTemplate, item)
					result.WriteString(itemStr)
				}
			} else {
				// Return original if we can't process it
				return match
			}
		}

		return result.String()
	})
}

// ServiceItem represents a service for template rendering
type ServiceItem struct {
	Name        string
	Slug        string
	Description string
}

// extractServicesFromContext gets services from RenderContext.ContentData
func extractServicesFromContext(ctx *RenderContext) []ServiceItem {
	var services []ServiceItem

	// Try direct services array
	if svcData, ok := ctx.ContentData["services"]; ok {
		services = parseServicesArray(svcData)
		if len(services) > 0 {
			return services
		}
	}

	// Try brief.services
	if brief, ok := ctx.ContentData["brief"].(map[string]interface{}); ok {
		if svcData, ok := brief["services"]; ok {
			services = parseServicesArray(svcData)
			if len(services) > 0 {
				return services
			}
		}
	}

	// Try reviewed_brief.services
	if brief, ok := ctx.ContentData["reviewed_brief"].(map[string]interface{}); ok {
		if svcData, ok := brief["services"]; ok {
			services = parseServicesArray(svcData)
		}
	}

	return services
}

// parseServicesArray converts various service formats to ServiceItem slice
func parseServicesArray(data interface{}) []ServiceItem {
	var services []ServiceItem

	switch v := data.(type) {
	case []interface{}:
		for _, item := range v {
			if svcMap, ok := item.(map[string]interface{}); ok {
				name, _ := svcMap["name"].(string)
				if name == "" {
					name, _ = svcMap["title"].(string)
				}

				slug, _ := svcMap["slug"].(string)
				if slug == "" {
					// Generate slug from name
					slug = strings.ToLower(strings.ReplaceAll(name, " ", "-"))
					slug = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(slug, "")
				}

				desc, _ := svcMap["description"].(string)

				if name != "" {
					services = append(services, ServiceItem{
						Name:        name,
						Slug:        slug,
						Description: desc,
					})
				}
			}
		}

	case string:
		// Might be a JSON string - try to parse it
		var items []map[string]interface{}
		if err := json.Unmarshal([]byte(v), &items); err == nil {
			for _, item := range items {
				name, _ := item["name"].(string)
				slug, _ := item["slug"].(string)
				if slug == "" {
					slug = strings.ToLower(strings.ReplaceAll(name, " ", "-"))
				}
				desc, _ := item["description"].(string)

				if name != "" {
					services = append(services, ServiceItem{
						Name:        name,
						Slug:        slug,
						Description: desc,
					})
				}
			}
		}
	}

	return services
}

// extractGenericArray gets a generic array from ContentData
func extractGenericArray(data map[string]interface{}, key string) []map[string]interface{} {
	if arr, ok := data[key].([]interface{}); ok {
		var result []map[string]interface{}
		for _, item := range arr {
			if m, ok := item.(map[string]interface{}); ok {
				result = append(result, m)
			}
		}
		return result
	}
	return nil
}

// renderGenericItem renders a template with a generic map item
func renderGenericItem(template string, item map[string]interface{}) string {
	result := template
	for key, value := range item {
		placeholder := "{{this." + key + "}}"
		if strVal, ok := value.(string); ok {
			result = strings.ReplaceAll(result, placeholder, strVal)
		} else {
			result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
		}
	}
	return result
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

// simplifyNavLabelForRender creates a clean nav label at render time.
// Most simplification happens upstream in populate_nav_tables (navSimplifyLabel)
// which has access to the page's nav_label. This function only handles cleanup
// of labels that still contain brand suffixes like "About Us | Brand Name".
//
// Trust the stored label when possible — aggressive render-time truncation
// produced labels like "Ai For" from "AI For Your Type Of Business".
func simplifyNavLabelForRender(label, url string) string {
	// Strip "|" and everything after (e.g. "About | Company Name" -> "About")
	if idx := strings.Index(label, "|"); idx > 0 {
		label = strings.TrimSpace(label[:idx])
	}

	// Strip " - " and everything after (e.g. "Services - Company" -> "Services")
	if idx := strings.Index(label, " - "); idx > 0 {
		label = strings.TrimSpace(label[:idx])
	}

	// For known short page names, use canonical labels regardless of input.
	// This handles e.g. index.html with title "Homepage | Company" -> "Home".
	pageName := strings.TrimSuffix(strings.TrimPrefix(url, "/"), ".html")
	pageNameLower := strings.ToLower(pageName)

	simpleLabels := map[string]string{
		"index": "Home",
		"home":  "Home",
	}
	if simple, ok := simpleLabels[pageNameLower]; ok {
		return simple
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
.site-header{background:var(--color-header-bg, var(--color-surface));padding:1rem 0;position:sticky;top:0;z-index:1000;box-shadow:0 2px 10px rgba(0,0,0,.1)}
.header-container{max-width:1200px;margin:0 auto;padding:0 2rem;display:flex;align-items:center;justify-content:space-between}
.logo{text-decoration:none;font-size:1.5rem;font-weight:700;color:var(--color-header-text, var(--color-text))}
.main-nav ul{display:flex;list-style:none;margin:0;padding:0;gap:2rem}
.main-nav a{color:color-mix(in srgb, var(--color-header-text, var(--color-text)) 90%%, transparent);text-decoration:none;font-weight:500;transition:color .2s}
.main-nav a:hover,.main-nav a.active{color:var(--color-accent)}
.mobile-menu-toggle{display:none;background:none;border:none;cursor:pointer;padding:.5rem}
.mobile-menu-toggle span{display:block;width:24px;height:2px;background:var(--color-header-text, var(--color-text));margin:5px 0}
@media(max-width:768px){.mobile-menu-toggle{display:block}.main-nav{position:absolute;top:100%%;left:0;right:0;background:var(--color-header-bg, var(--color-surface));padding:1rem;display:none}.main-nav.active{display:block}.main-nav ul{flex-direction:column;gap:0}.main-nav a{display:block;padding:.75rem 0;border-bottom:1px solid color-mix(in srgb, var(--color-header-text, var(--color-text)) 10%%, transparent)}}
</style>
<script>document.addEventListener("DOMContentLoaded",function(){var t=document.querySelector(".mobile-menu-toggle"),n=document.querySelector(".main-nav");t&&n&&t.addEventListener("click",function(){n.classList.toggle("active")})});</script>`,
		ctx.LogoText, navHTML)
}

// RenderFallbackFooter creates a basic footer when no component is available
func RenderFallbackFooter(ctx *RenderContext) string {
	navHTML := buildNavItemsHTML(ctx.NavItems)
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
.site-footer{background:var(--color-footer-bg, var(--color-surface));color:color-mix(in srgb, var(--color-footer-text, var(--color-text)) 90%%, transparent);padding:3rem 0 0;margin-top:auto}
.footer-container{max-width:1200px;margin:0 auto;padding:0 2rem;display:grid;grid-template-columns:2fr 1fr 1fr;gap:2rem}
.footer-brand h3,.footer-links h4,.footer-contact h4{color:var(--color-footer-text, var(--color-text));margin:0 0 1rem}
.footer-links ul{list-style:none;padding:0;margin:0}
.footer-links li{margin-bottom:.5rem}
.footer-links a{color:color-mix(in srgb, var(--color-footer-text, var(--color-text)) 70%%, transparent);text-decoration:none}
.footer-links a:hover{color:var(--color-footer-text, var(--color-text))}
.footer-bottom{margin-top:2rem;padding:1.5rem 0;border-top:1px solid color-mix(in srgb, var(--color-footer-text, var(--color-text)) 10%%, transparent);text-align:center}
.footer-bottom p{margin:0;color:color-mix(in srgb, var(--color-footer-text, var(--color-text)) 60%%, transparent);font-size:.9rem}
@media(max-width:768px){.footer-container{grid-template-columns:1fr}}
</style>`, ctx.LogoText, ctx.Tagline, navHTML, ctx.Email, year, ctx.CompanyName)
}

// ===========================================================================
// HTML INJECTION FUNCTIONS
// ===========================================================================

// InjectHeader replaces an existing header in HTML with a rendered component
func InjectHeader(ctx context.Context, db interface{}, html string, siteID uuid.UUID, renderCtx *RenderContext, logger *zap.Logger) string {
	// Skip injection if page content already contains a site-header.
	// This handles custom layouts where nav is part of a page section
	// (e.g. hero-with-integrated-nav, sidebar-nav adopted from crawl).
	if strings.Contains(html, `class="site-header`) ||
		strings.Contains(html, `class="site-header--`) {
		logger.Info("InjectHeader: page already contains site-header, skipping injection",
			zap.String("site_id", siteID.String()))
		return html
	}

	// Update nav items from deployed pages (not cached db_sync)
	if sqlDB, ok := db.(*sql.DB); ok && siteID != uuid.Nil {
		/*headerNav := GetHeaderNavFromPages(ctx, sqlDB, siteID, 6, logger)*/
		// NavFetchableOnly: this nav is injected into shipped page HTML. It was
		// `true` (deployedOnly), which meant the same intent but filtered on
		// build_status — dropping pages that are needs_rebuild yet still serving.
		// maxItems=0: no limit here â€” PopulateNavTablesAction already controls
		// which pages go into the primary group vs utility.
		headerNav := GetNavItems(ctx, sqlDB, siteID, []string{NavGroupPrimary}, NavFetchableOnly, 0, logger)
		if len(headerNav) > 0 {
			renderCtx.NavItems = headerNav
			logger.Debug("InjectHeader: Updated nav from deployed pages",
				zap.Int("items", len(headerNav)),
			)
		}
	}

	headerHTML, err := RenderHeader(ctx, db, siteID, renderCtx, logger)
	if err != nil {
		logger.Warn("Failed to render header", zap.Error(err))
		headerHTML = RenderFallbackHeader(renderCtx)
	}

	// Remove existing header AND its trailing <style> and <script> blocks
	// Pattern: <header...>...</header> optionally followed by <style>...</style> and/or <script>...</script>
	// Also capture any SOURCE comments that precede the header
	headerRe := regexp.MustCompile(`(?is)(?:<!--\s*HEADER\s+SOURCE:[^>]*-->\s*)*<header[^>]*>.*?</header>\s*(?:<style>.*?</style>\s*)?(?:<script>.*?</script>\s*)?`)
	html = headerRe.ReplaceAllString(html, "<!-- HEADER_REPLACED -->")

	// Insert after <body>
	bodyRe := regexp.MustCompile(`(?i)(<body[^>]*>)`)
	if bodyRe.MatchString(html) {
		html = bodyRe.ReplaceAllString(html, "$1\n"+headerHTML)
		html = strings.Replace(html, "<!-- HEADER_REPLACED -->", "", -1) // Remove ALL placeholders
	} else {
		html = strings.Replace(html, "<!-- HEADER_REPLACED -->", headerHTML, 1)
	}

	return html
}

// InjectFooter replaces an existing footer in HTML with a rendered component
func InjectFooter(ctx context.Context, db interface{}, html string, siteID uuid.UUID, renderCtx *RenderContext, logger *zap.Logger) string {
	// Skip injection if page content already contains a site-footer.
	if strings.Contains(html, `class="site-footer`) {
		logger.Info("InjectFooter: page already contains site-footer, skipping injection",
			zap.String("site_id", siteID.String()))
		return html
	}
	// Update nav items from deployed pages for footer
	// Footer includes legal pages (privacy, terms) that may not be in header
	if sqlDB, ok := db.(*sql.DB); ok && siteID != uuid.Nil {
		/*footerNav := GetFooterNavFromPages(ctx, sqlDB, siteID, logger)*/
		footerNav := GetNavItems(ctx, sqlDB, siteID, []string{NavGroupPrimary, NavGroupUtility, NavGroupLegal}, NavFetchableOnly, 0, logger)
		if len(footerNav) > 0 {
			renderCtx.FooterNavItems = footerNav
			logger.Debug("InjectFooter: Updated nav from deployed pages",
				zap.Int("items", len(footerNav)),
			)
		}
	}

	footerHTML, err := RenderFooter(ctx, db, siteID, renderCtx, logger)
	if err != nil {
		logger.Warn("Failed to render footer", zap.Error(err))
		footerHTML = RenderFallbackFooter(renderCtx)
	}

	// Remove existing footer AND its trailing <style> block
	// Pattern: <footer...>...</footer> optionally followed by <style>...</style>
	// Also capture any SOURCE comments that precede the footer
	footerRe := regexp.MustCompile(`(?is)(?:<!--\s*FOOTER\s+SOURCE:[^>]*-->\s*)*<footer[^>]*>.*?</footer>\s*(?:<style>.*?</style>\s*)?`)
	html = footerRe.ReplaceAllString(html, "<!-- FOOTER_REPLACED -->")

	// Also remove any orphaned footer styles that appear BEFORE a footer (from partial replacements)
	// This handles the case where styles were separated from their footer tag
	orphanedFooterStyleRe := regexp.MustCompile(`(?is)<style>\s*\.site-footer\s*\{.*?</style>\s*(<!--\s*FOOTER)`)
	html = orphanedFooterStyleRe.ReplaceAllString(html, "$1")

	// Insert before </body>
	bodyCloseRe := regexp.MustCompile(`(?i)(</body>)`)
	if bodyCloseRe.MatchString(html) {
		html = bodyCloseRe.ReplaceAllString(html, footerHTML+"\n$1")
		html = strings.Replace(html, "<!-- FOOTER_REPLACED -->", "", -1) // Remove ALL placeholders
	} else {
		html = strings.Replace(html, "<!-- FOOTER_REPLACED -->", footerHTML, 1)
	}

	// Clean up any content after </html> (malformed pages)
	html = cleanContentAfterHTML(html)

	return html
}

// cleanContentAfterHTML removes any content that appears after the closing </html> tag
// This handles malformed pages where footer or other content was appended incorrectly
func cleanContentAfterHTML(html string) string {
	// Find the LAST </html> tag (in case there are duplicates)
	htmlCloseRe := regexp.MustCompile(`(?i)</html>`)
	matches := htmlCloseRe.FindAllStringIndex(html, -1)

	if len(matches) > 0 {
		// Keep only up to and including the FIRST </html>
		firstClose := matches[0][1] // End of first </html>
		if firstClose < len(html) {
			html = html[:firstClose]
		}
	}

	return html
}

// RenderHead renders the head component for a site
// Looks up component by function "head", applies style collection colors,
// and renders with page-specific title/description from RenderContext
func RenderHead(ctx context.Context, db interface{}, siteID uuid.UUID, renderCtx *RenderContext, logger *zap.Logger) (string, error) {
	var coll *StyleCollection
	var err error
	var source string = "fallback"

	// Load style collection for colors
	if siteID != uuid.Nil {
		coll, err = GetStyleCollectionForSite(ctx, db, siteID, logger)
		if err != nil {
			logger.Warn("RenderHead: Failed to get style collection", zap.Error(err))
		}
	}

	if coll == nil && renderCtx.Domain != "" {
		coll, err = SelectStyleCollectionByDomain(ctx, db, renderCtx.Domain, logger)
		if err != nil {
			logger.Warn("RenderHead: Failed to select style collection by domain", zap.Error(err))
		}
	}

	// Apply colors from collection if not already set
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

	// Apply color defaults if still empty (head template uses these in CSS variables)
	if renderCtx.PrimaryColor == "" {
		renderCtx.PrimaryColor = "#1a1a2e"
	}
	if renderCtx.SecondaryColor == "" {
		renderCtx.SecondaryColor = "#2d2d44"
	}
	if renderCtx.AccentColor == "" {
		renderCtx.AccentColor = "#16a085"
	}
	if renderCtx.TextColor == "" {
		renderCtx.TextColor = "#333333"
	}
	if renderCtx.BackgroundColor == "" {
		renderCtx.BackgroundColor = "#ffffff"
	}

	// Get head component â€” lookup by function name "head"
	comp, err := GetComponentByFunction(ctx, db, "head", logger)
	if err != nil {
		logger.Warn("RenderHead: No head component found, using fallback")
		head := RenderFallbackHead(renderCtx)
		return fmt.Sprintf("<!-- HEAD SOURCE: fallback -->\n%s", head), nil
	}
	source = "component-db:" + comp.Name

	// Render template with context (title, description, colors etc.)
	rendered := RenderTemplate(comp.HTMLTemplate, renderCtx, logger)

	logger.Info("RenderHead: Rendered head component",
		zap.String("source", source),
		zap.String("title", renderCtx.Title),
		zap.Int("html_length", len(rendered)),
	)

	return fmt.Sprintf("<!-- HEAD SOURCE: %s -->\n%s", source, rendered), nil
}

// RenderFallbackHead creates a basic head section when no component is available
func RenderFallbackHead(ctx *RenderContext) string {
	title := ctx.Title
	if title == "" {
		title = ctx.CompanyName
	}
	description := ctx.Description
	if description == "" {
		description = ctx.Tagline
	}
	primary := defaultString(ctx.PrimaryColor, "#1a1a2e")

	return fmt.Sprintf(`<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <meta name="description" content="%s">
    <meta name="theme-color" content="%s">
    <link rel="stylesheet" href="/assets/css/styles.css">
</head>`, title, description, primary)
}

// InjectHead removes any existing <head> blocks and inserts the rendered head
// component in the correct position â€” always before <body>, never in-place.
// Previous version did in-place replacement which preserved wrong positioning
// when <head> had migrated inside <body> (e.g. via cleanHTMLStructure dedup).
func InjectHead(ctx context.Context, db interface{}, html string, siteID uuid.UUID, renderCtx *RenderContext, logger *zap.Logger) string {
	headHTML, err := RenderHead(ctx, db, siteID, renderCtx, logger)
	if err != nil {
		logger.Warn("InjectHead: Failed to render head, using fallback", zap.Error(err))
		headHTML = RenderFallbackHead(renderCtx)
	}

	// Step 1: Remove ALL existing <head>...</head> blocks regardless of position.
	// Use [\s>] after <head to avoid matching <header> tags.
	headRe := regexp.MustCompile(`(?is)<head[\s>].*?</head>`)
	existedBefore := headRe.MatchString(html)
	html = headRe.ReplaceAllString(html, "")

	if existedBefore {
		logger.Debug("InjectHead: Removed existing <head> block(s)")
	}

	// Step 2: Insert new head in correct position â€” always before <body>
	bodyRe := regexp.MustCompile(`(?i)(<body[^>]*>)`)
	if bodyRe.MatchString(html) {
		html = bodyRe.ReplaceAllString(html, headHTML+"\n$1")
		logger.Debug("InjectHead: Inserted head before <body>")
	} else {
		// No <body> â€” try after <html>
		htmlTagRe := regexp.MustCompile(`(?i)(<html[^>]*>)`)
		if htmlTagRe.MatchString(html) {
			html = htmlTagRe.ReplaceAllString(html, "$1\n"+headHTML)
			logger.Debug("InjectHead: Inserted head after <html>")
		} else {
			// No structure at all â€” prepend
			html = headHTML + "\n" + html
			logger.Warn("InjectHead: No <html> or <body> found, prepended head")
		}
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

// NavItemFromDB represents a navigation item from pages table
type NavItemFromDB struct {
	Label    string
	URL      string
	NavOrder int
	InHeader bool
	InFooter bool
}

// GetHeaderNavFromPages queries pages table for header navigation
// Only includes pages with in_header=true AND status in deployed/active
// DEPRECATED
func GetHeaderNavFromPages(ctx context.Context, db *sql.DB, siteID uuid.UUID, maxItems int, logger *zap.Logger) []NavItem {
	if db == nil || siteID == uuid.Nil {
		logger.Debug("GetHeaderNavFromPages: No DB or site_id, returning empty nav")
		return []NavItem{}
	}

	if maxItems <= 0 {
		maxItems = 6 // Default max header items
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
		  AND build_status = 'deployed'
		  AND deleted_at IS NULL
		ORDER BY nav_order ASC, created_at ASC
		LIMIT $2
	`

	rows, err := db.QueryContext(ctx, query, siteID, maxItems)
	if err != nil {
		logger.Warn("GetHeaderNavFromPages: Query failed", zap.Error(err))
		return []NavItem{}
	}
	defer rows.Close()

	var items []NavItem
	for rows.Next() {
		var label, url string
		var navOrder int
		if err := rows.Scan(&label, &url, &navOrder); err != nil {
			logger.Warn("GetHeaderNavFromPages: Scan failed", zap.Error(err))
			continue
		}

		// Clean up verbose labels
		label = datahelpers.SimplifyNavLabel(label, datahelpers.ExtractNameFromURL(url))

		items = append(items, NavItem{
			Label: label,
			URL:   url,
		})
	}

	logger.Debug("GetHeaderNavFromPages: Built nav",
		zap.Int("items", len(items)),
		zap.String("site_id", siteID.String()),
	)

	return items
}

// GetFooterNavFromPages queries pages table for footer navigation
// Includes pages with in_footer=true OR legal pages (privacy, terms)
// DEPRECATED
func GetFooterNavFromPages(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) []NavItem {
	if db == nil || siteID == uuid.Nil {
		return []NavItem{}
	}

	query := `
		SELECT 
			COALESCE(nav_label, title, name) as label,
			COALESCE(url, '/' || name || '.html') as url,
			COALESCE(nav_order, 0) as nav_order
		FROM pages 
		WHERE site_id = $1 
		  AND (in_footer = true OR LOWER(name) LIKE '%privacy%' OR LOWER(name) LIKE '%terms%')
		  AND status IN ('deployed', 'active')
		  AND build_status = 'deployed'
		  AND deleted_at IS NULL
		ORDER BY 
			CASE WHEN LOWER(name) LIKE '%privacy%' OR LOWER(name) LIKE '%terms%' THEN 1 ELSE 0 END,
			nav_order ASC
	`

	rows, err := db.QueryContext(ctx, query, siteID)
	if err != nil {
		logger.Warn("GetFooterNavFromPages: Query failed", zap.Error(err))
		return []NavItem{}
	}
	defer rows.Close()

	var items []NavItem
	for rows.Next() {
		var label, url string
		var navOrder int
		if err := rows.Scan(&label, &url, &navOrder); err != nil {
			continue
		}

		label = datahelpers.SimplifyNavLabel(label, datahelpers.ExtractNameFromURL(url))
		items = append(items, NavItem{Label: label, URL: url})
	}

	return items
}
