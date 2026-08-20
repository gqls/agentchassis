-- FILE: docs/agent_docs/sql_for_agents/502_bugfix_260_arm_mistyped_llm_fields_ROLLBACK.sql
--
-- Reverts 502: sets refuse_mistyped_llm_fields back to false on both of
-- page-content-writer's render_component steps. Hand-run only.
--
-- FALSE, not removed: the Go reader treats absent and false identically
-- (`armed, _ := config[key].(bool)`), so either would work — but leaving the key
-- present at false keeps the decision VISIBLE to a reader of the step config,
-- and an absent key is indistinguishable from a step that was never armed.
--
-- What this does NOT revert: the seam's unconditional hard error, or the
-- unconditional type ENRICHER on an already-failed render. Neither is gated by
-- this key. If the intent is to restore the silent regex fallback, this file is
-- the wrong instrument — that is a code change, and re-adding it would restore
-- the third state bugs_open/260 exists to remove (5 tests fail on the mutation).

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent('page-content-writer', '498_ROLLBACK: pre-revert');

UPDATE agent_definitions
   SET default_config = jsonb_set(
           jsonb_set(
               default_config,
               '{workflow,steps,process_sections_loop,config,sub_workflow,steps,render_section,config,refuse_mistyped_llm_fields}',
               'false'::jsonb,
               true),
           '{workflow,steps,process_sections_loop,config,sub_workflow,steps,render_from_template,config,refuse_mistyped_llm_fields}',
           'false'::jsonb,
           true),
       updated_at = now()
 WHERE type = 'page-content-writer' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE
    v_still_armed int;
BEGIN
    SELECT count(*) FILTER (WHERE k.value->>'action' = 'render_component'
                              AND (k.value->'config'->>'refuse_mistyped_llm_fields')::boolean IS TRUE)
      INTO v_still_armed
      FROM agent_definitions ad,
           LATERAL jsonb_path_query(ad.default_config, 'strict $.**.steps') AS steps,
           LATERAL jsonb_each(steps) AS k
     WHERE ad.type = 'page-content-writer' AND ad.is_active
       AND COALESCE(ad.is_snapshot, false) = false AND ad.deleted_at IS NULL;
    IF v_still_armed <> 0 THEN
        RAISE EXCEPTION '260/502 ROLLBACK: % render_component step(s) still armed', v_still_armed;
    END IF;
    RAISE NOTICE '260/502 ROLLBACK: declared-type refusal disarmed on every render_component step';
END $$;

COMMIT;
