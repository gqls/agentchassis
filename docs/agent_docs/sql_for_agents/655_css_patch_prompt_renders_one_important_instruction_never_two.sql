-- 655: the css-patch prompt renders ONE !important instruction, never two
--
-- bugs_open/390, follow-up to 616+635, decided on the pre-registered rule in the
-- lane's NOTES §(p)/(q): "if >=2 attributed repairs keep !important under
-- needs_important:false, write the fence migration". Sample as of 2026-08-26
-- evening: EVERY attributed repair so far kept !important against the measured
-- block's explicit "Do NOT" (remortgagecalculator 9b2b2ce9; garden-tools
-- fd933538, and the batch that followed - counts in the council submission).
-- Mechanism: 616's general guidance ("mark ONLY the single property ... as
-- !important") sits near the END of the prompt; 635's measured block with its
-- "Do NOT use !important ... supersedes the general guidance below" sits near
-- the START. The model follows the later instruction. "Supersedes" adjudicated
-- by the model is a comment, not a control - so make the contradiction
-- UNREPRESENTABLE: fence 616's three bullets on the ABSENCE of
-- override_requirement, and shorten 635's else-branch sentence whose
-- "guidance below" would otherwise dangle. Exactly one of the two instructions
-- can now render. needs_important:true keeps its MUST-carry branch; the
-- unattributed/legacy path is byte-identical to today's.
--
-- Literals below were SLICED OUT OF THE LIVE ROW's prompt_template
-- (2026-08-26), not retyped - the twice-bitten lesson of 616.
-- Drift anchor for any future migration: THIS file's shape, not 635's.

DO $mig$
DECLARE
  v_general      text := $g$- IMPORTANT: the declaration you must beat is usually NOT in the stylesheet above. Each page carries its own <style> block, and the platform emits that block inside <main>, always AFTER this file is linked in <head>. Your appended rule is therefore always earlier in source order, so it can never win on position: it wins only by higher specificity, or by !important.
- Measured across 40 completed repairs on 7 sites (2026-08-25, bugs_open/390): the winning declaration sat in the page block in 39 of them and in this stylesheet in 0, and it out-specified the audited selector in 33 - typically a section-scoped `.section-class .element-class` beating an audited `TAG.element-class`.
- So: repeat the audited selector exactly as the finding states it, and mark ONLY the single property you are correcting as !important, for example `color: #123456 !important`. No other property gets !important, and no other property changes.$g$;
  v_sentence_old text := $s$Do NOT use !important - it is not needed to win here; this measured section supersedes the general !important guidance below.$s$;
  v_sentence_new text := $t$Do NOT use !important - it is not needed to win here: a selector meeting the requirement above wins without it.$t$;
  v_fence_open   text := '{{if not .input_data.spec.override_requirement}}';
  v_fence_close  text := '{{end}}';
  v_635_fence    text := '{{if .input_data.spec.override_requirement}}';
  v_id uuid; v_prompt text; v_n int;
BEGIN
  SELECT count(*) INTO v_n FROM agent_definitions
   WHERE type='css-patch-agent' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF v_n <> 1 THEN RAISE EXCEPTION 'expected exactly ONE live css-patch-agent row, found %', v_n; END IF;

  SELECT id, default_config->'workflow'->'steps'->'plan_css_fix'->'config'->>'prompt_template'
    INTO v_id, v_prompt FROM agent_definitions
   WHERE type='css-patch-agent' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF position(v_fence_open in v_prompt) > 0 THEN
    RAISE EXCEPTION '655 already applied: not-fence present; refusing to double-apply'; END IF;
  IF (length(v_prompt) - length(replace(v_prompt, v_general, ''))) / length(v_general) <> 1 THEN
    RAISE EXCEPTION 'drift: the 616 passage is not present exactly once - re-anchor before applying'; END IF;
  IF (length(v_prompt) - length(replace(v_prompt, v_sentence_old, ''))) / length(v_sentence_old) <> 1 THEN
    RAISE EXCEPTION 'drift: the 635 else-sentence is not present exactly once - re-anchor before applying'; END IF;
  IF position(v_635_fence in v_prompt) = 0 THEN
    RAISE EXCEPTION 'drift: 635''s own fence is missing - this migration builds on 635''s shape'; END IF;

  UPDATE agent_definitions
     SET default_config = jsonb_set(default_config,
           '{workflow,steps,plan_css_fix,config,prompt_template}',
           to_jsonb(replace(replace(v_prompt, v_general, v_fence_open || v_general || v_fence_close),
                            v_sentence_old, v_sentence_new))),
         updated_at = now()
   WHERE id = v_id;
  GET DIAGNOSTICS v_n = ROW_COUNT;
  IF v_n <> 1 THEN RAISE EXCEPTION 'UPDATE touched % rows, expected exactly 1', v_n; END IF;

  SELECT default_config->'workflow'->'steps'->'plan_css_fix'->'config'->>'prompt_template'
    INTO v_prompt FROM agent_definitions WHERE id = v_id;
  IF position(v_fence_open || v_general || v_fence_close in v_prompt) = 0 THEN
    RAISE EXCEPTION 'post: fenced 616 passage absent after apply'; END IF;
  IF position(v_sentence_new in v_prompt) = 0 THEN RAISE EXCEPTION 'post: new sentence absent'; END IF;
  IF position(v_sentence_old in v_prompt) > 0 THEN RAISE EXCEPTION 'post: old sentence survived'; END IF;
  RAISE NOTICE 'migration 655 applied: prompt_template now % chars', length(v_prompt);
END $mig$;
