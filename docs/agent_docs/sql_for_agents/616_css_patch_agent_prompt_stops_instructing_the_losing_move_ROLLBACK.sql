-- 616_..._ROLLBACK.sql - restore 318's cascade bullet on css-patch-agent.
--
-- The inverse replace, not a snapshot restore: 616 changed exactly one key, and a
-- snapshot restore would also discard anything another lane has legitimately
-- changed on this row since. If you want the snapshot instead, it is the row
-- stamped '616_css_patch_agent_prompt_stops_instructing_the_losing_move: pre-update'.
--
-- NOTE what rolling back reinstates: an instruction that is FALSE for this agent's
-- situation and produces a rule that cannot take effect (bugs_open/390, 33 of 40
-- sampled completions). Roll back only to unblock something else, and say so.

DO $probe$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM agent_definitions
        WHERE type = 'css-patch-agent'
          AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
          AND default_config #>> '{workflow,steps,plan_css_fix,config,prompt_template}'
              LIKE '%can never win on position%'
    ) THEN
        RAISE EXCEPTION '390/616 ROLLBACK: not applied - the correction is not present';
    END IF;
END $probe$;

BEGIN;

SELECT snapshot_agent('css-patch-agent',
  '616_..._ROLLBACK: pre-restore');

UPDATE agent_definitions
   SET default_config = jsonb_set(
           default_config,
           '{workflow,steps,plan_css_fix,config,prompt_template}',
           to_jsonb(
               replace(
                   default_config #>> '{workflow,steps,plan_css_fix,config,prompt_template}',
$new$- IMPORTANT: the declaration you must beat is usually NOT in the stylesheet above. Each page carries its own <style> block, and the platform emits that block inside <main>, always AFTER this file is linked in <head>. Your appended rule is therefore always earlier in source order, so it can never win on position: it wins only by higher specificity, or by !important.
- Measured across 40 completed repairs on 7 sites (2026-08-25, bugs_open/390): the winning declaration sat in the page block in 39 of them and in this stylesheet in 0, and it out-specified the audited selector in 33 - typically a section-scoped `.section-class .element-class` beating an audited `TAG.element-class`.
- So: repeat the audited selector exactly as the finding states it, and mark ONLY the single property you are correcting as !important, for example `color: #123456 !important`. No other property gets !important, and no other property changes.$new$,
$old$- Rely on the cascade: an appended rule with the same or higher specificity overrides the earlier declaration. Repeat the offending selector exactly as it appears above (or more specifically) so your override wins.$old$
               )
           ),
           false
       ),
       updated_at = NOW()
 WHERE type = 'css-patch-agent'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $verify$
DECLARE v_prompt text;
BEGIN
    SELECT default_config #>> '{workflow,steps,plan_css_fix,config,prompt_template}'
      INTO v_prompt
      FROM agent_definitions
     WHERE type = 'css-patch-agent'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF position('can never win on position' in v_prompt) > 0 THEN
        RAISE EXCEPTION '390/616 ROLLBACK verify: the correction survived';
    END IF;
    IF position('Repeat the offending selector exactly as it appears above' in v_prompt) = 0 THEN
        RAISE EXCEPTION '390/616 ROLLBACK verify: 318 bullet was not restored';
    END IF;
    IF position('"css_added"' in v_prompt) = 0 THEN
        RAISE EXCEPTION '390/616 ROLLBACK verify: output contract damaged';
    END IF;
    RAISE NOTICE '390/616 ROLLBACK: restored';
END $verify$;

COMMIT;
