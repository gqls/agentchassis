# 194 — four of six `save_page_sections` callers never map `sections_metadata`, so every save through them NULLs `content_data`

**Filed:** 2026-08-04 · **By:** the `bugs_open/087` lane, from 087's own live acceptance run
**Status:** OPEN — mechanism confirmed first-hand, one of four instances fixed (see below)
**Severity:** Medium-high. `content_data` is what the rerender path regenerates a section from.
A save that NULLs it does not corrupt the served page — it silently removes the page's ability
to be rerendered, and nothing reports it.

## Symptom, observed live

`bugs_open/087`'s acceptance test dispatched `page-rebuild` at
`vetcomparison.uk` / `tool-cma-obligation-checker-guide`
(correlation `3fdf4acf-5f96-49f9-8801-28047aae92ef`, 2026-08-04 09:47–09:50Z, all steps
`COMPLETED`, page redeployed correctly). The rebuild worked. But:

| slot | `content_data` before | after |
|---|---|---|
| `hero` | 644 chars | **NULL** |
| `article-body` | 3,810 chars | **NULL** |
| `call-to-action` | 420 chars | **NULL** |

`rendered_html` is fine on all three (3,243 / 5,317 / 2,303 chars, fresh copy, page serves
HTTP 200 with all three `data-component` slots present). The row `id`s changed too — the
action deletes the agent-writable components and re-inserts them.

## Root cause — a config key that four of six callers simply do not set

`SavePageSectionsAction` (`platform/orchestration/actions/save_page_sections_action.go`)
writes `content_data` on insert (`:685-687`), taking it from the section's own
`content_data`, which reaches it via a **`sections_metadata_field`** config key. Callers that
supply that key preserve structured content; callers that do not write SQL NULL.

Read off the live rows (`jsonb_path_query(default_config, '$.**.steps.*')` — the step is
nested inside a loop `sub_workflow` in four of the six, so a top-level `jsonb_each` misses them):

| caller | `sections_metadata_field` | result |
|---|---|---|
| `page-build-handler` | `page_content.response.sections_metadata` | **preserved** |
| `page-rerender` | `rerender_sections.sections_metadata` | **preserved** |
| `page-rebuild` | *absent* | **NULLed** |
| `pageflow-builder` | *absent* | **NULLed** |
| `site-work-orchestrator` | *absent* | **NULLed** |
| `tool-recreation-handler` | *absent* | **NULLed** |

**The data is available on every one of those paths — it is simply not mapped.** The writer
returns it: `page-content-writer`'s `compile_page` output keys on the run above are
`page_html, page_name, section_count, sections_metadata`. The three `page-*`/`*-builder`
callers all store the writer's reply under `output_field: page_content`, so
`page_content.response.sections_metadata` — byte-identical to what `page-build-handler`
already uses — resolves for all of them.

Fleet-wide as of 2026-08-04: **161 of 1,201 `page_components` rows (13.4%) have NULL
`content_data`.** That figure is *exposure, not damage attribution* — it is a snapshot of
the population, and this case has not traced which of those 161 were NULLed by these four
callers versus never having had structured content in the first place. Do not quote it as
"161 pages damaged".

## Why it is not merely cosmetic

- `page-rerender` regenerates a section **from `content_data`** — with it NULL there is
  nothing to regenerate from. This is the same asymmetry recorded in MEMORY as *"a repro is
  destroyed by the render"*.
- The pre-overwrite value **is** archived, so this is recoverable in principle:
  `page_component_history` gets an insert (`:536-538`) with `source='save_page_sections_overwrite'`,
  and the three rows for the run above hold exactly the pre-run lengths (644 / 3,810 / 420).
  **But `component_id` on those rows is NULL** — the FK is `ON DELETE SET NULL` and the action
  deletes the components before re-inserting, so the archive survives while its link to a slot
  does not. Recovery has to match on content, not on id.
- **A naive restore is the wrong repair.** Putting the archived `content_data` back beside the
  *new* `rendered_html` pairs stale structured content with fresh HTML; the next rerender would
  regenerate the *old* page over the new one. The correct repair is to fix the mapping and
  re-run the build, so both halves are written together.

## Fix

**Config-only, one key per caller, live on apply.** Mirror `page-build-handler` exactly:

```json
"sections_metadata_field": "page_content.response.sections_metadata"
```

- **`page-rebuild` — APPLIED**, `docs/agent_docs/sql_for_agents/310_page_rebuild_preserves_content_data.sql`,
  and proven by re-running the same page (see "Verification" below). Done here rather than
  deferred because this lane's own acceptance run is what NULLed those three components, and
  leaving a live page unable to rerender would have been damage recorded and not repaired.
- **`pageflow-builder`, `site-work-orchestrator`, `tool-recreation-handler` — NOT DONE.**
  Deliberately: the first two are the same two callers `bugs_open/087` found broken in the
  *other* direction and both are dormant, and `tool-recreation-handler` runs a different
  writer flow whose response shape this case has **not** read — asserting
  `page_content.response.sections_metadata` for it would be an unmeasured claim. Read its
  writer's output keys before copying the line.

## Verification

Re-run a rebuild on a `rebuild_policy != 'owned'` page and require **both** halves:

```sql
SELECT slot_name, length(rendered_html) AS html, length(content_data::text) AS data, updated_at
FROM page_components WHERE page_id='<page>'::uuid ORDER BY position;
```

`data` must be non-NULL on every row, and `updated_at` must be the new run's. The
disconfirming outcome is stated so it cannot be retrofitted: **if `content_data` is still NULL
after the mapping is added, the writer's reply is not reaching the save step under that path
and the fix is wrong** — check `page_content.response` actually carries `sections_metadata` on
*that* caller before assuming the key name is at fault.

## Related

- `bugs_open/087` — the same agent, the same class (a caller that does not map a field its
  callee needs), found by 087's acceptance run. 087's fix makes the *writer* independent of
  its callers; this one is the *saver*'s equivalent and is still caller-supplied.
- `bugs_closed/125` — `page-rebuild` deploying to a name-derived path. Same agent, same
  "five implementations, four correct" shape.
- `bugs_open/098` — archiving does not retract; the rerender path this defect disables is one
  of the mechanisms 098 depends on.

---

## 2026-08-04 10:01Z — `page-rebuild`'s instance FIXED and PROVEN LIVE

Seed 310 applied, then the **same page re-run** (correlation
`76008af6-495d-4ca1-a1da-57b09048417e`, 09:58–10:01Z, all steps `COMPLETED`):

| slot | `rendered_html` | `content_data` before 310 | after 310 |
|---|---|---|---|
| `hero` | 3,132 | **NULL** | **626** |
| `article-body` | 4,119 | **NULL** | **2,797** |
| `call-to-action` | 2,335 | **NULL** | **433** |

Both halves written together at the new run's `updated_at`, `build_status` → `deployed`,
page serves HTTP 200 at its canonical url and 404 at the name-derived one. The damage this
lane's own acceptance run caused is repaired — by rebuilding, not by restoring the archive,
for the reason stated above.

The verify block was **induced before being trusted**: run alone against the unmodified row
it raised `194/310: sections_metadata_field is <NULL>, expected …`. A verify block you have
not seen fail is a claim, not a check.

**Still open:** `pageflow-builder`, `site-work-orchestrator`, `tool-recreation-handler`.
The first two are dormant and are the same pair `bugs_open/087` found broken in the other
direction; the third runs a different writer flow whose response shape is **[UNMEASURED]** —
read its writer's output keys before copying the line.
