# HANDOFF 2026-08-26 — bugfix 410 (silent scan loss): fix LIVE, ratchet LIVE, what remains and how to verify anything I claim

Written for a session with none of this context. Every claim carries its check. **Cite symbols,
never line numbers** — this lane's own corrections log a line number expiring the same afternoon
it was written.

## ⚠ FIRST: the number 410 is AMBIGUOUS

Two unrelated bugs share it (CLAUDE.md ambiguous-number list, 2026-08-26). THIS lane is
`bugs_open/410_HANDOFF_2026-08-26_three_seams_fail_toward_the_quiet_default_and_the_artefact_looks_freshly_built.md`.
The OTHER is the content-feed phase-lock case (`…next_fetch_at_stamped_at_fetch_time…`), owned by
its own session — do not touch it, do not read its commits as ours. Resolve by slug; `git log`
the FILE PATH.

## What this lane is

`bugs_open/410` is a pattern file: three seams, three lanes, one week, all failing toward the
quiet default (assemble/skip), all reporting `complete`, all leaving the artefact looking freshly
built. The filing's core insight: **the estate's safest mode is also its silent-failure mode**,
so every drift lands there and announces nothing by construction.

This lane took **instance 3** — `loadStoredSections`
(`platform/orchestration/actions/rerender_page_sections_action.go`): its `rows.Scan` error branch
was Warn-and-continue, so a scan failure returned fewer sections, or none, **with no error**. And
because `save_page_sections` replaces the page's rows **wholesale** (`DELETE FROM page_components
WHERE page_id = $1`, `save_page_sections_action.go`, symbol area ~:898), a dropped section was not
merely unrendered — **its row was deleted**. Destruction, not degradation: that asymmetry is the
whole design argument.

## Current state — everything below is DONE and verified

| what | state | evidence / re-check |
|---|---|---|
| the fix | **LIVE** | `datahelpers.ScanShortfall(offered, kept, subject)` applied strictly in `loadStoredSections`; commit `7c443aac6`; probe-verified post-roll in BOTH `agent-chassis` pods (three-way `/proc/1/exe` form, RUNBOOK §12): capability literal `refusing the partial result` PRESENT on both nodes, must-present control passed, must-absent control absent |
| council | **APPROVED round 1** | correlation `c8385154-17b4-43f5-94b2-41f552f43867`; verdict artifact 2026-08-26 11:20:59Z; "4 advisory objections — none high". Both commits credited (`7c443aac6` via `Council-Submitted`, `b93622995` via `Council-Reviewed`) |
| the advisories | **all actioned, adjudicated, or refuted** — commit `b93622995` | dispositions table in NOTES §evening; the debug_historian medium was REFUTED by measurement (all 8 projected columns now have controls: `2295 | 2295 | 1064 | 10 | 0 | 0` — the NULL-heavy columns are exactly the COALESCE'd ones) |
| the class ratchet | **LIVE in every build** | `scan_swallow_ratchet_test.go` (blocking, `platform/orchestration/actions/**`, 166 sites) + `check_scan_swallow` in `scripts/pattern-check.py` (advisory, tree-wide, the other 41) — both read `scan_swallow_baseline.txt` (207 sites / 127 files, census 2026-08-26). Counts only ever go DOWN; opt-out marker `// scan-loss:accepted: <reason>` at the site |
| guard silence | 0 refusals post-roll, **recorded as UNDEMANDED** | 0 rerender invocations in the window, so the zero distinguishes nothing yet; the demand control is the mutation-proved test suite (fires the guard every build) |
| docs | current | NOTES / RUNBOOK / README / SUMMARY in this directory; bug file scoreboard appended; register **DBI-027** (database-and-infrastructure.md + index row) says LIVE |

**Every guard is mutation-proved by execution, not assertion**: neuter `ScanShortfall` → its test
red; delete the loader's call → both refusal tests red, reproducing the defect verbatim
("returned 1 section(s) with no error"); remove a baseline line → "NEW silent scan loss"; inflate
a count → "FELL … ratchet down".

## What REMAINS — none of it is this fix incomplete; all of it is scoped-out and stated

1. **Candidate 1 (the expensive one): make "I did not understand this" a refusal fleet-wide.**
   The door-closing fix on the busiest pipeline. Needs its own design round and council review —
   the bug file says so, and no one has claimed it. If picked up: the strict-vs-graded asymmetry
   argued in `ScanShortfall`'s doc comment is the template (refuse where the consumer replaces
   wholesale; degrade loudly where it projects).
2. **Candidate 2 / `bugs_open/404` candidate 0: vocabulary↔reader parity at commit time.**
   UNCLAIMED (confirmed by the 384 lane 2026-08-26). Different mechanism from ours: a parity
   check would not notice a swallowed scan, and our scan ratchet cannot notice a reason the DB
   knows and Go does not. Whoever builds it: our ratchet reads Go source ONLY — that boundary is
   stated in its header so the two checks never look like each other's coverage.
3. ~~**The content-loss residual, in the very loop we fixed:** `_ = json.Unmarshal(cdJSON,
   &s.contentData)` keeps the row and EMPTIES it on corrupt JSON — `offered == kept`, invisible
   to any count guard. Needs a decision: may an unparseable section render as an empty one?
   Commented at the site, listed in the ratchet header's blind spots.~~
   **CLOSED 2026-08-31** — decision made by measurement (jsonb column, 0 of 2,751 live values
   non-object, 55 SQL-NULL rows stay loadable): decode failure now drops the row so the existing
   `ScanShortfall` refuses. Council corr `a69d82f2` (submitted; read the verdict). NOTES
   2026-08-31 has the full account. Item 2 also moved: shipped by the 404 lane (`ef4236b4d`).
4. **Convergence debt (reuse_agent advisory, tracked not owed now):** `scanBlogArticles`
   (`rebuild_blog_listing_action.go`) is the TRUE sibling — converge onto `ScanShortfall` the
   NEXT time that function is touched (needs a graded helper variant, shipped WITH that caller;
   note sits at its counters). `collectPageSections` is a **FALSE sibling** (in-memory array,
   degrade-not-refuse) — do NOT converge it; the reason is in `ScanShortfall`'s doc so pressure
   cannot force a false unification.
5. **The 207 census will go stale BY ADDITION.** Re-run before quoting:
   `git log --since=2026-08-26 --diff-filter=A -- platform/ internal/ pkg/ cmd/` — non-empty
   means re-census. The parity re-check recipe (Go vs Python classifier) is in the baseline
   file's header; run it whenever either classifier is edited.
6. **A standing constraint inherited by the `news_editorial` lane (they have accepted it):** any
   column added to `loadStoredSections`' SELECT must be NOT NULL or COALESCE'd, or the guard
   starts firing on their edit. Also the poison-row trap, stated in
   `rerender_page_sections_scan_completeness_test.go`'s header: the tests poison via NULL into
   `position` (an `int`); if `position` ever becomes nullable or COALESCE'd, **those tests go
   green for the wrong reason** — pick another non-nullable destination, never delete the test.

## Two pre-existing advisories in files we touched — NOT ours, passed to their owners

- `runtime-fill-scope` at `rerender_page_sections_action.go` (raw `data-runtime-fill` string
  test, bugs_open/137 family) — in the `news_editorial` lane's region; they have taken it and
  found their P1 extraction forces a re-graining decision there.
- `logged-model-output` in `rebuild_blog_listing_action.go` (~:707 area) — pre-existing,
  unowned, low.

## How to verify ANY of this from scratch (the fast path)

```bash
# the guard exists and is called (two greps):
grep -n "ScanShortfall" platform/orchestration/datahelpers/scan_completeness.go \
     platform/orchestration/actions/rerender_page_sections_action.go
# the whole suite, including the ratchet:
go test ./platform/orchestration/actions/... ./platform/orchestration/datahelpers/... -count=1
# it is in the RUNNING binary (full three-way form + gotchas: RUNBOOK §12):
POD=$(kubectl -n ai-persona-system get pod -l app=agent-chassis -o name | head -1)
kubectl -n ai-persona-system exec $POD -- grep -aq "refusing the partial result" /proc/1/exe && echo LIVE
# the verdict:
SELECT metadata->>'decision' FROM diagnosis_artifacts
 WHERE correlation_id='c8385154-17b4-43f5-94b2-41f552f43867' AND kind='council_report';
```

## The lane's transferable lessons (full versions in WRONG_CALLS `2026-08-26c` §1–5)

Five measurement errors in one day, all one error: **the measurement answered the question that
was ENCODED, not the one asked.** The distilled rules now in estate memory: *the number goes next
to its population, written down before the number, or it is not evidence yet*; *reproduction is
not verification when both parties share the same wrong assumption* (an exact match is the most
persuasive possible form of no evidence); and the only check that ever caught anything: **open one
member of the population and read it** — aggregates were looked at three times, a member zero
times, and the member settled it in ten seconds.

## Key artefacts

| what | where |
|---|---|
| the bug file (scoreboard at the tail) | `bugs_open/410_HANDOFF_2026-08-26_three_seams_fail_toward_the_quiet_default_and_the_artefact_looks_freshly_built.md` |
| lane docs (NOTES / RUNBOOK / README / SUMMARY / submission) | `docs/agent_docs/docs024_key_docs_latest/bugfix_410_silent_scan_loss/` |
| the helper + its tests | `platform/orchestration/datahelpers/scan_completeness{,_test}.go` |
| the fixed loader + its tests | `platform/orchestration/actions/rerender_page_sections_action.go` (`loadStoredSections`), `rerender_page_sections_scan_completeness_test.go` |
| the ratchet + baseline + advisory twin | `platform/orchestration/actions/scan_swallow_ratchet_test.go`, `scan_swallow_baseline.txt`, `scripts/pattern-check.py` (`check_scan_swallow`) |
| register entry | `docs/agent_docs/docs026_concept_register/register/database-and-infrastructure.md` **DBI-027** (+ index row) |
| commits | `7c443aac6` (fix + ratchet), `2840a8b79` (lane docs), `99d4b574e` (index row), `523b28cc0` (summary), `b93622995` (advisory close-out) |
| council | `c8385154-17b4-43f5-94b2-41f552f43867`, APPROVED r1 |
