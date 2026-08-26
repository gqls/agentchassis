-- 636_aiao_two_tool_templates_token_collisions_ROLLBACK.sql
--
-- Restores the byte-exact pre-636 `html_template` for the two aiao tool components
-- from the migration_backups rows 636 wrote.
--
-- ⚠ THE STATE THIS RETURNS TO IS 5 MEASURED CONTRAST DEFECTS — four the site audit
-- reports and one it CANNOT see (`.result-value`, inside a result panel that is hidden
-- until the visitor runs the calculator). Roll back only to isolate a worse regression
-- elsewhere, and say which.
--
-- ⚠ TEMPLATES ONLY, SO THIS IS NOT COMPLETE ON ITS OWN. Placements keep whatever HTML
-- they last rendered. If 636 has already been propagated, the OLD colours return only
-- after another page-scoped `template_changed` rerender of
-- /tools/automation-savings-estimator/index.html and /tools/build-vs-buy-analyzer/index.html.
-- Both pages are `rebuild_policy='generic'`, so that rerender is the ordinary one — do
-- NOT reach for 625's owned-page dance here.

BEGIN;

UPDATE content_components cc
SET html_template = (b.old_value->>'html_template'), updated_at = now()
FROM migration_backups b
WHERE b.migration_name='636_aiao_two_tool_templates_token_collisions'
  AND b.target_table='content_components' AND b.target_id = cc.id::text;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM content_components cc JOIN migration_backups b
    ON b.target_id=cc.id::text AND b.migration_name='636_aiao_two_tool_templates_token_collisions'
   WHERE cc.html_template = (b.old_value->>'html_template');
  IF n <> 2 THEN RAISE EXCEPTION 'rollback 636: % of 2 templates byte-identical to backup', n; END IF;
  RAISE NOTICE 'rollback 636 OK: both templates restored byte-exact. Live pages keep 636''s colours until re-rendered.';
END $$;

COMMIT;
