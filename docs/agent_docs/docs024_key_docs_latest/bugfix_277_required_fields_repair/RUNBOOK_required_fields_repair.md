# RUNBOOK — bugfix 277 (commands, with their gotchas attached)

## Census / dry run (read-only, idempotent)

```bash
cd docs/agent_docs/docs024_key_docs_latest/bugfix_277_required_fields_repair
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -f - < CENSUS_2026-08-15_predicted_routes.sql
```
Gotcha: the census re-implements the seed's classifier readably. When either changes, change
BOTH, and re-prove the seed's own string (below) — the census passing proves the census, not
the seed.

## Prove the seed's exact embedded SQL against real rows (before apply / after any edit)

Extract the query FROM THE SEED FILE (never retype it), unescape `''`→`'`, substitute
`$1::uuid`/`$2::uuid` with a real (site_id, work_item_id), run via psql. The 2026-08-15 run of
exactly this proved all five canary candidates route as the census predicts (NOTES).

## Apply the seed (DB config, live instantly, INERT — 0 rows assigned)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - \
  < docs/agent_docs/sql_for_agents/410_required_fields_missing_router.sql
```
Gotcha: the verify block RAISEs (and aborts the COMMIT) on any mis-wired branch, a retired
`spec` key, a non-triaged conversion status, or >0 rows already assigned. A re-run snapshots
first (conditional DO block) and is safe.

## Canary assignment (the first execution — treat any surprise as a stop)

```sql
UPDATE site_work_items
   SET handler_agent = 'required-fields-missing-handler',
       status = 'triaged', attempt_count = 0, updated_at = NOW()
 WHERE item_type = 'required_fields_missing'
   AND status IN ('needs_human_review','unresolved')
   AND COALESCE(handler_agent,'') = ''
   AND left(id::text,8) IN ('332bb3f6','4fa5b019','e512af8a','483fb749');
-- expect UPDATE 4
```
Gotchas: `attempt_count = 0` is load-bearing (claim gate requires attempt_count <
max_attempts). Dispatch cadence is 120s; no dispatch within ~300s of a chassis pod restart.

## Verify the canary (per arm)

```sql
-- each canary row: status + recorded route
SELECT left(id::text,8), status, result->'response'->'triage'->>'route' AS route,
       result->>'route' AS route2, left(COALESCE(error,''),60) AS err
FROM site_work_items WHERE left(id::text,8) IN ('332bb3f6','4fa5b019','e512af8a','483fb749');
-- expected: 332bb3f6 complete/stale · 4fa5b019 complete/converted (+ a new content_rewrite
-- row, source='required-fields-missing-handler', spec->>'mode'='edit_live') ·
-- e512af8a needs_human_review with the blob error message · 483fb749 needs_human_review
-- with the owned message
-- conversions:
SELECT item_type, status, item_key FROM site_work_items
WHERE source = 'required-fields-missing-handler';
-- blob dedup key still held (exactly 1 non-terminal row on the key):
SELECT count(*) FROM site_work_items
WHERE item_key = (SELECT item_key FROM site_work_items WHERE left(id::text,8)='e512af8a')
  AND status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved','cancelled');
```
Gotcha: `result` is overwritten by the loop's mark_complete with the saga response — the route
survives INSIDE `result->'response'->'triage'` (that is why `complete_workflow.output_fields`
includes `triage`). For PARKED rows the loop's complete no-ops, so the route is at
`result->>'route'` (written by update_work_item_status result_fields) and the message in `error`.

## Fleet assignment (after the canary verifies)

Same UPDATE without the id filter. Then the after-state:

```sql
SELECT status, COALESCE(result->'response'->'triage'->>'route', result->>'route') AS route, count(*)
FROM site_work_items WHERE item_type='required_fields_missing' GROUP BY 1,2 ORDER BY 3 DESC;
```

## Post-roll (after the producer Go change ships in a chassis image)

```bash
# whose commit is the service running? (startup line scrolls; binary probe is durable)
kubectl -n ai-persona-system exec <chassis-pod> -- \
  sh -c "grep -oa 'buildinfo.GitCommit=[0-9a-f-]*' /proc/1/exe | head -1"
git merge-base --is-ancestor <producer-commit> <stamp>   # exit 0 = shipped
```
Then re-run the fleet assignment UPDATE once — items the OLD producer filed between
assignment and roll are born parked/unassigned and need sweeping in.

## Churn guard (+7 days)

```sql
SELECT count(*) FROM site_work_items
WHERE item_type='required_fields_missing' AND status='unresolved'
  AND created_at > '<fleet-assignment-time>';
-- expect ~0; anything else means a close arm is churning against the producer
```

## Council

Submission corr `7b0e2833-715f-4a9a-897b-efd913073582`. Verdict:
```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='7b0e2833-715f-4a9a-897b-efd913073582' AND kind='council_report'
ORDER BY created_at;
```
Gotcha: publish→run start measured at 29 min under normal load — a missing orchestration row
is latency, not a dropped dispatch; find the run by payload, never re-trigger on absence.

## Rollback

`docs/agent_docs/sql_for_agents/410_required_fields_missing_router_ROLLBACK.sql` — refuses
while non-terminal rows still route at the handler (un-assign first; header has the UPDATE).
If the producer Go change has shipped, roll that back too or items go 'blocked' at claim
(bugs_closed/077 shape).
