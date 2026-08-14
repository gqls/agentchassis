# HANDOFF — `bugs_open/215` quiet mode, continue here (2026-08-12 afternoon)

Supersedes `HANDOFF_2026-08-11_215_quiet_mode_continue_here.md`, whose §4 ("DO THIS
FIRST — read the dark-launch population") is **done** and whose O1 is **decided and
executed**. Read that file only for history; its §6 traps and §7 loose ends are
carried forward below where still live.

**The pilot is live and inert. Nothing is half-finished in the tree.**

## 1. State

| | |
|---|---|
| code | **LIVE on chassis `v1.0.1297`** (rolled 2026-08-13 22:29Z; `ARCHIVED_PAGE_DEPLOY_REFUSED` present on BOTH replicas, one-letter near-miss absent, `OWNED_PAGE_GUARD` positive control present, re-probed 2026-08-14 07:40Z. **The provenance line had already rotated at ~9h — an empty grep there means OUT OF RANGE, not unstamped**; the literal probe is the fallback precisely because it has no shelf life). Previously: **LIVE on chassis `v1.0.1295`** (rolled 2026-08-13 13:53Z; lane literal + `ARCHIVED_PAGE_GUARD` both artefact-verified on BOTH replicas with a one-letter near-miss and a pre-lane control; provenance `69612d692a4a…`). Superseded the 1293 verification below, which read: **LIVE on chassis `v1.0.1293`** (rolled 2026-08-12 19:13–19:14Z), **re-verified on BOTH replicas 2026-08-12 between the 19:13Z roll and commit `580af7ff0` (19:46Z) by two independent methods.** (1) Literal probe of `/proc/1/exe`: `PLAN_PAGE_STEM_TWIN_REFUSED` **present**, one-letter near-miss `…REFUSEE` **absent** (so the probe can fail), pre-lane control `OWNED_PAGE_GUARD` **present** (so it works on this binary) — all three on each replica. (2) Provenance stamp, captured while still in log range: both replicas built from `7a1887e3163af75ce…`, and `git merge-base --is-ancestor` confirms `19acfc895` (the `carryForwardStructureSpecKeys` re-adoption fix) **is in the build**. Supersedes the `v1.0.1291` verification of ~16:00Z |
| council | **APPROVED** round 3, corr `56e13695-17cb-48ec-bc6b-0371fde8b717` |
| enabled on | **fundamentallyai.com only** — `honour_realised_identity` + `twin_identity_snap`. `stem_twin_snap` absent by design |
| dark-launch counters | **still 0/0/0/0** — re-read twice on 2026-08-12 — once **before** the `v1.0.1293` roll and once **after** it, both times with BOTH controls. Demand: **0 `site_plans` rows since the roll**, and the only post-1291 plan was noted.co.uk's first build (0 `pages` predating it, so it cannot exercise the reconciler). Instrument: `agent_error_log` took 3,503 rows in 24h and **13 since the roll**, so the query is not blind on the new binary either. **No replan has run through the new path yet — the zero is want of demand, not want of function** |
| damage remaining | **7 both-deployed twin pairs, 4 domains** — untouched, needs an owner call per pair (**O2, the only open decision**) |
| bug file | stays **OPEN**; newest sections at the bottom of `bugs_open/215_HANDOFF_...md` |
| register | **PLAN-048** in `docs026_concept_register/register/site-plan-and-reconciler.md` |

## 2. What changed on 2026-08-12

- Read the §4 population as instructed. **Zero, for want of demand** — the only plan
  since the roll was noted.co.uk's FIRST build (its `pages` created 0.65s *after* the
  plan row, so nothing to reconcile against); fundamentallyai's plan predates the
  21:53Z roll.
- **The fleet rolled `v1.0.1288` → `v1.0.1290` underneath this lane.** Re-verified:
  all four lane commits are ancestors of HEAD, literal probe present on both new
  replicas with a pre-lane positive control and two one-letter near-miss negatives.
- **O1 decided by the owner and executed** — see §3.
- Resolved the "five decomposed sites" exclusion to actual domains. **It is six**, and
  **finetuning.uk is both decomposed AND a twin domain** — an overlap no document had.
- Filed a LANDMINE on the counters' semantics (§4 below) and a `WRONG_CALLS.md` entry
  against my own mis-stamped measurement time.

## 3. The pilot, exactly as seeded

`SEED_2026-08-12_fundamentallyai_identity_gates.sql` in this directory. Spec row
`c4c6b829-8e70-4048-a8c2-a050112ff72d`, `created_by brochure_215_quiet_mode_thread`.

**It was an INSERT, not the carry-forward every sibling `SEED_*.sql` does** —
fundamentallyai had no `structure` spec row at all (framework-built, never adopted).
Safe because the only other reader of the aspect, `siteUsesFlatURLs`, states its own
contract as "absent spec, absent key … all mean false" (`site_url_shape.go:29-32`).
**If you seed a third site, re-check that** — a present row with a differently
defaulting key would re-shape live URLs as a side effect of enabling a page-identity gate.

`stem_twin_snap` is asserted **ABSENT**, not false, and the seed aborts if the key
exists at all. Absent and false are identical to the code and different to a human
reader, and O2 is open.

## 4. DO THIS FIRST next session — but read it correctly

```sql
SELECT error_code, count(*), min(occurred_at), max(occurred_at)
FROM agent_error_log
WHERE error_code IN ('PLAN_PAGE_IDENTITY_TWIN_OBSERVED','PLAN_PAGE_STEM_TWIN_OBSERVED',
                     'PLAN_PAGE_STEM_TWIN_REFUSED','PLAN_PAGE_IDENTITY_SNAPPED')
GROUP BY 1 ORDER BY 1;
```

**Two rules, both learned the hard way, both now in `LANDMINES.md`:**

1. **A zero means "no replan yet" until you prove otherwise.** The demand control is
   `SELECT max(created_at) FROM site_plans` **plus** checking whether that plan's site
   had `pages` predating it — a first build cannot exercise the reconciler at all.
2. **A non-zero must be classified before it is quoted.** With a gate off, the layer
   records and then *lets the duplicate through*; and an `*_OBSERVED` row means one of
   two opposite things — harmless re-detection of an existing twin, or a freshly minted
   one. Join the row's `plan_name` back against `pages` to tell them apart. The query
   is in the LANDMINES entry "The page-identity dark-launch counter is NOT a passive
   instrument".

**A roll LANDED at 2026-08-12 19:13–19:14Z (`v1.0.1293`) — two consequences for this query.**
The counters live in `agent_error_log`, which the roll does **not** clear, so a non-zero may
predate it: **bound the read with `occurred_at > '2026-08-12 19:13:00+00'`** if you are asking
what the NEW binary did. And CLAUDE.md's standing rule applies before you try to induce a
signal: **no orchestration dispatch within ~300s of a chassis pod (re)start** — the spawn is
silently dropped, which reads exactly like "the gate did not fire". (That window closed long before this was written; it will matter again on the next roll.)

**Expected first signal on fundamentallyai: ~2 `PLAN_PAGE_STEM_TWIN_OBSERVED` rows**,
the harmless kind (both sides of both pairs already realised). That is the evidence
`stem_twin_snap` is waiting on.

## 5. The one open decision — O2, the 7 pairs

Unchanged and owner's. Procedure, population and ordering:
`RUNBOOK_2026-08-11_duplicate_page_identity_remediation.md`. Needs a survivor decision
per pair before anything executes. Two constraints the runbook does not both state:

- The fundamentallyai sweep front owns that site's execution — route through its handoff.
- **finetuning.uk's pair carries the `bugs_open/204` constraint too** (it is decomposed).
- **Do not enable `stem_twin_snap` on fundamentallyai as a shortcut**: it would snap
  each bare plan entry onto the `tool-` page, i.e. pick the survivor by machine, and
  both pairs are 3 components against 3.

**The side-by-side is DONE and waiting on the owner:**
`DECISION_INPUT_2026-08-12_seven_twin_pairs.md` in this directory — all seven pairs
with what each version actually *serves* (inputs, forms, word counts, byte sizes
against per-domain 404 controls), a recommendation and confidence per pair, and the
one input I could not measure (search indexing).

**Read its headline before you use any earlier figure about these pairs.** Component
count — the runbook's step-1 "which side has content" input — is a **container** count
and it points the WRONG WAY on 3 of the 7. On 4 of 7 the bare side serves **no
interactive element at all** while its `tool-` twin is the working tool. Landmined.

Six of the seven now have a defensible recommendation; the seventh
(robot-hands `cycle-time-estimator`) is a genuine judgement call where both sides are
interactive. Nothing has been executed.

## 6. Traps still live (from the 08-11 handoff, plus new)

- **Verifying the deploy.** `logs -l app=agent-chassis | grep 'build provenance'`
  returns ~1.4MB of council payloads quoting the phrase, and the startup line rotates
  out within hours. Probing `/proc/1/exe` for **commit SHAs** returns absent for
  everything including a fabricated control — no positive control, proves nothing.
  **What works:** probe **added string literals**, with a **one-letter near-miss**
  negative control AND a pre-lane positive control, on **both** replicas.
  > **AMENDED 2026-08-12 ~20:05Z — the provenance line IS usable if you catch it fresh,
  > and it is strictly better than the literal probe when you can.** Two changes to the
  > recipe make it work: name the **pod**, never `-l app=` (the label selector is what
  > drags in the council payloads, and it reads one pod of N), and cap the read with
  > `--limit-bytes=400000` so you get the head of the log where startup lives:
  > `kubectl -n ai-persona-system logs <pod> --limit-bytes=400000 | grep -m2 -i 'build provenance'`
  > Done ~50 min after the roll, that returned the stamp cleanly on both replicas. Then
  > **"did my fix ship?" stops being an inference**:
  > `git merge-base --is-ancestor <your-commit> <the stamped sha>`.
  > **Its shelf life is the trap, not its accuracy** — run it in the first minutes after a
  > roll and record the sha in this table, because an empty result later means "out of
  > range", never "unstamped". The literal probe stays the fallback: it has no shelf life.
- **The identity marker's route** travels through `collected_data.site_plan`, NOT
  `site_plan_pages`. `TestReconcile_MarkerSurvivesTheStepBoundary` fails loudly if a
  future change makes `extractPagesFromPlan` field-selective.
- **Re-adoption dropping the three flags is FIXED and LIVE** as of `v1.0.1290`
  (`carryForwardStructureSpecKeys`, `19acfc895` — verified present on both replicas
  with a near-miss control 2026-08-12). The LANDMINES entry said "INERT UNTIL THE NEXT
  ROLL"; that is corrected. Still re-check the spec after any adoption run, with
  `data ? 'key'`, never `->>'key' = 'true'`.
- **Three `pages` upsert helpers with opposite policies.** Only `SyncPagesToDBAction`
  → `upsertPage` writes `pages`; `WriteSitePlanAction` writes the PLAN table.

## 7. Loose ends

- ~~**090 diagnosis `38099787-…`** — re-checked 2026-08-12 and **still verdict-less**:
  zero `doc_notes` mention the correlation. Nobody has read a root cause.~~
  **RESOLVED 2026-08-12 evening, and the loose end was never real.** The run completed on
  08-11 with **five `diagnosis_artifacts` bundles** and a correct root cause. My
  "verdict-less" reading came from querying `doc_notes`, **where no diagnosis run has ever
  written anything** — that check returns `0` for a successful run too — and from searching
  the wrong one of the item's two correlations. Landmine + `WRONG_CALLS.md` entry filed.
  **Read `bugs_open/266`**, which carries the verified root cause; the answer to *"should
  the build/deploy path refuse an archived page?"* is **yes, and not at the seam you would
  copy from the neighbouring guard** — `owned_page_guard` sits at `assemble_page`, which
  `page-rerender` and `section-editor` bypass, and those two produced the most recent
  re-deploys. **This no longer gates remediation step 5; `266` does**, and it is a live
  recurring defect (latest re-deploy 08-12 14:25Z), not the historical residue this file
  assumed.
- **Two `PLAN_PAGE_MERGE_LOSSY` rows** from the 08-11 census replan tripped the
  standing richer-wins revisit trigger — same underlying condition as O2.
- **`bugs_open/204`'s own census figure is stale** (says 5 sites, measures 6). Not this
  lane's bug; noted in the 215 file because the exclusion is consumed here.

---

# 8. Session log — 2026-08-12 evening. READ THIS FIRST if you are picking up cold

§§1–7 above have been edited in place and are current. This section says what changed
tonight, what is now true, and what to do next. **Nothing is half-finished in the tree; six
commits, all docs, no code.**

## 8.1 What is now settled (do not re-do)

| | |
|---|---|
| chassis | **`v1.0.1293`**, rolled 19:13–19:14Z. Lane code re-verified on **both** replicas two ways — see §1. Provenance sha `7a1887e316…`; `19acfc895` confirmed in the build by `merge-base` |
| counters | **0/0/0/0**, read twice (pre- and post-roll), both times with a demand control **and** an instrument control. **0 `site_plans` since the roll.** The zero is want of demand |
| §7's 090 loose end | **RESOLVED — it was never real.** My `doc_notes` check was blind; the run had answered on 08-11. See §7 |
| the archived-page defect | **filed as `bugs_open/266`**, root cause verified first-hand, population measured, two consumer lanes told |
| O2 (seven pairs) | **UNCHANGED — still the owner's, still the only open decision.** `DECISION_INPUT_2026-08-12_seven_twin_pairs.md` is ready and waiting |

## 8.2 `bugs_open/266` — the substantive find, in one paragraph

An `archived` page is rebuilt and re-stamped `deployed` by **at least four independent
producers**, none of which reads `pages.status`. The sharpest evidence: for
`tool-llm-cost-calculator`, `reconcile_site_plan` **correctly withheld** the build
(`owned_page_review`/`needs_human_review`, still uncompleted) and `image-build-handler`
rebuilt and deployed it **16 minutes later** by an unrelated path. **Live population: 5 pages
across 3 domains** (fundamentallyai ×2, leopardess ×1, robot-hands ×2), verified by curl with
per-domain 404 controls. Leopardess' has been archived-and-serving since **2026-07-17**.

**Two traps recorded in that file, both of which I nearly walked into:**
1. **Do not copy `owned_page_guard`'s placement.** It sits at `assemble_page` *deliberately
   not* at `git_commit`, because git_commit is how owned pages legitimately deploy — via
   `page-rerender` and `section-editor`, which are exactly two of the four producers here.
   Copying it closes 2 of 4 doors. `owned` ≠ `archived`.
2. **`status='archived' AND deployed_at IS NOT NULL` is a BLIND detector** — 18 rows, only 5
   serving. `deployed_at` is history, not liveness (098's retracted pages keep their stamps).
   Two-step only: SQL selects candidates, curl decides. And **a curl `000` is not a `404`** —
   one page gave 000 then 200 on three retries.

## 8.3 Why this matters to O2, which is the whole point

The remediation plan for the seven twin pairs is "keep one side, archive the other".
**`266` says an archive does not hold.** So the ordering changes: **fix `266` before executing
O2**, or the pairs will be remediated and quietly un-remediated. This is a sequencing change
only — it does not touch the survivor decisions, which are still what is needed from the owner.

## 8.4 DO THIS NEXT, in this order

1. **Nothing is owed to the counters.** They need a replan, and §5 says the fundamentallyai
   sweep front owns that site's execution — do not force one to make a signal appear.
2. ~~**`266` needs a fix**~~ **DONE — built, tested and committed `580af7ff0` (19:46Z), register
   `PBP-042` in the same commit, council submitted `2da9d905-25d8-4916-9b76-bc096679c6ab`
   (`Council-Submitted:` trailer, verdict pending).** Two refusals: `GitCommitAction` (stops
   the page serving) and `UpdatePageStatusAction` (stops the row claiming a deploy).
   **Three things are still owed on it, and they are the next session's:**
   - **Read the council verdict and act on a REVISE/REJECTED** — the code is already on the
     shared branch, so a bad verdict is a live problem, not a queued one.
   - **After the next roll, verify:** pod-probe the literal `ARCHIVED_PAGE_GUARD` on both
     replicas, then re-run the two-step population check (SQL candidates → curl verdict →
     fabricated control), then confirm no NEW `deployed_at` on any archived page. A zero in
     `agent_error_log` for `ARCHIVED_PAGE_%` needs a demand control exactly as §4 does —
     nothing has dispatched at an archived page means nothing fired.
   - **The 5 live pages are UNTOUCHED and the guard does not undo them.** It stops recurrence.
     Note it also now blocks repair-by-rebuild for those five, so **retraction is the route**.
3. **The 090 diagnosis artefacts expire 2026-09-10, unpinned.** Its findings are copied into
   `266`, so this is not urgent; pin them only if you want the raw bundles.
4. **O2 remains blocked on the owner.** Do not execute any pair.

## 8.5 Method notes worth carrying (they cost real time tonight)

- **Prove your instrument before believing a zero.** Both counter reads carry an instrument
  control now. This is what caught nothing — but the *absence* of one is what let §7's false
  claim stand for a day.
- **Ask where a SUCCESSFUL run puts its output before believing an unsuccessful-looking
  answer.** One control query against a known-complete run is what broke the `doc_notes`
  mistake. Landmine filed; incident in `WRONG_CALLS.md`.
- **A `needs_page` work item has no `page_id`** (the page may not exist yet) — filtering by
  `page_id` returns a confident, empty, wrong answer. Filter `spec->>'page_name'`.
- **The provenance line is usable if you catch it fresh** — per-pod, `--limit-bytes`, then
  `git merge-base --is-ancestor`. §6 has the amended recipe. Record the sha while it is in
  range, because later an empty result means "out of range", not "unstamped".

---

# 9. 2026-08-13 — `266` is APPROVED and LIVE. What is left is O2, which is the owner's.

- **`266` council APPROVED** round 1 (corr `2da9d905…`, 4 advisory, none high) and the fix
  is **LIVE on `v1.0.1295`**, artefact-verified on both replicas with controls.
  **Behaviourally unexercised** — `ARCHIVED_PAGE_%` counters are `0`, and that is want of
  demand, proven: **zero work items have targeted an archived page since the roll.**
  `266` stays OPEN until a build is dispatched at an archived page and the guard refuses it.
- **The objections were worth the round** — full answers in `bugs_open/266`. Two things
  they found: my sole-writer claim came from a literal grep (it survived for `deployed_at`,
  and would NOT have for `build_status`, which `UpsertPageForRole` writes via
  `Col("build_status", …)` — invisible to that grep); and a multi-page `git_commit`
  bypasses the guard, though the only such path is unexercised and names an unregistered
  action. `deployed_at` is now in `reservedPageColumns`, so the property is enforced.
- **Two named residuals stand** (neither closed, both in `266`): the multi-page commit
  path, and `deployer-agent`'s `index.html` commit which carries no page identity.
- **DARK-LAUNCH COUNTERS (§4) unchanged: still 0/0/0/0.** The demand control is unchanged
  too — no replan has run. §4's reading rules stand as written.
- **O2 is now the ONLY thing blocking this lane, and it is a decision, not work.** The
  sequencing argument that put `266` first is discharged: an archive now holds, so
  remediating a twin pair by archiving one side will stick once a build is dispatched.
  **`DECISION_INPUT_2026-08-12_seven_twin_pairs.md` is ready and unchanged.**
- **The 5 archived-and-serving pages are still serving.** The guard stops recurrence; it
  does not undo them, and it now also blocks repair-by-rebuild, so **retraction
  (`page-retraction`, which dispatches `delete_file` and is unaffected by the guard) is the
  route.** Two of the three domains belong to other lanes, both told in their own handoffs.

---

# 10. EXECUTION PLAN — O2 is DECIDED (owner, 2026-08-13). This is now work, not a question.

Decisions recorded verbatim in `DECISION_INPUT_2026-08-12_seven_twin_pairs.md` §"OWNER
RULING 2026-08-13". Procedure is the runbook's 8 steps per pair, unchanged. **Nothing has
been executed.**

## What the owner's choices changed about the procedure

- **Pairs 3+4 (fundamentallyai guides): he chose `/guides/`, which is the side I did NOT
  recommend — and he is right on merit; my recommendation was on execution cost.** The
  consequence is exactly the one the decision doc predicted: the loser (`/blog/`) is **IN
  PLAN**, so **runbook step 3 (remove the loser from the current plan) is now MANDATORY,
  not optional.** Archive it while its plan entry stands and the refile chain re-creates it.
- **Pair 7 (robot-hands `cycle-time-estimator`): he chose MERGE, not retire.** The bare page
  carries ~1,700 more words than the `tool-` page. **Retiring it before merging destroys
  them.** This is the only pair whose execution includes writing — and per CLAUDE.md's
  2026-08-06 ruling **the FRAMEWORK writes the content, not the session**: route the merge
  through the pipeline (`section_edit` / content_direction), do not hand-author it.
- **Pair 2 (finetuning): decided (`tool-`), execution HELD on `bugs_open/204`.**

## Suggested order, and why

1. **Pair 1 — ai-agent-orch `llm-cost-calculator`. Do this FIRST as the canary.** It is the
   only pair needing **no plan edit** (neither side is in the plan), and it is the clearest
   case. It exercises all 8 runbook steps on the lowest-risk pair before any plan surgery.
2. **Pairs 3+4 — fundamentallyai.** Plan edit first (step 3), then 4–8. **Route through the
   fundamentallyai sweep front**, which owns that site's execution (§5).
3. **Pairs 5+6 — robot-hands `payload-calculator`, `matchmatrix`.** Both sides in plan, so
   step 3 first for each.
4. **Pair 7 — robot-hands `cycle-time-estimator`.** Content merge through the framework
   FIRST, verify the merged tool page serves the prose, only then steps 3–8 on the bare side.
5. **Pair 2 — finetuning.** Blocked. Do not execute until `204` is fixed.

## Three things that are true now and were not when the runbook was written

- **Step 5 (archive) is DURABLE.** `bugs_open/266` is fixed, council-approved and live on
  `v1.0.1295`. The runbook's warning 3 ("archiving is not durable on its own") is
  **discharged for the rebuild path**.
- **Step 6 (retract) is still REQUIRED and is unaffected by the new guard.** Archiving stops
  a page being *re-built*; it has never removed the already-deployed *file*. Retraction
  dispatches `delete_file`, a different path from `git_commit`, deliberately not guarded.
- **A bonus falls out of step 5: this is how `266` gets its behavioural proof.** If anything
  tries to rebuild a freshly archived loser, the guard refuses and writes
  `ARCHIVED_PAGE_DEPLOY_REFUSED`. That is the evidence `266` is waiting on, and it needs no
  contrived test —
  `SELECT * FROM agent_error_log WHERE error_code LIKE 'ARCHIVED_PAGE_%' ORDER BY occurred_at DESC;`

## The five archived-and-serving pages — NOT this lane's

Owner: **"leave them for their own lanes."** leopardess and robot-hands were told in their
own handoffs. **fundamentallyai's two still need a call from its sweep front.** Separate
population from the seven pairs — no page is in both lists; do not conflate them.

## Still unmeasured, and the owner has not asked for it

**Search-engine indexing.** Named in the runbook as an input, no data source. On age alone
the flat URLs are older on 5 of 7 pairs and likelier indexed — **an inference, not a
measurement.** Pairs 3+4 now retire the flat `/blog/` side, which is the direction where
this matters most. **A redirect (step 7) is what makes it safe, so do not skip step 7.**

---

# 11. 2026-08-14 — O2 EXECUTION STARTED AND WAS HALTED AT A FINDING. Nothing was mutated.

**Chassis `v1.0.1297`** (rolled 08-13 22:29Z): `266`'s guard re-verified present on both
replicas with controls. **Still not behaviourally exercised** — `ARCHIVED_PAGE_%` counters
are `0` and **no archived page has acquired a new `deployed_at` in the ~18h since the guard
went live**, which is consistent but is still want of demand, not proof.

## What happened: pair 1 was worked to step 5 and stopped, read-only throughout

Pair 1 (ai-agent-orchestration `llm-cost-calculator`) was the agreed canary. Reconnaissance
confirmed the decision doc at the artefact (loser 11,312 b / survivor 46,368 b / control
404). Then **step 7 turned out to be inert** — full write-up in the RUNBOOK's
`⚠ CORRECTION 2026-08-14`. **There is no redirect mechanism**: `redirects` is empty
fleet-wide with no reader and no writer in the tree. Retiring a URL 404s it.

**Execution stopped there. No work item cancelled, no page archived, no file retracted, no
plan touched.** Everything done was a SELECT or a curl.

## Three things found on the way that the next session needs

1. **`site_work_items` page names are NOT unique fleet-wide.** A step-4 query filtering on
   `spec->>'page_name' IN ('llm-cost-calculator', …)` **without a site join** returned 29
   items across **four different `page_id`s on several sites** — including `051d46eb`,
   which is *fundamentallyai's* `owned_page_review` row from `bugs_open/266`. Scoped to the
   site it is **4 items, of which exactly one belongs to the loser.** Cancelling the
   unscoped set would have hit other sites' work. **Always join `sites` in step 4.**
2. **ai-agent-orchestration.com has NO `site_plans` row at all** — not "the pages are absent
   from the plan", which is what the decision doc's *"in current plan: no"* implies. Step 3
   is inapplicable for a stronger reason than recorded, and the refile loop has nothing to
   re-create from on this site.
3. **The survivor is `rebuild_policy='owned'`** and the loser is `generic` — so PBP-036's
   guard already protects the side we are keeping.

## The decision this puts back to the owner (see the RUNBOOK correction)

Retiring any loser produces a **404, not a redirect**. That is most consequential for
**pairs 3+4**, where he chose `/guides/` and therefore retires the older, likelier-indexed
`/blog/` URLs — **and he was told a redirect would make that safe.** It does not exist.
Options per pair: accept the 404, keep the loser published and de-duplicate another way, or
build a redirect capability first. **Do not resume execution until this is answered.**

---

# 12. PAIR 1 EXECUTED TO STEP 6 AND CORRECTLY REFUSED. Read this before touching pair 1 again.

**Owner reversed pairs 3+4 to `/blog/` on the redirect finding** (recorded in the decision
doc, 2026-08-14 ruling) and said carry on with the rest. Pair 1 was then executed.

## What was done, and what state pair 1 is in NOW

| step | action | result |
|---|---|---|
| 4 | cancel the loser's one open work item | **DONE** — `087c029a…` `page_rerender` → `cancelled`, reason recorded, `handled_by=brochure_215_o2_thread`. **Site-scoped**; the unscoped query would have cancelled 29 items across four sites |
| 5 | archive the loser | **DONE** — `llm-cost-calculator` (`6e66ff49…`) `active` → `archived`. Survivor `tool-llm-cost-calculator` untouched, still `active` |
| 6 | retract the deployed file | **REFUSED BY THE PLATFORM, correctly. Nothing deleted.** corr `3e526489…`, orchestration `9d85cdc2…`, `COMPLETED` with `1 considered, 0 dispatched` |
| 7 | write a redirect | **SKIPPED — inert.** See the RUNBOOK correction |
| 8 | verify at the artefact | **not reached** |

## Why it refused, and why that is the headline

> `retraction refused for page "llm-cost-calculator": still linked from live content —
> repair or remove those links first (see editorial_inbound)`

**The page is linked from three live surfaces**, named in the refusal's `context`:

- **nav** — `site_nav_items` row `c5738bd1-cb71-4f93-9f8e-d0fa166678a5`, label *"LLM Cost Calculator"*
- **chrome footer** — `{"slot":"footer","source":"chrome"}`
- **an article body** — page `llm-provider-abstraction-production-agent-systems`, slot `article-body`

**The runbook said this page had zero inbound links.** That claim came from `link_registry`,
which is **empty fleet-wide** — so it was never evidence (LANDMINES 2026-08-14). The real
inbound census, which `RetractPageDeploymentAction` runs itself, found three references on
the first page anyone tried. **Assume the same is true of the other six pairs until each is
checked by dispatching the retraction and reading the refusal** — that is now the cheapest
inbound census available, and it is free: it refuses rather than deleting.

## ⚠ The state pair 1 is in is SAFE but it is not FINISHED

`llm-cost-calculator` is now **`status='archived'` and still serving 200** — i.e. this
lane has deliberately created a sixth instance of exactly the condition `bugs_open/266`
tracks. It is safe: `266`'s guard stops it being rebuilt, nav and footer still point at a
page that still serves, so **nothing is broken for a visitor**. But it must not be left
here indefinitely.

**Two ways forward, and the first is the decided one:**

1. **Repair the three referrers to point at `/tools/tool-llm-cost-calculator.html`, then
   re-dispatch the retraction.** The nav row is a DB edit; the footer is **stored chrome**
   (`bugs_open/117` — no page rerender regenerates it, read that first); the article body is
   page content, so per the 2026-08-06 ruling **the framework rewrites it, not a session**.
2. **Or revert:** `UPDATE pages SET status='active' WHERE id='6e66ff49-3f0e-423f-9286-5ec3dc0c413c'`
   and re-open the cancelled work item, restoring the pre-execution state exactly.

## The transferable lesson for the remaining six pairs

**Dispatch the retraction FIRST, as a read-only inbound census.** It refuses without
deleting, and its refusal names every referrer by surface and slot. Doing that before
archiving would have told us pair 1's real link state in ~40 seconds, with no DB change at
all. **The runbook's step order (archive, then retract) is wrong in this respect** — retract
first to learn, then archive, then retract again to execute.

---

# 13. 2026-08-14 (afternoon) — pair 1's blocker is CONTENT WORK, by design. This resizes O2.

Chassis **`v1.0.1298`** (rolled 08:58Z): guard re-probed present on both replicas, controls
behaved. `ARCHIVED_PAGE_%` still 0 — and note the loser has now been `archived` for ~6h with
`deployed_at` unmoved, so nothing has tried to rebuild it either.

## Do NOT hand-edit the nav row. The retraction does it itself.

`retract_page_graph.go:16-38` splits the three inbound classes deliberately:

- **nav row (`site_nav_items`) → the retraction DEACTIVATES it** (to `'inactive'`, a stated
  convention) and reports it. *"A nav row is a pointer, not prose."* **Mechanised — editing it
  by hand is doing the machine's job, and would leave the machine's own step a no-op.**
- **editorial (body copy AND site chrome) → REFUSE, and name the referrers.** *"Repairing
  prose is a content decision… auditors raise work items and never rewrite content."* The
  refusal is what makes "dead link created by a retraction" unrepresentable rather than merely
  detected.
- **outbound orphans → report only**, owned by `check_orphan_pages`.

**So pair 1 is blocked on exactly TWO things, both editorial:**

1. **the site's footer chrome** — `{"slot":"footer","source":"chrome"}`. Site-wide: every page's
   footer links to the loser. Chrome is a **stored artefact** and no page rerender regenerates
   it — read `bugs_open/117` before touching it.
2. **one article body** — page `llm-provider-abstraction-production-agent-systems`, slot
   `article-body`.

**A useful thing found while checking: the survivor ALREADY has its own nav row**
(`f96754a7…`, position 12, `/tools/tool-llm-cost-calculator.html`). So the loser's nav row
(`c5738bd1…`, position 5) must be **deactivated, not repointed** — repointing would duplicate
an existing entry. The retraction does exactly that. (Cosmetic aside for whoever finishes:
the surviving row's label is `Llm Cost Calculator`, slug-cased; the dying row has the correct
`LLM Cost Calculator`. Worth copying the casing across — one field, and it is the label that
will remain in the nav.)

## The existing link-repair machinery does NOT apply here, and that is not a gap

`component_link_repair.go` / `repairSectionLinks` repair links to pages that are **dead or
unbuilt**. These links point at a page that exists and serves 200. What is wanted is to
**repoint an editorial link to a different destination**, which is a content decision — the
same reason the retraction refuses. The repair machinery would only engage *after* a
retraction created a dead link, which is precisely the state the refusal exists to prevent.

## ⚠ This resizes O2, and the owner should know before pairs 3–7

**Pair 1 was chosen as the canary because it was the cleanest of the seven** — no plan surgery,
a 727-word stub as the loser. It still cannot be retracted without a **chrome edit and a
content edit**. There is no reason to expect the other six to be cleaner, and robot-hands'
three additionally need plan surgery. **The runbook's "8 mechanical steps per pair" is not
what this is.** Realistic shape per pair: dispatch retraction as a census → repair N editorial
referrers through the framework → re-dispatch → verify.

**Cheapest next measurement, and it costs nothing:** dispatch the retraction at each remaining
loser *before* archiving. It refuses, deletes nothing, and returns the full referrer list per
page. Six dispatches would size the whole remaining job in minutes.

## State of pair 1 right now

`llm-cost-calculator` is `archived` and still serving 200; nav and footer still point at it, so
**nothing is broken for a visitor**. Cancelled work item `087c029a…`. Survivor untouched.
Exact revert if wanted:
`UPDATE pages SET status='active' WHERE id='6e66ff49-3f0e-423f-9286-5ec3dc0c413c';`
