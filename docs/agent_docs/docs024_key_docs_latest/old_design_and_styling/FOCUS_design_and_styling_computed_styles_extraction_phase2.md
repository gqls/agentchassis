# Phase 2: Computed Styles Extraction

Three changes: adapter passthrough, Go action, workflow SQL.

---

## 1. Adapter Change: Pass `actions` to firecrawl

In `webscrape-adapter` → `firecrawl_provider.go` → `Scrape()` function,
add this block after the `waitFor` check (around line 4715):

```go
	if waitFor > 0 {
		payload["waitFor"] = waitFor
	}

	// NEW: Pass through firecrawl actions (JS execution, wait, click, etc.)
	if actions, ok := config["actions"].([]interface{}); ok && len(actions) > 0 {
		payload["actions"] = actions
	}
```

This is safe — if no `actions` in config, nothing changes. The webscrape
adapter just needs a rebuild after this change.

---

## 2. Go Action: extract_computed_styles_action.go

New file in `platform/orchestration/actions/`:

```go
// FILE: platform/orchestration/actions/extract_computed_styles_action.go
//
// ExtractComputedStylesAction reads the rawHtml from a JS-augmented scrape
// result, finds the injected #__design_tokens__ element, parses the JSON,
// and merges the computed styles into the existing design fingerprint.
//
// This gives us the browser-resolved values — after cascade, specificity,
// media queries, and JS modifications have been applied. Much more reliable
// than parsing CSS source.
//
// Step Zero:
//   - extract_design_fingerprint: parses CSS source from rawHtml. Different approach.
//   - enrich_fingerprint_with_css: parses external CSS files. Source-level.
//   - No existing action extracts browser-computed styles.
//   Decision: New action needed.
//
// Registration (add to registry.go):
//   "extract_computed_styles": {
//       Handler:     ExtractComputedStylesAction,
//       Category:    "analysis",
//       Description: "Extract browser-computed design tokens from JS-augmented scrape",
//       IsLocal:     true,
//   },

package actions

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var ExtractComputedStylesInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{},
	Optional:   []string{"scrape_field", "fingerprint_field"},
	Defaults:   map[string]interface{}{"scrape_field": "computed_styles_scrape", "fingerprint_field": "design_fingerprint"},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("extract_computed_styles", ExtractComputedStylesInputSpec)
}

func ExtractComputedStylesAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "extract_computed_styles"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	scrapeField := "computed_styles_scrape"
	if sf, ok := config["scrape_field"].(string); ok && sf != "" {
		scrapeField = sf
	}

	fpField := "design_fingerprint"
	if ff, ok := config["fingerprint_field"].(string); ok && ff != "" {
		fpField = ff
	}

	// ── Load existing fingerprint ───────────────────────────────────
	fpRaw := datahelpers.ExtractNestedField(params.CollectedData, fpField)
	fp := make(map[string]interface{})
	if fpRaw != nil {
		fpUnwrapped := datahelpers.UnwrapDeep(fpRaw, logger)
		if fpMap, ok := fpUnwrapped.(map[string]interface{}); ok {
			fp = fpMap
		}
	}

	// ── Find rawHtml in scrape result ───────────────────────────────
	rawHtmlPaths := []string{
		scrapeField + ".response.data.raw_html",
		scrapeField + ".response.data.rawHtml",
		scrapeField + ".response.data.html_content",
		scrapeField + ".raw_html",
		scrapeField + ".rawHtml",
	}

	var rawHtml string
	for _, path := range rawHtmlPaths {
		raw := datahelpers.ExtractNestedField(params.CollectedData, path)
		if raw != nil {
			if s, ok := raw.(string); ok && len(s) > 100 {
				rawHtml = s
				logger.Info("Found rawHtml for computed styles",
					zap.String("path", path),
					zap.Int("length", len(s)))
				break
			}
		}
	}

	if rawHtml == "" {
		logger.Warn("No rawHtml found in scrape result")
		fp["computed_styles_status"] = "no_html"
		return fp, nil
	}

	// ── Parse HTML and find injected design tokens ──────────────────
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(rawHtml))
	if err != nil {
		logger.Warn("Failed to parse rawHtml", zap.Error(err))
		fp["computed_styles_status"] = "parse_error"
		return fp, nil
	}

	tokenEl := doc.Find("#__design_tokens__")
	if tokenEl.Length() == 0 {
		logger.Warn("No #__design_tokens__ element found — JS may not have executed")
		fp["computed_styles_status"] = "no_tokens_element"
		return fp, nil
	}

	tokenJSON := strings.TrimSpace(tokenEl.Text())
	if tokenJSON == "" {
		logger.Warn("#__design_tokens__ element is empty")
		fp["computed_styles_status"] = "empty_tokens"
		return fp, nil
	}

	var tokens map[string]interface{}
	if err := json.Unmarshal([]byte(tokenJSON), &tokens); err != nil {
		logger.Warn("Failed to parse design tokens JSON",
			zap.Error(err),
			zap.String("preview", tokenJSON[:min(200, len(tokenJSON))]))
		fp["computed_styles_status"] = "json_error"
		return fp, nil
	}

	logger.Info("Extracted computed design tokens",
		zap.Int("token_count", len(tokens)))

	// ── Merge computed styles into fingerprint ──────────────────────
	// Computed values are the ground truth — they override source-parsed values.

	// CSS variables (from all stylesheets, resolved)
	if cssVars, ok := tokens["css_variables"].(map[string]interface{}); ok && len(cssVars) > 0 {
		existingVars, _ := fp["css_variables"].(map[string]interface{})
		if existingVars == nil {
			existingVars = make(map[string]interface{})
		}
		for k, v := range cssVars {
			existingVars[k] = v // computed wins
		}
		fp["css_variables"] = existingVars
		logger.Info("Merged computed CSS variables",
			zap.Int("count", len(cssVars)))
	}

	// Computed colours from key elements
	if computed, ok := tokens["computed"].(map[string]interface{}); ok {
		fp["computed_styles"] = computed
	}

	// Element-specific computed styles
	if elements, ok := tokens["elements"].(map[string]interface{}); ok {
		fp["computed_elements"] = elements
	}

	// Build suggested mapping from computed data
	suggested, _ := fp["suggested_mapping"].(map[string]interface{})
	if suggested == nil {
		suggestedStr, _ := fp["suggested_mapping"].(map[string]string)
		suggested = make(map[string]interface{})
		for k, v := range suggestedStr {
			suggested[k] = v
		}
	}

	// Computed background/text/font override frequency-based guesses
	if computed, ok := tokens["computed"].(map[string]interface{}); ok {
		if bg, ok := computed["background"].(string); ok && bg != "" {
			hex := cssColorToHex(bg)
			if hex != "" {
				suggested["background"] = hex
			}
		}
		if text, ok := computed["text"].(string); ok && text != "" {
			hex := cssColorToHex(text)
			if hex != "" {
				suggested["text"] = hex
			}
		}
		if font, ok := computed["font_family"].(string); ok && font != "" {
			suggested["font_family"] = font
		}
	}

	// CSS variables override with correct names
	if cssVars, ok := tokens["css_variables"].(map[string]interface{}); ok {
		for varName, varValue := range cssVars {
			valStr, ok := varValue.(string)
			if !ok || valStr == "" {
				continue
			}
			lower := strings.ToLower(varName)
			switch {
			case strings.Contains(lower, "primary") && suggested["primary"] == nil:
				suggested["primary"] = strings.TrimSpace(valStr)
			case strings.Contains(lower, "accent") && suggested["accent"] == nil:
				suggested["accent"] = strings.TrimSpace(valStr)
			case (strings.Contains(lower, "bg") || strings.Contains(lower, "background")) && suggested["background"] == nil:
				suggested["background"] = strings.TrimSpace(valStr)
			case (strings.Contains(lower, "surface") || strings.Contains(lower, "card")) && suggested["surface"] == nil:
				suggested["surface"] = strings.TrimSpace(valStr)
			case strings.Contains(lower, "text") && !strings.Contains(lower, "muted") && suggested["text"] == nil:
				suggested["text"] = strings.TrimSpace(valStr)
			}
		}
	}

	fp["suggested_mapping"] = suggested
	fp["computed_styles_status"] = "extracted"
	fp["computed_styles_stats"] = map[string]interface{}{
		"css_vars_found":   countMapKeys(tokens, "css_variables"),
		"elements_sampled": countMapKeys(tokens, "elements"),
	}

	return fp, nil
}

// cssColorToHex converts rgb(r, g, b) or rgba(r, g, b, a) to #hex.
// Returns empty string if not a recognised format.
func cssColorToHex(color string) string {
	color = strings.TrimSpace(color)
	if strings.HasPrefix(color, "#") && (len(color) == 4 || len(color) == 7) {
		return color
	}
	if !strings.HasPrefix(color, "rgb") {
		return ""
	}
	// Extract numbers from rgb(r, g, b) or rgba(r, g, b, a)
	inner := color
	inner = strings.TrimPrefix(inner, "rgba(")
	inner = strings.TrimPrefix(inner, "rgb(")
	inner = strings.TrimSuffix(inner, ")")
	parts := strings.Split(inner, ",")
	if len(parts) < 3 {
		return ""
	}
	var r, g, b int
	fmt.Sscanf(strings.TrimSpace(parts[0]), "%d", &r)
	fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &g)
	fmt.Sscanf(strings.TrimSpace(parts[2]), "%d", &b)
	if r > 255 || g > 255 || b > 255 {
		return ""
	}
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func countMapKeys(m map[string]interface{}, key string) int {
	if sub, ok := m[key].(map[string]interface{}); ok {
		return len(sub)
	}
	return 0
}
```

Add to `registry.go`:
```go
"extract_computed_styles": {
    Handler:     ExtractComputedStylesAction,
    Category:    "analysis",
    Description: "Extract browser-computed design tokens from JS-augmented scrape",
    IsLocal:     true,
},
```

---

## 3. Workflow SQL

Adds two steps between `extract_fingerprint` and `check_has_external_css`:
- `scrape_computed_styles` — scrapes homepage with JS that injects design tokens
- `merge_computed_styles` — Go action reads tokens from rawHtml, merges into fingerprint

```sql
-- Add computed styles extraction to adoption workflow
-- Insert between extract_fingerprint and check_has_external_css

UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,extract_fingerprint,next_step}',
    '"scrape_computed_styles"'
)
WHERE type = 'site-adoption-agent';

-- Add the scrape step with JS injection
UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,scrape_computed_styles}',
    $step${
        "action": "firecrawl_scrape",
        "description": "Scrape homepage with JS to extract browser-computed design tokens",
        "output_field": "computed_styles_scrape",
        "next_step": "merge_computed_styles",
        "error_step": "check_has_external_css",
        "config": {
            "url_field": "input_data.url",
            "scrape_config": {
                "formats": ["rawHtml"],
                "only_main_content": false,
                "wait_for": 2000,
                "actions": [
                    {
                        "type": "executeJavascript",
                        "script": "try { const tokens = {}; const cs = getComputedStyle(document.documentElement); const bodyCs = getComputedStyle(document.body); tokens.computed = { background: bodyCs.backgroundColor, text: bodyCs.color, font_family: bodyCs.fontFamily, font_size: bodyCs.fontSize, line_height: bodyCs.lineHeight }; tokens.css_variables = {}; for (const sheet of document.styleSheets) { try { for (const rule of sheet.cssRules) { if (rule.selectorText && (rule.selectorText === ':root' || rule.selectorText === 'html' || rule.selectorText.includes(':root'))) { for (let i = 0; i < rule.style.length; i++) { const prop = rule.style[i]; if (prop.startsWith('--')) { tokens.css_variables[prop] = rule.style.getPropertyValue(prop).trim(); } } } } } catch(e) {} } tokens.elements = {}; const selectors = { header: 'header, .site-header, [class*=header]', hero: '.hero, [class*=hero], main > section:first-child', nav_link: '.main-nav a, nav a, .site-header a', footer: 'footer, .site-footer, [class*=footer]', heading: 'h1, h2', card: '.card, [class*=card], [class*=pillar]', accent_element: 'a, .btn, button, [class*=accent], [class*=primary]' }; for (const [name, sel] of Object.entries(selectors)) { const el = document.querySelector(sel); if (el) { const s = getComputedStyle(el); tokens.elements[name] = { background: s.backgroundColor, color: s.color, font_family: s.fontFamily, font_size: s.fontSize, font_weight: s.fontWeight, border_color: s.borderColor, padding: s.padding }; } } const tag = document.createElement('script'); tag.type = 'application/json'; tag.id = '__design_tokens__'; tag.textContent = JSON.stringify(tokens); document.body.appendChild(tag); } catch(e) { const tag = document.createElement('script'); tag.type = 'application/json'; tag.id = '__design_tokens__'; tag.textContent = JSON.stringify({error: e.message}); document.body.appendChild(tag); }"
                    }
                ]
            }
        }
    }$step$::jsonb
)
WHERE type = 'site-adoption-agent';

-- Add the merge step
UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,merge_computed_styles}',
    '{
        "action": "extract_computed_styles",
        "description": "Parse browser-computed design tokens and merge into fingerprint",
        "output_field": "design_fingerprint",
        "next_step": "check_has_external_css",
        "error_step": "check_has_external_css",
        "config": {
            "scrape_field": "computed_styles_scrape",
            "fingerprint_field": "design_fingerprint"
        }
    }'::jsonb
)
WHERE type = 'site-adoption-agent';
```

Verify the wiring:
```sql
SELECT
    default_config->'workflow'->'steps'->'extract_fingerprint'->>'next_step' as after_fingerprint,
    default_config->'workflow'->'steps'->'scrape_computed_styles'->>'next_step' as after_scrape_cs,
    default_config->'workflow'->'steps'->'scrape_computed_styles'->>'error_step' as scrape_cs_error,
    default_config->'workflow'->'steps'->'merge_computed_styles'->>'next_step' as after_merge_cs,
    default_config->'workflow'->'steps'->'merge_computed_styles'->>'error_step' as merge_cs_error
FROM agent_definitions WHERE type = 'site-adoption-agent';
```

Expected: `scrape_computed_styles` → `merge_computed_styles` → `check_has_external_css`
Error paths both fall through to `check_has_external_css` so the pipeline continues if JS execution fails.

---

## Updated adoption workflow

```
ensure_site_record
  → crawl_site (firecrawl — all pages, rawHtml + markdown)
  → format_crawl (summaries for LLM)
  → check_crawl_content
  → extract_fingerprint (Go — parse CSS source from <style> blocks)
  → scrape_computed_styles (firecrawl — homepage only, with JS)     ← NEW
  → merge_computed_styles (Go — parse tokens, merge into fingerprint) ← NEW
  → check_has_external_css → fetch_primary_css → enrich_fingerprint
  → analyze_site (LLM)
  → classify_archetype (LLM)
  → select_content → derive_content_direction (LLM)
  → apply_plan (Go — writes specs, pages, work items)
  → generate_design_intent → write_design_intent
  → complete
```

The computed styles step runs on just the homepage (one scrape, not a
full crawl). It waits 2s for JS to execute, runs the extraction script,
and the Go action reads the injected JSON from the HTML.

If JS execution fails (firecrawl doesn't support it, CORS blocks
stylesheets, the page errors), the error_step routes to
check_has_external_css and the pipeline continues with source-parsed
data only. No blocking failure.
