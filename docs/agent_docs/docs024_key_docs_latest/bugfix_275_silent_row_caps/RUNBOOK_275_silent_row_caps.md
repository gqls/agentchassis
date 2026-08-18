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

## ⚠ Census the caps from `collected_data`, NOT from the pod logs (added 2026-08-18)

**The log census that used to live in the handoff cannot work, and its zero is not a negative
result.** The chassis container log rotates on size and the coordinator emits whole-state dumps, so
the retrievable window is **15–90 seconds** (measured 2026-08-18 on pods with 0 restarts, up 27
minutes). There is no aggregator — `platform/logger/logger.go:37` is `OutputPaths: ["stdout"]`.

**Bound your window before you believe any log grep.** If the oldest line postdates your event, the
grep was theatre:

```bash
kubectl -n ai-persona-system logs <pod> 2>/dev/null | grep -ao '"ts":"[^"]*"' | head -1
```

**The durable census.** `QueryDatabaseAction` writes its result to the step's `output_field`, which
lands in `orchestration_states.collected_data` and survives rolls. Every fact the WARN reports is
there, retroactively:

```sql
WITH runs AS (
  SELECT owner_agent_type AS agent, created_at, current_step,
         CASE owner_agent_type
           WHEN 'internal-linker'         THEN collected_data->'candidate_pages'
           WHEN 'content-feed-trigger'    THEN collected_data->'news_sites'
           WHEN 'model-directory-trigger' THEN collected_data->'directory_sites'
         END AS out,
         CASE owner_agent_type WHEN 'internal-linker' THEN 15
                               WHEN 'content-feed-trigger' THEN 5
                               WHEN 'model-directory-trigger' THEN 12 END AS cap
  FROM orchestration_states
  WHERE owner_agent_type IN ('internal-linker','content-feed-trigger','model-directory-trigger'))
SELECT agent, cap, created_at, current_step,
       CASE jsonb_typeof(out) WHEN 'array'  THEN jsonb_array_length(out)
                              WHEN 'object' THEN (out->>'count')::int END AS rows_returned
FROM runs WHERE out IS NOT NULL ORDER BY created_at;
```

⚠ **Handle BOTH output formats or you will report zero hits.** `output_format: array` gives a bare
array (use `jsonb_array_length`); `object` gives `{rows, count, columns, …}` (use `->>'count'`). A
first pass that only counted arrays reported **0 of 4** for `content-feed-trigger`; the true answer
was **3 of 4**.

⚠ **`orchestration_states` is pruned to roughly 2 days** (5,701 rows; only 25 older than 2 days,
measured 2026-08-18). Better than the log by ~3,000×, still not "all history" — never report its
count as a lifetime total.

⚠ **Get the cap from live config, not from this file.** Re-run the cap census above; the map of
agent → cap is what makes the query correct, and it changes.

## Catch the WARN live (the only way to read it at all)

Because the window is under two minutes, you must be attached **before** the event:

```bash
kubectl -n ai-persona-system logs -f --since=1s -l app=agent-chassis --max-log-requests=10 --prefix \
  | grep --line-buffered -a -e "EQUALS the query.s LIMIT" -e "QueryDatabaseAction: Complete" \
  | while IFS= read -r line; do printf '%s %s\n' "$(date -u +%FT%TZ)" "${line:0:900}" >> capture.txt; done
```

Three traps, all of which cost me a rearm on 2026-08-18:

- ⚠ **NO `cut` (or any block-buffering filter) in the pipeline.** `cut` holds output in a 4 KB buffer,
  so a rare small WARN never reaches the file and the empty capture reads exactly like "it did not
  fire". Truncate with `${line:0:900}` inside the loop instead. **Foreground-test the exact pipeline
  against a pattern that MUST match** (`'"level":"info"'`) before arming it — that is what exposed it.
- ⚠ **`date -u -d '2026-08-18 20:45:00'` parses as LOCAL time.** `-u` sets the output zone, not the
  input's. The deadline came out an hour early, before the event it existed to catch. Write `…Z`.
- ⚠ **Never `pkill -f <name>` from the shell that also names it** — your own command line matches, and
  the shell kills itself mid-script (exit 144, edits silently not applied). The `[w]` bracket trick
  does NOT save you when the command line genuinely contains the word. Write the PID to a file.

## Ask the running binary what built it (when `build provenance` has scrolled)

On a busy chassis the startup line is out of `--tail=3000` within minutes. Probe the binary — one
exec, many candidates, **with both controls in the same breath**:

```bash
git log --format=%H -n 120 > cands.txt
echo "deadbeefcafe0000111122223333444455556666" >> cands.txt   # must be ABSENT
git log --format=%H -n 1 --skip=3000 >> cands.txt              # real but ancient: must be ABSENT
kubectl -n ai-persona-system exec -i <pod> -- grep -aoFf - /proc/1/exe < cands.txt | sort -u
git merge-base --is-ancestor <your-commit> <the one sha that matched>   # "did my fix ship?"
```

⚠ **`grep -m1 'build provenance'` is not safe on these logs** — a substring can match inside a 183 KB
state dump and you get megabytes back. Use `grep -ao '"msg":"build provenance","git_commit":"[0-9a-f]*"'`.

## Date the roll, so you know whether the detector was even running

```bash
kubectl -n ai-persona-system get rs -l app=agent-chassis \
  -o custom-columns='CREATED:.metadata.creationTimestamp,IMAGE:.spec.template.spec.containers[0].image,DESIRED:.spec.replicas' \
  --no-headers | sort
```

⚠ Older replicasets are pruned, so the oldest surviving one carrying your image is the **earliest
evidenced** roll, not provably the first. Measured 2026-08-18: `v1.0.1309` (the first release
containing the detector) shows **15:45:31Z** — so a "24h census" run that afternoon covered mostly
hours in which the detector did not exist in the fleet.
