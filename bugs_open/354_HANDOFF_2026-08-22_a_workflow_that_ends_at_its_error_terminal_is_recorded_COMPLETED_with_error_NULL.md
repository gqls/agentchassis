# 354 — a workflow that ends at its ERROR TERMINAL is recorded `COMPLETED` with `error` NULL

**Filed 2026-08-22** by the `bugs_open/260` lane (render-fallback), which met this while
establishing that a symptom it was blamed for was pre-existing. **Status: OPEN — OWNED by the `bugs_open/307` lane from 2026-08-22 (they asked, I confirmed I was not working it).**
It was filed so somebody could own it, and somebody has — the account now lives here, not in a fork.

> **On the 090 loop, stated plainly because the 2026-07-31 owner ruling requires it.**
> This is a cross-cutting structural claim, so it went to the diagnosis loop **before** this file
> asserted a cause. Intake correlation `2f17e660-70fd-4c69-8a73-4c5ab28b792c`, run correlation
> **`025968a3-6e2e-4f2c-9dbb-f351820a871a`** (the key the artifacts are written under).
> **VERDICT: CONFIRMED**, at iteration 1, landed 09:21Z the same morning (`stopped_by: confirmed`).
> Read it in full — it carries pointers this file does not:
> ```sql
> SELECT jsonb_pretty(result->'response'->'response') FROM site_work_items
>  WHERE id = '619ff617-ae57-4414-95f0-533a7216b0cd';
> ```
> It reached the conclusion **independently**, and its citations are its own, not this file's:
> `state.Status = StatusCompleted` in `completeWorkflow`; `state.CollectedData["__step_error"] = …`
> and `state.CurrentStep = errorStep; state.Status = StatusExecutingStep` in `routeToErrorStep`;
> plus a live row it fetched itself with its own SQL —
> `0e7762af-… | COMPLETED | complete_error | error NULL | __step_error.failed_step = save_tool`.
> Its coverage note is the sharpest statement of the defect: *"completeWorkflow's body (fully shown)
> reads CollectedData only to log its keys (`safeDataKeys`) and never inspects the `__step_error`
> value"*, and *"completeWorkflow never assigns to `state.Error` anywhere in its body"*.
>
> **It also named the scope a fixer should read next**, which this file had not: `continueExecution`,
> `handleOrchestrationStatus`, `getNextStepFromResult`, and — the two most relevant —
> `routeToErrorStepOrFail` and `failWorkflow`, which are the sibling arms that DO fail the workflow.
> Start there: the question a fix has to answer is why one arm records the failure and the other does not.
>
> Everything in §2 and §3 below is first-hand and was written before the verdict: a query whose output
> is quoted, or a function read at the deciding arm. The loop agrees with all of it; none of it is
> inferred from the loop.

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

> **ADDED 2026-08-22 by this lane at `bugs_open/307 [e24299]`'s prompting — the control's ONE
> exception is not noise, it is a third state.** I recorded "exactly 1" (they measured 2 on their
> sample) and never asked what it was. They did: those rows are `check_page_found`'s `else_step`
> firing — **a page genuinely not found, i.e. a SKIP, not a failure.** So the discriminator's only
> apparent false positive is a case that *should* be excluded, and requiring `__step_error` to be
> present excludes it for free. That is `bugs_closed/299`'s finding — **a skip is a third state,
> neither clean nor failed** — showing up inside this bug's own control. **The lesson is the same one
> §2(d) records against me:** I treated a small residue as rounding error rather than asking what it
> was, twice in one file, and both times the residue was load-bearing.

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
That table holds **08-21 and 08-22 only** (2,384 + 723 rows), plus 24 older rows.
> **CORRECTED 2026-08-22, same day, by `bugs_open/307 [e24299]`, the lane picking this bug up — I called those
> 24 rows "stragglers stuck since July" and never asked their status.** They are **all `CANCELLED`**
> (`GROUP BY status` on rows older than 48h: `CANCELLED | 24 | 2026-07-19 | 2026-07-24`, re-verified
> here before accepting it). They are not stuck — they are **unreaped**, which is a different defect
> and is the one described in §5 below. **The check I skipped was one clause long.** Filed in
> `WRONG_CALLS.md`: an anomalous row's AGE is not its diagnosis; ask what STATE it is in.
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

> **⚠ CORRECTED 2026-08-22 by `bugs_open/307 [e24299]`, the lane that took this bug over** — note
> that a SECOND live session shares the name `bugs_open/307` and is **not** the source of any of
> this; a bare-name message reaches one of them at random. Wording below is theirs, at their
> request. Both halves were independently re-verified by this lane before folding in.
>
> **CORRECTED 2026-08-22 — §5 candidate 1's cost estimate was reasoned from a world three days out
> of date.** Migration `466` (applied 2026-08-19) gave `orchestration_states.status` a FOREIGN KEY to
> a vocabulary table `orchestration_status_vocabulary`, carrying an `is_terminal` column. Adding a
> status is now one INSERT with a documented recipe, not the open-ended vocabulary change this
> section assumed. **But there is a trap directly underneath it:** `database-cleanup`'s arm 3 still
> deletes `WHERE status IN ('COMPLETED','FAILED')` — a literal — while arm 4 deletes
> `WHERE status NOT IN (… WHERE is_terminal OR is_pausable)`. **A new terminal status falls through
> both and is never deleted.** Verified against the live `scheduled_tasks.pre_query` row, not the
> migration file. This is not hypothetical: it is already happening to `CANCELLED` — every one of the
> 24 rows in the table older than 24h is `CANCELLED`, oldest **2026-07-19, 34 days**, while
> everything else is reaped at 24h. So candidate 1 must widen the cleanup arm in the same migration
> or it silently re-files `bugs_closed/294`.
>
> **So the net effect on §5's ordering is the OPPOSITE of what a first reading suggests, and this
> lane got it wrong out loud before being corrected.** The FK makes the *write* cheap; the *reaping*
> is the real cost, and it did not exist in the estimate at all. **Candidate 1 is more expensive than
> §5 says, not less.** The ordering may still hold — but nobody should inherit it as evidence, in
> either direction.
>
> **On `RFC_023`, and this lane was too broad.** `RFC_023` records a `COMPLETED`→`FAILED` change on
> this table drawing a REJECTED guardian veto **on scope**, and CLAUDE.md's 2026-07-28 ruling is that
> a scope veto is not answered by resubmitting with better measurements. This lane read that as
> covering both candidates. `[e24299]` narrowed it, correctly: **`RFC_023`'s vetoed change re-typed an
> existing status, altering its meaning for every consumer. Candidate 2 re-types nothing** — it fills
> a column that is currently always NULL on those rows (`0 of 3,086`, measured by that lane). So
> candidate 2 is genuinely contained and **candidate 1 is the architecture question** — which is §5's
> own split, but for a sharper reason than §5 gave.
>
> **The control replicated on disjoint traffic, which is stronger than a repeated count.** That lane
> independently re-measured §2(a)/(b): **36** COMPLETED runs at `complete_error` carrying
> `__step_error`, all `error` NULL, against **2 of ~3,020** clean COMPLETED runs ending there —
> against this file's 50 and 1 of 2,997. Different day, disjoint sample, same discrimination, so
> **the discriminator is a property of the system rather than of a window**. The counts differing is
> just §2(d)'s retention window moving.
>
> **A DEAD END, recorded so nobody re-walks it** (that lane's finding, and they asked for it to be
> written down): the tempting structural rule — *"an error terminal is one not reachable from the
> start on normal edges"* — **does not work**. `page-build-handler`, the biggest producer at 22 of
> 36, has a `check_page_found` `conditional` whose `else_step` is `complete_error` **directly**, so
> that terminal IS reachable on an ordinary edge and the rule misses **61% of the damage**. No purely
> structural rule can work, because a conditional's else-branch into a failure terminal is
> structurally identical to a success path. The redesign is around a **declared terminal outcome**
> instead.

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
