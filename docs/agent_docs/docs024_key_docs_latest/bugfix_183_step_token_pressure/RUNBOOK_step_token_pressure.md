# RUNBOOK — fleet-step-token-pressure (bugs_open/183 candidate 4, generalised)

All DB access:
`kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`

## Apply / remove the task

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db < SQL_2026-08-06_fleet_step_token_pressure_task.sql
# remove:
#   DELETE FROM scheduled_tasks WHERE name='fleet-step-token-pressure';
```

## Is it running? (a missing note is NOT "nothing is wrong")

```sql
SELECT name, enabled, interval_seconds, last_triggered_at, last_completed_at
  FROM scheduled_tasks WHERE name='fleet-step-token-pressure';
```

Gotcha: the schema has no `last_run_at`/`next_run` — those column names fail.
`last_triggered_at` stale by more than ~2 intervals means the SCHEDULER is not
running the task, which is a different finding from a quiet fleet. The task only
writes a note when the flagged set CHANGES (digest dedup, 30-day look-back), so
"no new note" with a fresh `last_triggered_at` is healthy.

## Read the findings

```sql
SELECT created_at, body FROM doc_notes
 WHERE subject_type='pipeline' AND subject_key LIKE 'fleet-step-token-pressure:%'
 ORDER BY created_at DESC LIMIT 3;
```

## Reading a finding: the KIND decides the fix, and two of them are opposites

| kind | means | the fix |
|---|---|---|
| **C** | died on the **clock**, not the cap | **Do NOT raise the cap** — it is already unreachable. Stream, shrink the unit, or use a faster model. |
| **T** | has truncated **at** this cap | Raise the cap, or shrink the unit. Check the SHAPE first (below). |
| **N** | near-miss, peak ≥ 95% of cap | Watch, or pre-emptively raise. |
| **P** | pressure, p95 ≥ 85% of cap | The body of the distribution is near the ceiling. |

**Before acting on a T, check the shape, not the count** — a big truncation number can
be one stuck record on a retry loop rather than cap drift, and those want opposite fixes
(`bugs_open/205` is the worked case):

```sql
SELECT md5(prompt_rendered), count(*), count(*) FILTER (WHERE success) AS ok
  FROM llm_call_log WHERE step_name='X' AND created_at > now() - interval '7 days'
 GROUP BY 1 ORDER BY 2 DESC;
-- many distinct prompts near the cap = genuine cap drift
-- one or two prompts repeating   = a stuck item; a bigger cap treats the symptom
```

**The clock ceiling in tokens.** Output runs ~98 tok/s on Sonnet 5 and 47–82 on Sonnet
4.6 against a 600s non-streaming HTTP timeout (`aiservice/anthropic.go:72`; ollama is
120s at `ollama.go:55`). So the reachable ceiling is ~58,000 tokens on Sonnet 5 and
~28,000–42,000 on Sonnet 4.6 — **a `max_tokens` above that cannot be reached whatever
the config says.**

## Re-run the measurement by hand (thresholds live in the pre_query, not here)

```sql
SELECT pre_query FROM scheduled_tasks WHERE name='fleet-step-token-pressure';
```

Run the body up to `flagged` as a plain SELECT (drop the `ins` CTE). Do not
re-encode the thresholds anywhere — two copies of a rule is the drift class
099/102 exist to fight. Gotchas that will bite a hand-rerun, all learned the
hard way (see NOTES):

- **A truncated call logs `output_tokens = NULL`** — the cut is stated only in
  `error_message`. Any p95/peak over `output_tokens IS NOT NULL` silently
  excludes every truncation and is blindest on the steps that truncate most.
- **`agent_type` is display-only.** Rows before 2026-07-26 ~15:00 log `generic`;
  keying a measurement (or a DISTINCT ON) by agent_type splits populations and
  resurrects retired caps.
- **`round(double precision, int)` does not exist** in this Postgres — cast the
  percentile to `::numeric` before `round(x, 1)`.

## The pinned known-case test (the check must flag its own motivating case)

The scratch harness parameterises `now()`; the canonical results (2026-08-06):

- As-of `2026-08-02 18:00+00`: `classify_and_extract@6000` flags **T** (n=21,
  5 truncations) — top of the list.
- As-of `2026-08-01 00:00+00` (before ANY truncation of that step):
  `classify_and_extract@6000` flags **P** (n=15, p95 90.0%) — the leading
  indicator fires BEFORE the failure. This is what the 90-day window buys; a
  14-day clone shows n=3 (under floor, silent) on the same date.
- Live same day: `classify_and_extract` correctly ABSENT (current cap 32000,
  last run 13% of cap) — the retired-cap population stays retired.

To re-run: take the pre_query body, replace every `now()` with
`timestamp '<as-of>'` and add `AND created_at <= timestamp '<as-of>'` beside
each window bound, select from `agg` with the `flagged` predicate.

## The stated-limit check (two agents, one step_name, different caps)

The current-cap rule (single most recent call per step_name) assumes no two
agents concurrently hold the same step_name at different caps. Verify it stays
empty; if this ever returns rows, the pre_query's `latest` CTE needs a per-agent
discriminator that does NOT reintroduce the generic-era resurrection:

```sql
WITH latest AS (
  SELECT DISTINCT ON (agent_type, step_name) agent_type, step_name, max_tokens
    FROM llm_call_log
   WHERE created_at > now() - interval '14 days'
     AND step_name NOT LIKE 'review_%' AND max_tokens IS NOT NULL
   ORDER BY agent_type, step_name, created_at DESC
)
SELECT step_name, count(DISTINCT max_tokens) AS caps,
       string_agg(DISTINCT agent_type || '@' || max_tokens, ', ') AS holders
  FROM latest GROUP BY 1 HAVING count(DISTINCT max_tokens) > 1;
-- 2026-08-06: 0 rows
```

(14-day window here on purpose: it asks about CONCURRENT holders, and stays
clear of resurrecting months-old eras; it can false-positive for a few days
around a genuine cap raise — read the dates before acting.)

## Relation to FIX-058 (council-seat-token-pressure)

The two tasks partition the fleet: FIX-058 owns `step_name LIKE 'review_%'`,
this task owns the rest, same thresholds. If a seat is renamed out of the
`review_%` convention it silently changes hands — the naming convention is the
contract.
