# HANDOFF — `bugs_open/315` → **CLOSED** (2026-08-21 ~19:45Z)

> **⚠ THE BUG IS CLOSED AND THE LANE'S WORK IS DONE.** `bugs_closed/315_HANDOFF_2026-08-18_…md`.
> Both migrations are APPLIED, the check is LIVE and PROVEN AT THE ARTEFACT, and nothing here is
> waiting on anyone. This file is kept as the record of how it got there.
>
> **What is live:** chassis `v1.0.1322` carries `page_content_divergence` (register `DGH-015`);
> migration `547` armed the last three unarmed deploy-stampers (**six of six armed, zero unarmed**);
> migration `526` enabled the check at 19:23Z. Council `Council-Reviewed:
> 9e8d73b8-f777-4404-a1c7-d8e06af897fb` (APPROVED round 3; rounds 1 and 2 each found a real defect,
> one of them in a guard written to answer the previous round).
>
> **The proof** — the discovery run's own record on a site with 21 judgeable pages, which is what
> distinguishes "ran and found nothing" from "never ran":
> `checks_run: [site_unreachable, page_content_divergence]`, `checks_failed: []`,
> `checks_unregistered: []`, `items_inserted: 0`.
>
> **TWO RESIDUALS, neither reopening the bug, both in `PLAN`:**
> - **D6** — make an unarmed `deployed` stamper NULL the fingerprint rather than leave a stale one.
>   The backstop for the *next* unarmed stamper; there are none today.
> - **D8** — widen the settle window from 30 to 60 minutes at the next build. **The observed delivery
>   tail is ~17 minutes, not the 14 seconds first measured** (see §3), so the current margin is ~1.8x.
>   Left as-is deliberately: a premature finding is flag-only and retracts itself on the next pass.
>
> **If you are picking this up to do anything, do D8** — and re-run the 40-page artefact sample first,
> because that is the measurement that found the error.

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

## 2. THE ONLY THING LEFT — one step, and it is not this lane's to take

**a. The chassis image — ✅ DONE 2026-08-21 ~17:00Z.** `v1.0.1322`, built from commit
`bac1899216fc6406f46cfcf8710f6a74c24276e0`, **contains all seven of this lane's commits** and the
running binary registers the check. Verified at the artefact, with controls both ways:

```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=400 | grep -m1 'build provenance'
#   -> git_commit bac1899216fc6406f46cfcf8710f6a74c24276e0
git merge-base --is-ancestor f715b8c1d bac18992   # YES
git merge-base --is-ancestor bac18992 f715b8c1d   # NO  <- reverse control: the test discriminates
P=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | head -1)
kubectl -n ai-persona-system exec ${P#pod/} -- grep -aq "page_content_divergence" /proc/1/exe          # present
kubectl -n ai-persona-system exec ${P#pod/} -- grep -aq "page_content_divergence_NOT_A_REAL_SYMBOL" /proc/1/exe  # absent (negative control)
```

**b. Apply `547`, then `526`. IN THAT ORDER — and `526` will refuse if you skip `547`.**

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/547_arm_the_three_unarmed_deploy_stampers.sql
# then, and only then:
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/526_enable_page_content_divergence_HOLD.sql
```

**Both are WRITTEN, COMMITTED and PROVEN against live data but NOT APPLIED.**

⚠ **`547` came back REVISE on round 1 and has been revised and resubmitted** (same trail,
`Council-Submitted: 9e8d73b8-f777-4404-a1c7-d8e06af897fb`). **Read the round-2 verdict before
applying.** All three HIGH objections in round 1 were correct, and one of them found this bug's own
defect reproduced inside the migration written to fix it:

- **`substeps` is the half that RUNS.** `LoopAction` reads `config["substeps"]` first and falls back
  to `sub_workflow.steps` only when substeps is absent/empty (`loop_actions.go:91-104`). Arming the
  fallback on a loop carrying both would have created a DEAD key while the executing step stayed
  unarmed — and the migration's own recursive verify would have found the armed dead copy and passed.
  `[MEASURED]` none of the three carries substeps, so the path is right; it is now **gated**, and the
  gate is proven to bite.
- **Duplicate active definition rows.** Four types fleet-wide carry two active rows where only the
  higher version loads. `[MEASURED]` our three carry one each; now **gated**, proven to bite.
- **"0 runs in ALL HISTORY" was FALSE** — read from `orchestration_states`, which reaps terminal rows
  after ~24h (`[MEASURED]` 24 of 3,154 rows older than 48h). Durable source `agent_run_stats`:
  page-rebuild **7**, pageflow-builder **3**, site-work-orchestrator **1**, last activity 2026-08-09.
  **Rare, not dead** — the "behaviourally inert" argument is withdrawn.

**What makes it safe anyway, and this is the better evidence:** the three last ran **2026-08-09 13:50**
and the first `content_hash` was written **2026-08-20 17:36**, eleven days later. They *cannot* have
stranded a stale fingerprint — the column had no values when they last ran. That is why the sweep
found 228 of 228 matching: structural, not lucky. Arming is protective for their NEXT run.

⚠ **One thing no resubmission can discharge:** `guardian` noted that this lane caused a 33-minute
full-estate outage on 2026-08-20 doing exactly this class of change, and that the recurrence warrants
**a second pair of eyes beyond the council**. That is a judgement for the owner, and it is why 547 sits
written-and-proven rather than applied.

⚠ **After EITHER apply, the first query is "what did I break?", not "did it work?"** Both files carry
their damage query in the header. This is `bugs_open/336`'s lesson, and this lane learned it by taking
every page-publish in the estate down for 33 minutes while confirming its config was right.

## 3. What the measurements say, so you do not have to re-derive them

- **The fleet is healthy.** `[MEASURED 2026-08-21 ~10:35Z]` **228 of 228** active pages carrying a
  `content_hash` serve bytes hashing exactly to it, across 12 domains. Expect the check to find
  **nothing** on day one; a finding is likelier to be a defect in the check than a divergence in the
  fleet, so re-run the comparison by hand before believing it (RUNBOOK Part 3).
- **That sweep is also the end-to-end proof of the fingerprint itself** — 228 independent
  confirmations that nothing between the stamp and the wire alters the bytes.
- **The 30-minute settle window is LOAD-BEARING, and its margin was OVERSTATED.** `[MEASURED
  10:38Z–13:20Z]` 1,099 re-probes, 85 pages, 95 deploy events: 3 DIVERGED at ages 1s/13s/14s, all
  converged by 140–156s, 0 of 995 readings at age ≥157s diverged.
  > **⚠ CORRECTED 19:36Z, by re-running the proof after go-live.** A random 40-page sample returned
  > **2 DIVERGED, aged 15 and 21 minutes**. Tracked to convergence: `/model-fine-tuning.html` read
  > MATCH @945s, **DIVERGED @1012s**, MATCH @1079s onward. **The largest observed divergence age is
  > ~17 MINUTES, not 14 seconds** — so the window is ~**1.8x** the worst case, not 128x. The earlier
  > watch sampled only fast deliveries; quoting its maximum as "the tail" was the error. Hence **D8**.
  > **And the shape is not a simple lag:** that page went MATCH → DIVERGED → MATCH, 67s apart.
  > Delivery lands **progressively across edge nodes**, which is why the confirmation fetch must AGREE
  > with the first before anything is filed.
  Still doubly demonstrated as load-bearing: 3 items prevented in the watch, 2 more in one 40-page
  sample.
- **Two intermittent 404s** in the same watch, both serving one shared edge error page, each
  surrounded by MATCH — 2 more false items, prevented by treating a non-200 as unjudgeable.
- **It will be exercised fast** — once it is enabled at all (see §4).
  `site-discovery-rotation-availability` fires every **300s** with a **4-hour** floor (NOT the quality
  rotation's 7 days), so the fleet is swept every ~4–5 hours.

## 4. ⚠ WHY 547 EXISTS — a live hazard this lane documented as impossible

**The check is sound only if every live step that stamps a page `deployed` also records what it
sent.** An unarmed stamper leaves a fingerprint describing an OLDER deploy, and the check then
convicts a healthy page — permanently, because nothing rewrites that row.

`[MEASURED 2026-08-21, RECURSIVE walk]` **six such steps, THREE UNARMED:** `page-rebuild`,
`pageflow-builder`, `site-work-orchestrator`, all at
`workflow.steps.<loop>.config.sub_workflow.steps.update_page_status` — **the page-BUILDING paths**,
i.e. the ones that actually emit new bytes.

> **This corrects an earlier claim of mine that said the opposite** ("exactly three, all armed, zero
> unarmed") which reached the check's header, the register, three commit messages and the council
> submission. My census walked one level. **The `guardian` seat caught it from the SHAPE of the claim
> without seeing the query.** Full account in `WRONG_CALLS.md` and NOTES.

**You do not have to remember this: 526 refuses to apply while any unarmed stamper exists** — proven
to bite ("3 of 6 live steps … do NOT declare deploy_result_field").

**`547` is the fix and it is WRITTEN, COMMITTED, PROVEN AND NOT APPLIED.** It arms all three by naming
their sibling `git_commit` step's `output_field` (`page_deployed`), mirroring 494.

- **Blast radius `[MEASURED]`: zero today.** Arming changes behaviour in exactly one case — the stamp
  is refused when the deploy reported a skip — and all three have **0 runs in ALL HISTORY**, 0
  scheduled tasks, 0 work items routed at them.
- **Reachability, which a run-count cannot answer:** exactly ONE live dispatch reference fleet-wide —
  `maintenance-triage` carrying `agent_type = page-rebuild`. (A substring search over `default_config`
  also hits `council-gate`, `fix-proposer` and `domain-research-classifier` — but there the names are
  **prose inside reviewer prompts**, not wiring. Match VALUES at dispatch keys, not substrings.)
- **Proven three ways against live data, each ending in ROLLBACK with rows untouched:** the migration
  alone; the round trip with its own rollback body; and **the composition with a control** — 547-then-526
  passes every gate, while **526 alone still refuses**. The sequence works *because* of 547.
- **Not a `_HOLD` file, deliberately:** `deploy_result_field` is already declared by the running binary
  (494 armed three agents yesterday; 232 fingerprints written since), so no image-before-config
  constraint applies.

**Why arming rather than PLAN D6's stamp-side NULLing:** arming RAISES fingerprint coverage where
NULLing lowers it. **D6 remains worth doing as the backstop for the NEXT unarmed stamper** — the one
nobody notices being added — and is still unbuilt.

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
