# PLAN 2026-08-22 — bugs_open/342 residuals: the refusal half, and arming what is built

**Lane:** `bugfix_342_absent_required`. Took ownership 2026-08-22 — the bug was filed UNOWNED by
the `bugfix_260_render_fallback` lane, whose own work on it (the seam report + the escalation)
is DONE, APPROVED (council `bb7f5d0e`, round 6) and LIVE. This lane picks up what that file
says it stays OPEN for.

## What was verified before planning (all 2026-08-22, first-hand)

- **The whole 342 detection+escalation stack is LIVE on v1.0.1323**, including the editor-route
  escalation the bug file's banner says is "NOT in v1.0.1322 and inert". Evidence: commits
  `cd90e8b27`, `65f1b0b95`, `af4743464` are all ancestors of the v1.0.1323 build stamp
  `70e7b4f9c` (`git merge-base --is-ancestor`), and the stamp was probed IN THE BINARY on both
  replicas (`grep -aq "70e7b4f9c" /proc/1/exe` → present; nonsense control → absent). The bug
  file banner is therefore STALE and gets corrected in this lane's commit.
- **`record_absent_required_fields` is armed NOWHERE.** `agent_definitions` scan for the key:
  0 rows. The chrome escalation is live code behind an opt-in nobody has opted into.
- **No refusal exists anywhere for the absent-required case** (code read: the editor routes
  emit the work item and then persist the blank section to the live page anyway).
- **`section-editor.apply_edit` has NO `error_step`** (live row read). A refusal failing that
  step meets `bugs_open/344`'s completion-trample. The live page is protected regardless —
  the write never happens — but the DRIVING item's terminal status may read `complete` until
  344 lands. Stated as an interaction, not fixed here; 344 owns it.
- **Populations, re-measured** (the bug file's figures had moved):
  - No-schema components: now **100 of 283** active — but **95 of the 100 are
    `component_level='tool'`**, which are self-contained HTML with no LLM fields BY DESIGN
    (the rerender gate's `isSelfContainedSection` codifies exactly this). The genuinely
    exposed class is **5 non-tool components, one page_components usage each, 2 with template
    placeholders** (`report-request-form`, `audience-check-form`). "The hard part" from the
    bug file's §5 has largely dissolved; what remains is small data work, not chassis work.
  - Chrome store (`site_components`): rows reference only `site-header`/`site-footer`/`head`
    class components which declare **0 required source:llm fields**, so arming the chrome
    record fires on **0 rows today** — measured with a NON-VACUOUS join (candidate-pairs
    count checked; see NOTES for the vacuous first attempt).
  - `page_components` writer-side census: 131 rows missing a required llm field, dominated by
    `ported-page/body` on `build_status='deployed'` rows — the post-deploy check's own
    population, which is already filing items for them (45 complete / 30 needs_human_review).

## Decisions

1. **The remaining fix is the REFUSAL HALF at the section-editor persist switch, opt-in.**
   The two editor routes write `rendered_html` straight to an already-live page with no
   `validate_content` between. Today, when the seam publishes `AbsentRequiredFields`, they
   file the item and ship the blank anyway. The refusal declines the persist and leaves the
   live section untouched. Shape per owner ruling 2026-08-02 §2: a new ConfigKey
   `refuse_absent_required_fields`, unsafe default OFF, fail-open on a mistyped value —
   exactly the `refuse_mistyped_llm_fields` precedent, in the same file.
   - **Placed at the ONE persist switch in `ApplySectionEditAction`, not in the two branch
     helpers** — the file's own idiom (link repair, envelope refusal) and the reason: a future
     edit branch inherits the gate instead of re-remembering it. The helpers publish
     `AbsentRequiredFields` on the outcome; the action gates once.
   - The work item is still filed BEFORE the refusal (emit happens inside the branch), so a
     refused edit leaves a queue entry saying why.
   - RFC_022 check: opt-in ✓, unsafe default OFF ✓, zero live consumers name it (scan run:
     0 rows) ✓ → not architecture-scope. Normal council gate.
2. **Arm `record_absent_required_fields: true` on the 7 live `render_site_components` steps
   now** (appliable migration, not _HOLD — the Go half is verified live). Measured to fire on
   0 rows today, which makes it FREE to arm and closes the door before a chrome component
   that declares required fields is ever adopted. The alternative — leave it OFF until it is
   needed — is the "mechanism rotting unexercised" failure the owner has ruled against
   requiring.
3. **Arm the editor refusal via a `_HOLD` migration** (ordering-critical: the Go half must
   roll first; a key naming behaviour the binary does not have is a no-op that LOOKS applied —
   502's exact reasoning, same shape).
4. **Out of scope, stated so the choice is a choice:**
   - Refusal at the SEAM: ruled out by owner 2026-08-02 §2 (new authority over content that
     renders today at sites that never asked); the per-path shape is the compliant one.
   - The 5 no-schema non-tool components: data work, one usage each; recorded in the bug
     file with the census so the next reader sizes it correctly. No chassis change closes it.
   - `bugs_open/344`'s completion-trample: owned elsewhere; interaction stated in code
     comment and bug file.
   - The six unwired call sites: re-verified individually (see NOTES) — each has a mechanism
     reason (raw templates with no component, a contact-info block with no schema, a stitched
     TEMPLATE whose content arrives later, audit probes that remove fields on purpose).
     No change.

## Edits (council submission mirrors these)

1. `platform/orchestration/actions/mistyped_llm_fields_gate.go` — add
   `refuse_absent_required_fields` key + `refuseAbsentRequiredFields()`, same fail-open
   semantics as both siblings.
2. `platform/orchestration/actions/section_editor_actions.go` — `sectionEditOutcome` gains
   `AbsentRequiredFields`; both render branches set it; ONE gate in `ApplySectionEditAction`
   before the repair/persist path refuses when armed.
3. `ApplySectionEditInputSpec.ConfigKeys` += the key (same edit as 2, same file).
4. Tests in `render_seam_absent_required_test.go` + a section-editor-side test: armed refusal
   refuses and names the fields; unarmed passes byte-for-byte; fail-open on mistyped values;
   mutation-proven (temporarily invert the guard, watch the test fail).
5. `docs/agent_docs/sql_for_agents/NNN_bugfix_342_arm_chrome_absent_required_record.sql` —
   appliable now; DO/RAISE verify blocks, not bare SELECTs.
6. `docs/agent_docs/sql_for_agents/NNN+1_bugfix_342_arm_editor_absent_required_refusal_HOLD.sql`
   — _HOLD until a chassis image built from this commit has rolled (verify recipe inside).
7. `bugs_open/342_…md` — correct the stale banner (v1.0.1323 evidence), update the residual
   list, record the re-measured populations, name this lane as owner.
8. Register/STY-057 touch + LANDMINES only if a new trap is actually found (none yet).

## Verification plan

- Unit tests above; `go build ./...` against `git archive HEAD` shape (shared-tree rule).
- After the next roll: probe the binary for the new refusal literal, then apply the _HOLD
  migration, then a live canary: an edit against a component with a required llm field
  deliberately absent must refuse (item filed, live section byte-identical), and a clean edit
  must persist (positive control).
- After arming the chrome record: re-run the site_components census; expect 0 items filed and
  0 seam reports — and a DEMAND control (a chrome rerender must have actually run).
