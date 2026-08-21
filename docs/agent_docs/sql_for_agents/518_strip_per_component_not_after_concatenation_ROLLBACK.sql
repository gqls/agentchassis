-- 518 ROLLBACK — restore concatenate-then-strip
--
-- ⚠ DO NOT RUN THIS TO "GO BACK TO A KNOWN GOOD STATE". The state it restores is the
-- defect: a <style> block consuming the next component's prose, which measured
-- noted.co.uk/index (3 components, hero + 8KB info-card grid + CTA) as ONE character of
-- visible text, and lost more than half the visible text on 349 of 693 active pages.
--
-- It exists only so the change is reversible if the per-component form turns out to have
-- its own problem. If you run it, the floor will wrongly exclude pages again AND the
-- writer's content_sample will go back to being a fragment on multi-component pages.

BEGIN;

CREATE OR REPLACE FUNCTION page_visible_text_len(p_page_id uuid) RETURNS integer
LANGUAGE sql STABLE AS $fn$
  SELECT length(regexp_replace(regexp_replace(regexp_replace(
           COALESCE(string_agg(pc.rendered_html, ' '), ''),
           '<(style|script)[^>]*>.*?</\1>', ' ', 'gis'),
         '<[^>]+>', ' ', 'g'),
       '\s+', ' ', 'g'))
  FROM page_components pc
  WHERE pc.page_id = p_page_id
    AND pc.rendered_html IS NOT NULL
    AND COALESCE(pc.slot_name, '') NOT IN ('header','footer','head')
$fn$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,load_pages_missing_meta,config,query}',
         to_jsonb(
           'SELECT p.id, p.name, p.url, p.title, p.page_type, ' ||
           'LEFT(regexp_replace(regexp_replace(regexp_replace(regexp_replace(' ||
           'string_agg(pc.rendered_html, '' ''), ''<(style|script)[^>]*>.*?</\1>'', '' '', ''gis''), ' ||
           '''<[^>]+>'', '' '', ''g''), ''&nbsp;|&amp;|&quot;|&#39;|&lt;|&gt;'', '' '', ''g''), ' ||
           '''\s+'', '' '', ''g''), 1200) AS content_sample ' ||
           'FROM pages p ' ||
           'JOIN page_components pc ON pc.page_id = p.id ' ||
           '  AND pc.rendered_html IS NOT NULL ' ||
           '  AND COALESCE(pc.slot_name, '''') NOT IN (''header'',''footer'',''head'') ' ||
           'WHERE p.site_id = $1 AND p.status = ''active'' ' ||
           '  AND COALESCE(p.meta_description, '''') = '''' ' ||
           '  AND page_visible_text_len(p.id) > 200 ' ||
           'GROUP BY p.id, p.name, p.url, p.title, p.page_type ' ||
           'ORDER BY p.name LIMIT 25'
         )
       ),
       updated_at = now()
 WHERE type='meta-description-backfiller' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE noted int;
BEGIN
  SELECT page_visible_text_len(p.id) INTO noted FROM pages p JOIN sites s ON s.id=p.site_id
   WHERE s.domain='noted.co.uk' AND p.name='index';
  RAISE NOTICE '518 ROLLBACK done — noted.co.uk/index measures % chars again (the defect: it has 1205 of real text)', noted;
END $$;

COMMIT;
