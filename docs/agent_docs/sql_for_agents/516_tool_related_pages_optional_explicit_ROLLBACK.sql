-- 516 ROLLBACK — restore the UNMARKED `related_pages` wire on both tool build
-- steps, i.e. put migration 211's shape back.
--
-- WHAT ROLLING BACK RESTORES, stated plainly so the decision is informed: the
-- unmarked wire resolves the same path when it exists AND falls through to the
-- whole-tree search when it does not — which is `bugs_open/330`, the defect 516
-- exists to close. Roll back only if the MARKED form is causing harm (e.g. a
-- consumer turns out to need the searched value after all), never as tidy-up.
--
-- Hand-run. Not swept by the runner (a _ROLLBACK sidecar is refused client-side).

BEGIN;

SELECT snapshot_agent('tool-generator',
                      '516_ROLLBACK_tool_related_pages_optional_explicit: pre-rollback');
SELECT snapshot_agent('tool-deployer',
                      '516_ROLLBACK_tool_related_pages_optional_explicit: pre-rollback');

DO $$
DECLARE
    tgt record;
    cfg jsonb;
BEGIN
    FOR tgt IN
        SELECT * FROM (VALUES
            ('tool-generator', 'save_tool'),
            ('tool-deployer',  'deploy_tool')
        ) AS v(agent_type, step_name)
    LOOP
        SELECT default_config #> ARRAY['workflow','steps',tgt.step_name,'config'] INTO cfg
          FROM agent_definitions
         WHERE type = tgt.agent_type AND is_active
           AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

        IF cfg IS NULL THEN
            RAISE EXCEPTION '516 ROLLBACK: no live %.% step config', tgt.agent_type, tgt.step_name;
        END IF;
        IF NOT (cfg ? 'related_pages?') THEN
            RAISE EXCEPTION '516 ROLLBACK: %.% does not carry related_pages? — nothing to roll back '
                '(516 was not applied, or someone has already reverted it)',
                tgt.agent_type, tgt.step_name;
        END IF;

        UPDATE agent_definitions
           SET default_config = jsonb_set(
                   default_config,
                   ARRAY['workflow','steps',tgt.step_name,'config','related_pages'],
                   to_jsonb('input_data.spec.related_pages'::text),
                   true),
               updated_at = NOW()
         WHERE type = tgt.agent_type AND is_active
           AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

        UPDATE agent_definitions
           SET default_config = default_config
                   #- ARRAY['workflow','steps',tgt.step_name,'config','related_pages?'],
               updated_at = NOW()
         WHERE type = tgt.agent_type AND is_active
           AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    END LOOP;
END $$;

DO $$
DECLARE
    tgt record;
    got text;
    marked boolean;
BEGIN
    FOR tgt IN
        SELECT * FROM (VALUES
            ('tool-generator', 'save_tool'),
            ('tool-deployer',  'deploy_tool')
        ) AS v(agent_type, step_name)
    LOOP
        SELECT default_config #>> ARRAY['workflow','steps',tgt.step_name,'config','related_pages'],
               default_config #> ARRAY['workflow','steps',tgt.step_name,'config'] ? 'related_pages?'
          INTO got, marked
          FROM agent_definitions
         WHERE type = tgt.agent_type AND is_active
           AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

        IF got IS DISTINCT FROM 'input_data.spec.related_pages' OR marked THEN
            RAISE EXCEPTION '516 ROLLBACK VERIFY FAILED: %.% reads related_pages=% , related_pages? present=%',
                tgt.agent_type, tgt.step_name, COALESCE(got, '<null>'), marked;
        END IF;
    END LOOP;
    RAISE NOTICE '516 ROLLBACK OK: both steps back to the unmarked 211 wire (330 is reachable again)';
END $$;

COMMIT;
