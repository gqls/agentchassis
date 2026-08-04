# HANDOFF — `bugs_open/185` lane, 2026-08-04

Supersedes `HANDOFF_2026-08-03_continue_here.md` (that one covers `bugs_closed/175`, which
is finished). **Read this, then `bugs_open/185` itself** — that file carries the full
per-tranche record, and every claim in it is backed by a query or a mutation.

## One paragraph

`bugs_closed/175` is closed, live and done. Its follow-up `bugs_open/185` — *"every
detector that selects `build_status = 'deployed'` is blind to 28 live pages"* — has **all
three tranches written, committed and LIVE on chassis `v1.0.1247`+** (currently `v1.0.1250`).
Tranche 1 is additionally **proven in behaviour**. What is outstanding is **two council
verdicts** (both resubmitted, answers already written into the bug file) and **two named
follow-ups that belong to other lanes**. No code is owed.

## Where each tranche stands

| tranche | what it did | code | council |
|---|---|---|---|
| **1** | alias-aware predicate builders + 3 detectors converged | LIVE, **PROVEN** (gaswholesalers orphan finding, new predicate 3 vs old 0) | **APPROVED** `66d07687` |
| **2** | render audit + both rerender queuers converged; 3 keeps recorded in the pattern-check allow-list | LIVE `v1.0.1247`+ | **APPROVED** `b563a61c` (round 2) |
| **3** | planner's empty-page gate → `realisedPageHasShipped` + migration 302 | LIVE (symbol present, old symbol absent); migration live | **APPROVED** `c881ef22` (round 2, 3rd attempt — two runs reaped) |

**ALL THREE TRANCHES ARE NOW COMMITTED, LIVE AND COUNCIL-APPROVED. No code is owed and
nothing is pending.** What follows is the residue, all of it owned elsewhere or optional.

**Superseded pick-up instruction:** checking those two correlations for a verdict.
`SELECT metadata->>'decision' FROM diagnosis_artifacts WHERE correlation_id='<corr>' AND kind='council_report' ORDER BY created_at;`
If either is APPROVED, nothing to do but note it. If REVISE again, the objections come with
the reviewers' own checks answered — read them against the bug file's answer tables first,
because most of what has come back so far was already true in the code.

## Follow-ups, none blocking

0. **The silent fallback in `realisedPageHasShipped`** (`bug_historian` [medium], tranche 3's
   approval). When `has_shipped` is absent the gate reverts to the old `build_status` test
   with **no log** — so a caller wired wrong, or migration 302 reverted, would quietly
   restore the buggy predicate. Deliberately not changed: live, approved, zero exposure, and
   the honest fix is a Debug line in a per-page hot path that wants its own small round.

## Outstanding, and neither is mine to fix

1. **`max_pages` on eight sites (config, audit-cadence owner).** Converging the render
   audit put 8 sites over the 25/site cap; **two crossed the line because of this change**
   (gamesdesign 25→34, leopardess 25→34), six were already over. The cap is NOT silent — it
   logs `TRUNCATED` and returns `pages_total` — but coverage genuinely shifted. Raising
   `max_pages` is config, not code.
2. **`tool-archetype-taster-quiz` subject-key mismatch (tool-acceptance lane).** Its plan is
   filed under `tool-archetype-taster-quiz`; the ladder computes `archetype-taster-quiz`
   for a `component_level='section'` component, so it never matches. **Censused: exactly 1
   of 24 current tool plans**, pre-dates this work, found by the tranche-1 verification.
3. **Fix candidate 2** (open by choice): merging `NeverDeployedPagePredicate` with
   `queryresolve.FetchablePageEligibilitySQL`. They cross-reference by comment now. Read
   `queryresolve.go:210-236` first — it documents a deliberate family of three.

## The seam is being adopted, which is the best signal in this lane

The **098 retraction lane** picked up `PageHasShippedPredicateFor` and added the
complementary axis themselves — `PageWantedLivePredicateFor` (`status = 'active'`), commit
`6a7ab87a8` — then applied it to the very render-audit query this lane converged. Two axes,
both named, in one place. That is the shape the whole bug was arguing for, arriving from
another session without being asked.

## What will bite you here

**Verification probes, in descending order of strength.** This lane got this wrong twice in
one session and both are in `WRONG_CALLS.md`:

1. **A function symbol** — survives compilation, greps cleanly, and a RENAME gives a free
   negative control (`realisedPageIsBuilt` = 0 *beside* `realisedPageHasShipped` = 1 is far
   stronger than either alone).
2. A long, distinctive **single** string literal.
3. Commit ancestry (`git merge-base --is-ancestor`) — sound, weaker, label it.

**Never** grep for a Go **comment** (it cannot be in the binary, so `0` is meaningless and
reads exactly like "not shipped"). **Never** chain two greps over `strings` output — Go's
string table concatenates unrelated constants onto one line, so `grep A | grep B` matches
noise. And **never predict a string COUNT**: a predicate built at runtime by a function is
not a literal at all.

**Three council rounds in a row caught a documentation gap, not a defect.** Each time the
code was already correct and the submission simply did not show it — an incomplete edit
list, an unquoted `spec.reason`, an unverified predicate match. **If a submission's safety
rests on something being true in the code, quote the line.** The reviewers cannot see the
diff, only what you describe.

**Other traps this lane hit:** the shared `/tmp` filled to 100% mid-commit (the commit
landed, only its output was lost — check `git log` before re-running anything); a council
run was **reaped** after a roll killed it mid-seat (`stale EXECUTING_STEP for >4h`), which
looks identical to latency; and bug numbers collide routinely — this file's own bug was
renumbered 181 → 185 because another lane filed a different 181 first. **Resolve by slug.**

## Docs map

- `bugs_open/185_…` — the record: census, both-ways diffs, per-tranche decisions, all
  council answer tables.
- `bugs_closed/175_…` — the parent, closed and live.
- `NOTES_page_role_upsert.md` — missteps in order, with the checks that would have caught them.
- `RUNBOOK_page_role_upsert.md` — commands with their gotchas, incl. post-roll pod-grep.
- `TALK_2026-08-03_…` / `TALK_2026-08-03b_…` — the review-machinery account and the
  judgement→mechanical ratchet (the second one's addendum records a check that measured
  itself out of existence).
- `architecture_review/RFC_010` — RATIFIED, all four owner rulings implemented and live.
- Register: **PBP-027**. Landmines: two, synced. `WRONG_CALLS.md`: six entries from this lane.
