# HANDOFF — bugs_open/138, continue here (2026-07-31 evening)

Cold-start doc for this lane. Read this, then `NOTES_degraded_gates.md` (newest at the
bottom) and `RUNBOOK_degraded_gates.md` §7–10 before touching anything.

---

## The bug in one paragraph

A council reviewer whose LLM reply exceeds `max_tokens` is recovered as a fragment,
marked `degraded`, and a degraded `object` **gates its round to REVISE regardless of the
severities that survived** — correctly, since a high objection may have been in the lost
tail. But the record then said `gating objection from <seat>`, as though the seat had
judged. It hadn't; it ran out of room. And a high object rate is also the documented
kill-switch for retiring a noisy seat. **So the failure hides inside its own evidence:
the fix is to give the seat room, and the invitation is to sack it.**

## State: all four candidates answered

| candidate | state | where |
|---|---|---|
| 1 — say WHY a round gated | **LIVE, pod-verified twice** (including surviving a fresh build) | FIX-055, commit `3a59b5012` |
| 2 — alert on the rate | **DONE, running, self-sourcing targets** | FIX-058, `104_REPORT_*` + `104_TASK_*` |
| 3 — right-size caps | **DONE** — owner closed the last divergence 07-31 | `sql_for_agents/277` |
| 4 — schema order / length | **reorder REFUTED; length budget applied to 10 seats and VERIFIED** | FIX-059, `scripts/apply-seat-length-budget.py` |

**The bug stays OPEN** for one reason only: candidate 1's `true` branch has never fired
in production. No post-roll round has been gated *by a truncation*, so the TRUNCATED
wording is proven by unit test and by the persistence path, not by a live instance.

## What is actually owed

1. **Catch the `true` branch when it happens.** Do not induce it with an artificial
   round; watch for it:
   ```sql
   SELECT created_at, body::jsonb->>'decided_by' FROM diagnosis_artifacts
   WHERE kind='council_report' AND metadata->>'gated_by_truncation'='true'
   ORDER BY created_at DESC LIMIT 5;
   ```
   When it fires, the wording is confirmed and **138 can close**.

2. **Let the other three seats accumulate a control arm.** `guardian`,
   `improvement_guardian` and `debug_historian` got the length budget at cutover
   `15:12:49`–`15:12:58` but had **zero** pre-cutover calls in the window, so there is no
   before/after comparison for them yet. Re-run RUNBOOK §10's spawn-time query in a few
   days. **Do NOT compare their post-cutover max against their 14-day historical max** —
   a maximum grows with the sample, and that comparison is a trap this lane has already
   documented.

3. ~~Two owner decisions~~ **BOTH DECIDED AND APPLIED, 2026-07-31 ~18:20.** Kept here
   because what they were is still the fastest way to understand the shape of the fix:
   - `feature-designer/review_architecture` now emits `notes` third, matching its two
     siblings (FIX-060). All three architecture seats report an identical key order.
   - The length budget now covers **48 of 51** seats. 2 refused (hand-authored blocks),
     1 excluded with a printed reason. 099 drift none.
   **What this creates, and it is new:** 47 seats changed behaviour with no per-seat
   evidence. Most were never near their cap, so expect no visible change in peaks. **The
   thing to watch is whether OBJECTION COUNTS fall** — that would mean coverage traded
   for brevity, which is the failure this rollout can cause and the narrow one could not.
   Query in `NOTES`, dated 2026-07-31 ~18:20; fleet cutover `18:16:46`–`18:16:53`.
   Compare rounds either side **by spawn time**.

## The result worth carrying forward

**The length budget works, and it works alone.**

| `review_editquality` @16000 | calls | peak | % of cap | mean |
|---|---|---|---|---|
| rounds spawned **BEFORE** the block | 10 | 15,721 | **98.3%** | 9,848 |
| rounds spawned **AFTER** the block | 8 | 8,793 | **55.0%** | 6,569 |

Same seat, same afternoon, **cap unchanged** (it was already 16000). That separates the
two halves `review_architecture` could not — there the cap raise and the budget shipped
together and nobody could say which mattered. The small-sample objection cuts the wrong
way: a maximum grows with n, and the arm with *more* calls has the *higher* peak.

**And the companion finding: a raised cap does not stay roomy.** `review_editquality`
grew into its doubled 16000 cap in three days with no prompt change — 13,115 → 13,592 →
15,721 tokens — so the 07-28 raise bought it about three days. Read every cap raise that
way, including `277`'s: time, not immunity. This is the strongest evidence the lane
produced for its own central claim, and it is clean of the confound the first version had.

## The instruments, and how to read them honestly

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/104_REPORT_seat_token_pressure_v1.sh [days]
./scripts/apply-seat-length-budget.py [--verify|--apply]
```
```sql
SELECT created_at, body FROM doc_notes WHERE categories ? 'seat-token-pressure'
ORDER BY created_at DESC LIMIT 3;                    -- the automatic half
SELECT pre_query FROM scheduled_tasks WHERE name='council-seat-token-pressure';  -- the ONLY copy of the thresholds
```

- The alert is an **event, not a heartbeat** — `subject_key` is an md5 of the flagged
  set, so a persisting condition speaks once. **Silence from a task that stopped running
  looks exactly like silence from a healthy fleet**: check `last_triggered_at` before
  trusting a quiet week.
- It flags on **peak ≥ 95%** (near-miss; truncation is a tail event) and **p95 ≥ 85%**
  (routine pressure). Keeping them separate is load-bearing — it found
  `review_debug_historian` at peak 99.8% with a p95 of **62.2%**, which a p95 rule could
  never see.
- **Read `n_holder` against `n` before acting.** Three of the original five flags had
  `n_holder ≤ 5`: inferences from a sibling council at the same cap, not measurements of
  the holder.

## Five traps this lane paid for. Read these before measuring anything here.

1. **`llm_call_log.agent_type` was relabelled 2026-07-26** — 1,836 rows say `generic`,
   and **`fix-proposer` has never appeared at all**. A 14-day per-agent figure is
   computed from four days. Key on `step_name`. (LANDMINES, footprint `llm_call_log`.)
2. **The denominator changes inside the window.** `max_tokens` is per call and caps get
   raised, so compute the ratio **per row** and filter to the current cap. The mixed
   version renders editquality as "95% of 16000" when those rows peak at 62.9%.
3. **Config cutover time ≠ a run's effective time.** An orchestration keeps the workflow
   it loaded **at spawn**, so every before/after measurement needs the
   `orchestration_states.created_at` join, not `llm_call_log.created_at`. I got this
   wrong today *with the warning written in my own runbook*.
4. **`prompt_template` and `max_tokens` sit at different depths** — `config.prompt_template`
   vs `config.ai_service.max_tokens` — and neither wrong path errors; both return a clean,
   uniform, false answer for every seat.
5. **`snapshot_agent()` writes to `agent_definitions_backup`**, not an `is_snapshot` row
   in `agent_definitions`. The obvious check returns 0 after three successful snapshots.
   And assert the backup does **not** contain your change, or it is a souvenir rather
   than a rollback.

## What I got wrong, since it is the cheapest thing to inherit

Three checks in one day answered a different question from the one I meant to ask: a
mixed-denominator ratio that made a working cap raise look useless; the wrong JSON depth;
and a retyped grep pattern that missed a backtick, made a present landmine read as
absent, and led me to commit a duplicate plus a false claim (withdrawn in `bcecb65e3`).

The pattern is not carelessness. **Each wrong answer fitted something I already had good
reason to believe** — the file really was being swept by other sessions; a cap raise
really does move the cliff. The checks that get through are the ones you have a reason to
believe, so the discipline is to re-derive the result that confirms your argument *first*,
not last. Logged in `WRONG_CALLS.md`.

## Related, deliberately not touched

- `experience-approval-council` holds `review_prior_art` at 8000 (16000 on the fix lane)
  and `review_honesty` at 8000 (16000 on experience-planner). Flagged as information;
  they belong to the experience-loop workstream. **Tell that lane rather than editing
  their council.**
- `bugs_open/119` is the sibling — `unreadable` (malformed JSON, round voided), not
  `degraded` (well-formed but incomplete, round *decided*). Different field, different
  `decided_by`, different fix.
