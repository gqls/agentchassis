-- 679 ROLLBACK — strip visual-designer's opt-in by exact inverse replace.
BEGIN;
DO $r$
DECLARE cfg jsonb; tpl text; n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='visual-designer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN RAISE EXCEPTION '679 ROLLBACK: % active rows, want 1', n; END IF;
  SELECT default_config INTO cfg FROM agent_definitions
   WHERE type='visual-designer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  tpl := cfg->'workflow'->'steps'->'design'->'config'->>'prompt_template';
  IF position(E'\n\n## {{.build_standard}}\n' IN tpl) = 0 THEN
    RAISE EXCEPTION '679 ROLLBACK: no opt-in section to strip';
  END IF;
  tpl := replace(tpl, E'\n\n## {{.build_standard}}\n', '');
  IF position('{{.build_standard}}' IN tpl) > 0 THEN
    RAISE EXCEPTION '679 ROLLBACK: placeholder survives the strip — remove by hand';
  END IF;
  cfg := jsonb_set(cfg, '{workflow,steps,design,config,prompt_template}', to_jsonb(tpl));
  UPDATE agent_definitions SET default_config = cfg
   WHERE type='visual-designer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
END $r$;
DELETE FROM schema_migrations WHERE filename='679_build_standard_optin_visual_designer_HOLD.sql';
COMMIT;
