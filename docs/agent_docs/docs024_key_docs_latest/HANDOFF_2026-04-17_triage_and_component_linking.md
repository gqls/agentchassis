# Handoff: Triage Fix + Component Linking + Topic Splitting

**Date**: 2026-04-17  
**Test site**: gaswholesalers.com (`5fe15466-4e2e-4ff2-981e-98c1b7074002`)  
**Homepage page_id**: `4ff0e0ff-fab2-423e-a59c-b9de4674a84f`  
**Nav groups**: primary `8fc6381a-e87b-44af-ad3a-532542e49854`, utility `ceccf615-d23c-4279-90b5-ff15215fd4d4`

---

## What Was Done This Session

### 1. Triage Scoring — FIXED, WORKING

**Problem**: 200+ ingested items never scored. `apply_feed_scores` returned "no scores found in collected_data" every cycle since April 2nd.

**Root causes** (two independent issues):

**A. JSON truncation**: `max_items: 50` produced ~10K tokens of LLM output, exceeding `max_tokens: 4000` (later raised to 8192). Truncated JSON = invalid = parse failure.

**B. `max_items` config ignored**: `LoadFeedItemsForTriageAction` used `inputs.GetInt("max_items", 50)` which reads from `ai.Values` (populated from `collectedData`, not from `StepConfig.Config`). Config literal numbers are invisible to `ExtractActionInputs`. The working pattern (used by `LoadWorkItemsAction`) is `datahelpers.GetIntField(params.StepConfig.Config, "max_items", 50)`.

**C. Wrapper map not unwrapped**: `execute_llm_prompt` returns `{"type": "text", "result": "[...]"}`. `apply_feed_scores` tried `scores.result` (nil via ExtractNestedField), fell back to `scores` (got the wrapper map), but didn't unwrap the `result` key from the map.

**Fixes deployed**:
- `feed_triage_actions.go` — changed `inputs.GetInt` → `datahelpers.GetIntField(params.StepConfig.Config, ...)` for max_items
- `feed_triage_actions.go` — added wrapper map unwrap for scores (map → result key → string → parse)
- SQL: `max_items` set to 15 in agent_definitions config
- SQL: `max_tokens` raised to 8192
- SQL: `scores_field` set to `"scores.result"` (belt-and-braces with the unwrap code)

**Current state**: Triage is working. As of last check: 41 relevant, 23 rejected, 232 ingested (backlog clearing at 15 items per cycle). Credibility populated: 9 high, 20 medium, 1 low.

**Backlog**: `content-feed-refresh` interval temporarily set to 1800s (30 min). **Reset to 21600s (6h) once ingested count drops below 30**:
```sql
UPDATE scheduled_tasks SET interval_seconds = 21600 WHERE name = 'content-feed-refresh';
```

### 2. Scheduler Routing — IDENTIFIED, WORKAROUND

**Problem**: `content-feed-refresh` scheduled task fires to `system.agent.generic.requests` with `target_agent_type: content-feed-trigger`. The generic agent consumes the message without routing it to the trigger agent — the trigger has never run (0 orchestration states).

**Workaround**: Manual kcat trigger script spawns `content-feed-orchestrator` directly. The script is in the transcript.

**Proper fix not yet applied**: Either change the scheduled task's `target_topic` to `system.agent.content-feed-trigger.process`, or restructure the trigger to work via the generic agent spawn path. The complication: the trigger agent doesn't exist until spawned, so nothing listens on its topic.

### 3. save_page_sections component_id — FILE READY, NOT YET DEPLOYED

**Problem**: Every page rebuild wipes `component_id` on `page_components`. The `enrichSectionsWithComponentIDs` function runs in the shared block (after both metadata and HTML paths) but fails to link because:

**A. Metadata path name mismatch**: Metadata produces `ComponentName` from `component_function` (e.g. `differentiators-section`) but `content_components.function` has `differentiators`. The enrichment tried exact match + underscore variant, but never stripped common suffixes.

**B. HTML data-component not consulted**: The metadata path sets `ComponentName` from metadata, not from the `data-component` attribute in the rendered HTML. The HTML value is the authoritative one that matches `content_components.function`.

**C. All logging was `logger.Debug`**: Invisible in production logs.

**Current DB state for gaswholesalers homepage**:
```
position | slot_name               | linked | html_func
1        | hero                    | f      | hero
2        | features                | f      | features
3        | services-grid           | f      | services-grid
4        | differentiators-section | f      | differentiators    ← metadata/HTML name mismatch
5        | social-proof            | f      | social-proof
6        | latest-news             | f      | latest-news
7        | call-to-action          | f      | call-to-action
```

**News page**:
```
position | slot_name      | linked | html_func
1        | hero           | t      | hero
2        | news-listing   | f      | (null)     ← template missing data-component attr
3        | call-to-action | t      | call-to-action
```

**Fix ready** (`save_page_sections_action.go` in outputs):
- Extracts `data-component` from HTML and prefers it over metadata ComponentName
- Strips `-section`, `-container`, `-wrapper`, `-block` suffixes as fallback candidates
- Tries all candidates with exact + underscore variants in priority order
- All `logger.Debug` → `logger.Info`
- 654 → 716 lines

**After deploying**, also fix the news-listing template:
```sql
-- Column is html_template, not template_html
SELECT LEFT(html_template, 200) FROM content_components
WHERE function = 'news-listing' AND is_active = true;

-- Add data-component if missing
UPDATE content_components
SET html_template = REPLACE(
    html_template,
    '<section class="news-listing-section"',
    '<section data-component="news-listing" class="news-listing-section"'
)
WHERE function = 'news-listing' AND is_active = true
  AND html_template NOT LIKE '%data-component="news-listing"%';
```

**Verification after deploy**: Trigger a content rewrite or rerender for gaswholesalers homepage, then check:
```sql
SELECT pc.position, pc.slot_name, pc.component_id IS NOT NULL as linked
FROM page_components pc
WHERE pc.page_id = '4ff0e0ff-fab2-423e-a59c-b9de4674a84f'
ORDER BY pc.position;
```
All 7 should show `linked = true`.

### 4. Topic Splitting — NOT YET STARTED

Next task after component linking. Split the single Grok `api_news` source into topic-focused sources for better diversity. SQL-only change — no Go code needed. Plan is in `031_news_content_diversity_plan.md` § "Topic-Focused Source Splitting".

---

## Files Ready to Deploy

| File | Target | Status |
|------|--------|--------|
| `save_page_sections_action.go` | `platform/orchestration/actions/save_page_sections_action.go` | Ready, not deployed |
| `feed_triage_actions.go` | `platform/orchestration/actions/feed_triage_actions.go` | **Deployed** |
| `feed_news_recommendation_action.go` | `platform/orchestration/actions/feed_news_recommendation_action.go` | **Deployed** |
| `page_growth_budget.go` | `platform/orchestration/actions/page_growth_budget.go` | **Deployed** |
| `apply_gap_plan_action.go` | `platform/orchestration/actions/apply_gap_plan_action.go` | **Deployed** |

---

## SQL Changes Applied This Session

```sql
-- Improvement-sweep threshold raised from 20 to 50
UPDATE scheduled_tasks SET pre_query = '...<50...' WHERE name = 'improvement-sweep';

-- Stale-work-item-reaper scheduled task created
INSERT INTO scheduled_tasks (name, ...) VALUES ('stale-work-item-reaper', ...);

-- Triage batch size reduced to 15
UPDATE agent_definitions SET default_config = jsonb_set(..., 'max_items', '15') WHERE type = 'feed-triage';

-- Triage max_tokens raised to 8192
UPDATE agent_definitions SET default_config = jsonb_set(..., 'max_tokens', '8192') WHERE type = 'feed-triage';

-- Triage scores_field changed to "scores.result"
UPDATE agent_definitions SET default_config = jsonb_set(..., 'scores_field', '"scores.result"') WHERE type = 'feed-triage';

-- News added to primary nav (header) at position 4
INSERT INTO site_nav_items (...) VALUES (..., 'News', '/news.html', ...);

-- Growth config for gaswholesalers (5 content, 3 blog, 3 structural)
UPDATE site_specs SET data = '{"initial_target":12,"weekly_content_pages_max":5,...}' WHERE ...;

-- missing_news_page and unlinked_page_components re-added to completeness-discovery-agent checks
-- (These keep falling off — may need investigation into what removes them)
```

---

## Monitoring Queries

```sql
-- Triage backlog progress (check if ready to reset to 6h)
SELECT status, COUNT(*) FROM content_feed_items
WHERE site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
GROUP BY status;

-- When backlog cleared, reset interval:
-- UPDATE scheduled_tasks SET interval_seconds = 21600 WHERE name = 'content-feed-refresh';

-- Component linking after save_page_sections deploy
SELECT pc.position, pc.slot_name, pc.component_id IS NOT NULL as linked
FROM page_components pc
WHERE pc.page_id = '4ff0e0ff-fab2-423e-a59c-b9de4674a84f'
ORDER BY pc.position;

-- Credibility distribution
SELECT credibility, COUNT(*)
FROM content_feed_items
WHERE site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND credibility IS NOT NULL
GROUP BY credibility;
```

---

## Known Issues Remaining

| Issue | Priority | Notes |
|-------|----------|-------|
| `save_page_sections` component_id | P1 | File ready, deploy + verify |
| news-listing template missing `data-component` | P1 | SQL fix above, do after Go deploy |
| content-feed-refresh scheduler routing | P2 | Trigger never runs via scheduler; manual kcat workaround |
| Checks keep falling off completeness-discovery-agent | P2 | `missing_news_page` and `unlinked_page_components` had to be re-added; something may be overwriting the checks list |
| Header nav missing News link | P3 | Footer has it, header rerender may not have included it |
| Improvement-sweep site starvation | P3 | Oldest `updated_at` always wins; sites with frequent rebuilds dominate |

---

## Key Schema Notes (mistakes to avoid)

- `site_specs` has no `updated_at` column (only `created_at`, `is_current`, `superseded_at`)
- `content_feed_items` has no `updated_at` column (only `created_at`, `processed_at`)
- `pages` has no `rendered_html` column (has `rendered_head`)
- `content_components` template column is `html_template`, not `template_html`
- `agent_error_log` has no `stack_trace` column (has `context` JSONB)
- `site_specs` INSERT requires `created_by` (NOT NULL)
- `site_nav_items` has no `in_header`/`in_footer` columns — nav is via `group_id` → `site_nav_groups`
- `logger.Debug` is invisible in production — always use `logger.Info`
- `inputs.GetInt` reads from `ai.Values` (collectedData), NOT from config literals — use `datahelpers.GetIntField(params.StepConfig.Config, ...)` for config numbers
