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
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
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

	// InputSchema is the component's declared field contract, set by any caller
	// that has the component in hand. NIL MEANS UNKNOWN, NOT VALID: the seam
	// then reports nothing and behaves exactly as before, which is fail-open and
	// is stated rather than hidden — a caller that does not set it gets no
	// absent-required-field report.
	//
	// It exists for bugs_open/342. Go's missingkey=zero renders a field the
	// content never supplied as EMPTY WITH NO ERROR, page assembly then drops a
	// visually-empty section, and the content vanishes — the mechanism behind
	// the fleet-wide blanking of article bodies (bugs_closed/004/005). The gate
	// that catches it, missingRequiredLLMFields, was called at 2 of the 15
	// render call sites. Carrying the schema on the context is what lets the
	// SEAM apply that same rule for every caller instead of each caller
	// remembering to — the same "make the guarantee mechanical here, where they
	// all arrive" move as the form_action seeding and the InstanceID report.
	//
	// It reuses the slot of the dead `SchemaSnapshot` field (with `SchemaMode`
	// and `RenderOptions`, deleted 2026-08-21): their only reader was
	// RenderTemplateWithValidation, which went with the regex fallback in
	// bugs_closed/260, leaving three declarations that described a strict-mode
	// validation this binary had stopped doing.
	//
	// ⚠ CONTROL FIELD — it must never be settable from content. A field the
	// content can supply would let content hand the renderer its own contract
	// and switch off its own check. It is a map, and the step contract is
	// derived by reflection over STRING fields only (renderContextScalarFields),
	// so it is excluded structurally rather than by an exclusion list — and
	// render_context_derivation_test.go asserts exactly that.
	InputSchema map[string]interface{} `json:"-"`

	// AbsentRequiredFields is an OUTPUT: RenderTemplate writes it, callers read
	// it. It names the schema-required source:"llm" fields that rendered EMPTY,
	// so a caller with a database handle and a site identity can escalate —
	// which the seam cannot, having neither.
	//
	// It is an out-field rather than a fifth return value because the seam's
	// signature was just unified to ONE spelling across sixteen call sites
	// (2026-08-21) and churning it again the same day would undo that; this
	// function already mutates ctx for the form_action seeding, so the shape has
	// precedent here. `json:"-"` for the same reason as InputSchema above.
	AbsentRequiredFields []string `json:"-"`

	// RenderedTemplateSHA is an OUTPUT: RenderTemplate writes it, callers read it.
	// It is the SHA-256 of the template text this call actually executed, and it
	// exists so that the bytes a render produces can be tied back to the exact
	// template that produced them (RFC_046, ruled 2026-08-22).
	//
	// The estate infers a component row's identity five different ways — from an
	// HTML attribute, from a sentinel meaning unknown, from POSITION in a plan,
	// from fuzzy name matching, and from slot-name equality during a rebuild — and
	// records it nowhere. bugs_open/357 is what that costs: a whole interactive
	// tool stored in a row declaring itself the shared `hero`, because hero was
	// first in the page's plan. A digest computed HERE is the one fact only the
	// seam knows, and it is knowable at no cost: this function has the template
	// text in hand and executes it.
	//
	// It is an out-field rather than a return value for exactly the reason
	// AbsentRequiredFields above is: the seam's signature was unified to ONE
	// spelling across sixteen call sites (owner ruling 2026-08-21) and churning it
	// again would undo that. Like that field, the seam only REPORTS — resolving the
	// digest to a component_versions row is the caller's job, because that needs a
	// database handle the seam does not have and must not acquire.
	//
	// EMPTY MEANS UNKNOWN, never "no template": an empty template returns early
	// above without setting it, and a caller that does not read it is unaffected.
	// `json:"-"` for the same reason as the two fields above — it must never be
	// settable from content, and the step contract is derived by reflection over
	// STRING fields, so this one IS reachable by that derivation and is excluded
	// by the tag deliberately. render_context_derivation_test.go is what tells you
	// if that stops being true.
	RenderedTemplateSHA string `json:"-"`
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
//
// ORDER BY name is load-bearing, not tidiness: without it this is `LIMIT 1` over a
// set that has TWO members for both `site-header` and `site-footer` (measured
// 2026-07-31), so the answer was whatever the plan happened to return. Measured
// before adding it — it returns the same row it returned unordered for both, so the
// fleet's answer is unchanged and is now guaranteed rather than incidental.
//
// This is the SECTION-shaped lookup and it stays that way: it has no
// component_level filter because its callers ask it for section functions
// (`generic-text-block`, via GetComponentWithFallback). For SITE CHROME use
// ResolveChromeComponent below — see the header there for why the two differ.
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
		ORDER BY name
		LIMIT 1
	`
	return queryComponent(ctx, db, query, function, logger)
}

// ===========================================================================
// SITE CHROME SELECTION — bugs_open/118
// ===========================================================================
//
// "Which library component serves chrome function F?" was asked in three places
// with three different predicates, and all three answers were wrong in a
// different way (measured live 2026-07-31):
//
//   render_site_components  no predicate at all   -> footer-4-column      (is_active=false)
//   link_site_components    is_active only        -> header-leopardess    (a client's FORK)
//   GetComponentByFunction  is_active + no fork   -> site-header          (component_level='section')
//
// So the predicate lives here once and the callers use it. The section selector
// (component_selector.go queryCandidates) has had the right shape all along —
// active, unforked, level-filtered — and is the reason this one is written the
// same way rather than invented.
//
// WHY NOT PARAMETERISE queryCandidates INSTEAD (the council's prior_art
// objection, corr 5bc232d6 — a fair one, because the plan cited it as correct
// and then did not reuse it). Four things would have to become conditional to
// share one WHERE clause: it keys on `section_type`, not `function`; it is
// hard-filtered to `component_level='section'`, which is the exact complement of
// the chrome set; it returns []ComponentCandidate with no html_template or
// input_schema, which is what chrome rendering actually needs; and it exists to
// SCORE a field of candidates, where chrome selection has one answer and the
// scoring is meaningless over a pool whose size is one. What is genuinely shared
// is the SHAPE of the eligibility rule, and that is what this copies —
// deliberately, and with the reasoning here rather than left to be rediscovered.

// chromeComponentLevels is the set of content_components.component_level values
// that may serve as SITE CHROME. It is a whitelist and not `<> 'section'` on
// purpose: 'section', 'tool' and 'element' are all page-body levels, and a new
// body level added later must not silently become eligible chrome.
//
// The estate uses four levels for chrome because the vocabulary grew twice:
// 'site' (the site-header/site-footer pool, 12 rows), and the singular 'header',
// 'footer' and 'head' levels that predate it.
//
// MEASURED 2026-07-31, answering the council's editquality objection (corr
// 5bc232d6) that 'header'/'footer' were in the list unverified: those two levels
// currently admit ZERO rows to any chrome pool. All six rows at them are legacy
// `*_pre_037` components carrying their OWN functions (`header-docs`,
// `header-minimal-tool`, `footer-with-disclaimer`, …), and `function = $1` is
// applied before the level filter is ever consulted. Exactly ONE row is both
// chrome-level and chrome-function — `head-seo-standard` (level 'head',
// function 'head', is_active=false) — which is what the 'head' entry is for.
// So 'header' and 'footer' are FORWARD-LOOKING, not observed. They stay because
// narrowing the list to ('site','head') would make a chrome component correctly
// filed at level 'header' silently ineligible, which is the worse failure: a
// selection that finds nothing looks exactly like a library with nothing in it.
const chromeComponentLevels = `'site', 'header', 'footer', 'head'`

// chromeEligibleSQL is THE chrome-eligibility predicate. Every chrome selection
// must be built from this string so a fourth hand-typed copy cannot drift back in
// — chrome_selection_test.go scans the package and fails if one does.
//
// forked_from IS NULL matters as much as is_active here and is the half that was
// missing everywhere: a fork carries its parent's `function`, so an ACTIVE fork of
// one site's header is a candidate to become every other site's header. That is
// not hypothetical — `header-leopardess` sorts first among active `site-header`
// rows and is what link_site_components would have assigned.
func chromeEligibleSQL(alias string) string {
	return alias + "is_active AND " +
		alias + "forked_from IS NULL AND " +
		alias + "component_level IN (" + chromeComponentLevels + ")"
}

// chromePinEligibleSQL is THE predicate for a style-collection chrome PIN
// (`style_collections.header_component_id` / `footer_component_id`). It is
// deliberately NOT chromeEligibleSQL, and the difference is one clause.
//
// A pin OMITS `forked_from IS NULL`. That clause is right for pool SELECTION — a
// fork carries its parent's `function`, so an active fork of one client's header
// is otherwise a candidate to become every other site's (bugs_open/118's
// `header-leopardess` finding) — and it is WRONG for a pin, because naming a
// site's own fork is exactly what a pin is for. Measured over all four live pins
// on 2026-07-31, that is the only row where the two predicates disagree, and the
// pin predicate is the one that gets it right:
//
//	ai-agent-orchestration.com  header-professional-dark  inactive        both false
//	finetuning.uk               header-professional-dark  inactive        both false
//	gaswholesalers.com          header-professional-dark  inactive        both false
//	leopardessconsulting.co.uk  header-leopardess         active FORK     pin TRUE, pool false
//
// Copying chromeEligibleSQL here would reject the single legitimate pin in the
// fleet while still catching the three real ones. `component_level` stays,
// because a pin decides what is SERVED as chrome on every page of every site on
// the collection, and a page-section component is not that (bugs_open/167).
//
// chrome_pin_test.go pins the asymmetry: make the two predicates equal and it
// goes red with the reason.
func chromePinEligibleSQL(alias string) string {
	return alias + "is_active AND " +
		alias + "component_level IN (" + chromeComponentLevels + ")"
}

// ChromeSlotFunction maps a site_components.slot_name to the
// content_components.function that serves it. One definition: the map used to be
// inline in render_site_components while link_site_components hard-coded the two
// function strings, which is how a slot can mean different things to two writers.
// An unknown slot maps to itself, preserving the existing behaviour for callers
// that pass a function name straight through.
func ChromeSlotFunction(slot string) string {
	switch slot {
	case "header":
		return "site-header"
	case "footer":
		return "site-footer"
	case "head":
		return "head"
	default:
		return slot
	}
}

// ResolveChromeComponent returns the library component that should serve a chrome
// function, and whether that component is actually ELIGIBLE chrome.
//
// eligible=false means the library holds no active, unforked, chrome-level row for
// this function at all, and the returned component is a last-resort pick. Callers
// must treat that as a defect to report, never as a default — it is how every site
// in the fleet ended up assigned to a deactivated footer.
//
// [UNCONSUMED] The report it produces — an ERROR log and `ineligible_chrome` in
// render_site_components' result — has NO automated reader today, and the council's
// bug_historian seat is right that this is the bugs_open/071 / bugs_open/083 shape
// (a signal computed and then discarded). It is stated rather than fixed, because
// the obvious remedy is to file a work item, and a work item for exactly this
// condition already exists, already fires, and already cannot be satisfied —
// see bugs_open/166. Adding a second unsatisfiable item is not closing a loop.
//
// It answers with the last-resort row rather than an error because refusing is not
// free: `head` has NO eligible component today (both candidates are is_active=false)
// and a site whose head slot goes unrendered also loses injectBrandHeadTags, i.e.
// its favicon and og-card. Louder, not more broken.
//
// The `eligible` flag comes back from the same query that picks the row, so the
// caller cannot be told the row is fine by one round trip and the truth by another.
func ResolveChromeComponent(ctx context.Context, db interface{}, function string, logger *zap.Logger) (*Component, bool, error) {
	eligible := chromeEligibleSQL("")
	query := `
		SELECT
			id,
			name,
			function,
			COALESCE(category, '') as category,
			html_template,
			input_schema,
			COALESCE(is_dark_section, false) as is_dark_section,
			(` + eligible + `) as chrome_eligible
		FROM content_components
		WHERE function = $1
		ORDER BY (` + eligible + `) DESC, name
		LIMIT 1
	`

	var comp Component
	var schemaJSON []byte
	var category sql.NullString
	var isEligible bool

	var err error
	switch d := db.(type) {
	case *sql.DB:
		err = d.QueryRowContext(ctx, query, function).Scan(
			&comp.ID, &comp.Name, &comp.Function, &category,
			&comp.HTMLTemplate, &schemaJSON, &comp.IsDarkSection, &isEligible,
		)
	case *pgxpool.Pool:
		err = d.QueryRow(ctx, query, function).Scan(
			&comp.ID, &comp.Name, &comp.Function, &category,
			&comp.HTMLTemplate, &schemaJSON, &comp.IsDarkSection, &isEligible,
		)
	default:
		return nil, false, fmt.Errorf("unsupported database type: %T", db)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, false, fmt.Errorf("no component serves chrome function %q", function)
		}
		return nil, false, fmt.Errorf("failed to resolve chrome component for %q: %w", function, err)
	}

	if category.Valid {
		comp.Category = category.String
	}
	if len(schemaJSON) > 0 {
		json.Unmarshal(schemaJSON, &comp.InputSchema)
	}

	if !isEligible && logger != nil {
		// ERROR, not Warn: there is no legitimate steady state in which the
		// library has nothing active to serve a chrome function, and the
		// symptom of the version that logged nothing was a deactivated
		// component rendering on every site for months.
		logger.Error("site chrome: no eligible component for function — falling back to an ineligible row",
			zap.String("function", function),
			zap.String("using_component", comp.Name),
			zap.String("component_id", comp.ID),
			zap.String("required", "is_active AND forked_from IS NULL AND component_level IN ("+chromeComponentLevels+")"),
		)
	}

	return &comp, isEligible, nil
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

// GetChromePinComponent dereferences a style-collection chrome PIN and reports
// whether the pinned component is still fit to serve chrome.
//
// This is the ONE way to read `style_collections.header_component_id` /
// `footer_component_id`. Both consumers — the build path (RenderHeader /
// RenderFooter) and the assignment path (link_site_components) — go through it,
// which is the whole point: until 2026-08-01 they each dereferenced the pin with
// GetComponentByID, whose SQL is `WHERE id = $1` with no predicate of any kind,
// so a pin bypassed bugs_open/118's predicate on BOTH paths (bugs_open/170).
//
// The assignment path is the half that made this urgent. link_site_components
// writes the pin into `site_components.component_id` and NULLs `rendered_html`,
// so an unguarded pin does not merely render a deactivated component — it
// overwrites the repair `repointRetiredChromeSlot` performed (bugs_open/166) and
// re-creates 118 in the column 118 fixed.
//
// It returns the component either way rather than an error, matching
// ResolveChromeComponent: the caller decides what an ineligible answer means, and
// both callers use it to fall through to the eligible-only pool lookup. Returning
// the row keeps that decision at the call site, where the fallback lives.
//
// NOT a change to GetComponentByID: that is a general by-id fetch, used by
// RenderComponentAction to render arbitrary page sections (v3_site_actions.go),
// where a chrome-level predicate would be wrong.
func GetChromePinComponent(ctx context.Context, db interface{}, id uuid.UUID, logger *zap.Logger) (*Component, bool, error) {
	eligible := chromePinEligibleSQL("")
	query := `
		SELECT
			id,
			name,
			function,
			COALESCE(category, '') as category,
			html_template,
			input_schema,
			COALESCE(is_dark_section, false) as is_dark_section,
			(` + eligible + `) as pin_eligible
		FROM content_components
		WHERE id = $1
		LIMIT 1
	`

	var comp Component
	var schemaJSON []byte
	var category sql.NullString
	var isEligible bool

	var err error
	switch d := db.(type) {
	case *sql.DB:
		err = d.QueryRowContext(ctx, query, id).Scan(
			&comp.ID, &comp.Name, &comp.Function, &category,
			&comp.HTMLTemplate, &schemaJSON, &comp.IsDarkSection, &isEligible,
		)
	case *pgxpool.Pool:
		err = d.QueryRow(ctx, query, id).Scan(
			&comp.ID, &comp.Name, &comp.Function, &category,
			&comp.HTMLTemplate, &schemaJSON, &comp.IsDarkSection, &isEligible,
		)
	default:
		return nil, false, fmt.Errorf("unsupported database type: %T", db)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			// A pin naming a component that no longer exists. The FK makes this
			// unreachable today; treated as "not eligible, no row" rather than a
			// hard error so a caller still falls through to the library.
			return nil, false, fmt.Errorf("chrome pin %s names no component", id)
		}
		return nil, false, fmt.Errorf("failed to dereference chrome pin %s: %w", id, err)
	}

	if category.Valid {
		comp.Category = category.String
	}
	if len(schemaJSON) > 0 {
		json.Unmarshal(schemaJSON, &comp.InputSchema)
	}

	// The eligibility flag comes back from the SAME query that fetched the row, so
	// a caller cannot be told the component is fine by one round trip and the truth
	// by another. Same construction, and same reason, as ResolveChromeComponent.
	return &comp, isEligible, nil
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
//
// SIBLING, kept honest by cross-reference (council corr b72a4029 r1):
// datahelpers.ContentDataCanFillTemplate asks the ROUTING form of this
// question — "does content_data hold ANY of the template's content?" — with a
// deliberately flat field scan, because this package imports its consumer and
// cannot be imported back. If you change what counts as "a field this
// template reads" here, check that helper's templateTopLevelFieldRe for the
// same case, and vice versa — the two must not silently disagree.
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

// RenderTemplate renders a component template. IT IS THE ONLY SPELLING — see
// render_seam_one_spelling_test.go, which fails the build if a second one
// appears (owner ruling 2026-08-21, RFC_041 §5).
//
// WHY ONE. Until 2026-08-21 there were two: this function, and a one-line
// `RenderTemplate` wrapper that discarded the two reports below and returned
// just (string, error). Nine of the twelve call sites used the short one, which
// is how bugs_open/238 shipped five <img src=""> to a live homepage while the
// call had the field names in hand — the discard was INSIDE the wrapper, where
// no reviewer of the call site could see it. Now a caller that does not want
// the reports must write `out, _, _, err :=` and the discard is in the diff,
// which is the whole difference between a convenience and a silent default.
//
// The 238 lane's council round named this move and declined to make it inside a
// bug fix ("changes the primitive every render flows through — the RFC-shaped
// move, not a rider"). It became cheap once bugs_open/260 gave every caller an
// error to handle, which is why it is done now and not then.
//
// It reports which bare output placeholders rendered empty (`missing`) and which
// of those sat inside an href=/src= attribute (`inURLAttr` — a dead control on
// a live page). A blanked URL attribute logs at Error with its field names; any
// other blanked field logs at Warn. This replaces the previous count-only
// <no value> log, which named nothing and let 30 dead controls ship silently on
// idea.uk (bugs_open/018). Uses Go's text/template for full support of {{if}},
// {{range}}, {{with}}, etc.
func RenderTemplate(templateStr string, ctx *RenderContext, logger *zap.Logger) (string, []string, []string, error) {
	if templateStr == "" {
		// Not an error: an intentionally empty template stub is a real thing on
		// this estate (rerender_page_sections carries such a section rather than
		// failing it). "Executed and produced nothing" is a different fact from
		// "could not execute", and only the second one is the error channel's.
		//
		// RenderedTemplateSHA is deliberately NOT set here. Empty means unknown,
		// and "there was no template" is a kind of unknown — stamping the digest of
		// the empty string would give a caller a provenance token pointing at a
		// template that renders nothing, which is worse than no token (RFC_046).
		return "", nil, nil, nil
	}

	// The one fact only this seam knows: WHICH template text produced the bytes it
	// is about to return (RFC_046, ruled 2026-08-22). Computed before execution so
	// it describes the input, not anything the render may mutate, and reported
	// rather than acted on — resolving it to a component_versions row needs a
	// database handle this function does not have. See the field's doc comment.
	if ctx != nil {
		sum := sha256.Sum256([]byte(templateStr))
		ctx.RenderedTemplateSHA = hex.EncodeToString(sum[:])
	}

	// sanitiseFormAction's presence gate ("Absent is only a defect if this
	// component actually has a form to point somewhere") is meant to be proxied
	// by the template itself, not by whether the content-generation schema
	// remembered to author a form_action field. contact-block (bugs_open/228)
	// never asked its schema for one, so ctx.ContentData["form_action"] was
	// never present, the gate silently declined to act, and the form shipped
	// with no action at all. Seeding an empty placeholder — already one of
	// nonDeliveringFormActions' recognised values — makes the gate's own
	// stated condition mechanical: any template that references form_action
	// gets the sanitiser's protection, with zero dependency on content
	// authoring remembering to supply the field. A template that never
	// mentions form_action gains nothing.
	if strings.Contains(templateStr, "form_action") {
		if ctx.ContentData == nil {
			ctx.ContentData = map[string]interface{}{}
		}
		if _, present := ctx.ContentData["form_action"]; !present {
			ctx.ContentData["form_action"] = ""
			logger.Debug("RenderTemplate: seeded empty form_action for sanitiser",
				zap.String("template_preview", datahelpers.TruncateString(templateStr, 100)))
		}
	}

	// A template that namespaces its element ids with {{.InstanceID}} and is
	// rendered by a path that never bound one gets missingkey=zero — an empty
	// string — so every instance on the page lands back on IDENTICAL ids, which
	// is the exact silent failure bugs_open/283 exists to remove. Measured
	// 2026-08-16: eight non-test files hold fourteen calls to a RenderTemplate*
	// helper, and five of those files bound nothing before this change.
	// Enumerating call sites is what goes stale, so the report lives HERE, where
	// every one of them arrives, and mirrors the form_action seeding above: the
	// guarantee is made mechanical rather than left to each caller remembering.
	//
	// It refuses and does not substitute. There is no value this layer could
	// invent that would be right: it cannot see the page, so any token it made up
	// would either collide (no better than empty) or disagree with the token the
	// page's other render paths use for the SAME instance — which is worse than
	// empty, because the ids would then depend on which action last touched the
	// section.
	//
	// ARMED FROM LOG-ONLY TO REFUSAL, 2026-08-24 (plan §B2, RFC_032 §10). It
	// stood here as a named logger.Error for eight days, and this estate has an
	// owner ruling — cited at the dead-control report below — that a named log
	// is not escalation. The output is structurally wrong for EVERY instance,
	// which is the bugs_open/260 class: execute, or fail; there is no third
	// state. It is deliberately unlike the absent-content report further down,
	// which does not refuse because that content legitimately renders today.
	//
	// UNCONDITIONAL, NOT OPT-IN, and the measurements are why. The owner ruling
	// of 2026-08-02 §2 requires new authority on a shared seam to ship opt-in
	// with the unsafe default OFF *or* to be measured inert; the second arm is
	// satisfied here, and a per-caller flag would re-create exactly the
	// per-call-site wiring this seam exists to remove. What was measured, all
	// dated 2026-08-24 and all disconfirmable:
	//
	//   - STATIC, tree-wide: 11 non-test files call a RenderTemplate* helper;
	//     6 bind a token, 5 are allow-listed in pattern-check.py's
	//     INSTANCE_TOKEN_ALLOWED as slots that occur once per document. Zero
	//     unbound. Demand control: deleting the bind call from each of the 6
	//     flips it to an unscoped-component-render finding, so the zero could
	//     have come out otherwise.
	//   - AT THE ARTEFACT: of 2,020 page_components rows, ZERO carry the
	//     unbound-token shape id="-…"; 374 carry a bound id="c-…". In the
	//     window since generic-text-block began spelling id="{{.InstanceID}}"
	//     exactly (2026-08-23 12:32), 155 of its rows were written and 155 of
	//     155 carry a bound token, 0 empty. That template is the one where an
	//     unbound render is visible as id="" rather than id="-…", so it is the
	//     sharpest available detector, and it says the live paths all bind.
	//   - BLAST RADIUS: zero chrome-level templates (header 4, footer 1, site
	//     6, element 1 — all active) spell {{.InstanceID}}, so the refusal
	//     cannot fail a header, footer or <head> render. All 140 that do spell
	//     it are section (30) or tool (110) level.
	//   - NOT MEASURED, stated rather than implied: the fleet log census the
	//     plan asked for is a BLIND instrument on this cluster and its zero is
	//     worth nothing — there is no log aggregator, spawned job pods carry
	//     ttlSecondsAfterFinished=3600, and across all 181 live and completed
	//     pods (176k log lines) not one line names component_library.go at all.
	//     The artefact census above is the substitute, because it can come out
	//     non-zero.
	//
	// SECOND DOOR, unclosed on purpose: RenderTemplateWithMap
	// (rerender_pages_actions.go:818) is a separate render path with no such
	// check (bugs_open/260 §13g — it does not even share this one's FuncMap).
	// Its callers render chrome only, and the chrome census above is zero, so
	// arming it would guard nothing today. If a section or tool template ever
	// reaches it, this refusal will not fire.
	if TemplateNeedsInstanceID(templateStr) {
		token, _ := ctx.ContentData[InstanceContentKey].(string)
		if token == "" {
			logger.Error("RenderTemplate: template namespaces ids with {{."+InstanceContentKey+"}} but no per-instance token was bound — refusing rather than rendering identical element ids",
				zap.String("template_preview", datahelpers.TruncateString(templateStr, 100)),
			)
			return "", nil, nil, fmt.Errorf(
				"template namespaces ids with {{.%s}} but no per-instance token is bound — bind via BindInstanceToken/BindSingleSectionInstanceToken (bugs_open/283 seam)",
				InstanceContentKey)
		}
	}

	// Convert context to map[string]interface{} - preserves nested structures
	data := contextToInterfaceMap(ctx)

	// Log what we're about to render (debug)
	logger.Debug("RenderTemplate: executing",
		zap.Int("data_fields", len(data)),
		zap.String("template_preview", datahelpers.TruncateString(templateStr, 100)),
	)

	// Execute, or fail. There is no third state (bugs_open/260).
	//
	// What stood here was a silent fallback to a regex renderer written for
	// HANDLEBARS syntax: on ANY error it logged at Warn and substituted
	// {{.field}} values while leaving every {{if}}, {{range}} and {{end}} it
	// could not execute sitting in the output. The result was well-formed HTML
	// with the control directives still in it and the values already resolved —
	// so nothing downstream could tell a rendered page from an unrendered one
	// except a final regex check, which refused the whole page. 26 occurrences
	// across 7 domains in the eight days to 2026-08-18, every one of them an
	// EXECUTE error (a content field of the wrong shape), never a parse error.
	//
	// Deleting it was measured, not argued: 0 of 251 active component templates
	// fail to Parse, 0 of 1,778 stored sections fail to Execute against their
	// own content_data, and 0 of 253 active components are even written in the
	// fallback's dialect ({{#…}}, {{nav_items_html}}, {{quick_links_html}}).
	// Each figure had a control that could have come out otherwise.
	//
	// ⚠ THE RETIRED DIALECT IS NOW UNRENDERABLE, not merely unused: a template
	// naming {{nav_items_html}} cannot Parse at all ("function not defined"),
	// so it hard-fails here rather than being quietly regex-patched. Nav HTML
	// reaches a template through ctx.NavItems and the FuncMap, nowhere else.
	result, err := executeGoTemplate(templateStr, data, logger)
	if err != nil {
		// Error, not Warn: this is a section that will not exist. The wrapped
		// error from executeGoTemplate already carries the template line, the
		// field and the offending value ("range can't iterate over …"), which
		// is the diagnosis the 26 occurrences never had.
		logger.Error("RenderTemplate: template execution failed — refusing to emit output that was not executed",
			zap.Error(err),
			zap.String("template_preview", datahelpers.TruncateString(templateStr, 100)),
		)
		// No missing-field report on this path: there is no output to report
		// against, and a list of empty placeholders would read as if a render
		// had happened.
		return "", nil, nil, fmt.Errorf("component template execution failed: %w", err)
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

	// ── bugs_open/342: an ABSENT REQUIRED field is not a Warn ─────────────
	//
	// missingkey=zero renders a field the content never supplied as empty, with
	// no error, and page assembly drops a visually-empty section — so the
	// content does not arrive broken, it does not arrive at all. That is the
	// mechanism that blanked article bodies fleet-wide (bugs_closed/004/005).
	//
	// The rule is not new and is deliberately NOT re-derived here:
	// missingRequiredLLMFields is the same function the two gated call sites
	// call before rendering, so the seam and the pre-check cannot disagree about
	// what "required" or "empty" means. What is new is WHERE it runs — at the
	// seam, so all fifteen call sites get it, instead of at the two that
	// remembered.
	//
	// REPORT ONLY, no refusal. Refusing here would be new authority over content
	// that renders successfully today at thirteen sites that never asked for it
	// (owner ruling 2026-08-02 §2), and the two paths that DO want to refuse
	// already do, before the render, where refusing is cheaper. What this closes
	// is the SILENCE — 342's actual complaint. Error, not Warn, because a
	// required field is a stated contract and the section is about to disappear:
	// the sibling report below is Warn for a merely-empty optional field and
	// Error for a dead URL control, and this belongs with the latter.
	//
	// ⚠ IT JUDGES `data`, NOT ctx.ContentData, AND THE TWO GIVE DIFFERENT
	// ANSWERS ON PURPOSE. contextToInterfaceMap supplies fleet defaults —
	// cta_text falls back to "Get Started", colours to the house palette — so a
	// required field the writer never produced can still be non-empty by the
	// time the template sees it. The pre-render gate asks "did the WRITER
	// supply it?" and refuses on that. This asks "will it RENDER EMPTY?",
	// because the damage 342 describes is empty -> page assembly drops the
	// section -> the content vanishes. A defaulted field renders something, so
	// the section survives and there is nothing to report.
	//
	// The consequence, stated so nobody reads a silence as agreement: this
	// report is a SUBSET of what the pre-render gate would name. Both are
	// correct about their own question, and render_seam_absent_required_test.go
	// pins the divergence with the reason so the next reader does not "fix" one
	// to match the other.
	if len(ctx.InputSchema) > 0 {
		if absentRequired := missingRequiredLLMFields(ctx.InputSchema, data); len(absentRequired) > 0 {
			// PUBLISHED, not just logged (council bb7f5d0e round 1,
			// bug_historian, GATING). The first version of this reported at
			// Error and stopped there — and this estate has an owner ruling that
			// a named log is NOT escalation (bugs_open/054, 2026-07-22), earned
			// on precisely this shape: bugs_open/018 shipped the observability
			// half for dead controls, 30 of them shipped anyway, and the owner
			// ruled "make it MEAN something". Writing the same half again and
			// noting the limit in a comment would have been the same mistake
			// with a footnote.
			//
			// The seam still cannot escalate — it has no database handle and no
			// site identity — so it publishes here and a caller that has both
			// acts. See render_site_components_action.go, which files a
			// required_fields_missing item: the item type that ALREADY EXISTS
			// for this defect, with a router seeded (bugs_open/277), rather than
			// a fourth type of my own.
			ctx.AbsentRequiredFields = absentRequired
			logger.Error("RenderTemplate: REQUIRED content field(s) absent — the section rendered empty and page assembly will drop it (bugs_open/342)",
				zap.Strings("absent_required_fields", absentRequired),
				zap.String("template_preview", datahelpers.TruncateString(templateStr, 100)),
			)
		}
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

	return result, missing, inURLAttr, nil
}

// ── RETIRED 2026-08-19 (bugs_open/260): the handlebars regex renderers ──
//
// contextToMap, renderEachBlocks, renderIfBlocks, renderGoStyleSubstitutions,
// renderHandlebarsSubstitutions, applyBidirectionalAliases,
// RenderTemplateWithValidation (already caller-less) and
// validateContentAgainstSchema were deleted with the silent fallback they
// existed to serve. They spoke a DIFFERENT DIALECT from every template on the
// estate — {{#each}}, {{nav_items_html}}, {{quick_links_html}} — so when
// text/template failed they substituted values and left the control directives
// in the output, which is bugs_open/260 in one sentence.
//
// Measured before deletion (2026-08-19, each with a control that could have
// come out otherwise): 0 of 253 active components use that dialect, 0 of 251
// active templates fail to Parse, 0 of 1,778 stored sections fail to Execute.
//
// contextToMap's death also removes bugs_open/203's last two fabricate-a-URL
// members (primary_cta_url → /contact.html, secondary_cta_url → /about.html).
// The surviving map builder is contextToInterfaceMap, which sets cta_url from
// the context and invents nothing.
//
// DO NOT REINSTATE ONE OF THESE AS A "SMALL" REPAIR. Any of them, re-added,
// restores the third state this seam exists to remove: output that no template
// engine produced.

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

	// The scalar half of the template contract is DERIVED from the struct's
	// json tags (bugs_open/109) — every tagged string field except the control
	// fields is advertised, so a new RenderContext field reaches templates
	// without anyone remembering this map. The decorations after the loop are
	// what tags cannot express: the computed logo_text, render-time defaults,
	// and the contact_email alias.
	result := make(map[string]interface{}, 40)
	for key, value := range renderContextScalarFields(ctx) {
		if _, control := renderContextControlFields[key]; control {
			continue
		}
		result[key] = value
	}
	result["logo_text"] = logoText
	result["company_name"] = defaultString(ctx.CompanyName, logoText)
	result["primary_color"] = defaultString(ctx.PrimaryColor, "#1a1a2e")
	result["secondary_color"] = defaultString(ctx.SecondaryColor, "#2d2d44")
	result["accent_color"] = defaultString(ctx.AccentColor, "#16a085")
	result["text_color"] = defaultString(ctx.TextColor, "#333333")
	result["background_color"] = defaultString(ctx.BackgroundColor, "#ffffff")
	result["cta_text"] = defaultString(ctx.CTAText, "Get Started")
	// Correct-or-absent (LNK-005) — see the matching comment in contextToMap.
	// This is the path RenderTemplate actually calls (render_site_components_
	// action.go:971), so this is where the phantom /contact.html link shipped
	// from: cta_text got overwritten from ContentData below, cta_url didn't
	// (ContentData had no such key), and the mismatched pair passed the
	// template's `{{if and .cta_text .cta_url}}` guard because the fake
	// default made .cta_url non-empty.
	result["cta_url"] = ctx.CTAUrl
	result["contact_email"] = ctx.Email

	// Composites — these need shaping, not copying, so they stay explicit.
	result["site_id"] = ctx.SiteID
	result["nav_items"] = ctx.NavItems
	result["services"] = ctx.Services

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
		// The pin is dereferenced through GetChromePinComponent, never
		// GetComponentByID: the latter is `WHERE id = $1` with no predicate, which
		// is how a deactivated component pinned here rendered anyway on three
		// deployed sites (bugs_open/170).
		//
		// An ineligible pin leaves comp nil and FALLS THROUGH to the by-function
		// branch below — the eligible-only pool lookup that bugs_open/167 fixed.
		// Falling through, rather than reporting and rendering it anyway, is what
		// makes the pin path agree with the assignment path: for the three sites
		// this affects, the pool answer is the very component bugs_open/166 already
		// repointed their `site_components` row to.
		pinned, eligible, pinErr := GetChromePinComponent(ctx, db, *coll.HeaderComponentID, logger)
		if pinErr != nil {
			logger.Warn("Failed to get header component", zap.Error(pinErr))
		} else if !eligible {
			logger.Warn("site chrome: style collection pins an ineligible header — using the library instead",
				zap.String("collection", coll.Name),
				zap.String("pinned_component", pinned.Name),
				zap.String("component_id", pinned.ID),
				zap.String("required", "is_active AND component_level IN ("+chromeComponentLevels+")"),
			)
		} else {
			comp = pinned
			source = fmt.Sprintf("component-db:%s", coll.Name)
		}
	}

	// Fallback: try by function name. ResolveChromeComponent, not
	// GetComponentByFunction — the latter has no component_level filter, so the
	// site-header pool it selects from includes `site-header` itself, a
	// component_level='section' page-section component (bugs_open/167).
	if comp == nil {
		chromeComp, eligible, resolveErr := ResolveChromeComponent(ctx, db, ChromeSlotFunction("header"), logger)
		if resolveErr != nil || !eligible {
			// !eligible is a gate, not advice: the resolver ALWAYS answers, and its
			// answer when nothing is eligible is the last-resort row — which for a
			// chrome function is precisely the section component this bug is about.
			// RenderFallbackHeader is at least a header.
			logger.Warn("No eligible header component in the library, using fallback",
				zap.String("function", ChromeSlotFunction("header")),
				zap.Bool("eligible", eligible),
				zap.Error(resolveErr))
			header := RenderFallbackHeader(renderCtx)
			return fmt.Sprintf("<!-- HEADER SOURCE: fallback -->\n%s", header), nil
		}
		comp = chromeComp
		source = "component-db:" + comp.Name
	}

	// Render template. An execution failure is returned, not swallowed
	// (bugs_open/260): InjectHeader already answers an error from here with
	// RenderFallbackHeader, so the page gets well-formed fallback chrome
	// instead of a header with {{if}} directives left in it. Chrome needed
	// error PLUMBING, not new mechanism — the ladder was already built.
	renderCtx.InputSchema = comp.InputSchema // bugs_open/342
	rendered, _, _, err := RenderTemplate(comp.HTMLTemplate, renderCtx, logger)
	if err != nil {
		return "", fmt.Errorf("header component %q failed to render: %w", comp.Name, err)
	}
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
		// See the note in RenderHeader. Four collections — one more than the header
		// case, because leopardess's header pin is legitimate and its footer pin is
		// not — pin `footer-4-column`, is_active=false (bugs_open/170).
		pinned, eligible, pinErr := GetChromePinComponent(ctx, db, *coll.FooterComponentID, logger)
		if pinErr != nil {
			logger.Warn("Failed to get footer component", zap.Error(pinErr))
		} else if !eligible {
			logger.Warn("site chrome: style collection pins an ineligible footer — using the library instead",
				zap.String("collection", coll.Name),
				zap.String("pinned_component", pinned.Name),
				zap.String("component_id", pinned.ID),
				zap.String("required", "is_active AND component_level IN ("+chromeComponentLevels+")"),
			)
		} else {
			comp = pinned
			source = fmt.Sprintf("component-db:%s", coll.Name)
		}
	}

	// ResolveChromeComponent, not GetComponentByFunction — see RenderHeader above
	// and bugs_open/167. `site-footer` the SECTION component sits in the same pool
	// as `footer-theme-chrome` the chrome component, and only ORDER BY name
	// currently separates them.
	if comp == nil {
		chromeComp, eligible, resolveErr := ResolveChromeComponent(ctx, db, ChromeSlotFunction("footer"), logger)
		if resolveErr != nil || !eligible {
			logger.Warn("No eligible footer component in the library, using fallback",
				zap.String("function", ChromeSlotFunction("footer")),
				zap.Bool("eligible", eligible),
				zap.Error(resolveErr))
			footer := RenderFallbackFooter(renderCtx)
			return fmt.Sprintf("<!-- FOOTER SOURCE: fallback -->\n%s", footer), nil
		}
		comp = chromeComp
		source = "component-db:" + comp.Name
	}

	// See RenderHeader: the error goes back to InjectFooter's existing fallback
	// branch rather than being rendered around (bugs_open/260).
	renderCtx.InputSchema = comp.InputSchema // bugs_open/342
	rendered, _, _, err := RenderTemplate(comp.HTMLTemplate, renderCtx, logger)
	if err != nil {
		return "", fmt.Errorf("footer component %q failed to render: %w", comp.Name, err)
	}
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

	// The container is gated on its contents, matching the DB footer
	// components' {{if}} gates (bugs_open/111): a site with no contact route
	// must not render a bare "Contact" heading over empty space.
	contactHTML := ""
	if ctx.Email != "" {
		contactHTML = fmt.Sprintf(`<div class="footer-contact"><h4>Contact</h4><p>%s</p></div>`, ctx.Email)
	}

	return fmt.Sprintf(`<footer class="site-footer">
    <div class="footer-container">
        <div class="footer-brand"><h3>%s</h3><p>%s</p></div>
        <div class="footer-links"><h4>Links</h4><ul>%s</ul></div>
        %s
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
</style>`, ctx.LogoText, ctx.Tagline, navHTML, contactHTML, year, ctx.CompanyName)
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
		// Labelled, because since bugs_open/260 this branch is also how a
		// FAILED RENDER lands here, not only a missing component — and an
		// operator looking at a plain-looking header needs to be able to grep
		// for which it was. The strip regex below matches
		// `<!--\s*HEADER\s+SOURCE:[^>]*-->`, which this text still satisfies.
		logger.Warn("Failed to render header, using fallback", zap.Error(err))
		headerHTML = "<!-- HEADER SOURCE: fallback (render error) -->\n" + RenderFallbackHeader(renderCtx)
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
		// See InjectHeader: labelled so a render failure is distinguishable
		// from an absent component (bugs_open/260).
		logger.Warn("Failed to render footer, using fallback", zap.Error(err))
		footerHTML = "<!-- FOOTER SOURCE: fallback (render error) -->\n" + RenderFallbackFooter(renderCtx)
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

	// Get head component — chrome selection, so ResolveChromeComponent (bugs_open/167).
	//
	// This is the call site that makes the `eligible` check a GATE rather than a
	// hint. Live 2026-07-31 the library has NO eligible head component (both
	// candidates are is_active=false), and the resolver answers anyway — by design,
	// so the gap can be reported rather than swallowed. Its answer is `Document
	// Head`, an 8,523-char component_level='section' component. Using it because a
	// row came back would render a page section as <head>, i.e. it would CREATE
	// bugs_open/167 on the one slot that did not have it. Ineligible ⇒ fallback,
	// which is byte-for-byte what runs today when the lookup finds nothing.
	comp, eligible, resolveErr := ResolveChromeComponent(ctx, db, ChromeSlotFunction("head"), logger)
	if resolveErr != nil || !eligible {
		logger.Warn("RenderHead: no eligible head component in the library, using fallback",
			zap.String("function", ChromeSlotFunction("head")),
			zap.Bool("eligible", eligible),
			zap.Error(resolveErr))
		head := RenderFallbackHead(renderCtx)
		return fmt.Sprintf("<!-- HEAD SOURCE: fallback -->\n%s", head), nil
	}
	source = "component-db:" + comp.Name

	// Render template with context (title, description, colors etc.). As with
	// the header and footer, an execution failure returns to InjectHead's
	// existing fallback branch (bugs_open/260) — a <head> carrying unexecuted
	// directives is worse than the plain fallback head.
	renderCtx.InputSchema = comp.InputSchema // bugs_open/342
	rendered, _, _, err := RenderTemplate(comp.HTMLTemplate, renderCtx, logger)
	if err != nil {
		return "", fmt.Errorf("head component %q failed to render: %w", comp.Name, err)
	}

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
		// No SOURCE marker here: unlike the header and footer, the head is
		// stripped and rebuilt by tag, and a comment before <head> would sit
		// outside the element. The log line is the provenance record
		// (bugs_open/260).
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
