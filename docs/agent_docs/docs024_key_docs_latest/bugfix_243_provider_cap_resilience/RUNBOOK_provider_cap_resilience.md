# RUNBOOK — 243-anthropic-cap

Every command that had to be got right, with its gotcha attached. Change it HERE, not in
scrollback.

## 1. Is the cap biting RIGHT NOW?

**Use the histogram, not `max(created_at)`.** A single `max(... WHERE success)` cannot tell you
whether failures are still arriving, and an **hourly bucket cannot see an outage that started
inside it** — that trap cost the 08-22 reporter a wrong first read ("intermittent, not a wall")
because the window straddled the cutover.

```sql
SELECT date_trunc('hour', created_at) AS hr,
       count(*) FILTER (WHERE success) AS ok,
       count(*) FILTER (WHERE NOT success) AS failed,
       count(*) FILTER (WHERE NOT success AND error_message ILIKE '%usage limit%') AS cap
FROM llm_call_log WHERE created_at > now() - interval '14 hours' GROUP BY 1 ORDER BY 1;
```

If `cap > 0` in the current hour, split at the last success to see which side you are on:

```sql
SELECT max(created_at) FROM llm_call_log WHERE success;   -- then count either side of it
```

**Do not announce an outage or defer work on one observation.** The same message accompanies a
3-minute blip (08-17) and a 3h20m exhaustion (08-10) — the text carries no duration information
at all. Re-run a few minutes later before concluding anything.

## 2. Is the DISPATCH QUEUE stopped? (invisible in `llm_call_log`)

A stopped queue does **not** show up in the histogram above — LLM calls keep succeeding while
nothing can be claimed. This is the seam-A wedge.

```sql
SELECT name, healthy, last_checked, last_healthy, check_interval_seconds, left(error,120)
  FROM ai_endpoint_health WHERE name = 'claude';
SELECT max(claimed_at) FROM site_work_items WHERE claimed_by = 'build-dispatch-loop';
```

`healthy=f` **plus** no claim since roughly `last_checked` = this bug's dispatch wedge, not
queue latency.

## 3. Days on which the cap actually bit (the rate, not the anecdotes)

```sql
SELECT date_trunc('day', created_at)::date AS d,
       count(*) FILTER (WHERE NOT success AND error_message ILIKE '%usage limit%') AS cap_failures,
       count(*) FILTER (WHERE success) AS ok
FROM llm_call_log WHERE created_at > now() - interval '16 days' GROUP BY 1 ORDER BY 1;
```

Counting **days with failures** rather than narrated incidents is what showed 7 of 15 days
rather than "three occurrences".

## 4. Effective spend, now that caching is live

Full-price input alone understates nothing and overstates nothing — but it is **not** the bill.
Weight it: reads 0.1×, writes 1.25×.

```sql
SELECT date_trunc('day', created_at)::date AS d,
       sum(input_tokens) AS full_price,
       sum(COALESCE(cache_read_input_tokens,0)) AS cache_read,
       sum(COALESCE(cache_creation_input_tokens,0)) AS cache_write
FROM llm_call_log WHERE provider='anthropic' AND created_at > now() - interval '16 days'
GROUP BY 1 ORDER BY 1;
```

⚠ Quoting `full_price` alone makes the caching fix look like a 7× win. On the weighting above
it is ~30% — and total prompt volume grew. Both numbers are true; only the second predicts a cap.

## 5. Is a council round dead, or merely queued?

```sql
SELECT current_step, status, collected_data->'__step_error'->>'message'
FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '<YOUR_SUBMISSION_CORR>';
```

`complete_invalid` + a `usage limits` message = **this bug, a transient — resubmit.** That is the
*opposite* of the standing advice for a missing orchestration row (latency — do not retry), so
tell the two apart before acting.

⚠ **Resubmit with `RESUBMIT_CORR=<corr>`.** Without it you mint a NEW correlation, which
silently orphans any commit already carrying the old one in a `Council-Submitted:` trailer —
`098` joins on that id and the old run has no verdict to resolve to, ever.

## 6. The council's error routing (the seam-B census)

⚠ **`error_step` is nested inside `config`, not at the step level.** Read at step level the same
query returns "(none) | 29" — clean, confident, wrong.

```sql
SELECT COALESCE(v->'config'->>'error_step','(none)'), count(*)
FROM agent_definitions, LATERAL jsonb_each(default_config #> '{workflow,steps}') AS s(k,v)
WHERE type='council-gate' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND (k LIKE 'review_%' OR k LIKE 'gate_%') GROUP BY 1;
```

To see each seat's own `next_step` (needed for the C4(b) repoint — each seat routes to its own
successor, they are not all the same):

```sql
SELECT k AS step, v->'config'->>'error_step' AS error_step, v->>'next_step' AS next_step
FROM agent_definitions, LATERAL jsonb_each(default_config #> '{workflow,steps}') AS s(k,v)
WHERE type='council-gate' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND k LIKE 'review_%' ORDER BY k;
```

## 7. Reading a 090 diagnosis outcome — check BOTH places

The settling query from LANDMINES tells you whether a verdict exists:

```sql
SELECT (SELECT count(*) FROM diagnosis_artifacts WHERE correlation_id='<RUN_CORR>' AND kind<>'bundle') AS verdicts,
       (SELECT status FROM site_work_items WHERE spec->>'dispatch_correlation_id'='<RUN_CORR>') AS item;
```

`0 | complete` ⇒ over, no verdict. **But that does not tell you WHY, and the landmine's stated
discriminator (the body-omission line) can come back clean:**

```sql
-- may be empty even on a no-verdict run — an iteration-cap stop truncates nothing
SELECT iteration, length(body), substring(body from '_\(body omitted[^)]*\)_')
FROM diagnosis_artifacts WHERE correlation_id='<RUN_CORR>' AND kind='bundle' ORDER BY iteration;

-- THE MISSING HALF: an iteration-cap stop records its conclusion HERE, not in artifacts
SELECT result->'response'->'response'->>'status',
       result->'response'->'response'->>'summary'
FROM site_work_items WHERE spec->>'dispatch_correlation_id'='<RUN_CORR>';
```

`UNVERIFIABLE` + "stopped: iteration-cap" = the loop neither confirmed nor refuted. **Do not
present it as support**; take the owner ruling's declared-substitute path explicitly.

⚠ The trigger prints **two** correlations. Artifacts are written under the **second**
(`RUN_CORRELATION_ID`), not the intake one.

## 8. Probing the endpoint yourself

**Never extract the key into the session** (owner 08-23). Probe from the pod. And remember
`pingClaude`'s own verdict is not evidence about the cap: it returns healthy on any non-auth
status, and the cap is a **400**.
