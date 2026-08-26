-- ROLLBACK for 644 — remove the 'image' arm and restore Illustrated Text Block's
-- original sources.
--
-- ⚠ DOES NOT DROP component_expresses(): migrations 591/592/593 depend on it and
-- three live planner menus call it. This restores the FOUR-arm body verbatim.
--
-- ⚠ WHAT A ROLLBACK CANNOT UNDO. Any page planned while 644 was live may have
-- had an Illustrated Text Block selected for it. Reverting the vocabulary does
-- not un-plan those, and reverting `image_url` to `site_assets.image` will make
-- them resolve to the PAGE HERO on their next build — which is the defect 644
-- was written to prevent. Prefer fixing forward. If you must roll back, census
-- first:
--   SELECT s.domain, p.url FROM page_components pc
--     JOIN content_components cc ON cc.id = pc.component_id
--     JOIN pages p ON p.id = pc.page_id JOIN sites s ON s.id = p.site_id
--    WHERE cc.name = 'Illustrated Text Block';
-- Six such rows existed on one site (apis.uk) before 644.

BEGIN;

CREATE OR REPLACE FUNCTION public.component_expresses(p_html_template text, p_input_schema jsonb)
 RETURNS text[]
 LANGUAGE sql
 IMMUTABLE
AS $function$
  SELECT COALESCE(array_agg(x ORDER BY x), ARRAY[]::text[]) FROM (
    SELECT 'html-block'::text AS x WHERE EXISTS (
      SELECT 1 FROM jsonb_each(COALESCE(p_input_schema->'fields', '{}'::jsonb)) f
       WHERE f.value->>'source' = 'llm' AND f.value->>'type' = 'html')
    UNION
    SELECT 'list' WHERE p_html_template ~* '<(ul|ol)[\s>]' OR EXISTS (
      SELECT 1 FROM jsonb_each(COALESCE(p_input_schema->'fields', '{}'::jsonb)) f
       WHERE f.value->>'source' = 'llm' AND f.value->>'type' = 'html')
    UNION
    SELECT 'table' WHERE p_html_template ~* '<table[\s>]' OR EXISTS (
      SELECT 1 FROM jsonb_each(COALESCE(p_input_schema->'fields', '{}'::jsonb)) f
       WHERE f.value->>'source' = 'llm' AND f.value->>'type' = 'html')
    UNION
    SELECT 'items' WHERE p_html_template ~* '\{\{[-\s]*range' AND EXISTS (
      SELECT 1 FROM jsonb_each(COALESCE(p_input_schema->'fields', '{}'::jsonb)) f
       WHERE f.value->>'source' = 'llm' AND f.value->>'type' IN ('array', 'list'))
  ) s;
$function$;

UPDATE content_components
   SET input_schema = jsonb_set(
                        jsonb_set(input_schema,
                          '{fields,image_url,source}', '"site_assets.image"'),
                          '{fields,image_alt,source}',  '"site_assets.image"'),
       updated_at = NOW()
 WHERE name = 'Illustrated Text Block'
   AND input_schema->'fields'->'image_url' IS NOT NULL
   AND input_schema->'fields'->'image_alt' IS NOT NULL;

DO $$
DECLARE
    still integer;
    back  integer;
BEGIN
    SELECT count(*) INTO still FROM content_components
     WHERE 'image' = ANY (component_expresses(html_template, input_schema));
    IF still <> 0 THEN
        RAISE EXCEPTION '644 ROLLBACK: % component(s) still express image — the arm was not removed, aborting', still;
    END IF;

    SELECT count(*) INTO back FROM content_components
     WHERE name = 'Illustrated Text Block'
       AND input_schema->'fields'->'image_url'->>'source' = 'site_assets.image'
       AND input_schema->'fields'->'image_alt'->>'source' = 'site_assets.image';
    IF back <> 1 THEN
        RAISE EXCEPTION '644 ROLLBACK: restored % rows, expected 1 — aborting', back;
    END IF;
    RAISE NOTICE '644 ROLLBACK OK: image arm removed, Illustrated Text Block sources restored';
END $$;

COMMIT;
