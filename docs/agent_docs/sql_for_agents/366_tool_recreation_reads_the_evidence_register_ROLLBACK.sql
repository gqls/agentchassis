-- 366_tool_recreation_reads_the_evidence_register_ROLLBACK.sql
--
-- Removes the "Verified facts" section 366 inserted into
-- tool-recreation-handler's `recreate_tool` prompt, by cutting from the section
-- heading up to (and not including) the `## Design Context` anchor it was
-- inserted before. Everything else in the template is untouched.
--
-- This is a text cut rather than a restore-from-snapshot on purpose: another
-- session may have amended a different part of the same prompt since 366 ran,
-- and restoring the whole pre-366 template would silently discard their change.
-- The snapshot 366 took remains available if a full restore is ever wanted.

BEGIN;

SELECT snapshot_agent('tool-recreation-handler',
  '366_ROLLBACK: pre-removal of the verified-facts section') AS snapshot_id;

UPDATE agent_definitions SET default_config = jsonb_set(
  default_config,
  '{workflow,steps,recreate_tool,config,prompt_template}',
  to_jsonb(
    left(default_config #>> '{workflow,steps,recreate_tool,config,prompt_template}',
         position('## Verified facts' in
                  default_config #>> '{workflow,steps,recreate_tool,config,prompt_template}') - 1)
    || substr(default_config #>> '{workflow,steps,recreate_tool,config,prompt_template}',
              position('## Design Context' in
                       default_config #>> '{workflow,steps,recreate_tool,config,prompt_template}'))
  )
)
WHERE type = 'tool-recreation-handler'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND position('## Verified facts' in
               default_config #>> '{workflow,steps,recreate_tool,config,prompt_template}') > 0;

DO $$
DECLARE t text;
BEGIN
  SELECT default_config #>> '{workflow,steps,recreate_tool,config,prompt_template}'
    INTO t FROM agent_definitions
   WHERE type = 'tool-recreation-handler'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF position('## Verified facts' in t) <> 0 THEN
    RAISE EXCEPTION 'the verified-facts section is still present after rollback';
  END IF;
  IF position('## Design Context' in t) = 0 THEN
    RAISE EXCEPTION 'the rollback removed the ## Design Context anchor itself';
  END IF;
END $$;

COMMIT;
