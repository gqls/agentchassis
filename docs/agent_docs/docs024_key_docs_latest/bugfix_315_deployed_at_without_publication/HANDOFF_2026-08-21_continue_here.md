# HANDOFF — `bugs_open/315`, continue here (2026-08-21 ~14:00Z)

**Supersedes `HANDOFF_2026-08-20_continue_here.md`.** That file is still accurate about everything
it describes; it is out of date only in that its one substantive open item — the D5 divergence
sweep — is now **BUILT**. Read this first, then that one for the background it does not repeat.

**The lane's work is now COMPLETE except for switching one thing on, which cannot be done from
here.** Nothing is broken, nothing is urgent, and the fleet is measurably healthy.

---

## 1. What changed today

The divergence sweep (`PLAN` decision **D5**, `bugs_open/315` fix candidate 4) is written, tested,
committed and registered. It is the first consumer of `pages.content_hash`: it fetches every
judgeable page with a cache-buster, sha256s the body, and compares against the fingerprint the
deploy stamp recorded. That turns *"is the origin serving what we sent?"* from six hours of nobody
noticing into a comparison a machine makes every few hours.

| piece | state |
|---|---|
| `check_page_content_divergence.go` + tests | **COMMITTED** `f715b8c1d` — 15 tests, 9 guards mutation-proved |
| Register `DGH-015` (+ index row) | **COMMITTED** `e05c38cdb`; also corrects `DGH-013`'s now-false "nothing consumes content_hash yet" |
| Migration `526_..._HOLD.sql` (+ `_ROLLBACK`) | **WRITTEN AND PROVEN, NOT APPLIED** — held until the image rolls |
| Council round | `SUBMISSION_CORR = be85a6d3-f2c0-4f7a-b791-e95087141fc8` — **read the verdict** |
| Docs (the standing five) + `WRONG_CALLS` + `LANDMINES` | **COMMITTED** `ba4abbd5d`, `ae210bef4` |

## 2. THE ONLY THING LEFT — and it is two steps, in this order

**a. A chassis image carrying `f715b8c1d` must roll.** Not this lane's call: releases are
whole-fleet and the owner runs `make release`. Confirm at the artefact, per service, never at git:

```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor f715b8c1d <the stamped sha> && echo SAFE-TO-ENABLE
```
An empty grep means "the startup line has scrolled", **not** "unstamped".

**b. THEN apply `docs/agent_docs/sql_for_agents/526_enable_page_content_divergence_HOLD.sql` by
hand — and expect it to REFUSE until §4 is dealt with.** The order is not a preference: `run_discovery_checks` resolves each name against the
binary's own registry, and a name the binary does not register **fails the whole `run_checks` step**
— taking `site_unreachable`, a live and useful check, down with it. The file carries the full apply
procedure, the damage query to run FIRST, and its own rollback.

## 3. What the measurements say, so you do not have to re-derive them

- **The fleet is healthy.** `[MEASURED 2026-08-21 ~10:35Z]` **228 of 228** active pages carrying a
  `content_hash` serve bytes hashing exactly to it, across 12 domains. Expect the check to find
  **nothing** on day one; a finding is likelier to be a defect in the check than a divergence in the
  fleet, so re-run the comparison by hand before believing it (RUNBOOK Part 3).
- **That sweep is also the end-to-end proof of the fingerprint itself** — 228 independent
  confirmations that nothing between the stamp and the wire alters the bytes.
- **The 30-minute settle window is LOAD-BEARING, not precautionary.** `[MEASURED 10:38Z–13:20Z]`
  1,099 re-probes, 85 pages, 95 deploy events: the only 3 DIVERGED readings were at ages **1s, 13s,
  14s**, all converged by 140–156s; **0 of 995** readings at age ≥157s diverged. Those 3 are 3 work
  items the check would have filed against healthy pages in under three hours without the window.
- **Two intermittent 404s** in the same watch, both serving one shared edge error page, each
  surrounded by MATCH — 2 more false items, prevented by treating a non-200 as unjudgeable.
- **It will be exercised fast** — once it is enabled at all (see §4).
  `site-discovery-rotation-availability` fires every **300s** with a **4-hour** floor (NOT the quality
  rotation's 7 days), so the fleet is swept every ~4–5 hours.

## 4. ⚠ THE ONE THING THAT BLOCKS ENABLING — and it is a LIVE hazard, not a theoretical one

**The check is sound only if every live step that stamps a page `deployed` also records what it
sent.** An unarmed stamper leaves a fingerprint describing an OLDER deploy, and the check then
convicts a healthy page — permanently, because nothing rewrites that row.

`[MEASURED 2026-08-21, RECURSIVE walk]` **there are SIX such steps and THREE ARE UNARMED:**
`page-rebuild`, `pageflow-builder` and `site-work-orchestrator`, all at
`workflow.steps.<loop>.config.sub_workflow.steps.update_page_status` — **the page-BUILDING paths**,
i.e. the ones that actually emit new bytes.

> **This corrects an earlier claim of mine that said the opposite** ("exactly three, all armed, zero
> unarmed") and which reached the check's header, the register, three commit messages and the council
> submission. My census walked one level. **The council gate's `guardian` seat caught it from the
> SHAPE of the claim without seeing the query.** Full account in `WRONG_CALLS.md` and NOTES.

**You do not have to remember this.** Migration **526 refuses to apply** while the recursive unarmed
count is non-zero — proven to bite against live data ("3 of 6 live steps … do NOT declare
deploy_result_field"). So §2b will simply fail until this is dealt with, which is the intended
behaviour, not a fault.

**Nothing is poisoned today** — the 228-page sweep would have shown it and found 228 MATCH — but that
is an observation about one moment, not a property of the system.

**The fix is `PLAN` D7: arm the three.** All three carry a `git_commit` step `deploy_page` with
`output_field: "page_deployed"`, so it is one migration in **494's exact shape**. Arming beats D6's
stamp-side NULLing for these three because it RAISES fingerprint coverage where NULLing lowers it;
**D6 stays as the backstop for the NEXT unarmed stamper**, the one nobody notices being added.

⚠ **D7 needs its own council round and its own damage query.** Arming changes behaviour on the main
build path — an armed stamper REFUSES the stamp when its commit reported a skip — and the last time
this lane armed stampers it took the fleet's page-publishing down for 33 minutes (`bugs_open/336`).
**Re-run the recursive enumeration first** (RUNBOOK Part 3): the count can change without anyone
touching this code.

## 4b. The council's other two objections — DISCHARGED, and one of them changed the code

Both from the same APPROVED round (`be85a6d3`), both medium, both checked rather than waved through.

**`reuse_agent` — "is `sites.published_hash` / migration 422 the same mechanism?"** No: 422 drives
`publish_site`, a DIRECT B2 upload from a spawned credentialed pod, while this check observes
commit-is-deploy (git → Actions → B2 sync). `sites.published_hash` is site-level and is not page
bytes — the one live value is `th1:05a06351`, a prefixed TREE digest, against a per-page sha256 here.
And 422 fires only for sites with `publish_target` set: `[MEASURED]` **1 of 45**.

⚠ **But the seat's instinct found a real hazard and the predicate changed.** `publish_site_action.go`
writes **neither** `content_hash` nor `deployed_at`. So a site with hashed pages that later opts into
`publish_target` would keep fingerprints the new seam never updates — the stale-fingerprint hazard
again, by a different door. The query now carries **`s.publish_target IS NULL`**. `[MEASURED]` the one
opted-in site has 12 active pages and 0 hashed, so there was no exposure; the guard makes it
structurally impossible rather than merely currently absent.

**`debug_historian` — "you hand-rolled a liveness predicate instead of reusing the shared one."**
Half right, and acted on: the "did this page ship" leg is now
**`queryresolve.DeployedPageEligibilitySQL`**, concatenated rather than re-typed (a test pins the
reuse itself, not just the resulting text, because re-typing it inline would stay green while forking
the platform's definition). But `status='active'` is NOT a liveness filter and appears in no shared
predicate — it excludes RETRACTED and ARCHIVED pages, which keep `deployed_at` by design and are
deliberately unserved; judging them would report every retraction as a divergence. The enumeration
that seat asked for, which was genuinely owed:

| status | build_status | n | hashed |
|---|---|---|---|
| active | deployed | 651 | **232** |
| active | needs_rebuild | 56 | 0 |
| active | planned | 42 | 0 |
| archived | (all three) | 69 | 0 |

Every hashed page is `active`+`deployed`; no archived page carries a hash; `status='deployed'` never
occurs at all.

**`tooling_provenance` — "no concept-register edit in the submission."** Right about the submission,
wrong about the change-set: `DGH-015` exists (`e05c38cdb`), filed one commit after the submission.

## 5. Still open from the previous handoff, unchanged

- **`RFC_038` can be closed or ratified.** Its survey is done and the change is shipped and proven.
  Someone other than this lane should make that call.
- **`bugs_open/336`'s durable guard** — a test that every key an action READS is declared in ITS OWN
  spec. Not this lane's to build; it should not grade its own homework there.
- **`bugs_open/315` itself can be closed** once someone independently re-runs the proof. Candidates
  1 and 2 are delivered and live; **4 is now built** and awaits only the roll; **3 remains
  undiagnosable from here** (the runner workflow lives in the private `gqls/sites` repo) — which is
  precisely why candidate 4 detects it from this side instead.
- **`collectUniqueValue` extraction** stays this lane's accepted follow-on **if** the
  `staged_component_build` lane takes the resolver route for `commit_sha`; their CONTRIB is answered
  in full at the bottom of `CONTRIB_2026-08-20_…md`. Delete `collectUniqueValue` when RFC_029 Phase
  2 ships.

## 6. Traps this session hit for real (the 08-20 handoff's nine still stand)

1. **A mutation table written from the DESIGN is fiction.** Four of nine rows were false — three
   because a later guard in SERIES absorbed the fault, one because the cap test sized both its
   fixture and its assertion from the very const under test, so it could never fail. Write the table
   from the RUN.
2. **Two guards could not be made load-bearing at all** (non-200, oversize): each has a second guard
   in series. The file says so rather than claiming a proof it does not have.
3. **`psql … | tee f | head -20` silently truncated 228 rows to 21.** No error. Redirect, then read.
4. **`config->>'build_status'` returns a clean, confident ZERO** — the live key is `status`. Now a
   `LANDMINES.md` entry, because zero is an ordinary answer to that question.
5. **`pkill -f lag_watcher.sh` kills its own shell** — the pattern matches the `bash -c` running it.
6. **Do not name a migration number before you have looked** — the register pointed at `495` for
   twenty minutes; the next free number was `526`.

## 7. Where the documents are

```
docs/agent_docs/docs024_key_docs_latest/bugfix_315_deployed_at_without_publication/
  PLAN_2026-08-19_…md       D1–D5, plus D5's build record and D6 (the open hole)
  NOTES_…md                 technical log; read from the BOTTOM. The missteps are the point
  RUNBOOK_…md               THREE parts now — Part 3 is the sweep: the D6 query, the by-hand
                            divergence measurement, the lag measurement, the post-apply order
  README_where_we_are.md    the owner's plain-prose log
  SUMMARY_2026-08-19_…md    what we believed at the first milestone
  SUMMARY_2026-08-20_…md    the second
  HANDOFF_2026-08-20_…md    the previous handoff — still the best background
  HANDOFF_2026-08-21_…md    this file
  submission_315_divergence_sweep.json   the council submission
```
Elsewhere: `bugs_open/315` · `bugs_open/336` · `architecture_review/RFC_038` · register `DGH-013`
(corrected) and **`DGH-015`** · `LANDMINES.md` (one new entry, one existing entry qualified) ·
`WRONG_CALLS.md` (one new entry) · `sql_for_agents/526_*`.
