-- 676 ROLLBACK — strip the three opt-ins by exact inverse replace, each asserted.
BEGIN;
DO $r$
DECLARE cfg jsonb; tpl text; t record;
BEGIN
  FOR t IN SELECT * FROM (VALUES ('build-site-planner','plan_site'), ('content-gap-planner','plan_gaps'), ('visual-designer','design')) AS x(ty, step) LOOP
    SELECT default_config INTO cfg FROM agent_definitions
     WHERE type=t.ty AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    tpl := cfg->'workflow'->'steps'->t.step->'config'->>'prompt_template';
    IF position(E'\n\n## {{.build_standard}}\n' IN tpl) = 0 THEN
      RAISE EXCEPTION '676 ROLLBACK: % carries no opt-in section — nothing to strip (already rolled back, or the template drifted)', t.ty;
    END IF;
    tpl := replace(tpl, E'\n\n## {{.build_standard}}\n', '');
    IF position('{{.build_standard}}' IN tpl) > 0 THEN
      RAISE EXCEPTION '676 ROLLBACK: % still names the placeholder after the strip — a second/edited occurrence exists, remove by hand', t.ty;
    END IF;
    cfg := jsonb_set(cfg, ('{workflow,steps,' || t.step || ',config,prompt_template}')::text[], to_jsonb(tpl));
    UPDATE agent_definitions SET default_config = cfg
     WHERE type=t.ty AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  END LOOP;
END $r$;
DELETE FROM schema_migrations WHERE filename='676_build_standard_optins_HOLD.sql';
COMMIT;
