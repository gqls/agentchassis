# RUNBOOK — bugs 235 / 155 / 071 close-out

## Fire / track the 209 behavioural proof (155's closure test, legacy pair)

```bash
cd docs/agent_docs/docs024_key_docs_latest/bugfix_209_deploy_purpose_keyed_source
./fire_209_proof.sh pageflow   # or: swo — prints CORRELATION_ID; PUBLISH_OK or it did not send
```
Gotcha: kcat exits 0 having sent nothing — only `PUBLISH_OK` counts. Queue latency to run
start can be ~30 min; a missing orchestration row is latency, not a drop. Find by payload:

```sql
SELECT id, current_step, status, created_at FROM orchestration_states
WHERE collected_data->'input_data'->>'domain'='cookly.uk'
ORDER BY created_at DESC LIMIT 5;
-- or by the fired correlation:
SELECT id, current_step, status FROM orchestration_states
WHERE correlation_id='<CORRELATION_ID>' OR collected_data->'headers'->>'correlation_id'='<CORRELATION_ID>';
```

Then `./verify_209_proof.sh` — asserts hero vs logo artefact sha256 DIFFER from
`origin/master` of ~/projects/sites. `RESULT: PASS` is the read.

## Live config reads (schema quirk)

`agent_definitions.default_config->'workflow'->'steps'` is an OBJECT keyed by step name,
NOT an array — `jsonb_array_elements` fails with "cannot extract elements from an object".
Read a step directly:

```sql
SELECT jsonb_pretty(default_config->'workflow'->'steps'->'deploy_asset')
FROM agent_definitions
WHERE type='asset-deployer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

## Migration 324 record-only (applied out-of-band 2026-08-06)

```bash
git add docs/agent_docs/sql_for_agents/324_asset_deployer_passes_asset_id.sql
git commit docs/agent_docs/sql_for_agents/324_asset_deployer_passes_asset_id.sql -m "..."
./scripts/migration/run-migrations.sh --record-only 324_asset_deployer_passes_asset_id.sql \
  --note 'applied out-of-band 2026-08-06 under GUC marker v1.0.1259+; verified 2026-08-22: live input_fields carries asset_id; bak_asset_deployer_20260806 present'
./scripts/migration/run-migrations.sh   # dry-run: 324 must no longer be pending
```
Gotcha: commit EXACTLY the applied bytes first — the ledger stores the file md5.

## 235 deletion (Phase 3)

Binary check (the action's own error literal as marker, both replicas + control):
```bash
for p in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name); do
  echo "$p:";
  kubectl -n ai-persona-system exec ${p#pod/} -- sh -c \
    'grep -aq "retract_asset_files requires a database handle" /proc/1/exe && echo MARKER_PRESENT || echo MARKER_ABSENT;
     grep -aq "retract_asset_filesZZZnotreal" /proc/1/exe && echo CONTROL_BAD || echo CONTROL_OK';
done
```
Dispatch: generalise `staged_component_build/scripts/RETRACT_gaswholesalers_logo_jpg.sh`
(dry-run default; `ARM=1` adds step_overrides `{retract:{dry_run:false}}` — propagation
UNVERIFIED for this shape, fails safe to dry-run). ⚠ its audit query column is
`occurred_at`, not `created_at`. Wire check per site is the CONTROL PAIR:
```bash
curl -s -o /dev/null -w '%{http_code}\n' https://<domain>/assets/images/logo.jpg   # want 404
curl -s -o /dev/null -w '%{http_code}\n' https://<domain>/assets/images/logo.png   # want 200
```
Use a generous `--max-time`; a `000` here is usually your own request rate.
