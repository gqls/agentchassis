# HANDOFF 2026-08-07 — 189 and 204 are DONE (verified, still holding at v1.0.1261); start a fresh pickup here

Supersedes `HANDOFF_2026-08-06c_…` and everything earlier in this series. Nothing
is outstanding on either bug. Standing owner brief:
`HANDOFF_2026-08-05_next_bug_pickup.md` — re-read it, it is the job description.

## Both bugs: finished, and re-confirmed after the v1.0.1261 roll

**`bugs_open/204`** — the build path could not resolve a positional slot name
(`prose-0`), so a decomposed page could never be rebuilt and each miss filed junk
work items. Fix `13252f714` (Path 0: resolve by `page_components.component_id`
first, reusing 182's helper), council APPROVED `d3e232b8`.

**`bugs_open/189`** — a resolving save renamed the slot to its component's
function, so the locked-row guard missed and DUPLICATED the row. Fix `92e14493b`
(+ the `v3_site_actions.go` half at `1d11827c1`), council APPROVED `87444080`,
registered **PBP-035**, plus a config change (`slot_name_from`) applied 08-06.

Both were **induced, not argued**:

- Re-render path (item `b4de13fb`, orch `b807b035`): 4 rows not 5, positional
  names intact, locked `tool-2` keeping its row id AND its 2026-08-02
  `updated_at`, `rerendered=4 / carried=0`.
- Build path (item `996b9619`, orch `fa89217a`): both positional slots resolved
  (`ready=2, deferred=0`), prose rebuilt 1993→2358 b and 192→471 b in the
  requested voice, **zero** junk items, `pages.sections` unchanged, and the
  served page carries the new prose with the pre-fix opening returning **0**.

**Re-verified 2026-08-07 at v1.0.1261** (a later roll than the one they were
proven on): `load page slot identities` = 1 and `stored_slot_name` = 1 on
agent-chassis / business-intel / vet-intel, fabricated control 0; config keys
still present on both writer steps — and note they **survived another session's
`page-content-writer` update at 08-06 19:53**, which was the live risk, since
re-running that seed's stale full-workflow block would have reverted them; both
verified pages still hold their rows and slot names.

**Both files stay in `bugs_open/`** — owner direction 2026-08-06, which overrides
CLAUDE.md's `/bugs_closed/` bar. The closure evidence is written inside each
file. Do not `git mv` them; ask before moving any bug file.

## Next pickup, mechanically

1. `ls bugs_open/` — newest first. As of 08-06, `205` and `206` were filed by
   other threads (206 = entity-directory/entity-page/section-index pages have no
   real builder) and are the freshest.
2. Standing four before touching anything: `scripts/who-owns.py <n|slug>`
   (resolve number collisions by SLUG — several numbers name two bugs), `git log`
   the FILE the bug's §2 names, grep the live `.jsonl` transcripts for the bug
   number AND the code path, and check `site_work_items` for open work on the
   target.
3. Re-verify the defect against the live system before planning. Twice in this
   arc that changed the task.
4. Then the brief as written: plan, fix, council (platform/internal/pkg only),
   commit per task with a pathspec, keep the docs current, missteps to
   `WRONG_CALLS.md`.

## Two loose ends deliberately left (not forgotten)

- **`tool-recreation-handler` is a third `save_page_sections` producer with no
  `stored_slot_name`** (council `bug_historian`, LOW). Safe today only because it
  regenerates single-tool HTML with no structured slot identity to offer — a fact
  about that producer, not an enforced mechanism. Recorded in PBP-035.
- **The tri-state id-resolution judgement is now written inline at two call
  sites** (`plan_sections`, `rerender_page_sections`) with call-site-specific
  consequences. The architecture seat raised it twice at MEDIUM. If a THIRD
  consumer appears, factor the DECISION into one shared helper first. Recorded in
  `doc_notes d9d67807` and PBP-035.

## Method lessons from this arc (the transferable part)

1. **A row count proves the damage, never that work happened.** 189's own
   documented pass condition ("4 rows, not 5") is satisfied identically by "fixed"
   and "did nothing" — the re-render path has a carry branch. The discriminators
   were **row identity** (DELETE+INSERT ⇒ new ids) and the action's own
   `rerendered`/`carried` counters. Before recording a behavioural pass, name the
   inaction that would give the same reading. In `WRONG_CALLS.md` and memory.
2. **Read the verdict you cite.** I called 182's semantics "council-reviewed"
   when only its SUBMISSION existed — no `council_report` for corr `80fbbe7d`.
   One query would have shown it: `SELECT kind FROM diagnosis_artifacts WHERE
   correlation_id='<corr>' AND kind='council_report'`.
3. **"Sole consumer" is a query, not a memory.** `plan_sections` has two live
   consumers, not the one the file header implies.
4. **After making a dead path live, trace where it terminates.** 204's fix armed
   189's trap on a second path; found by tracing compile→save after the council
   round, and disclosed rather than left in the diff.
5. **Filing is not dispatching.** A `page_rerender` work item at `approved` sits
   for ever — nothing polls it. Publish to `system.agent.generic.requests` and
   confirm at the DB, because `kcat -P` exits 0 having sent nothing. And
   `spec.reason` selects the branch: without `section_data_resolved` the rerender
   takes the assemble-only path and never reaches the save.
6. **Check pod age against the ~300s post-restart drop window** before
   dispatching, especially right after a roll.
7. **A pathspec commit cannot exclude a same-file edit, in either direction.**
   Half of 189 was swept into another session's commit mid-write; nothing lost,
   and both commits say so. Say so — a reader of either alone sees half a change.
8. **The commit-msg trailer gate is right.** A placeholder
   `Council-Submitted: pending` is refused: the trailer is a join key and
   forward-only forbids an amend. Submit first, then commit.
9. **Production config writes are classifier-gated.** Prepare the exact command
   with its verification and backup, and hand it to the operator; don't fight it.

## Pointers

Bug files (both carry full closure evidence): `bugs_open/189_HANDOFF_2026-08-03_…`,
`bugs_open/204_HANDOFF_2026-08-05_…` · plans: `PLAN_2026-08-06_204_…`,
`PLAN_2026-08-06_189_…` · milestone read-out:
`SUMMARY_2026-08-06_two_bugs_that_unblocked_the_framework_rerun.md` · register:
PBP-035 in `docs026_concept_register/register/page-build-pipeline.md` ·
`doc_notes d9d67807` (`action/plan_sections`) and `c23ce8cb`
(`pipeline/page-content-writer`) · config backup:
`scratchpad/pcw_default_config_backup_20260806.json`.

**Unblocked by this arc:** the owner's 2026-08-05 instruction to rerun
loancalculator's copy through the framework in the H voice. The mechanism is now
proven on that site's own pages — `guide-how-loans-are-calculated` was rebuilt
that way as the canary.
