-- NNN_fix_researcher_spawn_columns.sql  (§B4 incident: pod spawned as GENERIC)
--
-- ROOT CAUSE: the §B4 seed populated default_config (+ category/status from a
-- donor) but NOT the columns the SPAWNER consumes. getAgentDefinition
-- (spawn_actions.go:2120) SELECTs image_repository, image_tag, command,
-- resources, capabilities, topics, health_config, env_vars, is_active,
-- idle_timeout_seconds — and filters is_active = true. With command NULL the
-- container spec gets Command: nil → the image's DEFAULT ENTRYPOINT runs →
-- the generic chassis service boots (type "generic", topic
-- system.agent.generic.process, group generic-group — exactly the observed
-- pod) and never reads the injected AGENT_TYPE/KAFKA_TOPICS env. The
-- dispatcher's call to the per-instance job topic goes unheard; the item
-- stays claimed. (Liveness for spawn = is_active boolean, NOT status.)
--
-- FIX: copy the spawn-consumed infrastructure columns from a PROVEN donor —
-- page-build-handler, spawned by the same dispatcher daily. Deliberately NOT
-- copied: capabilities and topics (semantic/per-agent; spawned agents get
-- topics via env), default_config (ours), category/status (already set).
-- Run the comparison SELECT first; if the donor's topics/capabilities look
-- load-bearing for spawn on your \d output, pause and paste.

-- 0) comparison (adjust to \d agent_definitions if names differ)
SELECT type, is_active, image_repository, image_tag, command, idle_timeout_seconds,
       (resources IS NULL) AS res_null, (health_config IS NULL) AS health_null,
       (env_vars IS NULL) AS env_null
FROM agent_definitions
WHERE type IN ('vertical-exemplar-researcher','page-build-handler')
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL ORDER BY type;

BEGIN;

SELECT snapshot_agent('vertical-exemplar-researcher',
  'populate spawn-consumed columns (image/tag/command/resources/health/env/idle/is_active) from donor page-build-handler — seed omitted them; pod booted as generic via default entrypoint');

UPDATE agent_definitions v
SET image_repository     = d.image_repository,
    image_tag            = d.image_tag,
    command              = d.command,
    resources            = d.resources,
    health_config        = d.health_config,
    env_vars             = d.env_vars,
    idle_timeout_seconds = d.idle_timeout_seconds,
    is_active            = true,
    updated_at           = now()
FROM agent_definitions d
WHERE v.type = 'vertical-exemplar-researcher'
  AND COALESCE(v.is_snapshot,false) = false AND v.deleted_at IS NULL
  AND d.type = 'page-build-handler'
  AND COALESCE(d.is_snapshot,false) = false AND d.deleted_at IS NULL;

-- verify — expect command populated, is_active t, image matching the donor
SELECT type, is_active, image_repository, image_tag, command
FROM agent_definitions
WHERE type = 'vertical-exemplar-researcher'
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

COMMIT;

-- CLEANUP + RE-RUN (after applying):
-- 1) delete the stuck generic job/pod:
--    kubectl -n ai-persona-system get jobs | grep vertical-exemplar
--    kubectl -n ai-persona-system delete job <job-name>
-- 2) reset the item (or let claimed-item-timeout do it — no evidence rows
--    exist, so its evidence check will reset with attempts+1):
--    UPDATE site_work_items SET status='triaged'
--    WHERE item_type='needs_vertical_research' AND status='claimed'
--      AND site_id='5fe8785b-223d-41a3-88ee-c07187622381';
-- 3) the pump re-dispatches within ~30s; watch:
--    kubectl -n ai-persona-system logs -f -l agent-type=vertical-exemplar-researcher --tail=200
--    (expect: Agent starting type=vertical-exemplar-researcher, per-instance
--     job topic + <type>-group-<id8> consumer group)

-- ── REVERT (semantic; the snapshot above is the byte-exact path) ────────────
-- BEGIN;
-- SELECT snapshot_agent('vertical-exemplar-researcher','revert spawn columns to NULL');
-- UPDATE agent_definitions SET image_repository=NULL, image_tag=NULL,
--   command=NULL, resources=NULL, health_config=NULL, env_vars=NULL,
--   idle_timeout_seconds=NULL, updated_at=now()
-- WHERE type='vertical-exemplar-researcher'
--   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- COMMIT;
