# Running notes — gamesdesign.co.uk silent no-op rebuild

*Diagnosis-only. No fixes applied. Updated this session with runtime evidence
that reframes the root cause. Extends the reasoning state in
`PREAMBLE_gamesdesign_diagnosis_handoff.md`; where they disagree, this file is
newer.*

DB access used for all queries below:
`kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db`

---

## Symptom

The gamesdesign.co.uk root `index` rebuild reports success, but the live page
stays stale (deployed `page_components` last written 2026-06-06 16:59; nothing
since). It "presents as success" — the `page_rerender` work item completes, no
error surfaces — so it is invisible without inspecting stored components.

---

## The flow (call chain that reaches the block)

`page-build-handler` → `plan_sections` (emits `section_plan`: ready/deferred/
skipped + each ready section's `llm_field_specs`) → `spawn_content_writer` →
`page-content-writer`: `select_sections` (reads `input_data.section_plan.
sections_ready`) → `process_sections_loop` (per section: `generate_content`
[`execute_llm_prompt`, `max_tokens: 2000`] → `render_section`
[`render_component`, output `section_output`]) → `compile_page`
[`compile_page_sections`, `sections_from: processed_sections`, output
`page_content`] → returns `page_content` to `page-build-handler` → `save_sections`
step runs `save_page_sections` → **content-regression guard blocks**.

---

## Two faults (still two; one is reframed)

1. **The save-side shortfall (root cause — REFRAMED this session).** Originally
   filed as a *generation* shortfall. Generation is now shown to be sound (see
   ruled-out + confirmed). The real fault is that the `sections` payload the
   regression guard measures is ~3k of text while the writer's actual rendered
   sections are ~34k. The loss happens **between `section_output` and the guard**
   — in `page_content` assembly (`compile_page` → `page_content.response.
   sections_metadata`) and/or save-step path selection — not in the LLM.
2. **Status rollup (visibility fault — still open, unchanged).** The blocked save
   is logged as `step save_sections failed` in `agent_error_log`, yet the
   `page_rerender` row in `site_work_items` is `complete`. A failed step rolls up
   to a completed work item, which is why the rebuild looks successful. Separate
   from fault 1, worth fixing regardless: a genuine future regression-block must
   be visible, not silently `complete`.

---

## Ruled out (checked and falsified — do not re-derive)

- **"Generated sections never reach save."** FALSE. 15 rows in `agent_error_log`
  show sections reach save every rebuild; save fails on the guard, not on missing
  input.
- **Persisted section status in `site_plan_sections`.** Not a thing.
  `site_plan_sections` has no status column (`\d`: `plan_id, page_name, ordering,
  component_name, component_version_id, palette_id, layout_id,
  typography_set_id`). Readiness is computed at runtime in
  `plan_sections_action.go` and observable only in the run trace.
- **Triage starving the writer.** FALSE. Every recent index run is
  `ready_count=5, deferred=0, skipped=0, total=5`; the per-section plan confirms
  `hero, tool-list, guide-list, game-list, system-stats` all `ready` with field
  specs (`hero 4, tool-list 3, guide-list 4, game-list 3, system-stats 24`).
- **`max_tokens: 2000` truncation.** FALSE. Per-section `generated_content` is
  399/438/443/406/1420 chars of JSON (~100–400 tokens) — far under the 2000 cap.
  The LLM is asked for a few short fields; it is not being truncated. (Revises the
  preamble's "2000 tokens can't reproduce ~8.9k": the writer never had to — the
  component template expands the short fields into the large `rendered_html`.)
- **Recreate-mode-not-firing as the volume cause.** Recreate-mode IS broken
  (`build_mode=recreate`, `has_existing=false`, empty `raw_markdown` on every
  recreate run) — but it does not drive the shortfall: the build that deployed the
  live 34k page (`4e0b339a`, 06-06) also had `has_existing=false` and produced a
  full page. Same flag, opposite outcome. Track separately; it is not fault 1.
- **Query-resolved list data vanished in rebuilds.** FALSE. A/B of the deploying
  build vs a blocked build shows equivalent `resolved_data` per section
  (tool-list ~1962 vs ~1930, guide-list ~1624 vs ~1591, game-list ~1721 vs
  ~1689). The list bulk is present.
- **"Guard counts `content_data`, or `section_output.rendered_html` is short."**
  FALSE (this session). `section_output_N.rendered_html` is full (sum ~34,317;
  stripped ~23,038), `content_data` sums to ~8,717, and
  `extractSectionsFromMetadata` provably reads `rendered_html` (line 411,
  `m["rendered_html"]`) and skips empties. The guard's ~3k matches neither field.

---

## Confirmed (with evidence)

- **Deployed page intact and content-rich.** 5 `deployed` `page_components`:
  rendered_html 2426/8951/7513/8116/7369 (~34,375), content_data
  882/2261/2079/2045/1219 (~8,486). The guard returns before the
  snapshot/DELETE, so the live page is protected, not damaged.
- **Generation is sound.** Writer run `472eed7d` (06-15, index): 5 sections, all
  ready, full `resolved_data`, `section_output` rendered_html
  2288/9047/7458/8155/7369 — equivalent to the deploying build `4e0b339a`. The
  writer produces a full ~34k page.
- **The guard input is short, the writer output is not.** For `472eed7d`,
  `section_output` rendered_html stripped sums to ~23k, yet the guard reported
  ~3k for the corresponding save. So the `sections` slice the guard measured was
  NOT `section_output`. The loss is downstream of the loop.
- **Guard arithmetic is correct.** `existingTextLen` = SUM of stripped
  `rendered_html` from `page_components WHERE build_status='deployed'` (full, ~16k);
  `newTextLen` = SUM of stripped `s.HTML` over the in-memory `sections`; blocks
  when `newTextLen < existingTextLen/4`. The asymmetry is in what populates
  `sections`, not in the comparison.
- **The INSERT writes `s.HTML` into the `rendered_html` column** (line 352:
  `... rendered_html ... VALUES (..., section.HTML, ...)`). So `SectionData.HTML`
  is both what the guard measures AND what gets persisted. If it is the short
  field, the guard is correctly preventing a bad write, not merely miscounting.
  This is why we must distinguish "measurement-only" from "populate" before any fix.

### Evidence table — writer run `472eed7d` (index, blocked)

| section_output | rendered_html | rendered_html stripped | content_data |
|---|---|---|---|
| 0 (hero)         | 2288 | 1984 |  815 |
| 1 (tool-list)    | 9047 | 5306 | 2372 |
| 2 (guide-list)   | 7458 | 5672 | 2039 |
| 3 (game-list)    | 8155 | 5077 | 2099 |
| 4 (system-stats) | 7369 | 4999 | 1392 |
| **sum**          | **34,317** | **23,038** | **8,717** |

Guard reported `newTextLen` across runs: 2854 / 3005 / 3124 / 3143 / 3187 /
3212 / 3254 / 3323 / 3486 (existing 13,029–20,196; existing/4 ≈ 3,257–5,049).

---

## Current localized cause (open — leading hypothesis, NOT confirmed)

The guard's ~3k is close to the **LLM text total** (`generated_content`
399+438+443+406+1420 ≈ 3,106 of JSON), not to the rendered HTML. Leading
hypothesis: `compile_page` (`CompilePageSectionsAction`) builds
`page_content.response.sections_metadata` with `rendered_html` derived from the
LLM `content_data` (the short fields) rather than carrying the full
`section_output.rendered_html`. The save step then reads that short array via
`sections_metadata_field = page_content.response.sections_metadata`, so the
guard sums ~3k and blocks.

Alternatives still on the table until the next checks:
- `sections_metadata` carries full HTML for some sections and empty for others;
  `extractSectionsFromMetadata` skips empties, leaving a short subset. (The
  `new_sections` count is logged in `agent_error_log` callers but not in the
  message — the per-element check below settles it.)
- The metadata path is not firing and the HTML fallback ran. Less likely:
  `saveSectionsExtractFromHTML` captures full `<section>` blocks, and on a full
  assembled document with no `<section>` it returns zero (the `<html>/<!doctype>`
  guard), not ~3k.

`CompilePageSectionsAction` is not in either bundle — it must be read to confirm.

---

## Guardrails (the tempting wrong moves)

- **Do NOT loosen or remove the content-regression guard.** It is the only thing
  protecting the live 34k page, and because the INSERT persists `s.HTML`, a short
  payload would overwrite good `rendered_html` if the guard let it through.
- **Do NOT raise `max_tokens`.** Falsified as a cause; the cap is not hit.
- **Do NOT "fix" generation or recreate-mode to address the shortfall.**
  Generation is sound. (Recreate-mode is a real but separate defect.)
- **Do NOT conclude measurement-bug vs populate-bug without the per-element
  `sections_metadata` data.** They have different fixes and different blast radius.

---

## Plan — next diagnostic steps (still diagnosis-only)

Run against `472eed7d-9af7-4b32-ba7f-e8cab9ac97f2` (blocked index run). If any
returns no rows, suspect the path/id, not absence of data — re-check before
concluding.

**Step 1 — does `page_content` expose the metadata array, and is its HTML short?**
This reproduces the guard's input from `page_content` and is decisive.

```sql
-- 1a. shape of page_content
SELECT jsonb_typeof(collected_data->'page_content')                                   AS pc_type,
       (SELECT string_agg(k, ', ' ORDER BY k)
          FROM jsonb_object_keys(collected_data->'page_content') k)                   AS pc_keys,
       (SELECT string_agg(k, ', ' ORDER BY k)
          FROM jsonb_object_keys(collected_data #> '{page_content,response}') k)      AS response_keys
FROM orchestration_states
WHERE orchestration_id = '472eed7d-9af7-4b32-ba7f-e8cab9ac97f2';

-- 1b. per-element rendered_html in sections_metadata — does it sum to ~3k or ~23k?
SELECT ord,
       e->>'component_function'                                              AS fn,
       length(e->>'rendered_html')                                          AS rh_len,
       length(regexp_replace(COALESCE(e->>'rendered_html',''),'<[^>]+>','','g')) AS rh_stripped,
       length((e->'content_data')::text)                                    AS cd_len
FROM orchestration_states os,
     LATERAL jsonb_array_elements(
       os.collected_data #> '{page_content,response,sections_metadata}'
     ) WITH ORDINALITY AS t(e, ord)
WHERE os.orchestration_id = '472eed7d-9af7-4b32-ba7f-e8cab9ac97f2'
ORDER BY ord;
```

Reading: if `rh_stripped` sums to ~3k (≈ the guard's number) → `sections_metadata`
carries the short HTML → the bug is in `compile_page`. If it sums to ~23k → the
metadata path is fine and save did not use it (config mismatch → HTML fallback);
go to Step 2. If most rows have empty `rendered_html` → the "subset" variant.

**Step 2 — what does the page-build-handler save step actually read?**

```sql
-- confirm the save step name first
SELECT (SELECT string_agg(k, ', ' ORDER BY k)
          FROM jsonb_object_keys(default_config #> '{workflow,steps}') k) AS step_names
FROM agent_definitions WHERE type = 'page-build-handler' ORDER BY version DESC LIMIT 1;

-- then dump the save step config (replace save_sections if the name differs)
SELECT jsonb_pretty(default_config #> '{workflow,steps,save_sections}')
FROM agent_definitions WHERE type = 'page-build-handler' ORDER BY version DESC LIMIT 1;
```

Confirm `sections_metadata_field` = `page_content.response.sections_metadata`
(canonical) and what `html_field` is. A mismatch here would force the HTML
fallback.

**Step 3 — read `CompilePageSectionsAction`** (not in either bundle). Find how it
populates each `sections_metadata` entry's `rendered_html` — whether from the
section's full rendered output or from `content_data`/`generated_content`.

```bash
grep -rn 'func CompilePageSectionsAction\|sections_metadata\|rendered_html' \
  platform/orchestration/actions/ | grep -i compile
sed -n '/func CompilePageSectionsAction/,/^}/p' \
  platform/orchestration/actions/<compile file>.go
```

**After Step 1–3 (pending confirmation, do not apply yet):**
- If `sections_metadata` HTML is short → fix is in `compile_page` so each entry's
  `rendered_html` carries the full rendered section (the same value `section_output`
  holds), structural, in the Go action — reuse the existing render output, don't
  re-derive.
- If save isn't using the metadata path → fix the save step config / the
  `page_content` shape contract so the metadata path fires; then the existing
  `extractSectionsFromMetadata` (which already reads `rendered_html`) works.
- Where the fix lands "in both places" (deferred until cause is confirmed): the
  live `agent_definitions` config in `clients_db` AND the source SQL/migration in
  the repo that redeploys it, so a DB patch is not overwritten by the next apply.

**Fault 2 (separate track):** in `page-build-handler`, trace how a `save_sections`
step error maps to the `page_rerender` work-item terminal status. A step that
returns an error should drive the work item to failed/blocked, not `complete`.
Worth fixing regardless of fault 1.

---

## Verification when fixed (both faults)

- **Fault 1:** an index rebuild produces `sections_metadata` rendered_html within
  range of the deployed ~16k stripped, the guard does not fire, and
  `page_components.rendered_html` updates with new content (timestamps advance).
- **Fault 2:** a forced regression-block (or a real one) drives the `page_rerender`
  work item to failed/blocked, not `complete`.

---

## Key IDs / where things live

- Site: `gamesdesign.co.uk`, page `index`. Deployed: 5 sections, ~34k, written
  2026-06-06 16:59 (unchanged since — confirms no rebuild has saved).
- Writer runs (page-content-writer, index): `472eed7d` (06-15, blocked, use for
  drill-downs), `4e0b339a` (06-06, the build that deployed the live page; A/B
  baseline). `b29a6849` (06-14) is FAILED with no `processed_sections`.
- `orchestration_states` keys: writer per-iteration outputs are `section_output_0..N`
  and `generated_content_0..N`; loop output is `processed_sections` (object, not
  array — `jsonb_array_length` on it errors); terminal is `page_content` (object,
  ~81k; size is structure, not a content-volume proxy).
- Source read this session: `save_page_sections_action.go` (uploaded). Guard at
  ~L224–260; metadata path L142–156; HTML fallback `saveSectionsExtractFromHTML`
  L455; `extractSectionsFromMetadata` L394 (reads `rendered_html`); INSERT L352.
- Empty in project mount (could not read here): `002_intake_orchestrator.sql`,
  `003_site_classifier.sql`, `018_briefing_questionnaire.sql` (0 bytes).
