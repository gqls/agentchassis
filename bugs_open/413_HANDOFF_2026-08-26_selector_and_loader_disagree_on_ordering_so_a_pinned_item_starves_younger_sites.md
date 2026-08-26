# 413 — the dispatch selector and the item loader disagree on ordering, so one pinned item freezes its site's age and starves every younger site of trigger dispatch

**Filed 2026-08-26 by the dispatch_throughput lane.** Diagnosis loop run in flight:
`RUN_CORRELATION_ID=250188a7-29ae-4b3d-ace6-638694612c8b` (090, filed ~15:2xZ before this file
was written, per the 2026-07-31 owner ruling). First-hand verification below was done at the live
artefact this afternoon; the loop run is the independent check, not the substitute.

**Symptom that surfaced it** (handed over by the `bugs_open/391` lane, explicitly undiagnosed, in
`docs/agent_docs/docs024_key_docs_latest/dispatch_throughput/CONTRIB_2026-08-26_from_bugfix_391_priority_is_inert_between_sites.md`):
`finetuning.uk`, an unlocked site with **73** (as of 2026-08-26 ~15:0xZ) eligible items — every one
passing every clause of the live selector, verified clause-by-clause — received **zero** dispatch
for 10+ hours (last claim 05:09:30, exactly one `build-dispatch-loop` in 12 h), while the fleet ran
265–278 claims/h and sites with strictly younger oldest-items were served.

## Mechanism (two orderings, one contract each, incompatible jointly)

1. **Site selection** — `find_dispatchable_site`, a `query_database` step on the LIVE
   `agent_definitions` row for `build-pipeline-trigger` (read from the artefact 2026-08-26, not a
   mirror): picks the site owning the **globally oldest** eligible `site_work_items` row —
   `ORDER BY wi.created_at ASC, wi.priority ASC, wi.id ASC LIMIT 1`, with a skip-if-site-has-a-
   claimed-item clause. A site's standing in the queue IS the age of its single oldest eligible row.
2. **Item loading on the picked site** — `load_work_items`
   (`platform/orchestration/actions/load_work_item_actions.go`): `ORDER BY wi.priority ASC,
   wi.created_at ASC` and takes `max_items` (5). Priority is the within-site order; age is the
   tiebreak.

The selector ranks sites by AGE; the loader serves items by PRIORITY. An old row at a
numerically-high (= served-last) priority therefore does two things at once: it keeps **winning
site selection** for its site, and it **never gets loaded** while better-priority rows keep
arriving on that site. The site's oldest-eligible age freezes at the pinned row's `created_at`,
and every site whose oldest row is younger gets **nothing from the trigger** until the pin clears.

Two pin flavours, both measured live `[MEASURED 2026-08-26 ~15:1xZ]`:

- **Never-loaded pin:** `mortgagecalculator.co.uk` — 3× `audit_tool` at priority **140**, created
  **2026-08-25 23:53**, `attempt_count=0`, still `triaged` — through a 3-hour window in which the
  site received **22 loops and ~95 claims** at priorities 30–130. Same shape on `oufe.com` and
  `ai-agent-orchestration.com` (`audit_tool` @140, attempt 0, created 01:42/00:55).
- **Fail-bounce pin:** `loancash.co.uk` — 2× `required_fields_missing` @140, created **00:09**,
  `attempt_count=2`: loaded, failed, released back to `triaged`, still the site's oldest row.
  (This flavour self-clears at `max_attempts`; the never-loaded flavour clears only when the
  site's better-priority inflow pauses long enough to drain 5-at-a-time down to it.)

**Cross-check that the model, not just the snapshot, is right:** sites are served exactly while
they hold old rows and drop off the moment those drain — `fundamentallyai.com` took 35
trigger-spawned loops up to 14:13 and then none; `idea.uk` 19 up to 13:59 and then none. (Their
"ahead" rows are invisible to a NOW-census: claimed rows leave `triaged` and completed rows
archive out of `site_work_items` — the rolling-window trap.) At ~15:1x, **13 sites / ~570
eligible rows** stood older than finetuning.uk's oldest (05:03:29); several are pinned as above.

## Why the aggregate hides it

Fleet claims/h, distinct-sites-per-hour, and lost-claim share all read HEALTHY through this —
2026-08-26 afternoon: 265–278 claims/h across 13–17 distinct sites/h — because the starved site
contributes no failures, no losses, and no attempts. **The damage is an absence** (cf. the
`bugfix-213` memory class). A per-site floor — max hours-since-last-claim across sites with
eligible work — is the meter that sees it; distinct-sites-touched cannot.

## What is NOT the cause (each measured, between the 391 lane and this one)

site lock (NULL, no exceptions) · stuck claim holding the busy-skip (0 claimed rows) ·
`retry_after` deferral, incl. the bugs_open/307 shape (73/73 claimable) · attempt exhaustion
(all attempt_count=0, max_attempts not reached) · approval_mode / depends_on clauses (73/73 pass;
all depends_on NULL) · loader-drop black hole (every looped site loads ~5 and claims ~90%+;
bugs_closed/078's scan-drop shape specifically ruled out) · the second producer of
`build-dispatch-loop` spawns (NULL `task_name`, ~3-4/h fleet-wide) serves some younger sites but
neither causes nor masks the pin.

## Relation to design intent

Priority ~140 on `audit_tool` plausibly MEANS "run last within the site" — each ordering is
defensible alone; the defect is their JOINT effect at the seam, where "last within the site"
becomes "first for the fleet's site ranking, forever". `bugs_closed/078` fixed a *dropped-row*
variant of selector-counts-what-loader-won't-take; this is the *ordering* variant of the same
contract gap. The 391 lane's CONTRIB (§3) independently identified the ordering key as where the
per-site-latency trade-off lives, before this evidence existed.

## Fix candidates (ranked by what closes the door; none applied — owner/lever decisions live in the dispatch_throughput lane)

1. **Make the two orderings agree on what "represents" a site**: the selector ranks sites by the
   age of their oldest LOADABLE-NEXT row (i.e. min(created_at) over each site's top-`max_items`
   by the loader's own ordering) — the pin becomes unrepresentable. Costlier query; needs the
   loader's ordering visible to the selector (one shared SQL fragment is exactly what the
   `work_items_common.go` ⚠⚠ comment warns against hand-DRYing — do it as a stated contract).
2. **Age floor / anti-starvation term in site selection** (e.g. rank by
   least(oldest_eligible, now()-<cap>) or round-robin among sites idle > N hours): bounds the
   damage without touching the loader; the pin remains but cannot starve indefinitely.
3. **Cap how long a row may pin** — escalate/reprioritise rows older than N hours still at
   attempt 0 (touches item semantics, not the seam).
4. Documentation only (the 391 CONTRIB's option 1) — rejected as sufficient: callers setting
   priority cannot see the seam, and the starved site's owner has no signal at all.

## How to verify a fix

The finetuning.uk shape reproduces naturally within hours; the meter is the per-site floor above
(RUNBOOK, to be added to the 24h post-B read). Disconfirming result for candidate 1/2: with the
fix live, no site with eligible work goes > ~1h unserved while pinned rows exist elsewhere —
measured at the artefact (loops per site), not at the aggregate.

## Interaction with in-flight work (dispatch_throughput lane)

Phase 3 (`max_items` 5→8) cuts BOTH ways: deeper loads reach worse-priority rows sooner (weakens
never-loaded pins) but hold each site busy longer per pick. Do not treat Phase 3 as the fix or as
neutral — measure the per-site floor across it. Ruling B itself (migration 637) neither caused nor
cures this: the pair co-picked the deep site 94% pre-B, so pre-B starvation was WORSE, just
unmeasured.

## Addendum 2026-08-26 ~15:5xZ — pinned vs victim, measured (the distinction is the 391 lane's, from their own site's shape)

The 391 lane observed that finetuning.uk is NOT itself pinned (its oldest row is @60 and loads
fine the moment the site wins) and proposed the discriminator: **pinned** = the site's oldest
eligible row falls OUTSIDE the loader's top-`max_items` by (priority, created_at); **victim** =
the oldest would load, the site simply never wins. Census run `[MEASURED 2026-08-26 ~15:5xZ]`,
query in RUNBOOK (windowed rank comparison, all selector clauses applied):

- **25 sites hold eligible work; 13 are pinned, 12 are not.** Severe pins: loanzy.uk
  (oldest loads 60th), ai-agent-orchestration.com (51st), webdesign.co.uk (34th), loancash
  (28th), lendzy (28th), loancalculator (27th). finetuning.uk: oldest_load_rank **2** — pure
  victim, canary at rank 1 (their reading of their own queue, confirmed here).
- **Pin status is DYNAMIC — the file's own first example has already cleared.**
  mortgagecalculator.co.uk, pinned with 3× @140 behind a deep better-priority queue at ~15:1x,
  had drained to 5 eligible by ~15:5x: oldest_load_rank 3, unpinned, last claim 15:16. A pin
  clears when the site's better-priority inflow pauses long enough to drain within reach.
  Any census of this population is a snapshot; date it.
- **Sharpened mechanism: a site's own pin does not starve ITSELF — starvation is positional.**
  loanandmortgagecalculator.co.uk is pinned (rank 8) AND starving since 04:39 — because of the
  pins AHEAD of it in age order, not its own. A pinned site starves the same way a victim does
  while older pins exist; its own pin then makes it the next persistent age-blocker once it
  starts winning. Fix consequences: unpinning (candidate 1) frees victims for free, as the 391
  lane noted — but candidate 2 (age floor) is the only one that bounds the POSITIONAL wait,
  which is the harm both groups actually suffer.
- Currently starving > 10 h: gaswholesalers.com (victim, last claim 04:22),
  finetuning.uk (victim, 05:09); loanandmortgagecalculator (pinned, 04:39).

090 status at this addendum: run `250188a7` iterating (evidence bundles 15:16 / 15:20 in
`diagnosis_artifacts`), verdict not yet landed — will be appended here when read.

## Addendum 2026-08-26 ~17:1xZ — the 090 verdict: UNVERIFIABLE at iteration cap; mechanism unamended; one sentence of the symptom corrected

**Verdict verbatim** (item `result`, completed 15:40:27Z): *"Diagnosis NOT confirmed (stopped:
iteration-cap)"*, status **UNVERIFIABLE**, no fix proposed, best-effort trail attached. This is
neither CONFIRMED nor REFUTED. Per the 2026-07-31 owner ruling, this file therefore stands on
**declared first-hand verification**: the live selector text read from `agent_definitions` (not a
mirror), clause-by-clause eligibility 73/73/73 on the victim site, raw pinned rows observed
directly (the @140 trio at `attempt_count=0` through 22 loops / ~95 claims on the same site), the
per-site loop-cessation cross-check against history, and the repeatable censuses in the lane
RUNBOOK. The loop's five evidence bundles are on corr `250188a7` in `diagnosis_artifacts`.

**Reconciling the loop's two counter-points, from its own trail:**

1. *"A query built specifically to surface [the pin signature] returned zero rows."* The bundle
   records request DESCRIPTIONS and results but not SQL, so the query cannot be audited — but the
   trail's own "oldest rows" sample consists of **`status='detected'` rows** (dated back to
   2026-07-26, several on finetuning.uk itself), which are **ineligible by the selector's own
   filter** (`status IN ('triaged','approved')`). A pin test keyed to that population asks about
   rows the selector never sees. A zero from an unauditable query with no positive control is not
   a refutation ([[a-post-fix-zero-needs-a-demand-control]]); the 15:5x census in the addendum
   above, with the full clause set applied, found 13 pinned sites in the same hour.
2. *"Well over a dozen distinct site_ids cycling ... inconsistent with a single stale row
   monopolizing trigger-driven dispatch."* Correct observation, wrong target: 413 does not claim
   a single row monopolizes dispatch. It claims service cycles healthily among the sites at the
   OLD end of the age order while sites behind the pins wait on fall-through. The cycling the loop
   saw is what the mechanism predicts.

**> CORRECTED 2026-08-26: the symptom overclaimed, and the overclaim is what the loop graded.**
The filed symptom said younger sites "receive no trigger-driven dispatch **at all**" while a pin
holds. Too strong: they are served by FALL-THROUGH whenever every older-item site is
simultaneously busy — fundamentallyai.com took 35 trigger loops the same afternoon. The precise
claim, which all evidence supports: **a site behind pins is served only by fall-through, with no
bound on the wait** — measured 1 loop/12 h (finetuning.uk), >11 h (gaswholesalers.com). Caught by
the 090 round; logged in `WRONG_CALLS.md` 2026-08-26 ("a graded verdict grades the SENTENCE you
filed"). A re-file of the 090 with the corrected sentence is deliberately NOT queued: the
remaining open question is a fix-candidate choice, not a mechanism dispute, and a second run
would spend credits re-deriving the census this file already carries.

## Fix BUILT 2026-08-26 evening (bugs_open/413 session; NOT YET APPLIED — apply ≥12:00Z 08-27 by agreement)

**Candidate 1, generalised to the framework contract: a selector may only represent a
container by work its drainer will actually take — same filters, same ordering, same window.**
Chosen over candidate 2 because it closes the door (the pin becomes UNREPRESENTABLE) rather
than bounding the damage; candidate 2 (positional-wait bound) is a policy trade and goes to
the owner separately (README_where_we_are 2026-08-26 evening) with the per-site floor as its
evidence base. Candidate 3 untouched: priority 140 still means "run last within the site" —
that is its meaning, not damage.

**Shipped in this commit** (all inert until the _HOLD is hand-applied):

- `docs/agent_docs/sql_for_agents/657_selector_ranks_sites_by_loadable_work_HOLD.sql` —
  selector ranks sites by min(created_at) over each site's top-K eligible rows under the
  LOADER's ordering (`priority ASC, created_at ASC`, `load_work_item_actions.go:789`); K read
  LIVE from `build-dispatch-loop > load_items > max_items` (written K-agreement with 658, so
  Phase 3's 5→8 is absorbed automatically; COALESCE→1 degrades pin-free). Preflight refuses
  drift by whole-text md5; guards assert every eligibility clause individually + an EXECUTE
  probe. `_ROLLBACK` (md5-exact restore) + `_VERIFY` (md5 / ordering-mirror / K-resolution /
  probe; **fails by design pre-apply** — proven, exit 3).
- `platform/orchestration/actions/load_work_items_ordering_contract_test.go` — AST test
  pinning the loader's ORDER BY literal; mutation-proven (ordering swapped → FAILS with a
  message routing the editor to the selector window + VERIFY assertion 2; reverted; passes).
- Register WDS-002: "ordering contract CLOSED" bullet + visible correction of its own
  "bounded by ceil(backlog/5)" claim; same dated correction block added to 284's header.
- `LANDMINES.md` "ONE ordering contract" entry (verifier dispatched, corr 8aed215e).
- `bugs_open/415` — adjacent finding, scoped OUT: the fire-gate `pre_query` is NARROWER than
  the selector (`triaged`-only + `pipeline='build'`), so an approved-only backlog never fires
  the trigger; theoretical at today's volume, first-hand verified.

**Proofs at the artefact (2026-08-26 ~20:4xZ):** full dry run in a rolled-back transaction —
preflight, update, guards, VERIFY-passes-on-applied-state (K=5; census 28 eligible/15 pinned
at 20:46Z), rollback-restores-md5-exact, exit 0, nothing persisted. **Divergence, same
instant:** OLD text picks webdesign.co.uk (pinned, oldest row loads 135th at the 20:1x census
— 16/25 sites pinned, evening inflow deepening pins vs 13/25 at 15:5x), NEW picks
vetcomparison.uk (oldest genuinely-loadable work). Cost same class by EXPLAIN ANALYZE
(~120 vs ~115ms real work, JIT-dominated either way, one fire/60s).

**Council:** submitted 2026-08-26 ~21:0xZ, `SUBMISSION_CORR ecf2e542-7ba3-4574-92ed-35025aed5b27`
(edits: the _HOLD migration + the Go test; sidecars/prose named in the rationale). Commit
carries `Council-Submitted:`; verdict to be read and dispositioned here before apply.

**Apply plan (agreed in writing with the dispatch_throughput session ~20:4xZ):** NOT BEFORE
12:00Z 2026-08-27, after their 24h post-B read (~09:00Z) and 658/Phase 3 (~09:30Z), so each
lever gets an attributable window; stamp the apply time HERE and ping that session; they hand
the 09:00Z pre-fix floor baseline. Commands + acceptance meter: dispatch_throughput RUNBOOK
§657. **Disconfirming result for the fix** (unchanged from §"How to verify"): any site with
eligible work > ~1h unserved while pinned rows exist elsewhere, measured at the floor, worst
site, dated. Close this bug only when fixed AND live AND measured.

## Council round 1: REVISE — and the gating objection was right (2026-08-26 ~21:0xZ)

Verdict on corr `ecf2e542` landed ~12 min after submission: **REVISE, gating objection from
debug_historian (HIGH)** — the K subquery selected the build-dispatch-loop row by
`ORDER BY updated_at DESC`, but the runtime loader (`loadAgentDefinition`,
`platform/messaging/processor.go:371-389`; `loadAgentDefinitionForAction`,
`ai_actions.go:1400-1412`) selects `ORDER BY version DESC LIMIT 1`, and `updated_at` is
documented DEGENERATE — so under a duplicate-active-row shape (four types carry one) K could
be read from a row the runtime never loads. Moot today (their own check: 1 active row,
version {1}, both types) and fixed anyway: **the K subquery now mirrors the loader's
selection rule verbatim**; new stored-text md5 `af908ea3758814994d0f54b8506e9a70` (VERIFY
updated in lockstep). Their second objection also acted on: preflight now asserts exactly ONE
active row per type and pins the UPDATE by captured row id, never re-resolved by type.
Remaining seats' items answered with evidence in the r2 rationale (operation re-typed
`config_change`; 658's text verified non-competing — 0 ordering hits, jsonb_set-integers
only; test-absence claim re-stated as a run check; provenance shown: the landmine doc_notes
rows synced at 20:49 pre-round ARE the subject's travelling record). All r2 proofs re-run at
21:10Z: dry-run transaction green end-to-end, pre-apply VERIFY still fails by design.
**Round 2 submitted on the SAME correlation** (`RESUBMIT_CORR`, trail accumulates); the
throughput lane's tomorrow-session gates its all-clear on this verdict being read and acted
on. A REVISE round is cheaper than the defect it finds — this one found a real one.

## Council round 2: APPROVED (2026-08-26 ~21:2xZ) — one advisory, acted on; READY FOR APPLY

**Verdict read in full:** APPROVED, "1 advisory objection(s) — none high-severity", 4
abstained. (The report header says "(round 1)" — the known template literal; the reports
were counted.) Dispositions:

- **guardian MEDIUM (acted on):** the bare `::int` cast on max_items would THROW on a
  malformed value — halting dispatch FLEET-WIDE, a larger blast radius than the bug fixed;
  658's write-side guard binds only 658's own writer. The K expression is now TOTAL: regex
  guard (`~ '^[0-9]+$'`) → COALESCE → GREATEST(...,1). Proven with runtime values at the DB:
  `'all'/'0'/'-3'/NULL → 1`, `'8' → 8` — every failure direction lands on K>=1, pin-free,
  dispatch never stops. (Probe-construction note for the next reader: a LITERAL `'all'::int`
  is constant-folded before CASE can guard it and throws — that is a probe artifact, not the
  query's behaviour; test with column refs.) The deliberate divergence from the loader's own
  fallback (GetIntField → 50) keeps K <= M, the safe direction. Final stored-text md5
  **`d29807313a8f6ed543a541c35c1626c4`** (VERIFY in lockstep).
- **guardian LOW (recorded):** the cross-agent config read is now load-bearing for
  correctness — already tracked in WDS-002 as instance 2 of the pattern; a third warrants a
  shared helper.
- **debug_historian LOW (confirmed first-hand, as asked):** both `scheduled_tasks` rows
  carry `target_agent_type='build-pipeline-trigger'` `[MEASURED 2026-08-26 ~21:2xZ]` — they
  fire the SAME agent type resolving to the SAME single active `agent_definitions` row, so
  the 657 edit is automatically shared; the sibling's only per-row config is
  pre_query/interval (bugs_open/415's territory, unaffected).

Final proofs (21:24Z): dry-run transaction — preflight + id-pinned update + guards +
**VERIFY green on applied state (K=5; census 30 eligible / 14 pinned)**; live row proven
untouched afterwards (standalone VERIFY fails on md5 arm against the unchanged
d6f98acd... text, exit 3). **Everything is in place for the hand-apply ≥12:00Z 2026-08-27**
(RUNBOOK §657; stamp here + ping the throughput lane; floor at +2h/+6h vs their 09:00Z
baseline).
