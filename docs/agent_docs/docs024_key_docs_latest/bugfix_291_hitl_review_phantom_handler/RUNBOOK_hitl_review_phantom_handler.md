# RUNBOOK — bugfix 291

## Census (the bug's live population)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
 "SELECT count(*), count(DISTINCT site_id), min(created_at)::date, max(updated_at)
  FROM site_work_items
  WHERE status='blocked' AND error='Handler agent not registered: hitl-review';"
```
Gotcha: use EXACT equality on the error — claim writes `'Handler agent not registered: ' || $2`
verbatim; `LIKE` with no wildcard is equality spelt confusingly and a `%` would also match
other phantom handlers (which would be a DIFFERENT bug's rows).

## Where `hitl-review` lives in live config (should be ONE path; zero after Phase 3)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
WITH RECURSIVE walk(path, val) AS (
  SELECT ''::text, default_config FROM agent_definitions
   WHERE type='tool-auditor' AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false
  UNION ALL
  SELECT path || '.' || key, value FROM walk, LATERAL jsonb_each(val) AS e(key, value)
   WHERE jsonb_typeof(val)='object')
SELECT path, val FROM walk WHERE jsonb_typeof(val)='string' AND val::text LIKE '%hitl%';"
```
Gotcha: `jsonb_path_query` with `like_regex` also works but returns the VALUE only; the
recursive walk gives you the path, which is what the migration gate needs.

## The needs_human_review handler split (settles the bug file's corrected claim)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
 "SELECT handler_agent, status, count(*) FROM site_work_items
  WHERE item_type='needs_human_review' GROUP BY 1,2 ORDER BY 3 DESC;"
```

## Straggler sweep (re-runnable; for rows filed by in-flight OLD-config auditor runs after 448)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
UPDATE site_work_items
   SET status='needs_human_review', handler_agent='', error=NULL,
       result = COALESCE(result,'{}'::jsonb) || jsonb_build_object('repair_291',
         jsonb_build_object('repaired_at', now()::text, 'from_status','blocked',
           'from_handler','hitl-review', 'straggler', true)),
       updated_at=now()
 WHERE status='blocked' AND error='Handler agent not registered: hitl-review'
   AND created_by='tool-auditor';"
```
Run the census again the NEXT DAY — an auditor orchestration loaded before 447 applied can
file at the old shape for as long as it lives.

## Migration apply (hand-apply, never bare --apply on this shared tree)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/447_tool_auditor_review_items_park_at_needs_human_review.sql
./scripts/migration/run-migrations.sh --record-only 447_tool_auditor_review_items_park_at_needs_human_review.sql
```
Gotchas: `--apply` takes EVERY pending file (other sessions keep strays queued — 213/214/324
were sitting untracked at session start). An aborted psql session needs `ROLLBACK;` first.
Verify the snapshot in **`agent_definitions_backup` ordered by `snapshot_taken_at`** — NOT
`is_snapshot` rows in agent_definitions (landmine).

## Prove the config fix at the artefact

```bash
# after the next tool-auditor run (~hourly when its site has audit items):
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
 "SELECT status, handler_agent, created_at FROM site_work_items
  WHERE created_by='tool-auditor' AND item_type='needs_human_review'
  ORDER BY created_at DESC LIMIT 5;"
```
New rows must arrive at `needs_human_review` (handler still `hitl-review` until Phase 3 — that
is expected and inert).

## Prove the Go guard shipped (before Phase 3)

```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor <guard-commit-sha> <stamped-sha> && echo SHIPPED
```
Gotcha: the provenance line is a STARTUP line and scrolls; empty ≠ unstamped — fall back to
the known-value binary probe with a control (LANDMINES).
