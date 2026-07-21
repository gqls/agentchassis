# PLAN — bug 025: wire up `pages.content_direction`

**Started 2026-07-21.** Owner chose fix candidate **(2) wire it up as documented**
(not delete). Bug file: `/bugs_open/025_HANDOFF_2026-07-19_content_direction_column_documents_behaviour_that_does_not_exist.md`.

## What "wire it up" means, verified against the LIVE system (not the seeds)

Setting `pages.content_direction` (jsonb) on a page must steer that page's copy on
(re)build. The value must reach the `page-content-writer` prompt as
`.current_page.content_direction`.

### Data-flow — how `current_page` is populated (checked in live agent_definitions)

There are two loaders and (at least) three caller paths. `current_page` is a whole
map/object; the writer prompt dereferences `.current_page.<field>`.

| caller (live agent) | current_page source | loader that must carry the column |
|---|---|---|
| `page-build-handler` | `page_record` | **`load_page_record_action.go`** ← load-bearing for rebuild |
| `pageflow-builder` (loop) | loop var over `get_pages_to_build` | `get_pages_to_build_actions.go` |
| `page-rebuild` (loop) | loop var over `get_pages_to_build` | `get_pages_to_build_actions.go` |
| `site-work-orchestrator` | `current_item.spec` | spec = `json.Marshal(page)` from `get_pages_to_build` |

> **CORRECTION to the handoff's implicit model, 2026-07-21:** seed
> `055_page_build_handler.sql:85` shows `page-build-handler` mapping
> `"current_page": "input_data.spec"`. **The LIVE config does not** — it maps
> `"current_page": "page_record"`, the fresh output of `load_page_record`. So for
> the rebuild path (the handoff's own acceptance test — "rebuild the page") the
> load-bearing loader is `load_page_record`, NOT the spec. Caught by dumping the
> live `page-build-handler` default_config; the seed is migration history, not truth.

**Both loaders need the column.** load_page_record for page-build-handler; and
get_pages_to_build for pageflow-builder / page-rebuild / the spec that
site-work-orchestrator reads.

## Changes

1. **Go — `load_page_record_action.go`**: add `content_direction` to `selectCols`,
   scan it (nullable jsonb → parsed value), add to the result map. *(essential — feeds
   page-build-handler rebuild path)*
2. **Go — `get_pages_to_build_actions.go`**: add `content_direction` to both SELECTs
   (includeAll + filtered), scan in both `scanPageRowsForBuild` (sql) and
   `scanPageRowsForBuildPgx` (pgx), add to both page maps. *(feeds pageflow-builder,
   page-rebuild, and the work-item spec)*
3. **Prompt — `page-content-writer` default_config**: `current_page` is already an
   `input_field`; add a guarded `{{if .current_page.content_direction}}` block to the
   `generate_content` prompt_template rendering instruction/format/examples/avoid.
   Live DB UPDATE + a numbered migration for durability.
4. **Docs**: correct `pages.content_direction` COMMENT (now implemented) and the
   `003_contracts_and_standards` §content_direction entry.

## Ordering & delivery

- Go changes are inert until a chassis image roll. The prompt block is a guarded
  `{{if}}` — harmless before the binary ships (no data → block skipped, no error), so
  either order is safe. Ship image, then verify.
- Column stays; NULL for all 301 pages today, so no backfill.

## Acceptance (from the bug file)

Set `pages.content_direction` on one page, rebuild it, confirm the instruction's
effect in the **saved section / rendered artifact** — not a `complete` status.
