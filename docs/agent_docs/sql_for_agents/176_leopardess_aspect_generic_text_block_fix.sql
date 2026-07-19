-- 176: restore `generic-text-block` to leopardess's AUTHORITATIVE site_plan
--      aspect on index and case-studies, so the next rebuild stops deleting
--      two live editorial sections (bugs_open/002 error C, second instance)
--
-- Context (2026-07-19). Found while fixing 002/C on robot-hands: the same
-- three-source divergence bites leopardessconsulting.co.uk, but through the
-- OTHER source. leopardess has a current site_plans row with ZERO
-- site_plan_sections rows, so source 1 misses and the site_specs.site_plan
-- aspect (source 2) is authoritative for it
-- (load_page_sections_from_spec_action.go:164-200).
--
-- 16 leopardess pages carry a deployed `generic-text-block`. It appears in the
-- aspect on NONE of them — it was added at the page_components / pages.sections
-- level on 2026-07-18 and never written back up. For 14 of those pages that is
-- harmless: they are either absent from the aspect or their aspect entry has
-- `"sections": null`, so source 2 misses and source 3 (pages.sections) governs.
--
-- For exactly TWO pages the aspect holds a real array that omits the block, so
-- source 2 wins and syncs DOWN over pages.sections, deleting it:
--   index         aspect [hero, features, differentiators-section, call-to-action]
--                 live   [hero, features, generic-text-block, differentiators-section, call-to-action]
--   case-studies  aspect [hero-case-studies, case-studies-list, call-to-action]
--                 live   [hero-case-studies, generic-text-block, case-studies-list, call-to-action]
--
-- The content at risk is deliberate and verified live 2026-07-19:
--   index        -> "The whole thing on one page / Every figure here comes from
--                    our own database and is listed in the evidence base."
--   case-studies -> "The three systems, as a route map / Three systems, one
--                    interchange. Every line stops at a person before anything
--                    is written down."
-- Both return 200 and render the copy above at https://leopardessconsulting.co.uk/
-- and /case-studies.html.
--
-- DIRECTION: align the authoritative source UP to what is deployed and live.
-- This makes no editorial decision — it preserves exactly what is on the site
-- today. If the leopardess workstream later wants these blocks gone, they
-- should be removed from the aspect AND pages.sections together.
--
-- Verify after applying (expect both to include generic-text-block):
--   SELECT pg->>'name', pg->'sections'
--   FROM site_specs ss, jsonb_array_elements(ss.data->'pages') pg
--   WHERE ss.site_id='4851f6fc-71cf-4160-a270-e03d6d3e0732'
--     AND ss.aspect='site_plan' AND ss.is_current
--     AND pg->>'name' IN ('index','case-studies');

BEGIN;

-- Rebuild the pages array in place, preserving element order (WITH ORDINALITY)
-- and touching only the two drifted pages.
UPDATE site_specs ss
SET data = jsonb_set(
        ss.data,
        '{pages}',
        (
            SELECT jsonb_agg(
                       CASE pg->>'name'
                           WHEN 'index' THEN jsonb_set(pg, '{sections}',
                               '["hero","features","generic-text-block","differentiators-section","call-to-action"]'::jsonb)
                           WHEN 'case-studies' THEN jsonb_set(pg, '{sections}',
                               '["hero-case-studies","generic-text-block","case-studies-list","call-to-action"]'::jsonb)
                           ELSE pg
                       END
                       ORDER BY ord
                   )
            FROM jsonb_array_elements(ss.data->'pages') WITH ORDINALITY AS t(pg, ord)
        )
    ),
    updated_at = NOW()
WHERE ss.site_id = '4851f6fc-71cf-4160-a270-e03d6d3e0732'
  AND ss.aspect = 'site_plan'
  AND ss.is_current = true;

COMMIT;
