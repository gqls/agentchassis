# RUNBOOK — bugfix 220

## Read the live dispatcher mapping (the defect's home)
```sql
SELECT jsonb_pretty(default_config->'workflow'->'steps'->'process_item'->'config'->'sub_workflow'->'steps'->'call_handler')
FROM agent_definitions
WHERE type='build-dispatch-loop' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```
Gotcha: the call_handler step is nested under `process_item.config.sub_workflow.steps`,
not under `workflow.steps` directly. For agents where you don't know the nesting, use
the LATERAL jsonb_each form (see NOTES) — `jsonb_path_query` with `?(...)` filters
fails on this Postgres with "syntax error at or near (" for `@.keyvalue()`.

## Find every dispatcher that maps a given path (live, not seeds)
```sql
SELECT type FROM agent_definitions
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND default_config::text LIKE '%current_item.spec.page_name%';
```
Gotcha: this is a VALUE substring search, which works; the `jsonb::text LIKE
'%"k":"v"%'` KEY:VALUE form does NOT (jsonb renders a space after the colon). Induce a
non-zero control before trusting a zero (`input_data.spec.page_name` → 5 rows).

## The item-type census (who disagrees column vs spec)
Bug file § CONTRIB 2026-08-08 has the query shape and the measured answer
(only unbuilt_internal_link; re-run before shipping — the census moves daily).

## Apply + record the migration (after the Go is committed)
```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  < docs/agent_docs/sql_for_agents/340_unbuilt_link_dispatch_authoritative_page_id.sql
./scripts/migration/run-migrations.sh --record-only 340_unbuilt_link_dispatch_authoritative_page_id.sql \
  --note "<what you verified>"
```
Gotcha: record in the same motion — an idempotent file probes `ok` and reads as
pending for ever (mig 335 lesson). The `_ROLLBACK` sidecar is excluded by SIDECAR_RE.

## Prove the deploy (after the next fleet roll — owner runs make release)
```bash
kubectl -n ai-persona-system exec <chassis-pod> -- sh -c 'strings /app/agent-chassis | grep -c "authoritative_page_id"'
# expect >0 on EVERY replica; this change removes no string, so there is no negative
# control — corroborate with behaviour (below), not the tag.
```

## Behavioural acceptance
Find a site with a live link to a never-deployed page (`pages.deployed_at IS NULL` +
an href to its url in any rendered component), fire the one-shot improvement loop,
then assert: the minted `unbuilt_internal_link` item's `result` names the TARGET's
file, `pages.deployed_at` is now set on the target, and `result._verification.status`
= `verified`. The wrong outcome (pre-fix) deploys the CONTAINER's file and leaves the
target NULL.
