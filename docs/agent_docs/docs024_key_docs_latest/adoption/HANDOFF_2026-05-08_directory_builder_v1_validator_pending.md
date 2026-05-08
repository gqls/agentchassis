# HANDOFF — Directory-Builder v1 / Tier D, mid-build state

**Date:** 2026-05-08
**Author session:** Session 4 of agentchassis directory-builder work (Sessions 1-3 in the prior HANDOFF doc).
**Pipeline stage:** Mid-build — chassis deployed at v1.0.994 with Go changes 1-3 of 4. Component-creator is now reaching the pre-store validation step, which is the new failure point. One Go change pending build/deploy.

---

## Where we are

The goal of this work (per the focus doc `FOCUS_directory_builder_and_list_components.md`) is to make list-shaped components (tool-list, blog-listing, directory-grid, etc.) draw their items from a real database query instead of having the LLM fabricate them. This requires a Tier D field-classification in the component-creator prompt, a `query.*` source resolver, and validation that understands the new array+items schema shape.

We picked **tool-list** as the smallest viable end-to-end slice. Once it regenerates correctly with Tier D, we re-adopt gamesdesign.co.uk and verify the deployed `/tools/index.html` shows real tool entries from the pages table rather than fabricated ones.

The slice is roughly 80% landed. The current blocker is that the pre-store validation in `compute_component_quality.go::extractSchemaFields` doesn't understand array sub-schemas — when the LLM produces a Tier D template with `{{range .items}}{{.title}}{{end}}`, the validator sees `.title` as an "orphan template variable with no schema entry" because it only walks top-level schema fields, not the array's `items` sub-object.

A patched `compute_component_quality.go` is staged in `/mnt/user-data/outputs/` ready for a chassis build. Once deployed, the work item resets and the regen should complete.

---

## What landed in this session

### Migrations (all run, all verified)

| # | File | What it did |
|---|---|---|
| 038 | `038_backups_pre_directory_builder.sql` | Two per-row backups: agent_definitions (component-creator) and content_components (tool-list). Postgres truncated table names to 63 chars (NAMEDATALEN limit) — the actual stored names dropped the trailing letters of `_pre_directory_builder` to `_pre_directory_build` for the agent backup and `_pre_directory_buil` for the component backup. |
| 039 | `039_component_creator_prompt_tier_d.sql` | First attempt at adding Tier D block. Pre-output check (a) extension landed cleanly; **the Tier D block insertion silently failed** — anchor used 4-space indent, actual prompt uses 3+5 spaces. Reported `UPDATE 1` because `replace()` returns the unchanged string when no match is found. Verification caught it via the `has_tier_d` row. |
| 039b | `039b_tier_d_insertion_retry.sql` | Retry of the Tier D block insertion with byte-accurate 3-space + 5-space anchor (confirmed via hex dump). User ran this multiple times; resulted in **three duplicate copies of the Exception paragraph** in pre-output check (a) — see 039c. |
| 039c | `039c_dedupe_exception_paragraph.sql` | One `regexp_replace` with a backreference, collapsing N consecutive copies to 1. Verified end state has exactly 1 copy. |
| 039d | `039d_tier_d_use_funcmap_helpers.sql` | Rewrote the Tier D block to use the new funcMap helpers (`{{rangeStart "items"}}`, `{{placeholder "title"}}`, `{{rangeEnd}}`) instead of bare `{{range .items}}`, `{{.title}}`, `{{end}}`. Fixed the "missing value for range at line 102" parse error. Reset the failed work item. |
| 039e | `039e_exception_paragraph_use_funcmap_helpers.sql` | Same fix applied to the Exception paragraph in pre-output check (a) — it had the same bare-Go-template-literal pattern, which trips at line 215 once Tier D is past. Reset the work item again. |
| 040 (part 1) | `040_tool_list_rename_and_regen.sql` (rename only) | Renamed `tool-list_pre_037` → `tool-list` (cosmetic clean-up). Part 2 (work item INSERT) failed due to my schema mistake — see 040b. |
| 040b | `040b_tool_list_regen_work_item.sql` | Emitted the regen work item with correct schema (`priority INTEGER`, `severity TEXT`, etc.), DELETE+INSERT pattern instead of ON CONFLICT, DO block pre-flight (replacing the `(SELECT 1/0)` CASE pattern that didn't short-circuit). |

### Go code (deployed in chassis v1.0.994)

| File | Change |
|---|---|
| `platform/orchestration/actions/queryresolve/queryresolve.go` (NEW) | New package. Single entry point `Resolve(ctx, db, QueryRequest, logger)`. Currently handles `pages_where_type:<type>`. Hard cap of 24 items, default 12. |
| `platform/orchestration/actions/plan_sections_action.go` | Imports `queryresolve`; new `query.*` handling block in field-resolution loop runs the resolver and puts results into `resolvedData[fieldName]`; existing `case "query"` in `resolveSource` reduced to defensive `(nil, true)` return with a comment. |
| `platform/orchestration/datahelpers/data_helpers.go` | Added `templateRangeStart` and `templateRangeEnd` to the funcMap as `rangeStart` and `rangeEnd`. These let prompts illustrate `{{range .X}}` and `{{end}}` syntax for the LLM without the prompt template engine itself trying to execute them. |

### Go code (PENDING build & deploy)

| File | Change |
|---|---|
| `platform/orchestration/actions/compute_component_quality.go` | `extractSchemaFields` now traverses array sub-schemas. When a field has `type: "array"` with an `items` sub-object, the sub-object's keys are added to the returned schemaFields map. This fixes the pre-store validation rejection on Tier D templates. **This is the only Go change still pending build.** |

Diff staged at `/mnt/user-data/outputs/compute_component_quality.diff`. Full file at `/mnt/user-data/outputs/compute_component_quality.go`.

### Documents

| File | Change |
|---|---|
| `FOCUS_directory_builder_and_list_components.md` | Updated source-types table (query.{name} now resolves at plan_sections time, not render time). New "Contract change" section documenting the why. Implementation strategy section rewritten to reflect the **hybrid** chosen (resolver-as-package now, dedicated agent later). |
| `003_query_plantime.diff` | One-line patch to doc 003 line 669 — change "At render time" to "At plan_sections time" for `query.{name}`. Apply with `patch -p0` from the repo root or edit by hand. |
| `FOCUS_chrome_templates_and_page_shape.md` (NEW) | Captures the C-priority work for after the directory-builder. Two distinct fixes: (1) header/footer templates must use `{{.nav_items_html}}` etc. via algorithmic enforcement (validation gate at store time), (2) adoption + planner must produce same-shape pages via the canonicaliser. **Deferred until tool-list verifies.** |

---

## Current state of the slice

### Chassis

- Image: `docker.io/aqls/agent-chassis:v1.0.994`
- Pods running: 3, all on v1.0.994 (verified via `kubectl ... -o jsonpath='{...image}'`)
- Has: queryresolve package, plan_sections patch, data_helpers funcMap helpers
- **Lacks**: extractSchemaFields array-sub-schema traversal (the v1.0.995 build)

### Work item

- ID: `e88fa510-87f5-4a32-8b9d-68188dc7ebe3`
- item_key: `needs_component:tool-list`
- handler_agent: `component-creator`
- Last status: `claimed` then `failed` with attempt_count=2 (at 14:41 UTC)
- Last error: `step store_component failed: ... template variables and schema fields do not match`

### tool-list component row

- ID: `a68b52b7-61c5-4797-a701-8e8643684f75`
- name: `tool-list` (canonical, post-rename)
- function: `tool-list`
- is_active: `true`
- updated_at: `2026-05-07 16:45` — i.e. STILL the pre-Tier-D shape (50 fields like `tool_1_name`, `tool_2_name`, etc.; html_template has 6 hardcoded card slots; no array+items)
- The new chassis will overwrite this row UPDATE-in-place when regen succeeds, preserving `id` so any references stay valid.

### Backups

Both 038 backups are in place:
- `agent_def_component_creator_backup_20260507_pre_directory_build` (1 row)
- `content_components_backup_tool_list_20260507_pre_directory_buil` (1 row)

If the regen produces something worse than the pre-session state, restore with:

```sql
-- Restore agent prompt
UPDATE agent_definitions
SET default_config = (
  SELECT default_config
  FROM agent_def_component_creator_backup_20260507_pre_directory_build
  LIMIT 1
), updated_at = NOW()
WHERE id = '23720180-7a39-4e3d-92e1-ebdbf95b57f4';

-- Restore tool-list component
UPDATE content_components
SET html_template = (SELECT html_template FROM content_components_backup_tool_list_20260507_pre_directory_buil LIMIT 1),
    input_schema  = (SELECT input_schema  FROM content_components_backup_tool_list_20260507_pre_directory_buil LIMIT 1),
    updated_at = NOW()
WHERE function = 'tool-list' AND is_active = true;
```

---

## Next actions (in order)

### 1. Build chassis v1.0.995 with the validator fix

The only Go change is `compute_component_quality.go` — adds array sub-schema traversal in `extractSchemaFields`. Diff is small (~30 lines added).

The other Go changes are already in v1.0.994 and don't need rebuilding.

### 2. Deploy v1.0.995

```bash
kubectl -n ai-persona-system rollout restart deployment/agent-chassis
kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[*].spec.containers[*].image}{"\n"}'
```

Confirm all pods on v1.0.995.

### 3. Reset the failed work item

```sql
UPDATE site_work_items
SET status        = 'triaged',
    attempt_count = 0,
    error         = NULL,
    claimed_by    = NULL,
    claimed_at    = NULL,
    completed_at  = NULL,
    result        = '{}'::jsonb,
    source        = source || '+post_validator_fix_reset',
    updated_at    = NOW()
WHERE item_key = 'needs_component:tool-list'
  AND item_type = 'needs_component_regeneration'
  AND status = 'failed';
```

(This isn't a numbered migration; it's a one-line reset.)

### 4. Wait ~1-2 minutes, then verify the regenerated tool-list

```sql
SELECT id, name, function, is_active,
       (input_schema -> 'fields') ? 'items' AS schema_has_items_key,
       jsonb_pretty(input_schema -> 'fields' -> 'items') AS items_field,
       html_template LIKE '%{{range .items}}%' AS template_has_range,
       html_template LIKE '%{{end}}%'         AS template_has_end,
       LENGTH(html_template) AS template_len,
       updated_at
FROM content_components
WHERE function = 'tool-list';
```

**Expect to see:**
- `schema_has_items_key`: `t`
- `items_field`: an object with `type: "array"`, `source: "query.pages_where_type:tool"`, and an `items` sub-object containing field definitions
- `template_has_range`: `t`
- `template_has_end`: `t`
- `updated_at`: post-deploy timestamp

If those check out, the regeneration worked.

### 5. Re-adopt gamesdesign.co.uk

Use the existing adoption flow. Site ID for the current adoption: `2524997b-be2a-4a38-931d-ac068e109233`.

Once adoption + plan + build complete, fetch the deployed `/tools/index.html` and confirm the tool list contains the real tool pages from the pages table (drop-rate-simulator, ehp-calculator, jump-physics, lanchester-combat-calculator, etc.) — not fabricated entries.

### 6. If verified, broaden to other list components

Other formerly-truncated list components: game-list, guide-list, blog-listing, case-studies-grid. Pattern: another migration emitting `needs_component_regeneration` items for each, the new prompt produces Tier D shapes, the queryresolve package handles their queries (the resolver currently accepts any `pages_where_type:<X>` argument, so blog_post / guide / game / entity_page work without code changes).

---

## What we deferred

### C — Chrome templates and page-shape canonicalisation

`FOCUS_chrome_templates_and_page_shape.md` captures this fully. Two fixes:

1. **Algorithmic enforcement** of the "headers/footers must use `{{.nav_items_html}}` etc." contract, via a validation gate in `store_generated_component_action.go` that rejects header/footer templates with hardcoded internal `<a href="/foo.html">` links not inside any `{{range}}` block.

2. **Page-shape canonicalisation at the source.** `apply_adoption_plan_action.go` lines 52169-52176 hand-builds page URLs (`/tools/<name>.html`) instead of going through the canonicaliser; the planner-emitted shape goes through. Result: `games` (content) AND `games-index` (entity_directory) BOTH deployed for the same logical page. Fix: adoption calls the canonicaliser; reconciler prunes orphan pages absent from the latest plan.

User's instruction: do these AFTER the directory-builder verifies. No work in this session.

### D — Other quality issues observed on gamesdesign

- Dead BEM design system in `<head>` (no theme CSS variables landing)
- `built_from_plan_version` backfill needed
- Component-creator config audit (some fields may be redundant)

All deferred.

### E — The 41 stuck `needs_section_data` items

Across 6 sites: leopardessconsulting:17, gaswholesalers:6, robot-hands:6, ai-agent-orchestration:5, finetuning:4, vonc:3. Per the prior FOCUS doc these are mostly "component not found" cases that should have been emitted as `needs_component:*` items. Cleanup deferred until tool-list verifies.

---

## Gotchas worth remembering

### `replace()` and prompt anchors

The component-creator prompt has irregular indentation (3+5 spaces, not consistent). Postgres `replace()` does literal-byte matching — silently fails to a no-op when the anchor doesn't match. Always:

1. **Verify anchor bytes via hex dump** before writing a replace migration:
   ```sql
   SELECT ENCODE(CONVERT_TO(SUBSTRING(... FROM ... FOR ...), 'UTF8'), 'hex')
   ```
2. **Test verification matches presence with COUNT, not just LIKE.** A `LIKE '%foo%'` test passes if the substring is anywhere in the prompt; if a prior bad replace produced duplicates, you won't notice.
3. **Prefer shorter, lexically distinctive anchors** when possible. Multi-line whitespace-heavy anchors are fragile.

### `replace()` with re-entry

If a migration's anchor matches the original text but the replacement contains the original text again (e.g. anchor on a paragraph, replacement appends a new paragraph after the same text), running the migration twice appends twice. **Always verify count, not just presence.** 039b ran 3 times producing 3 Exception paragraphs.

### CASE expressions don't short-circuit sub-SELECTs

```sql
CASE WHEN COUNT(*) = 0 THEN (SELECT 1/0) ELSE 1 END
```

Postgres evaluates the sub-SELECT eagerly during planning even when the WHEN condition is false. Use DO blocks with `RAISE EXCEPTION` instead.

### Postgres NAMEDATALEN

Backup table names truncate at 63 chars. Keep names short or accept the truncation and reference the truncated name in subsequent migrations.

### `site_work_items` schema field types (easy to confuse)

- `priority` is `INTEGER`, default 100. Lower number = higher priority.
- `severity` is `TEXT`, default `'medium'`. Values: `low`, `medium`, `high`.
- `summary` is `NOT NULL`. Always provide.
- `created_by` is `NOT NULL`. Always provide.
- `pipeline` is `NOT NULL`, default `'build'`.
- The dedup index `idx_swi_dedup` is partial on (site_id, item_key) WHERE status not in (complete, verified, rejected, wont_fix, failed). Use DELETE+INSERT for clarity rather than ON CONFLICT WHERE.

### Prompt templates and Go's `text/template`

The component-creator's prompt itself goes through `RenderPromptTemplate` (`text/template` parse + execute) before being sent to the LLM. **Any literal `{{...}}` in the prompt will be parsed as a template action.** Use the funcMap helpers:

- `{{placeholder "X"}}` → renders to literal `{{.X}}` for the LLM to copy
- `{{rangeStart "X"}}` → renders to literal `{{range .X}}`
- `{{rangeEnd}}` → renders to literal `{{end}}`

If you write a prose mention of `{{.X}}` or `{{range}}` in the prompt without going through one of these helpers, the parser will explode.

### The pre-store validation has its own logic

The prompt's pre-output self-check (a) tells the LLM how to count fields. The Go pre-store validator in `store_generated_component_action.go` runs `scoreComponent` which has its own counting logic. **Updating one without the other breaks Tier D.** The prompt change in 039 was followed by the validator fix in `compute_component_quality.go` — these are two halves of the same change, both needed.

---

## Files staged for next session

In `/mnt/user-data/outputs/`:

**Pending build:**
- `compute_component_quality.diff` — the validator fix
- `compute_component_quality.go` — full file with patch

**Already deployed (v1.0.994, for reference):**
- `queryresolve/queryresolve.go`
- `plan_sections_action.go` + `plan_sections_action.diff`
- `data_helpers.go` + `data_helpers.diff`

**Already run (migrations 038-040b):**
- `038_backups_pre_directory_builder.sql`
- `039_component_creator_prompt_tier_d.sql`
- `039b_tier_d_insertion_retry.sql`
- `039c_dedupe_exception_paragraph.sql`
- `039d_tier_d_use_funcmap_helpers.sql`
- `039e_exception_paragraph_use_funcmap_helpers.sql`
- `040_tool_list_rename_and_regen.sql` (part 1 only — part 2 superseded by 040b)
- `040b_tool_list_regen_work_item.sql`

**Documents:**
- `FOCUS_directory_builder_and_list_components.md` (current)
- `FOCUS_chrome_templates_and_page_shape.md` (deferred, captured)
- `003_query_plantime.diff` (apply to doc 003)

---

## Key identifiers

- **Component-creator agent**: `23720180-7a39-4e3d-92e1-ebdbf95b57f4`
- **system.internal site**: `eac60db8-b032-432b-b36d-76f37632045d`
- **gamesdesign.co.uk current adoption**: `2524997b-be2a-4a38-931d-ac068e109233`
- **tool-list component**: `a68b52b7-61c5-4797-a701-8e8643684f75`
- **Current regen work item**: `e88fa510-87f5-4a32-8b9d-68188dc7ebe3`
- **Chassis image (deployed)**: `docker.io/aqls/agent-chassis:v1.0.994`
- **Chassis image (pending build)**: `docker.io/aqls/agent-chassis:v1.0.995` (suggested)
