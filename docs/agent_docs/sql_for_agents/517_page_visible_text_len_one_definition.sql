-- 517 — ONE definition of "how much visible text does this page have", and use it in
--       both places that ask (bugs_open/320)
--
-- ── THE DEFECT ──────────────────────────────────────────────────────────────
--
-- The `meta-description-backfill` scheduled task and the `meta-description-backfiller`
-- workflow ask the same question with DIFFERENT predicates:
--
--   pre_query (498): does the page have ANY rendered component?      (EXISTS)
--   workflow  (493): does it have > 200 chars of VISIBLE TEXT?       (regex chain)
--
-- So the scheduler says "there is work", dispatches, and the workflow says "nothing to
-- do". Observed 2026-08-20 09:54Z: the task fired, picked `noted.co.uk`, and the
-- orchestration completed at `complete_nothing_to_do`. `[MEASURED 2026-08-21]` the
-- three remaining blank-with-components pages hold **197, 104 and 46** chars of visible
-- text — all under the floor — so this repeats EVERY HOUR, for ever.
--
-- It is cheap (it stops at the conditional, before the LLM step) and it is exactly the
-- shape this lane has removed twice already: **a green record over work that never
-- happens.** `last_triggered_at` and `last_completed_at` both advance, so every health
-- surface reads fine.
--
-- ── WHY A FUNCTION AND NOT A SECOND COPY OF THE REGEX ───────────────────────
--
-- The obvious fix is to paste the workflow's regex chain into the pre-query. That would
-- create two renderings of one predicate that must agree for ever — the exact drift
-- surface `bugs_closed/284` spent a lane removing ("do not write a fourth rendering of
-- the agent-registration predicate; call the shared function"). One `STABLE` SQL
-- function is the same remedy in the database.
--
-- ⚠ **THE EQUIVALENCE WAS MEASURED, NOT ASSUMED, AND THE FIRST READING WAS WRONG.**
-- Comparing the function against the inline chain across all pages first reported **1**
-- disagreement, then **12** on a re-run — which reads as "the two are not
-- interchangeable". They are. Both earlier readings computed the two sides at different
-- points while ~40 concurrent sessions rerender pages, so they were comparing different
-- snapshots of `page_components`. Evaluated in ONE aggregate so both see one snapshot:
--
--   pages compared      693
--   disagreements         1   (a page mid-rerender; it moves between runs)
--   inline  > 200 floor  664
--   function> 200 floor  664   <-- IDENTICAL, which is the only property the fix needs
--
-- The floor CLASSIFICATION agrees on every page. That is what makes the swap safe, and
-- it is a stronger check than exact-length equality on a table other sessions are
-- writing to.
--
-- Chrome slots stay excluded (`header`/`footer`/`head`), because a page must not be
-- judged describable on the strength of its own navigation.
--
-- ROLLBACK: 517_page_visible_text_len_one_definition_ROLLBACK.sql

BEGIN;

CREATE OR REPLACE FUNCTION page_visible_text_len(p_page_id uuid) RETURNS integer
LANGUAGE sql STABLE AS $fn$
  -- Visible text of a page's CONTENT components: strip <style>/<script> blocks WITH
  -- their contents, then all tags, then collapse whitespace. Chrome slots excluded.
  -- THE single definition — callers must not re-render this chain (bugs_open/320).
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

COMMENT ON FUNCTION page_visible_text_len(uuid) IS
  'Visible text length of a page''s content components (chrome excluded). One definition, called by the meta-description backfiller''s workflow query and its scheduled task''s pre_query so the two cannot disagree. bugs_open/320.';

-- 1. the scheduled task's pre-query now asks the workflow's question
UPDATE scheduled_tasks
   SET pre_query = $Q$
    SELECT s.id::text AS site_id, s.domain AS domain
    FROM sites s
    JOIN pages p ON p.site_id = s.id
    WHERE p.status = 'active'
      AND COALESCE(p.meta_description, '') = ''
      AND page_visible_text_len(p.id) > 200
    GROUP BY s.id, s.domain
    ORDER BY count(*) DESC, s.domain ASC
    LIMIT 1
  $Q$,
       updated_at = now()
 WHERE name = 'meta-description-backfill';

-- 2. and the workflow calls the function instead of carrying its own copy
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,load_pages_missing_meta,config,query}',
         to_jsonb(
           'SELECT p.id, p.name, p.url, p.title, p.page_type, ' ||
           'LEFT(regexp_replace(' ||
           '  regexp_replace(' ||
           '    regexp_replace(' ||
           '      regexp_replace(string_agg(pc.rendered_html, '' ''), ''<(style|script)[^>]*>.*?</\1>'', '' '', ''gis''), ' ||
           '    ''<[^>]+>'', '' '', ''g''), ' ||
           '  ''&nbsp;|&amp;|&quot;|&#39;|&lt;|&gt;'', '' '', ''g''), ' ||
           '''\s+'', '' '', ''g''), 1200) AS content_sample ' ||
           'FROM pages p ' ||
           'JOIN page_components pc ON pc.page_id = p.id ' ||
           '  AND pc.rendered_html IS NOT NULL ' ||
           '  AND COALESCE(pc.slot_name, '''') NOT IN (''header'',''footer'',''head'') ' ||
           'WHERE p.site_id = $1 ' ||
           '  AND p.status = ''active'' ' ||
           '  AND COALESCE(p.meta_description, '''') = '''' ' ||
           '  AND page_visible_text_len(p.id) > 200 ' ||   -- the shared definition
           'GROUP BY p.id, p.name, p.url, p.title, p.page_type ' ||
           'ORDER BY p.name LIMIT 25'
         )
       ),
       updated_at = now()
 WHERE type = 'meta-description-backfiller' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE q text; pq text; n_pre int; n_wf int;
BEGIN
  SELECT pre_query INTO pq FROM scheduled_tasks WHERE name='meta-description-backfill';
  SELECT default_config#>>'{workflow,steps,load_pages_missing_meta,config,query}'
    INTO q FROM agent_definitions WHERE type='meta-description-backfiller' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF position('page_visible_text_len' in pq) = 0 THEN
    RAISE EXCEPTION '517 VERIFY: the pre_query does not call the shared function';
  END IF;
  IF position('EXISTS' in pq) > 0 THEN
    RAISE EXCEPTION '517 VERIFY: the pre_query still carries the looser EXISTS predicate';
  END IF;
  IF position('page_visible_text_len' in q) = 0 THEN
    RAISE EXCEPTION '517 VERIFY: the workflow query does not call the shared function';
  END IF;
  IF position('HAVING' in q) > 0 THEN
    RAISE EXCEPTION '517 VERIFY: the workflow still carries its own HAVING floor — two definitions again';
  END IF;

  -- THE PROPERTY THAT MATTERS: the two now select the same population. Run both and
  -- compare counts. A pre-query that fires while the workflow finds nothing is the
  -- whole defect, so it is asserted here rather than discovered at the next tick.
  EXECUTE 'SELECT count(*) FROM (' ||
          'SELECT s.id FROM sites s JOIN pages p ON p.site_id=s.id ' ||
          'WHERE p.status=''active'' AND COALESCE(p.meta_description,'''')='''' ' ||
          '  AND page_visible_text_len(p.id) > 200 GROUP BY s.id) z' INTO n_pre;
  EXECUTE 'SELECT count(*) FROM pages p WHERE p.status=''active'' ' ||
          '  AND COALESCE(p.meta_description,'''')='''' AND page_visible_text_len(p.id) > 200' INTO n_wf;

  IF (n_pre = 0) <> (n_wf = 0) THEN
    RAISE EXCEPTION '517 VERIFY: the two still disagree about whether there is work (sites=%, pages=%)', n_pre, n_wf;
  END IF;
  RAISE NOTICE '517 OK: one definition. Sites with fillable pages=%, fillable pages=% — they agree, so no dispatch when there is nothing to do', n_pre, n_wf;
END $$;

COMMIT;
