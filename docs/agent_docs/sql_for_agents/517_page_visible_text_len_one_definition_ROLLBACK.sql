-- 517 ROLLBACK — restore the two independent predicates
--
-- ⚠ THIS RE-ARMS THE HOURLY NO-OP: the pre-query goes back to the looser EXISTS test,
-- so the scheduler dispatches an orchestration every hour that the workflow then
-- concludes has nothing to do. It also restores TWO renderings of one predicate.
--
-- The function is left in place deliberately — dropping it would break any other caller
-- that has since adopted it, and an unused STABLE function costs nothing. If you really
-- want it gone, check for callers first:
--   SELECT name FROM scheduled_tasks WHERE pre_query LIKE '%page_visible_text_len%';
--   SELECT type FROM agent_definitions WHERE default_config::text LIKE '%page_visible_text_len%';

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM scheduled_tasks WHERE name='meta-description-backfill';
  IF n <> 1 THEN
    RAISE EXCEPTION '517 ROLLBACK: expected 1 meta-description-backfill task, found %', n;
  END IF;
END $$;

UPDATE scheduled_tasks
   SET pre_query = $Q$
    SELECT s.id::text AS site_id, s.domain AS domain
    FROM sites s
    JOIN pages p ON p.site_id = s.id
    WHERE p.status = 'active'
      AND COALESCE(p.meta_description, '') = ''
      AND EXISTS (
        SELECT 1 FROM page_components pc
        WHERE pc.page_id = p.id
          AND pc.rendered_html IS NOT NULL
          AND COALESCE(pc.slot_name, '') NOT IN ('header','footer','head')
      )
    GROUP BY s.id, s.domain
    ORDER BY count(*) DESC, s.domain ASC
    LIMIT 1
  $Q$,
       updated_at = now()
 WHERE name = 'meta-description-backfill';

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,load_pages_missing_meta,config,query}',
         to_jsonb(
           replace(
             default_config#>>'{workflow,steps,load_pages_missing_meta,config,query}',
             '  AND page_visible_text_len(p.id) > 200 GROUP BY',
             'GROUP BY'
           ) ||
           ' HAVING length(regexp_replace(regexp_replace(regexp_replace(string_agg(pc.rendered_html, '' ''), ''<(style|script)[^>]*>.*?</\1>'', '' '', ''gis''), ''<[^>]+>'', '' '', ''g''), ''\s+'', '' '', ''g'')) > 200'
         )
       ),
       updated_at = now()
 WHERE type='meta-description-backfiller' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE pq text;
BEGIN
  SELECT pre_query INTO pq FROM scheduled_tasks WHERE name='meta-description-backfill';
  IF position('EXISTS' in pq) = 0 THEN
    RAISE EXCEPTION '517 ROLLBACK VERIFY: the pre_query was not restored';
  END IF;
  RAISE NOTICE '517 ROLLBACK OK — two predicates again, and the hourly no-op dispatch is re-armed';
END $$;

COMMIT;
