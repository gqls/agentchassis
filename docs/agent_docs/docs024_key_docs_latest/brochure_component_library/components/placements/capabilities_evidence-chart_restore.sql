\set ON_ERROR_STOP on
-- Restore evidence-chart on `capabilities` at PLAN level.
--
-- WHY THIS IS A SEPARATE FILE from index_capabilities_evidence-chart.sql:
-- that script places BOTH pages, and `index` already carries the section at
-- ordering 2. Re-running it would insert a second one. This is the capabilities
-- half only, and it refuses to run if the section is already there.
--
-- WHY IT WAS REMOVED, and why it is safe to put back (2026-07-27):
-- the capabilities placement was pulled on 2026-07-26 as containment for
-- bugs_open/085 — no page identity reached a section template, so all three
-- charts rendered on every page carrying the section and the site would have
-- published the same three charts twice. 085 is now fixed on BOTH render paths
-- and verified live: `index` carries exactly the one chart declared for it
-- (v1.0.1173 build path, v1.0.1174 scoped path). The `pages` key in the
-- evidence_base decides which charts land here — news-pipeline-credibility and
-- council-review-outcomes — and nothing in this file repeats that decision.
--
-- idx_site_plan_sections_key is UNIQUE (plan_id, page_name, ordering), so an
-- in-place `ordering + 1` collides mid-shift. Shift high (+100), insert, bring
-- down (-99) — the RUNBOOK recipe.
--
-- capabilities: hero-services, hero-card-carousel, services-grid,
--               info-card-grid, [evidence-chart], call-to-action
--               The chart is the last beat before the ask, which is where the
--               "how we prove it" material belongs.
--
-- PLAN LEVEL ONLY. This does not create a page_components row and does not put
-- the section on the live page — the page must be built through the pipeline so
-- the writer authors the eyebrow/title/intro and the resolver supplies charts
-- and facts. Never hand-author content_data.

\set plan_id '81741260-6447-492c-bf98-4b3c185f8e7b'

BEGIN;

-- Refuse rather than duplicate.
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM site_plan_sections
              WHERE plan_id = '81741260-6447-492c-bf98-4b3c185f8e7b'
                AND page_name = 'capabilities' AND component_name = 'evidence-chart') THEN
    RAISE EXCEPTION 'evidence-chart is already in the capabilities plan — nothing to restore';
  END IF;
END $$;

UPDATE site_plan_sections SET ordering = ordering + 100
 WHERE plan_id = :'plan_id'::uuid AND page_name = 'capabilities' AND ordering >= 4;
INSERT INTO site_plan_sections (plan_id, page_name, ordering, component_name)
VALUES (:'plan_id'::uuid, 'capabilities', 4, 'evidence-chart');
UPDATE site_plan_sections SET ordering = ordering - 99
 WHERE plan_id = :'plan_id'::uuid AND page_name = 'capabilities' AND ordering >= 100;

-- Refuse to leave a gap or a duplicate behind.
DO $$
DECLARE bad int;
BEGIN
  SELECT count(*) INTO bad FROM (
    SELECT page_name FROM site_plan_sections
     WHERE plan_id = '81741260-6447-492c-bf98-4b3c185f8e7b' AND page_name = 'capabilities'
     GROUP BY page_name
    HAVING min(ordering) <> 0 OR max(ordering) <> count(*) - 1) q;
  IF bad > 0 THEN
    RAISE EXCEPTION 'capabilities ordering is no longer a dense 0..n-1 sequence';
  END IF;
END $$;

COMMIT;

SELECT page_name, ordering, component_name
  FROM site_plan_sections
 WHERE plan_id = :'plan_id'::uuid AND page_name = 'capabilities'
 ORDER BY ordering;
