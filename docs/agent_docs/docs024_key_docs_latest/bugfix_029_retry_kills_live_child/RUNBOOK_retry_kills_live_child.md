# RUNBOOK — `bugs_open/029`, retry-kills-live-child

Every query here had to be got right once. The gotcha is attached to the command,
not left in someone's scrollback.

All DB access:
`kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`

---

## Is this bug biting right now?

**Do not read the `AWAITING_RESPONSES` census** — this bug file's own 2026-07-21 section
records watching it go 16 → 24 → 13 and misreading it twice. It mixes healthy in-flight
rows with dead ones. Use the two below instead.

```sql
-- 1. the union of BOTH under-reporting tables (neither alone is a census)
SELECT agent_type, step_name, action, count(*) AS n,
       min(occurred_at)::timestamp(0) AS first, max(occurred_at)::timestamp(0) AS last
  FROM agent_error_log
 WHERE error_message ILIKE '%timed out after%' AND occurred_at > now() - interval '4 days'
 GROUP BY 1,2,3 ORDER BY 4 DESC;
```

```sql
-- 2. wedged children: EXECUTING_STEP rows the 4-hour reaper eventually kills
SELECT current_step, status, left(error,70) AS err, count(*)
  FROM orchestration_states
 WHERE owner_agent_type='build-dispatch-loop' AND status='FAILED'
   AND created_at > now() - interval '4 days'
 GROUP BY 1,2,3 ORDER BY 4 DESC;
```

> **Both tables under-report, in OPPOSITE directions** (`bugs_open/029`, 2026-07-28 and
> 2026-08-10 contributions). `orchestration_states` has shown 0 FAILED against 79 logged
> timeouts; a `page-rerender` run has shown the exact inverse — FAILED in
> `orchestration_states` with **zero** rows in `agent_error_log`. A clean read of either
> one alone is not evidence of health.

---

## The freeze time is `last_activity`, NEVER `updated_at`

`updated_at` on `orchestration_states` is bumped by the **reaper** when it writes FAILED,
so using it makes every wedged row look like it lived for ~4h26m — uniform, plausible and
wrong. I did exactly this first.

```sql
SELECT orchestration_id, current_step,
       created_at::timestamp(0)     AS created,
       last_activity::timestamp(0)  AS froze,      -- the real freeze
       updated_at::timestamp(0)     AS reaped,     -- when the reaper wrote FAILED
       (last_activity-created_at)::interval(0) AS ran_for
  FROM orchestration_states
 WHERE owner_agent_type='build-dispatch-loop' AND status='FAILED'
   AND error LIKE 'reaper: stale EXECUTING_STEP%' AND created_at > now() - interval '4 days'
 ORDER BY last_activity;
```

---

## The decisive join: child freeze against the parent's retry clock

This is the query that turned a correlation into a mechanism. **`correlation_id` is `uuid`
on `orchestration_states` and `varchar` on `awaited_requests`** — the join fails with
`operator does not exist: uuid = character varying` unless you cast. Cast the *uuid* side.

```sql
WITH child AS (
  SELECT orchestration_id, correlation_id::text AS corr, created_at, last_activity, current_step
    FROM orchestration_states
   WHERE owner_agent_type='build-dispatch-loop' AND status='FAILED'
     AND error LIKE 'reaper: stale EXECUTING_STEP%' AND created_at > now() - interval '4 days')
SELECT c.current_step,
       (c.last_activity-c.created_at)::interval(0)  AS child_ran_for,
       ar.retry_version, ar.status AS awaited_status,
       (ar.timeout_at-ar.sent_at)::interval(0)      AS last_window,
       (c.last_activity - ar.sent_at)::interval(0)  AS freeze_minus_last_send
  FROM child c
  LEFT JOIN awaited_requests ar
    ON ar.correlation_id = c.corr AND ar.step_name='call_dispatch'
 ORDER BY c.created_at;
```

A small **positive** `freeze_minus_last_send` (seconds, not minutes) is the signature.

---

## Prove the window truncation without reading any Go

```sql
SELECT step_name, retry_version, count(*) AS n,
       percentile_cont(0.5) WITHIN GROUP (ORDER BY (timeout_at-sent_at))::interval(0) AS median_window
  FROM awaited_requests WHERE sent_at > now() - interval '4 days'
 GROUP BY 1,2 HAVING count(*) > 5 ORDER BY 1,2;
```

Compare `retry_version=0` (the declared value) against `retry_version>=1` (always 05:00,
or 03:00 where the step declares more than 30 minutes).

---

## Size the blast radius of the truncation

```sql
WITH steps AS (
  SELECT ad.type AS agent_type, s.key AS step_name,
         (s.value->'config'->>'timeout_seconds')::int AS tmo
    FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s
   WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
     AND s.value->'config' ? 'timeout_seconds')
SELECT tmo AS declared_s, count(*) AS steps, count(DISTINCT agent_type) AS agents
  FROM steps WHERE tmo > 300 GROUP BY 1 ORDER BY 1 DESC;
```

> **Read the LIVE row, not the seed.** `SEED_*.sql` records what an agent *was*;
> `agent_definitions` is what it *is*. Filter
> `is_active AND NOT is_snapshot AND deleted_at IS NULL` or you will count snapshots.

---

## What a callee legitimately takes (before calling any timeout "generous")

```sql
SELECT owner_agent_type, count(*) AS completed_runs,
       round(100.0*count(*) FILTER (WHERE (updated_at-created_at) > interval '5 minutes')/count(*),1)  AS pct_over_300s,
       round(100.0*count(*) FILTER (WHERE (updated_at-created_at) > interval '15 minutes')/count(*),1) AS pct_over_900s
  FROM orchestration_states
 WHERE status='COMPLETED' AND created_at > now() - interval '36 hours'
 GROUP BY 1 HAVING count(*) > 50 ORDER BY 2 DESC;
```

`updated_at` is safe **here** — for a COMPLETED row it is the completion. It is only
misleading for reaped rows (see above).

---

## Before dispatching anything that costs credits

```bash
kubectl -n ai-persona-system get pods -l app=agent-chassis \
  -o custom-columns=NAME:.metadata.name,START:.status.startTime,READY:.status.containerStatuses[0].ready
```

Under ~300s of uptime, the spawn is silently dropped and you pay for nothing. This bug
file's 2026-07-26 contribution lost two council rounds to exactly that.

---

## Filing the diagnosis run

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/090_TRIGGER_needs_diagnosis_v1.sh "<symptom>"
```

Check the queue first — `SELECT summary, status FROM site_work_items WHERE
item_type='needs_diagnosis' AND status='awaiting_diagnosis';` — and note the script prints
**two** correlations. The **RUN** correlation (minted by `diagnose-dispatch-loop`, stamped
back as `spec.dispatch_correlation_id`) is the key the artifacts are written under. The
intake correlation is not.

```sql
-- find the run
SELECT owner_agent_type, current_step, status, created_at::timestamp(0), updated_at::timestamp(0)
  FROM orchestration_states WHERE correlation_id::text = '<RUN_CORRELATION_ID>' ORDER BY created_at;
```

---

## Censusing the wedge signature: do NOT filter by `item_type`

*Contributed by the `site_ai_agent_orchestration` lane, 2026-08-18, as a correction to its
own measurement — it is a correctness condition on this diagnostic, not just on theirs.*

The per-site mutex in `find_dispatchable_site` is
`NOT EXISTS (SELECT 1 FROM site_work_items active WHERE active.site_id = wi.site_id AND
active.status='claimed')`. It has **no `item_type` clause** — a claim of *any* type
excludes the whole site.

So a census filtered to one type (theirs was `item_type='page_rerender'`) **cannot
distinguish "this site holds no claim" from "this site holds a claim I filtered out"**.
Their number happened to be right and the inference from it was unsupported.

```sql
-- CORRECT: unfiltered by item_type, which is what the mutex actually sees
SELECT s.domain, count(*) FILTER (WHERE wi.status='claimed')  AS claimed_any_type,
       count(*) FILTER (WHERE wi.status='triaged')            AS triaged,
       min(wi.claimed_at) FILTER (WHERE wi.status='claimed')::timestamp(0) AS oldest_claim
  FROM sites s JOIN site_work_items wi ON wi.site_id = s.id
 GROUP BY 1 HAVING count(*) FILTER (WHERE wi.status='claimed') > 0
 ORDER BY 4;
```

**Reading it.** Sites rotating — a different site holding one claim each minute, taken and
released inside a minute — is the **healthy** baseline. The wedge is *one site holding one
claim STATIC for tens of minutes*. The count alone does not separate them; the **age** of
the claim does.

> ⚠ **Take claim age from `claimed_at`, not `updated_at`.** `trg_site_work_items_updated_at`
> bumps `updated_at` on every write, so a periodically-touched row looks perpetually fresh
> — the same trigger behind the unreapable-work-item landmine. This is the `site_work_items`
> twin of the `last_activity` vs `updated_at` trap above; two tables, one mistake.

---

## Does the INITIAL wait honour the declaration? (join per AGENT, never per step name)

Added 2026-08-18 when this query, written the lazy way, invented two defects that did not exist.

**A step name is not a key.** `call_handler` is declared **2100** by `diagnose-dispatch-loop`
and **1200** by `report-dispatch-loop`; `trigger_deploy` is 180 by `rerender-site` and 120 by
`section-editor`. Aggregate the declaration across agents (`max(...)`, `min(...)`, or a join on
`s.key` alone) and you will report a truncation that is only your own aggregate disagreeing
with a different agent's config. Join on **(agent, step)**, via the owning orchestration:

```sql
WITH obs AS (
  SELECT os.owner_agent_type AS agent,
         regexp_replace(ar.step_name,'^process_(item|sites)_iter_[0-9]+_','') AS base_step,
         count(*) AS n,
         min(extract(epoch from (ar.timeout_at-ar.sent_at)))::int AS min_win,
         max(extract(epoch from (ar.timeout_at-ar.sent_at)))::int AS max_win
    FROM awaited_requests ar
    JOIN orchestration_states os ON os.orchestration_id = ar.orchestration_id
   WHERE ar.retry_version=0 AND ar.sent_at > now() - interval '3 days'
   GROUP BY 1,2),
decl AS (
  SELECT ad.type AS agent, s.key AS base_step, (s.value->'config'->>'timeout_seconds')::int AS declared
    FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s
   WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
     AND s.value->'config' ? 'timeout_seconds')
SELECT o.agent, o.base_step, o.n, o.min_win, o.max_win, d.declared,
       CASE WHEN o.min_win BETWEEN d.declared AND d.declared+2 THEN 'OK' ELSE '*** MISMATCH ***' END
  FROM obs o JOIN decl d USING (agent, base_step) ORDER BY 7, o.n DESC;
```

Result 2026-08-18: **18 pairs, 18 OK.** The `+2` slack is not cosmetic — `SentAt` and
`TimeoutAt` are **two separate `time.Now()` reads**, so the stored difference is always
marginally over the declaration (the same arithmetic that puts a step declaring exactly 1800
into the old `> 30*time.Minute` arm).

> ⚠ **`agent_definitions` is not the whole universe of declarations.** `call_verifier`,
> `call_ingester`, `call_section_editor` and `call_planner` appear in `awaited_requests` with
> distinct non-default windows and exist in **no** active agent definition — they come from
> per-run plans in `orchestration_states.workflow_plan`, which is also what `retryWindow`
> reads. A step missing from the `decl` side means *"not declared here"*, never *"not declared"*.

### ⚠ `retry_version=0` is a SURVIVORSHIP filter, not a neutral one

`awaited_requests` is `PRIMARY KEY (request_id)` — **one row per request, rewritten in place on
retry.** `retry_version` is bumped and `sent_at`/`timeout_at` are overwritten, so a retried
request's earlier windows are **destroyed, not archived**.

So `WHERE retry_version=0` quietly means *"requests that never retried"*. Any defect that
shortens a window makes its own rows retry and **leave the population you are counting**. The
census is still usable — a short window that the callee answers inside still shows up at rv0 —
but state the residual: a truncation whose window **always** expires is invisible here, and the
only instrument that would see it is the log line at registration.

The same applies in reverse to reading a retried row: a 180s window at rv≥1 tells you nothing
about what that request was granted at rv0. Read `retry_version` before drawing any conclusion
about which code path set the number.
