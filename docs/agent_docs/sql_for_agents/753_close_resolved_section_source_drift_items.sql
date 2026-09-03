-- 753: close section_source_drift work items whose stores AGREE AGAIN, and
--      record WHICH SIDE won so a destroyed correction is not closed silently
--
-- Context (2026-09-03). check_section_source_drift is flag-only:
--   HandlerAgent: ""   Status: "needs_human_review"
--   "no handler auto-aligns planning sources; a human picks the intended layout"
-- Nothing ever closes one. [MEASURED 2026-09-03] six items were open, the
-- oldest filed 2026-07-28 -- 37 days. Four describe a drift that no longer
-- exists. The item's spec.authoritative / spec.pages_sections are frozen at
-- FILING time and read as current, so the backlog cannot be triaged by reading
-- the items.
--
-- Closing is SELF-CORRECTING, which is what makes it safe rather than tidy:
-- idx_swi_dedup is UNIQUE (site_id, item_key) WHERE status NOT IN
-- ('complete','verified','rejected','wont_fix','failed','unresolved',
-- 'cancelled'), so closing frees the key and the next completeness-discovery
-- pass re-files anything still real. It also UNBLOCKS re-filing: while a stale
-- item sits open, a genuinely NEW drift on the same page is invisible behind it.
--
-- CLOSED BY PREDICATE, NOT BY AN ID LIST. The set is re-derived inside this
-- transaction from the live stores, mirroring the check's own precedence
-- (tier 1 site_plan_sections for the current plan, else tier 2 the
-- site_specs.site_plan aspect) against pages.sections. A page that has drifted
-- between measurement and apply is therefore NOT closed, and apis.uk is
-- excluded by the DATA rather than by a hand-maintained exception list.
--
-- ⚠ THE DIRECTION TEST IS THE POINT. Comparing today's agreed list against the
-- item's two frozen lists says WHO WON:
--   agrees at spec.pages_sections -> the live composition held. Benign.
--   agrees at spec.authoritative  -> THE CACHE WAS OVERWRITTEN. Something a
--                                    person put on that page was destroyed by
--                                    the tier-1 sync-down. The drift is gone,
--                                    so the item must close -- but closing it
--                                    as a plain success would RATIFY the loss.
--   agrees at neither             -> a genuine re-plan moved both.
-- Every row therefore carries `direction` in its result. [MEASURED 2026-09-03]
-- three of the four closed here are `authority_won`: robot-hands.com/
-- gripper-catalog lost `gripper-spec-sheet` -- the very component migration 154
-- was written to rescue in July -- and idea.uk/guides-index lost `guide-list`.
-- Those two are filed as a bug in their own right; this migration only stops
-- them blocking the detector.
--
-- Raw comparison is sound even though the check compares through
-- datahelpers.MergeLockedPageSlots: the merge is applied to BOTH sides with the
-- SAME locked rows, so raw-equal implies merged-equal. Raw is a safe
-- over-approximation of "agrees", i.e. it can only ever close FEWER items than
-- the check would consider resolved, never more.
--
-- Verify after applying:
--   SELECT s.domain, wi.spec->>'page_name', wi.status, wi.result->>'direction'
--     FROM site_work_items wi JOIN sites s ON s.id=wi.site_id
--    WHERE wi.item_type='section_source_drift' ORDER BY wi.created_at;
--   -- expect apis.uk STILL needs_human_review; the rest complete with a direction

BEGIN;

DO $pre753$
DECLARE
    v_open int;
BEGIN
    SELECT count(*) INTO v_open FROM site_work_items
     WHERE item_type='section_source_drift' AND status NOT IN ('complete','cancelled','rejected');
    IF v_open = 0 THEN
        RAISE EXCEPTION '753 ABORT: no open section_source_drift items at all — nothing to close, and this is not the state this migration was written against.';
    END IF;
END
$pre753$;

WITH tier1 AS (
    SELECT sp.site_id, sps.page_name,
           jsonb_agg(sps.component_name ORDER BY sps.ordering) AS auth
      FROM site_plans sp
      JOIN site_plan_sections sps ON sps.plan_id = sp.id
     WHERE sp.is_current
     GROUP BY sp.site_id, sps.page_name
), tier2 AS (
    SELECT ss.site_id, pg->>'name' AS page_name, pg->'sections' AS auth
      FROM site_specs ss,
           LATERAL jsonb_array_elements(ss.data->'pages') pg
     WHERE ss.aspect = 'site_plan' AND ss.is_current
       AND jsonb_typeof(ss.data->'pages') = 'array'
), resolved AS (
    SELECT wi.id,
           COALESCE(t1.auth, t2.auth) AS live_auth,
           p.sections                 AS live_cache,
           CASE WHEN t1.auth IS NOT NULL THEN 'site_plan_sections'
                WHEN t2.auth IS NOT NULL THEN 'site_specs.site_plan'
                ELSE 'none' END       AS serving_source,
           CASE WHEN p.sections = wi.spec->'pages_sections' THEN 'cache_held'
                WHEN p.sections = wi.spec->'authoritative'  THEN 'authority_won'
                ELSE 'third_list' END AS direction
      FROM site_work_items wi
      JOIN pages p ON p.site_id = wi.site_id AND p.name = wi.spec->>'page_name'
      LEFT JOIN tier1 t1 ON t1.site_id = wi.site_id AND t1.page_name = wi.spec->>'page_name'
      LEFT JOIN tier2 t2 ON t2.site_id = wi.site_id AND t2.page_name = wi.spec->>'page_name'
     WHERE wi.item_type = 'section_source_drift'
       AND wi.status NOT IN ('complete','cancelled','rejected')
       AND COALESCE(t1.auth, t2.auth) IS NOT DISTINCT FROM p.sections
)
UPDATE site_work_items wi
   SET status = 'complete',
       updated_at = NOW(),
       result = jsonb_build_object(
           'closed_by',      'migration 753',
           'closed_on',      '2026-09-03',
           'resolution',     'the serving authority and pages.sections agree again; the drift this item describes no longer exists',
           'serving_source', r.serving_source,
           'live_list',      r.live_cache,
           'direction',      r.direction,
           'direction_note', CASE r.direction
               WHEN 'authority_won' THEN 'THE CACHE WAS OVERWRITTEN — whatever was edited into pages.sections was destroyed by the tier-1 sync-down. Closed only so it stops blocking the detector; the loss itself is filed separately.'
               WHEN 'cache_held'    THEN 'the live composition held; benign.'
               ELSE 'neither frozen list matches — a re-plan moved both sides.' END)
  FROM resolved r
 WHERE wi.id = r.id;

DO $post753$
DECLARE
    v_apis int;
    v_bad  int;
BEGIN
    -- apis.uk must still be OPEN: it is a live divergence and another lane owns it.
    SELECT count(*) INTO v_apis
      FROM site_work_items wi JOIN sites s ON s.id = wi.site_id
     WHERE wi.item_type='section_source_drift' AND s.domain='apis.uk'
       AND wi.status = 'needs_human_review';
    IF v_apis <> 1 THEN
        RAISE EXCEPTION '753 VERIFY FAILED: apis.uk/index should still be open (found % open items). It is a LIVE divergence owned by another lane and must not be closed.', v_apis;
    END IF;

    -- nothing closed here may still be divergent.
    SELECT count(*) INTO v_bad
      FROM site_work_items wi
      JOIN pages p ON p.site_id = wi.site_id AND p.name = wi.spec->>'page_name'
      LEFT JOIN LATERAL (
          SELECT jsonb_agg(sps.component_name ORDER BY sps.ordering) AS auth
            FROM site_plans sp JOIN site_plan_sections sps ON sps.plan_id = sp.id
           WHERE sp.site_id = wi.site_id AND sp.is_current AND sps.page_name = wi.spec->>'page_name'
      ) t1 ON true
      LEFT JOIN LATERAL (
          SELECT pg->'sections' AS auth
            FROM site_specs ss, LATERAL jsonb_array_elements(ss.data->'pages') pg
           WHERE ss.site_id = wi.site_id AND ss.aspect='site_plan' AND ss.is_current
             AND pg->>'name' = wi.spec->>'page_name'
      ) t2 ON true
     WHERE wi.item_type='section_source_drift'
       AND wi.result->>'closed_by' = 'migration 753'
       AND COALESCE(t1.auth, t2.auth) IS DISTINCT FROM p.sections;
    IF v_bad <> 0 THEN
        RAISE EXCEPTION '753 VERIFY FAILED: % item(s) closed while still divergent.', v_bad;
    END IF;

    -- every row closed must carry a direction.
    SELECT count(*) INTO v_bad FROM site_work_items
     WHERE item_type='section_source_drift' AND result->>'closed_by'='migration 753'
       AND COALESCE(result->>'direction','') = '';
    IF v_bad <> 0 THEN
        RAISE EXCEPTION '753 VERIFY FAILED: % closed item(s) carry no direction — the loss/benign distinction is the whole point of this migration.', v_bad;
    END IF;
END
$post753$;

COMMIT;
