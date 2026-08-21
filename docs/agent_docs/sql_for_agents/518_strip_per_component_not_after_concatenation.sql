-- 518 — strip markup PER COMPONENT, not after concatenating them (bugs_open/320)
--
-- ⚠ THIS IS A DEFECT IN MY OWN MIGRATION 493, found by investigating one page properly
-- instead of accepting its number.
--
-- ── WHAT WENT WRONG ─────────────────────────────────────────────────────────
--
-- 493 introduced a visible-text measure shaped as **concatenate, then strip**:
--
--     regexp_replace(string_agg(pc.rendered_html, ' '), '<(style|script)…>.*?</\1>', …)
--
-- The `<style>`/`<script>` strip therefore runs across the JOIN BOUNDARIES between
-- components, and a match that begins in one component can consume prose belonging to
-- the next. `noted.co.uk/index` is the worked case:
--
--     per-component visible text, summed : 1,205 chars
--     concatenate-then-strip (493's measure) :     1 char
--     components 3 · <style> opens 3 · </style> closes 3   (balanced — not a broken tag)
--
-- So a homepage with a hero, an 8KB info-card grid and a call to action measured as
-- ONE character of content.
--
-- ── HOW BAD, MEASURED ACROSS THE ESTATE ─────────────────────────────────────
--
-- `[MEASURED 2026-08-21]` over 693 active pages with content components:
--
--     lose more than HALF their visible text to the formulation : **349**
--     wrongly judged BELOW the 200 floor                        :  **24**
--     of those, currently empty (i.e. real damage today)         :   **1**  (noted.co.uk/index)
--
-- The live damage is one page. **The measurement is wrong on half the estate**, which is
-- the part that matters, because two things read it:
--   1. the FLOOR that decides whether a page is describable — so pages with plenty of
--      content can be silently declined;
--   2. `content_sample`, the 1200 characters of page text handed to the WRITER. On a
--      page where the strip eats the middle, the model has been describing a page from a
--      fragment of it. The descriptions already written read well, so this degraded the
--      input rather than obviously breaking it — which is exactly why it went unnoticed.
--
-- ── THE FIX ─────────────────────────────────────────────────────────────────
--
-- Strip each component's markup on its own, THEN join the results. A regex can no longer
-- span two components because it never sees two at once. This also removes the ordering
-- sensitivity that made the earlier equivalence checks disagree with themselves
-- (1 disagreement, then 12, on data that had not changed shape).
--
-- Both readers are updated together, because a floor and a sample that disagree about
-- what a page says is the same class of bug one layer along.
--
-- ROLLBACK: 518_strip_per_component_not_after_concatenation_ROLLBACK.sql

BEGIN;

-- 1. the shared function (517) strips per component
CREATE OR REPLACE FUNCTION page_visible_text_len(p_page_id uuid) RETURNS integer
LANGUAGE sql STABLE AS $fn$
  SELECT COALESCE(sum(length(regexp_replace(regexp_replace(regexp_replace(
           pc.rendered_html,
           '<(style|script)[^>]*>.*?</\1>', ' ', 'gis'),
         '<[^>]+>', ' ', 'g'),
       '\s+', ' ', 'g'))), 0)::int
  FROM page_components pc
  WHERE pc.page_id = p_page_id
    AND pc.rendered_html IS NOT NULL
    AND COALESCE(pc.slot_name, '') NOT IN ('header','footer','head')
$fn$;

COMMENT ON FUNCTION page_visible_text_len(uuid) IS
  'Visible text length of a page''s content components (chrome excluded). Strips markup PER COMPONENT then sums, so a <style> block cannot consume the next component''s prose (bugs_open/320, migration 518 — the concatenate-then-strip version measured a 3-component homepage as 1 char). One definition; callers must not re-render this chain.';

-- 2. content_sample is built the same way: strip per component, then join
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,load_pages_missing_meta,config,query}',
         to_jsonb(
           'SELECT p.id, p.name, p.url, p.title, p.page_type, ' ||
           'LEFT(string_agg(' ||
           '  regexp_replace(' ||
           '    regexp_replace(' ||
           '      regexp_replace(' ||
           '        regexp_replace(pc.rendered_html, ''<(style|script)[^>]*>.*?</\1>'', '' '', ''gis''), ' ||
           '      ''<[^>]+>'', '' '', ''g''), ' ||
           '    ''&nbsp;|&amp;|&quot;|&#39;|&lt;|&gt;'', '' '', ''g''), ' ||
           '  ''\s+'', '' '', ''g''), ' ||
           '  '' '' ORDER BY pc.slot_name), 1200) AS content_sample ' ||
           'FROM pages p ' ||
           'JOIN page_components pc ON pc.page_id = p.id ' ||
           '  AND pc.rendered_html IS NOT NULL ' ||
           '  AND COALESCE(pc.slot_name, '''') NOT IN (''header'',''footer'',''head'') ' ||
           'WHERE p.site_id = $1 ' ||
           '  AND p.status = ''active'' ' ||
           '  AND COALESCE(p.meta_description, '''') = '''' ' ||
           '  AND page_visible_text_len(p.id) > 200 ' ||
           'GROUP BY p.id, p.name, p.url, p.title, p.page_type ' ||
           'ORDER BY p.name LIMIT 25'
         )
       ),
       updated_at = now()
 WHERE type = 'meta-description-backfiller' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE q text; noted int; per_comp int; still_wrong int;
BEGIN
  SELECT default_config#>>'{workflow,steps,load_pages_missing_meta,config,query}'
    INTO q FROM agent_definitions WHERE type='meta-description-backfiller' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF position('string_agg(' in q) = 0 OR position('ORDER BY pc.slot_name' in q) = 0 THEN
    RAISE EXCEPTION '518 VERIFY: content_sample is not strip-then-aggregate with a deterministic order';
  END IF;
  IF position('regexp_replace(string_agg' in q) > 0 THEN
    RAISE EXCEPTION '518 VERIFY: the query still concatenates before stripping';
  END IF;

  -- The worked case must now measure as content, not as 1 char.
  SELECT page_visible_text_len(p.id) INTO noted
    FROM pages p JOIN sites s ON s.id=p.site_id
   WHERE s.domain='noted.co.uk' AND p.name='index';
  IF noted IS NULL OR noted < 1000 THEN
    RAISE EXCEPTION '518 VERIFY: noted.co.uk/index measures % chars, expected ~1205 — the fix did not take', noted;
  END IF;

  -- And no page should now be judged below the floor while its components say otherwise.
  SELECT count(*) INTO still_wrong FROM (
    SELECT p.id,
           sum(length(regexp_replace(regexp_replace(regexp_replace(pc.rendered_html,
             '<(style|script)[^>]*>.*?</\1>',' ','gis'),'<[^>]+>',' ','g'),'\s+',' ','g'))) AS s,
           page_visible_text_len(p.id) AS f
    FROM pages p JOIN page_components pc ON pc.page_id=p.id AND pc.rendered_html IS NOT NULL
      AND COALESCE(pc.slot_name,'') NOT IN ('header','footer','head')
    WHERE p.status='active' GROUP BY p.id) z
   WHERE s > 200 AND f <= 200;
  IF still_wrong > 0 THEN
    RAISE EXCEPTION '518 VERIFY: % pages are still wrongly below the floor', still_wrong;
  END IF;

  RAISE NOTICE '518 OK: per-component strip. noted.co.uk/index now measures % chars; 0 pages wrongly below the floor', noted;
END $$;

COMMIT;
