-- 693_lendzy_adopt_three_orphan_tool_components_ROLLBACK.sql
--
-- Returns lendzy's three adopted tool pages to their pre-693 state: the
-- page_components rows go back to component_id NULL / slot_name 'section', and
-- the three adopted components are removed.
--
-- ⚠ WHAT THIS DOES NOT UNDO. Rolling back restores the DEFECT — the three pages
-- go back to failing every re-render and never earning a deployed_at. It does
-- not remove any deployed_at stamp the framework wrote in between, and it does
-- not un-deploy an artefact. If a rebuild has already shipped new bytes for
-- these pages, rolling back leaves the record pointing at nothing while the new
-- artefact serves — which is a WORSE state than either side. Check first:
--
--   SELECT name, build_status, deployed_at FROM pages
--    WHERE site_id='8ff093d5-1f19-453b-9439-a10379bbcd76'
--      AND name IN ('tool-price-cap-checker','tool-true-cost-calculator',
--                   'tool-complaint-deadline-calculator');
--
-- If any row now carries a deployed_at, do NOT roll back blindly — the repair
-- has already done its work and the reason for reverting needs re-stating.

BEGIN;

UPDATE page_components pc
   SET component_id = NULL,
       slot_name    = 'section',
       updated_at   = NOW()
  FROM pages p, content_components cc
 WHERE pc.page_id = p.id
   AND pc.component_id = cc.id
   AND p.site_id = '8ff093d5-1f19-453b-9439-a10379bbcd76'
   AND cc.created_from = 'adopted'
   AND cc.name IN ('tool-price-cap-checker-lendzy-co-uk',
                   'tool-true-cost-calculator-lendzy-co-uk',
                   'tool-complaint-deadline-calculator-lendzy-co-uk');

DELETE FROM content_components
 WHERE name IN ('tool-price-cap-checker-lendzy-co-uk',
                'tool-true-cost-calculator-lendzy-co-uk',
                'tool-complaint-deadline-calculator-lendzy-co-uk')
   AND created_from = 'adopted'
   AND NOT EXISTS (SELECT 1 FROM page_components pc WHERE pc.component_id = content_components.id);

DO $$
DECLARE comps int; nulls int;
BEGIN
  SELECT count(*) INTO comps FROM content_components
   WHERE name IN ('tool-price-cap-checker-lendzy-co-uk',
                  'tool-true-cost-calculator-lendzy-co-uk',
                  'tool-complaint-deadline-calculator-lendzy-co-uk');
  IF comps <> 0 THEN
    RAISE EXCEPTION '693 ROLLBACK: % adopted component(s) survive — a page_components row still references one', comps;
  END IF;

  SELECT count(*) INTO nulls
    FROM pages p JOIN page_components pc ON pc.page_id = p.id
   WHERE p.site_id = '8ff093d5-1f19-453b-9439-a10379bbcd76'
     AND p.name IN ('tool-price-cap-checker','tool-true-cost-calculator',
                    'tool-complaint-deadline-calculator')
     AND pc.component_id IS NULL AND pc.slot_name = 'section';
  IF nulls <> 3 THEN
    RAISE EXCEPTION '693 ROLLBACK: expected 3 rows back at component_id NULL, found %', nulls;
  END IF;

  RAISE NOTICE '693 ROLLBACK OK: 3 rows restored, 3 components removed — the defect is back in place';
END $$;

COMMIT;
