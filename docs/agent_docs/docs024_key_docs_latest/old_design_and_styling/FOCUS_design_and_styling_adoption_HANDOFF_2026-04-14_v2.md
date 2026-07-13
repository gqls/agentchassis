# Handoff: Design Adoption Pipeline — 2026-04-14

---

## Current State

The adoption pipeline is partially working. The crawl, fingerprint extraction, external CSS detection, and CSS fetch all work. The pipeline fails at two points that are being fixed:

1. **`classify_archetype` template rendering** — LLM output from `analyze_site` is truncated JSON (~8K chars cut mid-value). `UnwrapDeep` can't parse it, so Go templates can't traverse the fields. Workaround: use `{{.adoption_analysis}}` to dump the raw wrapper for the LLM to read.

2. **`enrich_fingerprint_with_css` can't find CSS content** — the scrape result uses `response.data.raw_html` (underscore, nested under `response.data`) but the code looks for `rawHtml` (camelCase, flat paths). Fix: update `cssPaths` in the Go action.

### What's deployed

| Component | Status |
|-----------|--------|
| `extract_design_fingerprint` action + registry | ✅ Deployed, works. Finds colours, external CSS URLs. |
| `enrich_fingerprint_with_css` action + registry | ✅ Deployed, but **CSS path bug** — can't find content in scrape result |
| `UnwrapDeep` in `ExtractFields` (unified_extractor.go) | ✅ Deployed (v1.0.961), but JSON truncation means it can't parse |
| Adoption workflow SQL (all steps wired) | ✅ Applied |
| Webdesign-agent three-way prompt | ✅ Applied |
| `archive_completed_work_items` function | ✅ Created and tested |
| `work-item-archiver` agent + scheduled task | ✅ Created |

### What needs doing

| Priority | Fix | Type | Status |
|----------|-----|------|--------|
| 1 | Template workaround: `{{.adoption_analysis}}` in classify_archetype + generate_design_intent | SQL | **Apply now** |
| 2 | Fix `cssPaths` in `enrich_fingerprint_with_css_action.go` — add `response.data.raw_html` paths | Go | **Ready to deploy** |
| 3 | Investigate `analyze_site` JSON truncation — 7930 chars cut mid-value despite `max_tokens: 32000` | Investigation | Not started |
| 4 | Add `repairTruncatedJSON` fallback in `UnwrapDeep` | Go | Not started |
| 5 | Change `tryParseJSON` debug logging to Info level for failures | Go | Not started |
| 6 | Split `analyze_site` into focused steps (identity/pages/interactive) | Architecture | Planning |

---

## Adoption Workflow (current deployed state)

```
ensure_site_record
  → crawl_site (firecrawl — 20 pages found for gamedesign.uk)
  → format_crawl (summaries — 500 chars/page)
  → check_crawl_content
  → extract_fingerprint (Go — finds colours, layout, external CSS URLs)  ✅ works
  → check_has_external_css → routes to fetch_primary_css              ✅ works
  → fetch_primary_css (webscrape adapter — fetches global.css)         ✅ works
  → enrich_fingerprint (Go — ❌ can't find CSS in response.data.raw_html)
  → analyze_site (LLM — ⚠️ output truncated at ~8K chars)
  → classify_archetype (LLM — ❌ fails on template rendering)
  → select_content → derive_content_direction
  → apply_plan (writes specs, pages, work items)
  → generate_design_intent → write_design_intent
  → complete
```

---

## Bugs Found This Session

### Bug 1: `enrich_fingerprint_with_css` wrong field paths

The webscrape adapter returns scrape results at `response.data.raw_html` (underscore, nested). The action looks for `rawHtml` (camelCase, flat). Result: `css_enrichment: "no_css_content_found"` — the most valuable design tokens (`--bg-color: #121212`, `--primary-color: #00bcd4`) are never extracted.

**Fix:** Update `cssPaths` in `enrich_fingerprint_with_css_action.go`:
```go
cssPaths := []string{
    cssField + ".response.data.raw_html",
    cssField + ".response.data.html_content",
    cssField + ".response.data.markdown_content",
    cssField + ".raw_html",
    cssField + ".html_content",
    cssField + ".rawHtml",
    cssField + ".body.data.rawHtml",
    cssField + ".data.rawHtml",
    cssField + ".markdown",
    cssField + ".content",
}
```

### Bug 2: `analyze_site` LLM output truncated

The LLM describes 20 pages in structured JSON. Output reaches ~7930 chars then stops mid-value (`"page_type": "`). `tryParseJSON` fails silently (logged at Debug level, invisible). `UnwrapDeep` returns the raw string wrapper. Go templates can't traverse into it.

**Workaround (SQL):** Use `{{.adoption_analysis}}` instead of `{{.adoption_analysis.identity}}` etc. The LLM receiving the data can parse partial JSON.

**Root cause options:**
- LLM ran out of output space (unlikely at 32K tokens for 8K chars)
- API response truncated in transit
- The webscrape adapter or coordinator truncated the response body

**Long-term fix:** Split `analyze_site` into smaller steps. Page inventory should be a Go action (deterministic from crawl metadata), not LLM-generated.

### Bug 3: `enrich_fingerprint_with_css` not in registry

Was missing from `registry.go`. Workflow validator rejected the entire adoption workflow with `action 'enrich_fingerprint_with_css' requires a topic` because it thought it was a remote action.

**Fix:** Added to registry.go with `IsLocal: true`. Deployed in v1.0.960.

### Bug 4: Template field traversal into `interface{}`

Go templates can't access map keys through an `interface{}` type barrier. Even after `UnwrapDeep`, if the JSON is a string (truncated, can't parse), the template `{{.adoption_analysis.result}}` fails with `can't evaluate field result in type interface{}`.

**Fix:** `UnwrapDeep` added to `ExtractFields` general loop (deployed v1.0.961). Works when JSON is valid. Falls back to raw string when truncated.

---

## Fingerprint Data (from last successful extraction)

gamedesign.uk fingerprint found:
- 8 colours: `#111111` (bg, count 12), `#00bcd4` (cyan accent, count 8), `#ff9800`, `#ff5252`, `#2e7d32`, `#4caf50`, `#555555`, `#5d4037`
- External CSS: `https://gamedesign.uk/css/global.css` — detected but content not yet extracted
- Dark sections: predominant scheme "mixed", dark backgrounds `#111111`, `#333333`
- No fonts extracted (they're in global.css, not inline styles)
- No CSS variables extracted (they're in global.css `:root` block)
- `suggested_mapping`: only `primary: #111111`, `text: #ff9800` (from inline colours)

After the CSS path fix, enrichment should add:
- CSS variables: `--bg-color: #121212`, `--primary-color: #00bcd4`, `--surface-color: #1e1e1e`, `--text-main: #e0e0e0`, `--font-family: 'Segoe UI', Roboto, ...`
- Font families from the CSS
- Correct suggested_mapping with the original's actual palette

---

## Architecture Improvements Discussed

### Split `analyze_site` into focused steps
- `analyze_identity` (LLM — small JSON: company, industry, tone)
- `analyze_pages` (Go — from crawl metadata, URL pattern classification)
- `analyze_interactive` (LLM — identify tools/calculators)

### Large site support (map → classify → selective crawl)
- `firecrawl_map` discovers all URLs (lightweight, no content)
- Go action classifies URLs by pattern, picks ~15 representatives
- `firecrawl_scrape` fetches only representatives
- Analysis works on representative set, page inventory from map

---

## Diagnostic Queries

```sql
-- Check adoption orchestration state
SELECT orchestration_id, status, current_step, LEFT(error, 300) as error, created_at
FROM orchestration_states
WHERE collected_data::text LIKE '%site-adoption-agent%'
ORDER BY created_at DESC LIMIT 1;

-- Check fingerprint data from latest run
SELECT 
    (collected_data->'design_fingerprint'->'has_external_css')::text as has_ext,
    (collected_data->'design_fingerprint'->'primary_css_url')::text as primary_url,
    (collected_data->'design_fingerprint'->'css_enrichment')::text as enrichment,
    LEFT((collected_data->'design_fingerprint'->'css_variables')::text, 300) as css_vars,
    LEFT((collected_data->'design_fingerprint'->'suggested_mapping')::text, 300) as mapping
FROM orchestration_states
WHERE collected_data::text LIKE '%site-adoption-agent%'
  AND collected_data::text LIKE '%design_fingerprint%'
ORDER BY created_at DESC LIMIT 1;

-- Check what specs exist
SELECT aspect, source, LEFT(data::text, 200) as preview, created_at
FROM site_specs
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk')
  AND is_current = true
ORDER BY aspect;

-- Check work items
SELECT item_type, status, COUNT(*)
FROM site_work_items
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk')
GROUP BY item_type, status ORDER BY item_type, status;

-- Check archived items
SELECT count(*), min(completed_at), max(completed_at)
FROM site_work_items_archive
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk');
```

---

## Files Changed This Session

| File | Change | Status |
|------|--------|--------|
| `unified_extractor.go` | Added `UnwrapDeep` to general field extraction loop | ✅ Deployed v1.0.961 |
| `registry.go` | Added `enrich_fingerprint_with_css` entry | ✅ Deployed v1.0.960 |
| `enrich_fingerprint_with_css_action.go` | `cssPaths` fix needed — `response.data.raw_html` | **Ready to deploy** |
| `agent_definitions` (site-adoption-agent) | Template workarounds for truncated JSON | **Apply SQL** |
| `archive_completed_work_items` function | New — archives terminal work items | ✅ Applied |
| `work-item-archiver` agent definition | New — daily archiver agent | ✅ Applied |
| `scheduled_tasks` (work-item-archiver) | New — daily 86400s trigger | ✅ Applied |
| `site_work_items_archive` | Added `pipeline`, `approval_mode`, `updated_at` columns | ✅ Applied |
