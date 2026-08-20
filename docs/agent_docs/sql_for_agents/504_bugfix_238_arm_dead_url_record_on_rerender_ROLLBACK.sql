-- FILE: docs/agent_docs/sql_for_agents/504_bugfix_238_arm_dead_url_record_on_rerender_ROLLBACK.sql
--
-- Disarm bugs_open/238's RECORD-ONLY dead-URL emit on the re-render path.
-- Run BY HAND — the migration runner never applies a ROLLBACK sidecar.
--
-- Sets the key to FALSE rather than removing it, deliberately and for the same
-- reason 380's sidecar does: an explicit false is a visible decision in the
-- config, whereas an absent key is indistinguishable from "this seam was never
-- wired here", and the next reader cannot tell a deliberate disarm from an
-- oversight.
--
-- WHEN TO REACH FOR THIS. The record-only emit refuses nothing and blocks no
-- page, so the only real reason is VOLUME: `dead_url_control` items arriving
-- faster than they can be worked. Note the churn edge before concluding that is
-- what you are seeing — the item key embeds the sorted dead-field list, so a
-- PARTIAL repair mints a second item alongside the parked first one. Two items
-- for one slot is that edge, not a runaway.
--
-- WHAT YOU KEEP. The Go half stays in the binary and is inert with the flag
-- false. `resolution.DeadURLSlots` still populates and the pre-existing
-- per-section Error log still fires, so you keep the signal in the logs and lose
-- only the durable, queryable work item.
--
-- WHAT YOU DO NOT UNDO. Items already filed are NOT deleted by this file. A
-- record outlives the switch that produced it: each one names a real page and
-- slot serving a dead control, and deleting them would destroy the only durable
-- evidence of damage this estate has for the ungated class. Work them or close
-- them with evidence; do not sweep them.

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent('page-rerender', '504_ROLLBACK_bugfix_238: pre-disarm');

DO $$
DECLARE
    v_rows int;
BEGIN
    SELECT count(*) INTO v_rows
      FROM agent_definitions
     WHERE type = 'page-rerender' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF v_rows <> 1 THEN
        RAISE EXCEPTION '238/504 ROLLBACK: expected exactly 1 live page-rerender row, found % — target that row by id, do not update by type', v_rows;
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
           default_config,
           '{workflow,steps,rerender_sections,config,record_dead_url_controls}',
           'false'::jsonb,
           true),
       updated_at = now()
 WHERE type = 'page-rerender' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- Verify the disarm, treating absence as failure (IS NOT FALSE, not <> true).
DO $$
DECLARE
    v_armed boolean;
BEGIN
    SELECT (default_config #>> '{workflow,steps,rerender_sections,config,record_dead_url_controls}')::boolean
      INTO v_armed
      FROM agent_definitions
     WHERE type = 'page-rerender' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_armed IS NOT FALSE THEN
        RAISE EXCEPTION '238/504 ROLLBACK verify FAILED: record_dead_url_controls reads % (want explicit false)', COALESCE(v_armed::text, '(absent)');
    END IF;
    RAISE NOTICE '238/504 ROLLBACK verify OK: record_dead_url_controls is explicitly false; the Go half stays in the binary and is inert';
END $$;

COMMIT;

-- ---------------------------------------------------------------------------
-- REPORT what was already filed, so disarming is not mistaken for undoing.
--
--   SELECT s.domain, swi.status, swi.item_key, swi.created_at
--     FROM site_work_items swi JOIN sites s ON s.id = swi.site_id
--    WHERE swi.item_type = 'dead_url_control'
--    ORDER BY swi.created_at DESC;
