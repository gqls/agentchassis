-- 478_309_blog_listing_collection_dialect_ROLLBACK.sql
--
-- Restores blog-listing_pre_037's input_schema and html_template from the
-- backup table migration 478 created. Refuses if that table is missing.
--
-- NOTE: rolling back re-arms bugs_open/309 -- the cards lose their anchors
-- again on the next re-plan, and the component re-declares a source the
-- source-vocabulary birth gate now refuses.

BEGIN;

DO $rb478$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM information_schema.tables
     WHERE table_name = 'content_components_bak_20260818_309_blog_listing';
    IF n <> 1 THEN
        RAISE EXCEPTION 'ROLLBACK 478: backup table content_components_bak_20260818_309_blog_listing is missing -- cannot restore';
    END IF;
END
$rb478$;

UPDATE content_components cc
   SET input_schema  = b.input_schema,
       html_template = b.html_template,
       updated_at    = now()
  FROM content_components_bak_20260818_309_blog_listing b
 WHERE cc.id = b.id;

DO $rb478$
DECLARE phantom int;
BEGIN
    SELECT count(*) INTO phantom
      FROM content_components cc, jsonb_each(cc.input_schema->'fields') f
     WHERE cc.name = 'blog-listing_pre_037'
       AND f.value->>'source' LIKE 'site_specs.blog.%';
    IF phantom = 0 THEN
        RAISE EXCEPTION 'ROLLBACK 478 VERIFY: restore did not bring back the original schema';
    END IF;
    RAISE NOTICE 'rollback 478: restored (% phantom source(s) back, as expected)', phantom;
END
$rb478$;

COMMIT;
