# RUNBOOK — requires-backend section gate (bugs_open/276)

## DB access
```
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```

## Re-check migration numbering before EVERY new file write (this tree churns fast)
```
ls docs/agent_docs/sql_for_agents/ | grep -E "^4[0-9]{2}_" | sort | tail -20
```

## Live row identity (re-check id/version before touching — another session may have moved it)
```sql
SELECT id, version, updated_at FROM agent_definitions
WHERE type IN ('content-gap-planner','build-site-planner','site-planner')
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

## Byte-exact current query text (paste into the pre-state gate literal)
```sql
SELECT default_config#>>'{workflow,steps,load_available_components,config,query}'
FROM agent_definitions WHERE type='content-gap-planner'
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- swap type + step name (load_components for build-site-planner) as needed
```

## Tag / capability census (re-run fresh, cite the timestamp)
```sql
SELECT name, function, component_level FROM content_components
WHERE COALESCE(semantic_tags,'[]'::jsonb) ? 'requires-backend';

SELECT domain, deploy_config->'capabilities' AS caps FROM sites
WHERE COALESCE(deploy_config->'capabilities','[]'::jsonb) ? 'backend';
```

## Dispatch volume (30d) — the evidence for which call site matters most
```sql
SELECT owner_agent_type, count(*), max(created_at) AS most_recent
FROM orchestration_states
WHERE owner_agent_type IN ('site-planner','content-gap-planner','build-site-planner')
  AND created_at > now() - interval '30 days'
GROUP BY owner_agent_type;
-- for "0, ever" on site-planner: drop the created_at filter entirely, don't just widen it
```

## site_record.site_id binding proof (content-gap-planner)
```sql
SELECT count(*) AS total,
       count(*) FILTER (WHERE collected_data#>>'{site_record,site_id}' IS NOT NULL) AS nonnull
FROM orchestration_states WHERE owner_agent_type='content-gap-planner'
  AND created_at > now() - interval '30 days';
-- want total == nonnull
```

## Disagreeing-pair proof (run BEFORE apply against current text, and AFTER apply against live)
Real site ids: `relojistas.com` = `ecf15e75-a966-4900-bcb0-1c85f689dbfd` (backend capability),
`gamesdesign.co.uk` = `e33263f4-74f8-494f-b191-546845dbbddf` (static control).
```sql
-- OLD (paste current/pre-fix query text), then NEW (paste post-fix query text), each run
-- once per site id with the id substituted for $1 (or via PREPARE ... EXECUTE, untyped $1,
-- to also prove the text-param -> uuid coercion the Go driver path relies on).
-- Compare: md5(string_agg(row::text, ',' ORDER BY category, name)) old vs new.
```

## Applying a migration (each file is self-contained BEGIN..COMMIT)
```
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -f - < docs/agent_docs/sql_for_agents/418_content_gap_planner_requires_backend_gate.sql
```
GOTCHA: `psql -f -` needs the file piped via stdin redirect (`< file`), not as a `-c` string —
the file has a `DO $$ ... $$` block with embedded semicolons that `-c` will split wrongly.

## Council submission
```
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <submission.json>
```
This submission is config-only (no Go bytes change) — pass `FORCE=1` (stated reason: "config
ships as a docs-path migration", the same reason migration 406 used for the sibling tool-side
gate). Save the printed `SUBMISSION_CORR` into NOTES immediately.

## Coverage check who-owns before touching bugs_open/276 further in a resumed session
```
python3 scripts/who-owns.py 276
```
