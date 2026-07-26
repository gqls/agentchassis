\set ON_ERROR_STOP on
-- Place evidence-chart at PLAN level on index and capabilities.
--
-- site_plan_sections, not page_components: instance rows are dropped by any
-- plan-driven rebuild, which is why the first five components had to be placed
-- twice. This seam is the one proven to survive (five independent rebuilds,
-- 2026-07-25).
--
-- idx_site_plan_sections_key is UNIQUE (plan_id, page_name, ordering), so an
-- in-place `ordering + 1` collides mid-shift. Shift high (+100), insert, bring
-- down (-99) — the RUNBOOK recipe.
--
-- index:        hero, stat-band, [evidence-chart], differentiators, ...
--               The stat band already says "a few real numbers from our own
--               work" in text; the chart is the next beat of that argument.
-- capabilities: ..., info-card-grid, [evidence-chart], call-to-action
--               Two charts land here (news pipeline, council outcomes) because
--               that is where the "how we prove it" material sits.
--
-- Which chart appears on which page is decided by the chart definitions'
-- `pages` key in the evidence_base, not here.

\set plan_id '81741260-6447-492c-bf98-4b3c185f8e7b'

BEGIN;

-- index: open a gap at ordering 2
UPDATE site_plan_sections SET ordering = ordering + 100
 WHERE plan_id = :'plan_id'::uuid AND page_name = 'index' AND ordering >= 2;
INSERT INTO site_plan_sections (plan_id, page_name, ordering, component_name)
VALUES (:'plan_id'::uuid, 'index', 2, 'evidence-chart');
UPDATE site_plan_sections SET ordering = ordering - 99
 WHERE plan_id = :'plan_id'::uuid AND page_name = 'index' AND ordering >= 100;

-- capabilities: open a gap at ordering 4 (before call-to-action)
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
    SELECT page_name, count(*) AS n, max(ordering) AS mx, min(ordering) AS mn
      FROM site_plan_sections
     WHERE plan_id = '81741260-6447-492c-bf98-4b3c185f8e7b' AND page_name IN ('index','capabilities')
     GROUP BY page_name
    HAVING min(ordering) <> 0 OR max(ordering) <> count(*) - 1) q;
  IF bad > 0 THEN
    RAISE EXCEPTION 'ordering is no longer a dense 0..n-1 sequence on % page(s)', bad;
  END IF;
END $$;

COMMIT;

SELECT page_name, ordering, component_name FROM site_plan_sections
 WHERE plan_id = :'plan_id'::uuid AND page_name IN ('index','capabilities')
 ORDER BY page_name, ordering;
