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

---

## 2026-07-31 — OWNED and FIXED IN CODE (session "bugfix 27"). **Stays OPEN: committed, not live.**

**State:** Go half committed `337fdd9af`; config half written as
`docs/agent_docs/sql_for_agents/281_improvement_loop_branches_on_site_state.sql` and
**deliberately NOT applied**; council submission `757cc7be-8551-4e43-9d1e-705b0977be1d`
(trailer `Council-Submitted:`, verdict owed). Workstream:
`docs/agent_docs/docs024_key_docs_latest/bugfix_150_improvement_loop_false_clean/`.
**Two steps are owed before this can close** — see § "What is owed" at the foot.

### The `[INFERRED from a single run]` marker is DISCHARGED — second observation

§ Confidence says "that this happens on **every** run" is inferred, because
`orchestration_states` retention had cleared all history and orchestration `30692439` was
the only row in existence. A control run fired on the **pre-fix** binary (v1.0.1218) before
changing anything reproduces it on a different site, on a different day:

**Orchestration `911ecdd8-140f-402f-99fd-aa89700afed2`, vetcomparison.uk, 2026-07-31 21:12Z.**
The site started with **0** `detected` items; the discovery half created them during the run.

```
call_design_audit.response.triage_result = {"promoted": 24, "has_items": true}
call_site_review.response.triage_result  = {"promoted":  3, "has_items": true}
triage_result (the parent's own copy)    = {"promoted":  0, "has_items": false}
current_step = complete_clean              status = COMPLETED
```
```sql
-- promoted in the run's window
SELECT count(*) FROM site_work_items
WHERE site_id='72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND triaged_at > '2026-07-31 21:12:00+00';
--  27
-- closing rerenders the clean branch skipped
SELECT count(*) FROM site_work_items
WHERE site_id='72b9e3a6-872f-4528-a6d6-7f205ea60f4d'
  AND item_key LIKE 'improvement_rerender%' AND created_at > '2026-07-31 21:12:00+00';
--   0
```

**27 findings promoted, 0 closing rerenders, terminal message "No issues found — site is
clean".** Two sightings, two sites, two days.

### CORRECTION to § Confidence — the escape hatch opens and does NOT help

That section names `site-review-agent.write_strategic_findings` as the one candidate step
that could create `detected` items between the last child triage and the parent's, and notes
it created none in the observed run. **In this run it did**: site-review promoted **3** items
of its own. The parent still saw 0, because the child triages *after* it writes. So the hatch
exists, fires, and changes nothing — the defect is more robust than this file allowed for,
not less. Any future fix that relies on "sometimes the parent will have rows" is unsound.

### What was built — and why it is NOT this file's candidate 1

`triage_detected_items` now also returns the **site-scoped** answer beside its call-scoped
one: `site_dispatchable` (bool) and `site_dispatchable_count` (int), counting
`site_work_items` in a dispatchable status for the target pipeline — **whoever promoted them,
in whatever order, including a fourth caller that does not exist yet.** Migration 281 points
`check_has_findings` at it. Concept register **WDS-015**; six regression tests in
`triage_detected_items_site_state_test.go`.

**Candidate 1 ("one triage, one owner") was deliberately not taken.** It is ranked first
here and it is the wrong first move: it needs an audit of every other parent of those two
child agents, and it leaves the identical defect available to the next agent that gains a
triage step. A site-scoped signal makes the ordering **irrelevant** rather than making one
ordering **mandatory**. Both children keep their triage steps.

**Candidate 2 ("branch on a `query_database` count") is what this implements — but in Go,
not in config.** The count belongs in the action that already knows the site, so every
present and future consumer gets it, instead of each caller remembering to add a step. It
also keeps the predicate testable and greppable, which a SQL string in a config row is not.

**`has_items` is deliberately UNCHANGED.** Measured before deciding: **four** live
conditions read a `has_items`, across three actions — `build-dispatch-loop.check_has_items`
(`pending.has_items`), `site-work-orchestrator.check_has_items` (`work_items.has_items`) and
`.check_has_fix_items` (`fix_items.has_items`), plus this one. The other three read their own
loader's output and are correct. Redefining the word here would repair one branch by making a
shared convention mean two things.

### SIBLING FINDING, recorded not fixed — a SECOND route to the same false "clean"

`check_audit_pass_limit` sends a site straight to `complete_clean` when
`get_audit_pass_count(site) >= 3`:

```json
{"condition": "pass_count_data.limit_reached == true",
 "then_step": "notify_scheduler_clean",     // -> complete_clean, "No issues found — site is clean"
 "else_step": "spawn_quality_discovery"}
```

So a capped site is told it is clean when what happened is **"we skipped auditing"** — and
because that branch is upstream of `triage_findings`, its `detected` pile is never promoted
at all, by anyone. **`[MEASURED 2026-07-31]` 0 of 25 sites have `get_audit_pass_count >= 3`**,
so this is latent rather than live today. Not fixed here: it needs an honest terminal step
(a distinct `complete_audit_limit` with its own message, so the outcome is queryable), which
is a different change from this one and worth its own decision. **Anyone verifying 150 must
pick a site with `passes < 3`, or the run short-circuits before the branch under test.**

### What is owed before this closes

1. **A chassis image carrying `337fdd9af`, rolled**, then the pod-grep on **every** replica
   with its positive control (`site_dispatchable` ≥ 1 and
   `TriageDetectedItemsAction: Starting` ≥ 1 in the same exec). A roll is not evidence —
   `bugs_open/153`. This session deliberately did not roll (owner's call: rolling kills
   other sessions' in-flight councils).
2. **Then** apply migration 281 and re-fire the sweep. Expect `current_step = 'complete'`
   with `triage_result.promoted = 0` and `site_dispatchable = true` in the same object —
   that pair *is* the bug, routing the other way — plus an `improvement_rerender%` item
   from the run.

**Do not apply 281 before step 1.** On a binary without the key the condition resolves to
nil → false → **every** run takes the clean branch, including the ones that promote, which
is strictly worse than the bug. Pinned by `TestConditionOnAPreUpgradeBinaryDoesNotSilentlyInvert`
as well as by the migration's own banner.
