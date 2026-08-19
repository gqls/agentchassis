# RUNBOOK — bugfix 313/298 (internal linker)

Every command that was hard to get right, with its gotcha attached.

## Read the live workflow (the identity check — step list, not names)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db <<'SQL'
SELECT s.key AS step, s.value->>'action' AS action,
       s.value->'config'->>'output_format' AS output_format,
       s.value->'config'->>'condition' AS condition,
       s.value->'config'->>'then_step' AS then_step,
       s.value->'config'->>'else_step' AS else_step
FROM agent_definitions a, LATERAL jsonb_each(a.default_config->'workflow'->'steps') s
WHERE a.type='internal-linker' AND a.is_active AND COALESCE(a.is_snapshot,false)=false
  AND a.deleted_at IS NULL
ORDER BY s.key;
SQL
```
⚠ `internal-linker` ≠ `internal-link-resolver` (LANDMINES: one word apart, the busy one is not
the defective one). The step list above IS the identity — check it, not the name.
⚠ `agent_definitions` has no `name` column (`display_name`), and `updated_at` moves in
roll-window bulk writes (200 rows/minute) — never date an edit from it.

## 298's census, with the query's OWN predicate

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db <<'SQL'
WITH cand AS (
  SELECT p.site_id, count(*) AS n FROM (
    SELECT p.site_id, p.name
    FROM pages p LEFT JOIN page_components pc ON pc.page_id = p.id AND pc.rendered_html IS NOT NULL
    WHERE p.status='active' AND p.page_type IN ('content','service','landing','tool')
    GROUP BY p.site_id, p.name, p.url, p.title, p.page_type
    HAVING COUNT(pc.id) > 0
  ) p GROUP BY p.site_id)
SELECT count(*) sites, count(*) FILTER (WHERE n > 15) over_cap,
       percentile_cont(0.5) WITHIN GROUP (ORDER BY n) median, max(n) worst FROM cand;
SQL
```
⚠ Omitting the `HAVING COUNT(pc.id) > 0` over-counts (298's own first-pass misstep).

## Apply migration 490 (direct psql — NOT the runner)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 \
  < docs/agent_docs/sql_for_agents/490_internal_linker_candidates_object_uncapped_fail_loud.sql
```
⚠ `MIGRATIONS_DIR=… ./run-migrations.sh --apply` scopes NOTHING if the assignment lands on its
own line, and an unscoped run applies ~100 other threads' pending files (LANDMINES). Direct psql
per file is the 275-lane recipe.
A clean apply prints: `NOTICE` (snapshot), `BEGIN`, `DO`, `UPDATE 1`, `DO`, `COMMIT`.
**`UPDATE 0` + `COMMIT` is a silent no-op — read that line.**

Then record it (the runner's record-only mode writes `schema_migrations` with checksum):
```bash
scripts/migration/run-migrations.sh --record-only 490_internal_linker_candidates_object_uncapped_fail_loud.sql \
  "applied by direct psql, bugfix_313 lane, 2026-08-19"
# (check the exact record-only invocation in the script header before first use)
```

## Verify the fix at the artefact (never at the status)

```bash
# 1) the first llm_call_log row in all history = the disconfirming pair's "after" arm
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
 "SELECT created_at, step_name, LEFT(prompt_rendered, 200) FROM llm_call_log
  WHERE agent_type='internal-linker' ORDER BY created_at DESC LIMIT 3"

# 2) the prompt rendered PAGES, not the map's key names (the broken-prompt trap as evidence)
#    look for '### <page name> (' lines under '## Candidate Pages'; 'rows'/'count'/'columns' = FAIL

# 3) durable run-shape check (survives rolls, ~2-day retention)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
 "SELECT created_at, current_step, status,
         jsonb_array_length(COALESCE(collected_data->'candidate_pages'->'rows','[]'::jsonb)) AS cand_rows
  FROM orchestration_states WHERE owner_agent_type IS NOT DISTINCT FROM owner_agent_type
    AND collected_data ? 'candidate_pages' ORDER BY created_at DESC LIMIT 5"
```
⚠ `orchestration_states.owner_agent_type` is the PARENT for loop-dispatched handlers — do not
filter the linker's runs by it (275 lane trap #2); key on `collected_data ? 'candidate_pages'`.
⚠ The LCO-009 WARN line lives 15–90 s in the pod log — never census from logs.

## Run the new audit offline (after it exists; Go, so binary = this tree)

```bash
scripts/audit-array-producer-conditions.sh          # exit 1 on findings; clean fleet expected post-490
go test ./cmd/config-key-audit/ ./platform/orchestration/actions/ ./platform/orchestration/datahelpers/
```
