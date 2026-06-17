# Running notes — gamesdesign.co.uk silent no-op rebuild

*Root cause now CONFIRMED (this session): the writer→parent response is dropped
when it exceeds an inter-agent size limit, so the compiled sections never reach
the save step. No fix applied yet. Diagnosis-only, but ready to move to a fix
once the truncation mechanism/threshold is located. Supersedes the reasoning
state in `PREAMBLE_gamesdesign_diagnosis_handoff.md` where they disagree — that
preamble's "generation shortfall" framing is now falsified.*

DB access used for all queries:
`kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db`

---

## Symptom

The gamesdesign.co.uk root `index` rebuild reports success, but the live page
stays stale (deployed `page_components` last written 2026-06-06 16:59; nothing
since). It presents as success — the `page_rerender` work item completes, no
error surfaces — so it is invisible without inspecting stored state.

---

## ROOT CAUSE (confirmed) — writer→parent response exceeds a size limit

The `page-content-writer` produces a full compiled page, but its completion
response to the parent `page-build-handler` is **capped when the full result
exceeds an inter-agent size limit** and replaced with a summary envelope. The
compiled `sections_metadata` (and `page_html`) never cross the boundary, so the
parent's `save_page_sections` reads its configured path
`page_content.response.sections_metadata` and finds nothing.

### Evidence — A/B of the parent's received `page_content.response`

| writer run | date | parent `page_content.response` keys | sections_metadata |
|---|---|---|---|
| `4e0b339a` (deployed) | 06-06 | `page_html, page_name, section_count, sections_metadata` | array, **5** |
| `472eed7d` (no-op)    | 06-15 | `completed_at, completed_steps, message, orchestration_id, status` | **absent** |

On the 06-15 parent (`cd73eea6`): `completed_steps` is a bare number, and
`message` = **"Workflow completed successfully. Full result exceeded size
limit."** That message is the cap firing. The writer's own state for the same
run holds the full output (see Confirmed), so the content exists — it was not
propagated.

### The chain

`page-content-writer` compiles `page_content` (~81k: `page_html` ~34.5k +
`sections_metadata` 5×full `rendered_html` ~34k + `content_data` ~8.7k) →
`complete` step returns it → **framework detects the result exceeds the size
limit and substitutes a summary envelope** → parent stores that envelope under
`page_content.response` → `save_page_sections` resolves
`page_content.response.sections_metadata` → null.

### Two downstream manifestations (same cause)

- **No-sections skip.** Path null → save logs "no sections found", returns
  `sections_saved: 0`. (The `cd73eea6` / 06-15 mode.)
- **Short-fallback block.** Path null → save falls to `html_field =
  validation_result.clean_html`, which in the 06-14 runs supplied ~2.8–3.5k and
  tripped the content-regression guard (`agent_error_log`). `validation_result`
  is itself downstream of the same truncated `page_content`, so it is starved too.

Both leave the live page stale.

### Scope / impact

Not gamesdesign-specific. Any page whose compiled result exceeds the size limit
will silently fail to save. The fix must handle large compiled pages generally,
not just this site.

---

## The flow (call chain, with the break point)

`page-build-handler`: `plan_sections` → `spawn_content_writer` (`spawn_agent`,
role `content_writer`, `output_field: writer_agent`) → `call_content_writer` →
[child] `page-content-writer`: `select_sections` → `process_sections_loop`
(`generate_content` → `render_section` → `section_output`) → `compile_page`
(`compile_page_sections`, output `page_content` = `{page_html, page_name,
section_count, sections_metadata}`) → `complete` (`output_field: page_content`)
→ **[BOUNDARY — result capped if oversized]** → parent stores
`page_content.response` (envelope, not the result) → `validate_content` →
`save_sections` (`save_page_sections`, `sections_metadata_field:
page_content.response.sections_metadata`, `html_field:
validation_result.clean_html`) → null → no-op/short-fallback.

---

## Two faults (one is fully reframed)

1. **Truncated writer→parent handoff (ROOT CAUSE — reframed).** Was filed as a
   generation shortfall; generation is sound. The compiled result is dropped at
   the inter-agent boundary for exceeding a size limit.
2. **Status rollup (visibility fault — still open).** A save that no-ops
   (`sections_saved: 0`) or errors on the guard still rolls the `page_rerender`
   work item up to `complete`. This is why the rebuild looks successful. Separate
   from fault 1; worth fixing regardless so a real failure is visible.

---

## Ruled out (checked and falsified — do not re-derive)

- **"Generated sections never reach save."** FALSE — but for a subtler reason
  than first thought: the sections are produced in full; the *response carrying
  them* is truncated before save.
- **Persisted section status in `site_plan_sections`.** Not stored; computed at
  runtime in `plan_sections_action.go`.
- **Triage starving the writer.** FALSE. Every recent index run is
  `ready_count=5, deferred=0, skipped=0`; per-section plan shows all five ready.
- **`max_tokens: 2000` truncation.** FALSE. Per-section `generated_content`
  399/438/443/406/1420 chars (~100–400 tokens), far under the cap.
- **Recreate-mode-not-firing as the volume cause.** Recreate-mode IS broken
  (`build_mode=recreate`, `has_existing=false` on every recreate run) but does
  not drive the shortfall — the 06-06 build that deployed also had
  `has_existing=false`. Separate defect; track independently.
- **Query-resolved list data vanished.** FALSE. `resolved_data` equivalent in the
  deploying vs blocked build (tool-list ~1962 vs ~1930, etc.).
- **"Guard counts `content_data` / `section_output.rendered_html` is short."**
  FALSE. `section_output` rendered_html is full; `extractSectionsFromMetadata`
  reads `rendered_html` (L411).
- **`compile_page` builds short `sections_metadata`.** FALSE (this session).
  `CompilePageSectionsAction` → `extractSectionFromMap` → `extractHTMLFromSectionMap`
  sets `meta.rendered_html` to the FULL rendered HTML (v3_site_actions.go
  L1838/L1917); `content_data` is attached separately, not as the HTML. Writer's
  own `page_content.sections_metadata` is a 5-element array, per-section stripped
  ~23k total (see below).
- **The save path is misconfigured.** FALSE. `sections_metadata_field =
  page_content.response.sections_metadata` is correct for the un-truncated case
  (it resolves to the full array on the 06-06 build). The defect is that the
  array is absent when the response is capped.

---

## Confirmed (with evidence)

- **Writer output is full** (run `472eed7d`, writer-side `page_content`):
  `section_count=5`, `page_html` 34,523; `sections_metadata` 5 entries —
  hero 2288/1984, tool-list 9047/5306, guide-list 7458/5672, game-list
  8155/5077, system-stats 7369/4999 (rendered_html / stripped). Stripped total
  ~23k. (Writer-side path is `page_content.sections_metadata`; the `response`
  wrapper exists only on the parent — the earlier 0-rows on
  `page_content.response.sections_metadata` was the wrong path for the writer
  row, not absent data.)
- **Parent received an envelope, not the result** (06-15): see A/B table and the
  "Full result exceeded size limit" message.
- **Parent received the full result on the deploying build** (06-06): array of 5.
- **Deployed page intact** — 5 `deployed` components, rendered_html
  2426/8951/7513/8116/7369 (~34,375). Guard returns before the snapshot/DELETE.
- **Guard arithmetic correct** and the INSERT writes `s.HTML` into the
  `rendered_html` column — so a short payload would also corrupt the column; the
  guard is protecting the live page, not merely miscounting.

---

## Guardrails (the tempting wrong moves)

- **Do NOT loosen or remove the content-regression guard.** It protects the live
  34k page, and the INSERT persists `s.HTML`.
- **Do NOT raise `max_tokens`.** Falsified as a cause.
- **Do NOT "fix" generation, compile, recreate-mode, or the save path config** to
  address the shortfall — all exonerated. (Recreate-mode is a separate real defect.)
- **Do NOT simply raise the inter-agent size limit** without weighing the bus and
  storage implications — pushing ~81k+ payloads through the message path is the
  thing the limit exists to prevent. Prefer a structural fix that keeps large
  artifacts out of the response (see plan).

---

## Plan — confirm the mechanism, then choose a structural fix

Still diagnosis-first. Locate the cap before changing anything.

**Step A — find where the response is capped and its threshold.** The literal
message is the anchor.

```bash
grep -rn 'exceeded size limit\|Full result exceeded\|size limit' \
  platform/ internal/ pkg/ 2>/dev/null
# expect the cap in the complete/complete_workflow action or the response publisher.
```
Establish: the threshold (bytes), what it measures (final_result vs the full
collected_data), and whether it truncates or substitutes the envelope we observed.

**Step B — structural fix options (decide after Step A; not yet applied).**
The architecture currently expects the full `sections_metadata` to cross the
boundary, which is inherently large. Candidates, with trade-offs:

1. *Save reads from the writer's persisted state, not the message.* The compiled
   `sections_metadata` already lives in the child's
   `orchestration_states.collected_data->page_content`. The parent holds the
   child id (`writer_agent` output / `orchestration_id` in the response), so
   `save_page_sections` (or a small load step) could fetch the array from the DB
   by orchestration_id. Keeps large artifacts off the bus; aligns with
   "complexity in actions, messages small." Requires a new read action / config.
2. *Writer persists its own sections and returns a reference.* The writer is an
   orchestrator; it could persist sections to a durable store and return a row/id.
   Parent's save reads by reference. Largest change to responsibilities.
3. *Writer owns the save.* Move `save_page_sections` into the writer so the
   payload never crosses the boundary. Changes which agent owns persistence —
   weigh against distinct-responsibilities.
4. *Raise the limit* — quick, but see guardrail; only if Step A shows the
   threshold is unreasonably low and payloads are bounded.

Recommendation to evaluate first: option 1 (read by orchestration_id) — smallest
structural change that removes the size pressure and reuses existing state.

**Step C — Fault 2 (separate track).** In `page-build-handler`, trace how a
`save_sections` no-op (`sections_saved: 0`) or guard error maps to the
`page_rerender` work-item terminal status. A no-op/error must drive the work item
to failed/needs-review, not `complete`. Steps present that bear on this:
`check_content_produced`, `check_has_ready_sections`, `validate_content`,
`mark_needs_review`, `update_status`.

---

## Verification when fixed

- **Fault 1:** an index rebuild lands the full `sections_metadata` at the save
  step (or save fetches it from the writer's state), the guard does not fire, and
  `page_components.rendered_html` updates (timestamps advance past 06-06).
- **Fault 1 (negative):** a deliberately oversized compiled page still saves —
  the size limit no longer drops the content.
- **Fault 2:** a no-op/blocked save drives the `page_rerender` work item to
  failed/needs-review, not `complete`.

---

## Key IDs / where things live

- Site `gamesdesign.co.uk`, page `index`. Deployed: 5 sections, ~34k, written
  2026-06-06 16:59 (unchanged — confirms no rebuild has saved).
- Writer runs (page-content-writer, index): `472eed7d` (06-15, no-op; drill-down)
  with parent `cd73eea6` (page-build-handler, holds the truncated envelope);
  `4e0b339a` (06-06, deployed; parent holds the full result — A/B baseline).
- The cap surfaces as `page_content.response.message = "Workflow completed
  successfully. Full result exceeded size limit."` with `completed_steps` a number.
- Save step config (`page-build-handler`): action `save_page_sections`,
  `sections_metadata_field = page_content.response.sections_metadata`,
  `html_field = validation_result.clean_html`, `output_field = sections_saved`.
- `page-build-handler` steps: call_content_writer, check_content_produced,
  check_has_ready_sections, check_page_found, complete, complete_error,
  deploy_page, ensure_site_record, load_existing_content, load_page_record,
  load_spec_sections, mark_needs_review, plan_sections, save_sections,
  spawn_content_writer, spawn_rerender_agent, update_status, validate_content.
- Code read this session: `v3_site_actions.go` — `CompilePageSectionsAction`
  L1612, `extractSectionFromMap` L1838, `extractHTMLFromSectionMap` L1917 (output
  is top-level `{page_html, page_name, section_count, sections_metadata}`, no
  `response` wrapper). `save_page_sections_action.go` instrumented last session
  (Info logs at the save boundary; still useful to confirm the no-op path live).
- Not yet read: the complete/complete_workflow action or response publisher that
  enforces the size cap (Step A target).
- Empty in project mount: `002_intake_orchestrator.sql`, `003_site_classifier.sql`,
  `018_briefing_questionnaire.sql` (0 bytes).
