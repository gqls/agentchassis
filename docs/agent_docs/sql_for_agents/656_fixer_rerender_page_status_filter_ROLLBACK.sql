-- ROLLBACK for 656 — remove the page-status filter from the fixer's
-- create_rerender query.
--
-- WHEN YOU WOULD RUN THIS: the filter is excluding pages you need re-rendered
-- (an archived page being deliberately repaired before un-archiving would be the
-- realistic case), or the rewrite mangled the query.
--
-- ⚠ NOTE WHAT IT RESTORES. Without the filter the fixer files re-renders for
-- ARCHIVED pages again — re-publishing retired pages and making a retraction
-- self-undoing (bugs_open/098). That is the state 656 exists to end, so this is a
-- deliberate step backwards, not a neutral undo.
--
-- ⚠ AND IT WILL MAKE THE DAILY AUDITOR FIRE. livespec declares this query as
-- containing `p.status = 'active'`, so `live-declaration-drift-check` will exit 1
-- at 07:00 naming the missing fragment. That is correct behaviour: if you mean to
-- keep the rollback, remove the fragment from the Declaration in the same commit,
-- or the alarm is permanent and will be learned-to-ignore.

BEGIN;

DO $$
DECLARE q text;
BEGIN
  SELECT default_config #>> '{workflow,steps,create_rerender,config,query}' INTO q
    FROM agent_definitions
   WHERE type = 'component-template-fixer' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF q IS NULL THEN
    RAISE EXCEPTION '656_ROLLBACK: no live component-template-fixer create_rerender query';
  END IF;
  IF q NOT LIKE '%p.status = ''active''%' THEN
    RAISE NOTICE '656_ROLLBACK: filter already absent — nothing to do';
    RETURN;
  END IF;

  PERFORM snapshot_agent('component-template-fixer',
    '656_ROLLBACK pre-image: removing the page-status filter');

  UPDATE agent_definitions
     SET default_config = jsonb_set(default_config,
           '{workflow,steps,create_rerender,config,query}',
           to_jsonb(replace(q, ' AND p.status = ''active''', ''))),
         updated_at = NOW()
   WHERE type = 'component-template-fixer' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
END $$;

DO $$
DECLARE q text;
BEGIN
  SELECT default_config #>> '{workflow,steps,create_rerender,config,query}' INTO q
    FROM agent_definitions
   WHERE type = 'component-template-fixer' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF q LIKE '%p.status = ''active''%' THEN
    RAISE EXCEPTION '656_ROLLBACK: the filter is still present';
  END IF;
  -- The surgical inverse must have taken exactly its own clause and nothing else.
  IF q NOT LIKE '%rebuild_policy IS DISTINCT FROM ''owned''%'
     OR q NOT LIKE '%''reason'',''template_changed''%'
     OR q NOT LIKE '%NOT EXISTS%' THEN
    RAISE EXCEPTION '656_ROLLBACK: the removal took more than its own clause with it';
  END IF;
END $$;

COMMIT;
