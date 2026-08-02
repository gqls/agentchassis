# RUNBOOK — bugs_open/172

Every command here had a gotcha attached. Read the gotcha, not just the SQL.

## 1. Is the type cap still latent? (count the types the ARTEFACT lists)

The section renders one `agent_definitions[<type>]: root ai_service` line per
gathered type, so coverage is countable from the bundle text.

```sql
WITH b AS (
  SELECT (SELECT count(*) FROM regexp_matches(body, 'agent_definitions\[[^\]]+\]: root ai_service', 'g')) AS types_listed
    FROM diagnosis_artifacts WHERE kind='bundle' AND body LIKE '%### agent state (auto-gathered%'
) SELECT types_listed, count(*) FROM b GROUP BY 1 ORDER BY 1 DESC;
```

**Gotcha:** `orchestration_states` retains ~1 day here and cannot bound this
historically. The 30-day `diagnosis_artifacts` corpus is the instrument. Re-measured
2026-08-02: still max **4** listed against a default cap of 5 — latent.

## 2. The measurement that found the SECOND cap

Counting types is not enough — you must count **distinct types in the log lines**,
which is a different question from how many types were gathered.

```sql
WITH b AS (
  SELECT
    (SELECT count(*) FROM regexp_matches(body, 'agent_definitions\[[^\]]+\]: root ai_service', 'g')) AS types_listed,
    (SELECT count(*) FROM regexp_matches(body, '- llm_call_log \[', 'g')) AS log_lines,
    (SELECT count(DISTINCT m[1]) FROM regexp_matches(body, '- llm_call_log \[[^]]*\] ([^/]+)/', 'g') m) AS types_in_log
  FROM diagnosis_artifacts WHERE kind='bundle' AND body LIKE '%### agent state (auto-gathered%'
) SELECT types_listed, log_lines, types_in_log, count(*) FROM b GROUP BY 1,2,3 ORDER BY 1 DESC;
```

**Gotcha, and it is the whole finding:** `log_lines` alone reads as healthy —
10 rows, the cap, looks like plenty of evidence. It is `types_in_log` that shows
those 10 rows all belong to **one** agent. A count of rows cannot see a
distribution; only the `count(DISTINCT)` can. 2026-08-02: **23 bundles with
`types_listed > 1 AND log_lines > 0`, and `types_in_log = 1` in every one.**

## 3. Reproduce the starvation directly (do this before believing the artefact)

```sql
-- TODAY's shape: one shared budget, allocated by global recency
SELECT agent_type, count(*) FROM (
  SELECT agent_type FROM llm_call_log
  WHERE agent_type = ANY('{page-content-writer,council-gate,diagnose-agent}'::text[])
  ORDER BY created_at DESC LIMIT 10) t GROUP BY 1;
-- => council-gate 10.  page-content-writer (18,286 rows) and diagnose-agent (324): NOTHING.

-- AFTER: per-type budget
SELECT agent_type, count(*) FROM (
  SELECT agent_type, ROW_NUMBER() OVER (PARTITION BY agent_type ORDER BY created_at DESC, id DESC) rn
  FROM llm_call_log WHERE agent_type = ANY('{page-content-writer,council-gate,diagnose-agent}'::text[])
) t WHERE rn <= 10 GROUP BY 1;
-- => 10 each.
```

**Gotcha:** pick types with *disparate volumes*. Three equally-chatty types return
a mix under both queries and the differential shows nothing — the bug hides behind
a well-chosen example.

## 4. Negative control — single-type bundles must not move

```sql
SELECT count(*) AS rows_differing FROM (
  (SELECT id FROM llm_call_log WHERE agent_type = ANY('{diagnose-agent}'::text[]) ORDER BY created_at DESC LIMIT 10)
  EXCEPT
  (SELECT id FROM (SELECT id, ROW_NUMBER() OVER (PARTITION BY agent_type ORDER BY created_at DESC, id DESC) rn
     FROM llm_call_log WHERE agent_type = ANY('{diagnose-agent}'::text[])) t WHERE rn <= 10)
) d;   -- must be 0
```

36 of the 72 retained bundles name a single type; this is what says their baseline
does not move.

## 5. PREPARE the SQL before it goes into Go

`go build` cannot parse a SQL string. The window-function query was PREPAREd against
the live schema first:

```sql
PREPARE p(text[], int) AS SELECT created_at, agent_type, COALESCE(step_name,''), model,
  COALESCE(max_tokens,0), COALESCE(output_tokens,0), success
FROM (SELECT id, created_at, agent_type, step_name, model, max_tokens, output_tokens, success,
        ROW_NUMBER() OVER (PARTITION BY agent_type ORDER BY created_at DESC, id DESC) AS rn
      FROM llm_call_log WHERE agent_type = ANY($1)) t
WHERE rn <= $2 ORDER BY agent_type, created_at DESC, id DESC;
```

**Gotcha:** `id` must be projected in the SUBQUERY to be usable in the outer
`ORDER BY`. It is not in the outer SELECT list, and that is legal — but only
because the subquery emits it.

## 6. The tests, and the one thing a mock cannot do

```bash
go test ./platform/orchestration/actions/ -run 'AgentState|CallLogCoverage|CappedType' -v
```

Mutation-prove them (a passing test proves nothing until you have watched it fail).
Copy the file aside, mutate, run, restore:

| mutation | caught? |
|---|---|
| remove `cappedTypeNotice` call | ✅ |
| heading re-asserts full coverage | ✅ |
| remove `callLogCoverageNotice` call | ✅ |
| remove `ORDER BY type` | ❌ **first attempt** — see below |

**Gotcha, and it cost a wrong assumption:** sqlmock replays rows in whatever order
the *test* supplies, so no mock-driven test can observe the database's ordering.
The first determinism test passed unchanged with `ORDER BY type` deleted. The fix
is a strict `ExpectQuery` on the query TEXT including the clause — after which the
mutation fails correctly. The ordering *guarantee* itself is Postgres's and is
verified by §3 above, not by the unit test.

## 7. Verify at the running pod (never at git, never at the tag)

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | head -1)
kubectl -n ai-persona-system exec -n ai-persona-system $POD -- sh -c \
  'strings /app/agent-chassis | grep -c "were NOT gathered (agent_state_cap"'   # positive: >=1
kubectl -n ai-persona-system exec -n ai-persona-system $POD -- sh -c \
  'strings /app/agent-chassis | grep -c "ORDER BY created_at DESC$"'            # negative control: expect 0
```

**Gotcha:** run BOTH, on EVERY replica. A positive control proves the pipeline; only
the negative control (a string the change REMOVED) proves the new binary is not the
old one carrying both. And `grep -c` is case-sensitive — mis-cased patterns read as
"not shipped" (`bugs_open/153`).
