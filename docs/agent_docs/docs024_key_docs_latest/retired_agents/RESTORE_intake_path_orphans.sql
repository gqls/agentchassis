-- RESTORE: intake-orchestrator + site-classifier (retired 2026-08-02, late)
--
-- NOT a migration. This file lives outside docs/agent_docs/sql_for_agents/ on
-- purpose: the migration runner applies EVERY pending file in that directory, so
-- parking this there would silently un-retire both agents on the next --apply.
-- Run it by hand, deliberately, or not at all.
--
-- Run with:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db < RESTORE_intake_path_orphans.sql
--
-- These two are a PAIR. intake-orchestrator's workflow spawns site-classifier
-- (step spawn_classifier / call_classifier); site-classifier is reachable from
-- nothing else. Restoring the orchestrator alone gives you a workflow whose
-- classifier step resolves to an inactive agent. Restore BOTH or neither.

-- ---------------------------------------------------------------------------
-- CASE 1 — the rows are still present (this is how they were retired). One line.
-- ---------------------------------------------------------------------------
UPDATE agent_definitions
SET is_active  = true,
    deleted_at = NULL,
    updated_at = now()
WHERE type IN ('intake-orchestrator','site-classifier');

-- Verify: expect 2 rows, is_active=t, deleted_at NULL.
SELECT type, id, version, is_active, deleted_at
FROM agent_definitions
WHERE type IN ('intake-orchestrator','site-classifier') ORDER BY type;

-- ---------------------------------------------------------------------------
-- CASE 2 — the rows were later PHYSICALLY deleted. Reimport from the JSON backup.
-- ---------------------------------------------------------------------------
-- Do NOT retype default_config by hand. Load
-- BACKUP_2026-08-02_intake_path_orphans.json (a JSON array of 2 full rows) and
-- let Postgres do the mapping:
--
--   \set backup `cat BACKUP_2026-08-02_intake_path_orphans.json`
--   INSERT INTO agent_definitions
--   SELECT * FROM jsonb_populate_recordset(NULL::agent_definitions, :'backup'::jsonb)
--   ON CONFLICT (id) DO NOTHING;
--
-- Then confirm the config survived the round trip rather than trusting the row
-- count — a truncated paste produces an agent that looks restored and is not:
--
--   SELECT type,
--          length(default_config::text) AS cfg_bytes,   -- expect 4795 and 1854
--          (SELECT count(*) FROM jsonb_object_keys(default_config->'workflow'->'steps')) AS steps
--   FROM agent_definitions WHERE type IN ('intake-orchestrator','site-classifier');
--
-- Expected: intake-orchestrator 13 steps — call_briefer, call_builder,
-- call_classifier, call_rerender, complete, fetch_available_builders,
-- fetch_questionnaire, hitl_confirm_type, hitl_review_brief, spawn_briefer,
-- spawn_builder, spawn_classifier, spawn_rerender.
-- site-classifier 2 steps — classify_site, complete.

-- ---------------------------------------------------------------------------
-- BEFORE YOU RESTORE — what you are turning back on, and what has moved under it
-- ---------------------------------------------------------------------------
-- 1. THE MENU IT READS HAS BEEN GUTTED. fetch_available_builders runs
--    query_agent_definitions with type_pattern '%-builder' and active_only, then
--    spawns via agent_type_field 'confirmed_type.recommended_builder'. On
--    2026-08-02 that menu went 7 -> 2: only pageflow-builder and report-builder
--    are active. The HITL dynamic_select will therefore offer a human two
--    options, one of which (report-builder) is not a site builder at all.
--    Un-retire the builders you actually want offered first — see
--    RESTORE_multipage-website-builder.sql.
--
-- 2. ITS ENTRY POINT IS A HUMAN, NOT AN AGENT. Nothing in the database spawns
--    intake-orchestrator. It is triggered by an operator publishing to Kafka
--    topic system.agent.generic.requests with
--    {"action":"orchestrate","config":{"agent_type":"intake-orchestrator"}} —
--    see scripts/initial_messages/090_new_build/ (last touched 2026-06-21).
--    Restoring the rows does not restore that habit, and the current operator
--    script (scripts/initial_messages/020_build_pipeline/082_submit_domain_unified.sh,
--    2026-07-30) deliberately routes to domain-submitter instead.
--
-- 3. YOU PROBABLY WANT THE REPLACEMENT, NOT THIS. The live path is
--    domain-submitter -> domain-research-classifier -> domain-strategist ->
--    vertical-exemplar-researcher -> site-design-planner -> build-briefing-agent
--    -> build-site-planner. If the reason you are reading this file is "how do I
--    build a site", use 082_submit_domain_unified.sh. Restore these two only if
--    you specifically want the old orchestrator-spawns-a-builder shape back.
