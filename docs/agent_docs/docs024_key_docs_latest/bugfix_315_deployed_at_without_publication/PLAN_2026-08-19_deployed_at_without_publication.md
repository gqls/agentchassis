# PLAN — bug 315: make `pages.deployed_at` mean something, and make publication checkable

**Written 2026-08-19.** Design, phasing, decisions and their reasons. Corrections live here, marked.

## The problem, restated after measuring it

`pages.deployed_at` is read across the estate as "this page has shipped". It is written by five
agents. **Two of them write it before the deploy is even dispatched**; the other three write it
after a commit **whose result they discard**. So there is no arrangement of the current workflows
under which the column could be evidence of publication.

Underneath that sits a second fact that decides the design: **"commit is deploy" is asynchronous and
batched.** The origin is rewritten per changed domain directory by one `b2 sync` on a self-hosted
runner, tens of minutes after the commit (measured: 40 sampled pages sharing a single three-second
`last-modified` window while their stamps spread across the following hour). A synchronous stamp
*cannot* honestly assert publication, however well guarded — at the moment it is written, the claim
is reliably false and becomes true later if nothing goes wrong.

And a third, from register `DGH-009`: **an unchanged file commits as an EMPTY commit** and the
adapter still reports success. So neither "the step succeeded" nor even "a commit exists" implies
bytes moved. Only content does.

## Decisions

### D1 — Split the fact. Keep `deployed_at` as the ever-shipped floor; carry publication evidence separately.

`deployed_at` keeps its current meaning — *this page's artefact was committed to the deploy repo* —
strengthened so it is only written downstream of a real commit result. Publication evidence goes
into a content hash plus a commit sha, and publication itself is confirmed **asynchronously**, by a
sweep, never inside the stamp.

**Why not just redefine/narrow it in the docs:** it leaves 315's actual question permanently
unanswerable, and even the narrowed reading is false for the two agents that stamp pre-deploy.

**Why not make it mean "served bytes confirmed":** the delivery half is outside this repo's process
boundary and minutes away in time. Worse, ~15 consumers depend on the current semantic as an
*ever-shipped floor* (`queryresolve.go` eligibility, `check_componentless_pages`,
`maintenance_actions` staleness). Redefining it would let a transient CDN failure start delisting
working pages — worse than the bug.

**Why the split is cheap:** it reuses `pages.content_hash` (dead, 0/786, `varchar(64)` — exactly a
sha256 hex), `page_components.deploy_commit` (dead, 0/1,775), the `publish_site` served-bytes
acceptance idiom, and the discovery-check framework. The site-level publish seam already made this
exact design choice one level up; this is the same idea at page granularity, with acceptance
detached in time because delivery is detached in time.

> **Note, and it is the strongest argument for the shape:** the platform already *retired a config
> key for this feature*. `UpdatePageStatusInputSpec.RemovedConfigKeys["commit_from"]` says the intent
> was *"recording which git commit a page's content was DEPLOYED in, from the git_commit step's
> output … unimplemented — pages has no such column. **Implement it as a feature if wanted**, do not
> re-add the key."* And the page-rerender seed's output contract already promises
> `"deploy_result": "git commit result with commit_sha"` — a promise the adapter has never kept.
> This plan is that feature, under a new name as the retirement note requires.

### D2 — Make the layer that knows say what it knows.

`CommitToRepo` computes `newCommitSHA` and returns `repo.HTMLURL` (a per-repo constant). Change it
to return an outcome struct carrying the sha, a `no_change` flag (new tree sha == parent's tree sha)
and, for deletions, the absent paths. The adapter's success payload then carries `commit_sha`,
`no_change` and `files_sha256` (path → sha256 of the exact bytes committed; the adapter holds them).

Cost is small and was verified rather than assumed: **3 production callers, all inside the
git-adapter**; `CreateBranch`/`CreatePullRequest` do not call it; the delete tests are `_, err :=`
and compile unchanged.

**`no_change` is REPORT-ONLY in v1.** Skipping the empty commit would change what the shared seam
guarantees for all 19 `git_commit` steps (today every success fires the deploy workflow and a cache
purge). That is architecture-scope and is deliberately excluded.

### D3 — The stamp guard reads a CONFIG-NAMED field, never a literal one.

New optional config key on `update_page_status` naming the deploy step's output field. When set, the
`deployed` branch refuses the stamp on a skip/failure and writes the evidence columns on success.

**This is not stylistic.** `[MEASURED]` the 19 live `git_commit` steps carry **nine different
`output_field` names** and two set none; `deploy_result` names only three of them. A guard
hard-coding `deploy_result` would be blind on 16 of 19 and — since it must fail open — would wave
them through silently, which is this bug reintroduced inside its own fix.

Fail-open is required (a config typo must not freeze deploy stamps fleet-wide) but must be
**countable**: a durable `agent_error_log` row whenever the field is set and unreadable. A silent
fail-open is how this bug survived four rerenders.

⚠ **The verdict must be resolved through the estate's existing resolver, not by indexing a path.**
`deploy_result` has two shapes — inline (`…response.data.success`) and, when the deploy is done by a
called sub-agent, one level deeper (`…response.deploy_result.response.data.success`). That second
shape is **57 of 744 rows (7.7%)** over 7 days. `datahelpers`' whole-tree search and
`tryUnwrapMapPatterns`/`UnwrapDeep` exist for exactly this; do not hand-roll a second unwrapper.

### D4 — The two pre-deploy stampers are a CONFIG reorder, not a guard.

`page-build-handler` and `tool-recreation-handler` stamp before any deploy exists *by construction* —
no guard can read evidence that is not there. Both then call `page-rerender`, **which already
contains its own post-commit `update_status`**. So the fix is to delete their early stamp step and
let the rerender's stamp stand.

DB-only, live immediately, and **on its own it takes "2 of 5 stamp before the deploy" to zero.**

### D5 — The divergence sweep compares HASHES, not timestamps, and only after a settle window.

A new discovery check comparing `pages.content_hash` against the sha256 of served bytes fetched with
a cache-buster.

**Not timestamps.** The bug file's §5 candidate 4 proposes `deployed_at` vs origin `last-modified`.
That check would have convicted the two *healthy* pages in the bug's own §2 table, because a
byte-identical rerender legitimately rewrites nothing. And `[MEASURED]` it flags 40 of 40 sampled
pages at any given instant, because the origin lags in batches. **It cannot separate "not synced
yet" from "will never sync"** — only elapsed time can, and the known bad case took six hours.

False-positive sources and their mitigations: mid-delivery window → require `deployed_at` older than
a configurable settle window; CDN edge → cache-buster and assert the origin-header shape; retracted
and archived pages (which deliberately keep `deployed_at`) → predicate on `status='active'`; pages
with no hash yet → predicate `content_hash IS NOT NULL`, so the check is structurally inert until D2
has shipped and reports nothing before then.

> **HARD CONSTRAINT, measured on this lane 2026-08-19 — a settle window is NOT sufficient.**
> I ran candidate 4's own method (origin `last-modified` vs `deployed_at`) across 40 live pages and
> got a clean, confident **40 of 40 stale**. Every one of them was **fine**: the served bytes matched
> the current database content, the pages simply had not needed rewriting, and the apparent staleness
> persisted for **85 minutes** — well past any settle window one would reasonably configure. A sweep
> built on timestamps, with or without a grace period, would have filed 40 false work items on the
> fleet's busiest site. **The hash is not an optimisation over the timestamp comparison; it is the
> only version of this check that works at all**, because "did not need publishing" and "failed to
> publish" are indistinguishable in every signal the platform exposes today.

Emission: a new item type, per-page dedup key, **no handler agent in v1** — re-filing a rerender is
the loop that already failed four times; the honest first move is visibility.

## Phasing

Ordering matters in a way that is the inverse of the usual rule. `UpdatePageStatusInputSpec` is
**`StrictConfig: true`**, so a config naming a key the running binary does not declare fails
validation. **The image must ship before the config**, and that is a real, stated ordering
constraint rather than a convenience.

| phase | kind | live when | useful alone? |
|---|---|---|---|
| 0a — drop the pre-deploy stamp from the two handlers (D4) | config migration | immediately | **yes** — kills the worst half by itself |
| 0b — schedule the already-built, undriven `deployed_zero_components` detector via its existing `emit_checks` override | config | immediately | yes — detection exists and has never run |
| 1 — `CommitOutcome` + payload `commit_sha`/`files_sha256`/`no_change` (D2) | Go, **git-adapter image** | on rebuild + roll | yes — every deploy becomes auditable |
| 2 — the guard + evidence writes; migration adding `pages.deploy_commit` (D3) | Go, **chassis image** + migration | on rebuild + roll | with 1 |
| 3 — name the deploy field on the three post-commit stamping steps | config migration | **must follow phase 2's roll** (StrictConfig) | — |
| 4 — the divergence check + enabling it (D5) | Go + config | chassis roll, then config | gated on 1–3 |

## Out of scope, and why

- **The `gqls/sites` GitHub Actions workflow** — it lives outside this repo and the repo is private,
  so §5 candidate 3 ("why is this page skipped by the batch") cannot be diagnosed from here. D5 is
  deliberately designed to *detect* that failure from this side without reading the runner.
- **The 42 `planned` componentless pages** — never built, nothing served; a different defect already
  owned by `check_componentless_pages` / `nav_linked_never_built`. Phase 0b starts surfacing them
  incidentally; repairing them is the build pipeline's job.
- **`bugs_open/083`'s stranded `needs_human_review` items** — that bug's disease. D5 deliberately
  does not emit into `needs_human_review`, so it cannot inherit the stranding.

## Architecture-scope assessment

**Not architecture-scope**, with the reasons: the new config key is opt-in with the unsafe default
OFF (absent ⇒ today's behaviour) and **zero live steps name it** — the exact RFC_022 shape, and the
enumeration is the evidence, not the assertion. The adapter payload additions are
additive-and-inert (no chassis code reads any payload field today). The new item type has no
automated consumer, which is fine under the 2026-08-02 §1 ruling **provided** the producer and key
shape are named in the concept register in the same commit.

**Deliberately excluded as RFC-shaped:** suppressing the empty commit on `no_change`, and any
synchronous served-bytes gate inside the stamp.

**Consumers to TELL, not merely measure** (2026-07-29 ruling 3): the 16 agents carrying `git_commit`
steps; the `webdesign_tool_rebuilds` lane that filed this; retraction owners (`deployed_at`
semantics for non-active pages are unchanged, and D5 depends on that); and `deployment-github.md`
DGH-001, updated in the same commit as each shipping phase.

## Status — updated 2026-08-19 late evening (council **APPROVED**, round 3 of trail `377167cd`)

| phase | state |
|---|---|
| **0a** — drop the pre-deploy stamp (D4) | **APPLIED and LIVE** (migration 491, 15:20Z). Verified at the config AND at runtime: 31 pages since, all `deployed`, none stranded, stamps coming from the rerender |
| **0b** — drive the silenced `deployed_zero_components` detector | **not done** — still undriven |
| **1** — adapter returns `CommitOutcome` + `commit_sha`/`files_sha256` (D2) | **BUILT, committed `0c5b94725`, NOT ROLLED** (needs a git-adapter image) |
| **2** — `deploy_result_field` guard + `pages.content_hash` (D3) | **BUILT and APPROVED.** `086f9b7b7` shipped in `v1.0.1316` but its resolver was unsafe; **corrected by `f0dd97c71`, which is NOT yet rolled.** One further commit (the derived placeholder + `buildPageDeployStampQuery`) is written, tested and **BLOCKED** — another session's uncommitted symbol shares `v3_site_actions.go` |
| **3** — arm the key on the three stamping steps | **WRITTEN and HELD** — `sql_for_agents/494_stamp_reads_deploy_evidence_HOLD.sql`. Must not be applied until the rebuilt chassis runs: `StrictConfig` makes an undeclared key a validation FAILURE, not a no-op |
| **4** — the divergence sweep (D5) | **not built.** Gated on `content_hash` being populated, which is gated on the roll |

**Two deliberate narrowings of this plan, made while building it:**

- **No `no_change` flag.** It needs the parent commit's tree sha, which is off the hot path
  (`getBaseTreeSHA` is only `getLatestCommitSHA`'s error fallback), so it would cost a GitHub
  round-trip on every commit across 19 live `git_commit` steps — to populate a field the council
  ruled report-only. `files_sha256` answers the same question at the grain the site serves.
- **`DEPLOY_EVIDENCE_UNREADABLE` logs at `warning`, not `high`.** Chassis and git-adapter are
  separate images, so a chassis carrying the key against an older adapter resolves nothing on
  *every* deploy. That window is expected and bounded; `high` would make the fleet error log
  useless for a day.

**And one edit withdrawn** (round-1 `prior_art_librarian`): the plan's `ALTER TABLE pages ADD COLUMN
deploy_commit` is gone. `sql_for_agents/356` records that column was dropped **deliberately** as
"belongs in page_components", and that wiring it "is an owner call, not a bug fix". Owner ruled
2026-08-19: wire `pages.content_hash`, leave `page_components.deploy_commit` alone.

⚠ **D5's hard constraint, measured rather than assumed** — see the block under D5: a settle window
does NOT rescue a timestamp comparison. Only the hash separates "did not need publishing" from
"failed to publish".
The diagnosis loop produced no verdict (LLM usage cap, twice) — see NOTES for the first-hand
substitute, stated plainly per the owner ruling of 2026-07-31.

---

## Council gate

Submitted 2026-08-19 (phases 0–2 — the config reorder, the adapter return contract, the guard).

**`SUBMISSION_CORR = 377167cd-6324-4bc7-a866-87ad8c435132`**

Five edits, twelve `grounded_in` evidence quotes, risks stated including the two `[UNVERIFIED]`
items (whether the two handlers' `complete_workflow` output fields tolerate the removed
`status_updated`, and whether anything downstream of the commit rewrites page bytes — the private
`gqls/sites` workflow could not be read).

Read the verdict with:
```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='377167cd-6324-4bc7-a866-87ad8c435132' AND kind='council_report' ORDER BY created_at;
```

---

## D5 — BUILT 2026-08-21 (phase 4), and what the build changed about the design

`check_page_content_divergence.go` + its test file, committed `f715b8c1d`; registered as **DGH-015**;
enabling migration held at `sql_for_agents/526_enable_page_content_divergence_HOLD.sql`.
Council round: `SUBMISSION_CORR = be85a6d3-f2c0-4f7a-b791-e95087141fc8`.

**The design above survived contact, with three additions the build forced.** Recording them here
because a plan that only ever gets ticked off is not a record of anything.

1. **Two race guards D5 did not name.** The plan's false-positive list had four entries
   (mid-delivery, CDN edge, retracted pages, no hash). Building it surfaced two more, both of which
   file a work item against a perfectly healthy page:
   - *the page is redeployed while we are probing it* — we read hash H1, a deploy writes H2 and new
     bytes, we fetch the new bytes and convict H1. Closed by re-reading `content_hash` AND
     `deployed_at` after a mismatch and discarding the candidate if either moved.
   - *the origin is mid-write* — a sync in progress can answer with one body and then another.
     Closed by confirming a mismatch with a second fetch and requiring the two served hashes to
     AGREE. Two different bodies is a moving target, not a divergence.

2. **The settle window is now MEASURED, and it turns out to be load-bearing rather than
   precautionary.** [MEASURED 2026-08-21, 10:38Z–13:20Z] 1,099 re-probes over 85 pages and 95
   deploy events: the only 3 DIVERGED readings were at ages **1s, 13s and 14s**, all converged by
   140–156s, and **0 of 995** readings at age ≥157s diverged. Those three readings are three work
   items this check would have filed against healthy pages in under three hours had it judged them
   at the moment of stamping. Two intermittent 404s in the same watch (both serving one shared edge
   error page, each surrounded by MATCH readings) are two more, had a non-200 been judged as content
   rather than skipped. **The window stays at 30 minutes** — roughly 128x the largest lag observed —
   deliberately conservative because one afternoon on one estate is not a census and the sync is
   BATCHED, which is exactly the case a small sample under-represents.

3. **The check has NO live positive, and that is now a measured fact rather than a hope.**
   [MEASURED 2026-08-21] all **228 of 228** active pages then carrying a `content_hash` served bytes
   hashing exactly to their stored fingerprint, across 12 domains. So it ships as a REGRESSION GUARD
   (the `check_asset_reference_404` posture) with every branch proved by an induced fault. That same
   sweep is also the end-to-end proof of D2/D3: stored fingerprint and served bytes agree on 228
   independent pages, so no encoding, transform or path-keying error sits anywhere between the stamp
   and the wire.

### D6 / D7 — an UNARMED `deployed` stamper poisons this check, and THREE OF THEM ARE LIVE

> **⚠ THIS SECTION WAS REWRITTEN 2026-08-21, HOURS AFTER IT WAS FIRST WRITTEN AND BEFORE THE CHECK
> WAS EVER ENABLED. The version it replaces was built on a FALSE measurement**, and the wrong version
> is kept here in outline because how it was wrong is the useful part. It said: *"It cannot arise
> today, and that is a query rather than an argument: exactly three live steps set `status='deployed'`
> ... and all three declare `deploy_result_field`. Zero unarmed stampers."* It then reasoned, at
> length and quite correctly given that premise, that the structural fix was inert and could safely
> wait.
>
> **The premise was wrong. There are SIX stampers and THREE are unarmed.** The census walked
> `default_config.<workflow>.steps.*` — one level. The three it missed sit at
> `workflow.steps.<loop>.config.sub_workflow.steps.update_page_status`.
>
> **What caught it: the council gate's `guardian` seat**, round 1 of `be85a6d3`, which said the claim
> "was almost certainly measured with a top-level `workflow.steps` census — documented elsewhere on
> this platform as blind to actions nested inside `sub_workflow`/substeps", and asked for
> re-verification "against a query that walks `sub_workflow` too before it's trusted as an
> invariant". It never saw the query. It inferred the blindness from the SHAPE of the claim — a
> confident small integer about "every step that does X" — and it was right. The verdict was
> APPROVED; had I read only the decision and not the objections, this would have shipped.

**The mechanism.** The stamp assigns `pages.content_hash` only when the deploy-evidence guard RAN
(`v3_site_actions.go`); an unarmed `update_page_status` leaves the column alone. That is correct for
an unarmed path that changes nothing, and WRONG for one that deploys NEW BYTES: the fingerprint then
describes an older deploy, and D5's check convicts a healthy page — permanently, since nothing
rewrites that row.

**`[MEASURED 2026-08-21, RECURSIVE walk]` — six stampers, three unarmed:**

| agent | path | armed with |
|---|---|---|
| page-rerender | `workflow.steps.update_status` | `deploy_result` |
| report-builder | `workflow.steps.update_status` | `deploy_result` |
| section-editor | `workflow.steps.update_page_status` | `git_result` |
| **page-rebuild** | `workflow.steps.build_pages_loop.config.sub_workflow.steps.update_page_status` | **UNARMED** |
| **pageflow-builder** | `workflow.steps.build_pages_loop.config.sub_workflow.steps.update_page_status` | **UNARMED** |
| **site-work-orchestrator** | `workflow.steps.build_items_loop.config.sub_workflow.steps.update_page_status` | **UNARMED** |

The three unarmed ones are the page-BUILDING paths — the ones that actually emit new bytes. This is
the main road, not a corner case.

**Nothing is poisoned today, and that is luck, not safety.** A stale fingerprint presents as exactly
the mismatch D5 looks for, and the 228-page sweep found 228 MATCH — so at 2026-08-21 10:35Z no page
was in that state. That is an observation about one moment.

#### What was done about it immediately (D6a) — the precondition is now ENFORCED

`526_enable_page_content_divergence_HOLD.sql` carries an **unarmed-stamper gate**: it RAISEs and
refuses to apply while the recursive count is non-zero, and *also* refuses if the walk returns zero
TOTAL stampers, since a jsonpath that matches nothing is indistinguishable from a fleet with none —
the same false-zero shape that produced the wrong claim. **Proven to bite against live data**: it
currently refuses with "3 of 6 live steps ... do NOT declare deploy_result_field". So the check
cannot be switched on into the hazard, and no one has to remember this.

#### D7 — ARM THE THREE. The preferred fix, and NOT taken in this round.

All three carry a `git_commit` step `deploy_page` with `output_field: "page_deployed"`, so arming
them is one migration in **494's exact shape**, needle-gated the same way.

**Why arming beats D6's stamp-side NULLing as the answer to *these three*:** arming RAISES fingerprint
coverage (three more paths start recording what they sent, on the busiest routes in the estate),
whereas NULLing LOWERS it (a page rebuilt by `page-rebuild` would lose its fingerprint until an armed
path redeployed it). D6 remains worth doing as the **backstop for the NEXT unarmed stamper** — the
one nobody notices being added — but it is not the fix for a stamper we can simply arm.

**Why it is not in this round.** Arming changes behaviour on the main build path: an armed stamper
REFUSES the `deployed` stamp when its commit reported a skip. That is the intended semantics and it
is a real change to the busiest path in the estate, and the last time this lane armed stampers it
took the fleet's page-publishing down for 33 minutes (`bugs_open/336`). It deserves its own council
round and its own damage query, not a paragraph inside a detection change.

**Whoever takes D7:** re-run the RECURSIVE enumeration first (RUNBOOK Part 3) — the count can change
without anyone touching this code, which is the whole reason 526 gates on it rather than trusting a
number written in a file.


### D7 — STATUS 2026-08-21 evening: WRITTEN, COMMITTED, PROVEN, NOT APPLIED

`sql_for_agents/547_arm_the_three_unarmed_deploy_stampers.sql` (+ `_ROLLBACK`), council round
`Council-Submitted: 9e8d73b8-f777-4404-a1c7-d8e06af897fb`.

Also settled today, and it is what makes 547 low-risk rather than a change to the busiest path in the
estate — the fear the first draft of this section was written under:

- `[MEASURED]` **0 runs in ALL HISTORY** for all three; 0 scheduled tasks; 0 work items routed at them.
- `[MEASURED]` exactly **one** live dispatch reference fleet-wide: `maintenance-triage` →
  `agent_type = page-rebuild`. The other three "references" a substring search reports are PROSE in
  reviewer prompts. **Match values at dispatch keys, not substrings** — the same class of error as the
  one-level census that produced the original false claim.
- So the "it changes behaviour on the main build path" caution stands in principle and is **empty in
  practice today**: the path is dormant. Arming it is protective for the moment it wakes.

The chassis image carrying the check is live (`v1.0.1322` / `bac18992`, verified at the artefact with
controls), so the only thing between this lane and a working divergence sweep is applying 547 then 526.


---

## 2026-08-21 ~19:40Z — APPLIED AND PROVEN LIVE, plus a correction to D5's own margin

**Both migrations are APPLIED.** `547` (arm the three) then `526` (enable the check), in that order,
each followed by its damage query before its benefit query.

- After `547`: **all six deployed-stampers armed, zero unarmed**; 0 `deploy_result_field` errors,
  0 `DEPLOY_EVIDENCE_UNREADABLE`, 0 FAILED orchestrations, `page_rerender` claimed/failed unchanged.
- After `526`: the availability agent ran at 19:24 (agritec.uk) and 19:30
  (loanandmortgagecalculator.co.uk), both COMPLETED. The second is the one that counts, because the
  first site has **0 active pages** and a zero from a site with no candidates is not evidence.

**THE PROOF AT THE ARTEFACT** — the run's own record, which distinguishes "ran and found nothing"
from "never ran":

```
checks_requested: [site_unreachable, page_content_divergence]
checks_run:       [site_unreachable, page_content_divergence]
checks_failed:    []        checks_unregistered: []
items_inserted:   0         (site had 21 judgeable pages)
```

`checks_unregistered: []` is the field that would have shown the failure mode `526`'s hold existed to
prevent. It is empty.

### ⚠ D5 CORRECTION — the settle window's margin was overstated by two orders of magnitude

D5's build record above says the window is "roughly 128x the largest lag observed". **That is wrong,
and it was found by accident** while re-running the artefact proof after go-live.

A random 40-page sample returned **2 DIVERGED**, both on fundamentallyai.com, aged **15 and 21
minutes**. Tracked to convergence:

| page | trace |
|---|---|
| `/model-fine-tuning.html` | MATCH @945s → **DIVERGED @1012s** → MATCH @1079s onward |
| `/tools/automation-savings-estimator/index.html` | MATCH @1293s onward |

**Largest observed divergence age is ~1012s (~17 minutes), not 14 seconds.** The window is therefore
about **1.8x** the worst observed case, not 128x. The original 2h42m watch was not wrong about what it
saw — it happened to catch only fast deliveries, and quoting its maximum as "the tail" was the error.
Same shape as reading a retention-bounded table as a lifetime, which this lane also did today.

**And the shape is not a simple lag.** `/model-fine-tuning.html` went MATCH → DIVERGED → MATCH,
non-monotonic, 67 seconds apart. Delivery lands **progressively across edge nodes**, so a probe gets
whichever version the answering node holds. That is precisely why the confirmation fetch must AGREE
with the first before anything is filed — two probes seconds apart can legitimately hit different
nodes. The guard was designed for "origin mid-write" and turns out to be load-bearing for edge
propagation too.

**The window is still load-bearing, and now doubly demonstrated:** those 2 pages are 2 work items the
check would have filed against healthy pages in a single 40-page sample.

### D8 — widen the settle window to 60 minutes at the next build. NOT TAKEN YET.

30 minutes against a 17-minute worst case is a thin margin, and nothing sampled so far bounds a
*slow* delivery batch. Widening costs a rebuild and a fleet roll, and the cost of the wider window is
that a real divergence stays invisible for its first hour — still comfortable against the six-hour
case this check exists for.

It is left at 30 for now because **the failure mode is bounded and self-clearing**: a premature
finding is FLAG-ONLY (no handler, nothing acts on it) and is RETRACTED on the next pass's positive
re-observation. So the cost of being wrong is a work item that clears itself, not damage. Whoever
takes D8 should re-run the 40-page sample first — it is the measurement that found this.


---

## D8 — DONE 2026-08-22: settle window widened 30 → 60 minutes

At the owner's instruction. `divergenceSettleWindow = 60 * time.Minute`, pinned by
`TestDivergenceSettleWindowIsPinnedAndReachesTheQuery` — which asserts the LITERAL 60 (not the const,
which would be self-referential and unfailable) **and** that the value reaches the query as `$2`,
because a constant nothing binds is a documented intention rather than a behaviour. Both halves
mutation-proved: reverting the const fails the test, and swapping the bound value for a fixed 1800
fails it too.

**What it buys, stated honestly:** it clears the ~17-minute propagation tail by ~3.5x instead of ~1.8x.
**What it does not buy:** immunity from the 1h07 case observed on 2026-08-22, which would still be
filed at 60 minutes. **What it costs:** close to nothing — a real divergence is invisible for its first
hour, but detection latency is set by the **4-hour discovery rotation**, not by this constant.

Inert until the next chassis build and roll.

### ⚠ D9 — THE WINDOW IS THE WRONG INSTRUMENT, and the 2026-08-22 catch is why. NOT BUILT.

Three measurements, each an order of magnitude apart, now bound the same phenomenon:

| observation | age at which the page was still wrong |
|---|---|
| 2026-08-21 watch (95 deploy events) | 1s, 13s, 14s |
| 2026-08-21 40-page sample | 15 min, 21 min (max observed 1012s) |
| 2026-08-22 vetcomparison, deploy B | **1h07** |
| 2026-08-22 vetcomparison, deploy A | **9+ hours** (the true positive) |

**These OVERLAP the failure they are supposed to be distinguished from.** No value of
`divergenceSettleWindow` separates "still arriving" from "never arriving", because on this estate a
healthy delivery and a broken one look identical at every age up to hours. Widening trades
false-positive margin for detection latency and never resolves the ambiguity; 60 minutes is a better
blunt instrument than 30, not a fix.

**What DOES separate them is PERSISTENCE ACROSS PASSES.** A page still diverged on a *later* pass —
4 hours on, a different probe, a different moment — is not in flight. The check already has this
information and discards it: the 2026-08-22 item was re-detected on three consecutive passes
(`items_skipped: 1`, `findings: 1` each time) and the dedup index silently absorbed every repeat.

**Shape of the fix (unbuilt):** file at `detected` on first observation as now, but only escalate —
severity, or a handler, or simply a `strikes` count in the spec — once a *second* pass observes the
same `stored_hash`/`served_hash` pair. That distinguishes "slow" from "stuck" using evidence the
system already produces, and it would have marked the 9-hour case as stuck at 01:54 while leaving
every sub-hour case as a single quiet flag. It also makes the settle window much less load-bearing,
which is the point: **a threshold that has been wrong by two orders of magnitude twice should not be
the thing the mechanism rests on.**

⚠ Whoever builds it: `CheckResult.Resolved` retracts on a positive re-observation, so "strikes" must
survive a retraction cycle or a flapping page will oscillate. And re-detection is NOT verification
(RFC_017) — a second observation of the same defect is evidence of PERSISTENCE, not of correctness,
and the entry must say which of the two it is claiming.


### D10 — an empty or error-page 200 is HASHED rather than skipped. NOT BUILT.

`fetchServedPage` hashes whatever a 200 returns. `[MEASURED 2026-08-22]` live responses on this estate
include zero-length bodies (`sha256 = e3b0c44298fc…`, the empty string) and Cloudflare error pages
(`e3ebaa16dd9d…`). Both are STABLE hashes, so the double-fetch agreement guard does **not** filter
them: two consecutive empty responses agree with each other and the check would file a divergence
against a healthy page.

Not yet observed in a filed item — the single item to date is a genuine stale-content case — but it is
the one remaining way this check can manufacture a false positive, and it is cheap to close: treat a
zero-length 200 as unjudgeable (the same posture as the oversize-body skip, which already refuses to
hash what it cannot hash honestly), and consider refusing a body with no `<html`. Both belong in
`fetchServedPage`, beside the oversize check, with the same "skip, do not guess" comment.
