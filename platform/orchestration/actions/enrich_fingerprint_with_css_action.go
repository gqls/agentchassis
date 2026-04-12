// FILE: platform/orchestration/actions/enrich_fingerprint_with_css_action.go
//
// EnrichFingerprintWithCSSAction parses externally-fetched CSS file content
// and merges the extracted design data into the existing fingerprint.
//
// This runs after firecrawl_scrape fetches the CSS file(s) through the
// webscrape adapter. It reuses the fp* parsing functions from
// extract_design_fingerprint_action.go (same package).
//
// Step Zero:
//   - extract_design_fingerprint: extracts from inline styles only. Different scope.
//   - format_crawl_for_analysis: markdown summaries, no CSS. Different purpose.
//   - No existing action merges external CSS data into a fingerprint.
//   Decision: New action needed.
//
// Registration (add to registry.go):
//   "enrich_fingerprint_with_css": {
//       Handler:     EnrichFingerprintWithCSSAction,
//       Category:    "analysis",
//       Description: "Parse fetched CSS and merge into design fingerprint",
//       IsLocal:     true,
//   },

package actions

import (
	"context"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var EnrichFingerprintWithCSSInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{},
	Optional:   []string{"css_scrape_field", "fingerprint_field"},
	Defaults:   map[string]interface{}{"css_scrape_field": "css_scrape_result", "fingerprint_field": "design_fingerprint"},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("enrich_fingerprint_with_css", EnrichFingerprintWithCSSInputSpec)
}

func EnrichFingerprintWithCSSAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "enrich_fingerprint_with_css"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	cssField := "css_scrape_result"
	if cf, ok := config["css_scrape_field"].(string); ok && cf != "" {
		cssField = cf
	}

	fpField := "design_fingerprint"
	if ff, ok := config["fingerprint_field"].(string); ok && ff != "" {
		fpField = ff
	}

	// ── Load existing fingerprint ───────────────────────────────────
	fpRaw := datahelpers.ExtractNestedField(params.CollectedData, fpField)
	if fpRaw == nil {
		logger.Warn("No existing fingerprint found, nothing to enrich")
		return map[string]interface{}{"status": "no_fingerprint"}, nil
	}

	fpUnwrapped := datahelpers.UnwrapDeep(fpRaw, logger)
	fp, ok := fpUnwrapped.(map[string]interface{})
	if !ok {
		logger.Warn("Fingerprint is not a map")
		return map[string]interface{}{"status": "invalid_fingerprint"}, nil
	}

	// ── Extract CSS content from scrape result ──────────────────────
	// firecrawl_scrape returns various formats. Try common paths.
	var cssContent string

	cssPaths := []string{
		cssField + ".rawHtml",
		cssField + ".body.data.rawHtml",
		cssField + ".data.rawHtml",
		cssField + ".body.rawHtml",
		cssField + ".markdown",
		cssField + ".body.data.markdown",
		cssField + ".content",
		cssField + ".body.content",
	}

	for _, path := range cssPaths {
		raw := datahelpers.ExtractNestedField(params.CollectedData, path)
		if raw != nil {
			if s, ok := raw.(string); ok && len(s) > 20 {
				cssContent = s
				logger.Info("Found CSS content",
					zap.String("path", path),
					zap.Int("length", len(s)))
				break
			}
		}
	}

	if cssContent == "" {
		logger.Info("No CSS content found in scrape result, fingerprint unchanged")
		// Return existing fingerprint unmodified
		fp["css_enrichment"] = "no_css_content_found"
		return fp, nil
	}

	// ── Clean up CSS content ────────────────────────────────────────
	// firecrawl might wrap CSS in HTML tags if it treated it as a page
	if strings.Contains(cssContent, "<html") {
		// Try to extract content between <body> tags or <pre> tags
		if bodyStart := strings.Index(cssContent, "<body"); bodyStart >= 0 {
			if bodyEnd := strings.Index(cssContent[bodyStart:], ">"); bodyEnd >= 0 {
				cssContent = cssContent[bodyStart+bodyEnd+1:]
			}
		}
		if bodyClose := strings.Index(cssContent, "</body>"); bodyClose >= 0 {
			cssContent = cssContent[:bodyClose]
		}
		// Strip remaining HTML tags
		cssContent = stripHTMLTags(cssContent)
	}

	// ── Parse CSS using existing fp* functions ──────────────────────
	bgColors := make(map[string]int)
	txtColors := make(map[string]int)
	allColors := make(map[string]int)
	fonts := make(map[string]int)
	fontSources := make(map[string]string)
	cssVars := make(map[string]string)
	maxWidths := make(map[string]int)
	displayPatterns := make(map[string]int)
	gapValues := make(map[string]int)

	fpExtractColors(cssContent, bgColors, txtColors, allColors)
	fpExtractFonts(cssContent, fonts, fontSources)
	fpExtractCSSVars(cssContent, cssVars)
	fpExtractLayout(cssContent, maxWidths, displayPatterns, gapValues)

	logger.Info("Parsed external CSS",
		zap.Int("colors", len(allColors)),
		zap.Int("fonts", len(fonts)),
		zap.Int("css_vars", len(cssVars)),
		zap.Int("max_widths", len(maxWidths)),
	)

	// ── Merge into existing fingerprint ─────────────────────────────
	// External CSS takes priority for variables and fonts (it's the
	// authoritative source). Colours are additive.

	// Merge CSS variables (most valuable — these are design tokens)
	existingVars, _ := fp["css_variables"].(map[string]interface{})
	if existingVars == nil {
		existingVars = make(map[string]interface{})
	}
	for k, v := range cssVars {
		existingVars[k] = v
	}
	fp["css_variables"] = existingVars

	// Merge colours (additive — add to existing lists)
	mergeColorEntries(fp, "colors", "all", allColors)
	mergeColorEntries(fp, "colors", "background", bgColors)
	mergeColorEntries(fp, "colors", "text", txtColors)

	// Merge fonts
	if len(fonts) > 0 {
		typoMap := map[string]interface{}{
			"fonts": fpSortedFontEntries(fonts, fontSources),
		}
		// Preserve existing google_fonts_urls if present
		if existingTypo, ok := fp["typography"].(map[string]interface{}); ok {
			if gfu, ok := existingTypo["google_fonts_urls"]; ok {
				typoMap["google_fonts_urls"] = gfu
			}
		}
		fp["typography"] = typoMap
	}

	// Merge layout
	if len(maxWidths) > 0 || len(displayPatterns) > 0 {
		fp["layout"] = map[string]interface{}{
			"max_widths":       fpTopKeys(maxWidths, 5),
			"display_patterns": displayPatterns,
			"gap_values":       fpTopKeys(gapValues, 5),
		}
	}

	// ── Re-derive suggested mapping with enriched data ──────────────
	suggested := make(map[string]string)

	// CSS variables first (most authoritative)
	for varName, varValue := range cssVars {
		lower := strings.ToLower(varName)
		val := strings.TrimSpace(varValue)
		switch {
		case strings.Contains(lower, "primary") && suggested["primary"] == "":
			suggested["primary"] = val
		case strings.Contains(lower, "accent") && suggested["accent"] == "":
			suggested["accent"] = val
		case strings.Contains(lower, "secondary") && suggested["secondary"] == "":
			suggested["secondary"] = val
		case (strings.Contains(lower, "bg") || strings.Contains(lower, "background")) && suggested["background"] == "":
			suggested["background"] = val
		case (strings.Contains(lower, "surface") || strings.Contains(lower, "card")) && suggested["surface"] == "":
			suggested["surface"] = val
		case strings.Contains(lower, "text") && !strings.Contains(lower, "muted") && suggested["text"] == "":
			suggested["text"] = val
		case (strings.Contains(lower, "muted") || strings.Contains(lower, "dim")) && suggested["text_muted"] == "":
			suggested["text_muted"] = val
		case strings.Contains(lower, "border") && suggested["border"] == "":
			suggested["border"] = val
		case strings.Contains(lower, "font-family") || strings.Contains(lower, "font_family"):
			suggested["font_family"] = val
		case strings.Contains(lower, "heading") && strings.Contains(lower, "font"):
			suggested["heading_font"] = val
		}
	}

	// Fall back to frequency for anything not found in vars
	if suggested["primary"] == "" && len(bgColors) > 0 {
		suggested["primary"] = fpTopKey(bgColors)
	}
	if suggested["text"] == "" && len(txtColors) > 0 {
		suggested["text"] = fpTopKey(txtColors)
	}
	if suggested["font_family"] == "" && len(fonts) > 0 {
		entries := fpSortedFontEntries(fonts, fontSources)
		if arr, ok := entries.([]map[string]interface{}); ok && len(arr) > 0 {
			if family, ok := arr[0]["family"].(string); ok {
				suggested["font_family"] = family
			}
		}
	}

	fp["suggested_mapping"] = suggested
	fp["css_enrichment"] = "enriched"
	fp["css_enrichment_stats"] = map[string]interface{}{
		"colors_added":   len(allColors),
		"fonts_added":    len(fonts),
		"css_vars_added": len(cssVars),
	}

	logger.Info("Fingerprint enriched with external CSS",
		zap.Int("total_css_vars", len(existingVars)),
		zap.Int("total_suggested", len(suggested)),
		zap.Int("fonts_from_css", len(fonts)),
	)

	return fp, nil
}

// mergeColorEntries adds new colour entries to the fingerprint's nested colour lists.
func mergeColorEntries(fp map[string]interface{}, section, subsection string, newColors map[string]int) {
	if len(newColors) == 0 {
		return
	}

	colors, _ := fp[section].(map[string]interface{})
	if colors == nil {
		colors = make(map[string]interface{})
		fp[section] = colors
	}

	existing, _ := colors[subsection].([]interface{})
	existingMap := make(map[string]int)
	for _, e := range existing {
		if entry, ok := e.(map[string]interface{}); ok {
			if hex, ok := entry["hex"].(string); ok {
				if count, ok := entry["count"].(float64); ok {
					existingMap[hex] = int(count)
				}
			}
		}
	}

	// Merge counts
	for hex, count := range newColors {
		existingMap[hex] += count
	}

	colors[subsection] = fpTopEntries(existingMap, 15)
}

// stripHTMLTags removes HTML tags from content (for CSS wrapped in HTML by firecrawl).
func stripHTMLTags(s string) string {
	var result strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(r)
		}
	}
	return result.String()
}
