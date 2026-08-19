# PLAN — bugs_open/301: refuse an owned page at `load_page_record`, before the LLM writer runs

**Lane opened 2026-08-19.** Bug file:
`bugs_open/301_HANDOFF_2026-08-18_page_build_handler_runs_the_llm_writer_and_link_resolver_before_the_owned_page_guard_so_the_work_is_thrown_away.md`
Filed by `vigilant_designer_offer_analysis`, explicitly offered unowned ("Take `bugs_open/301`",
their HANDOFF_2026-08-18 §"immediate next moves"). `scripts/who-owns.py 301` run 2026-08-19:
the two active lanes around it (`bugfix_277_required_fields_repair`, `vigilant_designer_offer_analysis`)
have both shipped their halves (Tier 1 status change; the filing) and neither claims the ordering fix.

## What we are fixing (and what we are NOT)

**Fixing: the ordering.** `page-build-handler`'s owned-page refusal lives only in
`save_page_sections` — the last step that touches the page — so a build item aimed at a
`pages.rebuild_policy='owned'` page runs `page-content-writer` (an LLM call) and
`internal-link-resolver` to completion and is *then* refused. Bug 301 measured 39 discarded
chains in 2.5 h on one site; re-verified live 2026-08-19 ~10:20 UTC (chains still burning
post-roll, `NOTES` has the rows). **146 open findings on owned pages are routed at
`page-build-handler` right now** [MEASURED 2026-08-19 12:0x UTC, live table] — each is a
guaranteed wasted chain under the current ordering.

**Not fixing here** (owned elsewhere, per their handoffs):
- Tier 1 (refusal writes `wont_fix`, not `failed`) — SHIPPED by the 277/083 lane, live on
  `v1.0.1314` (migration 480 + Go halves). This plan *composes* with it and relies on it.
- Tier 2 (routing MODIFY-shaped findings to a `field_updates` producer) — answered
  "different agent" by `copy_quality_two_stage`; partly superseded by the 184 lane's `473`.
- Candidate 3 (producers should not file generic content items against owned pages at all)
  — the real upstream repair, entangled with Tier 2; explicitly left on the table.

## The fix, and why this shape

**Opt-in early refusal at the page-load seam.** New config key `refuse_owned_page` on the
`load_page_record` action (Go), plus migration `488` (config half) that sets it on
`page-build-handler`'s `load_page_record` step and gives that step
`error_step: mark_item_failed`.

When the key is set and the loaded page is owned:
1. `pageIsOwnedForGuard` decides — the **single** ownership predicate (unified for
   `bugs_open/208` at the council's `reuse_agent` seat's insistence; no second predicate).
2. `emitOwnedPageReviewItem` files the same deduped `owned_page_review` row the save-path
   guard files (`refused_by='load_page_record'`; same `item_key` namespace, so it converges
   with reconcile's and save's rows rather than competing).
3. The action returns an error **leading with `ownedPageSkipReasonPrefix`** — the exact
   marker `update_work_item_status`'s `owned_page_refusal_status` (Tier 1, migration 480)
   matches in `__step_error.message`. The refusal therefore lands `wont_fix` with zero new
   vocabulary, via `error_step → mark_item_failed`, which already carries the Tier 1 key.

**Decisions and their reasons:**
- **Go predicate, not a config conditional** — bug 301's own candidate 2 caveat: a config
  predicate duplicating a Go one is the drift class this estate keeps filing bugs about,
  and a conditional cannot emit the review row or set `__step_error`.
- **Opt-in, default OFF** — owner ruling 2026-08-02 §2 (new authority on a shared seam
  ships as an opt-in field with the unsafe default OFF). The carrier census is the clincher:
  exactly two live agents carry `load_page_record` — `page-build-handler` (must refuse) and
  **`tool-recreation-handler` (the tool pipeline — the legitimate owner of owned pages,
  which must NEVER refuse them)**. A default-ON rule in code would have broken the tool
  pipeline; the opt-in makes the unsafe-for-whom question per-caller.
- **Declared in `ConfigKeys`, not `Optional`** — it is a setting (boolean read from config,
  like `skip_if_missing`), not a data reference. Named trade-off: the RFC_022 optional-key
  budget counts `Optional` only, so this key is not budget-counted; disclosed in the council
  submission so the seat can object. `load_page_record`'s Optional count stays 3 (of N=10).
- **The save-path guard STAYS** — it is the backstop for any other caller and for the
  fail-open window below; removing it would re-open `bugs_closed/295`.
- **Fail-open on an unreadable policy** (matching `pageIsOwnedForGuard`'s posture): here it
  is *cheap* fail-open — the save-path guard is still downstream, so an unchecked page gets
  one wasted chain, not a clobber. (The assemble-seam guard's fail-open window is loud
  because nothing sits behind it; this one has a backstop and a Warn log.)

**The structural claim that licenses the move** (what the 090 run is verifying): every
content-writing path in `page-build-handler`'s live workflow ends at `save_page_sections`.
Read from the live `agent_definitions` row 2026-08-19: the graph's only write path is
`validate_content → save_sections → update_status → spawn_rerender_agent → deploy_page`,
and the non-write arms (`check_page_found` else, `mark_no_ready_sections`,
`mark_writer_skipped`) park the item without touching the page. So an early refusal refuses
exactly the population the save guard refuses today — it cannot cut off a legitimate
owned-page write because the workflow contains none.
090 intake `7281193f-59c2-489a-a9f2-fd4d58408cf5`, run `dd61df1b-0d93-46e6-9065-1e0b9623379a`.

## Named behaviour changes (the honest list for the council)

1. **Genuine `load_page_record` errors** (DB error, malformed authoritative id) now route
   `error_step → mark_item_failed` (item visibly `failed`, attempt counted) instead of
   failing the workflow. This aligns the step with every sibling
   (`plan_sections`/`call_content_writer`/`save_sections`/`deploy_page` all route there) and
   is active from config-apply time, on the old binary too.
2. **Owned-page items that would have taken a no-op arm** (`mark_no_ready_sections` /
   `mark_writer_skipped`, both parking at `needs_human_review` with a no-op message) are now
   refused up front: `wont_fix` + `owned_page_review` row instead. Both outcomes are
   visible parks; the refusal is the truer record (the review row's `spec.fix` names the
   route that works) and `wont_fix` keeps the promoter's floor blind to it (Tier 1's whole
   point). Note the historical "74 completions on owned pages" in the bug file is NOT
   evidence of a live success path: `rebuild_policy` is mutable and those joins are
   query-time (the 277 lane's §8 lesson) — and the current workflow has no write path that
   bypasses the save guard.
3. **`tool-recreation-handler`: no change by construction** — it does not carry the key.

## Ordering (image vs config) — order-TOLERANT, deliberately

- Config before binary: old `LoadPageRecordAction` never reads the key (spec has
  `CheckConfig`, no `StrictConfig` — `platform/validation/workflow.go:185-195` warns once,
  never rejects). `error_step` is honoured by the existing coordinator (only change 1 above
  activates early).
- Binary before config: the code is dormant until the migration names the key.
So migration 488 is applied immediately after commit (removing the LANDMINES "loaded gun"
pending state), and the refusal itself activates at the next chassis roll.

## Verification plan (positive AND negative control, per the bug file's own §How to verify)

After the next roll, on a fresh dispatch burst:
- **Owned page** (positive): a content item at an owned page → `wont_fix` with
  `result ? 'owned_page_refusal'`, an `owned_page_review` row with
  `refused_by='load_page_record'`, and **no `page-content-writer` orchestration spawned for
  it** (check `orchestration_states` children).
- **Generic page** (negative control): writer runs, page saves, item completes. Without
  this, "no writer ran" is equally consistent with having broken the writer.
- Binary probe: `refuse_owned_page` PRESENT in `/proc/1/exe`, alongside the long-lived
  `OWNED_PAGE_GUARD` control and a nonsense-needle ABSENT control.

## Phasing

1. ✅ Verify bug still valid + no owner (this file, NOTES).
2. ✅ File 090 (running).
3. Go change + tests (`load_page_record_action.go`, new test file).
4. Migration `488` + `_ROLLBACK`, guard + verify blocks in the 480 house style.
5. Council submission (097) — commit with `Council-Submitted:` trailer, per the norm.
6. Commit narrow pathspec; register the mechanism in the concept register same commit.
7. Apply 488; verify live config; leave the roll to the fleet cadence (no same-tag rebuild).
8. Post-roll verification per above; then move 301 towards closed (fixed AND live bar).
