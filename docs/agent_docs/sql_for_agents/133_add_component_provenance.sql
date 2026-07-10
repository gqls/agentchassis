-- 0NN_add_component_provenance.sql — unblock tool creation (schema drift).
-- DRAFT 2026-07-08. Renumber 0NN.
--
-- ROOT CAUSE (run 00688389 / orchestration 9f93a988, agent_error_log
-- 2026-07-08 16:14:44):
--   step save_tool failed: create_tool_component: ERROR: column
--   "source_agent_type" of relation "content_components" does not exist
--   (SQLSTATE 42703)
-- generate_tool_html SUCCEEDED; the failure is one step later. The binary
-- shipped ahead of its migration — create_tool_component_action.go says so
-- itself, directly above the INSERT:
--   "source_* = creation provenance (NNN_add_component_provenance.sql),
--    mirroring knowledge_base's pair ... apply that migration before this
--    binary deploys."
-- Latent since ~2026-05-16 (last 'generated' tool component); component-creator
-- inserts a different column set, so nothing else broke. The Task-3 proof run
-- was simply the first caller in two months.
--
-- REUSE FIRST — before applying this, look for the canonical file:
--   find ~/projects/agentchassis -name '*provenance*'
--   git -C ~/projects/agentchassis grep -l add_component_provenance
-- If it exists, apply THAT (and diff it against this draft). Use this only if
-- the file the comment names was never written; then commit it to the repo so
-- the code's reference resolves.
--
-- No snapshot_agent() call: this touches a DATA table, not agent_definitions.
-- Additive + nullable + idempotent, so it is safe with the current binary
-- (which needs the columns) and with any older binary (which ignores them).

BEGIN;

-- 1) The mirror source must exist; its types are copied, never guessed.
DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n
    FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'knowledge_base'
      AND column_name IN ('source_agent_type', 'source_orchestration_id');
    IF n <> 2 THEN
        RAISE EXCEPTION
          'knowledge_base provenance pair not found (found % of 2) — cannot mirror types. Paste \d knowledge_base before proceeding.', n;
    END IF;
END $$;

-- 2) Add the pair to content_components with knowledge_base's exact types.
DO $$
DECLARE
    t_agent text;
    t_orch  text;
BEGIN
    SELECT format_type(a.atttypid, a.atttypmod) INTO t_agent
    FROM pg_attribute a
    WHERE a.attrelid = 'public.knowledge_base'::regclass
      AND a.attname = 'source_agent_type' AND NOT a.attisdropped;

    SELECT format_type(a.atttypid, a.atttypmod) INTO t_orch
    FROM pg_attribute a
    WHERE a.attrelid = 'public.knowledge_base'::regclass
      AND a.attname = 'source_orchestration_id' AND NOT a.attisdropped;

    EXECUTE format('ALTER TABLE content_components ADD COLUMN IF NOT EXISTS source_agent_type %s', t_agent);
    EXECUTE format('ALTER TABLE content_components ADD COLUMN IF NOT EXISTS source_orchestration_id %s', t_orch);

    RAISE NOTICE 'content_components provenance added: source_agent_type %, source_orchestration_id % (mirrored from knowledge_base)', t_agent, t_orch;
END $$;

COMMENT ON COLUMN content_components.source_agent_type IS
  'Creation provenance: agent type that created this component (mirrors knowledge_base.source_agent_type). NULL for rows predating this migration.';
COMMENT ON COLUMN content_components.source_orchestration_id IS
  'Creation provenance: orchestration that created this component (mirrors knowledge_base.source_orchestration_id). NULL for rows predating this migration.';

-- 3) Guard.
DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n
    FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'content_components'
      AND column_name IN ('source_agent_type', 'source_orchestration_id');
    IF n <> 2 THEN
        RAISE EXCEPTION 'content_components provenance columns missing after add (found %)', n;
    END IF;
END $$;

COMMIT;

-- Verify after apply:
--   SELECT column_name, data_type, is_nullable
--   FROM information_schema.columns
--   WHERE table_name = 'content_components' AND column_name LIKE 'source_%'
--   ORDER BY column_name;
--
-- Then RE-RUN the Task-3 proof (the function name is still free — the failed
-- run inserted nothing; the component INSERT was the first statement):
--   ./drafts/085_TRIGGER_toolgen_gamesdesign_v1.sh
-- Expect: tool component + page + nav, THEN compose_plan -> write_plan ->
-- index_plan, i.e. a doc_plans row with source='tool-generator' (Task-3 proof),
-- and the new component stamped source_agent_type='tool-generator'.
--
-- Watch for (nothing else drifts — the action's other INSERTs were checked
-- column-by-column against production schema and match):
--   pages(id, site_id, name, url, title, page_type, status, build_status,
--         nav_order, meta_description)                       OK
--   page_components(page_id, component_id, position, slot_name,
--         rendered_html, content_data, build_status)          OK
--   site_work_items(...)                                      OK
--
-- Rollback:
--   ALTER TABLE content_components DROP COLUMN IF EXISTS source_agent_type;
--   ALTER TABLE content_components DROP COLUMN IF EXISTS source_orchestration_id;
--   (Only with a binary that does not reference them — the current one does.)
