# HANDOFF — `bugs_closed/315` — state at 2026-08-22 19:00Z

> **THE BUG IS CLOSED.** `bugs_closed/315_HANDOFF_2026-08-18_…md`. The check is live, and it has now
> been graded against the whole fleet. This file is the record of how it got there plus what is open.

> # ⚠⚠ CORRECTED 2026-08-23 — READ THIS BEFORE §0. THE HEADLINE BELOW IS WRONG.
>
> **§0's "1 true positive, 0 false positives" is INVERTED. It was 0 true positives and
> 1 FALSE positive, and §2's "THE ONLY THING LEFT" was already done before it was written.**
>
> 1. **`vetcomparison.uk/index.html` was never stale.** Cloudflare Web Analytics is enabled
>    on that zone, and Cloudflare injects a ~359-byte
>    `static.cloudflareinsights.com/beacon.min.js` `<script>` into anything it treats as
>    browser HTML. `[MEASURED 2026-08-23, 5 fetches per header]` `Accept: */*` returns
>    **exactly the stored fingerprint 5/5**; the browser header and the check's own
>    `Accept: text/html,*/*` return one other hash 5/5; **a diff of the two bodies is TWO
>    LINES** — the beacon tag and its close. The served `last-modified` was **15 seconds
>    AFTER `deployed_at`**, so the object was current. Every visitor was being served the
>    current page for the whole ~21 hours.
> 2. **The six "independent observations" were one unconditional fact.** On an injecting
>    zone this check can never match, so it re-flags every pass for as long as the feature
>    is on. Repetition was not corroboration.
> 3. **§0's rule 1 is what hid it.** "Send a browser `Accept`" made the difference VISIBLE
>    and then `≠ stored` was read as *"the old page"* — the label "8/8 OLD" in §0 is that
>    inference, not an observation. **A hash tells you two things differ; it cannot tell
>    you HOW.** The missing step cost a day: `diff` the two bodies.
> 4. **FIXED at source** — commit `14a50e533` adds a raw-object probe: after a mismatch
>    confirms itself, a third fetch with `Accept: */*`; if that hashes to the fingerprint,
>    nothing is filed. It cannot hide a real divergence (a stale delivery serves the old
>    object under every header). The false item is `rejected`, with the mechanism in its
>    `result`.
> 5. **§2 is STALE, not pending: `547` and `526` were BOTH applied on or before
>    2026-08-21 21:53Z** — see §2's own corrected banner. Their ledger rows were missing
>    and were recorded 2026-08-23.
> 6. **D8 (60-minute window) is LIVE** — settled at the artefact 2026-08-23; see §0b.

## 0. THE HEADLINE — the check works, and my hand-checks did not

`[MEASURED 2026-08-22 ~18:50Z]` 311 judgeable pages swept, then every flagged page re-probed **5 times**:

**`page_content_divergence` scored 1 true positive, 0 false positives, 0 misses.**

> **⚠ CORRECTED 2026-08-23: this line is FALSE — it was 0 true positives, 1 false
> positive.** The single finding was Cloudflare beacon injection, not a stale page. See
> the banner at the top of this file. The 311-page sweep's *negative* result stands; only
> the grading of the one positive is withdrawn.

The one real case, `vetcomparison.uk/index.html`, has been serving visitors the **old page for ~21
hours** — through a republish at 08:50Z — and the check flagged it on **six consecutive passes**,
unprompted, starting 1h04 after the deploy.

⚠ **That is a LIVE CUSTOMER-FACING FAULT and it is not this lane's to fix** — the delivery step lives
in the private `gqls/sites` runner. What this lane can now do, which it could not before, is name it
exactly: site, page, publish timestamp, and six independent observations.

> **⚠ CORRECTED 2026-08-23: there was NO fault, live or otherwise.** The page was serving
> the current bytes to everyone the whole time. "Six independent observations" was one
> deterministic fact observed six times — on a beacon-injecting zone the check cannot
> match, so it re-flags every pass. **Repetition of an unconditional result is not
> corroboration**, and that it kept firing "unprompted" made it feel like evidence
> accumulating when nothing was accumulating.

### ⚠ Reproducing a finding by hand — read this before disagreeing with the check

I disagreed with it three times in one day and was wrong three times. Both rules are needed:

1. **Send a browser `Accept`.** A Cloudflare Worker (`server-timing: cfWorker`) serves a *different
   body per `Accept`*, from the same B2 object (identical `x-amz-version-id`). On the stale page,
   **8 fetches per header**: browser `Accept` 8/8 OLD, `Accept: text/html,*/*` (what the check sends)
   8/8 OLD, `Accept: */*` (curl's default) 8/8 CURRENT. **A plain curl says the page is fine.**

   > **⚠ CORRECTED 2026-08-23 — THE OBSERVATION IS RIGHT AND EVERY LABEL ON IT IS WRONG.**
   > The bodies really do differ by `Accept`, from one B2 object. But **"OLD" was never
   > observed — it was inferred from `≠ stored hash`.** The browser body was the CURRENT
   > page plus a Cloudflare Insights beacon; the `*/*` body was the same page without it.
   > So "a plain curl says the page is fine" is not curl being fooled — **curl was right**.
   > The rule that actually generalises: **a hash comparison can only tell you THAT two
   > bodies differ. Before you name the difference, `diff` them.** One `diff` here would
   > have cost seconds and saved a day; the whole delta was two lines.
   >
   > The check now does this itself (raw-object probe, `14a50e533`), so a by-hand
   > reproduction should compare **three** things: stored hash, browser body, `*/*` body —
   > and if the `*/*` body matches stored, the delivery is FINE.
2. **Fetch N times and state N.** A single fetch is noise-dominated here — live responses include
   zero-length bodies (`e3b0c44298fc…`, the sha256 of the empty string) and Cloudflare error pages
   (`e3ebaa16dd9d…`). One-fetch sweeps of mine produced "228 of 228 healthy" (under-read) and
   "6 pages stale" (over-read). The truth on 5 fetches each: **one page**.

```bash
BROWSER='Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8'
for i in 1 2 3 4 5; do curl -s -H "$BROWSER" "https://<domain><url>?cb=$RANDOM$RANDOM$$" | sha256sum; done | sort | uniq -c
```

**The check already does both** — it sends an HTML `Accept` and fetches twice requiring agreement.
That is why it was right. I built that guard and applied it to none of my own spot-checks.

## 0b. What is OPEN

| item | state |
|---|---|
| **D8 — window 30→60 min** | ✅ **SETTLED 2026-08-23: LIVE.** Not by the item route (no new item has been filed since, and the one that exists predates D8 — its spec still reads `1800`, which is correct-and-irrelevant). Settled at the artefact instead, via the RFC_040 capability table nobody in this lane had used: `SELECT git_commit FROM service_binary_capabilities WHERE service='agent-chassis' AND name='page_content_divergence'` → `bd454eb93`, and `git merge-base --is-ancestor 971178638 bd454eb93` = YES with the reverse direction NO (so the test discriminates). **That table is the answer to "has my Go change rolled" and it has no shelf life** — unlike the provenance log line, which had already scrolled out of a full `kubectl logs` here. |
| **D9 — escalate on PERSISTENCE across passes** | Unbuilt, and the day's evidence is the argument for it: convergence times (seconds → ~17 min → 1h20 → 21h) OVERLAP the failure, so no settle-window value separates them. The check already re-detects on every pass and the dedup index absorbs it. |
| **D11 — an EDGE THAT ADDS BYTES** | ✅ **BUILT AND COMMITTED 2026-08-23** (`14a50e533`), and it was not hypothetical — it had already produced this check's only production finding. Raw-object probe: after a mismatch confirms itself, refetch with `Accept: */*`; if that hashes to the fingerprint, discard. Three mutations, three distinct test failures. |
| **D10 — an empty/error 200 is HASHED, not skipped** | Unbuilt. Both artefact bodies are STABLE hashes, so the double-fetch guard does **not** filter them: two consecutive empty 200s would agree and file against a healthy page. The one remaining way this check can manufacture a false positive. Cheap. |
| **D6 — unarmed stamper should NULL the hash** | Unbuilt; no unarmed stampers exist today (6 of 6 armed). |
| **Retraction half** | **Still never fired, and still correctly so** — its precondition is a pass that OBSERVES A MATCH, and the page genuinely still diverges. Not evidence of a defect. |

⚠ **`items_inserted: 0` does NOT mean "found nothing"** — read `items_skipped` **and** `findings`. The
stale page has been re-detected on every pass with `items_inserted: 0, items_skipped: 1, findings: 1`,
because `idx_swi_dedup` correctly refuses a duplicate.

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

## 2. THE ONLY THING LEFT — one step, and it is not this lane's to take

> **⚠ CORRECTED 2026-08-23 — ALL OF §2 IS DONE. Nothing here is outstanding.**
> `547` and `526` were both applied on or before **2026-08-21 21:53:04Z** (the moment the
> first `page_content_divergence` item was filed — which requires the check to be enabled,
> and `526` refuses while any unarmed stamper exists, so `547` necessarily preceded it).
> Re-proved 2026-08-23: the recursive census returns **6 of 6 stampers armed**, the
> `availability-discovery-agent` checks array is `[site_unreachable,
> page_content_divergence]`, and rehearsing `547` in a rolled-back transaction aborts on
> its own guard — *"547: already applied"*.
> **Neither had a `schema_migrations` row**; both were recorded 2026-08-23 with notes
> saying the applying session is unknown and that `applied_at` is a proven upper bound.
> That gap mattered: `547` is not a `_HOLD` file, so the next `run-migrations.sh --apply`
> would have taken it and **aborted the whole pending batch** on its already-applied guard.
> ⚠ **Do not re-run either file.**

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
