# RUNBOOK — bugs_open/275 silent row caps

## Census every row cap in live agent config

```sql
SELECT a.type, s.key AS step,
       substring(s.value->'config'->>'query' from 'LIMIT[[:space:]]+[0-9]+') AS limit_clause
FROM agent_definitions a,
     LATERAL jsonb_each(COALESCE(a.default_config->'workflow'->'steps','{}'::jsonb)) s
WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
  AND s.value->'config' ? 'query'
  AND s.value->'config'->>'query' ~* 'LIMIT[[:space:]]+[0-9]+'
ORDER BY 1,2;
```

⚠ **Most hits are `LIMIT 1` and are NOT defects** — that is the fetch-one/claim-one idiom. Read the
multi-row ones. 26 hits / 7 multi-row as of 2026-08-16.

⚠ **A cap only matters if it BITES** — i.e. the population exceeds it. That is a second query per
step, and nobody has run it for the other six. LCO-009's WARN is what will answer it post-roll.

## Size a prompt payload before choosing a cap (the method that decided this fix)

Never argue about "prompt size" — measure per column, because one column is usually most of it:

```sql
WITH r AS (SELECT <the step's exact SELECT list> FROM <table> WHERE <the step's exact predicate>)
SELECT sum(length(colA)) a_ch, sum(length(colB)) b_ch, ...,
       round(avg(length(bigcol))) avg_big,
       percentile_disc(0.5) WITHIN GROUP (ORDER BY length(bigcol)) median_big,
       count(*) FILTER (WHERE length(bigcol) > 200) over_200
FROM r;
```

Here `description` was **80%** of the payload, so bounding it bought the whole library for +22%.
**Then read the prompt template** to confirm which columns are actually rendered — `category` is
selected and never used, which no amount of SQL would have told me.

## Applying an `agent_definitions` migration safely on a shared tree

**DB config is LIVE ON APPLY.** No roll gates it; no redeploy undoes it. So:

1. Commit the migration AND its `_ROLLBACK` sidecar **before** applying.
2. Apply the single file directly — this avoids the unscoped-runner landmine entirely
   (`MIGRATIONS_DIR=… ./run-migrations.sh --apply` scopes NOTHING if the assignment lands on its own
   line, and then applies ~100 other threads' pending files):
   ```bash
   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < docs/agent_docs/sql_for_agents/NNN_x.sql
   ```
3. **Scope by id with a PRE-STATE GATE** (the shape 406's council round asked for): refuse unless the
   row is the pinned one AND still carries the text you wrote against. On a tree this many sessions
   share, an unguarded `UPDATE … WHERE type='x'` silently clobbers whatever landed since you looked.
4. **Verify with `DO`/`RAISE`, never bare `SELECT`s** — `ON_ERROR_STOP` ignores a non-empty result, so
   a SELECT-only verify block cannot stop the `COMMIT`.

A clean apply prints: `NOTICE: Snapshot captured…`, `BEGIN`, the snapshot row, `DO`, `UPDATE 1`, `DO`,
`COMMIT`. **`UPDATE 0` with a `COMMIT` is a silent no-op** — the gate passed but the WHERE matched
nothing. Read that line.

## Verify a suggester-visibility fix (the bug's own disconfirming pair)

```sql
-- which tools were previously unreachable?
WITH ranked AS (
  SELECT display_name, row_number() OVER (ORDER BY display_name) rn
  FROM content_components
  WHERE component_level='tool' AND forked_from IS NULL AND is_active AND html_template != '')
SELECT count(*) FILTER (WHERE rn > 30) AS previously_invisible,
       min(display_name) FILTER (WHERE rn = 31) AS first_newly_visible
FROM ranked;
```

The end-to-end proof the bug asks for is `llm_call_log.prompt_rendered` for a `suggest_tools` call:
a late-alphabet tool must be absent before and present after. **That needs a real tool-suggester run**,
so it is owed until one happens naturally — the config check above is necessary, not sufficient.
