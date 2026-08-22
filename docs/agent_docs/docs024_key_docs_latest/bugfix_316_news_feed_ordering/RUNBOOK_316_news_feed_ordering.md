# RUNBOOK — bugs_open/316

Every query/command that was hard to get right, with its gotcha attached. Change it HERE.

---

## Read the live step config (the fact; the repo seed is history)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At -c "
SELECT jsonb_pretty(default_config->'workflow'->'steps'->'find_news_sites')
FROM agent_definitions
WHERE type='content-feed-trigger' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;"
```

⚠ **All four predicates matter.** Dropping `COALESCE(is_snapshot,false)=false` returns snapshot rows
that are not what the runtime reads; dropping `deleted_at IS NULL` returns tombstones.

## Census which sites the trigger actually picked — from `collected_data`, NEVER from the logs

The chassis log retains **15-90 seconds** (measured, `bugs_open/275` lane), so a pod sweep for a
6-hourly event returns a clean zero whether the cap bites constantly or never.
`QueryDatabaseAction` writes each result to the step's `output_field`, which survives rolls and reaches
back ~2 days:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT created_at,
       jsonb_array_length(COALESCE(collected_data->'news_sites'->'rows','[]'::jsonb)) AS n_rows,
       (SELECT string_agg(r->>'domain', ', ' ORDER BY r->>'domain')
          FROM jsonb_array_elements(COALESCE(collected_data->'news_sites'->'rows','[]'::jsonb)) r) AS domains
FROM orchestration_states
WHERE collected_data ? 'news_sites'
ORDER BY created_at DESC LIMIT 12;"
```

⚠ **`->'rows'` is required** and is the step's `output_format: object` shape. For a step declaring
`output_format: array` the payload is a bare array and `->'rows'` yields NULL — i.e. the same clean zero
you were trying to avoid. Handle both, or check the step's declared format first.

## The lateness table — the bug's disconfirming pair, in units of each site's OWN cadence

This is the query that settles whether a fix landed. **Before the fix**, overdue-ness correlates with
alphabetical rank. **After**, it must not.

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
WITH elig AS (
 SELECT DISTINCT s.id, s.domain FROM sites s
 JOIN site_specs ss ON ss.site_id=s.id AND ss.aspect='classification' AND ss.is_current=true
   AND (ss.data->'content_features'->'news_feed'->>'recommended')::boolean=true
 WHERE EXISTS (SELECT 1 FROM pages p WHERE p.site_id=s.id AND p.build_status='deployed')
)
SELECT row_number() OVER (ORDER BY e.domain) AS alpha_rank, e.domain,
   min(cs.fetch_interval) AS cadence,
   min(cs.next_fetch_at)  AS min_next_fetch,
   CASE WHEN min(cs.next_fetch_at) <= now()
        THEN justify_interval(now()-min(cs.next_fetch_at)) END AS overdue_by,
   round(100*EXTRACT(epoch FROM (now()-min(cs.next_fetch_at)))
           /NULLIF(EXTRACT(epoch FROM min(cs.fetch_interval)),0)) AS pct_of_own_cycle,
   max(cs.last_fetched_at) AS last_fetched
FROM elig e JOIN content_sources cs ON cs.site_id=e.id AND cs.is_active=true
GROUP BY e.domain ORDER BY e.domain;"
```

⚠ **A negative `pct_of_own_cycle` means not-yet-due, not early.** Read the `overdue_by` column, which is
NULL in that case, rather than the signed percentage.

⚠ **This query's `elig` CTE deliberately does NOT reproduce the trigger's own eligibility filter** — it
lists every news-feed site with a deployed page, including ones the trigger would skip. That is on
purpose: it is the denominator. Filtering it the way the trigger filters would be the "census filtered on
the very column it exists to test" trap.

## Fleet census of capped `query_database` steps (how big is the class?)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At -F '|' -c "
WITH steps AS (
  SELECT a.type AS agent_type, s.key AS step_name,
         regexp_replace(s.value->'config'->>'query', E'[\\\\s]+', ' ', 'g') AS q
  FROM agent_definitions a,
       LATERAL jsonb_each(a.default_config->'workflow'->'steps') s
  WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
    AND s.value->>'action' = 'query_database'
)
SELECT agent_type, step_name,
       COALESCE(substring(q from '(?i)ORDER BY (.*?) LIMIT'), '(none)') AS order_by,
       substring(q from '(?i)LIMIT +([0-9]+)') AS lim
FROM steps WHERE q ~* '\\mLIMIT\\M' ORDER BY 3, 1, 2;"
```

⚠ **The `regexp_replace` collapsing whitespace is load-bearing** — step queries are stored with embedded
newlines, and `substring(… from 'ORDER BY (.*?) LIMIT')` will not match across them without it.

⚠ **`substring` returns the FIRST match**, so a query with a subquery `LIMIT` reports that one. Classify
by reading the full query, not by trusting this summary — two of the fleet's hits are subquery limits
whose outer result is one row.

## Capacity arithmetic (the owner-decision half — verify, do not change)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
WITH elig AS (
 SELECT DISTINCT s.id FROM sites s
 JOIN site_specs ss ON ss.site_id=s.id AND ss.aspect='classification' AND ss.is_current=true
   AND (ss.data->'content_features'->'news_feed'->>'recommended')::boolean=true
 WHERE EXISTS (SELECT 1 FROM pages p WHERE p.site_id=s.id AND p.build_status='deployed'))
SELECT count(*) AS eligible_sites,
       round(sum(86400.0/EXTRACT(epoch FROM iv))) AS demanded_fetches_per_day
FROM (SELECT e.id, min(cs.fetch_interval) AS iv
      FROM elig e JOIN content_sources cs ON cs.site_id=e.id AND cs.is_active=true
      GROUP BY e.id) t;"
```

Supply is `runs_per_day x cap`. The trigger is 6-hourly, so 4 x 5 = 20.

## The house shape for an agent-config migration (copied from `549`, which is the current best example)

```
SELECT snapshot_agent('<type>', 'migration NNN: pre-update (<why>)');   -- BEFORE the transaction
BEGIN;
DO $$ ... RAISE EXCEPTION 'MIGRATION NNN: ...' ... $$;   -- PRE-state guards
UPDATE agent_definitions SET ... WHERE type='<type>' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
DO $$ ... RAISE EXCEPTION 'MIGRATION NNN: ...' ... $$;   -- POST-state verify
COMMIT;
```

⚠ **A verify block of bare `SELECT`s cannot stop the `COMMIT`** — `ON_ERROR_STOP` ignores a non-empty
result set. Use `DO` / `RAISE EXCEPTION`, and induce a failure once to prove the block actually aborts.
(Standing landmine; `549` follows it.)

⚠ **`snapshot_agent` has two overloads writing to two different tables.** The two-arg form used above
writes to **`agent_definitions_backup`**; the one-arg form inserts an `is_snapshot=true` row into
`agent_definitions`. Verify the snapshot by asking whether it holds the **pre-change** value, not whether
a row exists:

```sql
SELECT snapshot_taken_at,
       (default_config #>> '{workflow,steps,find_news_sites,config,query}') LIKE '%ORDER BY s.domain%' AS has_old
FROM agent_definitions_backup WHERE type='content-feed-trigger'
ORDER BY snapshot_taken_at DESC LIMIT 1;
```

### The pre-state guard is this migration's answer to the concurrent-edit problem

This row's `updated_at` moved at 08:36Z on the day of the fix for reasons nothing accounts for, on a tree
~30 sessions share. **Gate the `UPDATE` on the old query text being exactly what was measured** — e.g.
`RAISE EXCEPTION` unless the live `find_news_sites` query still ends `ORDER BY s.domain LIMIT 5`. Then a
concurrent edit **aborts the migration** instead of being silently overwritten, and the abort message
tells the next session to re-derive against the live value. That is a mechanical control where the
"I looked and found no snapshot" check was only an inference.

## Verifying the detector actually SHIPPED — at the running pod, never at git or the tag

Raised by the council's `debug_historian` seat (corr `703dbe2f`, low): a CronJob image is subject to the
same-tag-rebuild trap as any other. **A pod restart, a green build and a bumped tag all prove nothing.**

```bash
# 1. Did the CronJob run at all? (a MISSING doc_notes row means the job did not run —
#    which is NOT the same as "nothing is wrong")
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT created_at, left(body,120) FROM doc_notes
WHERE source='capped_schedule_ordering_check' ORDER BY created_at DESC LIMIT 3;"

# 2. Is the running image the one carrying the mode? Ask the BINARY, with controls.
POD=$(kubectl -n ai-persona-system get pods -l app=capped-schedule-ordering-check \
        --sort-by=.metadata.creationTimestamp -o name | tail -1)
kubectl -n ai-persona-system exec "$POD" -- grep -aq "capped-schedule-ordering" /proc/1/exe && echo PRESENT
kubectl -n ai-persona-system exec "$POD" -- grep -aq "capped-schedule-ordering-XXXX" /proc/1/exe && echo "CONTROL FAILED — matches anything"
```

⚠ **Run the negative control in the same breath.** A grep that matches everything and a grep against a
missing binary are indistinguishable from success behind the customary `2>/dev/null`. And never use
`strings` — it is absent from these images.

⚠ **The job is short-lived**, so there may be no pod to exec into between runs. In that case the
`doc_notes` row IS the evidence: it is written on clean runs too, precisely so its absence means
something.

## Proving the `doc_notes` write path without leaving a misleading row

`doc_notes.subject_type` is CHECK-constrained to eight values, and the council flagged that daily-check
inserts *"routinely fail live despite passing locally"*. `writeDocNote` uses `'pipeline'`, which is
allowed (1,878 live rows use it) — but prove it rather than reading the constraint:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 <<'SQL'
BEGIN;
INSERT INTO doc_notes (subject_type, subject_key, body, categories, source)
VALUES ('pipeline','capped-schedule-ordering','probe', jsonb_build_array('config-integrity'::text),'capped_schedule_ordering_check');
ROLLBACK;
SQL
```

⚠ **Roll it back.** Writing a real row before the cron exists would later read as "the cron ran that
day" — the row's whole meaning is that a run happened.
