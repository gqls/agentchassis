-- 520 ROLLBACK — remove the two prompt rules and disarm the birth guard.
-- Hand-run only (uppercase sidecar). Restores the rule-20→Structure seam and
-- deletes the enforce key (absent = unarmed = pre-519 behaviour exactly).
BEGIN;
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,generate_tool_html,config,prompt_template}',
      to_jsonb(regexp_replace(
        default_config#>>'{workflow,steps,generate_tool_html,config,prompt_template}',
        E'cannot tell that from a tool that is broken\\.\n21\\..*?\n\n## Structure',
        E'cannot tell that from a tool that is broken.\n\n## Structure')),
      false),
    updated_at = now()
WHERE type='tool-generator' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

UPDATE agent_definitions
SET default_config = default_config #- '{workflow,steps,save_tool,config,enforce_instance_scope}',
    updated_at = now()
WHERE type='tool-generator' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE p text;
BEGIN
  SELECT default_config#>>'{workflow,steps,generate_tool_html,config,prompt_template}'
    INTO p FROM agent_definitions
   WHERE type='tool-generator' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF p LIKE '%exactly one IIFE%' OR p LIKE '%22. Give every element%' THEN
    RAISE EXCEPTION '520 ROLLBACK verify: rules still present';
  END IF;
END $$;
COMMIT;
