# HANDOFF — 2026-08-02 — `bugfix_136_sibling_link_repair` lane · continue here

Written for a cold start in a new session. **Read this file, then `bugs_open/180`.**
Everything below is committed; nothing is left dirty in the tree by this lane.

---

## 1. State in one paragraph

`bugs_open/136` (**section-editor slug**) is **CLOSED, LIVE on v1.0.1229, pod-verified on both
replicas with a negative control**, council-APPROVED at round 1 (`0275f9c2-035f-4c9e-8a50-83836dfeffd9`),
registered as **LNK-027**, and moved to `bugs_closed/`. On the way out it produced a NEW bug,
**`bugs_open/180`**, which **reverses the next step this lane had written down**. If you do
only one thing with this handoff, make it that reversal.

## 2. What shipped (all live on v1.0.1229)

| thing | where |
|---|---|
| `repairComponentHTMLBeforePersist` — one-component dead-link repair, a *wrapper* over `repairSectionLinks` | `platform/orchestration/actions/component_link_repair.go` |
| wired into `ApplySectionEditAction` (both edit types) and `CreateReportPageAction` | `section_editor_actions.go`, `create_report_page_action.go` |
| `writeLinkRepairSkipLog` extracted (it existed twice; this would have been the third) | `component_link_repair.go`, callers updated in `save_sections_link_repair.go`, `rerender_link_repair.go` |
| `check_unrepaired_component_write` — advisory pre-commit rule | `scripts/pattern-check.py` |
| `sectionEditOutcome` — the swap's `UPDATE` moved to the caller, so the action has ONE persist point | `section_editor_actions.go` |

Commits: `66998d300` (fix), `a1aa1d421` (council responses), `d4fef13db` (summary),
`41138cacb` (close + files 180), `d289933c5` (landmine + 016b).

## 3. ⚠ THE ONE THING THAT CHANGED: do NOT wire the seam into the tool writers yet

`bugs_closed/136` says — in its own fix-candidate list, and this lane repeated it — that the
ranked next step is to give `create_tool_component_action.go` and `deploy_tool_action.go` the
same repair. **That is now the wrong order.** `bugs_open/180`: the shared `RepairPageLinks`
reads `href="' + q.link + '"` (JavaScript building an anchor at runtime) as an anchor with an
**empty** href and unlinks it — deleting a working link from the program, with output that is
still valid JavaScript. Tool markup is precisely where JS-built anchors live.

**Order: fix 180, then the tool writers.** Reproduce 180 in three lines (probe in the bug file;
run it against `platform/orchestration/datahelpers`, then delete the probe).

## 4. What is still open, ranked

1. **`bugs_open/180`** — the JS-anchor corruption. Fix candidate 1 (reuse LNK-025's tag
   scanner, which already jumps `script`/`style` spans) is almost certainly right. Exposure
   measured: **1 component, 1 site** (`vetcomparison.uk` / `tool-cma-obligation-checker`), not
   covered by any runtime-fill marker on its page — **and that is the size of one SPELLING**,
   not of the class (a template literal `href="${url}"` takes the phantom arm instead).
2. **The tool-markup writers** — 7 of the 35 census hrefs sit in tool-shaped slots. They are
   **deliberately not allow-listed** in `check_unrepaired_component_write`, so it keeps firing
   on them. Do not silence it; that would turn a live debt into a false all-clear.
3. **`architecture_review/RFC_008`** — the mandatory write seam (candidate 3). Four council
   seats converged on "advisory is the wrong ceiling". The RFC states the case against too,
   and names the measurement that would settle it: **nobody has ever checked whether advisory
   `pattern-check` findings are read and acted on.** That measurement is cheap and it governs
   every rule in that script, not just mine.
4. **The standing stock is untouched** — 18 unlinkable hrefs are live 404s today. Detection
   belongs to `bugs_open/116` (the phantom-link check has never run on any site).

## 5. Verification recipes (all in `RUNBOOK_sibling_link_repair.md`)

- **§2 the census** — use the AGGREGATE, not a listing. ⚠ It is an **upper bound**: a regex
  over `href="…"` counts href-shaped byte sequences, not links (that is 180).
- **§7 the pod-grep triple** — new marker, **negative control**, positive control, one exec,
  every replica. Post-roll result recorded in `bugs_closed/136`.
- **§6 the pattern-check controls** — all four, including (c) "a comment naming the seam must
  not silence it". The `strip_comments` landmine is live on that file.

## 6. Things this lane got wrong, so you do not repeat them

1. **Twice in one day I let a query's answer wear the noun I wanted.** "(30 rows)" from a
   `GROUP BY` listing became "30 links" (real: 35 occurrences, 6 sites); then "35" itself
   turned out to include a non-anchor. Both in `WRONG_CALLS.md`. Get numbers from
   `count(*)`/`count(DISTINCT …)` and **name the unit**.
2. **A council seat caught a vacuous negative from the PLAN alone** — a
   `mock.ExpectationsWereMet()` with nothing registered proves nothing, and the sibling assert
   could not catch it either because the fail-open path returns the input unchanged. Fix:
   REGISTER the call that must not happen, require it to go unmatched.
3. **A mutation that PASSES is not automatically a weak test.** Removing one guard changed
   nothing because an identical guard sits one level down — they are in SERIES. Check you
   actually changed the behaviour before believing your own mutation.
4. **`who-owns.py` is a lagging indicator.** It reads commits; a session mid-fix has none.
   Grep the live `.jsonl` transcripts for CODE SYMBOLS, not just the bug number.
5. **A probe beats a fetch.** Curling the live page to see whether the repair had corrupted it
   404'd and stalled; running the real function over the real bytes took two minutes and was
   deterministic.

## 7. Where everything lives

- Standing five: `docs/agent_docs/docs024_key_docs_latest/bugfix_136_sibling_link_repair/`
  (PLAN, RUNBOOK, NOTES, README_where_we_are, SUMMARY_2026-08-02, and this handoff)
- Closed case: `bugs_closed/136_HANDOFF_2026-07-28_section_editor_and_three_siblings_persist_links_with_no_repair.md`
- New bug: `bugs_open/180_HANDOFF_2026-08-02_link_repair_rewrites_javascript_that_builds_anchors.md`
- RFC: `docs/agent_docs/docs024_key_docs_latest/architecture_review/RFC_008_a_mandatory_write_seam_for_page_components_rendered_html.md`
- Register: **LNK-027** in `docs026_concept_register/register/link-management.md` (status `deployed`)
- Landmines: two added — the writer-set one (136) and the JS-anchor one (180)
- 016b §9: two patterns added — "count an action's WRITES, not its BRANCHES" and "a regex that
  reads MARKUP cannot tell markup from a string literal that CONTAINS markup"
