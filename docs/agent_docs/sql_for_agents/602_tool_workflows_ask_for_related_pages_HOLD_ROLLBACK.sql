-- ROLLBACK for 602 — remove the picker from both tool workflows and restore the
-- direct route to the saving step.
--
-- What this restores: the state of 2026-08-24, i.e. a tool whose request names
-- no `related_pages` gets NO cross-mentions and records
-- `tool_crosslink_not_emitted:no_related_pages` at info. That was measured at
-- 13 of 13 tool births over 08-22→08-24. Roll back only if the picker is doing
-- harm — a weak or wrong page choice is the shape to look for — not merely
-- because it is quiet: a quiet picker on a site with no topical match is it
-- working (prompt rule 5).
--
-- It does NOT touch `related_pages?` (migration 516) and does NOT touch the
-- reader in the binary. `related_pages_fallback` simply goes unwired again,
-- which is inert.

BEGIN;

SELECT snapshot_agent('tool-generator', '602_ROLLBACK: pre-rollback');
SELECT snapshot_agent('tool-deployer', '602_ROLLBACK: pre-rollback');

UPDATE agent_definitions SET default_config =
    jsonb_set(
        (default_config
            #- '{workflow,steps,load_site_page_names}'
            #- '{workflow,steps,suggest_related_pages}'
            #- '{workflow,steps,save_tool,config,related_pages_fallback?}'),
        '{workflow,steps,generate_tool_html,next_step}', '"save_tool"'::jsonb)
WHERE type='tool-generator' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

UPDATE agent_definitions SET default_config =
    jsonb_set(
        (default_config
            #- '{workflow,steps,load_site_page_names}'
            #- '{workflow,steps,suggest_related_pages}'
            #- '{workflow,steps,deploy_tool,config,related_pages_fallback?}'),
        '{workflow,steps,ensure_site_record,next_step}', '"deploy_tool"'::jsonb)
WHERE type='tool-deployer' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE
    r record;
BEGIN
    FOR r IN
        SELECT type,
               default_config->'workflow'->'steps' AS steps,
               CASE type WHEN 'tool-generator' THEN 'save_tool' ELSE 'deploy_tool' END AS saver,
               CASE type WHEN 'tool-generator' THEN 'generate_tool_html' ELSE 'ensure_site_record' END AS anchor
          FROM agent_definitions
         WHERE type IN ('tool-generator','tool-deployer') AND is_active
           AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
    LOOP
        IF r.steps ? 'suggest_related_pages' OR r.steps ? 'load_site_page_names' THEN
            RAISE EXCEPTION '602 ROLLBACK: % still carries a picker step', r.type;
        END IF;
        IF r.steps->r.anchor->>'next_step' <> r.saver THEN
            RAISE EXCEPTION '602 ROLLBACK: %.%.next_step is %, expected % — the workflow is now BROKEN, not merely un-rolled-back',
                r.type, r.anchor, COALESCE(r.steps->r.anchor->>'next_step','<absent>'), r.saver;
        END IF;
        IF (r.steps->r.saver->'config') ? 'related_pages_fallback?' THEN
            RAISE EXCEPTION '602 ROLLBACK: %.% still carries the fallback wire', r.type, r.saver;
        END IF;
        -- 516 must survive the rollback untouched.
        IF NOT ((r.steps->r.saver->'config') ? 'related_pages?') THEN
            RAISE EXCEPTION '602 ROLLBACK: %.% lost migration 516''s related_pages? wire — restore from the snapshot', r.type, r.saver;
        END IF;
    END LOOP;
    RAISE NOTICE '602 ROLLBACK: both tool workflows route straight to their saving step again; 516 intact.';
END $$;

COMMIT;
