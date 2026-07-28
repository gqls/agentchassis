# Bugs sweep — CONTINUE HERE (2026-07-28b, end of the second day)

**Supersedes `HANDOFF_2026-07-28_continue_here.md`** — that file's NEXT list is now
wrong on every item. Read this, then `NOTES_bugs_sweep.md` for the misstep log (two
threads write it; entries are signed).

## THE CONSTRAINT THAT SHAPES EVERYTHING UNTIL 08-01

**The Anthropic usage cap is exhausted; access returns 2026-08-01 00:00 UTC**
(`bugs_open/130` + the gauntlet thread own the write-up; owner decision needed).
Until then: **no council reviews, no diagnosis runs, no LLM content generation.**
Orchestrations DIE at their first LLM step in ~3s showing `complete_invalid` /
`COMPLETED` with `error` NULL — read `__step_error`, and do not misread the death as
queue latency (this session nearly did). What still works: everything non-LLM — feed
ingestion, page deploys, Go builds/tests, migrations, pod operations. Platform-code
commits can proceed trailer-less with the outage noted in the commit body (precedent:
`f78cf8125`); do NOT resubmit council runs before 08-01, they die at seat one.

## STATE — what changed on 2026-07-28 (two sweep threads working in parallel)

| bug | state | proof |
|---|---|---|
| **127** (news search was web search) | **CLOSED → `bugs_closed/`** | live both sides (adapter v1.0.1185, chassis v1.0.1187); 13:50 refresh delivered dated news, `source_published_at` populated after NULL-forever; council APPROVED r1 `a7ae8ce8` |
| **109** (render-context allowlists) | **all four maps derived, LIVE v1.0.1187** | pod-grep `setRenderContextScalarsFromData` 2 / `renderContextControlFields` 1; council corr `1d082754` died on the CAP not the plan — trailer owed, resubmit is optional post-08-01. Close/hold decision noted in-file (per-page trio residue) |
| **108** (code index stale/no bodies) | **CLOSED** (arch-review thread) | index mirrors pushed `d98010e8b`; banner names ref+sha and goes loud-STALE on drift |
| **097** (CTA integrity misses card links) | repair half FIXED+LIVE v1.0.1187 (sibling): 23 dead links cleared, 6 on the live page | stays OPEN for the detector half: `ctaFieldNames` still blind to card arrays; `check_phantom_internal_links` zero-findings on robot-hands unexplained (likely 083) |
| **129** (child swallows spawn) | FILED + diagnosis CONFIRMED (sibling) | bursty, 2/3 on the index lane |

**Key mechanism finding (097's diagnosis, CONFIRMED):** the BULK rerender path
(`rerenderSinglePage`) deploys **without** the `validate_page_content` gate, so
`RepairPageLinks` never sees bulk-rebuilt pages. Also: `RerenderSitePagesAction` is
unregistered dead code — the live path is `RerenderSinglePageAction` via per-page items.

## WHAT PAID OFF TODAY (add to the method, not instead of it)

- **Commit-per-task made a sibling's roll deliver your fix.** 109 + 127's chassis side
  went live on v1.0.1187 without this thread deploying anything — whatever HEAD a
  sibling builds carries whole tasks. The INVERSE also happened: a sibling's registry
  push overwrote a sibling's image under its own tag — **check the pod's imageID
  DIGEST, not the tag, when two threads build the same service** (their NOTES entry).
- **The bug's own acceptance check, read BEFORE building, changed the fix.** 127's
  "assert source_published_at populated" is what surfaced that the primary provider
  returns relative dates the feed writer can't parse — without reading the verifier
  first, the fix would have passed every test and failed its acceptance.
- **Verify external APIs against their docs, not memory** (WebFetch both providers'
  docs; one assumption from memory was wrong). Mark what stays documented-not-witnessed.
- **A foreign in-progress MERGE blocks all partial commits** ("cannot do a partial
  commit during a merge"). Don't touch their merge; queue behind it with a watcher on
  `.git/MERGE_HEAD`. Expect your working-tree doc edits to be SWEPT into their merge
  conclusion — content survives, attribution doesn't; record attribution in your next
  commit. The branch can also change under you (086 → `087_towards_multiple_domains`).

## NEXT TARGETS (re-verify ownership per §3 of the old handoff — both checks)

1. **091 candidate 1** — coordinate with `work_item_completion_integrity` first; the
   dropped-finding fix is theirs by prior claim.
2. **097 detector half** — unowned now the sibling finished the repair half; but its
   "why zero findings on robot-hands" question may need a diagnosis run → post-08-01.
3. **Unowned untouched:** 100, 104, 111, 114, 115, 117, 118, 128. Mind: 100+101 are
   vetcomparison's blocker (coordinate); 117+118 are one chrome pair (memory:
   [[bugfix-117-chrome-stored-artefact]]).
4. **Post-08-01 housekeeping:** optionally resubmit 109's council run; the 098 report
   will list several of today's commits un-reviewed — all cap casualties, documented
   in their commit bodies.

*Numbering note: bugs_open/127 resolved by slug is the news-search case (now in
bugs_closed/); the relojistas docs mention an unrelated "Action-127".*
