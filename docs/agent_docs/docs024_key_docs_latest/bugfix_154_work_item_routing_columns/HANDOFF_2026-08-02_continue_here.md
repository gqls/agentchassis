# HANDOFF 2026-08-02 — bugfix-19 session: dispatch fixed end-to-end, content restored, guard pending roll

**Read this first; then `NOTES_…` (2026-08-02 entries) for evidence.** The
2026-07-31 handoff in this directory is SUPERSEDED (154 closed).

## State: what is DONE and LIVE
- `bugs_closed/154` — routing columns; witnessed end-to-end; CLOSED.
- `bugs_closed/176` — selector↔loader disagreement; `284`+`285` applied; fleet
  dispatching fairly (verified: claims in exact FIFO order across 6 sites).
- `286` — triage: `0733a7a4` wont_fix (spurious), `93f2a3b7` dependency cleared,
  crosslink ran (link correct on all 3 slots).
- `287` — restores applied byte-exact from history: robot-hands 4655,
  fundamentallyai 32444, vetcomparison 7206+3043. Dispositions + not-restored
  list in `bugs_open/178`'s update block.

## OPEN — in priority order for whoever continues
1. **Verify the 2 rerender items** `item_key LIKE 'restore_287:%'` (robot-hands
   how-to-specify-a-gripper, vetcomparison /about) complete AND check the
   ARTEFACT: robot-hands rendered generic-text-block must contain BOTH
   `ISO 9409-1` AND `gripper-safety-factor-calculator`; vet faq ~3× its 2197.
   ⚠ fundamentallyai `/tools/review-council-simulator.html` must NOT be
   rerendered until the restored content_data shape is checked vs the renderer
   — its live render is the only known-good copy.
2. **The shrink guard is INERT until an image roll** (commit `2da3e08e5`).
   After the next roll, pod-grep BOTH replicas in one exec:
   `strings /app/agent-chassis | grep -c "SECTION SHRINK"` (expect ≥1) +
   control `grep -c "CONTENT REGRESSION BLOCKED"` (expect ≥1). Then INDUCE:
   fire a save that shrinks a big slot >50% and watch it refuse + emit the
   refusal work item.
3. **Council verdict `e64f8576`** (shrink guard). Find by payload:
   `SELECT current_step,status FROM orchestration_states WHERE
   collected_data->'input_data'->>'fix_correlation_id'='e64f8576-e056-4c2d-87f9-4e27c65aee08';`
   REVISE → answer into the code, resubmit with RESUBMIT_CORR.
4. **`bugs_open/178` root cause** — why does a link-insertion item regenerate
   the whole section? Undiagnosed; run `090`. Candidates 1 (edit-not-regenerate)
   and 3 (emit the delta) open. Also: relojistas' deleted DefinedTermSet slot
   (snapshot `b0e119a4`, needs a slot_name decision); vetcomparison's discarded
   same-day edits (re-detect via discovery).
5. **`bugs_open/177`** — tool-generator raises spurious tool_content items
   (8/8 failed identically). Unstarted. The 7 remaining rows sweep with its fix.
6. Watch list: `bugs_open/169` part A (spawn hang) untouched · scheduler
   pre_query vs selector asymmetry (`triaged`-only + locked_at) [UNMEASURED] ·
   loader's dependency subquery is SITE-SCOPED (cross-site depends_on never
   resolves; 285 copies it deliberately — fix means Go + both queries together).

## Landmines specific to this work
- 284+285 are only safe TOGETHER (FIFO + selector/loader agreement) — never
  revert one alone.
- A dependency releases ONLY on complete/verified — wont_fix blocks for ever;
  that is why 286 cleared the array instead of faking complete.
- Dispatch "quiet spells" are never baseline: read
  `collected_data->'load_items'` (`item_count:0` + `rows_dropped:0` = the 176
  signature), not time-since-last-claim.
