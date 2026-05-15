// FILE: platform/orchestration/actions/extract_interactive_fingerprint_action.go
//
// ExtractInteractiveFingerprintAction parses rawHTML from crawled pages to
// extract concrete signals about interactive elements: <canvas> presence,
// inline <script> content, external <script src="..."> URLs, inline event
// handlers, form/input counts, and library/framework signatures (canvas API,
// requestAnimationFrame, fetch, jQuery, Three.js, Phaser, p5, React, Vue).
// Pure Go — no LLM, no cost, deterministic.
//
// Parallels extract_design_fingerprint_action.go (same package, same pattern).
// Reads crawl pages the same way; produces a structured interactive
// fingerprint instead of a design fingerprint. External JS URLs are returned
// in the result for the workflow to fetch via firecrawl_scrape (handled by
// the enrich_fingerprint_with_js action, parallel to enrich_fingerprint_with_css).
//
// The output will be written as an "interactive_reference" spec aspect by
// apply_adoption_plan (once the workflow integration in step C2/C6 is done)
// and passed to the LLM brief generator that produces "interactive_intent".
//
// Step Zero (2026-05-15):
//   - extract_design_fingerprint: extracts design data only. Different scope.
//   - check_tool_health.scriptTagRe: regex check on our deployed tool output,
//     not on crawled source. Different direction.
//   - No existing action parses <script> or <canvas> from rawHTML.
//   Decision: new action needed.
//
// Registration (add to registry.go):
//   "extract_interactive_fingerprint": {
//       Handler:     ExtractInteractiveFingerprintAction,
//       Category:    "analysis",
//       Description: "Extract interactive element signals (scripts, canvas, forms) from crawled HTML",
//       IsLocal:     true,
//   },

package actions

import (
	"context"
	"regexp"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ── Constants ──────────────────────────────────────────────────────────

// Inline script truncation threshold. Real tool/calculator/game source code
// is typically well under this; analytics bundles and minified frameworks
// blow past it. Truncation is logged so we know if it's biting in practice.
// Revisit in C4 if downstream LLM step finds itself missing context.
const interactiveScriptTruncateBytes = 12 * 1024

// ── Input spec ─────────────────────────────────────────────────────────

var ExtractInteractiveFingerprintInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{},
	Optional:   []string{"crawl_field"},
	Defaults:   map[string]interface{}{"crawl_field": "crawl_result"},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("extract_interactive_fingerprint", ExtractInteractiveFingerprintInputSpec)
}

// ── Regex patterns ─────────────────────────────────────────────────────

var (
	// Canvas API and animation
	ifpCanvas2DRe  = regexp.MustCompile(`(?i)getContext\s*\(\s*['"]2d['"]`)
	ifpCanvasGLRe  = regexp.MustCompile(`(?i)getContext\s*\(\s*['"]webgl`)
	ifpRAFRe       = regexp.MustCompile(`\brequestAnimationFrame\b`)
	ifpAddEventRe  = regexp.MustCompile(`\.addEventListener\s*\(`)
	ifpFetchRe     = regexp.MustCompile(`\bfetch\s*\(`)
	ifpWebSocketRe = regexp.MustCompile(`\bnew\s+WebSocket\b`)
	// Libraries / frameworks
	ifpJqueryRe = regexp.MustCompile(`\bjQuery\b|\$\(\s*document|\$\.ajax`)
	ifpThreeRe  = regexp.MustCompile(`\bTHREE\.`)
	ifpPhaserRe = regexp.MustCompile(`\bPhaser\.`)
	// p5: characteristic setup()+draw() signature
	ifpP5SetupRe = regexp.MustCompile(`function\s+setup\s*\(`)
	ifpP5DrawRe  = regexp.MustCompile(`function\s+draw\s*\(`)
	ifpReactRe   = regexp.MustCompile(`\bReact\.(?:createElement|Component|useState|useEffect)\b|from\s+['"]react['"]`)
	ifpVueRe     = regexp.MustCompile(`\bnew\s+Vue\s*\(|Vue\.createApp`)
)

// Inline event handler attributes we look for.
var ifpEventAttrs = []string{
	"onclick", "onsubmit", "oninput", "onchange",
	"onload", "onmouseover", "onkeydown", "onkeyup",
}

// ── Main action ────────────────────────────────────────────────────────

func ExtractInteractiveFingerprintAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "extract_interactive_fingerprint"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	// Resolve inputs via the canonical extractor.
	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		ExtractInteractiveFingerprintInputSpec,
		logger,
	)
	if err != nil {
		return nil, err
	}

	crawlField := inputs.Get("crawl_field")
	if crawlField == "" {
		crawlField = "crawl_result"
	}

	// ── Find pages (same pattern as extract_design_fingerprint) ─────
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
				logger.Info("Found crawl pages for interactive fingerprinting",
					zap.String("path", path),
					zap.Int("count", len(arr)))
				break
			}
		}
	}

	if len(pages) == 0 {
		logger.Warn("No pages found for interactive fingerprint extraction")
		return map[string]interface{}{
			"status":  "no_pages",
			"message": "No crawl pages with rawHtml available for interactive fingerprint extraction",
		}, nil
	}

	// ── Aggregate state ────────────────────────────────────────────
	var (
		perPage               []map[string]interface{}
		externalJSURLs        = make(map[string]bool)
		librarySignals        = make(map[string]int)
		pagesAnalyzed         int
		inlineScriptBlocks    int
		externalScriptCount   int
		canvasPagesCount      int
		interactivePagesCount int
		truncatedScripts      int
	)

	// ── Per-page analysis ──────────────────────────────────────────
	for pageIdx, pageRaw := range pages {
		page, ok := pageRaw.(map[string]interface{})
		if !ok {
			continue
		}

		rawHTML, _ := page["rawHtml"].(string)
		if rawHTML == "" {
			continue
		}
		pagesAnalyzed++

		pageURL := ""
		if metadata, ok := page["metadata"].(map[string]interface{}); ok {
			pageURL, _ = metadata["url"].(string)
		}

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(rawHTML))
		if err != nil {
			logger.Warn("Failed to parse page HTML for interactive fingerprint",
				zap.Int("page_index", pageIdx),
				zap.String("url", pageURL),
				zap.Error(err))
			continue
		}

		var (
			canvasCount       int
			canvasDetails     []map[string]interface{}
			inlineScripts     []string // truncated samples
			inlineScriptCount int
			externalScripts   []string
			eventHandlerCount int
			formCount         int
			inputCount        int
			pageSignals       = make(map[string]bool)
		)

		// ── <canvas> elements ──────────────────────────────────
		doc.Find("canvas").Each(func(i int, s *goquery.Selection) {
			canvasCount++
			id, _ := s.Attr("id")
			width, _ := s.Attr("width")
			height, _ := s.Attr("height")
			canvasDetails = append(canvasDetails, map[string]interface{}{
				"id":     id,
				"width":  width,
				"height": height,
			})
		})

		// ── <script> blocks: inline content and external src ──
		doc.Find("script").Each(func(i int, s *goquery.Selection) {
			if src, exists := s.Attr("src"); exists && src != "" {
				resolved := fpResolveURL(src, pageURL)
				if resolved != "" {
					externalScripts = append(externalScripts, resolved)
					externalJSURLs[resolved] = true
					externalScriptCount++
				}
				return
			}

			// Inline script
			text := s.Text()
			if strings.TrimSpace(text) == "" {
				return
			}
			inlineScriptCount++
			inlineScriptBlocks++

			// Library signal detection on the full text
			ifpDetectLibrarySignals(text, librarySignals, pageSignals)

			// Capture truncated sample for downstream LLM
			if len(text) > interactiveScriptTruncateBytes {
				truncatedScripts++
				logger.Info("Inline script truncated for fingerprint",
					zap.Int("page_index", pageIdx),
					zap.String("url", pageURL),
					zap.Int("original_bytes", len(text)),
					zap.Int("kept_bytes", interactiveScriptTruncateBytes))
				text = text[:interactiveScriptTruncateBytes]
			}
			inlineScripts = append(inlineScripts, text)
		})

		// ── Inline event handler attributes ─────────────────────
		for _, attr := range ifpEventAttrs {
			doc.Find("[" + attr + "]").Each(func(i int, s *goquery.Selection) {
				eventHandlerCount++
				pageSignals[attr] = true
			})
		}

		// ── Forms and input elements ────────────────────────────
		doc.Find("form").Each(func(i int, s *goquery.Selection) {
			formCount++
		})
		doc.Find("input, select, textarea").Each(func(i int, s *goquery.Selection) {
			inputCount++
		})

		// ── Type hint heuristic ─────────────────────────────────
		typeHint := ifpClassifyPage(canvasCount, inlineScriptCount, eventHandlerCount,
			formCount, inputCount, pageSignals)

		if canvasCount > 0 {
			canvasPagesCount++
		}
		if typeHint != "static" {
			interactivePagesCount++
		}

		// Materialise signals list (deterministic order)
		signalsList := make([]string, 0, len(pageSignals))
		for sig := range pageSignals {
			signalsList = append(signalsList, sig)
		}
		sort.Strings(signalsList)

		pageEntry := map[string]interface{}{
			"page_index":          pageIdx,
			"url":                 pageURL,
			"type_hint":           typeHint,
			"canvas_count":        canvasCount,
			"inline_script_count": inlineScriptCount,
			"external_scripts":    externalScripts,
			"event_handler_count": eventHandlerCount,
			"form_count":          formCount,
			"input_count":         inputCount,
			"signals":             signalsList,
		}
		if len(canvasDetails) > 0 {
			pageEntry["canvas_details"] = canvasDetails
		}
		if len(inlineScripts) > 0 {
			pageEntry["inline_script_samples"] = inlineScripts
		}
		perPage = append(perPage, pageEntry)
	}

	// ── Build result ────────────────────────────────────────────────
	externalURLList := make([]string, 0, len(externalJSURLs))
	for u := range externalJSURLs {
		externalURLList = append(externalURLList, u)
	}
	sort.Strings(externalURLList)

	result := map[string]interface{}{
		"status": "extracted",
		"meta": map[string]interface{}{
			"pages_analyzed":          pagesAnalyzed,
			"inline_script_blocks":    inlineScriptBlocks,
			"external_script_count":   externalScriptCount,
			"canvas_pages_count":      canvasPagesCount,
			"interactive_pages_count": interactivePagesCount,
			"scripts_truncated":       truncatedScripts,
		},
		"per_page":         perPage,
		"external_js_urls": externalURLList,
		"library_signals":  librarySignals,
	}

	logger.Info("Interactive fingerprint extracted",
		zap.Int("pages_analyzed", pagesAnalyzed),
		zap.Int("inline_scripts", inlineScriptBlocks),
		zap.Int("external_scripts", externalScriptCount),
		zap.Int("canvas_pages", canvasPagesCount),
		zap.Int("interactive_pages", interactivePagesCount),
		zap.Int("external_js_urls", len(externalURLList)),
		zap.Int("scripts_truncated", truncatedScripts),
	)

	return result, nil
}

// ── Library signal detection ───────────────────────────────────────────

// ifpDetectLibrarySignals scans inline JS for known framework/API signatures.
// Increments aggregate counters and marks the per-page signals map.
func ifpDetectLibrarySignals(jsText string, aggregate map[string]int, pageSignals map[string]bool) {
	if ifpCanvas2DRe.MatchString(jsText) {
		aggregate["canvas_2d_context"]++
		pageSignals["canvas_2d_context"] = true
	}
	if ifpCanvasGLRe.MatchString(jsText) {
		aggregate["webgl_context"]++
		pageSignals["webgl_context"] = true
	}
	if ifpRAFRe.MatchString(jsText) {
		aggregate["requestAnimationFrame"]++
		pageSignals["requestAnimationFrame"] = true
	}
	if ifpAddEventRe.MatchString(jsText) {
		aggregate["addEventListener"]++
		pageSignals["addEventListener"] = true
	}
	if ifpFetchRe.MatchString(jsText) {
		aggregate["fetch"]++
		pageSignals["fetch"] = true
	}
	if ifpWebSocketRe.MatchString(jsText) {
		aggregate["websocket"]++
		pageSignals["websocket"] = true
	}
	if ifpJqueryRe.MatchString(jsText) {
		aggregate["jquery"]++
		pageSignals["jquery"] = true
	}
	if ifpThreeRe.MatchString(jsText) {
		aggregate["three_js"]++
		pageSignals["three_js"] = true
	}
	if ifpPhaserRe.MatchString(jsText) {
		aggregate["phaser"]++
		pageSignals["phaser"] = true
	}
	// p5 is characterised by BOTH setup() AND draw() functions present
	if ifpP5SetupRe.MatchString(jsText) && ifpP5DrawRe.MatchString(jsText) {
		aggregate["p5"]++
		pageSignals["p5"] = true
	}
	if ifpReactRe.MatchString(jsText) {
		aggregate["react"]++
		pageSignals["react"] = true
	}
	if ifpVueRe.MatchString(jsText) {
		aggregate["vue"]++
		pageSignals["vue"] = true
	}
}

// ── Type hint heuristic ────────────────────────────────────────────────

// ifpClassifyPage produces a starting-point classification from the extracted
// signals. The LLM does the real categorisation downstream — this is just a
// hint to guide the brief generation.
//
// Order matters: more specific patterns checked first.
func ifpClassifyPage(canvasCount, inlineScriptCount, eventHandlerCount, formCount,
	inputCount int, signals map[string]bool) string {

	// Game / animation: canvas with an animation loop
	if canvasCount > 0 && (signals["requestAnimationFrame"] ||
		signals["canvas_2d_context"] || signals["webgl_context"] ||
		signals["phaser"] || signals["three_js"] || signals["p5"]) {
		return "game_or_animation"
	}

	// Calculator: form with inputs and a script with event listeners
	if formCount > 0 && inputCount >= 2 &&
		(signals["addEventListener"] || eventHandlerCount > 0) {
		return "calculator"
	}

	// Interactive widget: script with event handlers but no form
	if inlineScriptCount > 0 &&
		(signals["addEventListener"] || eventHandlerCount > 0) {
		return "interactive_widget"
	}

	// Form-only (no script logic detected — likely a contact form)
	if formCount > 0 && inputCount > 0 {
		return "form"
	}

	return "static"
}
