# HANDOFF — 2026-08-19 (evening): 301 is CLOSED. Nothing is owed on this lane except one owner decision

**Lane:** `bugfix_301_owned_guard_ordering` — bug file NOW AT
`bugs_closed/301_HANDOFF_2026-08-18_page_build_handler_runs_the_llm_writer_and_link_resolver_before_the_owned_page_guard_so_the_work_is_thrown_away.md`
(moved this evening; resolve by SLUG — several numbers name two bugs on this tree).
**Read this file, then `NOTES_owned_guard_ordering.md` from the bottom** (session 3 entry holds the
closing measurements with their queries). PLAN = decisions and reasons; RUNBOOK = every query with
its gotcha; `SUMMARY_2026-08-19_owned_guard_ordering.md` = the read-aloud milestone.

---

## 0. STATE IN ONE PARAGRAPH

`page-build-handler` used to run the LLM writer + link resolver and only THEN refuse an owned page
(at `save_sections`, the last step). The fix — opt-in `refuse_owned_page` on `load_page_record` +
migration `488` — is committed (`6be66bceb`), applied (488 ledgered 11:05:25Z), **live on all 22
chassis-image pods at one digest (`v1.0.1316`)**, **proven on live demand both ways** (4 owned
refusals at load, 0 writer children, 1 fresh `refused_by='load_page_record'` review row; 2 generic
builds completed writer→save→deploy; 0 save-path refusals), and **council APPROVED at round 2**
(corr `c7bc1b9e-97c8-4f3e-8a4f-b3a7029505ee`, 16:19:04Z; every advisory answered by an
independent check — see NOTES session 3). The bug file carries a dated closing section and sits in
`bugs_closed/`. 016b has a §10 row and a §9 pattern ("a guard at the LAST step makes every refusal
cost the whole pipeline"). `MEMORY_closed.md` has a line. The save-path guard was deliberately KEPT.

Commits this lane made: `6be66bceb` (fix), `25ca816c7`/`5949d9ce3`/`4384ebe12` (docs),
`1c16eb692` (side-find: budget-cron parity), and the close commit (this evening — `git log -- <bug
file path>` for it).

## 1. THE ONE OPEN ITEM — an OWNER decision, flagged not taken

**Candidate 3 of the bug file — the real upstream defect.** Twelve producer sites in
`platform/orchestration/actions/` hard-code `handler_agent='page-build-handler'` for content
findings without reading `pages.rebuild_policy` (`apply_adoption_plan_action.go:693`,
`apply_gap_plan_action.go:780`, `create_tool_cross_link_items.go:312`,
`flag_page_image_rebuild_action.go:182`, `load_work_item_actions.go:236-240,264`,
`render_directory_action.go:429`, `reconcile_section_data_action.go:209`, …), so owned pages keep
accumulating findings that can only ever be refused — **142 open at ~20:55Z 08-19** (84 failed /
36 unresolved / 13 needs_human_review / 9 detected; re-run, keep the split — RUNBOOK "queue that
guarantees future waste"). It is now the "not addressed here" footnote of TWO closed files (295 and
301) and **no open bug file carries it as its subject** (`grep -l rebuild_policy bugs_open/*` —
208 is the rebuild-route sibling, not this).

Options put to the owner in `README_where_we_are.md` (bottom): **(a)** file it as its own small
`bugs_open` entry, cross-referencing the Tier 2 / `copy_quality_two_stage` exchange; **(b)** hand
it to that exchange. Lane recommendation: **(a)**. If the owner picks (a), the filing session
should: grep both bug dirs again first (a day is a long time here), name the 12 producers, state
the dedup key shape per the 2026-08-02 ruling, and cite 295 + 301 as the two files whose cost it
explains. `scripts/who-owns.py` the Tier 2 lane before routing anything at it.

## 2. IF YOU ARE HERE BECAUSE THE BUG CAME BACK

Re-run the RUNBOOK's post-roll verification with BOTH controls. The traps, all hit on this lane:
- `refused_by='load_page_record'` can read ZERO for ever — dedup per page; verify at
  `collected_data->'__step_error'` (`failed_step` + `OWNED_PAGE_GUARD`) and at the ABSENCE of a
  `page-content-writer` child.
- `error` is NULL on a routed failure — key on `__step_error.message`, never `error LIKE`.
- The step key is `save_sections`, not the action name `save_page_sections` — read the step names
  off the live row before keying `collected_data`.
- Never classify by joining `pages.rebuild_policy` (mutable).
- Binary probe, per pod, with both controls; the provenance line scrolls.
- `488` is applied + ledgered: do NOT edit it. Any new `agent_definitions` migration from this lane
  OPENS with `snapshot_agent()` (WRONG_CALLS 2026-08-19).
- The 090 run `dd61df1b` ended UNVERIFIABLE — cite neither way.
