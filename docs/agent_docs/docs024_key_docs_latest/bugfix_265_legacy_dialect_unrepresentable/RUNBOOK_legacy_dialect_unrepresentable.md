# RUNBOOK — bug 265

## Census (the number this bug is about)
```sql
SELECT count(*), max(created_at)::date FROM content_components WHERE input_schema ? 'properties';
-- provenance — the query that changed the design; run it BEFORE choosing a seam
SELECT created_from, source_agent_type, count(*), sum((input_schema ? 'properties')::int) legacy
  FROM content_components GROUP BY 1,2 ORDER BY 3 DESC;
-- top-level key census (does anything else legitimately use "properties"?)
SELECT k, count(*) FROM content_components, jsonb_object_keys(input_schema) k
 WHERE jsonb_typeof(input_schema)='object' GROUP BY 1 ORDER BY 2 DESC;
```

## Probe a migration without applying it (COMMIT → SELECT + ROLLBACK)
```bash
F=docs/agent_docs/sql_for_agents/437_content_components_refuse_legacy_input_schema_dialect.sql
sed -e 's/^COMMIT;$/SELECT …converted rows…; ROLLBACK;/' $F > $SCRATCH/437_probe.sql
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At -v ON_ERROR_STOP=1 < $SCRATCH/437_probe.sql
# then READ BACK that nothing changed: legacy count, pg_constraint, to_regclass(backup)
```
Gotcha: psql `-At` still prints command tags (BEGIN/DO/UPDATE 3…) — filter `^{` for JSON rows.

## Apply 437 (hand, scoped) and record
```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 \
  < docs/agent_docs/sql_for_agents/437_content_components_refuse_legacy_input_schema_dialect.sql
./scripts/migration/run-migrations.sh --record-only docs/agent_docs/sql_for_agents/437_content_components_refuse_legacy_input_schema_dialect.sql --note "<what you verified>"
```
Never a bare `--apply` — it takes every pending file (other lanes').

## Induce the refusal (the only proof a zero census is evidence)
```sql
BEGIN;
INSERT INTO content_components (name, html_template, function, input_schema)
VALUES ('zz','<section></section>','zz-scratch-265','{"type":"object","properties":{"x":{"type":"string"}}}');
-- expect: ERROR … violates check constraint "chk_input_schema_no_legacy_dialect"
ROLLBACK;
```

## Go
```bash
go build ./platform/... && go test ./platform/orchestration/datahelpers/ -run 'LegacyDialect|SchemaContentFields' -v
```
Fixture re-capture (never re-type): the probe above → `platform/orchestration/datahelpers/testdata/legacy_dialect_conversion_437.json`.

## Rollback
`437_…_ROLLBACK.sql` — drops the constraint FIRST, restores 3 rows by id from
`content_components_bak_20260816_265_legacy_dialect`, verifies 3.
