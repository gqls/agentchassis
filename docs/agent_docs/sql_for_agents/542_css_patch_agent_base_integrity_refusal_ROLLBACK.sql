-- 542_css_patch_agent_base_integrity_refusal_ROLLBACK.sql
--
-- Reverses 542: removes the base-integrity gate, the three status-stamping steps
-- and the refusal terminal, restores every rewired edge to its pre-542 value,
-- restores load_current_css's query to the three-column form, and drops the
-- file_shrink_floor opt-in.
--
-- ⚠ WHAT ROLLING BACK COSTS, stated plainly: css-patch-agent goes back to
-- appending onto a possibly-empty base and deploying it wholesale over the site's
-- real stylesheet, and its refusals and failures go back to reading `complete`.
-- That is the state that clobbered nine sites across three waves. Roll back only
-- if the gate itself is misfiring on healthy sites, and prefer lowering the floor
-- (a one-key jsonb_set) to removing the mechanism.

BEGIN;

SELECT snapshot_agent('css-patch-agent',
  '542_ROLLBACK: pre-revert');

-- restore the edges first, so no step is orphaned mid-transaction
UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           jsonb_set(
             jsonb_set(
               jsonb_set(
                 jsonb_set(
                   default_config,
                   '{workflow,steps,check_has_css,config,then_step}',
                   to_jsonb('plan_css_fix'::text)
                 ),
                 '{workflow,steps,check_has_css,config,else_step}',
                 to_jsonb('complete_no_css'::text)
               ),
               '{workflow,steps,load_current_css,error_step}',
               to_jsonb('complete_no_css'::text)
             ),
             '{workflow,steps,plan_css_fix,error_step}',
             to_jsonb('complete_error'::text)
           ),
           '{workflow,steps,save_css_to_db,error_step}',
           to_jsonb('complete_error'::text)
         ),
         '{workflow,steps,deploy_css,error_step}',
         to_jsonb('complete_error'::text)
       ),
       updated_at = NOW()
 WHERE type = 'css-patch-agent'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- restore the original three-column query
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,load_current_css,config,query}',
         to_jsonb('SELECT ct.id::text as theme_id, ct.css_content, ct.name as theme_name FROM css_themes ct JOIN style_collections sc ON sc.css_theme_id = ct.id JOIN sites s ON s.style_collection_id = sc.id WHERE s.id = $1::uuid'::text)
       ),
       updated_at = NOW()
 WHERE type = 'css-patch-agent'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- drop the added steps and the opt-in key
UPDATE agent_definitions
   SET default_config = default_config
         #- '{workflow,steps,check_base_integrity}'
         #- '{workflow,steps,mark_base_unsafe}'
         #- '{workflow,steps,mark_no_css}'
         #- '{workflow,steps,mark_step_failed}'
         #- '{workflow,steps,complete_refused}'
         #- '{workflow,steps,deploy_css,config,file_shrink_floor}',
       updated_at = NOW()
 WHERE type = 'css-patch-agent'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE
    v_steps jsonb;
BEGIN
    SELECT default_config #> '{workflow,steps}'
      INTO v_steps
      FROM agent_definitions
     WHERE type = 'css-patch-agent'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_steps ? 'check_base_integrity' OR v_steps ? 'mark_base_unsafe'
       OR v_steps ? 'mark_no_css' OR v_steps ? 'mark_step_failed'
       OR v_steps ? 'complete_refused' THEN
        RAISE EXCEPTION '198/542 ROLLBACK: a 542 step survives';
    END IF;
    IF v_steps #> '{deploy_css,config}' ? 'file_shrink_floor' THEN
        RAISE EXCEPTION '198/542 ROLLBACK: file_shrink_floor survives';
    END IF;
    IF v_steps #>> '{check_has_css,config,then_step}' <> 'plan_css_fix'
       OR v_steps #>> '{check_has_css,config,else_step}' <> 'complete_no_css'
       OR v_steps #>> '{load_current_css,error_step}' <> 'complete_no_css'
       OR v_steps #>> '{plan_css_fix,error_step}' <> 'complete_error'
       OR v_steps #>> '{save_css_to_db,error_step}' <> 'complete_error'
       OR v_steps #>> '{deploy_css,error_step}' <> 'complete_error' THEN
        RAISE EXCEPTION '198/542 ROLLBACK: an edge was not restored';
    END IF;
    IF position('css_len' in (v_steps #>> '{load_current_css,config,query}')) > 0 THEN
        RAISE EXCEPTION '198/542 ROLLBACK: query still measures css_len';
    END IF;

    RAISE NOTICE '198/542 ROLLBACK: verified — pre-542 shape restored (and the clobber door is open again)';
END $$;

COMMIT;
