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
