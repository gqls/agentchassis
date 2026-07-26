# 083 — discovery findings written as `status='detected'` never reach a handler: the promoter runs only inside a task disabled since May

**Filed:** 2026-07-26 · **Branch:** `086_experience_loop` · **Status:** OPEN, diagnosed with evidence, not fixed
**Severity:** high, structural and silent. Nothing errors. Detectors report success, rows are
written, and the work is never done. **98 items are parked fleet-wide.**
**Class:** dormant-machinery / fail-silent delivery gap — the same family as `bugs_open/071`
(a gate that detects every broken link then discards the finding) and `063`.
**Found by:** the bugs_open/049 session, while establishing why three months of correct
phantom-link detection had produced no change on any live site.

---

## Symptom

`phantom_internal_link` has been detected 22 times. It has been **fixed zero times, ever.**

```sql
SELECT item_type,
       count(*) FILTER (WHERE status='detected') AS stuck_detected,
       count(*) FILTER (WHERE status='complete') AS ever_complete,
       count(*) AS total
FROM site_work_items
WHERE item_type IN ('phantom_internal_link','unbuilt_internal_link','empty_internal_href')
GROUP BY 1;
```
```
 phantom_internal_link  | 18 |  0 | 22
 empty_internal_href    |  7 |  2 | 14
 unbuilt_internal_link  |  — |  — |  0     <- has never fired at all; see "second-order" below
```

Fleet-wide, **98 rows sit in `status='detected'`**, the oldest from 2026-07-14, spanning 21
item types — `page_rerender` (28), `undeployed_asset` (19), `phantom_internal_link` (18),
`empty_internal_href` (7), `empty_section` (4), and 16 others.

## Root cause — one promoter, and it lives inside a disabled scheduled task

Discovery checks write findings with `Status: "detected"` (the convention across
`platform/orchestration/actions/discovery_checks/*.go`). The dispatch loop cannot see that
status: both `claim_work_item_action.go:102` and `load_work_item_actions.go:559` filter
`status IN ('triaged','approved')`.

Exactly one thing bridges the gap — `TriageDetectedItemsAction`
(`platform/orchestration/actions/triage_detect_items_action.go:91-103`), whose own header says so:

> `TriageDetectedItemsAction promotes discovery findings from status='detected' to
> status='triaged' with pipeline='build' so the dispatch loop picks them up.`
> `Used by: improvement-loop agent (after all discovery agents complete)`

It is registered (`registry.go:819`) and it is correct — no type filter, it promotes everything
for the site. **It is simply never called**, because the only workflow carrying that step is the
`improvement-loop` agent, and the only thing that fires that agent is the `improvement-sweep`
scheduled task, which is **disabled**:

```sql
SELECT name, enabled, last_triggered_at FROM scheduled_tasks WHERE name='improvement-sweep';
--  improvement-sweep | f | 2026-05-02 10:11
```

The agent definition itself is alive and correctly configured — this is not a broken workflow,
it is an unscheduled one:

```sql
SELECT type, is_active FROM agent_definitions WHERE type='improvement-loop'
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;             --  improvement-loop | t
```

**Two other code paths write `status='triaged'` and neither helps**: `claim_work_item_action.go:186`
and `load_work_item_actions.go:917` both *release an already-claimed item* back to the queue when
an AI endpoint is unhealthy. Nothing else promotes `detected`.

So: **a detector that files a finding as `detected` is writing to a queue with no consumer.**

## Why this is not `bugs_open/033`

033 is the `needs_human_review` pile (376 rows) — items routed to a human surface that nobody
works. This is a different status with a different intended consumer: `detected` is meant to be
picked up by *machinery*, automatically, and the machinery is switched off.

The distinction matters because **`detected` was deliberately chosen over `needs_human_review`
precisely to avoid 033.** `bugs_closed/054`'s record states it plainly:

> On that basis `bugs_open/054` wired the new **chrome dead-control** finding into a
> *draining* pathway (`status='detected'` + `handler_agent='nav-link-fixer'`, the
> phantom-links convention) rather than adding a 125th row to the unread
> `needs_human_review` pile.

That reasoning was sound and the premise was false: the "draining pathway" does not drain, and
the "phantom-links convention" it copied has a lifetime completion count of zero. **Anyone
choosing `detected` today to avoid 033 is choosing a quieter version of the same problem** —
which is the real cost of this bug, because it keeps being chosen.

## Second-order effect: detector work ships and changes nothing

This is why it is worth fixing rather than noting.

`unbuilt_internal_link` (bugs_open/049 candidate 4) was built, tested, committed and is **live in
the running pod** — `docker.io/aqls/agent-chassis:v1.0.1165`, pod-grep count 2. It has **never
produced a single row**. Not because it is broken: because discovery only runs when a human fires
it, and the sites carrying its target class had not been swept. Last discovery item written:

```
finetuning.uk      2026-05-01     gaswholesalers.com  2026-04-25     vetcomparison.uk  never
```

Those were the three worst sites in 049's audit. Coverage and delivery fail together and for the
same reason: `improvement-sweep` was the periodic driver of *both* the discovery run and the
triage that follows it.

## Fix candidates

1. **Re-enable `improvement-sweep`** (`UPDATE scheduled_tasks SET enabled=true`). Live
   immediately, no image roll. ⚠️ **Owner decision, not a platform one** — it fires fix agents
   fleet-wide on a 180s interval and spends credits, and something disabled it deliberately on
   2026-05-02. `[UNVERIFIED]` why it was disabled; I did not find a record and did not assume one.
   Read that reason before flipping it.
2. **Give triage its own scheduled task**, decoupled from the fix loop: promote `detected` →
   `triaged` on a slow cadence for item types whose handlers are known-good, leaving discovery
   itself manual. Smaller blast radius than (1); it drains the existing pile without also
   re-arming fleet-wide auto-fixing.
3. **Refuse the write instead of parking it.** Make `insertWorkItem` reject (or loudly warn on)
   `status='detected'` when nothing is scheduled to promote it — the platform's own
   silent-empty-render class, applied to queues. Turns a silent park into a visible error at the
   point the finding is filed.
4. **Check the handler is real before celebrating the detector.** Related but distinct:
   `bugs_open/077` (detector predicate wider than its handler) and `078` (NULL handler_agent
   livelocks the dispatcher). A promoted `phantom_internal_link` routes to `page-build-handler`,
   and `[UNVERIFIED]` whether a whole-page rebuild actually repairs an LLM-authored bad href.
   **Do not enable (1) or (2) without checking that**, or the pile moves from `detected` to
   `failed` and nothing improves.

## How to verify a fix

1. `SELECT count(*) FROM site_work_items WHERE status='detected'` falls and **stays** fallen —
   a one-off drain that refills is not a fix.
2. `phantom_internal_link` reaches a non-zero `complete` count for the first time.
3. The items that complete are verified against the **live page**, not the row status —
   `complete` is not proof the work happened (016b, and `bugs_open/017`'s whole subject).

## Landmines

- **A zero completion count is not evidence the handler is broken.** Here it means the items
  were never dispatched at all. Distinguish "never claimed" from "claimed and failed" before
  blaming a handler: `claimed_at IS NULL` on all 18.
- **`approved` is a status no Go code ever writes** (recorded in `bugs_open/033`), so the
  dispatch filter `status IN ('triaged','approved')` is effectively `= 'triaged'`.
- Do not "fix" this by having detectors write `triaged` directly. That bypasses the triage
  stage's judgement and would auto-dispatch every finding the moment it is detected — including
  the ones a human should see first. The gap is the missing promoter, not the status.

## Related

- `bugs_open/049` — where this was found; its mechanism 2 detector is the worked example of a
  live, correct, never-fired check.
- `bugs_open/033` — the `needs_human_review` pile. Sibling, different status, different consumer.
- `bugs_closed/054` — chose `detected` believing it drained. The quote above is from its record.
- `bugs_open/071` — the build-time gate that detects broken links then discards them. Same
  fail-silent shape one stage earlier.
- `bugs_open/077` / `078` — handler-side hazards that fix candidates 1 and 2 must clear first.
