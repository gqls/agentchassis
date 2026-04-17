# Handoff: Design Adoption Pipeline — 2026-04-16 v2

## Session Transcript
Full conversation transcript at: `/mnt/transcripts/2026-04-16-09-13-00-design-adoption-pipeline-debug.txt`

---

## Victory: Design Fingerprint Now Correct

The adoption pipeline ran end-to-end successfully against the original gamedesign.uk site. The fingerprint captured the correct design tokens:

```
background: #121212    (was #080B10 / #0f3460 from our builds)
primary:    #00bcd4    (was #C8FF00 / #16a085 from our builds)  
surface:    #1e1e1e
text:       #e0e0e0
text_muted: #a0a0a0
border:     #333
font:       'Segoe UI', Roboto, Helvetica, Arial, sans-serif
```

CSS variables extracted from global.css:
```
--bg-color: #121212
--primary-color: #00bcd4
--surface-color: #1e1e1e
--text-main: #e0e0e0
--text-dim: #a0a0a0
--border-color: #333
--font-family: 'Segoe UI', Roboto, Helvetica, Arial, sans-serif
```

Latest successful orchestration: `a5a34254-8c79-45be-a229-8f37e739e295` (completed 2026-04-16, 5m37s runtime)

---

## Current Site State: gamedesign.uk

- **Site ID:** `15a6cb16-5a86-4541-a8e4-d7106239b6a4`
- **Build status:** pending
- **Pages:** 11 (clean — wiped and recreated by latest adoption)
- **Specs:** 6 current specs from adoption (all correct source data)
- **Style collection:** NULL (no CSS theme linked yet)
- **Image repo:** `docker.io/aqls/agent-chassis`, latest tag: `v1.0.961`

### Pages Created by Latest Adoption

```
name                        | page_type  | in_header | in_footer
----------------------------+------------+-----------+----------
games-index                 | content    | t         | t
guide-fairness-in-rng       | blog-post  | f         | f
guide-p2p-architecture      | blog-post  | f         | f
guide-rng-design            | blog-post  | f         | f
guides-index                | blog-index | t         | t
guide-skinner-box           | blog-post  | f         | f
index                       | landing    | t         | t
tool-ehp-calculator         | tool       | f         | f
tool-lanchester-sim         | tool       | f         | f
tool-progression-architect  | tool       | f         | f
tools-index                 | tool       | t         | t
```

### Current Specs

| Aspect | Source | Status |
|--------|--------|--------|
| identity | adoption | ✅ Correct — "The Utility Engine for Game Developers" |
| design_reference | adoption | ✅ Correct — #121212, #00bcd4, Segoe UI |
| design_intent | site-adoption-agent | ✅ Generated from correct fingerprint |
| site_archetype | adoption | ✅ "game design utility platform" |
| content_direction | adoption | ✅ Writing style guide extracted |
| structure | adoption | ✅ Page list with correct pages |

### Work Items

Work items were created by `apply_plan` and some may have been unpaused. Check status:
```sql
SELECT item_type, status, COUNT(*)
FROM site_work_items
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk')
  AND status NOT IN ('complete', 'wont_fix', 'failed')
GROUP BY item_type, status ORDER BY item_type, status;
```

---

## Problems Remaining (Prioritised)

### P1 — Header/Footer Components Have Hardcoded Colours

**What:** The header component template contains inline CSS with hardcoded `#1a1a2e` (gradient bg), `#16a085` (accent/CTA), not CSS variables. The footer has the same issue with `#1a1a2e`. These come from the component DB, not the CSS theme.

**Evidence:** In the deployed HTML:
```css
.site-header--gradient { background: linear-gradient(135deg, #1a1a2e 0%, #2d2d44 100%); }
.site-header--gradient .logo-icon { color: #16a085; }
.site-header--gradient .header-cta { background: #16a085; }
```

**Impact:** Even with correct CSS theme, header and footer show wrong colours.

**Fix options:**
- Option A: Header/footer templates should use `var(--color-header-bg)` etc. This requires updating the component templates in the content_components table.
- Option B: The webdesign-agent generates header/footer CSS that overrides the component's inline styles using the design_intent colours.
- Option C: The `needs_design` work item handler should also update header/footer component rendered HTML.

### P2 — Navigation Has Too Many Links

**What:** The header shows 9 nav links including sub-pages ("Tools/Tool Loot Probability Calculator", "Tools/Bayesian Ranking"), "Home", "About", "Get Started". The original had 3 (Tools, Guides, Games).

**Evidence:** The `in_header` flags in the pages table. Only `index`, `tools-index`, `guides-index`, `games-index` have `in_header: true`, which is better than before (was 14). But the header component template also pulls nav items from the old `site_nav_items` table or has them hardcoded.

**Root cause:** The header component's rendered HTML includes links from before the cleanup. The `rendered_header` column on pages contains the old nav. The header component template needs to be re-rendered with only the 4 `in_header: true` pages.

**Fix:** The `needs_rerender` work item should regenerate headers. Or check how the header component gets its nav links — likely from `site_nav_items` table or from the pages query filtered by `in_header = true`.

```sql
-- Check what nav items exist
SELECT * FROM site_nav_items 
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk')
ORDER BY position;
```

### P3 — Empty `<main>` Content

**What:** The index page and other pages have `<main></main>` with no content inside.

**Root cause:** The `needs_content_page` work items haven't been processed yet (they were paused/just unpaused). The content writer agent needs to run for each page.

**Fix:** Let the `needs_content_page` work items complete. They should generate section content based on the specs. Monitor:
```sql
SELECT item_type, status, LEFT(summary, 60) 
FROM site_work_items 
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk')
  AND item_type = 'needs_content_page'
ORDER BY status;
```

### P4 — CSS Theme Not Yet Generated

**What:** `sites.style_collection_id` is NULL. No CSS theme exists for the site yet.

**Root cause:** The `needs_design` work item either hasn't run or failed. The webdesign-agent needs to generate CSS from the `design_intent` spec.

**Fix:** Check the `needs_design` work item status:
```sql
SELECT status, LEFT(error, 200) as error, LEFT(summary, 100) as summary
FROM site_work_items
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk')
  AND item_type = 'needs_design';
```

If it failed, check why. If triaged, let it run. The webdesign-agent should read the `design_intent` (which now has the correct `#121212`/`#00bcd4` palette) and generate a CSS theme.

### P5 — `analyze_site` LLM Output Truncation

**What:** The `analyze_site` step produces ~8K chars of JSON that gets truncated mid-value. `UnwrapDeep` can't parse invalid JSON, so downstream templates can't traverse fields.

**Workaround applied:** Templates use `{{.adoption_analysis}}` to dump the raw text. The receiving LLM parses what it can.

**Long-term fix:** Split `analyze_site` into:
- `analyze_identity` (LLM — small JSON: company, industry, tone)
- `analyze_pages` (Go — from crawl metadata, URL pattern classification)
- `analyze_interactive` (LLM — identify tools/calculators)

Also add `repairTruncatedJSON` fallback in `UnwrapDeep`, and change `tryParseJSON` failure logging from Debug to Info level.

### P6 — `fpExtractCSSVars` BEM False Positives

**What:** The regex `(--[\w-]+)\s*:\s*([^;}{]+)` captures BEM class selectors like `.btn--primary:hover` as `--primary: hover`.

**Fix ready:** Replacement code at `/mnt/user-data/outputs/fp_extract_css_vars_replacement.go` with integration instructions at `/mnt/user-data/outputs/fp_extract_css_vars_integration.md`. Uses `:root` block targeting with semicolon-splitting instead of regex.

**Status:** Not yet deployed. The current extraction was ok because the original's CSS is clean, but this will break on other sites.

---

## Fixes Deployed This Session

| Fix | Type | Deployed |
|-----|------|----------|
| `enrich_fingerprint_with_css` registry entry | Go (registry.go) | ✅ v1.0.960 |
| `UnwrapDeep` in `ExtractFields` general loop | Go (unified_extractor.go) | ✅ v1.0.961 |
| `cssPaths` fix — `response.data.raw_html` paths | Go (enrich_fingerprint_with_css_action.go) | ✅ Deployed |
| `maxAge` passthrough in adapter Scrape() and Crawl() | Go (firecrawl.go) | ✅ Deployed |
| Template workaround — `{{.adoption_analysis}}` | SQL (agent_definitions) | ✅ Applied |
| `classify_archetype` structured field access | SQL (agent_definitions) | ✅ Applied (reverted to workaround) |
| `generate_design_intent` structured field access | SQL (agent_definitions) | ✅ Applied (reverted to workaround) |
| `archive_completed_work_items` function | SQL | ✅ Applied |
| `work-item-archiver` agent + scheduled task | SQL | ✅ Applied |
| `site_work_items_archive` schema sync | SQL | ✅ Applied |
| `max_age: 600000` in crawl config | SQL (agent_definitions) | ✅ Applied |

## Fixes Ready But Not Deployed

| Fix | Type | Location |
|-----|------|----------|
| `fpExtractCSSVars` `:root` block replacement | Go | `/mnt/user-data/outputs/fp_extract_css_vars_replacement.go` |
| Computed styles extraction (Phase 2) | Go + SQL | `/mnt/user-data/outputs/computed_styles_extraction_phase2.md` |

---

## Adoption Workflow (Current Deployed State)

```
ensure_site_record
  → crawl_site (firecrawl_crawl — limit 30, max_age 600000ms)
  → format_crawl (summaries — 350 chars/page)
  → check_crawl_content
  → extract_fingerprint (Go — colours, layout, external CSS URLs from inline styles)
  → check_has_external_css → fetch_primary_css (firecrawl_scrape)
  → enrich_fingerprint (Go — parses external CSS, merges into fingerprint) ✅ WORKING
  → analyze_site (LLM — identity, design, pages, interactive features) ⚠️ truncates
  → classify_archetype (LLM — site character, purpose, constraints) ✅ workaround
  → select_content → derive_content_direction (LLM — writing style guide)
  → apply_plan (Go — writes specs, creates pages, creates work items)
  → generate_design_intent (LLM — semantic design brief) ✅ workaround
  → write_design_intent → complete
```

Error paths: CSS fetch failure → skips to analyze_site. Content selection failure → skips to apply_plan. Design intent failure → completes anyway.

---

## Key Database Queries

```sql
-- Check site state
SELECT domain, build_status, style_collection_id FROM sites WHERE domain = 'gamedesign.uk';

-- Check pages
SELECT name, page_type, build_status, in_header FROM pages
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk') ORDER BY name;

-- Check nav items (may be separate from pages)
SELECT * FROM site_nav_items 
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk') ORDER BY position;

-- Check work items
SELECT item_type, status, COUNT(*), LEFT(string_agg(DISTINCT summary, ' | '), 100)
FROM site_work_items
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk')
  AND status NOT IN ('complete', 'wont_fix', 'failed')
GROUP BY item_type, status ORDER BY status, item_type;

-- Check specs
SELECT aspect, source, LEFT(data::text, 200), created_at
FROM site_specs
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk')
  AND is_current = true ORDER BY aspect;

-- Check fingerprint from latest adoption
SELECT 
    (collected_data->'design_fingerprint'->'css_enrichment')::text as enrichment,
    LEFT((collected_data->'design_fingerprint'->'suggested_mapping')::text, 400) as mapping,
    LEFT((collected_data->'design_fingerprint'->'css_variables')::text, 400) as css_vars
FROM orchestration_states
WHERE orchestration_id = 'a5a34254-8c79-45be-a229-8f37e739e295';

-- Check CSS theme (when generated)
SELECT ct.name, LEFT(ct.css_content, 500), ct.updated_at
FROM sites s
JOIN style_collections sc ON sc.id = s.style_collection_id
JOIN css_themes ct ON ct.id = sc.css_theme_id
WHERE s.domain = 'gamedesign.uk';

-- Check header component template
SELECT id, name, LEFT(html_content, 500) 
FROM content_components 
WHERE name LIKE '%header%' AND is_active = true;

-- Full cleanup procedure (for re-adoption)
BEGIN;
SELECT take_site_snapshot((SELECT id FROM sites WHERE domain = 'gamedesign.uk'), 'manual', NULL, 'Pre-wipe', 'admin');
UPDATE site_work_items SET status = 'wont_fix' WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk') AND status NOT IN ('complete', 'wont_fix', 'failed');
SELECT archive_completed_work_items(0, 10000);
UPDATE site_specs SET is_current = false, superseded_at = NOW() WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk') AND is_current = true;
DELETE FROM research_results WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk');
DELETE FROM page_component_history WHERE page_id IN (SELECT id FROM pages WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk'));
DELETE FROM redirects WHERE source_page_id IN (SELECT id FROM pages WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk'));
DELETE FROM pages WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk');
UPDATE sites SET build_status = 'pending', style_collection_id = NULL WHERE domain = 'gamedesign.uk';
COMMIT;
```

---

## Architecture Decisions Made

1. **`UnwrapDeep` in ExtractFields** — one-line fix in unified_extractor.go. All LLM outputs automatically unwrapped before templates see them. Works when JSON is valid; falls through to raw string when truncated.

2. **Template workaround for truncated JSON** — `{{.adoption_analysis}}` dumps the raw wrapper text. The LLM receiving it can parse partial JSON. Not ideal but unblocks the pipeline without Go changes.

3. **`:root` block targeting for CSS vars** — replacement for the regex-based approach. Semicolon-splitting within identified token blocks avoids BEM false positives and handles minified CSS.

4. **`maxAge` for firecrawl cache control** — adapter passes through `max_age` from config to firecrawl API. Set to 600000 (10 min) for adoption crawls. `skipCache` does NOT work with v2 API.

5. **Computed styles (Phase 2) deferred** — JS execution via firecrawl actions to extract `getComputedStyle()` values. Spec written but not implemented. Not needed while source CSS parsing works for the current site.

6. **Adopt-from vs deploy-to separation** — discussed but not implemented. Options: snapshot to S3, stage to subdomain, or store crawl artifacts. The current workaround is pause work items after adoption, verify specs, then unpause.

---

## Files Reference

| File | What | Status |
|------|------|--------|
| `/mnt/project/production_agent-chassis-full_context.txt` | Main Go source | Reference |
| `/mnt/project/production_agent-adapters_context.txt` | Adapter source | Reference |
| `/mnt/project/bk_agent_definitions_backup.sql` | DB backup | Reference |
| `/mnt/user-data/outputs/fp_extract_css_vars_replacement.go` | `:root` CSS extraction | Ready to deploy |
| `/mnt/user-data/outputs/fp_extract_css_vars_integration.md` | Integration instructions | Reference |
| `/mnt/user-data/outputs/computed_styles_extraction_phase2.md` | Phase 2 spec | Deferred |
| `/mnt/user-data/outputs/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-14.md` | Earlier handoff | Superseded by this doc |

---

## What To Do Next

1. **Check work items are running** — the `needs_design` and `needs_content_page` items should process
2. **When `needs_design` completes**, check the CSS theme has `#121212`/`#00bcd4` palette
3. **Fix the header/footer component** — hardcoded colours need to use CSS variables or be regenerated
4. **Fix the nav links** — check `site_nav_items` table, ensure only Tools/Guides/Games in header
5. **Deploy `fpExtractCSSVars` fix** — before adopting any other sites
6. **Consider the `analyze_site` split** — reduce truncation risk for larger sites
7. **Remove `max_age: 600000` from scrape_config** — or increase it, after adoption is stable. 0 forces fresh crawl every time which is expensive.
