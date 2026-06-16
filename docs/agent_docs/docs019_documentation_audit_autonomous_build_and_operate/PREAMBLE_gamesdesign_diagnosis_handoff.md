# Diagnosis so far — gamesdesign.co.uk silent no-op rebuild
*(Paste this ABOVE the bundle when handing to a fresh chat. The bundle carries
the code, schema, and runtime evidence; this carries the reasoning state — what
we concluded, what we ruled out, and how — which the bundle format does not.)*

## The symptom
The gamesdesign.co.uk root `index` page rebuild reports success, but the live
page stays stale. It "presents as success" — work item completes, no error
surfaced — so it's invisible without inspecting the stored components.

## Two faults (this is NOT one bug)
1. **Generation shortfall (the root cause).** The page-content-writer
   regenerates ~3,000 chars of text for a page whose existing content is
   ~13,000–20,000. The content-regression guard in `SavePageSectionsAction`
   (save_page_sections_action.go) then CORRECTLY blocks the overwrite
   (`newTextLen < existingTextLen/4` → hard error). The guard is doing its job;
   the defect is upstream in generation.
2. **Status rollup (the visibility fault).** The blocked save is logged as
   `step save_sections failed` in `agent_error_log`, YET the `page_rerender`
   row in `site_work_items` is `complete`. A failed step rolls up to a completed
   work item — which is why the rebuild looks successful. Separate from fault 1
   and worth fixing regardless: a future genuine regression-block should be
   VISIBLE, not silently `complete`.

## What we RULED OUT (do not re-derive these — they were checked and falsified)
- **"Generated sections never reach save_page_sections."** FALSE. The runtime
  evidence (15 rows in `agent_error_log`) shows sections DO reach save every
  rebuild; save fails on the regression guard, not on missing input. The
  original task framing said "never reach save" — that hypothesis is dead.
- **"Check the persisted section status (ready/deferred/skipped) in the plan
  table."** Not possible. `site_plan_sections` has NO status column (verified by
  `\d`: its columns are `plan_id, page_name, ordering, component_name,
  component_version_id, palette_id, layout_id, typography_set_id`). The
  `Status`/`LLMFieldSpecs` machinery is computed AT RUNTIME in
  `plan_sections_action.go` and passed to the writer — never stored. So the
  planning decision is only observable in the runtime trace, not a column.

## What we CONFIRMED (with evidence)
- **The deployed page is content-rich and intact** — 5 healthy sections: hero
  (2,426), tool-list (8,951), guide-list (7,513), game-list (8,116),
  system-stats (7,369), ~34k total. So the ~3k regeneration is a fraction of the
  WHOLE page, not one failed section. The page is being PROTECTED by the guard,
  not damaged (the guard returns before the DELETE/INSERT).
- **The regeneration is short across the board**, not a single broken section.

## Leading hypothesis (strong, from the prompt config — see the bundle)
The content-writer's `process_sections_loop` calls `generate_content`
(`execute_llm_prompt`) ONCE PER SECTION, capped at **`max_tokens: 2000`**
(`claude-sonnet-4-6`, JSON output). ~2000 tokens ≈ ~8k chars of raw JSON before
tag-stripping; the deployed sections are ~7.5k–9k chars EACH. A single
2000-token call cannot reproduce an ~8.9k section, so across 5 sections the
regeneration is starved to ~3k. This is a CONFIG-level cause, not necessarily a
writer bug.

## The open discriminator (decide THIS next — needs runtime, not static code)
The prompt has a recreate branch: `{{if .existing_content.has_existing}}`
injects the original page content and instructs "adapt it." Which case holds
changes the fix:
- **recreate-mode NOT firing** → writer composes fresh against the 2000 cap with
  no original → fix = ensure recreate/re-adoption passes `existing_content`,
  and/or the cap is too low for these section sizes.
- **recreate-mode firing** → model has the ~8k original and still returns ~3k →
  the cap is truncating the adaptation, or the per-section JSON output structure
  can't carry the volume → fix is on the cap / output structure.

**Next evidence (from the content-writer's run logs for one failed
orchestration, NOT the static prompt):** (a) was `existing_content.has_existing`
true? (b) how many sections did `sections_ready` contain? (c) the actual
per-section generated lengths. The prompt shows the CEILING; the logs show what
was hit.

## Guardrails (the tempting wrong moves)
- **Do NOT loosen or remove the content-regression guard.** It is correct and is
  the only thing currently protecting the live 34k page from a 3k overwrite.
- **Do NOT raise `max_tokens` blindly** before confirming the discriminator — if
  recreate-mode simply isn't firing, the cap is a red herring and the real fix
  is passing the existing content through.
- **Do NOT conclude the generation failure mode** (fresh-vs-recreate, cap-vs-
  structure) without the run logs. The two hypotheses above were both reached by
  checking evidence and discarding what it contradicted; hold the same bar.

## Where the full detail lives
`RUNBOOK_gamesdesign_silent_norebuild_bug.md` (the bug runbook): evidence,
reproduce/inspect commands, the bundle-2 build command, and the
verification-when-fixed criteria for both faults.
