# HANDOFF 2026-08-04 — 178's fix is DONE and LIVE; the only thing left is blocked on an unrelated bug (192)

**Read this first; then `NOTES_work_item_routing_columns.md`'s 2026-08-03/04
tail (three entries: "178 FIX IMPLEMENTED...", "dispatch RESULT...") for the
full evidence trail. `bugs_open/178` carries its own final update too.**

## State: 178's fix is DONE, LIVE, and proven correct where it matters

- Root cause (confirmed 2026-08-03): `content_rewrite` items never set
  `spec.mode`, so the writer got a rewrite instruction with nothing to edit
  and fabricated a replacement section, dropping most of the page's prose
  (measured -59% on one real page).
- Fix: a third `spec.mode` value, `"edit_live"`. New Go step
  `load_current_section_content` (between `check_has_ready_sections` and
  `spawn_content_writer` in `page-build-handler`) attaches the page's current
  `rendered_html` to each ready section, matched by slot name, when and only
  when the item opts in. Both live `content_rewrite` emitters
  (`create_tool_cross_link_items.go`, `apply_gap_plan_action.go`) now set it.
  Default OFF for everything else — proven by two passthrough tests, not
  just argued.
- Committed `08d0515f3` (code+SQL+register) + `0a2a94b89` (tag bump).
  Registered as **PBP-028** in `docs026_concept_register/register/page-build-pipeline.md`.
- Built `v1.0.1244`, deployed by the owner's whole-fleet release as part of
  `v1.0.1247`. **Pod-verified on both replicas** (binary symbol grep, not the
  tag — `LoadCurrentSectionContent: attached current content for edit mode`
  present, negative control absent).
- Migration `299_edit_live_channel_for_content_rewrite_writer.sql` applied
  by hand (both `DO $$ RAISE EXCEPTION $$` verify blocks passed) and
  recorded (`run-migrations.sh --record-only`).
- **Live-tested on a real production dispatch**, not a synthetic one: the
  two crosslink items `bugs_open/178`'s own note left parked for this
  purpose. Confirmed directly from `orchestration_states` that
  `section_plan.sections_ready[0].existing_content_html` held the page's
  exact current prose (the CMA-compliance content on
  `vetcomparison.uk/guides/cma-compliance`), matched to the right slot,
  unmodified. **This is the part of the mechanism 178 was about, and it
  works.**

## What's NOT done, and why

The full before/after `content_data`-length assertion (178's own "how to
verify a fix" test) could not be completed, because **both live dispatches
then failed one step later inside `page-content-writer`'s own
`select_sections`/`process_sections_loop`** — a completely separate,
pre-existing bug, filed as **`bugs_open/192`**. Evidence it is NOT this fix's
doing: the same failure hit an unrelated tool page on a different site in
the same run, and `orchestration_states` shows the failure wave started at
2026-08-03 21:00 — hours before this fix's image had been deployed anywhere.
**No content was lost** (the failure is upstream of any save) and the two
work items are `failed`/attempt_count=1/non-terminal — safe to retry.

192 itself is NOT diagnosed (a `090` run is owed, not run — flagged rather
than chased, to keep this session's scope to 178). Root cause, per the
evidence gathered: `select_sections` is `extract_fields` trying
`resolved_links.response.link_resolution.sections_ready` (present but
explicitly `null` in the failing runs) before falling back to
`input_data.section_plan.sections_ready` (which DOES hold the real data at
the same instant, confirmed directly). Why the fallback doesn't fire is
unexplained — `ExtractFieldsAction` does null-check its candidates on its
face, so there's a real puzzle here for whoever picks it up.

## OPEN — in priority order

1. **Diagnose and fix `bugs_open/192`.** This blocks 178's final acceptance
   test AND, per the measured failure rate (11-14 failures/hour for several
   hours on 08-03 night), is plausibly blocking a large fraction of ALL page
   content builds right now, fleet-wide, whenever
   `resolved_links.response.link_resolution.sections_ready` comes back
   explicitly `null` rather than absent. Recommend the `090` diagnosis
   trigger given it's cross-cutting and non-obvious — this handoff's author
   read `ExtractFieldsAction` far enough to rule out the obvious theory (no
   null-check) but did not find the real cause.
2. **Once 192 is fixed, re-dispatch the two parked items** and complete
   178's own verification: `SELECT length(content_data::text) FROM
   page_components WHERE page_id IN ('d8c51ace-9286-4e53-95f9-efd02152568b',
   '2a347990-c152-4fa2-8acb-a39a5f74f4a9') AND slot_name='generic-text-block'`
   — expect ~6034 and ~3637 respectively **plus** roughly the length of one
   inserted link anchor (~90-150 chars), NOT a wholesale replacement. Item
   ids: `9e9ec430-ff92-4264-83cc-6072840faad8` (guide-cma-compliance),
   `18bc832c-c937-4608-9a05-718772d44c88` (guide-independent-strategy), both
   `vetcomparison.uk`, both `status=failed attempt_count=1` currently — just
   needs a fresh `build-dispatch-loop` fire for site `72b9e3a6-872f-4528-a6d6-7f205ea60f4d`
   once 192 no longer breaks the writer.
3. **`build-dispatch-loop` has no scheduled task at all** — confirmed via
   `SELECT name, target_agent_type, enabled FROM scheduled_tasks` (only
   `report-dispatch` and `diagnose-pipeline-trigger` exist). This session
   fired it by hand (`kcat` with `config.agent_type:"build-dispatch-loop"`,
   `input_data:{site_id, domain}` — confirmed its `input_contract` first).
   Not fixed, just flagged — third instance of the "detection works,
   schedule/dispatch doesn't" pattern family already in memory. If this is
   why the queue has felt slow/stalled to other lanes, this is why.
4. **Council submission stalled**, not rejected: `Council-Submitted:
   97ebadcf-bbe6-485f-8231-ff16fc4e679f` on the 178 commit; the run reached
   `review_constitution` at 20:09:59Z on 08-03 and never advanced, no
   `council_report` was ever written. Advisory only — doesn't block — but
   flagging for whoever next looks at council-gate reliability. Do NOT
   re-submit for 178 unless you have a reason to think a fresh submission
   would fare differently; the change itself is unrelated to the stall.

## Landmines specific to this lane (carry-forward + new)

- All landmines from the 2026-08-03 handoff (shrink guard, dependency
  release rules, dispatch quiet-spell reading, `orchestration_states`
  retention) still apply unchanged.
- **NEW — a `content_rewrite` item created BEFORE the emitter fix has no
  `mode` key, so releasing it as-is still hits the OLD destructive path.**
  If you find more parked/gated crosslink or gap-plan items predating
  `08d0515f3`, patch their `spec` with `mode:"edit_live"` before releasing
  them, the same way this session did for the two it used.
- **NEW — `select_sections`'s null-fallback (bugs_open/192) can fail ANY
  content build**, not just `edit_live` ones. If you dispatch anything
  through `page-build-handler` and it fails at `process_sections_loop` with
  `key 'sections_ready' not found`, this is very likely 192, not a new bug —
  check 192 before filing a duplicate.
- **NEW — `bugs_open/087`'s error string is not unique to 087.** Two
  different causes now produce the identical `sections_for_render.sections_ready
  not found` message (087: `page-rebuild` supplies no section_plan at all;
  192: build-handler path, null link-resolution). Read which AGENT the
  failing orchestration is (page-rebuild vs page-content-writer via
  build-handler) before assuming which bug you're looking at.

## Cold-start pointers

- This file → `NOTES_work_item_routing_columns.md` (08-03/04 tail, three
  entries) → `bugs_open/178`'s final update → `bugs_open/192` (new, evidence
  only) → register entry PBP-028
  (`docs026_concept_register/register/page-build-pipeline.md`).
- Commits: `08d0515f3` (178 fix), `0a2a94b89` (tag bump), `75fceb501` (192
  filing).
- Migration: `docs/agent_docs/sql_for_agents/299_edit_live_channel_for_content_rewrite_writer.sql`,
  applied and recorded.
