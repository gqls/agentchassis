# 326 — a failed build can NEVER be retried: `create_work_item` dedups on `item_key` in ANY status, so re-submitting a domain reports COMPLETED and does nothing

**Filed 2026-08-19** by the `loanzy_uk_example_site` lane, from a greenfield one-shot build
watched end to end. **Status: OPEN, UNOWNED. Live. Customer-facing.**

> **On the 090 loop (owner ruling 2026-07-31):** not run, and this states why. The claim is not
> a structural theory — it is one action's own return value, read from the orchestration row,
> plus the absence of the row it should have created. Both are quoted below. Nothing here is
> inferred from a symptom.

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
