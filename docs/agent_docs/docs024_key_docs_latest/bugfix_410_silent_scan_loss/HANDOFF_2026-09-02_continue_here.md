# HANDOFF 2026-09-02 — bugfix 410 (silent scan loss): LANE COMPLETE on both axes; what remains is not this lane's work

Written for a session with none of this context. Every claim carries its check. Cite symbols,
never line numbers. Supersedes `HANDOFF_2026-08-26_continue_here.md` (kept — its per-item state
was corrected in place as things closed; this file is the current picture).

## ⚠ FIRST: the number 410 is AMBIGUOUS

Two unrelated bugs share it (CLAUDE.md ambiguous-number list). THIS lane is
`bugs_open/410_HANDOFF_2026-08-26_three_seams_fail_toward_the_quiet_default_and_the_artefact_looks_freshly_built.md`.
The OTHER is the content-feed phase-lock case (`…next_fetch_at_stamped_at_fetch_time…`), owned
by its own session. Resolve by slug; `git log` the FILE PATH, never the number.

## What this lane was, in one paragraph

`loadStoredSections` (`platform/orchestration/actions/rerender_page_sections_action.go`) feeds
`save_page_sections`, which replaces a page's rows WHOLESALE (`DELETE FROM page_components WHERE
page_id = $1`). So anything the loader silently loses is DESTROYED on save, under a fresh deploy
stamp, reported complete. Two silent-loss axes lived in that one loop: a row that failed to
scan was skipped (fixed 2026-08-26), and a row whose `content_data` failed to decode was kept
but EMPTIED (fixed 2026-08-31). Both now refuse via one guard, `datahelpers.ScanShortfall`.

## Current state — everything below is DONE, live, and re-verifiable

| what | state | re-check |
|---|---|---|
| row-axis fix (parent) | FIXED AND LIVE since 2026-08-26 roll | commit `7c443aac6`; council APPROVED r1 `c8385154-17b4-43f5-94b2-41f552f43867`; probe literal `refusing the partial result` (RUNBOOK §12) |
| content-axis fix (residual) | **FIXED AND LIVE since 2026-09-02 roll** | commit `359503af0`; council APPROVED r1 `a69d82f2-9859-4c33-98d9-e791fade2974` (all reviewers); probe literal `content_data does not parse into a section object` (RUNBOOK §13 — the parent literal does NOT prove this one; any build of `7c443aac6..359503af0` carries parent-only) |
| live verification 2026-09-02 | both pods, ReplicaSet `744cfb4bf`, different nodes | three-way form: both capability literals PRESENT, must-absent control clean, on BOTH pods |
| guard demand | **DEMANDED zero on both axes** | since 2026-09-01: 1,376 rerender-plan orchestrations, 176 executing `rerender_sections`, 0 refusals, 0 decode warns. ⚠ `orchestration_states.execution_path` is `[]` on this pipeline even on failed runs — count executions by `collected_data ? 'rerender_sections'` (the step's output_field), never by execution_path |
| the class ratchet | LIVE in every build | `scan_swallow_ratchet_test.go` (blocking, `platform/orchestration/actions/**`) + `check_scan_swallow` in `scripts/pattern-check.py` (advisory, tree-wide), one shared baseline `scan_swallow_baseline.txt` — 207 sites / 127 files, censused 2026-08-26, **re-censused 2026-08-31: identical**, 0 classifier disagreements. Re-census before quoting: `git log --since=2026-08-31 --diff-filter=A -- platform/ internal/ pkg/ cmd/`; parity recipe in the baseline header |
| the boundary tests | green, mutation-proved TWICE over | `rerender_page_sections_scan_completeness_test.go`: non-object content_data refuses ("kept 1 of 2"); SQL-NULL and jsonb-`null` rows STAY loadable (55 live such rows — the guard is exactly as strict as the parse, and that test is what stops a future over-tightening). Mutations: restore `_ = json.Unmarshal` → NonObjectContentData test red alone; delete `ScanShortfall` → all three refusal tests red |
| docs / register | current | NOTES / RUNBOOK (§13 added) / README / SUMMARY series (latest: `SUMMARY_2026-09-02_lane_complete.md`) in this directory; DBI-027 corrected visibly twice (residual closed → live); bug file scoreboard appended through 2026-09-02 |
| commits this phase | `359503af0` (fix), `01fcbbac6` (verdict close-out), `8b0d34c21` (RUNBOOK §13), plus this handoff's commit | all pathspec commits, scope reports clean |

Decision trail for the residual, since four council seats asked: the deferral was IN the
approved material (`7c443aac6` shipped the "KNOWN RESIDUAL … Not fixed here" comment and the
ratchet's "CONTENT loss … NOT covered" bullet), and the closure decision was made by
measurement: `content_data` is jsonb (broken syntax unstorable), 0 of 2,751 live values
non-object `[2026-08-31]`, re-confirmed 2,759/0 post-verdict — so the refusal is unreachable on
today's data and exists for the first writer that changes that.

## What REMAINS before the BUG FILE can close — none of it is lane work

1. **Candidate 1: unknown/not-understood → refusal, fleet-wide.** The door-closing design round.
   UNCLAIMED. ⚠ It now OVERLAPS the 404 lane's open question ("what happens to a reason nobody
   declared?" — tail of `bugs_open/404_HANDOFF_2026-08-25_…`, point 2). Whoever picks either up
   must read BOTH files first, or the estate gets two competing refusal designs on one seam.
   Template if picked up: strict-vs-graded asymmetry in `ScanShortfall`'s doc comment (refuse
   where the consumer replaces wholesale; degrade loudly where it projects).
2. **Instances 1 and 2 close in their own lanes** (`bugs_open/384`, `bugs_open/404` — both still
   open 2026-09-02). This pattern file cites them, does not own them.
3. **An OWNER DECISION on the closure shape**: spin candidate 1 into its own file (it is a
   design proposal now, not a diagnosis) and close the pattern file — the lane's recommendation,
   argued at the bug file's 2026-09-02 tail — or keep the pattern file open as candidate 1's
   tracker. Nobody but the owner should make this call; the file move itself is cheap
   (`git mv` + BOTH paths on the commit — see the LANDMINES entry on half-shipped moves).

Explicitly NOT closure blockers (stated so they are not mistaken for open work): the 41
advisory-only sites outside the blocking package; `scanBlogArticles` convergence-on-touch debt
(graded helper variant ships WITH that first caller — `collectPageSections` is a FALSE sibling,
do not converge it); the ratchet's class blind spot for decode-swallows at OTHER sites; the
news_editorial lane's inherited SELECT-column constraint and the poison-row trap (both stated in
the test file header).

## How to verify ANY of this from scratch

```bash
# both guards exist and are called:
grep -n "ScanShortfall\|json.Unmarshal(cdJSON" platform/orchestration/datahelpers/scan_completeness.go \
     platform/orchestration/actions/rerender_page_sections_action.go
# the suite + ratchet:
go test ./platform/orchestration/actions/... ./platform/orchestration/datahelpers/... -count=1
# both axes in the RUNNING binary (full three-way form + gotchas: RUNBOOK §12 and §13):
POD=$(kubectl -n ai-persona-system get pod -l app=agent-chassis -o name | head -1)
kubectl -n ai-persona-system exec $POD -- grep -aq "content_data does not parse into a section object" /proc/1/exe && echo CONTENT-LIVE
kubectl -n ai-persona-system exec $POD -- grep -aq "refusing the partial result" /proc/1/exe && echo ROW-LIVE
# the verdicts:
#   SELECT metadata->>'decision' FROM diagnosis_artifacts WHERE correlation_id='<corr>' AND kind='council_report';
#   c8385154-17b4-43f5-94b2-41f552f43867 (row axis) · a69d82f2-9859-4c33-98d9-e791fade2974 (content axis)
```

## Key artefacts

| what | where |
|---|---|
| the bug file (2026-09-02 closure section at the tail) | `bugs_open/410_HANDOFF_2026-08-26_three_seams_fail_toward_the_quiet_default_and_the_artefact_looks_freshly_built.md` |
| lane docs | `docs/agent_docs/docs024_key_docs_latest/bugfix_410_silent_scan_loss/` (NOTES = technical log incl. every misstep; RUNBOOK §12/§13 = the probes; SUMMARY series = the milestones) |
| helper + tests | `platform/orchestration/datahelpers/scan_completeness{,_test}.go`, `platform/orchestration/actions/rerender_page_sections_scan_completeness_test.go` |
| ratchet + baseline + advisory twin | `platform/orchestration/actions/scan_swallow_ratchet_test.go`, `scan_swallow_baseline.txt`, `scripts/pattern-check.py#check_scan_swallow` |
| register | DBI-027 in `docs/agent_docs/docs026_concept_register/register/database-and-infrastructure.md` |
| commits | `7c443aac6` · `b93622995` · `359503af0` · `01fcbbac6` · `8b0d34c21` (+ lane-doc commits `2840a8b79` `99d4b574e` `523b28cc0` `14fb629e0`) |
| council | `c8385154` APPROVED r1 (row) · `a69d82f2` APPROVED r1 all reviewers (content) |
