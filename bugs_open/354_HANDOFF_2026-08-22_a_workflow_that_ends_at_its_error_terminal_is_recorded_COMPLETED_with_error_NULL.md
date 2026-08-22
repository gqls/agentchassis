# 354 — a workflow that ends at its ERROR TERMINAL is recorded `COMPLETED` with `error` NULL

**Filed 2026-08-22** by the `bugs_open/260` lane (render-fallback), which met this while
establishing that a symptom it was blamed for was pre-existing. **Status: OPEN, UNOWNED.**
It is filed so somebody can own it; this lane is not working it.

> **On the 090 loop, stated plainly because the 2026-07-31 owner ruling requires it.**
> This is a cross-cutting structural claim, so it went to the diagnosis loop **before** this file
> asserted a cause. Intake correlation `2f17e660-70fd-4c69-8a73-4c5ab28b792c`, run correlation
> **`025968a3-6e2e-4f2c-9dbb-f351820a871a`** (the key the artifacts are written under).
> **The verdict had not landed when this file was written** — read it before acting:
> ```sql
> SELECT current_step, status FROM orchestration_states
>  WHERE correlation_id = '025968a3-6e2e-4f2c-9dbb-f351820a871a'::uuid;
> SELECT * FROM diagnosis_artifacts
>  WHERE correlation_id = '025968a3-6e2e-4f2c-9dbb-f351820a871a'::uuid;
> ```
> Everything in §2 and §3 below is first-hand: a query whose output is quoted, or a function read
> at the deciding arm. Nothing is inferred from the loop.

## 1. The one-paragraph version

`SagaCoordinator.completeWorkflow` sets `state.Status = StatusCompleted` **unconditionally** and
never looks at `collected_data["__step_error"]`, which `routeToErrorStep` wrote when a step failed.
It also never sets `state.Error`. So a run that failed a step, routed to its `error_step`, and
stopped at a terminal named `complete_error` lands in `orchestration_states` with **exactly the
same two columns** as a run that executed every step: `status='COMPLETED'`, `error IS NULL`. The
`current_step` value does distinguish them and **nothing reads it**; the only detail — the message
and the failing step name — lives in `collected_data`, which is reaped in about two days.

**This is a LEGIBILITY defect, not a routing defect.** `routeToErrorStep`'s own doc comment names
the legitimate case (`coordinator.go:3933`): "the dispatch loop can mark a work item as failed and
continue to the next item". That run genuinely completed its work. The defect is that the record
cannot tell that run apart from one that gave up at the error terminal, so **every consumer keying
on `status` or `error` reads both as success, permanently.**

## 2. Evidence — measured 2026-08-22, with a control that could have come out otherwise

**(a) The population.** `[MEASURED 2026-08-22 09:1xZ]`

```sql
SELECT current_step, count(*), count(*) FILTER (WHERE error IS NOT NULL) AS with_error_col
FROM orchestration_states
WHERE collected_data ? '__step_error' AND status='COMPLETED'
GROUP BY 1 ORDER BY 2 DESC;
```

| `current_step` | COMPLETED runs carrying `__step_error` | of which `error` column is non-NULL |
|---|---|---|
| `complete_error` | **50** | **0** |
| `complete` | 8 | 0 |
| `complete_idle` | 3 | 0 |
| `complete_invalid` | 2 | 0 |

**(b) The control — and this is the load-bearing half.** If `complete_error` were merely a common
terminal, the table above would say nothing. It is not:

```sql
SELECT current_step, count(*) FROM orchestration_states
WHERE NOT (collected_data ? '__step_error') AND status='COMPLETED'
GROUP BY 1 ORDER BY 2 DESC LIMIT 8;
```

Of **2,997** COMPLETED runs carrying no `__step_error`, exactly **1** ends at `complete_error`
(2,783 `complete`, 126 `complete_idle`, 38 `complete_approved`, 32 `complete_revise`, …).
So `current_step='complete_error'` is a near-perfect discriminator — **the information needed to
tell the two populations apart is already on the row.** This control could have come out the other
way, and would have refuted the whole entry if `complete_error` were a routine terminal.

**(c) What is actually being swallowed.** Grouped by the failing step (`status='COMPLETED'`):

| failed step | runs | sample message (truncated) |
|---|---|---|
| `load_page_record` | 22 | `either page_name …` |
| `save_tool` | 13 | `create_tool_component: tool birth refused (instance scope): …` |
| `process_item_iter_0_call_handler` | 8 | `apply_fix failed: fix_component_template: scope_component_in…` |
| `validate_content` | 5 | `validate_page_content: content vali…` |
| `call_content_writer` | 3 | `…_render_section failed: render_c…` |
| `audit` | 3 | `Request timed out (code: TIMEOUT)` |
| `deploy_css` | 2 | `failed to create blob for noted.co.uk/assets/css/styles.css: context deadline exceeded` |
| `spawn_dispatch` | 2 | `spawn_agent: failed to setup topics: …` |
| `save_sections`, `mark_item_failed`, `call_dispatch`, `review_prior_art`, `review_guardian` | 1 each | — |

`deploy_css` is the shape that shows the cost plainly: a site's stylesheet was not written, and
the run says `COMPLETED` with `error` NULL. `mark_item_failed` is the recursive one — the step that
exists to record a failure failed, and that too reads as success.

**(d) Retention, so nobody sizes this from `orchestration_states`.** `[MEASURED 2026-08-22]`
That table holds **08-21 and 08-22 only** (2,384 + 723 rows), plus 24 stragglers stuck since July.
So the counts above are **~2 days of traffic, not a census** — roughly 25–30 swallowed runs a day at
current load. Anyone quoting "50" as a total is quoting a retention window.

**(e) The failure IS durably recorded — just not where anyone reads.** `routeToErrorStep` calls
`logAgentError` with `Severity = "error"` (`coordinator.go:3963-3964`), and the join lands:

```sql
WITH swallowed AS (
  SELECT orchestration_id::text AS oid FROM orchestration_states
  WHERE status='COMPLETED' AND current_step='complete_error' AND collected_data ? '__step_error')
SELECT count(*), count(*) FILTER (WHERE EXISTS (
  SELECT 1 FROM agent_error_log l WHERE l.orchestration_id = s.oid)) FROM swallowed s;
```
→ **50 of 50**. `agent_error_log` retains a full month (45,417 rows back to 2026-07-23).
**This is the good news and it shapes the fix: no evidence is being lost, it is being filed
somewhere no consumer joins to.**

## 3. Root cause, read at the deciding arm

`platform/orchestration/coordinator.go`, line numbers valid at **`f11851c27`** (the file is
unmodified in the tree, so they are HEAD's, not a working copy's):

- `routeToErrorStep` (**:3956**) writes `state.CollectedData["__step_error"] = {failed_step, message}`,
  logs to `agent_error_log` at severity `error` (**:3963-3964**), sets `state.CurrentStep = errorStep` and
  `state.Status = StatusExecutingStep`. Correct, and deliberate.
- `completeWorkflow` (**:4443**) then runs when that error step reaches its terminal, and its whole
  status decision is one line — **:4456**:
  ```go
  state.Status = StatusCompleted
  ```
  No read of `CollectedData["__step_error"]`. No write to `state.Error`. There is no branch here
  that could distinguish the two cases, which is why the `error` column is NULL in **all 63** rows
  in §2(a) rather than in most of them.

## 4. Why this is not one of the bugs already open

The **symptom** is well known and is written down in several places — `LANDMINES.md` carries it at
least four times (`complete_invalid`, the build-pipeline surface at "A THIRD SURFACE, and it is the
quietest", the experience-planner entry, the 090-run entry). Those entries are all *"do not be
fooled by this"* notes for a reader who already has a suspicion.

What does not exist is a bug that owns **making the record honest**. The per-symptom cases are
filed and several are closed — `bugs_closed/344` (the dispatch loop overwriting a re-triaged row),
`bugs_open/348` (its named natural-damage case, whose own mechanism was corrected to point at 344),
`bugs_open/210`, `bugs_open/220`. Each fixed or described **one** consumer. This entry is the
shared cause underneath them: as long as `status`/`error` cannot express "ended at the error
terminal", every future consumer inherits the same blindness, and the next one will be filed as its
own bug too. **Do not re-fix those from here; they are the damage, this is the record.**

## 5. Fix candidates, ordered by what closes the door

1. **Distinguish the terminal status.** `completeWorkflow` consults `__step_error` (or, better,
   whether the terminal step was reached via `routeToErrorStep` — a flag set there, not a string
   match on a step name) and records a distinct status. This makes the bad state
   **unrepresentable**: no consumer can read success by accident.
   ⚠ **This adds to a shared vocabulary that many consumers switch on, so by the 2026-07-29 owner
   ruling §1 it changes what the mechanism GUARANTEES and is architecture-scope — RFC, not a bug
   patch.** Every reader of `orchestration_states.status` and of `StatusCompleted` must be
   enumerated first, not asserted about.
2. **Populate `orchestration_states.error` at completion** from `__step_error.message` when the run
   reached a terminal via `routeToErrorStep`. Contained, additive, no new status vocabulary, and it
   puts the message on the row a reader already has open. Does **not** close the door — a consumer
   testing `status='COMPLETED'` alone still reads success — but it removes the "there is no message
   anywhere after 48h" half. **Enumerate what currently relies on `error` being NULL for a COMPLETED
   run before shipping** ("no consumer breaks" is a query, not an argument).
3. **Reporting only** — a view or sweep joining `agent_error_log` on `orchestration_id`, which
   §2(e) proves resolves 50/50. Cheapest, loses nothing, and closes no door at all: it depends on
   somebody choosing to look, which is the property that produced this bug.

**Whoever takes it: 1 and 2 are not alternatives.** 2 is the contained thing shippable now; 1 is
the RFC that makes it stop recurring. Shipping 2 and calling it done leaves the door open.

## 6. How to verify a fix

- **Induce, do not wait.** Dispatch any workflow whose step declares an `error_step` and make that
  step fail (`load_page_record` with no `page_name` is the commonest natural case in §2(c)).
  Then read `status`, `current_step` and `error` on that orchestration row.
- **Run the control in the same breath** — a run that completes *normally* must be unchanged.
  A fix that stamps every run with a non-NULL `error` passes the first check and is worse than the
  bug.
- **Re-run §2(a) and §2(b) together.** The population number alone cannot tell you whether the fix
  worked or whether traffic simply dropped; the control is what makes the reading mean anything.
- **Do not verify from `orchestration_states` alone after ~48h** — §2(d). Use `agent_error_log`,
  which retains a month.
