-- Migration NNN — content_components creation provenance.
--
-- Grounded in \d content_components (2026-06-11): the table records the
-- creation MECHANISM (created_from, CHECK manual/generated/adopted/tool/forked)
-- but NOT the creating actor. agent_type / agent_workflow / data_sources are
-- the dynamic-component RENDER hooks, not provenance — do not overload them.
--
-- Column names deliberately MIRROR knowledge_base's existing provenance pair
-- (source_agent_type varchar(100), source_orchestration_id varchar(255)) so
-- one mental model and one query pattern covers both tables. NOTE the
-- near-collision with the unrelated render-hook column `agent_type`: the
-- comments below are load-bearing.
--
-- PRE-CHECKS (run before applying; abort and discuss if either surprises):
--   1. Columns absent (idempotence guard — plain ADD COLUMN errors if present):
--        SELECT column_name FROM information_schema.columns
--        WHERE table_name='content_components' AND column_name LIKE 'source_%';
--      Expect: 0 rows.
--   2. Whether existing creation events already give a partial backfill trail:
--        SELECT event_type, count(*) FROM system_events
--        WHERE event_type ILIKE '%component%' GROUP BY 1;
--        SELECT entity_type, count(*) FROM entity_state_log
--        WHERE entity_type ILIKE '%component%' GROUP BY 1;
--      If rich: a later backfill migration can join on these; this migration
--      does NOT backfill (columns stay NULL for existing rows — honest
--      absence beats guessed provenance).

BEGIN;

ALTER TABLE content_components
    ADD COLUMN source_agent_type       varchar(100),
    ADD COLUMN source_orchestration_id varchar(255);

COMMENT ON COLUMN content_components.source_agent_type IS
    'Creation provenance: agent type that created this row (e.g. tool-generator, tool-deployer). Mirrors knowledge_base.source_agent_type. NOT the render hook — that is the unrelated agent_type column.';
COMMENT ON COLUMN content_components.source_orchestration_id IS
    'Creation provenance: orchestration that created this row. Mirrors knowledge_base.source_orchestration_id. NULL on rows predating this migration.';

COMMIT;

-- No index: the query pattern ("what did orchestration X create") is forensic
-- and rare; a seq scan over ~10⁴ rows is fine. Add a partial index only if a
-- hot path appears.
--
-- SET-ON-INSERT (code changes, same release as this migration):
--   * create_tool_component_action.go  — novel tools: source_agent_type =
--     'tool-generator' (or the executing agent type from the execution
--     context), source_orchestration_id = the orchestration id.
--   * deploy_tool_action.go            — forks: stamp the DEPLOYING
--     orchestration (created_from='forked' already records the mechanism;
--     forked_from records the lineage).
--   Both values come from the execution context the actions already hold —
--   no new plumbing.
--
-- VERIFICATION after the first deploys:
--   SELECT name, created_from, source_agent_type, source_orchestration_id
--   FROM content_components
--   WHERE component_level='tool' ORDER BY created_at DESC LIMIT 10;
