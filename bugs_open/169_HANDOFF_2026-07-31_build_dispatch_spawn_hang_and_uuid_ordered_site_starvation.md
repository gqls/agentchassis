# 169 — a `build-dispatch-loop` spawn hung 38+ min mid-step, and separately, `build-pipeline-trigger`'s site-selection query starves sites by raw UUID order, not by wait time

> **PART B IS FIXED AND LIVE — 2026-08-02 (session "bugfix 19"). PART A IS
> UNTOUCHED AND THIS BUG STAYS OPEN FOR IT.**
>
> **Your part B was right in every particular, and the owner ruling you asked for
> was given on 2026-08-02: FIFO by oldest-eligible-item.** Shipped as
> `sql_for_agents/284` — `ORDER BY wi.created_at ASC, wi.priority ASC, wi.id ASC`,
> `DISTINCT ON` dropped as redundant under `LIMIT 1`. Cross-site priority stays
> deliberately unimplemented (a priority-major order just re-keys the starvation
> from UUID to priority; an aging scheme needs an owner-agreed scale constant).
>
> Two refinements to your write-up, both in your favour:
> - It is **not** "possibly working as designed": the order is not merely
>   unfair-by-accident but **deterministic**. `DISTINCT ON (site_id)` FORCES the
>   sort to lead with `site_id`, and `priority` is projected away before any
>   cross-site comparison, so lowest-UUID wins every tick and priority could never
>   influence which SITE was chosen. Starvation was certain, not possible.
> - Your dartsonline observation reproduced exactly: gamesdesign.co.uk, 14th of 17
>   by UUID, held the fleet's oldest eligible item for **3 days 10 hours** and was
>   selected zero times. First tick after 284 picked it.
>
> **You also flagged the `wi.domain = 'build'` seed drift** in
> `sql_for_agents/052`. Fixed the same day — and it was worse than a stale
> comment: one *operative* `UPDATE scheduled_tasks` in that file would have
> written the non-existent column back into the LIVE scheduler `pre_query` if
> re-run.
>
> **A defect your part B could not have seen, found while verifying the fix:** the
> selector and `LoadWorkItemsAction` disagree about what "dispatchable" means (the
> loader adds `approval_mode` and `depends_on` clauses), so the selector could
> hand the loop a site whose only item the loader refused — loop loads 0,
> completes cleanly, claims nothing, site stays eligible, picked again for ever.
> **FIFO ordering converts that from intermittent to permanent**, so 284 alone was
> a fleet stall. Both are fixed; see `bugs_closed/176`. If you are reading part B
> to understand dispatch, read 176 alongside it — neither change is safe alone.
>
> **PART A (the 38-minute spawn hang) has NOT been investigated by this session**
> and no part of the above bears on it. It remains exactly as you left it,
> including your instruction to run `090` before committing to a cause.

**Filed 2026-07-31, ~21:00 UTC (dartsonline.com site-fix session).** **Status: OPEN,
UNDIAGNOSED — this file is a handoff, not a root-cause claim.** Per this repo's own
owner ruling (CLAUDE.md, "Diagnosis before debugging"), a bugs_open file asserting a
cross-cutting/structural cause needs to go through `090_TRIGGER_needs_diagnosis_v1.sh`
or state plainly why it substituted equivalent first-hand verification. **Neither
happened here — I ran out of session budget mid-investigation and am handing off
deliberately before drawing a conclusion I haven't earned.** Everything below is
first-hand observed evidence (queries run, timestamps, correlation IDs), not a
diagnosis. Run `090` on this before committing to a fix.

## Two things were found, possibly related, possibly not — don't conflate them

### A) A specific spawn hung mid-step for 38+ minutes with zero progress

While fixing real content/nav bugs on dartsonline.com (see `PLAN`/context below), I
directly dispatched `build-dispatch-loop` (bypassing the normal `build-pipeline-trigger`
site-selection) for `site_id=5fe8785b-223d-41a3-88ee-c07187622381` (dartsonline.com),
correlation `9da39de8-eeb6-41a3-ab53-9640cc372e7a`. It chained through work items
correctly (iter_0 through iter_3 all `COMPLETED` in ~1.5–2 min each — index,
guides-index, news-index, new-arrivals), then got stuck:

```sql
SELECT correlation_id, status, current_step, EXTRACT(EPOCH FROM (now()-last_activity))::int AS since_s
FROM orchestration_states WHERE correlation_id = '9da39de8-eeb6-41a3-ab53-9640cc372e7a'::uuid;
-- one row: EXECUTING_STEP | process_item_iter_4_spawn_handler | since_s climbing steadily
--   (71s at first check, 2173s ~ 36min at last check, no step change in between)
```

The underlying work item never left `claimed`:

```sql
SELECT id, status, claimed_at, claimed_by, attempt_count
FROM site_work_items WHERE id='b286f2f5-6f7f-4205-ac70-f65bb9197633';
-- claimed | 2026-07-31 20:18:09 | build-dispatch-loop | attempt_count=0
-- (still claimed at 20:58:00 — 39m51s, right at the documented 40-min claimed-item-timeout)
```

That item was `Rerender page: tool-setup-builder` — targeting a page that, AT THE TIME
of this claim, had **zero `page_components` rows** (empty page, `build_status='planned'`,
never actually built — see part B of this session's other work for why). I do NOT know
whether the emptiness of the target page caused the hang, is coincidental, or whether
this class of hang would happen on any page. **That is exactly the kind of question 090
should answer, not this file.**

**What this is NOT**: I checked `bugs_open/029` (hung spawns saturate the dispatch
`concurrency_group` and halt ALL builds fleet-wide) first. That bug's signature is
`AWAITING_RESPONSES` after a spawn succeeds and the child never replies, and its
smoking gun is `build-pipeline-trigger` completely stopping (no new orchestrations at
all). Neither matched here: this orchestration was `EXECUTING_STEP` (stuck *inside* the
spawn action itself, before any child could even exist to not-reply), and
`build-pipeline-trigger` kept ticking normally the whole time
(`last_triggered_at`/`last_completed_at` advancing every ~120s throughout). **Different
phase, possibly a different bug** — don't assume it's 029 without checking.

I worked around it rather than debug it live: dispatched a SECOND, independent
`build-dispatch-loop` call for the same site (correlation
`09b7d25c-5d67-47ff-b0db-5968c684c708`), which correctly claimed and processed a
different set of triaged items (9 rows, all `COMPLETED` including one
`complete_error` worth a look) without colliding with the stuck one — confirming
per-item claiming is atomic and a second loop is a safe workaround, not a race.

### B) `build-pipeline-trigger`'s site-selection query orders by raw `site_id`, not urgency

Read the LIVE query (not the possibly-stale seed file — `docs/agent_docs/sql_for_agents/052_build_pipeline_trigger.sql`
on disk still has an `AND wi.domain = 'build'` clause that **does not exist** in the
live `agent_definitions` row; the live query has no domain/pipeline filter at all —
seed vs system drift, note it if you touch that file):

```sql
SELECT default_config #>> '{workflow,steps,find_dispatchable_site,config,query}'
FROM agent_definitions WHERE type='build-pipeline-trigger' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```
```sql
SELECT DISTINCT ON (wi.site_id) wi.site_id::text, s.domain
FROM site_work_items wi JOIN sites s ON s.id = wi.site_id
WHERE wi.status IN ('triaged','approved') AND wi.attempt_count < wi.max_attempts
  AND NOT EXISTS (SELECT 1 FROM site_work_items active WHERE active.site_id=wi.site_id AND active.status='claimed')
ORDER BY wi.site_id, wi.priority ASC LIMIT 1;
```

`ORDER BY wi.site_id, wi.priority` sorts candidate sites by **raw UUID**, then
priority only within a site's own rows. There is no age/wait-time/priority ordering
**across** sites. Observed consequence, live, not hypothetical: with 13 sites
simultaneously backlogged (measured 2026-07-31 ~19:30 UTC — robot-hands.com 1,
relojistas.com 21, vetcomparison.uk 12, webdesign.co.uk 1, gamesdesign.co.uk 35,
gaswholesalers.com 31, vonc.com 19, ai-agent-orchestration.com 38, dartsonline.com 22,
finetuning.uk 42, fundamentallyai.com 14, oufe.com 9, leopardessconsulting.co.uk 34),
`dartsonline.com` (site_id starting `5fe8...`) sat with 22 fully-eligible triaged items
completely untouched by the scheduler for over 20 minutes (from ~20:33 to past 20:56),
while `build-pipeline-trigger` ticked reliably every ~120s the entire time — because
sites with lexicographically smaller UUIDs kept having eligible work and kept winning
the `LIMIT 1`. **Re-run the query above yourself before trusting this figure — it is a
snapshot, and the backlog composition changes constantly.**

I do not know whether this is "working as designed" (accepted unfairness for a small
number of sites) or a real defect. It is at minimum worth an owner opinion: should
site selection be FIFO by oldest-eligible-item, or priority-weighted across sites,
rather than UUID order?

## What's already fixed and NOT part of this handoff

Real dartsonline.com content/nav bugs (missing Tools nav link, an empty never-built
tool page, thin "Start Here" content) were diagnosed and fixed properly this session —
new tool page built via the correct `tool-generator` pipeline, content-gap-planner
dispatched for the thin page, "News" already confirmed live in nav. Those fixes are
sound regardless of what's in this file; they are just **stuck in the queue described
in part B** along with everything else. See this session's own transcript / whatever
workstream doc captures it for that side — **not repeated here**, this file is
specifically the platform-mechanism finding, not the site fix.

## Recommended next step

Run the diagnosis loop rather than hand-theorizing further:

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/090_TRIGGER_needs_diagnosis_v1.sh \
  "a chained build-dispatch-loop orchestration (corr 9da39de8-eeb6-41a3-ab53-9640cc372e7a) hung at EXECUTING_STEP on process_item_iter_4_spawn_handler for 38+ minutes with zero last_activity change, for site_work_items id b286f2f5-6f7f-4205-ac70-f65bb9197633 (page_rerender, target page had zero page_components at claim time)"
```
Consider whether to file part B (site-selection ordering) as its own diagnosis target,
or fold it in — they may share a cause (e.g. if the hang itself is what's tying up a
concurrency slot that would otherwise let dartsonline's OWN items get claimed faster),
or may be unrelated. **Don't assume either way; that's what the loop is for.**

Before dispatching, re-check current state — by the time you read this, the 40-minute
claimed-item-timeout may have already reset item `b286f2f5-...` back to `triaged`, in
which case it may have already retried and either hung again (attempt_count 1) or
resolved (page now exists under a different id, built via `tool-generator` — see
above — so a retry against the OLD page_id may now correctly fail fast rather than
hang, which would itself be informative):

```sql
SELECT id, status, claimed_at, attempt_count, updated_at FROM site_work_items WHERE id='b286f2f5-6f7f-4205-ac70-f65bb9197633';
SELECT status, current_step FROM orchestration_states WHERE correlation_id='9da39de8-eeb6-41a3-ab53-9640cc372e7a'::uuid ORDER BY created_at DESC LIMIT 1;
SELECT s.domain, count(*) FROM site_work_items wi JOIN sites s ON s.id=wi.site_id
  WHERE wi.status IN ('triaged','approved') GROUP BY s.domain ORDER BY count(*) DESC;
```

## Related

- `bugs_open/029` — hung spawns saturate the `dispatch` concurrency group; checked and
  ruled out as an exact match (different phase, different scheduler symptom) but same
  general class ("a stuck spawn breaks fleet throughput silently") — worth reading
  alongside this.
- `bugs_closed/048` — no-op `pre_query` starves its concurrency group forever; a
  precedent for "the scheduler ticks happily while doing nothing useful."
- `docs/agent_docs/sql_for_agents/052_build_pipeline_trigger.sql` — the SEED file is
  stale relative to the live `agent_definitions` row (see part B); don't trust it as
  documentation of current behaviour without checking live.

---

# PART A — ROOT CAUSE FOUND AND FIXED IN CODE, 2026-08-02 (session "bugfix 27")

> **STATUS: STILL OPEN.** The fix is committed (`fe34fd04f`) and **inert until a
> chassis image is built past it and rolled**. This repo's bar for `/bugs_closed/` is
> **fixed AND live**, so part A stays here until the roll and the post-roll check
> below. Part B remains fixed and live (`sql_for_agents/284`, `bugs_closed/176`).

## The cause: a local action has no deadline, anywhere

Nothing in the chain bounds the handler call:

```
continueExecution → executeStep → executeLocalAction → executeAction → handler(ctx, params)
                                                        coordinator.go:1563
```

`executeAction` recovers panics and bounds nothing; no caller wraps the context. So an
action that blocks on a network call parks its orchestration at `EXECUTING_STEP`
indefinitely, and for `build-dispatch-loop` holds its `site_work_items` row in
`claimed` until the 120s `claimed-item-timeout` reaper expires it. That is exactly the
shape you observed: `process_item_iter_4_spawn_handler`, `since_s` climbing, item stuck
`claimed`, right up against the documented timeout.

**`timeout_seconds` looks like the bound and is not.** It is parsed
(`coordinator.go:1348-1353`) into `execCtx.TimeoutSeconds` with the comment *"leave it
for the action to set a default"* — and **no action reads it as a deadline on its own
execution** (checked across all 271 action files).

## Your instance, re-checked

- orchestration `9da39de8` is gone (terminal rows are reaped ~24h).
- item `b286f2f5` is now **`failed` at `attempt_count=3`** — it retried and exhausted.
- so the instance cannot be re-examined; the class was measured instead.

## The class, measured (`orchestration_state_audit`, 2026-07-31 → 08-02)

| measure | value |
|---|---|
| distinct runs entering a `*_spawn_handler` step | **165** |
| runs that ENDED at one | **1** (status `FAILED`) |
| `spawn_*` step executions | **6,951** |
| p50 / p95 / p99 | **0s / 18s / 24s** |
| executions above 300s | **exactly 1**, at 14,475s |

Bimodal with nothing in between: healthy spawns finish in seconds, the pathological one
ran four hours. So the class is real, rare, and a generous bound catches it without
touching anything healthy.

## A thing worth knowing that is NOT a fix

Five orchestrations across three days each stalled for almost exactly **4h01m**. That is
`coordinator.go:831`'s `maxAge = WorkflowPlan.TimeoutSeconds × 3` stale-orchestration
guard. It is an *orchestration-age* check inside the message-handling path: it fires only
when a message next arrives, marks the row failed, and **never interrupts the blocked
goroutine**. It is why these eventually clear, and it is not a step timeout.

## The fix (`fe34fd04f`, register **RSH-004**, council `2c6800e6`)

`executeLocalAction` derives a deadline and passes it to the handler. Default **600s**
(~25× the measured all-step p99.9 and 25× the spawn p99); per-step override
`local_action_timeout_seconds`; **`<=0` = explicitly unbounded**, logged Warn every time;
`DISABLE_LOCAL_ACTION_TIMEOUT=true` is a fleet-wide kill switch restoring the previous
behaviour with no rebuild. A malformed value falls back to the **default**, never to
unbounded. A `DeadlineExceeded` is wrapped to name the action, step and elapsed time,
then routed through the existing `handleActionError` path so `error_step` is unchanged.

**Deliberately NOT wired to `timeout_seconds`** — 53 of the 64 live steps carrying that
key are `call_agent` and most of the rest are waiting semantics (`await_approval` carries
86400). It means "how long to wait for something EXTERNAL"; conflating the two is the
class RFC 006 was decided on, and would hand `await_approval` a 24-hour execution
deadline. **Deliberately NOT goroutine-abandonment** — it would regain control from a
ctx-ignoring action but leaks the goroutine, and a late write into an already-failed step
is a worse hazard than the hang.

## On your instruction to run 090 first — done, and here is what it returned

Filed as you asked: `RUN_CORRELATION_ID=3ca53d45-4826-4935-96a3-a0af4d194d91`. It ran
five iterations and **produced five `bundle` artifacts and no diagnosis** — no `fix_plan`,
no `council_report`, nothing under its correlation, and the work item completed. (Note
for anyone repeating this: there is **no `verdict` artifact kind** — the kinds that exist
are `fix_plan`, `council_report`, `bundle`, `escalation`, `iteration_note`. Waiting on
`kind='verdict'` waits for ever.)

So, per CLAUDE.md's owner ruling of 2026-07-31, stating plainly what was substituted:
**the root cause above rests on first-hand verification, not on the loop's verdict** —
the call chain was read end to end, every one of the 271 action files was checked for a
reader of `TimeoutSeconds`, and the distribution was measured over all 96,047 recorded
step executions. The loop's bundles corroborate the surrounding facts (every other
`build-dispatch-loop` run in the window completed in 20–50s) but did not deliver a
verdict.

**That the diagnosis loop completed without producing one is itself worth a look, and is
NOT claimed here as a defect** — one run is not a rate, and no other diagnosis run exists
in the last 10 days to compare against.

## What is owed before this can close

1. A chassis image built past `fe34fd04f`, rolled.
2. Pod-grep both replicas for a string this change ADDED and one it did not, in the same
   exec (a roll is not evidence — `bugs_open/153`):
   `strings /app/agent-chassis | grep -c local_action_timeout_seconds`  (expect ≥1)
   `strings /app/agent-chassis | grep -c "Executing local action"`      (control, ≥1)
3. Re-measure: no run should END at a `*_spawn_handler` step; a step that exceeds its
   deadline must FAIL with an error naming the action and elapsed time.

## Council verdict, and the scope objection it raised (2026-08-02)

**APPROVED** — correlation `2c6800e6-91c5-4344-acc9-50e909894a40`, *"approved with 2
advisory objection(s) — none high-severity"*, round 1, 8 seats abstained.

Every checkable objection was run rather than argued:

- **`reuse_agent` / `prior_art_librarian`** — is there already a context-deadline wrapper
  this should extend rather than duplicate? **No.** The package's only
  `context.WithTimeout` calls are `helpers.go`'s fixed 5s/10s query timeouts and one 60s
  at `coordinator.go:3972`, **all on `context.Background()`** — deliberately detached, a
  different pattern. Nothing to extend.
- **`guardian` (low)** — does an existing test assume no-timeout behaviour? **No.** The
  only test file referencing `executeLocalAction`/`executeAction` is the new one.
- **`guardian` (medium) / `architecture`** — has this class been deflected upward before?
  **No.** The only mention of `executeLocalAction` in `bugs_open/`, `bugs_closed/` or
  `architecture_review/` is this file.
- **`debug_historian` (medium)** — no deploy-verification named in the submission. Fair;
  it is in "What is owed" above. Its operational half is worth repeating: **a chassis roll
  kills an in-flight council run** — check for one before rolling this.

**The `architecture` seat's scope objection is upheld and routed, not rebutted.** It says
a new reserved step-config key plus a new fleet-wide env contract on the coordinator's
step-execution seam is architecture-scope, and that my inline argument for why RFC_010's
"opt-in, default OFF" ruling does not apply *belongs in an RFC where a human can check
it*. That is right, and it is the same shape as `bugs_closed/124`. Filed as
**`architecture_review/RFC_011_a_fleet_wide_execution_deadline_on_the_step_seam.md`**
with three costed options. Per the 2026-07-28 ruling, a scope objection is not answered
by better measurements — so nothing here was resubmitted.
