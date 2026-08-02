-- RESTORE: multipage-website-builder (retired 2026-08-02)
--
-- NOT a migration. This file lives outside docs/agent_docs/sql_for_agents/ on
-- purpose: the migration runner applies EVERY pending file in that directory, so
-- parking this there would silently un-retire the agent on the next --apply.
-- Run it by hand, deliberately, or not at all.
--
-- Run with:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db < RESTORE_multipage-website-builder.sql

-- ---------------------------------------------------------------------------
-- CASE 1 — the rows are still present (this is how it was retired). One line.
-- ---------------------------------------------------------------------------
UPDATE agent_definitions
SET is_active  = true,
    deleted_at = NULL,
    updated_at = now()
WHERE type = 'multipage-website-builder';

-- Verify: expect 2 rows, is_active=t, deleted_at NULL.
SELECT id, version, is_active, deleted_at
FROM agent_definitions WHERE type = 'multipage-website-builder' ORDER BY created_at;

-- Verify it is back in the builder MENU (this is the query intake-orchestrator
-- actually runs — type_pattern '%-builder' with active_only defaulting true):
SELECT type FROM agent_definitions
WHERE is_active = true AND type LIKE '%-builder' ORDER BY type;

-- ---------------------------------------------------------------------------
-- CASE 2 — the rows were later PHYSICALLY deleted. Reimport from the JSON backup.
-- ---------------------------------------------------------------------------
-- Do NOT retype default_config by hand: it is ~4.4 KB of workflow per row, and a
-- silently-truncated paste produces an agent that looks restored and is not.
-- Load BACKUP_2026-08-02_multipage-website-builder.json (a JSON array of 2 full
-- rows) and let Postgres do the mapping:
--
--   \set backup `cat BACKUP_2026-08-02_multipage-website-builder.json`
--   INSERT INTO agent_definitions
--   SELECT * FROM jsonb_populate_recordset(NULL::agent_definitions, :'backup'::jsonb)
--   ON CONFLICT (id) DO NOTHING;
--
-- Then re-run the two verification SELECTs above, and additionally confirm the
-- config survived the round trip rather than trusting the row count:
--
--   SELECT id,
--          length(default_config::text) AS cfg_bytes,        -- expect 4389 and 4469
--          jsonb_object_keys(default_config->'workflow'->'steps') AS step
--   FROM agent_definitions WHERE type = 'multipage-website-builder';
--
-- Expected 13 steps per row: assemble_site, call_strategist, complete, deploy,
-- ensure_site_record, generate_pages_loop, populate_nav, spawn_content_creator,
-- spawn_deployer, spawn_html_developer, spawn_strategist, sync_pages_to_db,
-- update_timestamps.

-- ---------------------------------------------------------------------------
-- AFTER RESTORING, note what comes back with it
-- ---------------------------------------------------------------------------
-- This agent is the SOLE carrier of extract_and_sync_links, which contains a
-- known success-shaped failure: on an unresolved site_id it returns
-- {"links_extracted": N, "persisted": false} with no error, so a caller reading
-- links_extracted has every reason to think the registry was written
-- (site_db_actions.go:396; audited in bugs_closed/092). That exposure is dormant
-- only because this agent does not run. Reviving the agent re-arms it — fix the
-- site_id check first.
--
-- The completeness floor guarding the same action's reconciliation delete
-- (link_registry_prune_floor.go, bugs_closed/165 site C) is mutation-proven but
-- has never been induced live, because link_registry has never held a row. If
-- this agent is revived, that induction becomes possible and should be run:
-- RUNBOOK_reconciliation_deletes.md § R-B2 transfers.
