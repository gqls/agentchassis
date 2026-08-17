# RUNBOOK — bugfix 287 (spawn_record)

DB access:
`kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`

## Watch the defect / the fix land (per agent — the roll is per SERVICE)

```sql
-- Resolver instrument (RFC_029/CTS-060). A correct fix takes field='result' to ZERO for the
-- agent WHILE it has traffic. work_item_id/current_page conflict rows are NOT this bug
-- (search noise for already-resolved fields / another extraction) — do not judge the fix by them.
SELECT date_trunc('hour', occurred_at) hr, error_code, context->>'field' AS field, count(*)
FROM agent_error_log
WHERE error_code LIKE 'RESOLVER_%' AND agent_type='build-dispatch-loop'
  AND occurred_at > now() - interval '24 hours'
GROUP BY 1,2,3 ORDER BY 1 DESC, 4 DESC;
```

```sql
-- Item census (287 §2/§8). Fix = spawn_record 0 WHILE own_envelope > 0 (demand pair).
SELECT date_trunc('hour', updated_at) hr,
       count(*) FILTER (WHERE result ? 'response')                       AS own_envelope,
       count(*) FILTER (WHERE result ? 'topics' AND result ? 'agent_id') AS spawn_record,
       count(*) total
FROM site_work_items
WHERE status='complete' AND updated_at > now() - interval '24 hours' AND handler_agent IS NOT NULL
GROUP BY 1 ORDER BY 1 DESC;
```

Demand control (zero rows also means no traffic):
```sql
SELECT count(*) FROM orchestration_states
WHERE owner_agent_type='build-dispatch-loop' AND created_at > now() - interval '24 hours';
```

## Census: loop substep config strings referencing sibling outputs

Gotcha that cost a wrong zero: the sub_workflow lives at `s.value->'config'->'sub_workflow'`
— NOT `s.value->'sub_workflow'`. Pretty-print one row before trusting a zero-row walk.

```sql
WITH substeps AS (
  SELECT ad.type AS agent, s.key AS loop_step, sub.key AS substep,
         sub.value->>'action' AS action, sub.value->'config' AS cfg,
         s.value->'config'->'sub_workflow'->'steps' AS siblings
  FROM agent_definitions ad,
       LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s(key,value),
       LATERAL jsonb_each(s.value->'config'->'sub_workflow'->'steps') sub(key,value)
  WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
    AND s.value->>'action'='loop'
), sibling_outputs AS (
  SELECT agent, loop_step,
         array_agg(DISTINCT sib.value->>'output_field')
           FILTER (WHERE sib.value->>'output_field' IS NOT NULL) AS outs
  FROM (SELECT DISTINCT agent, loop_step, siblings FROM substeps) x,
       LATERAL jsonb_each(x.siblings) sib(key,value)
  GROUP BY 1,2
)
SELECT ss.agent, ss.loop_step, ss.substep, ss.action, kv.key, kv.value
FROM substeps ss
JOIN sibling_outputs so ON so.agent=ss.agent AND so.loop_step=ss.loop_step,
LATERAL (SELECT key, value #>> '{}' AS value FROM jsonb_each(ss.cfg)
         WHERE jsonb_typeof(value)='string') kv
WHERE split_part(kv.value,'.',1) = ANY(so.outs)
ORDER BY 1,2,3,5;
```

## Prove the roll before lifting migration 450 (_HOLD)

```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
# startup line scrolls — empty means "not in range", fall back to the binary probe with a
# known-present AND known-absent sha (never a discovery grep):
kubectl -n ai-persona-system exec <chassis-pod> -- grep -aq "<expected-sha>" /proc/1/exe
git merge-base --is-ancestor <half-1-commit> <the stamp>
```

## Migrations

Dir: `docs/agent_docs/sql_for_agents/`. Runner: `scripts/migration/run-migrations.sh`
(dry-run per session; `--apply` takes EVERY pending file — scope it). `_HOLD.sql` and
`_ROLLBACK.sql` match `SIDECAR_RE` and are excluded from `--apply` (still listed).
Hand-apply + `--record-only <file> --note "..."` per migration-runner practice.
Template for a strict-marker HOLD migration: `417_image_build_handler_asset_id_goes_strict_HOLD.sql`.

## Tests

```bash
cd /home/ant/projects/agentchassis
go test ./platform/orchestration/ -run TestLoopConfigReferenceSuffixing -count=1
# shared-tree rule: prove HEAD compiles from an archive, not the dirty tree
T=$(mktemp -d) && git archive HEAD | tar -x -C "$T" && (cd "$T" && go build ./... ) && rm -rf "$T"
```
