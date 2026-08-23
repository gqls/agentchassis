# 326 — a failed build can NEVER be retried: `create_work_item` dedups on `item_key` in ANY status, so re-submitting a domain reports COMPLETED and does nothing

**Filed 2026-08-19** by the `loanzy_uk_example_site` lane, from a greenfield one-shot build
watched end to end. **Status: OPEN, UNOWNED. Live. Customer-facing.**

> **On the 090 loop (owner ruling 2026-07-31):** not run, and this states why. The claim is not
> a structural theory — it is one action's own return value, read from the orchestration row,
> plus the absence of the row it should have created. Both are quoted below. Nothing here is
> inferred from a symptom.

> ## ⚠ CORRECTED 2026-08-23 — THE ROOT CAUSE IN THIS FILE'S TITLE AND BODY IS WRONG
>
> **`create_work_item` does NOT dedup on `item_key` in any status.** The live index
> excludes terminal ones, `complete` and `cancelled` among them:
>
> ```
> "idx_swi_dedup" UNIQUE, btree (site_id, item_key)
>   WHERE item_key IS NOT NULL
>     AND status <> ALL (ARRAY['complete','verified','rejected','wont_fix',
>                              'failed','unresolved','cancelled'])
> ```
>
> `writeWorkItem`'s `ON CONFLICT (site_id, item_key) … DO NOTHING` names exactly that
> predicate, so a completed predecessor **cannot** hold the dedup slot. The index arm is
> innocent and it works: all **36** dedup events retained in `orchestration_states` over
> 24h resolve as legitimate open-holder dedups when read with timestamps.
>
> **The real mechanism is the anti-churn brake at the TOP of `writeWorkItem`**
> (`platform/orchestration/actions/load_work_item_actions.go`), which counts
> `complete`/`failed` siblings on `(site_id, item_key)` inside 7 days and has two arms:
>
> - **within-cycle** — newest terminal sibling under **3.0 hours** old ⇒
>   `return workItemWrite{}, nil`: no row, no error, and `Inserted:false` is
>   byte-identical to a genuine dedup. **This is what bit this build.**
> - **two-strike** — ≥2 terminal siblings ⇒ status rewritten to `unresolved`, which is
>   terminal *and* not dispatchable, so the row is born dead and outside the dedup index.
>
> **The timing proves which arm fired.** `domain-submitter` writes a `submission` spec
> *before* the deduping step, so `site_specs` dates every submission even though
> `orchestration_states` reaped correlation `3296ac3a-…` long ago:
>
> | # | submission spec | outcome |
> |---|---|---|
> | 1 | 12:53:00.04Z | filed `research_loanzy.uk` at 12:53:00.25 → `complete` 13:36:36 |
> | 2 | **15:21:17.23Z** | **the deduped one** — no row |
> | 3 | 20:16:12.61Z | filed a new row at 20:16:13.85 |
>
> Submission 2 landed **2h28m** after the terminal sibling was **created**. The window is
> 3.0 hours. At 3h01m the insert would have succeeded and there would be no bug — so the
> measurement could have come out otherwise.
>
> **⚠ THE WINDOW KEYS ON `created_at`, NOT `completed_at`,** and the gap is not small: on
> `garden-tools.uk` (live greenfield, 2026-08-23) the research item was created 17:17:15Z
> and completed 17:44:59Z — **27m44s apart**, so its window closed at 20:17, not 20:45. An
> operator reasoning from when the stage *finished* gets the boundary wrong in the
> **unsafe** direction on any long-running item.
>
> ### What this means for the claims in this file
>
> - **"Consumed for ever" is FALSE.** The window expires after three hours. `[MEASURED]`
> - **The 78-row hand-rename of `item_key`s was very likely NOT what made submission 3
>   possible** — by 20:16 the sibling was 7.4h old and `complete` sits outside the index,
>   so waiting past 15:53Z would have done the same job. `[INFERRED — the counterfactual
>   is unmeasurable after the fact, because the rename removed the very rows it would be
>   measured against. Proven instead by test: a `complete` sibling older than the window
>   inserts.]` **Do not repeat the hand-rename; it is surgery for a three-hour timer.**
> - **The two-strike arm is not literally permanent either** — its counter reads only 7
>   days. What IS permanently lost is each request that arrived inside the poisoned
>   window, because an action request has no detector to re-file it. **635 of 747** live
>   `unresolved` rows carry the two-strike brand (2026-08-23); the largest populations are
>   `page_rerender` (212) and `improve_tool` (205), both action requests.
> - **Fix candidate 1 ("put the attempt in the key") is REJECTED.** It defeats the one arm
>   that works, which is this file's own required negative control.
> - **Fix candidate 2 is right about where to look but wrong about the target** — the
>   index already excludes terminal rows; it is the brake above it that needed aligning.
>
> ### What this actually is
>
> **`bugs_closed/024`'s defect class, recurring on a pipeline 024 never touched.**
> `work_item_recurrence_test.go` already says the brake *"is wrong for an ACTION REQUEST,
> where a terminal predecessor means the request SUCCEEDED"*, and `recurrenceExpected` is
> the proven opt-out that waives the heuristics **without** waiving dedup. 024 fixed the Go
> call sites it touched; the config-driven build chain reaches the same helper through
> `create_work_item` and nobody classified it. **19 of 21** keyed `create_work_item` steps
> had never declared either way (2026-08-23).
>
> **On the 090 loop:** filed, as the standing rule requires for a contradicted root cause
> (intake `df0b3d97-…`, run `655c8508-…`). It returned **5 bundles and no verdict** — the
> documented over-60KB shape (`load_work_item_actions.go` is 82,669 bytes), with
> `doc_notes` taking 21 rows in the same window as the demand control. First-hand
> verification substituted and stated, per the ruling's escape hatch. Full evidence and
> every query: `docs/agent_docs/docs024_key_docs_latest/bugfix_326_retry_the_front_door/NOTES_326_retry_the_front_door.md`.
>
> ### ✅ PROVEN LIVE 2026-08-23 19:23Z — a re-submission INSIDE the window queued work
>
> The fix is verified at the artefact on a real build, by two observers, with the outcome
> meanings agreed **before** the result existed.
>
> `garden-tools.uk` was a greenfield build that died at its second hop (`bugs_open/376`, an
> unrelated `on_error` gap). The natural operator response — re-submit — is exactly the case
> this bug is about, and it landed **2h05m51s** after the terminal sibling was created, well
> inside the 3.0h brake that would have eaten it that morning:
>
> ```
> 07b589a9-025f-4cae-a454-809ddf4584f5 | research_garden-tools.uk | complete | 17:17:15.482481+00
> 3921bde4-968e-464d-8c2f-f682f495edf4 | research_garden-tools.uk | triaged  | 19:23:06.330863+00
> ```
>
> **A distinct id, exactly as this file's "How to verify a fix" section demands** — asserted at
> the row, never at the orchestration status. Re-queried independently minutes later, the new
> row had moved `triaged` → **`claimed`**: not merely filed but **dispatched**, which is the
> stronger statement. `retry_after` is NULL on both rows, correctly — the deferral change was
> vetoed and is not shipped, so what is proven here is the classification and nothing else.
>
> **Attribution control:** `recurrence_expected` re-read on the live definition at the same
> moment and still `true`.
>
> > **CORRECTED 2026-08-23, hours after writing it** — this control was first described here as
> > ruling out "the window simply having elapsed". **It does not, and that rival was never
> > live:** the threshold is 3.0h and the offset was 2h05m51s, so elapse is excluded by the
> > arithmetic alone, by 54 minutes. What the control actually rules out is that something
> > OTHER than the declaration stopped the brake biting — the threshold retuned, the block
> > disabled, or 572 rolled back between the two snapshots. All of those are consistent with an
> > inside-window insert; the live flag is what picks this fix out of that set. So it
> > establishes **attribution, not elapse**. Caught by the loanzy.uk lane: as first written, the
> > first reader to do the subtraction would find a rival that was never possible and trust the
> > rest of the block less.
>
> **Two caveats recorded by the lane that took the measurement, because a clean result is when
> to check the instrument:**
> - The `claimed`-items snapshot (0 before and after) was an **unused control**. It would only
>   have mattered on a null result; it corroborates nothing here.
> - **The key was genuinely free.** `needs_vertical_research` reached `failed` at 19:22:13Z,
>   ~40s before the test. Had the re-submission run while that item was still `triaged` —
>   non-terminal, and therefore inside `idx_swi_dedup` — the classifier's own `create_next_item`
>   would have conflicted and produced a **false negative on the fix's first live test**. My own
>   earlier instruction ("re-submit whenever, whatever the offset") would have caused exactly
>   that; the lane's own timing discipline is what avoided it.
>
> **Incidental finding, worth knowing if anything reads `aspect='submission'`:** a re-submission
> is *not* inert on the specs. The prior `submission` spec was superseded (`is_current` t→f) and
> a second written, so that aspect can hold more than one row per site.
>
> ### STATUS 2026-08-23: the customer path is FIXED AND LIVE; the framework half is vetoed and routed
>
> **Migration 572 is applied.** The five build-chain handoffs now declare
> `recurrence_expected: true`, so a re-submission after a finished build queues work instead of
> vanishing. Verified at the artefact: the new census (`scripts/audit-undeclared-recurrence.sh`)
> went 19 → 14 findings and **no build-chain step is undeclared any more**. The negative control
> is unaffected and enforced by the database, not by config — `idx_swi_dedup` still refuses a
> second OPEN item, so two simultaneous submissions of one domain still produce one build.
>
> **The wider fix — making the brake DEFER rather than destroy, for all 36 Go call sites and the
> 14 still-undeclared config steps — was REJECTED by the council gate on a guardian hard veto**
> (corr `f610741f-5054-41e8-b0b7-54915d79ba92`), on the ground that it is a fleet-wide
> architecture change bundled into an urgent point fix, and that 572 alone closes this bug. Both
> points are right. It is routed to
> `docs/agent_docs/docs024_key_docs_latest/architecture_review/RFC_048_the_anti_churn_brake_may_delay_work_but_may_not_destroy_it.md`
> with the patch beside it and three options costed. **Not resubmitted** — a scope veto is not
> answered with better measurements.
>
> **So this bug stays OPEN**, and what is still open is precisely: an unclassified caller's
> request can still be destroyed silently. `bugs_open/327` (the trigger's own silence) and
> migration 573 (`_HOLD`, the loud front door) are the other two residuals.
>
> *Caught by the bugs_open/326 fix lane. The loanzy.uk lane, which filed this, has recorded
> the matching correction in its own NOTES and runbook rather than having it forked here.*

## The one-paragraph version

`082_submit_domain_unified.sh <domain>` on a domain that has been submitted before produces a
**COMPLETED** orchestration and **no work at all**. `domain-submitter`'s `create_research_item`
step matches the previous attempt's item by `item_key` — **regardless of that item's status,
including `complete` and `cancelled`** — and returns `deduped`. Nothing is queued, nothing runs,
and the operator sees a success. Every pipeline stage key (`research_<domain>`,
`strategy_<domain>`, `briefing_<domain>`, `site_plan_<domain>`) is consumed for ever by the
first attempt, so **a build that fails halfway can never be re-run through the front door.**

## Evidence (measured 2026-08-18)

Re-submission of `loanzy.uk`, correlation `3296ac3a-30bd-4db5-84ae-16260394f3bc`:

```
orchestration_states: status=COMPLETED  current_step=complete   error=NULL
collected_data->'research_item':
{"deduped": true, "inserted": false, "item_key": "research_loanzy.uk",
 "item_type": "needs_domain_research", "site_id": "55213ded-…", "born_blocked": false,
 "handler_agent": "domain-research-classifier"}
```

`site_work_items` gained **no row**. The only artefact of the whole run was a new `submission`
spec written by the submitter before the dedup.

**The prior item was `complete`, not open** — i.e. dedup is not protecting against a duplicate
in flight, which would be legitimate. It is refusing because the domain was *ever* researched.

**What made a second run possible at all**, and it is not a fix: renaming the previous run's
keys by hand — `UPDATE site_work_items SET item_key = item_key || '_run1' WHERE site_id = …`,
**78 rows**. After that, the identical trigger filed `needs_domain_research` /
`research_loanzy.uk` / `triaged` immediately.

## Why this is worse than an inconvenience

1. **It is the customer path.** webdesign.uk sells a one-shot build. When a build fails — and
   `bugs_open/260` and `bugs_open/311` both cost pages in the same run — the natural operator
   response is to re-submit. That reports success and does nothing.
2. **The failure is silent in the only place anyone looks.** `COMPLETED` at step `complete`,
   no error. The `deduped: true` flag exists but is buried in `collected_data`, and no trigger,
   log line or item surfaces it.
3. **It compounds with a stalled or partial build.** A site left half-built cannot be resumed by
   the same route that created it; it needs hand surgery on `item_key`s by someone who knows
   this bug exists.

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **Put the attempt in the key.** `research_<domain>_<attempt|submission_id>` — a re-submission
   is then a new unit of work *by construction*, and no dedup predicate has to be correct for
   the retry to work. This is the only candidate where "retry is impossible" stops being
   representable.
2. **Dedup only against non-terminal rows.** `idx_swi_dedup`'s partial-index contract already
   implies terminal statuses are not duplicates (see the "dedup index ↔ Go list lockstep"
   landmine — `workItemTerminalStatuses` is the same contract). Align `create_work_item` with it.
   ⚠ Check the lockstep before changing either half; drift there is a fleet-wide `42P10`.
3. **Make the silence loud.** Whatever the keying, a submission that queues nothing must not
   report COMPLETED: fail the trigger, or file a visible item saying the domain is already built
   and naming what to do instead. This is the cheap half and it can ship independently.
4. **Give the route an explicit rebuild verb** — a supported "rebuild this site from scratch"
   that supersedes prior stage items rather than colliding with them. Today that operation
   exists only as undocumented hand-surgery.

## How to verify a fix

Submit a domain, let it complete, submit it again, and assert a **new** `needs_domain_research`
row exists with a distinct id — at the row, never at the orchestration status, which says
COMPLETED either way. Include a **negative control**: a genuine duplicate submitted while the
first is still `triaged` must still dedup, or candidate 2 has simply removed the protection.
