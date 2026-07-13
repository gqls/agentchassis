# Handoff: Component Linking Triage, News-Listing Template, Discovery Checks Investigation

**Date**: 2026-04-19
**Continuation of**: `HANDOFF_2026-04-17_triage_and_component_linking.md`
**Test site**: gaswholesalers.com (`5fe15466-4e2e-4ff2-981e-98c1b7074002`)
**Homepage page_id**: `4ff0e0ff-fab2-423e-a59c-b9de4674a84f`

---

## Session Summary

Worked through the P1/P2 list from the 04-17 handoff. Two items closed (one as a non-issue), one parked, one prepared but not executed.

| Issue | 04-17 state | 04-19 state |
|---|---|---|
| `save_page_sections` component_id | P1 — file ready, not deployed | **Parked** — code is on production but not linking. Investigation overlaps with another chat. |
| news-listing template missing `data-component` | P1 — SQL written in handoff | **SQL verified, transaction ready, not yet run** |
| Checks falling off completeness-discovery-agent | P2 — keeps needing re-adding | **Closed** — investigation found no overwriter |
| content-feed-refresh scheduler routing | P2 | Not yet started |
| Header nav missing News link | P3 | Not yet started |
| Improvement-sweep site starvation | P3 | Not yet started |
| Topic splitting | Feature work | Not yet started |

---

## 1. `save_page_sections` component_id — PARKED

### What the 04-17 handoff said
File ready in outputs folder (716 lines), not deployed. Enrichment runs in HTML-only block so metadata path wipes `component_id` every rebuild.

### What we found today
The fix **is** on production — user confirmed after seeing the extracted file from `production_agent-chassis-full_context.txt` lines 36397–37114. All four changes are present:
- `data-component` preferred over metadata `ComponentName`
- Suffix stripping (`-section`, `-container`, `-wrapper`, `-block`)
- Enrichment runs outside the HTML-only block (both metadata and HTML paths)
- All `logger.Debug` → `logger.Info`

**But linking still isn't working.** Diagnostic queries showed:

```
gaswholesalers homepage (page_id 4ff0e0ff-fab2-423e-a59c-b9de4674a84f):
  page_components last written 2026-04-14 20:58 — 3 days stale
  page.updated_at 2026-04-16 16:11 (page-level metadata, not component rewrite)
  All 7 rows: linked = false
  All 7 rows: html_data_component populated correctly
  All 7 rows: matching content_components.function exists (EXISTS = true)

Broader test — any page rebuilt since the fix was deployed:
  15 pages rebuilt after 2026-04-15, across multiple sites
  Every single one: linked_components = 0 / total_components
```

### What this means
Several possibilities, not yet narrowed:
1. Deployed binary doesn't contain the code we extracted (repo has it, running image is older)
2. Code runs but an early return before enrichment fires on these code paths
3. DB writes happen on a different path that skips enrichment
4. Enrichment links in memory but the INSERT drops `component_id`

### Next diagnostic (not run — overlaps with another chat)
```bash
# Pick any chassis pod
kubectl -n ai-persona-system get pods -o wide | grep -i chassis
POD=<name>

# Check for enrichment log lines in the last 3 hours
kubectl -n ai-persona-system logs $POD --since=3h \
  | grep -E "SavePageSectionsAction|enrichSectionsWithComponentIDs|enrichSectionsWithPlannedNames" \
  | tail -50

# And verify the binary actually has the fix strings
kubectl -n ai-persona-system exec $POD -- sh -c \
  'strings <path-to-agent-chassis> | grep -E "preferring data-component|enrichSectionsWithPlannedNames"'
```

Three log outcomes, three different fixes:
- **No `SavePageSectionsAction` lines** → components written by a different action; our fix is irrelevant to those rows
- **`SavePageSectionsAction` lines, no `enrichSectionsWithComponentIDs` lines** → enrichment skipped (DB nil, code path skipped)
- **`linked component` lines but DB still unlinked** → in-memory linking works, INSERT loses `component_id`

### Verification once resolved
```sql
SELECT pc.position, pc.slot_name, pc.component_id IS NOT NULL AS linked
FROM page_components pc
WHERE pc.page_id = '4ff0e0ff-fab2-423e-a59c-b9de4674a84f'
ORDER BY pc.position;
```
All 7 should be `true`.

---

## 2. news-listing template — SQL ready, NOT YET RUN

### Current template state (verified today)
One active variant:
- `id`: `11d4dc21-1ccc-40ef-93bc-b9e26bd95e9f`
- `function`: `news-listing`
- Template starts: `<section class="news-listing-section" id="news-listing">`
- **No `data-component` attribute** — this is why the news-listing slot never links

Exactly one `page_components` row uses it (gaswholesalers `/news.html`, position 2, unlinked).

### The transaction to run
```sql
BEGIN;

-- Preview
SELECT id, name,
       html_template LIKE '%data-component="news-listing"%' AS already_has_attr,
       LEFT(html_template, 200) AS before_snippet
FROM content_components
WHERE id = '11d4dc21-1ccc-40ef-93bc-b9e26bd95e9f';

-- Apply (idempotent via NOT LIKE guard)
UPDATE content_components
SET html_template = REPLACE(
    html_template,
    '<section class="news-listing-section"',
    '<section data-component="news-listing" class="news-listing-section"'
),
updated_at = now()
WHERE id = '11d4dc21-1ccc-40ef-93bc-b9e26bd95e9f'
  AND html_template NOT LIKE '%data-component="news-listing"%';

-- Verify
SELECT id, name,
       html_template LIKE '%data-component="news-listing"%' AS now_has_attr,
       LEFT(html_template, 200) AS after_snippet
FROM content_components
WHERE id = '11d4dc21-1ccc-40ef-93bc-b9e26bd95e9f';

-- If now_has_attr = true and after_snippet starts with
--   <section data-component="news-listing" class="news-listing-section" id="news-listing">
-- then: COMMIT;
-- otherwise: ROLLBACK;
```

### Note on existing row
The gaswholesalers `/news.html` position 2 row has HTML snapshotted before this template change. It will only pick up `data-component` after the page next re-renders through `SavePageSectionsAction` — **same dependency as item 1**. Both items are blocked on the same re-render path working.

### Schema notes used
- `content_components.html_template` is the correct column name (not `template_html`)
- Schema also has a `page_component_history` table — tracks `page_components`, not `content_components`, so no history concern for this template change

---

## 3. Checks falling off completeness-discovery-agent — CLOSED (not a real issue)

### What was suspected (04-17 handoff)
"Something may be overwriting the checks list" — `missing_news_page` and `unlinked_page_components` reportedly had to be re-added.

### What we found
No evidence of any overwriter. Full investigation:

**Live state intact:**
```
11 checks, updated 2026-04-17 15:44:
["all_sources_erroring", "empty_blog", "empty_sections",
 "missing_news_page", "missing_news_section", "missing_news_sources",
 "missing_structure", "orphan_pages", "stale_news_section",
 "unlinked_page_components", "unresolved_sections"]
```

**Backups show monotonic growth, not regression:**
| Snapshot | Check count | Missing vs live |
|---|---|---|
| `agent_definitions_backup_20260322` (23 Mar) | 3 | Pre-dated news checks, pre-dated news expansion |
| `agent_definitions_backup_20260326` (2 Apr) | 8 | Pre-dated `missing_news_page`, `unlinked_page_components`, `unresolved_sections` |
| Live (17 Apr) | 11 | — |

Each backup is a snapshot from **before** certain checks were added. There's no "after they fell off" snapshot anywhere.

**`agent_default_configs` ruled out:**
Separate table, one row only (`domain-analyst` model/temp/retries). Unrelated to workflow config.

**Only runtime path that could overwrite workflow:**
`updateAgentWorkflow()` at context line 61056 uses `jsonb_set(default_config, '{workflow}', $1)` which replaces the entire workflow subtree. Called only from `applyChange("modify_workflow")`, which runs only from `ApproveImprovementAction` (human approval required).

**`improvement_proposals` table is empty:**
```
SELECT status, COUNT(*) FROM improvement_proposals GROUP BY status;
→ 0 rows
```
Never had a proposal. The modify_workflow path has never executed.

### Likely cause of the original "falling off" experience
A past manual SQL edit used an incomplete `UPDATE` that replaced the full checks array rather than appending to it. The most plausible culprit: a **stale comment** in the discovery_checks Go source at chassis context line ~43070 showing an 8-item `UPDATE` example. Someone running that verbatim would drop the 3 newest checks.

### Recommended small cleanup
Replace the stale comment header in the discovery_checks file (and similar comments in `design-discovery-agent` / `quality-discovery-agent` if present) with a safer pattern that uses array append rather than replacement:

```go
// The live checks array is maintained in the DB via manual SQL.
// To inspect current list:
//   SELECT default_config->'workflow'->'steps'->'run_checks'->'config'->'checks'
//   FROM agent_definitions WHERE type = 'completeness-discovery-agent' AND deleted_at IS NULL;
//
// To add a new check, use jsonb array append (do NOT replace the array):
//   UPDATE agent_definitions
//   SET default_config = jsonb_set(
//       default_config,
//       '{workflow,steps,run_checks,config,checks}',
//       (default_config->'workflow'->'steps'->'run_checks'->'config'->'checks')
//         || '["new_check_name"]'::jsonb
//   )
//   WHERE type = 'completeness-discovery-agent' AND deleted_at IS NULL;
```

The `||` pattern is additive — it can't accidentally shorten the list.

---

## 4. Latent risk logged for future work

`updateAgentWorkflow` at chassis context line 61056:
```go
UPDATE agent_definitions
SET default_config = jsonb_set(default_config, '{workflow}', $1::jsonb),
    updated_at = NOW()
WHERE type = $2 AND is_active = true
```

Currently safe because nothing fires it. When an automated improvement-proposal generator is eventually wired up, any proposal containing a partial `workflow` object will silently erase the rest of the workflow tree (all steps, configs, handlers). The function should be changed to a deep-merge pattern before that generator ships, or proposal authors constrained to edit specific step paths rather than the whole workflow.

Not fixing now — no caller exists yet. Noting so it doesn't become a surprise later.

---

## Open Items

| Issue | Priority | State |
|---|---|---|
| `save_page_sections` component_id not linking despite fix being deployed | P1 | Parked — overlaps other chat. Diagnostic plan documented above. |
| news-listing template `data-component` | P1 | Transaction ready to run. Blocked on user executing. |
| content-feed-refresh scheduler routing | P2 | Trigger never runs via scheduler; manual kcat workaround in use. Two possible fixes sketched in 04-17 handoff. |
| Header nav missing News link | P3 | Footer has it, header re-render didn't include it. |
| Improvement-sweep site starvation | P3 | Oldest `updated_at` site always wins. |
| Topic splitting (split single Grok source into geopolitical / regulatory / pricing / transition) | Feature | SQL-only change, no Go needed. Plan in `031_news_content_diversity_plan.md`. |
| Stale discovery_checks comments showing old `UPDATE` patterns | Cleanup | Low priority cosmetic / copy-paste hazard fix. |

---

## Schema Notes Confirmed This Session

- `content_components.html_template` (not `template_html`) — confirmed against schema dump
- `improvement_proposals` columns: `target_type`, `target_id`, `proposed_changes`, `applied_changes` (not `agent_type`, `changes`)
- `agent_default_configs` exists but is unrelated to workflow config (model/retries/temperature only, keyed by `config_name` + `agent_type` + `environment`)
- `page_component_history` exists as an audit table for `page_components` changes — not triggered by `content_components` updates

---

## Monitoring Queries

```sql
-- Component linking rate across recent rebuilds (cf. item 1 diagnostic)
SELECT
    p.site_id,
    p.id AS page_id,
    p.name,
    MAX(pc.updated_at) AS last_rebuild,
    COUNT(*) AS total_components,
    COUNT(pc.component_id) AS linked_components
FROM page_components pc
JOIN pages p ON p.id = pc.page_id
WHERE pc.updated_at > now() - interval '7 days'
GROUP BY p.site_id, p.id, p.name
HAVING COUNT(*) > COUNT(pc.component_id)  -- only pages with unlinked rows
ORDER BY last_rebuild DESC
LIMIT 20;

-- Checks list on each discovery agent — verify nothing has shortened
SELECT
    type,
    jsonb_array_length(default_config->'workflow'->'steps'->'run_checks'->'config'->'checks') AS check_count,
    default_config->'workflow'->'steps'->'run_checks'->'config'->'checks' AS checks,
    updated_at
FROM agent_definitions
WHERE type IN ('completeness-discovery-agent', 'design-discovery-agent', 'quality-discovery-agent')
  AND deleted_at IS NULL
ORDER BY type;

-- Triage backlog progress (from 04-17 handoff — check if ready to reset to 6h)
SELECT status, COUNT(*) FROM content_feed_items
WHERE site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
GROUP BY status;
-- When ingested drops below 30:
--   UPDATE scheduled_tasks SET interval_seconds = 21600 WHERE name = 'content-feed-refresh';
```
