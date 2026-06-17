# Running notes — gamesdesign.co.uk silent no-op rebuild

*Root cause CONFIRMED at the framework boundary: the writer's `complete` step
declares `output_field` (singular) instead of `output_fields` (plural), so the
coordinator dumps the writer's whole state, exceeds the 900k result cap, and
returns a `status:completed` stub that drops the compiled page. No fix applied
yet — the exact `output_fields` value must reproduce the parent's read path and
is pending recovery of the pre-06-13 config. Supersedes
`PREAMBLE_gamesdesign_diagnosis_handoff.md` (its "generation shortfall" framing
is falsified). A generalised entry for this failure mode has been added to
`016_debugging_guide_v2_50.md` §9.*

DB access for all queries:
`kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db`

---

## Symptom

The gamesdesign.co.uk `index` rebuild reports success, but the live page stays
stale (deployed `page_components` last written 2026-06-06 16:59; nothing since).
Work item completes, no error surfaces — invisible without inspecting stored state.

---

## ROOT CAUSE (confirmed) — child result dropped by the size-limit fallback

The `page-content-writer` produces a full compiled page, but its result never
reaches the parent `page-build-handler`. Two compounding defects:

1. **Wrong output key (data, the trigger).** The writer's `complete` step is
   `{"action":"complete_workflow","config":{"output_field":"page_content"}}`.
   `output_field` (singular) only controls where a step stores its own output in
   `collected_data`. The coordinator's `extractWorkflowResult`
   (`coordinator.go` L3357–3367) reads `output_fields` (**plural**). The singular
   key is invisible, so it takes the fallback (L3380): serialise nearly all of
   `collected_data` minus `skipPatterns`. Those patterns skip `page_content_`
   (trailing underscore → `page_content_0`), `assembled_page`, etc. — but NOT
   `page_content`, `processed_sections`, or `section_output_0..N`. For a
   multi-section page that dump exceeds `MaxResultSizeBytes` (900000, L3432).
2. **Lossy size guard (code, the silent failure).**
   `extractWorkflowResultWithSizeLimit` (L3434) tries to shrink it, but the
   trimmer (L3454) only truncates top-level **string** fields > 10k (source
   comment: "Could also recurse into maps/arrays if needed" — it does not). The
   bulk is nested objects/arrays, so nothing shrinks and it falls to
   `extractMinimalResult` (L3475) — a stub `{orchestration_id, completed_steps,
   completed_at, status:"completed", message:"Workflow completed successfully.
   Full result exceeded size limit."}`. The compiled page is gone, and the
   `status:"completed"` is a false success.

The parent stores that stub under `page_content.response`. `save_page_sections`
reads `page_content.response.sections_metadata` → null → no-op ("no sections
found", `sections_saved:0`) or falls back to a short `validation_result.clean_html`
that trips the regression guard. Page stays stale; the false success rolls the
`page_rerender` work item up to `complete` (this is the visibility fault, same
root).

### Scope / impact

Not gamesdesign-specific. Any agent whose `complete` step uses `output_field`
(singular), or whose legitimate result exceeds 900k, hits this. A sibling
Kafka-size failure was seen on `image-build-handler` `complete_workflow`.

---

## The flow (call chain, with the break point)

`page-build-handler`: `ensure_site_record` → `load_page_record` →
`load_existing_content` → `load_spec_sections` → `plan_sections` →
`check_has_ready_sections` → `spawn_content_writer` → `call_content_writer`
(`call_agent`, `output_field: page_content`) → [child] `page-content-writer`:
… → `compile_page` (output `{page_html, page_name, section_count,
sections_metadata}`) → `complete` (`output_field: page_content` ← WRONG KEY) →
**[coordinator `extractWorkflowResult`: no `output_fields` → dump state → >900k →
trim fails → minimal stub]** → parent stores stub at `page_content.response` →
`check_content_produced` (passes — `skipped` not set) → `validate_content` →
`save_sections` (`sections_metadata_field: page_content.response.sections_metadata`,
`html_field: validation_result.clean_html`) → null → no-op/short-fallback →
`update_status` → deploy → `complete`.

---

## Two faults (one fully reframed)

1. **Child result dropped at `complete` (ROOT CAUSE).** Was filed as a generation
   shortfall; generation is sound. The compiled result is discarded by the
   coordinator's size-limit fallback because the writer never declared
   `output_fields`.
2. **Status rollup (visibility fault).** A no-op/blocked save still rolls the
   `page_rerender` work item to `complete` — because the dropped-result stub
   reports `status:"completed"` and the save returns success on 0 sections. Same
   root; the coordinator hardening below fixes the silent-success half.

---

## Ruled out (checked and falsified — do not re-derive)

- **"Generated sections never reach save."** The sections are produced in full;
  the *response carrying them* is replaced by a stub before save.
- **Persisted section status in `site_plan_sections`.** Not stored; runtime only.
- **Triage starving the writer.** FALSE — every recent index run is
  `ready_count=5, deferred=0, skipped=0`; all five sections ready.
- **`max_tokens: 2000` truncation.** FALSE — per-section `generated_content`
  ~100–400 tokens, far under the cap.
- **Recreate-mode-not-firing as the volume cause.** Recreate-mode IS broken
  (`has_existing=false` on every recreate run) but does not drive the shortfall —
  the 06-06 deploying build also had `has_existing=false`. Separate defect.
- **Query-resolved list data vanished.** FALSE — `resolved_data` equivalent in the
  deploying vs blocked build.
- **"Guard counts `content_data` / `section_output` is short."** FALSE — full
  rendered HTML; guard reads `rendered_html`.
- **`compile_page` builds short `sections_metadata`.** FALSE — `extractSectionFromMap`
  / `extractHTMLFromSectionMap` (v3_site_actions.go L1838/L1917) set
  `rendered_html` to the FULL HTML; writer-side `page_content.sections_metadata`
  is a 5-element array, per-section stripped ~23k total.
- **The save path is misconfigured.** FALSE — `page_content.response.sections_metadata`
  resolves to the full array on the 06-06 build; it is null only because the
  response is a stub.
- **The bundle's writer def was stale.** FALSE — live `version 2` (updated
  2026-06-13 12:16) matches: `complete = {output_field: page_content}`.

---

## Confirmed (with evidence)

- **Writer output full** (child `472eed7d`): `section_count=5`, `page_html`
  34,523; `page_content.sections_metadata` 5 entries, per-section
  rendered_html/stripped 2288/1984, 9047/5306, 7458/5672, 8155/5077, 7369/4999
  (~23k stripped). `page_content` ~81k.
- **Parent received a stub** (06-15, parent `cd73eea6`): `page_content.response`
  = `{completed_at, completed_steps(number), message, orchestration_id, status}`,
  `message` = "Workflow completed successfully. Full result exceeded size limit."
- **Parent received the real output on the deploying build** (06-06, `4e0b339a`):
  `page_content.response` = `{page_html, page_name, section_count,
  sections_metadata(array,5)}` — the compile output **flattened**.
- **Coordinator mechanism** (coordinator.go): `MaxResultSizeBytes=900000` (L3432);
  `extractWorkflowResult` reads `output_fields` plural (L3359), fallback dumps
  state minus `skipPatterns` that don't cover `page_content`/`processed_sections`/
  `section_output_*` (L3383–3414); trimmer top-level strings only (L3454–3460);
  `extractMinimalResult` message matches (L3481).
- **Live writer def:** one row, `version 2`, updated 2026-06-13 12:16,
  `complete = {output_field: page_content}`. `page-build-handler`'s own `complete`
  correctly uses `output_fields: ["sections_saved","deploy_result"]`.
- **Regression dating (by elimination):** the 06-06 flattened response is
  incompatible with the current singular-key fallback (which would expose
  `processed_sections`/`section_output_*` as sibling keys, not the four compile
  keys alone). So the prior version's `complete` used a correct, flattening output
  contract; the 06-13 v2 update regressed it. Symptom onset == def update date.

---

## Guardrails (the tempting wrong moves)

- **Do NOT raise `MaxResultSizeBytes`** — the limit guards the Kafka ceiling;
  pushing the whole working state across the bus is the thing to stop.
- **Do NOT loosen the content-regression guard** — it correctly protects the live
  page; the INSERT persists `s.HTML`.
- **Do NOT "fix" generation, compile, recreate-mode, or the save path config** —
  all exonerated. (Recreate-mode is a separate real defect.)
- **Do NOT assume `output_fields: ["page_content"]` is the fix** — it nests under
  a `page_content` key (array → `page_content.response.page_content.sections_metadata`),
  which does not match the parent's read path. Recover the pre-06-13 config.

---

## Plan

**Step 1 — recover the correct output contract (blocks the exact patch).** Only
`version 2` is in `agent_definitions`; the pre-06-13 config that produced the
06-06 flattened response is in the snapshot mechanism (debugging guide §6.1).
Pull the prior `complete` step for `page-content-writer` and confirm what
`output_fields`/output shape yields `page_content.response.sections_metadata`
directly (flattened). That value is the patch.

**Step 2 — apply the contract fix (data, both places, snapshot first).** Set the
writer's `complete` step to the recovered `output_fields` so the coordinator
returns only the compiled page (~81k < 900k). Land in the live `agent_definitions`
row AND the source SQL migration that defines `page-content-writer`. Snapshot
before the live UPDATE (§6.1). Verify the parent then receives the full
`sections_metadata` and the page saves.

**Step 3 — harden the coordinator (code, separate, affects all agents).**
`extractWorkflowResultWithSizeLimit`: recurse the trimmer into maps/arrays, or
persist the large artefact and return a reference; and never return
`status:"completed"` when configured `output_fields` were dropped — surface a
failure instead. This also fixes Fault 2's silent-success half.

---

## Verification when fixed

- **Fault 1:** an index rebuild lands the full `sections_metadata` at
  `page_content.response.sections_metadata`, the guard does not fire, and
  `page_components.rendered_html` updates (timestamps advance past 06-06).
- **Fault 1 (negative):** a deliberately oversized result no longer collapses to
  a stub — it either fits via reference or fails loudly; it never reports success
  while dropping output.
- **Fault 2:** a no-op/blocked save drives the work item to failed/needs-review,
  not `complete`.

---

## Key IDs / where things live

- Site `gamesdesign.co.uk`, page `index`. Deployed: 5 sections ~34k, written
  2026-06-06 16:59 (unchanged — confirms no rebuild has saved).
- Child writer runs (index): `472eed7d` (06-15, dropped; parent `cd73eea6` holds
  the stub) and `4e0b339a` (06-06, deployed; parent holds the real output — A/B).
- Stub signature: `page_content.response.message = "Workflow completed
  successfully. Full result exceeded size limit."`, `completed_steps` a number.
- Save step config (`page-build-handler`): `save_page_sections`,
  `sections_metadata_field = page_content.response.sections_metadata`,
  `html_field = validation_result.clean_html`, `output_field = sections_saved`.
- Writer `complete` (live, v2): `{output_field: page_content}` ← the bug.
  Correct sibling pattern: `page-build-handler` `complete` uses
  `output_fields: ["sections_saved","deploy_result"]`.
- Code read this session: `coordinator.go` `extractWorkflowResult` L3351,
  `MaxResultSizeBytes` L3432, `extractWorkflowResultWithSizeLimit` L3434,
  `extractMinimalResult` L3475; `v3_site_actions.go` `CompilePageSectionsAction`
  L1612, `extractSectionFromMap` L1838; `save_page_sections_action.go`
  instrumented earlier (Info logs at the save boundary; on a live rebuild now
  shows `metadata_array_len = -1` then "no sections found", corroborating the
  stub from the running system).
- Not yet read: the snapshot/revert table for `agent_definitions` (Step 1 target).
- Empty in project mount: `002_intake_orchestrator.sql`, `003_site_classifier.sql`,
  `018_briefing_questionnaire.sql` (0 bytes).
- Docs updated this session: this file; `016_debugging_guide_v2_50.md` §9 (new
  entry "Child workflow result silently replaced by a stub — `output_field` vs
  `output_fields` …").
