// FILE: platform/orchestration/actions/extract_design_fingerprint_action.go
//
// ExtractDesignFingerprintAction parses rawHTML from crawled pages to extract
// concrete design data: colour palette, font families, CSS variables, layout
// patterns, and dark/light section detection. Pure Go — no LLM, no cost,
// deterministic.
//
// Reads crawl pages using the same pattern as FormatCrawlForAnalysisAction
// (tries multiple paths in collected_data, falls back to DB).
//
// The output is written as a "design_reference" spec aspect by
// apply_adoption_plan, and passed to the webdesign-agent so it can reproduce
// the original site's visual identity.
//
// Step Zero:
//   - site-scraper: single-page LLM analysis, no CSS parsing. Different scope.
//   - format_crawl_for_analysis: markdown summaries for LLM. No CSS extraction.
//   - RenderCSSFromSpecAction: consumes design spec, doesn't extract one.
//   - No existing action parses <style> blocks from rawHTML.
//   Decision: New action needed.
//
// Registration (add to registry.go):
//   "extract_design_fingerprint": {
//       Handler:     ExtractDesignFingerprintAction,
//       Category:    "analysis",
//       Description: "Extract concrete design data (colours, fonts, layout) from crawled HTML",
//       IsLocal:     true,
//   },

package actions

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var ExtractDesignFingerprintInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{},
	Optional:   []string{"crawl_field"},
	Defaults:   map[string]interface{}{"crawl_field": "crawl_result"},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("extract_design_fingerprint", ExtractDesignFingerprintInputSpec)
}

// ── Regex patterns ──────────────────────────────────────────────────────

var (
	fpHexColorRe    = regexp.MustCompile(`#([0-9a-fA-F]{3,8})\b`)
	fpRgbColorRe    = regexp.MustCompile(`rgba?\(\s*(\d{1,3})\s*,\s*(\d{1,3})\s*,\s*(\d{1,3})(?:\s*,\s*[\d.]+)?\s*\)`)
	fpFontFamilyRe  = regexp.MustCompile(`font-family\s*:\s*([^;}{]+)`)
	fpGoogleFontsRe = regexp.MustCompile(`fonts\.googleapis\.com/css2?\?family=([^"&]+)`)
	fpMaxWidthRe    = regexp.MustCompile(`max-width\s*:\s*(\d+(?:\.\d+)?(?:px|rem|em|%))`)
	fpCSSVarRe      = regexp.MustCompile(`(--[\w-]+)\s*:\s*([^;}{]+)`)
	fpBgColorRe     = regexp.MustCompile(`background(?:-color)?\s*:\s*([^;}{]+)`)
	fpTextColorRe   = regexp.MustCompile(`(?:^|[;{}\s])color\s*:\s*([^;}{]+)`)
	fpDisplayRe     = regexp.MustCompile(`display\s*:\s*(grid|flex|inline-grid|inline-flex)`)
	fpGapRe         = regexp.MustCompile(`(?:gap|row-gap|column-gap)\s*:\s*([^;}{]+)`)
)

// ── Main action ─────────────────────────────────────────────────────────

func ExtractDesignFingerprintAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "extract_design_fingerprint"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	crawlField := "crawl_result"
	if cf, ok := config["crawl_field"].(string); ok && cf != "" {
		crawlField = cf
	}

	// ── Find pages (same pattern as format_crawl_for_analysis) ──────
	var pages []interface{}
	paths := []string{
		crawlField + ".pages",
		crawlField + ".body.data.pages",
		crawlField + ".data.pages",
		crawlField + ".body.pages",
	}

	for _, path := range paths {
		raw := datahelpers.ExtractNestedField(params.CollectedData, path)
		if raw != nil {
			if arr, ok := raw.([]interface{}); ok && len(arr) > 0 {
				pages = arr
				logger.Info("Found crawl pages for fingerprinting",
					zap.String("path", path),
					zap.Int("count", len(arr)))
				break
			}
		}
	}

	if len(pages) == 0 {
		logger.Warn("No pages found for design fingerprint extraction")
		return map[string]interface{}{
			"status":  "no_pages",
			"message": "No crawl pages with rawHtml available for fingerprint extraction",
		}, nil
	}

	// ── Aggregate across all pages ──────────────────────────────────

	bgColors := make(map[string]int)
	txtColors := make(map[string]int)
	allColors := make(map[string]int)
	fonts := make(map[string]int)
	fontSources := make(map[string]string) // font → "google_fonts" or "css"
	cssVars := make(map[string]string)
	maxWidths := make(map[string]int)
	displayPatterns := make(map[string]int)
	gapValues := make(map[string]int)
	googleFontsURLs := []string{}

	var styleBlockCount, inlineStyleCount int
	var darkBgColors []string
	hasDark := false
	hasLight := true // assume light unless only dark found
	pagesAnalyzed := 0

	for _, pageRaw := range pages {
		page, ok := pageRaw.(map[string]interface{})
		if !ok {
			continue
		}

		rawHTML, _ := page["rawHtml"].(string)
		if rawHTML == "" {
			continue
		}
		pagesAnalyzed++

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(rawHTML))
		if err != nil {
			logger.Warn("Failed to parse page HTML", zap.Error(err))
			continue
		}

		// ── <style> blocks ──────────────────────────────────────
		doc.Find("style").Each(func(i int, s *goquery.Selection) {
			styleBlockCount++
			css := s.Text()
			fpExtractColors(css, bgColors, txtColors, allColors)
			fpExtractFonts(css, fonts, fontSources)
			fpExtractCSSVars(css, cssVars)
			fpExtractLayout(css, maxWidths, displayPatterns, gapValues)
		})

		// ── <link> tags (Google Fonts) ──────────────────────────
		doc.Find("link").Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if !exists {
				return
			}
			if matches := fpGoogleFontsRe.FindStringSubmatch(href); len(matches) > 1 {
				googleFontsURLs = append(googleFontsURLs, href)
				fpParseFontFamiliesFromGoogleURL(matches[1], fonts, fontSources)
			}
		})

		// ── Inline styles ───────────────────────────────────────
		doc.Find("[style]").Each(func(i int, s *goquery.Selection) {
			style, _ := s.Attr("style")
			if style == "" {
				return
			}
			inlineStyleCount++
			fpExtractColors(style, bgColors, txtColors, allColors)
		})

		// ── Dark section detection ──────────────────────────────
		doc.Find("section, header, footer, div, nav").Each(func(i int, s *goquery.Selection) {
			style, _ := s.Attr("style")
			classes, _ := s.Attr("class")

			if bgMatches := fpBgColorRe.FindStringSubmatch(style); len(bgMatches) > 1 {
				bgVal := strings.TrimSpace(bgMatches[1])
				if fpIsDarkColor(bgVal) {
					hasDark = true
					norm := fpNormalizeColor(bgVal)
					if norm != "" {
						darkBgColors = append(darkBgColors, norm)
					}
				}
			}

			classLower := strings.ToLower(classes)
			if strings.Contains(classLower, "dark") || strings.Contains(classLower, "bg-dark") ||
				strings.Contains(classLower, "bg-black") {
				hasDark = true
			}
		})
	}

	// ── Build result ────────────────────────────────────────────────

	result := map[string]interface{}{
		"status": "extracted",
		"meta": map[string]interface{}{
			"pages_analyzed":      pagesAnalyzed,
			"style_blocks_found":  styleBlockCount,
			"inline_styles_found": inlineStyleCount,
		},
		"colors": map[string]interface{}{
			"background": fpTopEntries(bgColors, 10),
			"text":       fpTopEntries(txtColors, 10),
			"all":        fpTopEntries(allColors, 15),
		},
		"typography": map[string]interface{}{
			"fonts":             fpSortedFontEntries(fonts, fontSources),
			"google_fonts_urls": fpDedup(googleFontsURLs),
		},
		"layout": map[string]interface{}{
			"max_widths":       fpTopKeys(maxWidths, 5),
			"display_patterns": displayPatterns,
			"gap_values":       fpTopKeys(gapValues, 5),
		},
		"css_variables": cssVars,
		"dark_sections": map[string]interface{}{
			"has_dark_sections":  hasDark,
			"has_light_sections": hasLight,
			"dark_bg_colors":     fpDedup(darkBgColors),
			"predominant_scheme": fpPredominantScheme(hasDark, hasLight),
		},
	}

	// ── Derive suggested values ─────────────────────────────────────
	// These are the "best guess" mappings to our CSS variable convention.
	// The webdesign-agent can use these as starting points.

	suggested := make(map[string]string)

	// Check CSS variables first — they're the most authoritative source
	for varName, varValue := range cssVars {
		lower := strings.ToLower(varName)
		switch {
		case strings.Contains(lower, "primary") && suggested["primary"] == "":
			suggested["primary"] = strings.TrimSpace(varValue)
		case strings.Contains(lower, "accent") && suggested["accent"] == "":
			suggested["accent"] = strings.TrimSpace(varValue)
		case strings.Contains(lower, "secondary") && suggested["secondary"] == "":
			suggested["secondary"] = strings.TrimSpace(varValue)
		case (strings.Contains(lower, "bg") || strings.Contains(lower, "background")) && suggested["background"] == "":
			suggested["background"] = strings.TrimSpace(varValue)
		case (strings.Contains(lower, "surface") || strings.Contains(lower, "card")) && suggested["surface"] == "":
			suggested["surface"] = strings.TrimSpace(varValue)
		case strings.Contains(lower, "text") && !strings.Contains(lower, "muted") && suggested["text"] == "":
			suggested["text"] = strings.TrimSpace(varValue)
		case (strings.Contains(lower, "muted") || strings.Contains(lower, "dim")) && suggested["text_muted"] == "":
			suggested["text_muted"] = strings.TrimSpace(varValue)
		case strings.Contains(lower, "border") && suggested["border"] == "":
			suggested["border"] = strings.TrimSpace(varValue)
		case strings.Contains(lower, "font-family") || strings.Contains(lower, "font_family"):
			suggested["font_family"] = strings.TrimSpace(varValue)
		case strings.Contains(lower, "heading") && strings.Contains(lower, "font"):
			suggested["heading_font"] = strings.TrimSpace(varValue)
		}
	}

	// Fall back to frequency-based suggestions if CSS vars didn't provide them
	if suggested["primary"] == "" && len(bgColors) > 0 {
		suggested["primary"] = fpTopKey(bgColors)
	}
	if suggested["text"] == "" && len(txtColors) > 0 {
		suggested["text"] = fpTopKey(txtColors)
	}

	// Font suggestions from extracted fonts
	if suggested["font_family"] == "" {
		fontList := fpSortedFontEntries(fonts, fontSources)
		if arr, ok := fontList.([]map[string]interface{}); ok && len(arr) > 0 {
			if family, ok := arr[0]["family"].(string); ok {
				suggested["font_family"] = family
			}
			if len(arr) > 1 {
				if family, ok := arr[1]["family"].(string); ok {
					suggested["heading_font"] = family
				}
			}
		}
	}

	result["suggested_mapping"] = suggested

	logger.Info("Design fingerprint extracted",
		zap.Int("pages_analyzed", pagesAnalyzed),
		zap.Int("colors_found", len(allColors)),
		zap.Int("fonts_found", len(fonts)),
		zap.Int("css_vars_found", len(cssVars)),
		zap.Bool("has_dark_sections", hasDark),
		zap.Int("suggested_fields", len(suggested)),
	)

	return result, nil
}

// ── CSS parsing helpers ─────────────────────────────────────────────────

func fpExtractColors(css string, bgColors, txtColors, allColors map[string]int) {
	for _, match := range fpBgColorRe.FindAllStringSubmatch(css, -1) {
		val := strings.TrimSpace(match[1])
		for _, hex := range fpHexColorRe.FindAllString(val, -1) {
			if norm := fpNormalizeHex(hex); norm != "" && !fpIsBoringColor(norm) {
				bgColors[norm]++
				allColors[norm]++
			}
		}
	}

	for _, match := range fpTextColorRe.FindAllStringSubmatch(css, -1) {
		val := strings.TrimSpace(match[1])
		for _, hex := range fpHexColorRe.FindAllString(val, -1) {
			if norm := fpNormalizeHex(hex); norm != "" && !fpIsBoringColor(norm) {
				txtColors[norm]++
				allColors[norm]++
			}
		}
	}

	for _, hex := range fpHexColorRe.FindAllString(css, -1) {
		if norm := fpNormalizeHex(hex); norm != "" && !fpIsBoringColor(norm) {
			allColors[norm]++
		}
	}
}

func fpExtractFonts(css string, fonts map[string]int, sources map[string]string) {
	for _, match := range fpFontFamilyRe.FindAllStringSubmatch(css, -1) {
		for _, f := range strings.Split(match[1], ",") {
			f = strings.TrimSpace(strings.Trim(strings.TrimSpace(f), "'\""))
			if f == "" || fpIsGenericFont(f) {
				continue
			}
			fonts[f]++
			if sources[f] == "" {
				sources[f] = "css"
			}
		}
	}
}

func fpExtractCSSVars(css string, vars map[string]string) {
	for _, match := range fpCSSVarRe.FindAllStringSubmatch(css, -1) {
		name := strings.TrimSpace(match[1])
		value := strings.TrimSpace(match[2])
		if fpIsDesignVar(name) {
			vars[name] = value
		}
	}
}

func fpExtractLayout(css string, maxWidths, displayPatterns, gapValues map[string]int) {
	for _, match := range fpMaxWidthRe.FindAllStringSubmatch(css, -1) {
		maxWidths[match[1]]++
	}
	for _, match := range fpDisplayRe.FindAllStringSubmatch(css, -1) {
		displayPatterns[match[1]]++
	}
	for _, match := range fpGapRe.FindAllStringSubmatch(css, -1) {
		gapValues[strings.TrimSpace(match[1])]++
	}
}

func fpParseFontFamiliesFromGoogleURL(familyParam string, fonts map[string]int, sources map[string]string) {
	for _, fam := range strings.Split(familyParam, "&family=") {
		if colonIdx := strings.Index(fam, ":"); colonIdx > 0 {
			fam = fam[:colonIdx]
		}
		fam = strings.TrimSpace(strings.ReplaceAll(fam, "+", " "))
		if fam != "" {
			fonts[fam]++
			sources[fam] = "google_fonts" // upgrade source
		}
	}
}

// ── Colour helpers ──────────────────────────────────────────────────────

func fpNormalizeHex(hex string) string {
	hex = strings.ToLower(strings.TrimSpace(hex))
	if !strings.HasPrefix(hex, "#") {
		return ""
	}
	hex = hex[1:]
	if len(hex) == 3 {
		hex = string(hex[0]) + string(hex[0]) + string(hex[1]) + string(hex[1]) + string(hex[2]) + string(hex[2])
	}
	if len(hex) == 8 {
		hex = hex[:6]
	}
	if len(hex) != 6 {
		return ""
	}
	return "#" + hex
}

func fpNormalizeColor(val string) string {
	val = strings.TrimSpace(val)
	if hexMatch := fpHexColorRe.FindString(val); hexMatch != "" {
		return fpNormalizeHex(hexMatch)
	}
	return ""
}

func fpIsBoringColor(hex string) bool {
	boring := map[string]bool{
		"#000000": true, "#ffffff": true,
		"#333333": true, "#666666": true, "#999999": true, "#cccccc": true,
		"#f5f5f5": true, "#eeeeee": true, "#dddddd": true, "#e5e5e5": true,
		"#fafafa": true, "#f0f0f0": true, "#f8f8f8": true,
	}
	return boring[strings.ToLower(hex)]
}

func fpIsDarkColor(val string) bool {
	hex := fpNormalizeColor(val)
	if hex == "" || len(hex) != 7 {
		return false
	}
	var r, g, b int
	fmt.Sscanf(hex[1:], "%02x%02x%02x", &r, &g, &b)
	luminance := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
	return luminance < 80
}

func fpIsGenericFont(f string) bool {
	generic := map[string]bool{
		"serif": true, "sans-serif": true, "monospace": true,
		"cursive": true, "fantasy": true, "system-ui": true,
		"ui-serif": true, "ui-sans-serif": true, "ui-monospace": true,
		"inherit": true, "initial": true, "unset": true,
		"-apple-system": true, "blinkmacfontsystem": true,
		"segoe ui": true, "helvetica neue": true, "arial": true,
		"helvetica": true,
	}
	return generic[strings.ToLower(f)]
}

func fpIsDesignVar(name string) bool {
	prefixes := []string{
		"--color", "--colour", "--bg", "--text", "--font", "--heading",
		"--spacing", "--gap", "--radius", "--border", "--shadow",
		"--primary", "--secondary", "--accent", "--surface", "--card",
		"--max-width", "--container", "--transition",
	}
	lower := strings.ToLower(name)
	for _, p := range prefixes {
		if strings.HasPrefix(lower, p) {
			return true
		}
	}
	return false
}

func fpPredominantScheme(hasDark, hasLight bool) string {
	if hasDark && !hasLight {
		return "dark"
	}
	if hasDark && hasLight {
		return "mixed"
	}
	return "light"
}

// ── Output formatting helpers ───────────────────────────────────────────

func fpTopEntries(m map[string]int, limit int) []map[string]interface{} {
	type kv struct {
		Key   string
		Count int
	}
	kvs := make([]kv, 0, len(m))
	for k, v := range m {
		kvs = append(kvs, kv{k, v})
	}
	sort.Slice(kvs, func(i, j int) bool { return kvs[i].Count > kvs[j].Count })

	result := make([]map[string]interface{}, 0, limit)
	for i, kv := range kvs {
		if i >= limit {
			break
		}
		result = append(result, map[string]interface{}{
			"hex":   kv.Key,
			"count": kv.Count,
		})
	}
	return result
}

func fpTopKeys(m map[string]int, limit int) []string {
	type kv struct {
		Key   string
		Count int
	}
	kvs := make([]kv, 0, len(m))
	for k, v := range m {
		kvs = append(kvs, kv{k, v})
	}
	sort.Slice(kvs, func(i, j int) bool { return kvs[i].Count > kvs[j].Count })
	result := make([]string, 0, limit)
	for i, kv := range kvs {
		if i >= limit {
			break
		}
		result = append(result, kv.Key)
	}
	return result
}

func fpTopKey(m map[string]int) string {
	best, bestCount := "", 0
	for k, v := range m {
		if v > bestCount {
			best, bestCount = k, v
		}
	}
	return best
}

func fpSortedFontEntries(fonts map[string]int, sources map[string]string) interface{} {
	type fe struct {
		Family string
		Count  int
		Source string
	}
	entries := make([]fe, 0, len(fonts))
	for f, c := range fonts {
		src := sources[f]
		if src == "" {
			src = "css"
		}
		entries = append(entries, fe{f, c, src})
	}
	// Google Fonts first, then by count
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Source == "google_fonts" && entries[j].Source != "google_fonts" {
			return true
		}
		if entries[i].Source != "google_fonts" && entries[j].Source == "google_fonts" {
			return false
		}
		return entries[i].Count > entries[j].Count
	})

	result := make([]map[string]interface{}, len(entries))
	for i, e := range entries {
		result[i] = map[string]interface{}{
			"family": e.Family,
			"count":  e.Count,
			"source": e.Source,
		}
	}
	return result
}

func fpDedup(ss []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(ss))
	for _, s := range ss {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}
