-- FILE: docs/agent_docs/sql_for_agents/380_bugfix_238_arm_dead_url_guard_ROLLBACK.sql
--
-- Disarm the bugs_open/238 dead-URL refusal. Run BY HAND — the runner never
-- applies a ROLLBACK sidecar.
--
-- Sets the key to FALSE rather than removing it, deliberately: an explicit false
-- is a visible decision in the config, whereas an absent key is indistinguishable
-- from "this seam was never wired here" and the next reader cannot tell a
-- deliberate disarm from an oversight.
--
-- Reach for this if the refusal is blocking rebuilds faster than the
-- dead_url_control queue can be worked. The Go half stays in the binary; with
-- the flag false it is inert, and the pre-existing Error log ("URL attribute
-- rendered empty — dead control") still fires, so you keep the signal and lose
-- only the blocking.

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent('page-content-writer', '380_ROLLBACK_bugfix_238: pre-disarm');

UPDATE agent_definitions
   SET default_config = jsonb_set(
           default_config,
           '{workflow,steps,process_sections_loop,config,sub_workflow,steps,render_section,config,refuse_dead_url_controls}',
           'false'::jsonb,
           true),
       updated_at = now()
 WHERE type = 'page-content-writer' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE
    v_armed boolean;
BEGIN
    SELECT (jsonb_path_query_first(default_config,
              '$.**.steps.render_section.config.refuse_dead_url_controls'))::text::boolean
      INTO v_armed
      FROM agent_definitions
     WHERE type = 'page-content-writer' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF v_armed IS NOT FALSE THEN
        RAISE EXCEPTION '238/380 rollback: the flag is not false after the update (%) — aborting rather than reporting a disarm that did not happen', v_armed;
    END IF;
    RAISE NOTICE '238/380: dead-URL refusal DISARMED — the guard is inert, the Error log still fires';
END $$;

COMMIT;
