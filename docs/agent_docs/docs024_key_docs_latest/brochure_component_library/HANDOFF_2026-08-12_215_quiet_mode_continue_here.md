# HANDOFF — `bugs_open/215` quiet mode, continue here (2026-08-12 afternoon)

Supersedes `HANDOFF_2026-08-11_215_quiet_mode_continue_here.md`, whose §4 ("DO THIS
FIRST — read the dark-launch population") is **done** and whose O1 is **decided and
executed**. Read that file only for history; its §6 traps and §7 loose ends are
carried forward below where still live.

**The pilot is live and inert. Nothing is half-finished in the tree.**

## 1. State

| | |
|---|---|
| code | **LIVE on chassis `v1.0.1291`**, artefact-verified on BOTH replicas 2026-08-12 ~16:00Z (re-probed after the roll from 1290; positive control + one-letter near-miss both behaved). **⚠ STALE AS OF 2026-08-12 ~20:00Z — the owner has a fresh chassis building and about to deploy.** That verification described 1291 and says nothing about what replaces it. **Re-probe before relying on this row** (§6 has the only method that works), and read the stamp of `agent-chassis` specifically — a release can straddle other sessions' commits and ship several revisions under one tag, so a fleet-level tag is not an answer about this service. All four lane commits are ancestors of HEAD, so a build from HEAD carries them; that is an expectation, not a measurement |
| council | **APPROVED** round 3, corr `56e13695-17cb-48ec-bc6b-0371fde8b717` |
| enabled on | **fundamentallyai.com only** — `honour_realised_identity` + `twin_identity_snap`. `stem_twin_snap` absent by design |
| dark-launch counters | **still 0/0/0/0** — re-read 2026-08-12 ~19:00Z with BOTH controls: demand control unchanged (only post-roll plan is noted.co.uk's first build, 0 `pages` predating it), and an instrument control proves the query is not blind (`agent_error_log` took 3,503 rows in 24h). No replan has run through the new path yet |
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

**A roll is landing (owner, 2026-08-12 ~20:00Z) — two consequences for this query.** The
counters live in `agent_error_log`, which the roll does not clear, so a non-zero after the
roll may predate it: **bound the read with `occurred_at > '<roll time>'`** if you are asking
what the NEW binary did. And CLAUDE.md's standing rule applies before you try to induce a
signal: **no orchestration dispatch within ~300s of a chassis pod (re)start** — the spawn is
silently dropped, which reads exactly like "the gate did not fire".

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
