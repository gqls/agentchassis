# 205 — `extract_and_reconcile` has NO configured cap, runs at the hardcoded 2048, and two poisoned records have been re-dispatched every few minutes for 34 hours

**Filed:** 2026-08-06 · **Status:** OPEN · **Severity:** live and burning right now
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
- [INFERRED — NOT YET READ IN CODE] that the batch selects records by "not yet
  verified", so a record whose verification step can never succeed is re-selected
  every cycle, indefinitely. The evidence for it is strong but circumstantial: fresh
  correlation ids, a constant 5-minute-driven cadence, byte-identical prompts, and a
  second record taking over from the first at 00:34 on 08-06.
  **The next session must read `vet-batch-processor`'s selection step before
  asserting this.** If confirmed, the class is a *poison-pill loop*: a deterministic
  failure that a sweep re-drives for ever, with no attempt ceiling because each
  dispatch is a new attempt-1.

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
   the next oversized record loops again. Gated on part 3 being read in code.
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
