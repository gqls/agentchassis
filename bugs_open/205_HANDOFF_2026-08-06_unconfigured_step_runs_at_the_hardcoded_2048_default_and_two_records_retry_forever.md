# 205 — `extract_and_reconcile` has NO configured cap, runs at the hardcoded 2048, and two poisoned records have been re-dispatched every few minutes for 34 hours

**Filed:** 2026-08-06 · **Status:** CLOSED in substance 2026-08-08 — fixed AND live, all four owner decisions executed; stays in `bugs_open/` per the owner's 08-06 ruling · **Severity:** ~~live and burning right now~~ burn stopped 2026-08-07 01:40Z; poisoned record verified 2026-08-08

> **FIX STATUS 2026-08-07 ~01:45 UTC (bugfix-205 session):**
> - **Config half LIVE and BEHAVIOURALLY PROVEN:** the reaper's `reset_tasks` CTE
>   now counts each reset in `retry_count`, backs off via `scheduled_for`, and
>   PARKS a task as `'failed'` on what would be the 5th reset. Applied 01:26 UTC
>   with a row backup; the 33 measured loopers seeded to `retry_count=4` and
>   **all 33 parked in one reaper pass at 01:40:48Z**. Table now failed=33 /
>   pending=0 / in_progress=0; zero LLM calls on the step since; both scheduled
>   tasks' stamps still advance (quiet-because-parked, not dead). Apply script +
>   backup + ROLLBACK + runbook (incl. un-parking):
>   `docs024_key_docs_latest/bugfix_205_poison_pill_reaper/`.
> - **OWNER NOTE (from the council's guardian seat, and it is right):** the reset
>   arm is generic over `collection_tasks`, so a FUTURE task_type would inherit
>   park-at-5 silently. Today the census says the whole table is ONE task_type in
>   ONE vertical (`initial_verification`/`veterinary`, 3,134 rows all-time), so
>   the population of surprised consumers is empty — but whoever adds a second
>   task_type should choose its ceiling deliberately.
> - **Go half committed `d1eb3a6b5`, INERT until the next chassis roll** (the
>   08-06 build predates it): `ensure_collection_tasks.go` refuses to re-task a
>   parked business; `ai_actions.go` WARNs when `max_tokens` resolves to the
>   hardcoded transport fallback (candidate 1). Verify at the pod with
>   `strings /app/agent-chassis | grep -c "max_tokens not configured at any level"`.
> - **Council:** `Council-Submitted: 2db88f8f-11ea-47ed-b37d-35a6096d5597`
>   (round 1 `complete_invalid` — schema, not substance; resubmitted same corr).
> - ~~**Candidate 2 (the step's own cap) remains an OWNER CALL, deliberately not
>   taken here.** Parking stops the burn; the parked task is the prompt.~~
>   **ALL FOUR OWNER DECISIONS RULED AND EXECUTED 2026-08-08:** (1) cap set to
>   **8000** (nested-path write, type-verified) — the poisoned record then
>   verified on its FIRST attempt (`output_tokens=3135`; the 2048 fallback could
>   never have passed it); task `completed`, business `verified`. (2) The other
>   32 parked tasks **cancelled** (RETURNING all ids, message dated). (3+4)
>   **Per-task_type park ceilings + the shared reaper-accounting mechanism are
>   live**: `reaper_policies` + `business_intel.reap_stale_collection_tasks()`,
>   migration `sql_for_agents/335`, contract + migration-invitation in
>   `architecture_review/RFC_018`, register SCH-024, induced-test proven.
>   **This bug is CLOSED in substance** (stays in `bugs_open/` per the owner's
>   08-06 filing ruling); the only live watch-item left is the framework WARN
>   announcing whichever of the 7 remaining uncapped steps runs first.
>   **CORRECTED 2026-08-08 evening: it already HAS — and the announcement
>   expired unread.** `med-price-collector/scrape_prices` (Ollama-backed, no cap
>   at any config level) ran 7× on 08-07 15:14–21:23Z; the WARN fired into pod
>   logs that two subsequent fleet rolls have since destroyed, so every later
>   log-grep honestly read 0. Caught by `llm_call_log`: an uncapped Ollama call
>   logs `max_tokens NULL` (an uncapped Anthropic call logs `2048` — the
>   transport fallback; 112 pre-fix verifier rows prove the shape). **The
>   durable watch is the DB, not the logs**: `WHERE max_tokens = 2048 OR
>   max_tokens IS NULL`, cross-checking any 2048 hit against config. All 7
>   calls succeeded (local model, no paid spend, no truncation observed) — the
>   step still needs a deliberately chosen cap, which is the med-price lane's /
>   owner's call. 6 uncapped steps remain unheard-from. Full account:
>   WRONG_CALLS 2026-08-08 + lane NOTES.
>   **RE-CORRECTED 2026-08-09 — the 08-08 correction was itself wrong about the
>   mechanism.** `scrape_prices` never passes the WARN site: its action
>   (`vet_med_price_scrape_action.go`, `llmExtractPriceVariants`) calls Ollama
>   directly and hardcodes `num_predict: 500` — it is capped in CODE, the
>   tightest in the estate (max output ever observed ≈150–200 tokens, ~3×
>   headroom). The NULL `max_tokens` rows are that action's logging omission,
>   not uncapped calls, and **the WARN has in fact never fired.** What was
>   missed: this file's own census DEFINES the WARN's population, and
>   `scrape_prices` was never in it. WRONG_CALLS 2026-08-09 has the lesson.
>   **CLOSING THE CLASS, 2026-08-09 (owner decision):** the remaining 7
>   uncapped step-rows (6 distinct steps, all Anthropic) got explicit caps via
>   migration `sql_for_agents/347` — 32000 site-architect/design (design class
>   measured max 20,189), 16000 chief-strategist/generate_build_plan +
>   content-creator/create_content ×2, 8000 brand-designer/analyze_brand,
>   domain-analyst/analyze, provocation-gate-calibration/gate (sonnet-5 thinks
>   into output_tokens — never a small cap). Guards asserted the pre-existing
>   chief-strategist 8192 untouched and the fleet census EMPTY. **No active
>   LLM step can now fall to the 2048 fallback; the WARN's remaining job is
>   announcing any FUTURE step added uncapped.**
**Found by:** the first run of `fleet-step-token-pressure` (bugs_open/183 candidate 4).
It was the top line of the check's first note — 64 truncations, the largest in the
fleet — and nothing else was watching it. See
`docs024_key_docs_latest/bugfix_183_step_token_pressure/`.

## Symptom

`vet-practice-verifier` / `extract_and_reconcile` is failing **100% of calls** and has
been since 2026-08-05 ~03:00 UTC:

```
response truncated: stop_reason=max_tokens
(output_tokens=2048 reached the configured cap, 5795 chars recovered)
```

[MEASURED 2026-08-06 ~10:40 UTC] 19 of 19 calls on 08-06 truncated; 45 of 62 on 08-05;
0 of 54 on 08-04. Roughly 2 calls/hour, continuously, all failing.

## Root cause part 1 — the cap is nobody's decision

The step definition sets **no `max_tokens` at any level**. Read live:

```sql
SELECT jsonb_pretty(default_config #> '{workflow,steps,extract_and_reconcile}'
                    #- '{config,prompt_template}')
  FROM agent_definitions WHERE type='vet-practice-verifier' AND is_active
    AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- config.ai_service = {model: claude-haiku-4-5, provider: anthropic, api_key_env_var: ...}
-- no max_tokens. And no root ai_service block either (so bugs_open/009 shadowing is NOT
-- the mechanism here — there is simply nothing to shadow).
```

So the resolution chain (LCO-002) falls all the way through to the **hardcoded
platform default at `platform/aiservice/anthropic.go:109` — `"max_tokens": 2048`**.

This is a distinct and worse case than `bugs_open/183`'s. 183's step had a cap that
was *chosen once and outgrown*. This step's cap was **never chosen by anyone**: it is
a transport-layer fallback that happens to be the smallest number in the estate, and
nothing in the definition tells a reader it is in force.

## Root cause part 2 — it is TWO poisoned records, not a distribution drift

This is the part a headroom number alone gets wrong, and it changes the fix. All 64
failures come from **two byte-identical prompts**:

```sql
SELECT md5(prompt_rendered) AS prompt_md5, count(*), min(created_at), max(created_at),
       count(*) FILTER (WHERE success) AS ok
  FROM llm_call_log WHERE step_name='extract_and_reconcile' AND created_at > '2026-08-05'
 GROUP BY 1 ORDER BY 2 DESC;
--  33749de2… |  46 | 08-05 00:30 | 08-06 00:08 | 0
--  105ca46f… |  18 | 08-06 00:34 | 08-06 09:38 | 0   <- still running
--  every other md5 |  1-4 |  … | … | all succeeded
```

Successful calls emit **468–639 output tokens** — a quarter of the cap. The two
failing records emit >2048 (≈5,790 chars recovered before the cut). So the step is
not near its ceiling in general; two specific inputs produce a 4× larger document and
fail **deterministically**, every time, forever.

**Each failure carries a DISTINCT `correlation_id`** — so this is not one
orchestration retrying inside `max_attempts`. It is a fresh dispatch each time.

## Root cause part 3 — the loop, and what is verified vs inferred

- [VERIFIED, live] `scheduled_tasks` row **`vet-batch-verify`**: `enabled=true`,
  `interval_seconds=300`, `target_agent_type='vet-batch-processor'`,
  `fire_message=true`, last triggered minutes before this file was written.
- ~~[INFERRED — NOT YET READ IN CODE] that the batch selects records by "not yet
  verified", so a record whose verification step can never succeed is re-selected
  every cycle, indefinitely.~~

> **CORRECTED 2026-08-06 (bugfix-205 session) — the inference above was WRONG in
> detail; the loop is real but the door is elsewhere. Read in code and config:**
> - The batch honestly selects **`status='pending'` only** and claims to
>   `in_progress` (`LoadBusinessBatchAction`, `business_intel_actions.go:557-598`).
>   Success writes `completed`; **failure writes NOTHING** — a whole-repo grep finds
>   no Go writer of `pending`.
> - The `pending` is manufactured by the **`stale-orchestration-reaper`** scheduled
>   task (every 180 s): its `pre_query` carries a `reset_tasks` CTE —
>   `UPDATE business_intel.collection_tasks SET status='pending', started_at=NULL,
>   orchestration_id=NULL WHERE status='in_progress' AND started_at < NOW() -
>   INTERVAL '20 minutes'` — **unconditional; no retry ceiling, no backoff**. The
>   table's `retry_count` column is written by nothing (still 0 on the poisoned task
>   after ~50 dispatches). No repo seed defines this row; the live row is the source.
> - Loop period = 20 min reap + ≤5 min schedule + batch time — matches the observed
>   25–55 min between byte-identical failures, each under a fresh
>   `vet-batch-processor` parent (verified by joining `llm_call_log` →
>   `orchestration_states`: all 08-06 truncations are task `ea489aed-…`, business
>   `926410bc-…`, every parent distinct).
> - **[MEASURED 2026-08-06 ~20:30 UTC] The class is 33 tasks, not 2 records:**
>   `vet-practice-verifier` ran **1,576 times in 24 h, 1,575 FAILED, 33 distinct
>   task_ids** (~50 dispatches per task per day). Only ONE reaches
>   `extract_and_reconcile` and burns an LLM call; the other 32 die earlier at
>   `scrape_website` on external API/URL errors — **invisible to the token-pressure
>   check that found this bug**, which watches only LLM calls.
> - The second poisoned prompt (`33749de2…`) no longer appears in the window; only
>   `105ca46f…` loops as of 08-06 evening.
>
> So the class is confirmed as a *poison-pill loop*, with the sharper statement:
> **a reaper that resurrects stale claims without counting resurrections converts
> any deterministic failure into an infinite loop.** 016b §9 already documents the
> sibling defect (`stale-work-item-reaper`, 2026-07-25) and prescribes
> attempt-counting — this is that pattern's second instance.
> Fix in flight: `docs024_key_docs_latest/bugfix_205_poison_pill_reaper/PLAN_2026-08-06_poison_pill_reaper.md`.

## Fix candidates, ordered by what closes the door

1. **Stop unconfigured steps inheriting a silent 2048.** The transport default exists
   for a reason, but a *workflow step* reaching it means nobody sized that step. Make
   it visible rather than silent — log at WARN when `max_tokens` resolves to the
   hardcoded fallback, and surface it in the `fleet-step-token-pressure` note. This
   is the framework-level fix and it generalises past this agent.

   **[MEASURED 2026-08-06] The census is run, and it bounds the work.** Over every
   active, non-snapshot definition, counting steps whose `config` carries an
   `ai_service` block:

   ```sql
   WITH steps AS (
     SELECT ad.type, s.key AS step_name,
            (s.value->'config'->'ai_service'->>'max_tokens') AS step_cap,
            (ad.default_config #>> '{ai_service,max_tokens}') AS root_cap
       FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s
      WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
        AND s.value->'config' ? 'ai_service')
   SELECT count(*) AS llm_steps, count(*) FILTER (WHERE step_cap IS NOT NULL) AS has_step_cap,
          count(*) FILTER (WHERE step_cap IS NULL AND root_cap IS NOT NULL) AS falls_to_root,
          count(*) FILTER (WHERE step_cap IS NULL AND root_cap IS NULL) AS falls_to_hardcoded
     FROM steps;
   -- 126 | 114 | 4 | 8
   ```

   **8 of 126 steps set no cap at any level.** Only **two** have run in 90 days:
   this one (observed cap 2048 — the fallback confirmed from the call log, not
   inferred) and `site-architect/design`. The other **six are dormant, and that is
   the point of candidate 1**: they are latent, and the first real workload through
   any of them meets a 2048 ceiling nobody chose. That is precisely how this bug
   presented — 54 clean calls on 08-04, then two larger records.

   ⚠ **Caveat on the observed-cap column, stated because it bit this very query:**
   `llm_call_log` has no definition id, so joining observations by `step_name` alone
   conflates agents that share a step name. It is safe for `extract_and_reconcile`
   (only `vet-practice-verifier` defines it) and NOT safe for `design`, which several
   agents define — `site-architect/design`'s observed 32000 is very likely another
   agent's rows and must not be read as "this step is fine".
2. **Give this step an explicit cap.** ~8000 matches the fleet mode and is ~4× the
   observed success distribution. **Owner call**: every cap raise on this estate so
   far has been one, and `vet-practice-verifier` belongs to the vet-intel lane, not
   to the 183 lane that found this. Deliberately NOT done here.
3. **Break the poison-pill loop** — a record that has failed verification N times
   must be parked, not re-selected. Without this, candidate 2 only moves the cliff:
   the next oversized record loops again. ~~Gated on part 3 being read in code.~~
   **UN-GATED 2026-08-06: part 3 is read (see the correction above), and this is now
   the PRIMARY fix** — the reaper's `reset_tasks` CTE parks at `retry_count` 5
   instead of resetting for ever, plus `ensure_collection_tasks.go` treats a parked
   task as blocking re-creation. Plan and verification:
   `docs024_key_docs_latest/bugfix_205_poison_pill_reaper/`.
4. Weakest: `tolerate_truncation` on the step. **Probably wrong here** for 183's own
   reason — a salvaged partial ends at the last complete value, so trailing fields go
   silently ABSENT and the record would be reconciled from a fragment and marked
   done. Failing loudly beats writing a half-record into a practice directory.

## How to verify a fix

```sql
-- the loop has stopped: no new failures for this step after the fix time
SELECT date_trunc('hour',created_at) AS hr, count(*),
       count(*) FILTER (WHERE success) AS ok
  FROM llm_call_log WHERE step_name='extract_and_reconcile'
   AND created_at > now() - interval '6 hours' GROUP BY 1 ORDER BY 1 DESC;
-- and the two poisoned records specifically now succeed (or are parked):
SELECT md5(prompt_rendered), count(*) FILTER (WHERE success) AS ok, count(*)
  FROM llm_call_log WHERE step_name='extract_and_reconcile'
   AND created_at > now() - interval '2 hours' GROUP BY 1;
```

⚠ A quiet hour is NOT proof — the batch fires every 5 minutes, so absence of calls
means the *sweep* stopped, which is a different (and also bad) finding. Require
non-zero calls WITH `ok = count(*)`.

## Diagnosis loop (owner ruling 2026-07-31 compliance)

**Not run through `090` for parts 1 and 2**, per the ruling's named escape hatch, and
this states why: both are local and self-evidencing. The cap's absence is one jsonb
read against the live row; the hardcoded 2048 is one grep with a file:line; the
"two identical prompts" finding is a whole-population `md5()` group-by over every call
since 08-05, not a sample, and it could have come out otherwise (the competing
hypothesis — a general distribution drift — predicts many distinct prompts near the
cap, and is refuted by the same query that establishes the finding).

**Part 3 is explicitly NOT claimed as diagnosed.** It is marked [INFERRED] above with
the exact unread hop named. Whoever takes this should either read
`vet-batch-processor`'s selection step or file `090` on the loop specifically —
"a scheduled sweep re-dispatches a deterministically-failing record indefinitely" is
exactly the cross-cutting shape the loop is for, and it is very unlikely to be the
only sweep that does it.

## Related

- `bugs_open/183` — same family (a cap too small for the document its step emits),
  and the check filed under 183's candidate 4 is what found this. 183's step at least
  HAD a cap; this one never did.
- `bugs_closed/067` — "when a cap defect is found on ONE step, sweep EVERY step of
  that agent". Not yet done for `vet-practice-verifier`.
- `platform/aiservice/truncation.go:26-29` — the platform's own warning that raising
  a cap moves the cliff rather than closing the class. It applies to candidate 2.
