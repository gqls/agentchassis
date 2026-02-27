// FILE: platform/orchestration/actions/fix_forced_text_colors_action.go
//
// FixForcedTextColorsAction removes explicit color declarations from child
// text elements (h1-h6, p, li, blockquote, strong, em, cite) in component
// inline <style> blocks. These forced colors override the --section-*
// inheritance model and break dark sections.
//
// Before removing a forced color, the action validates that the resulting
// text/background combination meets WCAG AA contrast ratio (4.5:1) using
// the site's actual color palette.
//
// For dark-section components missing the --section-* contract, the action
// injects the contract variables before removing forced text colors.
//
// Two levels:
//   1. content_components.html_template — permanent fix for future renders
//   2. page_components.rendered_html    — immediate fix for stored HTML
//
// Used by: color-variable-fixer agent (dispatched by improvement loop)
//
// Config:
//   - fix_templates: bool (default true) — fix content_components.html_template
//   - fix_rendered:  bool (default true) — fix page_components.rendered_html
//   - min_contrast:  float (default 4.5) — minimum WCAG contrast ratio
//
// Data inputs (via ActionInputSpec):
//   - site_id (required)

package actions

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ============================================================================
// ActionInputSpec
// ============================================================================

var FixForcedTextColorsInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id"},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("fix_forced_text_colors", FixForcedTextColorsInputSpec)
}

// ============================================================================
// ACTION: fix_forced_text_colors
// ============================================================================

func FixForcedTextColorsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "fix_forced_text_colors"),
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	logger.Info("FixForcedTextColorsAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	// --- Resolve site_id ---
	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		FixForcedTextColorsInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	config := params.StepConfig.Config
	fixTemplates := true
	fixRendered := true
	minContrast := 4.5
	if v, ok := config["fix_templates"].(bool); ok {
		fixTemplates = v
	}
	if v, ok := config["fix_rendered"].(bool); ok {
		fixRendered = v
	}
	if v, ok := config["min_contrast"].(float64); ok {
		minContrast = v
	}

	// --- Load site's color palette ---
	palette, err := loadSitePalette(ctx, params.DB, siteID, logger)
	if err != nil {
		logger.Warn("FixForcedTextColorsAction: could not load palette, using defaults",
			zap.Error(err))
		palette = defaultPalette()
	}

	logger.Info("FixForcedTextColorsAction: Palette loaded",
		zap.String("site_id", siteIDStr),
		zap.String("text", palette.text),
		zap.String("primary", palette.primary),
		zap.String("background", palette.background),
	)

	templatesFixed := 0
	renderedFixed := 0
	skippedContrast := 0
	contractsAdded := 0
	var details []map[string]interface{}

	// =========================================================================
	// 1. Fix content_components.html_template
	// =========================================================================
	if fixTemplates {
		tf, ca, sc, td := fixTemplateTextColors(ctx, params.DB, siteID, palette, minContrast, logger)
		templatesFixed = tf
		contractsAdded += ca
		skippedContrast += sc
		details = append(details, td...)
	}

	// =========================================================================
	// 2. Fix page_components.rendered_html
	// =========================================================================
	if fixRendered {
		rf, ca, sc, rd := fixRenderedTextColors(ctx, params.DB, siteID, palette, minContrast, logger)
		renderedFixed = rf
		contractsAdded += ca
		skippedContrast += sc
		details = append(details, rd...)
	}

	totalFixed := templatesFixed + renderedFixed

	logger.Info("FixForcedTextColorsAction: Complete",
		zap.Int("templates_fixed", templatesFixed),
		zap.Int("rendered_fixed", renderedFixed),
		zap.Int("contracts_added", contractsAdded),
		zap.Int("skipped_low_contrast", skippedContrast),
	)

	return map[string]interface{}{
		"site_id":              siteIDStr,
		"templates_fixed":      templatesFixed,
		"rendered_fixed":       renderedFixed,
		"contracts_added":      contractsAdded,
		"skipped_low_contrast": skippedContrast,
		"total_fixed":          totalFixed,
		"details":              details,
		"needs_rerender":       totalFixed > 0,
	}, nil
}

// ============================================================================
// Palette loading
// ============================================================================

type sitePalette struct {
	primary    string // e.g. "#1a1a2e"
	text       string // e.g. "#333333"
	background string // e.g. "#ffffff"
	accent     string // e.g. "#0f3460"
}

func defaultPalette() sitePalette {
	return sitePalette{
		primary:    "#1a1a2e",
		text:       "#333333",
		background: "#ffffff",
		accent:     "#0f3460",
	}
}

func loadSitePalette(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) (sitePalette, error) {
	p := defaultPalette()

	// Query via style_collections joined through sites
	var primaryVal, textVal, bgVal, accentVal sql.NullString
	err := db.QueryRowContext(ctx, `
		SELECT
			sc.color_palette->>'primary',
			sc.color_palette->>'text',
			sc.color_palette->>'background',
			sc.color_palette->>'accent'
		FROM sites s
		JOIN style_collections sc ON s.style_collection_id = sc.id
		WHERE s.id = $1
	`, siteID).Scan(&primaryVal, &textVal, &bgVal, &accentVal)

	if err != nil {
		return p, fmt.Errorf("query palette: %w", err)
	}

	if primaryVal.Valid && primaryVal.String != "" {
		p.primary = primaryVal.String
	}
	if textVal.Valid && textVal.String != "" {
		p.text = textVal.String
	}
	if bgVal.Valid && bgVal.String != "" {
		p.background = bgVal.String
	}
	if accentVal.Valid && accentVal.String != "" {
		p.accent = accentVal.String
	}

	return p, nil
}

// ============================================================================
// Template fixing (content_components)
// ============================================================================

func fixTemplateTextColors(
	ctx context.Context, db *sql.DB, siteID uuid.UUID,
	palette sitePalette, minContrast float64, logger *zap.Logger,
) (fixed, contractsAdded, skippedContrast int, details []map[string]interface{}) {

	// Find content_components used by this site that have forced text colors
	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT cc.id, cc.function, cc.html_template, cc.is_dark_section
		FROM content_components cc
		WHERE cc.is_active = true
		  AND cc.component_level = 'section'
		  AND cc.html_template ~ '[\{;\s]color:\s*#[0-9a-fA-F]{3,8}'
		  AND (
		    EXISTS (SELECT 1 FROM site_components sc WHERE sc.site_id = $1 AND sc.component_id = cc.id)
		    OR EXISTS (
		      SELECT 1 FROM page_components pc
		      JOIN pages p ON pc.page_id = p.id
		      WHERE p.site_id = $1 AND pc.component_id = cc.id
		    )
		  )
	`, siteID)
	if err != nil {
		logger.Error("fixTemplateTextColors: query failed", zap.Error(err))
		return
	}
	defer rows.Close()

	for rows.Next() {
		var compID uuid.UUID
		var function, template string
		var isDarkSection bool
		if err := rows.Scan(&compID, &function, &template, &isDarkSection); err != nil {
			continue
		}

		result := processComponentCSS(template, function, isDarkSection, palette, minContrast, logger)
		if !result.changed {
			if result.skippedContrast {
				skippedContrast++
				details = append(details, map[string]interface{}{
					"layer":    "template",
					"function": function,
					"action":   "skipped_low_contrast",
					"ratio":    result.worstRatio,
				})
			}
			continue
		}

		_, err := db.ExecContext(ctx, `
			UPDATE content_components SET html_template = $1, updated_at = NOW() WHERE id = $2
		`, result.html, compID)
		if err != nil {
			logger.Error("fixTemplateTextColors: update failed",
				zap.String("function", function), zap.Error(err))
			continue
		}

		fixed++
		if result.contractAdded {
			contractsAdded++
		}

		details = append(details, map[string]interface{}{
			"layer":          "template",
			"function":       function,
			"id":             compID.String(),
			"changed":        true,
			"colors_removed": result.colorsRemoved,
			"contract_added": result.contractAdded,
			"worst_ratio":    result.worstRatio,
		})

		logger.Info("fixTemplateTextColors: Fixed component",
			zap.String("function", function),
			zap.Int("colors_removed", result.colorsRemoved),
			zap.Bool("contract_added", result.contractAdded),
			zap.Float64("worst_ratio", result.worstRatio),
		)
	}
	return
}

// ============================================================================
// Rendered HTML fixing (page_components)
// ============================================================================

func fixRenderedTextColors(
	ctx context.Context, db *sql.DB, siteID uuid.UUID,
	palette sitePalette, minContrast float64, logger *zap.Logger,
) (fixed, contractsAdded, skippedContrast int, details []map[string]interface{}) {

	rows, err := db.QueryContext(ctx, `
		SELECT pc.id, pc.slot_name, pc.rendered_html, p.name as page_name,
		       COALESCE(cc.is_dark_section, false) as is_dark_section,
		       COALESCE(cc.function, pc.slot_name) as function
		FROM page_components pc
		JOIN pages p ON pc.page_id = p.id
		LEFT JOIN content_components cc ON pc.component_id = cc.id
		WHERE p.site_id = $1
		  AND pc.rendered_html ~ '[\{;\s]color:\s*#[0-9a-fA-F]{3,8}'
		  AND pc.rendered_html LIKE '%<style%'
	`, siteID)
	if err != nil {
		logger.Error("fixRenderedTextColors: query failed", zap.Error(err))
		return
	}
	defer rows.Close()

	for rows.Next() {
		var pcID uuid.UUID
		var slotName, renderedHTML, pageName, function string
		var isDarkSection bool
		if err := rows.Scan(&pcID, &slotName, &renderedHTML, &pageName, &isDarkSection, &function); err != nil {
			continue
		}

		result := processComponentCSS(renderedHTML, function, isDarkSection, palette, minContrast, logger)
		if !result.changed {
			if result.skippedContrast {
				skippedContrast++
			}
			continue
		}

		_, err := db.ExecContext(ctx, `
			UPDATE page_components SET rendered_html = $1 WHERE id = $2
		`, result.html, pcID)
		if err != nil {
			logger.Error("fixRenderedTextColors: update failed",
				zap.String("slot", slotName), zap.String("page", pageName), zap.Error(err))
			continue
		}

		fixed++
		if result.contractAdded {
			contractsAdded++
		}

		details = append(details, map[string]interface{}{
			"layer":          "rendered",
			"page":           pageName,
			"slot_name":      slotName,
			"changed":        true,
			"colors_removed": result.colorsRemoved,
			"worst_ratio":    result.worstRatio,
		})

		logger.Info("fixRenderedTextColors: Fixed rendered component",
			zap.String("page", pageName),
			zap.String("slot", slotName),
			zap.Int("colors_removed", result.colorsRemoved),
		)
	}
	return
}

// ============================================================================
// CSS processing
// ============================================================================

type cssFixResult struct {
	html            string
	changed         bool
	colorsRemoved   int
	contractAdded   bool
	skippedContrast bool
	worstRatio      float64
}

// textElementRe matches CSS selectors that target text elements (as descendants).
// Matches: ".foo h2", ".foo p", ".foo li", ".foo blockquote" etc.
// Does NOT match: ".foo" alone (container), ".foo a" (links keep explicit color).
var textElementRe = regexp.MustCompile(`\b(h[1-6]|p|li|blockquote|strong|em|cite|span|dt|dd)\b`)

// cssRuleRe splits CSS into selector + declarations blocks.
var cssRuleRe = regexp.MustCompile(`([^{}]+)\{([^{}]+)\}`)

// textColorDeclRe matches a standalone color: #hex declaration (not background-color, border-color).
// The leading character ensures we don't match inside a compound property name.
var textColorDeclRe = regexp.MustCompile(`(^|[{;\s])color:\s*(#[0-9a-fA-F]{3,8})\s*;?\s*`)

// styleBlockRe finds <style> blocks.
var styleBlockRe = regexp.MustCompile(`(?is)(<style[^>]*>)(.*?)(</style>)`)

// sectionContractRe checks if --section-text is defined in the CSS.
var sectionContractRe = regexp.MustCompile(`--section-text\s*:`)

func processComponentCSS(
	html, function string, isDarkSection bool,
	palette sitePalette, minContrast float64, logger *zap.Logger,
) cssFixResult {
	result := cssFixResult{html: html, worstRatio: 99.0}

	// Determine what the text and heading colors WILL BE after removing forced colors
	var expectedTextColor, expectedHeadingColor, bgColor string

	if isDarkSection {
		// Dark section: text gets --section-text (white-ish) or inherits body color
		// Headings get --section-heading (#fff) or --color-primary
		expectedTextColor = "#ffffff"    // --section-text resolves to ~white
		expectedHeadingColor = "#ffffff" // --section-heading resolves to white
		bgColor = palette.primary        // dark sections typically use primary as bg
	} else {
		// Light section: text inherits from body (--color-text)
		// Headings get --color-primary (fallback from --section-heading)
		expectedTextColor = palette.text
		expectedHeadingColor = palette.primary
		bgColor = palette.background
	}

	// Try to extract actual background color from the component's CSS
	if extracted := extractBgColor(html); extracted != "" {
		bgColor = extracted
	}

	// --- Contrast pre-check ---
	textRatio, textErr := wcagContrastRatio(expectedTextColor, bgColor)
	headingRatio, headingErr := wcagContrastRatio(expectedHeadingColor, bgColor)

	if textErr != nil || headingErr != nil {
		logger.Warn("processComponentCSS: contrast calc failed, skipping",
			zap.String("function", function),
			zap.Error(textErr),
		)
		result.skippedContrast = true
		result.worstRatio = 0
		return result
	}

	worst := math.Min(textRatio, headingRatio)
	result.worstRatio = math.Round(worst*100) / 100

	if worst < minContrast {
		logger.Warn("processComponentCSS: resulting contrast too low, skipping",
			zap.String("function", function),
			zap.Float64("text_ratio", textRatio),
			zap.Float64("heading_ratio", headingRatio),
			zap.Float64("min_required", minContrast),
			zap.String("bg", bgColor),
			zap.String("expected_text", expectedTextColor),
		)
		result.skippedContrast = true
		return result
	}

	// --- For dark sections missing --section-* contract, add it ---
	if isDarkSection {
		newHTML := ensureSectionContractInHTML(html, logger)
		if newHTML != html {
			html = newHTML
			result.contractAdded = true
		}
	}

	// --- Remove forced text colors from child element rules ---
	newHTML := removeChildTextColorsFromHTML(html)
	if newHTML != html {
		result.html = newHTML
		result.changed = true
		// Count how many declarations were removed
		result.colorsRemoved = countTextColorDeclarations(html) - countTextColorDeclarations(newHTML)
	}

	return result
}

// removeChildTextColorsFromHTML processes <style> blocks in HTML, removing
// color: #hex from rules that target text elements (not container rules).
func removeChildTextColorsFromHTML(html string) string {
	return styleBlockRe.ReplaceAllStringFunc(html, func(block string) string {
		matches := styleBlockRe.FindStringSubmatch(block)
		if len(matches) < 4 {
			return block
		}
		openTag := matches[1]
		cssContent := matches[2]
		closeTag := matches[3]

		newCSS := removeForcedTextColorsFromCSS(cssContent)
		return openTag + newCSS + closeTag
	})
}

// removeForcedTextColorsFromCSS processes CSS rules. For each rule whose
// selector targets a text element, removes color: #hex declarations.
// Leaves container-level color declarations alone.
func removeForcedTextColorsFromCSS(css string) string {
	return cssRuleRe.ReplaceAllStringFunc(css, func(rule string) string {
		parts := cssRuleRe.FindStringSubmatch(rule)
		if len(parts) < 3 {
			return rule
		}
		selector := parts[1]
		declarations := parts[2]

		// Only process rules where selector targets a text element
		// (has h1-h6, p, li, blockquote etc. as a descendant)
		if !textElementRe.MatchString(selector) {
			return rule // container rule or non-text — leave alone
		}

		// Skip rules targeting 'a' elements (links keep explicit color)
		selectorTrimmed := strings.TrimSpace(selector)
		if strings.HasSuffix(selectorTrimmed, " a") ||
			strings.HasSuffix(selectorTrimmed, " a:hover") ||
			strings.HasSuffix(selectorTrimmed, " a:visited") {
			return rule
		}

		// Remove color: #hex declarations from this rule
		newDeclarations := textColorDeclRe.ReplaceAllStringFunc(declarations, func(match string) string {
			// Preserve the leading character ({, ;, or whitespace)
			sub := textColorDeclRe.FindStringSubmatch(match)
			if len(sub) >= 2 {
				return sub[1] // just the prefix char, color decl removed
			}
			return ""
		})

		if newDeclarations == declarations {
			return rule // nothing changed
		}

		// Clean up double semicolons, leading semicolons
		newDeclarations = cleanupDeclarations(newDeclarations)

		return selector + "{" + newDeclarations + "}"
	})
}

// cleanupDeclarations removes artifacts from declaration removal:
// double semicolons, leading/trailing semicolons, excess whitespace.
func cleanupDeclarations(decl string) string {
	// Collapse multiple semicolons
	multiSemiRe := regexp.MustCompile(`;\s*;`)
	result := multiSemiRe.ReplaceAllString(decl, ";")
	// Remove leading semicolon after opening
	result = regexp.MustCompile(`\{\s*;`).ReplaceAllString(result, "{")
	// Clean up whitespace
	result = regexp.MustCompile(`\s{2,}`).ReplaceAllString(result, " ")
	return result
}

// countTextColorDeclarations counts color: #hex declarations on text element rules.
func countTextColorDeclarations(html string) int {
	count := 0
	for _, block := range styleBlockRe.FindAllStringSubmatch(html, -1) {
		if len(block) < 3 {
			continue
		}
		css := block[2]
		for _, rule := range cssRuleRe.FindAllStringSubmatch(css, -1) {
			if len(rule) < 3 {
				continue
			}
			selector := rule[1]
			declarations := rule[2]
			if textElementRe.MatchString(selector) {
				count += len(textColorDeclRe.FindAllString(declarations, -1))
			}
		}
	}
	return count
}

// ============================================================================
// Dark section contract injection
// ============================================================================

// sectionContractCSS is the --section-* variable block for dark sections.
const sectionContractCSS = `
    --section-text: rgba(255,255,255,0.9);
    --section-text-muted: rgba(255,255,255,0.7);
    --section-heading: #ffffff;
    --section-surface: rgba(255,255,255,0.05);
    --section-border: rgba(255,255,255,0.2);`

// ensureSectionContractInHTML checks if the HTML's <style> blocks contain
// --section-text. If not, finds the container rule (first class-only selector
// that sets background) and injects the contract variables.
func ensureSectionContractInHTML(html string, logger *zap.Logger) string {
	return styleBlockRe.ReplaceAllStringFunc(html, func(block string) string {
		matches := styleBlockRe.FindStringSubmatch(block)
		if len(matches) < 4 {
			return block
		}
		openTag := matches[1]
		cssContent := matches[2]
		closeTag := matches[3]

		// Already has the contract
		if sectionContractRe.MatchString(cssContent) {
			return block
		}

		newCSS := injectSectionContract(cssContent, logger)
		if newCSS == cssContent {
			return block
		}
		return openTag + newCSS + closeTag
	})
}

// containerBgRe finds CSS rules with a dark background (container-level).
// Matches rules like: .hero-section { background: #1a1a2e; ... }
var containerBgRe = regexp.MustCompile(
	`(\.[a-z][a-z0-9-]*-section\s*\{[^}]*)(background(?:-color)?:\s*(?:#[0-4][0-9a-fA-F]{5}|var\(--color-primary\)))([^}]*)(\})`,
)

func injectSectionContract(css string, logger *zap.Logger) string {
	// Find the container rule that sets a dark background
	if !containerBgRe.MatchString(css) {
		// No recognisable dark container rule — can't safely inject
		logger.Debug("injectSectionContract: no dark container rule found")
		return css
	}

	// Insert --section-* variables just before the closing } of the container rule
	return containerBgRe.ReplaceAllStringFunc(css, func(match string) string {
		// Only inject once (first match)
		closingIdx := strings.LastIndex(match, "}")
		if closingIdx < 0 {
			return match
		}
		return match[:closingIdx] + sectionContractCSS + "\n" + match[closingIdx:]
	})
}

// ============================================================================
// Background color extraction
// ============================================================================

// bgColorRe extracts hex color from background declarations.
var bgExtractRe = regexp.MustCompile(`background(?:-color)?:\s*#([0-9a-fA-F]{3,8})`)

func extractBgColor(html string) string {
	for _, block := range styleBlockRe.FindAllStringSubmatch(html, -1) {
		if len(block) < 3 {
			continue
		}
		css := block[2]
		// Look for container-level background (first rule with background)
		for _, rule := range cssRuleRe.FindAllStringSubmatch(css, -1) {
			if len(rule) < 3 {
				continue
			}
			selector := rule[1]
			declarations := rule[2]
			// Container rule: no descendant text element in selector
			if textElementRe.MatchString(selector) {
				continue
			}
			if match := bgExtractRe.FindStringSubmatch(declarations); len(match) >= 2 {
				return "#" + match[1]
			}
		}
	}
	return ""
}

// ============================================================================
// WCAG Contrast Ratio
// ============================================================================

// parseHexColor parses a hex color string (#rgb, #rrggbb) into 0-255 RGB values.
func parseHexColor(hex string) (r, g, b uint8, err error) {
	hex = strings.TrimPrefix(hex, "#")
	switch len(hex) {
	case 3:
		// Expand #rgb → #rrggbb
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	case 6:
		// already fine
	case 8:
		// #rrggbbaa — ignore alpha
		hex = hex[:6]
	default:
		return 0, 0, 0, fmt.Errorf("invalid hex color: #%s", hex)
	}

	rr, err := strconv.ParseUint(hex[0:2], 16, 8)
	if err != nil {
		return 0, 0, 0, err
	}
	gg, err := strconv.ParseUint(hex[2:4], 16, 8)
	if err != nil {
		return 0, 0, 0, err
	}
	bb, err := strconv.ParseUint(hex[4:6], 16, 8)
	if err != nil {
		return 0, 0, 0, err
	}

	return uint8(rr), uint8(gg), uint8(bb), nil
}

// sRGBToLinear converts an sRGB channel value (0-255) to linear light.
func sRGBToLinear(c uint8) float64 {
	s := float64(c) / 255.0
	if s <= 0.04045 {
		return s / 12.92
	}
	return math.Pow((s+0.055)/1.055, 2.4)
}

// relativeLuminance computes WCAG 2.x relative luminance from 0-255 RGB.
func relativeLuminance(r, g, b uint8) float64 {
	return 0.2126*sRGBToLinear(r) + 0.7152*sRGBToLinear(g) + 0.0722*sRGBToLinear(b)
}

// wcagContrastRatio computes the WCAG contrast ratio between two hex colors.
// Returns a value >= 1.0 (1:1 = identical, 21:1 = max contrast).
func wcagContrastRatio(hex1, hex2 string) (float64, error) {
	r1, g1, b1, err := parseHexColor(hex1)
	if err != nil {
		return 0, fmt.Errorf("parse fg color %q: %w", hex1, err)
	}
	r2, g2, b2, err := parseHexColor(hex2)
	if err != nil {
		return 0, fmt.Errorf("parse bg color %q: %w", hex2, err)
	}

	l1 := relativeLuminance(r1, g1, b1)
	l2 := relativeLuminance(r2, g2, b2)

	// Ensure l1 is the lighter value
	if l1 < l2 {
		l1, l2 = l2, l1
	}

	return (l1 + 0.05) / (l2 + 0.05), nil
}

// isDarkColor returns true if a hex color has relative luminance < 0.2
// (perceptually dark).
func isDarkColor(hex string) bool {
	r, g, b, err := parseHexColor(hex)
	if err != nil {
		return false
	}
	return relativeLuminance(r, g, b) < 0.2
}
