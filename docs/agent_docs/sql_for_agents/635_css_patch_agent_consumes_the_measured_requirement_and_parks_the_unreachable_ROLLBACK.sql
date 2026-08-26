-- 635_..._ROLLBACK.sql - remove the check_repair_surface gate, the park step and
-- the measured-requirement prompt block; restore 616's shape exactly.
--
-- The inverse edit, not a snapshot restore: a restore would discard anything
-- another lane has legitimately changed on this shared row since. The snapshot,
-- if wanted: '635_css_patch_agent_consumes_the_measured_requirement: pre-update'.

DO $probe$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM agent_definitions
        WHERE type = 'css-patch-agent'
          AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
          AND default_config #> '{workflow,steps}' ? 'check_repair_surface'
    ) THEN
        RAISE EXCEPTION '390/635 ROLLBACK: not applied';
    END IF;
END $probe$;

BEGIN;

SELECT snapshot_agent('css-patch-agent', '635_..._ROLLBACK: pre-restore');

DO $restore$
DECLARE
    v_block  text := $tpl$## Audit Finding{{if .input_data.spec.override_requirement}}
## The declaration you must BEAT (measured in the browser, proven by removal)
{{with .input_data.spec.winning_rule}}{{if .selector}}Winner: `{{.selector}}` ({{.surface}}), declaring `{{.decl}}`{{if .important}}, marked !important{{end}}.{{else}}Winner: the element's own style attribute, declaring `{{.decl}}`.{{end}}
{{end}}{{with .input_data.spec.override_requirement}}{{.why}}.
Your selector must reach specificity ({{.min_specificity_text}}){{if .strictly_greater}} and EXCEED it strictly - an equal-specificity rule loses on source order{{end}}.
{{if .needs_important}}The corrected property MUST carry !important to match the winner.{{else}}Do NOT use !important - it is not needed to win here; this measured section supersedes the general !important guidance below.{{end}}
{{end}}{{with .input_data.spec.override_example}}A selector the platform has VERIFIED satisfies this requirement: `{{.}}`
{{end}}{{end}}$tpl$;
    v_rows   int;
    v_steps  jsonb;
    v_prompt text;
BEGIN
    SELECT count(*) INTO v_rows FROM agent_definitions
     WHERE type='css-patch-agent' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF v_rows <> 1 THEN
        RAISE EXCEPTION '390/635 ROLLBACK: expected exactly ONE live row, found %', v_rows;
    END IF;

    SELECT default_config #> '{workflow,steps}' INTO v_steps
      FROM agent_definitions
     WHERE type='css-patch-agent' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

    IF v_steps #>> '{check_base_integrity,config,then_step}' <> 'check_repair_surface' THEN
        RAISE EXCEPTION '390/635 ROLLBACK drift: check_base_integrity.then_step is %, expected check_repair_surface',
            v_steps #>> '{check_base_integrity,config,then_step}';
    END IF;
    v_prompt := v_steps #>> '{plan_css_fix,config,prompt_template}';
    IF position(v_block in v_prompt) = 0 THEN
        RAISE EXCEPTION '390/635 ROLLBACK drift: the inserted block is not present verbatim - the prompt has been edited since 635; roll back by hand';
    END IF;

    UPDATE agent_definitions
       SET default_config =
           jsonb_set(
             (default_config #- '{workflow,steps,check_repair_surface}')
                             #- '{workflow,steps,mark_cascade_unreachable}',
             '{workflow,steps,check_base_integrity,config,then_step}',
             to_jsonb('plan_css_fix'::text), false
           ),
           updated_at = NOW()
     WHERE type='css-patch-agent' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    UPDATE agent_definitions
       SET default_config = jsonb_set(
             default_config,
             '{workflow,steps,plan_css_fix,config,prompt_template}',
             to_jsonb(replace(default_config #>> '{workflow,steps,plan_css_fix,config,prompt_template}', v_block, '## Audit Finding')),
             false
           )
     WHERE type='css-patch-agent' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

    GET DIAGNOSTICS v_rows = ROW_COUNT;
    IF v_rows <> 1 THEN
        RAISE EXCEPTION '390/635 ROLLBACK: UPDATE touched % rows', v_rows;
    END IF;
END $restore$;

DO $verify$
DECLARE v_steps jsonb; v_prompt text;
BEGIN
    SELECT default_config #> '{workflow,steps}' INTO v_steps
      FROM agent_definitions
     WHERE type='css-patch-agent' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF (v_steps ? 'check_repair_surface') OR (v_steps ? 'mark_cascade_unreachable') THEN
        RAISE EXCEPTION '390/635 ROLLBACK verify: a 635 step survived';
    END IF;
    IF v_steps #>> '{check_base_integrity,config,then_step}' <> 'plan_css_fix' THEN
        RAISE EXCEPTION '390/635 ROLLBACK verify: check_base_integrity not restored';
    END IF;
    v_prompt := v_steps #>> '{plan_css_fix,config,prompt_template}';
    IF position('{{if .input_data.spec.override_requirement}}' in v_prompt) > 0 THEN
        RAISE EXCEPTION '390/635 ROLLBACK verify: the block survived in the prompt';
    END IF;
    IF position('## Audit Finding' in v_prompt) = 0
       OR position('can never win on position' in v_prompt) = 0
       OR position('"css_added"' in v_prompt) = 0 THEN
        RAISE EXCEPTION '390/635 ROLLBACK verify: prompt damaged (616 shape not restored)';
    END IF;
    RAISE NOTICE '390/635 ROLLBACK: restored to 616''s shape';
END $verify$;

COMMIT;
