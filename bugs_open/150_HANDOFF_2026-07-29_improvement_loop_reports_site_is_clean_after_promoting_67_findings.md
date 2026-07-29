# 150 — the improvement loop ends at "No issues found — site is clean" after promoting 67 findings, and skips its own dispatch

**Filed:** 2026-07-29, from a hand-fired run of `improvement-sweep` (owner
instruction). **Status:** OPEN, unowned.
**Not to be confused with** `bugs_open/083` BY SLUG
(`…detected_findings_never_reach_a_handler`), which is about the loop never
RUNNING. This is about what it does WHEN it runs. Different mechanism,
different fix; 083 is a prerequisite for anyone ever seeing this one.

## Symptom

One firing of `improvement-loop` against gamesdesign.co.uk
(orchestration `30692439-43d2-4406-9fe8-9734c3f5689a`, 17:05:43→17:10:01Z
2026-07-29) promoted **67 work items** from `detected` to `triaged` and then
terminated at step **`complete_clean`**, whose configured success message is
**"No issues found — site is clean"**.

Because `complete_clean` is reached via `notify_scheduler_clean`, the run also
skipped `insert_rerender_item` → `spawn_dispatch` → `call_dispatch`. **The loop
dispatched none of the fixes it had just queued.**

## Root cause: three agents share one triage responsibility, and the parent's copy runs last

`triage_detected_items` appears as a step in **three** live agent definitions:

```
improvement-loop  : triage_findings
design-audit-agent: triage
site-review-agent : triage
```

The parent calls both children before running its own copy
(`call_design_audit` → `spawn_site_review` → `call_site_review` →
`triage_findings`). The action promotes **every** detected row for the site
(`triage_detect_items_action.go:108` — the WHERE clause is `site_id` + `status`
only), so the first copy to run takes them all and every later copy sees zero.

Then the branch reads the LAST writer:

```json
"check_has_findings": {"action": "conditional",
  "config": {"condition": "triage_result.has_items == true",
             "then_step": "insert_rerender_item",
             "else_step": "notify_scheduler_clean"}}
```

`triage_result` is the output_field of the parent's own `triage_findings`, which
by then has nothing to promote.

**Evidence, from the run's own `collected_data`:**

| key | value |
|---|---|
| `call_design_audit.response.triage_result` | `{"promoted": 67, "has_items": true}` |
| `call_site_review.response.triage_result` | `{"promoted": 0, "has_items": false}` |
| `triage_result` (parent's own step) | `{"promoted": 0, "has_items": false}` |
| `current_step` / `status` | `complete_clean` / `COMPLETED` |

And the promotion was a single statement, not a trickle — all 67 rows carry the
identical `triaged_at`:

```
 2026-07-29 17:08:32.778827+00 | 67
```

which is inside the `design-audit-agent` child's window (17:06:59→17:08:35),
20 seconds BEFORE the site-review child was even spawned (17:08:52).

## What is and is not damaged

- **The fixes still happen.** `build-pipeline-trigger` is enabled, ticks every
  120s and selects `triaged`+`build` independently of the loop, so the promoted
  items were dispatched anyway (first claim 17:09:48, `rerender-pages` running
  by 17:10:08). This is why the defect is invisible: the loop's dispatch branch
  is redundant with a scheduled task that works.
- **The closing rerender is lost.** `insert_rerender_item` would have created a
  priority-99 `needs_rerender` ("Re-assemble and deploy pages after improvement
  fixes") guaranteed to run after every fix. Nothing else guarantees it. In this
  run `rerender-pages` happened to create 32 rerender items of its own, but that
  is a side effect of the fixes it was given, not a closing pass.
- **The completion message is false**, and `final_result` carries it. Anything
  that reads the loop's outcome — a human, a report, a future gate — is told a
  site with 67 open findings is clean. This is the `016b §9` "trust the
  rendered artefact, not the status" pattern with the status being the artefact.

## Confidence

`[MEASURED]` — everything above is from the one run's stored state and the live
agent definitions, quoted inline.

`[INFERRED from a single run]` — that this happens on **every** run.
`orchestration_states` retention has cleared all history (all-history query for
`owner_agent_type='improvement-loop'` returns exactly **1** row — this one), so
there is no base rate to check against. The static argument is strong: the
parent's `triage_findings` is reachable only from `call_site_review`
(`next_step` AND `error_step`) and `call_completeness_discovery.error_step`, all
of which are downstream of `call_design_audit`, so a child always triages first.
The escape hatch would be a step that CREATES detected items between the last
child triage and the parent's — `site-review-agent.write_strategic_findings` is
the only candidate, and in this run it created no work items (`created_by` has
no site-review entries). **Confirm with a second run before asserting "always".**

## Fix candidates, ordered by what closes the door

1. **Make the bad state unrepresentable: one triage, one owner.** Remove the
   `triage` step from `design-audit-agent` and `site-review-agent` and let the
   parent own it. Then `triage_result` describes the run. Risk: those two agents
   are also called from other parents — check every caller before cutting.
2. **Decide the branch on the site, not on a step's output.** Replace
   `triage_result.has_items` with a `query_database` count of
   `status='triaged' AND pipeline='build'` for the site. Correct regardless of
   who promoted, and immune to a fourth triage appearing later.
3. **Accumulate instead of overwrite.** Have the branch read
   `triage_result.has_items OR call_design_audit.response.triage_result.has_items
   OR …`. Cheapest, and the worst of the three — it hardcodes today's agent list
   into a condition, so the next agent to gain a triage step reintroduces the bug
   silently.

Not a candidate: changing `complete_clean`'s message. The message is honest
about the branch it is on; the branch is what is wrong.

## How to verify a fix

Fire one sweep at a site with a non-empty `detected` pile and assert BOTH:

```sql
-- the loop must not claim clean when it promoted
SELECT current_step, collected_data->'triage_result'->>'promoted'
FROM orchestration_states WHERE orchestration_id = '<id>';
--    expect current_step='complete', not 'complete_clean'

-- and the closing rerender must exist
SELECT id, item_type, priority FROM site_work_items
WHERE site_id='<site>' AND item_key LIKE 'improvement_rerender%'
  AND created_at > '<run start>';
```

A green happy path is not enough here: the failing branch is the one that must
be exercised, so the site must genuinely have findings.

**Repro tool:** `docs/agent_docs/docs024_key_docs_latest/gauntlet_dead_cta/scripts/run_improvement_sweep_once.sh <site_id> <domain>`
— fires one sweep without enabling the disabled task. Read its header for the
blast radius first: the discovery half outnumbered the promotion half ~60:1 on
an unswept site.
