-- 655 ROLLBACK: restore the unfenced 616 passage and 635's original else-sentence
DO $mig$
DECLARE
  v_general      text := $g$- IMPORTANT: the declaration you must beat is usually NOT in the stylesheet above. Each page carries its own <style> block, and the platform emits that block inside <main>, always AFTER this file is linked in <head>. Your appended rule is therefore always earlier in source order, so it can never win on position: it wins only by higher specificity, or by !important.
- Measured across 40 completed repairs on 7 sites (2026-08-25, bugs_open/390): the winning declaration sat in the page block in 39 of them and in this stylesheet in 0, and it out-specified the audited selector in 33 - typically a section-scoped `.section-class .element-class` beating an audited `TAG.element-class`.
- So: repeat the audited selector exactly as the finding states it, and mark ONLY the single property you are correcting as !important, for example `color: #123456 !important`. No other property gets !important, and no other property changes.$g$;
  v_sentence_old text := $s$Do NOT use !important - it is not needed to win here; this measured section supersedes the general !important guidance below.$s$;
  v_sentence_new text := $t$Do NOT use !important - it is not needed to win here: a selector meeting the requirement above wins without it.$t$;
  v_fence_open   text := '{{if not .input_data.spec.override_requirement}}';
  v_fence_close  text := '{{end}}';
  v_id uuid; v_prompt text; v_n int;
BEGIN
  SELECT count(*) INTO v_n FROM agent_definitions
   WHERE type='css-patch-agent' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF v_n <> 1 THEN RAISE EXCEPTION 'expected exactly ONE live css-patch-agent row, found %', v_n; END IF;
  SELECT id, default_config->'workflow'->'steps'->'plan_css_fix'->'config'->>'prompt_template'
    INTO v_id, v_prompt FROM agent_definitions
   WHERE type='css-patch-agent' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF position(v_fence_open || v_general || v_fence_close in v_prompt) = 0 THEN
    RAISE EXCEPTION 'rollback: fenced passage not found - 655 not applied, or drifted'; END IF;
  IF (length(v_prompt) - length(replace(v_prompt, v_sentence_new, ''))) / length(v_sentence_new) <> 1 THEN
    RAISE EXCEPTION 'rollback: 655 sentence not present exactly once'; END IF;
  UPDATE agent_definitions
     SET default_config = jsonb_set(default_config,
           '{workflow,steps,plan_css_fix,config,prompt_template}',
           to_jsonb(replace(replace(v_prompt, v_fence_open || v_general || v_fence_close, v_general),
                            v_sentence_new, v_sentence_old))),
         updated_at = now()
   WHERE id = v_id;
  GET DIAGNOSTICS v_n = ROW_COUNT;
  IF v_n <> 1 THEN RAISE EXCEPTION 'rollback UPDATE touched % rows', v_n; END IF;
  RAISE NOTICE '655 rolled back';
END $mig$;
