# HANDOFF — `bugs_open/215` quiet mode, continue here (2026-08-12 afternoon)

Supersedes `HANDOFF_2026-08-11_215_quiet_mode_continue_here.md`, whose §4 ("DO THIS
FIRST — read the dark-launch population") is **done** and whose O1 is **decided and
executed**. Read that file only for history; its §6 traps and §7 loose ends are
carried forward below where still live.

**The pilot is live and inert. Nothing is half-finished in the tree.**

## 1. State

| | |
|---|---|
| code | **LIVE on chassis `v1.0.1293`** (rolled 2026-08-12 19:13–19:14Z), **re-verified on BOTH replicas 2026-08-12 ~20:05Z by two independent methods.** (1) Literal probe of `/proc/1/exe`: `PLAN_PAGE_STEM_TWIN_REFUSED` **present**, one-letter near-miss `…REFUSEE` **absent** (so the probe can fail), pre-lane control `OWNED_PAGE_GUARD` **present** (so it works on this binary) — all three on each replica. (2) Provenance stamp, captured while still in log range: both replicas built from `7a1887e3163af75ce…`, and `git merge-base --is-ancestor` confirms `19acfc895` (the `carryForwardStructureSpecKeys` re-adoption fix) **is in the build**. Supersedes the `v1.0.1291` verification of ~16:00Z |
| council | **APPROVED** round 3, corr `56e13695-17cb-48ec-bc6b-0371fde8b717` |
| enabled on | **fundamentallyai.com only** — `honour_realised_identity` + `twin_identity_snap`. `stem_twin_snap` absent by design |
| dark-launch counters | **still 0/0/0/0** — re-read twice on 2026-08-12, at ~19:00Z and again at ~20:05Z **after** the `v1.0.1293` roll, both times with BOTH controls. Demand: **0 `site_plans` rows since the roll**, and the only post-1291 plan was noted.co.uk's first build (0 `pages` predating it, so it cannot exercise the reconciler). Instrument: `agent_error_log` took 3,503 rows in 24h and **13 since the roll**, so the query is not blind on the new binary either. **No replan has run through the new path yet — the zero is want of demand, not want of function** |
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
silently dropped, which reads exactly like "the gate did not fire". (That window is long past
as of ~20:05Z; it will matter again on the next roll.)

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
2. **`266` needs a fix, and it is a shared seam** → council gate before/alongside the commit,
   and a concept-register entry in the same commit (the 2026-07-28 ruling's condition 2).
   Fix candidate 1 (refuse at the commit seam) is the one that makes the bad state
   unrepresentable; read §8.2 trap 1 before writing a line.
   **`266` is NOT owned by anyone yet** — `who-owns.py` will say nothing useful, as it reads
   commits and the file is hours old.
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
