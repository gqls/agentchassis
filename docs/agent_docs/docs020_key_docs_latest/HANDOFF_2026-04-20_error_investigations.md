# Handoff — Error Investigations (post-Track-2)

**Date:** 2026-04-20
**Session end state:** Track 2 (component regeneration) complete and verified.
This handoff inventories open error threads surfaced by a queue + recent-
errors scan. Bring a fresh set of code files to a new chat to work through
these.

---

## What to bring into the next chat

1. `/mnt/project/production_agent-chassis-full_context.txt` (always, for code lookup)
2. `/mnt/project/bk_agent_definitions_backup.sql` (for agent workflow lookups)
3. Fresh copies of whatever action files match the investigations below.
   Specific paths are called out per investigation.
4. Any live DB diagnostics requested (run against clients_db, paste output).

---

## Queue snapshot (2026-04-20)

```
25 distinct (domain, item_type, status) buckets currently active.
```

### Distribution

| Status | Count |
|---|---|
| needs_human_review | 22 |
| unresolved | 5 |
| failed | 5 |
| claimed | 1 |

### By item_type

| Type | Count | Notes |
|---|---|---|
| needs_section_data | 19 | All human-gated — awaiting real client data |
| content_rewrite | 6 | 4 human-review (validation-blocked), 0 active |
| empty_section | 5 | 3 unresolved, 2 human-review |
| improve_tool | 2 | Both failed — old `component_id` resolution bug (was P1 in 105) |
| needs_composition | 2 | NEW — failed 2026-04-20 on gamedesign.uk |
| orphan_blog_posts | 2 | Unresolved on leopardess + robot-hands |
| needs_design | 1 | Claimed, stuck on gamedesign.uk (timeout) |
| needs_content_page | 2 | Human-review |
| needs_new_layout_candidate | 1 | Human-review |
| audit_tool | 1 | Failed on gaswholesalers |

---

## Error frequency (last 7 days)

| Occurrences | Sites | Error preview |
|---|---|---|
| **66** | **6** | Auto-completed: work verified done despite lost response |
| **47** | **7** | Claim timed out — handler pod likely died |
| 12 | 4 | step validate_content failed: validation failed: 1 blockers, 0 errors |
| 2 | 1 | **install_site_composition: column "industry_tags" is text[] but expression is ...** (fresh 2026-04-20) |
| 2 | 2 | load_tool: query_database resolved to nil (improve_tool items, old bug) |
| 1 | 1 | store_component: provocation-feed empty input_schema (fixed, historical) |
| 1 | 1 | validate_page_content: 0 blockers, 1 errors |
| 1 | 1 | Claim timed out (attempts exhausted) |
| 1 | 1 | store_component: gauntlet-interface empty input_schema (fixed, historical) |
| 1 | 1 | weekly_content_limit_reached (feature, not bug) |

---

## Investigation 1 — `install_site_composition` schema mismatch

**Priority: HIGH** — fresh failure today, blocking gamedesign.uk's new composable theme install.

### The error

```
step install_site_composition failed:
  failed to execute action install_site_composition:
  insert style_collections:
  ERROR: column "industry_tags" is of type text[] but expression is [truncated]
```

### Context

Part of the **composable theme migration** (doc `025_palette_layout_typography_migration_1_.md`). `install_site_composition` is a new action that installs a composed theme (palette + layout + typography) for a site, creating a `style_collections` row.

### Hypothesis

The action is passing a `jsonb` array or a `text` value where `style_collections.industry_tags` expects `text[]`. Classic Postgres array-cast mismatch.

### Diagnostic steps

```sql
-- Confirm the column type
\d style_collections

-- See the last attempted insert via error details
SELECT id, error, spec, updated_at
FROM site_work_items
WHERE item_type = 'needs_composition'
  AND error LIKE '%industry_tags%'
ORDER BY updated_at DESC
LIMIT 5;

-- Any recent successful inserts?
SELECT id, name, industry_tags, created_at
FROM style_collections
ORDER BY created_at DESC
LIMIT 5;
```

### Where the action lives

Search `production_agent-chassis-full_context.txt` for `InstallSiteComposition` or `install_site_composition`. The fix is likely:
- Convert Go `[]string` to `pq.StringArray` (or pgx equivalent) before passing to `ExecContext`
- OR use `unnest($1::text[])` or `ARRAY[...]::text[]` cast in SQL

### Recovery

After fix deployed, reset the 2 failed gamedesign.uk work items:

```sql
UPDATE site_work_items
SET status = 'triaged', attempt_count = 0, claimed_by = NULL,
    claimed_at = NULL, error = NULL, updated_at = NOW()
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk')
  AND item_type = 'needs_composition'
  AND status = 'failed';
```

---

## Investigation 2 — 66× "Auto-completed: work verified done despite lost response"

**Priority: HIGH** — highest-frequency error, 6 sites affected.

### Hypothesis

This is an orchestration-layer message (not an action failure). Likely raised by the chassis when a child agent's response message is lost (kafka timeout, pod crash, topic mismatch) but the child's DB state shows the work succeeded. The dispatcher decides to mark the work item complete anyway rather than retry.

Two possibilities:

**(a) Benign recovery** — the child did the work, the parent just didn't get the reply. DB state consistent. Marking complete is correct.

**(b) False recovery** — the child crashed mid-way or the DB state isn't actually consistent. Marking complete hides the failure.

### Diagnostic steps

```sql
-- Which work item types hit this?
SELECT item_type, handler_agent, COUNT(*) AS occurrences,
       array_agg(DISTINCT (SELECT domain FROM sites WHERE id = site_id)) AS domains
FROM site_work_items
WHERE error = 'Auto-completed: work verified done despite lost response'
  AND updated_at >= NOW() - INTERVAL '7 days'
GROUP BY item_type, handler_agent
ORDER BY occurrences DESC;

-- Were the results actually delivered? Check a sample of "auto-completed"
-- items — did the page deploy, component create, etc.?
SELECT swi.id, swi.item_type, swi.spec, swi.result, swi.updated_at,
       s.domain
FROM site_work_items swi
JOIN sites s ON s.id = swi.site_id
WHERE swi.error = 'Auto-completed: work verified done despite lost response'
ORDER BY swi.updated_at DESC
LIMIT 10;
```

### Where the message is raised

Search `production_agent-chassis-full_context.txt` for `Auto-completed` or `despite lost response`. Understand the conditions under which the orchestrator promotes `claimed` → `complete` without a response message.

### Open questions for next session

- Is this recovery too aggressive (marking false-positive completes)?
- Should "auto-completed" work items be visibly flagged for audit rather than silently treated as successful?
- Is the root cause pod instability, kafka message loss, or something else?

---

## Investigation 3 — 47× "Claim timed out — handler pod likely died"

**Priority: HIGH** — second-most-frequent error, 7 sites affected.

### Hypothesis

A handler agent claims a work item but the pod dies before marking it complete. The heartbeat/claim-expiry scan notices the claim is stale and releases it. Message text suggests the system knows why — pod died — but how does it know?

Possibilities:
- A TTL on `claimed_at` (X minutes)
- A pod health check that notices a missing container
- Kafka consumer group rebalance detecting a dead consumer

### Diagnostic

```sql
-- Which handlers die most often?
SELECT handler_agent, COUNT(*) AS timeouts,
       array_agg(DISTINCT (SELECT domain FROM sites WHERE id = site_id)) AS domains
FROM site_work_items
WHERE error LIKE '%Claim timed out%handler pod likely died%'
  AND updated_at >= NOW() - INTERVAL '7 days'
GROUP BY handler_agent
ORDER BY timeouts DESC;

-- Are these clustered in time (suggesting cluster-wide events) or spread out?
SELECT date_trunc('hour', updated_at) AS hour,
       COUNT(*) AS timeouts
FROM site_work_items
WHERE error LIKE '%Claim timed out%handler pod likely died%'
  AND updated_at >= NOW() - INTERVAL '7 days'
GROUP BY 1 ORDER BY 1 DESC
LIMIT 20;
```

### Actions

- Check `kubectl -n ai-persona-system get pods` — any agents in CrashLoopBackOff?
- Check pod memory limits — agents with `memory: 512Mi` may OOM on LLM-heavy workflows
- `webdesign-agent` is one we know hit this (claim b8951daa on gamedesign.uk)

### Where to look

In chassis: find the heartbeat loop that promotes stale claims back to triaged or marks them failed. Understand its definition of "stale" and whether it's too aggressive (healthy pods running long LLM calls) or too lenient (dead pods holding claims for ages).

---

## Investigation 4 — 12× content_validation blockers

**Priority: MEDIUM** — moderate frequency, affects 4 sites.

### The error

```
step validate_content failed:
  failed to execute action validate_page_content:
  content validation failed: 1 blockers, 0 errors
```

### Known categories (from handoff docs)

- Placeholder emails
- Unsubstantiated claims
- Empty pages
- Cross-site contamination

### Diagnostic

```sql
-- What content_rewrite / content page items hit validation blockers?
SELECT swi.id, swi.item_type, swi.page_id, 
       left(swi.result::text, 400) AS result_preview,
       s.domain,
       swi.updated_at
FROM site_work_items swi
JOIN sites s ON s.id = swi.site_id
WHERE swi.error LIKE '%content validation failed%blockers%'
ORDER BY swi.updated_at DESC
LIMIT 10;
```

### Status

These are currently `needs_human_review`, which is the correct path. Question: do we need better LLM prompts to catch the blocker classes BEFORE writing, or additional auto-repair attempts before handing to humans?

Likely defer until after the higher-frequency items (1, 2, 3).

---

## Investigation 5 — 2× improve_tool `load_tool` nil

**Priority: LOW** — old bug, already reported in 105. 2 occurrences in 7 days suggests the fix from P1 may not have fully landed, or there's a residue of pre-fix items still retrying.

### Diagnostic

```sql
-- Look at these specific items
SELECT id, site_id, spec, error, attempt_count, updated_at
FROM site_work_items
WHERE item_type = 'improve_tool'
  AND status = 'failed'
  AND error LIKE '%load_tool%resolved to nil%'
ORDER BY updated_at DESC;
```

If spec has `component_id` at top level — P1 fix is working, but these are pre-fix items that need resetting.
If spec has `component_id` only inside `spec.component_id` — P1 fix didn't fully land.

### Recovery

If P1 works, just reset:

```sql
UPDATE site_work_items
SET status = 'triaged', attempt_count = 0, error = NULL, updated_at = NOW()
WHERE item_type = 'improve_tool' AND status = 'failed'
  AND error LIKE '%load_tool%resolved to nil%';
```

---

## Investigation 6 — orphan_blog_posts (unresolved on 2 sites)

**Priority: LOW** — was addressed in 103_blog_nav_handoff. Check current state.

### Diagnostic

```sql
SELECT s.domain, swi.id, swi.status, swi.error, swi.updated_at
FROM site_work_items swi
JOIN sites s ON s.id = swi.site_id
WHERE swi.item_type = 'orphan_blog_posts'
  AND swi.status NOT IN ('complete', 'verified');

-- And check if the expected blog-listing slots actually populated
SELECT s.domain, p.name, pc.slot_name,
       LENGTH(pc.rendered_html) AS html_len
FROM pages p
JOIN sites s ON p.site_id = s.id
JOIN page_components pc ON pc.page_id = p.id
WHERE p.page_type = 'blog-index'
  AND s.domain IN ('leopardessconsulting.co.uk', 'robot-hands.com');
```

Expected: blog-listing slot populated with recent rendered_html. If empty → the `rebuild_blog_listing_action.go` rewrite didn't help these 2 sites and needs more investigation.

---

## Investigation 7 — nav duplicate header/footer residue

**Priority: MEDIUM** — dirty data persists, may regenerate issues.

### Finding

10 rows in `page_components` on **ai-agent-orchestration.com** with `slot_name` `header-professional-dark` or `footer-standard` and `component_id = NULL`, updated 2026-04-13/14.

The earlier nav handoff (2026-04-11) ran a cleanup but these rows appeared AFTER that cleanup. Either:
- Something is still creating these rows (plan_sections filter NOT deployed)
- A different content-rewrite or rebuild path reintroduces them

### Required code fix (from nav handoff, still needed)

**File:** `platform/orchestration/actions/plan_sections_action.go` (or wherever PlanSectionsAction lives)

Add a filter function that strips site-level component names from the sections list before passing to content writer:

```go
func filterSiteLevelSections(sections []string) []string {
    filtered := make([]string, 0, len(sections))
    for _, s := range sections {
        lower := strings.ToLower(s)
        if strings.Contains(lower, "header") ||
           strings.Contains(lower, "footer") ||
           lower == "site-header" ||
           lower == "site-footer" {
            continue
        }
        filtered = append(filtered, s)
    }
    return filtered
}
```

### Required data cleanup

After the code fix deploys, clean the dirty rows again:

```sql
BEGIN;
DELETE FROM page_components
WHERE (slot_name ILIKE '%header%' 
       OR slot_name ILIKE '%footer%'
       OR slot_name IN ('site-header', 'site-footer'))
  AND component_id IS NULL;
-- Then trigger rerender of affected pages
COMMIT;
```

### Also needed (from nav handoff)

- **InjectHeader/InjectFooter skip-if-present guard** — check if HTML already contains `class="site-header"` before injecting
- **component-template-fixer case-insensitivity** for responsive CSS idempotency check (`"responsive fix"` vs `"Responsive fix"`)

---

## Investigation 8 — `needs_design` claim stuck (mystery work item)

**Priority: LOW** — single instance, may self-resolve.

### Finding

Work item `b8951daa-f882-4bf3-92cd-d5d233b066af` on gamedesign.uk has been `claimed` by `build-dispatch-loop` since 10:57. Unusual timestamp: `updated_at = 10:56:15` (earlier than `claimed_at = 10:57:08`).

Could be:
- Legitimate in-flight work (webdesign-agent doing a long LLM call)
- Stuck claim (the pod may be dead but the timeout hasn't fired yet)

### Diagnostic

```sql
SELECT id, site_id, item_type, status, claimed_by, claimed_at,
       NOW() - claimed_at AS claim_age,
       attempt_count, max_attempts, spec, updated_at
FROM site_work_items
WHERE id = 'b8951daa-f882-4bf3-92cd-d5d233b066af';
```

If claim_age > 30 min and attempt_count > 0, the heartbeat loop should have released it. If not, tie to Investigation 3 (claim timeout mechanism).

---

## Investigation 9 — `empty_section` unresolved (3 items)

**Priority: LOW** — 3 `empty_section` items unresolved across robot-hands, finetuning, gaswholesalers (oldest updated 2026-04-13).

These are legitimately empty sections on deployed pages. Resolution is either:
- Content team fills them in (human-gated)
- Remove the section from the page plan
- Auto-regenerate if content is derivable

Probably needs to be categorised as human-gated and aged out via the heartbeat, not a code issue.

---

## Consolidated priority ordering for next session

1. **Investigation 1** — `install_site_composition` schema mismatch (fresh, blocking, small fix)
2. **Investigation 7** — nav plan_sections filter (structural, preventive, small)
3. **Investigation 2** — "Auto-completed despite lost response" (orchestration-layer, high-frequency)
4. **Investigation 3** — "Claim timed out" (orchestration-layer, high-frequency)
5. **Investigation 4** — content_validation blockers (medium, probably defer)
6. Everything else — individual triage or aged-out

---

## Historical context — what Track 2 closed

For reference when working through the above:
- Component prompt escape mechanism (`placeholder` funcMap) — deployed
- Layer 1 pre-store validation — deployed
- `StoreGeneratedComponentAction` regeneration path (create OR regenerate, not skip) — deployed
- `change_source` column on `component_versions` — migration applied
- `update_component_html_action` column rename (version_note → change_description) and rendered_html corruption fix — deployed
- Docs updated: 001, 003, 014, 020 + new 026
- Dispatcher source forwarding patch — staged, not yet applied: `patch_dispatcher_source_forwarding.sql`

The deployed `StoreGeneratedComponentAction` does NOT yet implement the "resurrect deactivated row" behaviour. If a deactivated row shares function/name, regen falls through to CREATE and hits `content_components_name_key` unique violation. Today's fix was an ad-hoc DELETE. Structural fix still pending.

---

## Schema notes (live, verified 2026-04-20)

- `sites` has NO `deleted_at` — use `status` field or omit filter
- `content_components.name` has a plain `UNIQUE` constraint (not partial) — collisions possible across `is_active` rows
- `component_versions.change_source` is new (April 2026), nullable text
- `site_work_items` columns: site_id, source, pipeline, item_type, severity, summary, priority, handler_agent, status, created_by, spec, item_key, page_id, component_id, depends_on, attempt_count, max_attempts, approval_mode, claimed_by, claimed_at, completed_at, result, error, updated_at, created_at

---

## Files staged in /mnt/user-data/outputs/ (recent)

Ready to apply / reference:
- `patch_dispatcher_source_forwarding.sql` — apply for change_source trace
- `store_generated_component_action.go` + `store_generated_regen.diff` — already deployed
- `update_component_html_action.go` + `update_component_html.diff` — already deployed
- `migration_component_versions_change_source.sql` — already applied
- `026_component_regeneration_flow.md` — merged into project
- Updated `001_development_guide.md`, `003_contracts_and_standards_v7.md`, `014_site_snapshots_and_revert.md`, `020_tool_lifecycle.md` — merged
