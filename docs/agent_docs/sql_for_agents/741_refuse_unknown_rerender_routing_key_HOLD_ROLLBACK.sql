-- 741_refuse_unknown_rerender_routing_key_HOLD_ROLLBACK.sql
--
-- Undoes 741: the refusal door comes back out of the front of the page-rerender
-- workflow, the gate returns to reading spec.reason alone, and the write-door CHECK
-- is dropped.
--
-- ⚠ WHAT ROLLING BACK COSTS. It restores bugs_open/440: an unrecognised routing key
-- goes back to completing GREEN having changed nothing. It does NOT un-stamp anything
-- — `spec.routing_reason` keeps being written by the converted producers (phase 2,
-- live since 12:13 on 2026-09-03), and the restored gate simply ignores it, which is
-- the pre-741 behaviour and is safe. So this is a real rollback of the REFUSAL, not
-- of the split.
--
-- WHEN YOU WOULD RUN IT: the refusal is parking items it should not. The diagnostic
-- to run FIRST, because it distinguishes the two causes and they have different fixes:
--
--   SELECT spec->>'routing_reason' AS key, count(*)
--     FROM site_work_items
--    WHERE item_type='page_rerender' AND status='needs_human_review'
--      AND error LIKE '%not in the sections-rerender vocabulary%'
--    GROUP BY 1 ORDER BY 2 DESC;
--
-- A key that SHOULD be in the vocabulary -> do not roll back; add the value to
-- RerenderSectionReasons and cut a migration (that is the designed path). A NULL or
-- empty key appearing there -> the guard clause has lost its `== null` / `== ''`
-- disjunct, which is the one failure mode that would park the whole legacy population;
-- roll back at once, then fix the clause.

BEGIN;

SET LOCAL lock_timeout = '5s';

DO $$
DECLARE n int; ss text;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='page-rerender' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '741 ROLLBACK REFUSED: expected exactly 1 active page-rerender row, found %', n;
  END IF;

  SELECT default_config #>> '{workflow,start_step}' INTO ss FROM agent_definitions
   WHERE type='page-rerender' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF ss <> 'check_routing_key_known' THEN
    RAISE EXCEPTION '741 ROLLBACK REFUSED: start_step is %, not check_routing_key_known — 741 is not applied, or another lane has restructured this workflow since', ss;
  END IF;

  PERFORM snapshot_agent('page-rerender',
    '741_..._HOLD_ROLLBACK.sql: pre-rollback (refusal door live)');
END $$;

-- 1. Front door back to the gate itself.
UPDATE agent_definitions
   SET default_config = jsonb_set(default_config, '{workflow,start_step}',
                                  '"check_rerender_mode"'::jsonb),
       updated_at = now()
 WHERE type='page-rerender' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- 2. Gate back to the five-value spec.reason clause (livespec.CheckRerenderModeConditionClause()).
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config, '{workflow,steps,check_rerender_mode,config,condition}',
         to_jsonb($cond$input_data.spec.reason == 'image_landed' OR input_data.spec.reason == 'section_data_resolved' OR input_data.spec.reason == 'cta_links_stale' OR input_data.spec.reason == 'template_changed' OR input_data.spec.reason == 'literal_markdown'$cond$::text)),
       updated_at = now()
 WHERE type='page-rerender' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- 3. Remove the three steps 741 added.
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config, '{workflow,steps}',
         (default_config #> '{workflow,steps}')
           - 'check_routing_key_known' - 'refuse_unknown_routing_key' - 'complete_refused'),
       updated_at = now()
 WHERE type='page-rerender' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- 4. The write door.
ALTER TABLE site_work_items
  DROP CONSTRAINT IF EXISTS chk_page_rerender_routing_reason_vocabulary;

DO $$
DECLARE cfg jsonb;
BEGIN
  SELECT default_config #> '{workflow}' INTO cfg FROM agent_definitions
   WHERE type='page-rerender' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF cfg #>> '{start_step}' <> 'check_rerender_mode' THEN
    RAISE EXCEPTION '741 ROLLBACK VERIFY FAILED: start_step is %', cfg #>> '{start_step}';
  END IF;
  IF cfg #> '{steps,check_routing_key_known}' IS NOT NULL
     OR cfg #> '{steps,refuse_unknown_routing_key}' IS NOT NULL
     OR cfg #> '{steps,complete_refused}' IS NOT NULL THEN
    RAISE EXCEPTION '741 ROLLBACK VERIFY FAILED: one of the three added steps is still present';
  END IF;
  IF cfg #>> '{steps,check_rerender_mode,config,condition}' LIKE '%routing_reason%' THEN
    RAISE EXCEPTION '741 ROLLBACK VERIFY FAILED: the gate condition still tests routing_reason';
  END IF;
  -- The steps the workflow needs in order to work at all must have survived the
  -- key-deletion in step 3 — `-` on a jsonb object takes a key name, and a typo
  -- there deletes nothing, but a wrong key name deletes the wrong step silently.
  IF cfg #> '{steps,check_rerender_mode}' IS NULL
     OR cfg #> '{steps,render_page}' IS NULL
     OR cfg #> '{steps,rerender_sections}' IS NULL THEN
    RAISE EXCEPTION '741 ROLLBACK VERIFY FAILED: a pre-existing step was removed';
  END IF;
  IF EXISTS (SELECT 1 FROM pg_constraint
              WHERE conname='chk_page_rerender_routing_reason_vocabulary'
                AND conrelid='site_work_items'::regclass) THEN
    RAISE EXCEPTION '741 ROLLBACK VERIFY FAILED: the CHECK constraint is still present';
  END IF;

  RAISE NOTICE '741 ROLLED BACK: gate reads spec.reason only; refusal door removed; CHECK dropped. bugs_open/440 is REOPENED in production — an unknown routing key assembles silently again.';
END $$;

COMMIT;
