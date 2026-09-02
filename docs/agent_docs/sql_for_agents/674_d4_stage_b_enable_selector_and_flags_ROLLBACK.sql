-- 674_d4_stage_b_enable_selector_and_flags_ROLLBACK.sql — removes stage B's config half.
-- Selector text restored to the post-657 shape (md5 d29807313... asserted after); the two
-- flags are DELETED — correct for THESE keys, stated against the 658 trap: an absent
-- honour_spend_governor reads false via Go's `.(bool)` (off — today's behaviour), whereas
-- 658's knobs fell back to Go defaults 50/20 on deletion and had to be restored explicitly.
-- Refuses while the governor is ENABLED: disable deliberately first, never yank under load.

BEGIN;

DO $$
DECLARE q text; n int;
BEGIN
  SELECT default_config#>>'{workflow,steps,find_dispatchable_site,config,query}' INTO q
  FROM agent_definitions
  WHERE type='build-pipeline-trigger' AND is_active AND COALESCE(is_snapshot,false)=false
    AND deleted_at IS NULL ORDER BY version DESC LIMIT 1;
  IF q IS NULL OR position('governor_work_class_map' in q) = 0 THEN
    RAISE EXCEPTION '674 ROLLBACK REFUSED: selector does not carry the governor clause — 674 is not applied.';
  END IF;
  n := (length(q) - length(replace(q, 'governor_work_class_map', ''))) / length('governor_work_class_map');
  IF n <> 1 THEN
    RAISE EXCEPTION '674 ROLLBACK REFUSED: clause appears % times — hand-edited since apply, investigate.', n;
  END IF;
  IF EXISTS (SELECT 1 FROM governor_config WHERE id=1 AND enabled) THEN
    RAISE EXCEPTION '674 ROLLBACK REFUSED: governor is ENABLED — UPDATE governor_config SET enabled=false first, deliberately.';
  END IF;
END $$;

SELECT snapshot_agent('build-pipeline-trigger', '674_..._ROLLBACK.sql: pre-restore');
SELECT snapshot_agent('build-dispatch-loop', '674_..._ROLLBACK.sql: pre-restore');

UPDATE agent_definitions
SET default_config = jsonb_set(default_config,
  '{workflow,steps,find_dispatchable_site,config,query}',
  to_jsonb(replace(
    default_config#>>'{workflow,steps,find_dispatchable_site,config,query}',
    'AND NOT COALESCE((SELECT gc.enabled AND COALESCE(m.llm_bearing, true) AND gs.shed_level >= CASE COALESCE(m.class, ''maintenance'') WHEN ''maintenance'' THEN 1 WHEN ''build'' THEN 2 ELSE 3 END FROM governor_config gc JOIN governor_state gs ON gs.id = 1 LEFT JOIN governor_work_class_map m ON m.item_type = wi.item_type WHERE gc.id = 1), false) AND NOT EXISTS (SELECT 1 FROM site_work_items active WHERE active.site_id = wi.site_id AND active.status = ''claimed'')',
    'AND NOT EXISTS (SELECT 1 FROM site_work_items active WHERE active.site_id = wi.site_id AND active.status = ''claimed'')'
  ))
)
WHERE type='build-pipeline-trigger' AND is_active AND COALESCE(is_snapshot,false)=false
  AND deleted_at IS NULL;

UPDATE agent_definitions
SET default_config = (default_config
  #- '{workflow,steps,load_items,config,honour_spend_governor}')
  #- '{workflow,steps,process_item,config,sub_workflow,steps,claim,config,honour_spend_governor}'
WHERE type='build-dispatch-loop' AND is_active AND COALESCE(is_snapshot,false)=false
  AND deleted_at IS NULL;

DO $$
DECLARE m text;
BEGIN
  SELECT md5(default_config#>>'{workflow,steps,find_dispatchable_site,config,query}') INTO m
  FROM agent_definitions
  WHERE type='build-pipeline-trigger' AND is_active AND COALESCE(is_snapshot,false)=false
    AND deleted_at IS NULL ORDER BY version DESC LIMIT 1;
  IF m <> 'd29807313a8f6ed543a541c35c1626c4' THEN
    RAISE EXCEPTION '674 ROLLBACK VERIFY: selector md5 % is not the post-657 text — restore incomplete', m;
  END IF;
  PERFORM 1 FROM agent_definitions
  WHERE type='build-dispatch-loop' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
    AND default_config::text LIKE '%honour_spend_governor%';
  IF FOUND THEN RAISE EXCEPTION '674 ROLLBACK VERIFY: a flag survived deletion'; END IF;
  RAISE NOTICE '674 ROLLBACK OK: selector byte-restored (md5 d29807313), flags gone (absent = off in Go).';
END $$;

COMMIT;
