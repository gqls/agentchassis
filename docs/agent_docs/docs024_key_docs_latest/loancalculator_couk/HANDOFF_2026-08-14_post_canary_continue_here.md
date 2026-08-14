# HANDOFF — loancalculator.co.uk · post-canary continue point (2026-08-14 night)

> Supersedes `HANDOFF_2026-08-13_planner_half_continue_here.md`. The planner half is
> DONE, council-APPROVED, and canary-run. The rebuild now waits on TWO OWNER
> ANSWERS, not on code.

```
site      loancalculator.co.uk   0162cde4-633e-45e9-8ca6-87a6b2fe1d26
serving   27/27 clean (verified tonight post-canary; calculators + cut claims + locked
          footer all intact through today's fleet rerender wave)
planner   407 LIVE + APPROVED (corr 508fe8eb, round 2; trail: REVISE -> APPROVED).
          Flag seeded on this site only. PLAN-049. Canary corr b23b19c7: menu 140,
          param resolved, locks untouched, pages byte-identical bar two contained
          planner-inventions (archived).
state     REBUILD READY TO DISPATCH once the two owner questions below are answered.
```

## What changed since the 08-13 handoff (all committed)

| thing | proof |
|---|---|
| 407 applied: load_components widened, params `site_record.site_id` | md5 control pre+post on STORED text: 129 unflagged / 140 flagged / 129 ghost; snapshot in `agent_definitions_backup` (NB: the two-arg `snapshot_agent` writes THERE, not agent_definitions) |
| `plan_includes_tools` seeded (structure aspect, fifth opt-in key) | seed guard passed (both keys present, 27-entry pages list intact); carry-forward on re-adoption covered by `19acfc895` (probed at both replicas tonight) |
| Council APPROVED, round 2 | `Council-Reviewed: 508fe8eb…` legitimate from `a212b3470` on; round-1 objections all dispositioned in NOTES 08-14 (late) |
| Canary planner run | corr `b23b19c7`, orch `7d9d9b6d` (row purges in ~2 days — evidence copied into NOTES) |
| Register | PLAN-049 (entry + index row); PLAN-048 ⚠ corrected, five-key census lives in PLAN-048 |
| Follow-up | `features_open/033` (structure-aspect key counter, review-at-8 budget) |
| 08-13 puzzle resolved | orchestration_states is a ~2-day working set; planner ran 3× ever; nothing exotic |

## The canary's two findings — THE OWNER QUESTIONS

1. **A bare replan does not propose compositions for built pages** (every realised
   page carried `sections: []` into the plan — verified in the run's collected_data).
   So the rebuild will NOT regenerate the 26 pages from the plan alone.
   **Q1: how should the 26 existing pages be regenerated?** The honest route the
   canary points at: explicit `recompose_pages` — and per the recompose landmine the
   intent must be in BOTH `spec.recompose_pages` AND the briefing prose, then judged
   via `agent_error_log` `RECOMPOSE_INTENT_NOT_REALISED`. The calculators' placement
   test happens THERE (this is where the 407 menu + the identity arm finally meet a
   real composition), with the 12 locks as the floor: worst case is repositioning,
   caught by the acceptance check, never deletion.
2. **A bare replan INVENTS pages** — `about` + `guides-index` appeared despite the
   run converging on the 27 realised pages. Contained tonight: 7 follow-on items
   deferred (incl. a needs_rerender that would have deployed the two not_built rows
   as EMPTY SHELLS), both rows archived (guarded, reversible — zero components).
   **Q2: is the mission brief's keep-pages pin trusted to suppress invention?** The
   canary ran bare so this is untested. Whatever the answer: the fire runbook gains a
   step — *immediately after the planner phase, check for new `active` rows*
   (`SELECT name FROM pages WHERE site_id='0162cde4…' AND created_at > <fire time>`).

## THEN the fire sequence (updated from 08-13 step list)

1. Owner answers Q1/Q2 above.
2. Release the 8 non-calculator locks (chrome 3 + css carriers 4 + homepage prose-0);
   the 12 calculator locks STAY. Pre-release state in NOTES 08-11.
3. Extract the MISSION block to plain .txt; fire `082_submit_domain_unified.sh
   loancalculator.co.uk --email uk@websy.uk --mission-file <txt>` — WITH the Q1
   recompose decision reflected in the dispatch.
4. Monitor via `parent_orchestration_id`; publish→start can be ~30 min; no dispatch
   within ~300s of a chassis restart. Query by correlation, never by now()-interval.
5. Verify per original step 8 (purity vs the 08-11 backups — the live digest-NULL
   baseline is eroding by design under the rerender wave, use snapshot `0d1b55f0` /
   bak tables), 27/27 serving (NB now 27, incl. the gap-planner's guide-loan-faqs of
   08-14), pre/post URL diff EMPTY, toolgolden 11/11, calculators IN PLACE, and the
   new-active-rows check from Q2.

## Standing cautions (carried + new)

- The 090 on matchLockedRow: still not re-run (08-12 attempt died on API 529); the
  owner-ruling substitute + council a625c326 APPROVED stand in. Re-run only if the
  mechanism is doubted again.
- 14 rerender-wave items were still working through this site tonight — expected,
  proven content-safe; do NOT read their digest-stamping as baseline damage.
- `lock_blocked_change` items are "the lock was exercised", never "the copy differed".
- The canary's 7 deferred items and 2 archived rows: if the owner WANTS an about
  page / guides index, un-archive + un-defer beats re-planning.
- Fetch pages at `pages.url`, never at a name-derived path (WRONG_CALLS 08-14 —
  the 404 body greps clean and reads as catastrophe).
